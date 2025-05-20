package service

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/hashicorp/go-multierror"
	"gorm.io/gorm/logger"

	logging "github.com/ipfs/go-log/v2"
	"gorm.io/gorm"

	"github.com/storacha/storage/pkg/pdp/ethereum"
	"github.com/storacha/storage/pkg/pdp/scheduler"
	"github.com/storacha/storage/pkg/pdp/service/contract"
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

type EthClient interface {
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
	ethClient EthClient,
	contractClient contract.PDP,
) (*PDPService, error) {
	var (
		startFns []func(context.Context) error
		stopFns  []func(context.Context) error
	)
	chainScheduler := scheduler.NewChain(chainClient)

	db, err := SetupDatabase(dialector)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	var t []scheduler.TaskInterface
	sender, senderTask := tasks.NewSenderETH(ethClient, wallet, db)
	t = append(t, senderTask)

	pdpInitTask, err := tasks.NewInitProvingPeriodTask(db, ethClient, contractClient, chainClient, chainScheduler, sender)
	if err != nil {
		return nil, fmt.Errorf("creating init proving period task: %w", err)
	}
	t = append(t, pdpInitTask)

	pdpNextTask, err := tasks.NewNextProvingPeriodTask(db, ethClient, contractClient, chainClient, chainScheduler, sender)
	if err != nil {
		return nil, fmt.Errorf("creating next proving period task: %w", err)
	}
	t = append(t, pdpNextTask)

	pdpNotifyTask := tasks.NewPDPNotifyTask(db)
	t = append(t, pdpNotifyTask)

	pdpProveTask, err := tasks.NewProveTask(chainScheduler, db, ethClient, contractClient, chainClient, sender, bs)
	if err != nil {
		return nil, fmt.Errorf("creating prove period task: %w", err)
	}
	t = append(t, pdpProveTask)

	if err := tasks.NewWatcherCreate(db, ethClient, contractClient, chainScheduler); err != nil {
		return nil, fmt.Errorf("creating watcher root create: %w", err)
	}

	if err := tasks.NewWatcherRootAdd(db, chainScheduler, contractClient); err != nil {
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
		startFns:       startFns,
		stopFns:        stopFns,
		engine:         engine,
		chainScheduler: chainScheduler,
	}, nil
}

func SetupDatabase(d gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(d, &gorm.Config{
		// No need to run every operation in a transaction, we are explicit about where transactions are required.
		SkipDefaultTransaction: true,
		Logger:                 NewGormLogger(log),
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

// GormLogger adapts the project's logging system to GORM's logging interface.
// It ensures consistent logging across the application regardless of whether
// the log is coming from GORM or the application code.
type GormLogger struct {
	log     *logging.ZapEventLogger
	level   logger.LogLevel
	slowSQL time.Duration // threshold for slow SQL logging
}

// NewGormLogger creates a new GormLogger with appropriate defaults.
func NewGormLogger(log *logging.ZapEventLogger) *GormLogger {
	return &GormLogger{
		log:     log,
		level:   logger.Info, // Default to Info level
		slowSQL: time.Second, // Default threshold for slow SQL
	}
}

// LogMode sets the log level for GORM and returns an updated logger.
// This allows dynamic configuration of logging level.
func (g *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *g
	newLogger.level = level
	return &newLogger
}

// Info logs info messages using the application's logger.
func (g *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	if g.level >= logger.Info {
		g.log.Infof(s, i...)
	}
}

// Warn logs warning messages using the application's logger.
func (g *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	if g.level >= logger.Warn {
		g.log.Warnf(s, i...)
	}
}

// Error logs error messages using the application's logger.
func (g *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	if g.level >= logger.Error {
		g.log.Errorf(s, i...)
	}
}

// getCallerInfo retrieves file, line, and function information from the call stack
// skipFrames specifies how many call frames to skip upward in the stack
// maxFrames specifies how many call frames to capture
func getCallerInfo(skipFrames, maxFrames int) []string {
	var callers []string

	for i := skipFrames; i < skipFrames+maxFrames; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Get function name
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// Extract just the package and function name, not the full path
		funcName := fn.Name()

		// Extract just the filename, not the full path
		_, fileName := path.Split(file)

		// Format as "file:line function"
		callerInfo := fmt.Sprintf("%s:%d %s", fileName, line, funcName)
		callers = append(callers, callerInfo)
	}

	return callers
}

// Trace logs SQL execution information.
// It adapts to the current log level and includes different details based on:
// - Whether there was an error
// - How long the query took (for slow query detection)
// - The configured log level
// It now includes call stack information to help identify where queries originate
func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Capture call stack information (skip GormLogger frames)
	// The skipFrames value may need adjustment based on GORM's internal call depth
	callStack := getCallerInfo(4, 3) // Skip 4 frames, capture 3 frames

	// Find caller that isn't in gorm package
	caller := "unknown"
	if len(callStack) > 0 {
		caller = callStack[0]
		// Try to find first caller outside of gorm package
		for _, frame := range callStack {
			if !strings.Contains(frame, "gorm.io/gorm") {
				caller = frame
				break
			}
		}
	}

	switch {
	case err != nil && g.level >= logger.Error:
		// Always log SQL errors with call stack
		g.log.Errorw("SQL Error",
			"error", err,
			"elapsed", elapsed,
			"sql", sql,
			"rows", rows,
			"caller", caller,
			"call_stack", callStack,
		)
	case elapsed > g.slowSQL && g.level >= logger.Warn:
		// Log slow SQL as warnings with call stack
		g.log.Warnw("Slow SQL",
			"elapsed", elapsed,
			"sql", sql,
			"rows", rows,
			"caller", caller,
			"call_stack", callStack,
		)
	case g.level >= logger.Info:
		// Standard SQL logs at Debug level with caller information
		g.log.Debugw("SQL",
			"elapsed", elapsed,
			"sql", sql,
			"rows", rows,
			"caller", caller,
		)
	}
}
