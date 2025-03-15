package jobqueue

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue/queue"
)

func NewMemoryDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file::memory")
	if err != nil {
		return nil, fmt.Errorf("failed to open memory database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// isntall the schema
	if err := queue.Setup(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to setup memory database: %w", err)
	}

	return db, nil
}

func NewMemory[T any](db *sql.DB, ser Serializer[T], name string, option ...Option) (*JobQueue[T], error) {
	// Make a new queue for the jobs. You can have as many of these as you like, just name them differently.
	q := queue.New(queue.NewOpts{
		DB:   db,
		Name: name,
	})

	return New[T](q, ser, option...), nil
}
