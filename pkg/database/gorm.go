package database

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/storacha/storage/pkg/pdp/service/models"
)

func NewGORMDb(d gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(d, &gorm.Config{
		// No need to run every operation in a transaction, we are explicit about where transactions are required.
		SkipDefaultTransaction: true,
		Logger:                 newGormLogger(log),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}

	// Execute pragmas directly using raw SQL
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 500000",
		"PRAGMA cache_size = -64000", // Approximately 64MB (negative means KB)
	}

	log.Infof("setting SQLite PRAGMAs...")
	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			return nil, fmt.Errorf("failed to execute pragma %s: %w", pragma, err)
		}
		log.Debugf("executed: %s", pragma)
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

// gormLogger adapts the project's logging system to GORM's logging interface.
// It ensures consistent logging across the application regardless of whether
// the log is coming from GORM or the application code.
type gormLogger struct {
	log     *logging.ZapEventLogger
	level   logger.LogLevel
	slowSQL time.Duration // threshold for slow SQL logging
}

// newGormLogger creates a new gormLogger with appropriate defaults.
func newGormLogger(log *logging.ZapEventLogger) *gormLogger {
	return &gormLogger{
		log:     log,
		level:   logger.Info, // Default to Info level
		slowSQL: time.Second, // Default threshold for slow SQL
	}
}

// LogMode sets the log level for GORM and returns an updated logger.
// This allows dynamic configuration of logging level.
func (g *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *g
	newLogger.level = level
	return &newLogger
}

// Info logs info messages using the application's logger.
func (g *gormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	if g.level >= logger.Info {
		g.log.Infof(s, i...)
	}
}

// Warn logs warning messages using the application's logger.
func (g *gormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	if g.level >= logger.Warn {
		g.log.Warnf(s, i...)
	}
}

// Error logs error messages using the application's logger.
func (g *gormLogger) Error(ctx context.Context, s string, i ...interface{}) {
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
func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Capture call stack information (skip gormLogger frames)
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
