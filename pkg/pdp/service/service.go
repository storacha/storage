package service

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/hashicorp/go-multierror"

	logging "github.com/ipfs/go-log/v2"
	"gorm.io/gorm"

	"github.com/storacha/storage/pkg/pdp/ethereum"
	"github.com/storacha/storage/pkg/pdp/scheduler"
	"github.com/storacha/storage/pkg/pdp/service/models"
	"github.com/storacha/storage/pkg/pdp/store"
	"github.com/storacha/storage/pkg/pdp/tasks"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/wallet"
)

var log = logging.Logger("pdp/service")

type PDPService struct {
	address   common.Address
	blobstore blobstore.Blobstore
	storage   store.Stash
	sender    ethereum.Sender

	db   *gorm.DB
	name string

	chainScheduler *scheduler.Chain
	engine         *scheduler.TaskEngine

	stopFns  []func(ctx context.Context) error
	startFns []func(ctx context.Context) error
}

func (p *PDPService) Storage() blobstore.Blobstore {
	return p.blobstore
}

func (p *PDPService) Start(ctx context.Context) error {
	for _, startFn := range p.startFns {
		if err := startFn(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *PDPService) Stop(ctx context.Context) error {
	var errs error
	for _, stopFn := range p.stopFns {
		if err := stopFn(ctx); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

type ChainClient interface {
	ChainHead(ctx context.Context) (*types.TipSet, error)
	ChainNotify(ctx context.Context) (<-chan []*api.HeadChange, error)
	StateGetRandomnessDigestFromBeacon(ctx context.Context, randEpoch abi.ChainEpoch, tsk types.TipSetKey) (abi.Randomness, error)
}

type ContractClient interface {
	tasks.SenderETHClient
	tasks.MessageWatcherEthClient
	bind.ContractBackend
}

func NewPDPService(
	dialector gorm.Dialector,
	address common.Address,
	wallet wallet.Wallet,
	bs blobstore.Blobstore,
	ss store.Stash,
	chainClient ChainClient,
	ethClient ContractClient,
) (*PDPService, error) {
	var (
		startFns []func(context.Context) error
		stopFns  []func(context.Context) error
	)
	chainScheduler := scheduler.NewChain(chainClient)

	db, err := setupDatabase(dialector)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	var t []scheduler.TaskInterface
	sender, senderTask := tasks.NewSenderETH(ethClient, wallet, db)
	t = append(t, senderTask)

	pdpInitTask, err := tasks.NewInitProvingPeriodTask(db, ethClient, chainClient, chainScheduler, sender)
	if err != nil {
		return nil, fmt.Errorf("creating init proving period task: %w", err)
	}
	t = append(t, pdpInitTask)

	pdpNextTask, err := tasks.NewNextProvingPeriodTask(db, ethClient, chainClient, chainScheduler, sender)
	if err != nil {
		return nil, fmt.Errorf("creating next proving period task: %w", err)
	}
	t = append(t, pdpNextTask)

	pdpNotifyTask := tasks.NewPDPNotifyTask(db)
	t = append(t, pdpNotifyTask)

	pdpProveTask, err := tasks.NewProveTask(chainScheduler, db, ethClient, chainClient, sender, bs)
	if err != nil {
		return nil, fmt.Errorf("creating prove period task: %w", err)
	}
	t = append(t, pdpProveTask)

	if err := tasks.NewWatcherCreate(db, ethClient, chainScheduler); err != nil {
		return nil, fmt.Errorf("creating watcher root create: %w", err)
	}

	if err := tasks.NewWatcherRootAdd(db, chainScheduler); err != nil {
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
	stopFns = append(stopFns, func(ctx context.Context) error {
		engine.GracefullyTerminate()
		return nil
	})

	// TODO this needs to be manually stopped
	ethWatcher, err := tasks.NewMessageWatcherEth(db, chainScheduler, ethClient)
	if err != nil {
		return nil, fmt.Errorf("creating message watcher: %w", err)
	}

	startFns = append(startFns, func(ctx context.Context) error {
		go chainScheduler.Run(ctx)
		return nil
	})
	stopFns = append(stopFns, ethWatcher.Stop)

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

func setupDatabase(d gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(d)
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
