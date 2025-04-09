package aggregator

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/ucan"

	"github.com/storacha/storage/internal/ipldstore"
	"github.com/storacha/storage/pkg/pdp/aggregator/aggregate"
	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue"
	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue/serializer"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

var log = logging.Logger("pdp/aggregator")

const workspaceKey = "workspace/"
const aggregatePrefix = "aggregates/"

const (
	LinkQueueName  = "link"
	PieceQueueName = "piece"
)

// task names
const (
	PieceAggregateTask = "piece_aggregate"
	PieceSubmitTask    = "piece_submit"
	PieceAcceptTask    = "piece_accept"
)

const queueBuffer = 16

func handleError(err error) {
	log.Errorf(err.Error())
}

// LocalAggregator is a local aggregator running directly on the storage node
// when run w/o cloud infra
type LocalAggregator struct {
	pieceQueue *jobqueue.JobQueue[piece.PieceLink]
	linkQueue  *jobqueue.JobQueue[datamodel.Link]
}

// Startup starts up aggregation queues
func (la *LocalAggregator) Startup(ctx context.Context) error {
	go la.pieceQueue.Start(ctx)
	go la.linkQueue.Start(ctx)
	return nil
}

// Shutdown shuts down aggregation queues
func (la *LocalAggregator) Shutdown(ctx context.Context) {
}

// AggregatePiece is the frontend to aggregation
func (la *LocalAggregator) AggregatePiece(ctx context.Context, pieceLink piece.PieceLink) error {
	log.Infow("Aggregating piece", "piece", pieceLink.Link().String())
	return la.pieceQueue.Enqueue(ctx, PieceAggregateTask, pieceLink)
}

// NewLocal constructs an aggregator to run directly on a machine from a local datastore
func NewLocal(
	ds datastore.Datastore,
	db *sql.DB,
	client *curio.Client,
	proofSet uint64,
	issuer ucan.Signer,
	receiptStore receiptstore.ReceiptStore,
) (*LocalAggregator, error) {

	aggregateStore := ipldstore.IPLDStore[datamodel.Link, aggregate.Aggregate](
		store.SimpleStoreFromDatastore(namespace.Wrap(ds, datastore.NewKey(aggregatePrefix))),
		aggregate.AggregateType(), types.Converters...)
	inProgressWorkspace := NewInProgressWorkspace(store.SimpleStoreFromDatastore(namespace.Wrap(ds, datastore.NewKey(workspaceKey))))

	linkQueue, err := jobqueue.New(
		LinkQueueName,
		db,
		&serializer.IPLDSerializerCBOR[datamodel.Link]{
			Typ:  &schema.TypeLink{},
			Opts: types.Converters,
		},
		jobqueue.WithLogger(logging.Logger("jobqueue").With("queue", LinkQueueName)),
		jobqueue.WithMaxRetries(10),
		jobqueue.WithMaxWorkers(uint(runtime.NumCPU())),
	)
	if err != nil {
		return nil, fmt.Errorf("creating link job-queue: %w", err)
	}

	pieceQueue, err := jobqueue.New(
		PieceQueueName,
		db,
		&serializer.IPLDSerializerCBOR[piece.PieceLink]{
			Typ:  aggregate.PieceLinkType(),
			Opts: types.Converters,
		},
		jobqueue.WithLogger(logging.Logger("jobqueue").With("queue", PieceQueueName)),
		jobqueue.WithMaxRetries(10),
		jobqueue.WithMaxWorkers(uint(runtime.NumCPU())),
	)
	if err != nil {
		return nil, fmt.Errorf("creating piece_link job-queue: %w", err)
	}

	// construct queues -- somewhat frstratingly these have to be constructed backward for now
	pieceAccepter := NewPieceAccepter(issuer, aggregateStore, receiptStore)
	aggregationSubmitter := NewAggregateSubmitteer(proofSet, aggregateStore, client, linkQueue)
	pieceAggregator := NewPieceAggregator(inProgressWorkspace, aggregateStore, linkQueue)

	if err := linkQueue.Register(PieceAcceptTask, func(ctx context.Context, msg datamodel.Link) error {
		return pieceAccepter.AcceptPieces(ctx, []datamodel.Link{msg})
	}); err != nil {
		return nil, fmt.Errorf("registering %s task: %w", PieceAcceptTask, err)
	}

	if err := linkQueue.Register(PieceSubmitTask, func(ctx context.Context, msg datamodel.Link) error {
		return aggregationSubmitter.SubmitAggregates(ctx, []datamodel.Link{msg})
	}); err != nil {
		return nil, fmt.Errorf("registering %s task: %w", PieceSubmitTask, err)
	}

	if err := pieceQueue.Register(PieceAggregateTask, func(ctx context.Context, msg piece.PieceLink) error {
		return pieceAggregator.AggregatePieces(ctx, []piece.PieceLink{msg})
	}); err != nil {
		return nil, fmt.Errorf("registering %s task: %w", PieceAggregateTask, err)
	}

	return &LocalAggregator{
		pieceQueue: pieceQueue,
		linkQueue:  linkQueue,
	}, nil
}
