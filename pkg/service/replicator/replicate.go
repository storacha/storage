package replicator

import (
	"context"
	"database/sql"
	"runtime"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/client"
	"github.com/storacha/go-ucanto/principal"

	"github.com/storacha/storage/lib/jobqueue"
	"github.com/storacha/storage/lib/jobqueue/serializer"
	"github.com/storacha/storage/pkg/pdp"
	"github.com/storacha/storage/pkg/service/blobs"
	"github.com/storacha/storage/pkg/service/claims"
	replicahandler "github.com/storacha/storage/pkg/service/storage/handlers/replica"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

const (
	QueueName         = "replicator"
	TransferQueueTask = "transfer"
)

var log = logging.Logger("replicator")

type Replicator interface {
	Replicate(context.Context, *replicahandler.TransferRequest) error
}

type Service struct {
	queue *jobqueue.JobQueue[*replicahandler.TransferRequest]
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
	db *sql.DB,
) (*Service, error) {

	replicationQueue, err := jobqueue.New(
		QueueName,
		db,
		&serializer.JSON[*replicahandler.TransferRequest]{},
		jobqueue.WithLogger(log),
		jobqueue.WithMaxRetries(10),
		jobqueue.WithMaxWorkers(uint(runtime.NumCPU())),
	)
	if err != nil {
		return nil, err
	}
	if err := replicationQueue.Register(TransferQueueTask, func(ctx context.Context, request *replicahandler.TransferRequest) error {
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
	return r.queue.Enqueue(ctx, TransferQueueTask, task)
}

func (r *Service) Start(ctx context.Context) error {
	r.queue.Start(ctx)
	return nil
}

func (r *Service) Stop(_ context.Context) error {
	// TODO consider closing the database connection on *sql.DB
	return nil
}
