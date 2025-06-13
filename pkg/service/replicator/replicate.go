package replicator

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/client"
	"github.com/storacha/go-ucanto/principal"

	"github.com/storacha/piri/pkg/database"
	"github.com/storacha/piri/pkg/database/sqlitedb"
	"github.com/storacha/piri/pkg/pdp"
	"github.com/storacha/piri/pkg/pdp/aggregator/jobqueue"
	"github.com/storacha/piri/pkg/pdp/aggregator/jobqueue/serializer"
	"github.com/storacha/piri/pkg/service/blobs"
	"github.com/storacha/piri/pkg/service/claims"
	replicahandler "github.com/storacha/piri/pkg/service/storage/handlers/replica"
	"github.com/storacha/piri/pkg/store/receiptstore"
)

var log = logging.Logger("replicator")

type Replicator interface {
	Replicate(context.Context, *replicahandler.TransferRequest) error
}

type Service struct {
	queue  *jobqueue.JobQueue[*replicahandler.TransferRequest]
	cancel context.CancelFunc
}

type adapter struct {
	id         principal.Signer
	pdp        pdp.PDP
	blobs      blobs.Blobs
	claims     claims.Claims
	receipts   receiptstore.ReceiptStore
	uploadConn client.Connection
}

func (a adapter) ID() principal.Signer                { return a.id }
func (a adapter) PDP() pdp.PDP                        { return a.pdp }
func (a adapter) Blobs() blobs.Blobs                  { return a.blobs }
func (a adapter) Claims() claims.Claims               { return a.claims }
func (a adapter) Receipts() receiptstore.ReceiptStore { return a.receipts }
func (a adapter) UploadConnection() client.Connection { return a.uploadConn }

func New(
	id principal.Signer,
	p pdp.PDP,
	b blobs.Blobs,
	c claims.Claims,
	rstore receiptstore.ReceiptStore,
	uploadConn client.Connection,
	stateDir string,
) (*Service, error) {

	db, err := sqlitedb.New(filepath.Join(stateDir, "replicator.db"),
		database.WithJournalMode("WAL"),
		database.WithTimeout(5*time.Second),
		database.WithSyncMode(database.SyncModeNORMAL),
	)
	if err != nil {
		return nil, fmt.Errorf("creating jobqueue database: %w", err)
	}
	replicationQueue, err := jobqueue.New[*replicahandler.TransferRequest](
		"replication",
		db,
		&serializer.JSON[*replicahandler.TransferRequest]{},
		jobqueue.WithLogger(log),
		jobqueue.WithMaxRetries(10),
		jobqueue.WithMaxWorkers(uint(runtime.NumCPU())),
	)
	if err != nil {
		return nil, err
	}
	if err := replicationQueue.Register("transfer-task", func(ctx context.Context, request *replicahandler.TransferRequest) error {
		return replicahandler.Transfer(ctx,
			&adapter{
				id:         id,
				pdp:        p,
				blobs:      b,
				claims:     c,
				receipts:   rstore,
				uploadConn: uploadConn,
			},
			request)
	}); err != nil {
		return nil, err
	}

	return &Service{queue: replicationQueue}, nil
}

func (r *Service) Replicate(ctx context.Context, task *replicahandler.TransferRequest) error {
	return r.queue.Enqueue(ctx, "transfer-task", task)
}

func (r *Service) Start(ctx context.Context) error {
	// Create a cancelable context for the queue that lives beyond fx startup
	queueCtx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	go r.queue.Start(queueCtx)
	return nil
}

func (r *Service) Stop(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
