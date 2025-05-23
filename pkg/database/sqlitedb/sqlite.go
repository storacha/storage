package sqlitedb

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	logging "github.com/ipfs/go-log/v2"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/storacha/storage/pkg/database"
)

var log = logging.Logger("database")

var (
	DefaultJournalMode                 = database.JournalModeWAL
	DefaultTimeout                     = 3 * time.Second
	DefaultSyncMode                    = database.SyncModeNORMAL
	DefaultForeignKeyConstraintsEnable = true
)

// New creates a new SQLite database connection with the given options
func New(path string, opts ...database.Option) (*sql.DB, error) {
	cfg := &database.Config{
		JournalMode:                 DefaultJournalMode,
		Timeout:                     DefaultTimeout,
		ForeignKeyConstraintsEnable: DefaultForeignKeyConstraintsEnable,
		SyncMode:                    DefaultSyncMode,
	}

	// Apply user-provided options (can override defaults)
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("failed to apply option %T: %w", opt, err)
		}
	}

	// Build connection string with pragmas
	var pragmas []string
	pragmas = append(pragmas, fmt.Sprintf("_pragma=journal_mode(%s)", cfg.JournalMode))
	pragmas = append(pragmas, fmt.Sprintf("_pragma=busy_timeout(%d)", cfg.Timeout.Milliseconds()))
	pragmas = append(pragmas, fmt.Sprintf("_pragma=synchronous(%s)", cfg.SyncMode))
	pragmas = append(pragmas, fmt.Sprintf("_pragma=foreign_keys(%d)", bool2int(cfg.ForeignKeyConstraintsEnable)))

	// Build connection string with pragmas - use file: URI format for pragma support
	connStr := fmt.Sprintf("file:%s", path)
	if len(pragmas) > 0 {
		connStr = fmt.Sprintf("%s?%s", connStr, strings.Join(pragmas, "&"))
	}

	log.Infof("connecting to sqlite at %s", connStr)
	// Open connection
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set reasonable defaults
	db.SetMaxOpenConns(1) // SQLite supports only one writer
	db.SetMaxIdleConns(1)

	return db, nil
}

func NewMemory(opts ...database.Option) (*sql.DB, error) {
	cfg := &database.Config{
		JournalMode:                 database.JournalModeMEMORY,
		Timeout:                     DefaultTimeout,
		ForeignKeyConstraintsEnable: DefaultForeignKeyConstraintsEnable,
		SyncMode:                    DefaultSyncMode,
	}

	// Apply user-provided options (can override defaults)
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("failed to apply option %T: %w", opt, err)
		}
	}

	if cfg.JournalMode != database.JournalModeMEMORY {
		return nil, fmt.Errorf("invalid option, for sqlite memory database JournalMode must be MEMORY")
	}

	// Build connection string with pragmas
	var pragmas []string
	pragmas = append(pragmas, fmt.Sprintf("_pragma=journal_mode(%s)", cfg.JournalMode))
	pragmas = append(pragmas, fmt.Sprintf("_pragma=busy_timeout(%d)", cfg.Timeout.Milliseconds()))
	pragmas = append(pragmas, fmt.Sprintf("_pragma=synchronous(%s)", cfg.SyncMode))
	pragmas = append(pragmas, fmt.Sprintf("_pragma=foreign_keys(%d)", bool2int(cfg.ForeignKeyConstraintsEnable)))

	connStr := "file::memory:"
	if len(pragmas) > 0 {
		connStr = fmt.Sprintf("%s?%s", connStr, strings.Join(pragmas, "&"))
	}

	log.Infof("connecting to sqlite at %s", connStr)
	// Open connection
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set reasonable defaults
	db.SetMaxOpenConns(1) // SQLite supports only one writer
	db.SetMaxIdleConns(1)

	return db, nil
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}
