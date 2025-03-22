package service

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/filecoin-project/lotus/api/client"

	logging "github.com/ipfs/go-log/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/storacha/storage/pkg/pdp/ethereum"
	"github.com/storacha/storage/pkg/pdp/scheduler"
	"github.com/storacha/storage/pkg/pdp/service/models"
	"github.com/storacha/storage/pkg/pdp/store"
	"github.com/storacha/storage/pkg/pdp/tasks"
	"github.com/storacha/storage/pkg/store/blobstore"
)

var log = logging.Logger("pdp/service")

type PDPService struct {
	address   common.Address
	blobstore blobstore.Blobstore
	storage   *store.LocalStashStore
	sender    ethereum.Sender

	db   *gorm.DB
	name string

	chainScheduler *scheduler.Chain
	engine         *scheduler.TaskEngine

	stopFns []func()
}

func (p *PDPService) Stop(ctx context.Context) error {
	// TODO either use uber fx, or a multierror
	for _, stopFn := range p.stopFns {
		stopFn()
	}
	return nil
}

func NewPDPService(ctx context.Context, address common.Address, bs blobstore.Blobstore, ss *store.LocalStashStore) (*PDPService, error) {
	var stopFns = []func(){}
	// NB: must use web socket to get chain head notifications
	chainClient, closer, err := client.NewFullNodeRPCV1(ctx, "ws://127.0.0.1:1234/rpc/v1", nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to lotus client: %w", err)
	}
	headChan, err := chainClient.ChainNotify(ctx)
	if err != nil {
		panic(err)
	}
	chainHead := <-headChan
	log.Infow("Got a chain head", "head", chainHead)

	stopFns = append(stopFns, closer)

	eClient, err := ethclient.Dial("http://127.0.0.1:1234/rpc/v1")
	if err != nil {
		return nil, fmt.Errorf("connecting to eth client: %w", err)
	}
	stopFns = append(stopFns, eClient.Close)

	chainScheduler := scheduler.NewChain(chainClient)

	dsn := "host=localhost user=postgres dbname=postgres port=5432 sslmode=disable"
	db, err := setupDatabase(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	var t []scheduler.TaskInterface
	sender, senderTask := tasks.NewSenderETH(eClient, db)
	t = append(t, senderTask)

	pdpInitTask, err := tasks.NewInitProvingPeriodTask(db, eClient, chainClient, chainScheduler, sender)
	if err != nil {
		return nil, fmt.Errorf("creating init proving period task: %w", err)
	}
	t = append(t, pdpInitTask)

	pdpNextTask, err := tasks.NewNextProvingPeriodTask(db, eClient, chainClient, chainScheduler, sender)
	if err != nil {
		return nil, fmt.Errorf("creating next proving period task: %w", err)
	}
	t = append(t, pdpNextTask)

	pdpNotifyTask := tasks.NewPDPNotifyTask(db)
	t = append(t, pdpNotifyTask)

	pdpProveTask, err := tasks.NewProveTask(chainScheduler, db, eClient, chainClient, sender, bs)
	if err != nil {
		return nil, fmt.Errorf("creating prove period task: %w", err)
	}
	t = append(t, pdpProveTask)

	if err := tasks.NewWatcherCreate(db, eClient, chainScheduler); err != nil {
		return nil, fmt.Errorf("creating watcher root create: %w", err)
	}

	if err := tasks.NewWatcherRootAdd(db, eClient, chainScheduler); err != nil {
		return nil, fmt.Errorf("creating watcher root add: %w", err)
	}

	// TODO this needs configuration and or tuning
	maxMoveStoreTasks := 8
	pdpStorePieceTask, err := tasks.NewStorePieceTask(db, bs, maxMoveStoreTasks)
	if err != nil {
		return nil, fmt.Errorf("creating pdp store piece task: %w", err)
	}
	t = append(t, pdpStorePieceTask)

	engine, err := scheduler.NewEngine(db, t)
	if err != nil {
		return nil, fmt.Errorf("creating engine: %w", err)
	}
	stopFns = append(stopFns, engine.GracefullyTerminate)

	// TODO this needs to be manually stopped
	ethWatcher, err := tasks.NewMessageWatcherEth(db, chainScheduler, eClient)
	if err != nil {
		return nil, fmt.Errorf("creating message watcher: %w", err)
	}

	go chainScheduler.Run(ctx)
	stopFns = append(stopFns, func() {
		if err := ethWatcher.Stop(context.TODO()); err != nil {
			log.Errorf("failed to stop eth watcher: %v", err)
		}
	})

	return &PDPService{
		address:        address,
		db:             db,
		name:           "storacha",
		blobstore:      bs,
		storage:        ss,
		sender:         sender,
		stopFns:        stopFns,
		engine:         engine,
		chainScheduler: chainScheduler,
	}, nil
}

func setupDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}
	if err := db.AutoMigrate(
		&models.Machine{},
		&models.Task{},
		&models.TaskHistory{},
		&models.TaskFollow{},
		&models.TaskImpl{},

		&models.ParkedPiece{},
		&models.ParkedPieceRef{},

		&models.PDPService{},
		&models.PDPPieceUpload{},
		&models.PDPPieceRef{},
		&models.PDPProofSet{},
		&models.PDPProveTask{},
		&models.PDPProofsetCreate{},
		&models.PDPProofsetRoot{},
		&models.PDPProofsetRootAdd{},
		&models.PDPPieceMHToCommp{},

		&models.EthKey{},
		&models.MessageSendsEth{},
		&models.MessageSendEthLock{},
		&models.MessageWaitsEth{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %s", err)
	}
	if err := db.Exec(models.Triggers).Error; err != nil {
		return nil, fmt.Errorf("failed to install database triggers: %s", err)
	}
	return db, nil
}
