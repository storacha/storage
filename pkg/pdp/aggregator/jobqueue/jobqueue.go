package jobqueue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue/serializer"
	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue/worker"

	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue/queue"
)

type Service[T any] interface {
	Start(ctx context.Context)
	Register(name string, fn func(context.Context, T) error) error
	Enqueue(ctx context.Context, name string, msg T) error
}

type Config struct {
	Logger     worker.StandardLogger
	MaxWorkers uint
	MaxRetries uint
	MaxTimeout time.Duration
}
type Option func(c *Config) error

func WithLogger(l worker.StandardLogger) Option {
	return func(c *Config) error {
		if l == nil {
			return errors.New("job queue logger cannot be nil")
		}
		c.Logger = l
		return nil
	}
}

func WithMaxWorkers(maxWorkers uint) Option {
	return func(c *Config) error {
		if maxWorkers < 1 {
			return errors.New("job queue max workers must be greater than zero")
		}
		c.MaxWorkers = maxWorkers
		return nil
	}
}

func WithMaxRetries(maxRetries uint) Option {
	return func(c *Config) error {
		c.MaxRetries = maxRetries
		return nil
	}
}

func WithMaxTimeout(maxTimeout time.Duration) Option {
	return func(c *Config) error {
		if maxTimeout == 0 {
			return errors.New("max timeout cannot be 0")
		}
		c.MaxTimeout = maxTimeout
		return nil
	}
}

type JobQueue[T any] struct {
	worker *worker.Worker[T]
	queue  *queue.Queue
}

func New[T any](name string, db *sql.DB, ser serializer.Serializer[T], opts ...Option) (*JobQueue[T], error) {
	// set defaults
	c := &Config{
		Logger:     &worker.DiscardLogger{},
		MaxWorkers: 1,
		MaxRetries: 3,
		MaxTimeout: 5 * time.Second,
	}
	// apply overrides of defaults
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// instantiate queue schema in the database, this should be fairly quick
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if err := queue.Setup(ctx, db); err != nil {
		return nil, err
	}

	// instantiate queue
	q, err := queue.New(queue.NewOpts{
		DB:         db,
		MaxReceive: int(c.MaxRetries),
		Name:       name,
		Timeout:    c.MaxTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	// instantiate worker which consumes from queue
	w := worker.New[T](q, ser, worker.WithLog(c.Logger), worker.WithLimit(int(c.MaxWorkers)))

	return &JobQueue[T]{
		queue:  q,
		worker: w,
	}, nil
}

func (j *JobQueue[T]) Start(ctx context.Context) {
	j.worker.Start(ctx)
}

func (j *JobQueue[T]) Register(name string, fn func(context.Context, T) error) error {
	return j.worker.Register(name, fn)
}

func (j *JobQueue[T]) Enqueue(ctx context.Context, name string, msg T) error {
	return j.worker.Enqueue(ctx, name, msg)
}

func NewInMemoryDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("failed to open memory database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return db, nil
}
