package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("database")

// SQLiteOptions holds all configuration options for SQLite connection
type SQLiteOptions struct {
	path   string
	params url.Values
}

// Option is a function that configures SQLiteOptions
type Option func(*SQLiteOptions)

// NewSQLite creates a new SQLite database connection with the given options
func NewSQLite(path string, opts ...Option) (*sql.DB, error) {
	options := &SQLiteOptions{
		path:   path,
		params: url.Values{},
	}

	// Apply user-provided options
	for _, opt := range opts {
		opt(options)
	}

	// Build connection string
	connStr := options.path
	if len(options.params) > 0 {
		connStr = fmt.Sprintf("%s?%s", options.path, options.params.Encode())
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

// WithJournalMode sets the journal mode for the database
// Possible values:
// - "DELETE": Default - delete journal after commit
// - "TRUNCATE": Truncate journal instead of deleting (faster on some filesystems)
// - "PERSIST": Zero journal file instead of deleting (faster but uses more disk)
// - "MEMORY": Store journal in memory (fast but no crash recovery)
// - "WAL": Write-Ahead Logging - allows concurrent reads and writes
// - "OFF": No journaling (dangerous - risks corruption)
func WithJournalMode(mode string) Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_journal", mode)
	}
}

// WithBusyTimeout sets how long SQLite will wait if the database is locked
// before returning a busy error. Higher values are better for concurrent access
// but may delay error reporting.
func WithBusyTimeout(duration time.Duration) Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_timeout", fmt.Sprintf("%d", duration.Milliseconds()))
	}
}

// WithForeignKeys enables or disables foreign key constraint enforcement.
// Should typically be enabled to ensure data consistency.
func WithForeignKeys(enable bool) Option {
	return func(o *SQLiteOptions) {
		if enable {
			o.params.Set("_fk", "true")
		} else {
			o.params.Set("_fk", "false")
		}
	}
}

// WithSharedCache enables shared cache mode, allowing multiple connections
// from the same process to share the SQLite page cache. Improves performance
// when multiple goroutines access the same database.
func WithSharedCache() Option {
	return func(o *SQLiteOptions) {
		o.params.Set("cache", "shared")
	}
}

// WithPrivateCache disables shared cache mode, giving each connection
// its own private cache. Better isolation but potentially higher memory usage.
func WithPrivateCache() Option {
	return func(o *SQLiteOptions) {
		o.params.Set("cache", "private")
	}
}

// WithSynchronous sets how aggressively SQLite writes to disk.
// Possible values:
// - "OFF": No syncs, fastest but dangerous - can corrupt database on power loss
// - "NORMAL": Sync at critical moments, good balance of safety/performance
// - "FULL": Full durability, syncs at critical points
// - "EXTRA": Maximum durability, additional syncs
func WithSynchronous(level string) Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_synchronous", level)
	}
}

// WithReadOnly opens the database in read-only mode.
// Useful for accessing databases on read-only media or preventing modifications.
func WithReadOnly() Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_mode", "ro")
	}
}

// WithReadWrite opens the database in read-write mode but doesn't create
// the database if it doesn't exist.
func WithReadWrite() Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_mode", "rw")
	}
}

// WithReadWriteCreate opens the database in read-write mode and creates
// the database if it doesn't exist (default behavior).
func WithReadWriteCreate() Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_mode", "rwc")
	}
}

// WithImmutable marks database as immutable, enabling additional optimizations.
// Should only be used when you're certain the database file won't be modified
// externally during the connection lifetime.
func WithImmutable() Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_immutable", "true")
	}
}

// WithAutoVacuum sets the auto-vacuum mode:
// - "NONE": No auto-vacuum (default)
// - "FULL": Vacuum on every transaction commit (slow but keeps file compact)
// - "INCREMENTAL": Vacuum gradually over time (better performance)
func WithAutoVacuum(mode string) Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_auto_vacuum", strings.ToLower(mode))
	}
}

// WithSecureDelete causes SQLite to overwrite deleted content with zeros.
// Prevents forensic recovery of deleted data but reduces performance.
func WithSecureDelete(enable bool) Option {
	return func(o *SQLiteOptions) {
		if enable {
			o.params.Set("_secure_delete", "true")
		} else {
			o.params.Set("_secure_delete", "false")
		}
	}
}

// WithCaseSensitiveLike makes the LIKE operator case sensitive.
// By default, LIKE in SQLite is case insensitive for ASCII characters.
func WithCaseSensitiveLike(enable bool) Option {
	return func(o *SQLiteOptions) {
		if enable {
			o.params.Set("_case_sensitive_like", "true")
		} else {
			o.params.Set("_case_sensitive_like", "false")
		}
	}
}

// WithTempStore configures where temporary tables and indices are stored:
// - "DEFAULT": Uses the compiler-defined default
// - "FILE": Store temporary data in files
// - "MEMORY": Store temporary data in memory (faster but uses more RAM)
func WithTempStore(location string) Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_temp_store", location)
	}
}

// WithRecursiveTriggers enables or disables recursive triggers.
// When enabled, triggers can fire other triggers recursively.
func WithRecursiveTriggers(enable bool) Option {
	return func(o *SQLiteOptions) {
		if enable {
			o.params.Set("_recursive_triggers", "true")
		} else {
			o.params.Set("_recursive_triggers", "false")
		}
	}
}

// WithDeferForeignKeys defers foreign key constraint checking until
// transaction commit. Improves performance for large transactions.
func WithDeferForeignKeys(enable bool) Option {
	return func(o *SQLiteOptions) {
		if enable {
			o.params.Set("_defer_foreign_keys", "true")
		} else {
			o.params.Set("_defer_foreign_keys", "false")
		}
	}
}

// WithIgnoreCheckConstraints enables or disables CHECK constraints.
// Should typically remain enabled to ensure data integrity.
func WithIgnoreCheckConstraints(ignore bool) Option {
	return func(o *SQLiteOptions) {
		if ignore {
			o.params.Set("_ignore_check_constraints", "true")
		} else {
			o.params.Set("_ignore_check_constraints", "false")
		}
	}
}

// WithQueryOnly disables all database modifications.
// Provides additional safety for read-only operations.
func WithQueryOnly(enable bool) Option {
	return func(o *SQLiteOptions) {
		if enable {
			o.params.Set("_query_only", "true")
		} else {
			o.params.Set("_query_only", "false")
		}
	}
}

// WithTransactionLock sets the transaction locking mode:
// - "DEFERRED": Acquire locks only when needed (default)
// - "IMMEDIATE": Acquire a reserved lock immediately
// - "EXCLUSIVE": Acquire an exclusive lock immediately
func WithTransactionLock(mode string) Option {
	return func(o *SQLiteOptions) {
		o.params.Set("_txlock", strings.ToLower(mode))
	}
}
