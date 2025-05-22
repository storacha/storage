package gormdb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/storacha/storage/pkg/database"
)

func TestGORMOptions(t *testing.T) {
	tests := []struct {
		name                string
		options             []database.Option
		expectedJournalMode string
		expectedForeignKeys string
		expectedBusyTimeout string
		expectedSynchronous string
	}{
		{
			name:                "default options",
			options:             nil,
			expectedJournalMode: strings.ToLower(string(DefaultJournalMode)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: fmt.Sprintf("%d", DefaultTimeout.Milliseconds()),
			expectedSynchronous: fmt.Sprintf("%d", DefaultSyncMode),
		},
		{
			name:                "custom journal mode DELETE",
			options:             []database.Option{database.WithJournalMode(database.JournalModeDELETE)},
			expectedJournalMode: strings.ToLower(string(database.JournalModeDELETE)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: fmt.Sprintf("%d", DefaultTimeout.Milliseconds()),
			expectedSynchronous: fmt.Sprintf("%d", DefaultSyncMode),
		},
		{
			name:                "custom journal mode PERSIST",
			options:             []database.Option{database.WithJournalMode(database.JournalModePERSIST)},
			expectedJournalMode: strings.ToLower(string(database.JournalModePERSIST)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: fmt.Sprintf("%d", DefaultTimeout.Milliseconds()),
			expectedSynchronous: fmt.Sprintf("%d", DefaultSyncMode),
		},
		{
			name:                "custom journal mode OFF",
			options:             []database.Option{database.WithJournalMode(database.JournalModeOFF)},
			expectedJournalMode: strings.ToLower(string(database.JournalModeOFF)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: fmt.Sprintf("%d", DefaultTimeout.Milliseconds()),
			expectedSynchronous: fmt.Sprintf("%d", DefaultSyncMode),
		},
		{
			name:                "custom journal mode MEMORY",
			options:             []database.Option{database.WithJournalMode(database.JournalModeMEMORY)},
			expectedJournalMode: strings.ToLower(string(database.JournalModeMEMORY)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: fmt.Sprintf("%d", DefaultTimeout.Milliseconds()),
			expectedSynchronous: fmt.Sprintf("%d", DefaultSyncMode),
		},
		{
			name:                "foreign keys disabled",
			options:             []database.Option{database.WithForeignKeyConstraintsEnable(false)},
			expectedJournalMode: strings.ToLower(string(DefaultJournalMode)),
			expectedForeignKeys: "0",
			expectedBusyTimeout: fmt.Sprintf("%d", DefaultTimeout.Milliseconds()),
			expectedSynchronous: fmt.Sprintf("%d", DefaultSyncMode),
		},
		{
			name:                "custom timeout",
			options:             []database.Option{database.WithTimeout(10 * time.Second)},
			expectedJournalMode: strings.ToLower(string(DefaultJournalMode)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: "10000",
			expectedSynchronous: fmt.Sprintf("%d", DefaultSyncMode),
		},
		{
			name:                "sync mode OFF",
			options:             []database.Option{database.WithSyncMode(database.SyncModeOFF)},
			expectedJournalMode: strings.ToLower(string(DefaultJournalMode)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: fmt.Sprintf("%d", DefaultTimeout.Milliseconds()),
			expectedSynchronous: fmt.Sprintf("%d", database.SyncModeOFF),
		},
		{
			name:                "sync mode FULL",
			options:             []database.Option{database.WithSyncMode(database.SyncModeFULL)},
			expectedJournalMode: strings.ToLower(string(DefaultJournalMode)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: fmt.Sprintf("%d", DefaultTimeout.Milliseconds()),
			expectedSynchronous: fmt.Sprintf("%d", database.SyncModeFULL),
		},
		{
			name: "multiple options combined",
			options: []database.Option{
				database.WithJournalMode(database.JournalModeTRUNCATE),
				database.WithForeignKeyConstraintsEnable(false),
				database.WithTimeout(5 * time.Second),
				database.WithSyncMode(database.SyncModeEXTRA),
			},
			expectedJournalMode: strings.ToLower(string(database.JournalModeTRUNCATE)),
			expectedForeignKeys: "0",
			expectedBusyTimeout: "5000",
			expectedSynchronous: fmt.Sprintf("%d", database.SyncModeEXTRA),
		},
		{
			name: "performance optimized settings",
			options: []database.Option{
				database.WithJournalMode(database.JournalModeWAL),
				database.WithSyncMode(database.SyncModeNORMAL),
				database.WithTimeout(1 * time.Second),
			},
			expectedJournalMode: strings.ToLower(string(database.JournalModeWAL)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: "1000",
			expectedSynchronous: fmt.Sprintf("%d", database.SyncModeNORMAL),
		},
		{
			name: "safety optimized settings",
			options: []database.Option{
				database.WithJournalMode(database.JournalModeWAL),
				database.WithSyncMode(database.SyncModeFULL),
				database.WithTimeout(30 * time.Second),
			},
			expectedJournalMode: strings.ToLower(string(database.JournalModeWAL)),
			expectedForeignKeys: fmt.Sprintf("%d", bool2int(DefaultForeignKeyConstraintsEnable)),
			expectedBusyTimeout: "30000",
			expectedSynchronous: fmt.Sprintf("%d", database.SyncModeFULL),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for the database
			tempDir, err := os.MkdirTemp("", "gorm-test-*")
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, os.RemoveAll(tempDir))
			})

			dbPath := filepath.Join(tempDir, "test.db")

			// Create a new GORM database with the specified options
			db, err := New(dbPath, tt.options...)
			require.NoError(t, err)

			// Verify PRAGMAs are applied correctly
			pragmaTests := []struct {
				pragma   string
				expected string
			}{
				{"PRAGMA journal_mode", tt.expectedJournalMode},
				{"PRAGMA foreign_keys", tt.expectedForeignKeys},
				{"PRAGMA busy_timeout", tt.expectedBusyTimeout},
				{"PRAGMA synchronous", tt.expectedSynchronous},
			}

			for _, pragmaTest := range pragmaTests {
				var result string
				err := db.Raw(pragmaTest.pragma).Scan(&result).Error
				require.NoError(t, err, "failed to query %s", pragmaTest.pragma)
				assert.Equal(t, pragmaTest.expected, result, "pragma %s has wrong value", pragmaTest.pragma)
			}
		})
	}
}
