package jobqueue

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue/queue"
)

func NewDB(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_journal_mode=WAL", path))
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

func NewInMemoryDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
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
