package aggregator

import (
	"context"
	"database/sql"
	"fmt"

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
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

var log = logging.Logger("pdp/aggregator")

const workspaceKey = "workspace/"
const aggregatePrefix = "aggregates/"

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
	return la.pieceQueue.Enqueue(ctx, "piece_aggregate", pieceLink)
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

	linkQueue, err := jobqueue.NewMemory[datamodel.Link](
		db,
		&jobqueue.IPLDSerializerCBOR[datamodel.Link]{
			Typ:  &schema.TypeLink{},
			Opts: types.Converters,
		},
		"link",
		jobqueue.WithLog(log),
	)
	if err != nil {
		return nil, fmt.Errorf("creating link job-queue: %w", err)
	}

	pieceQueue, err := jobqueue.NewMemory[piece.PieceLink](
		db,
		&jobqueue.IPLDSerializerCBOR[piece.PieceLink]{
			Typ:  aggregate.PieceLinkType(),
			Opts: types.Converters,
		},
		"piece_link",
		jobqueue.WithLog(log),
	)
	if err != nil {
		return nil, fmt.Errorf("creating piece_link job-queue: %w", err)
	}

	// construct queues -- somewhat frstratingly these have to be constructed backward for now
	pieceAccepter := NewPieceAccepter(issuer, aggregateStore, receiptStore)
	linkQueue.Register("piece_accept", func(ctx context.Context, msg datamodel.Link) error {
		return pieceAccepter.AcceptPieces(ctx, []datamodel.Link{msg})
	})

	aggregationSubmitter := NewAggregateSubmitteer(proofSet, aggregateStore, client, linkQueue)
	linkQueue.Register("piece_submit", func(ctx context.Context, msg datamodel.Link) error {
		return aggregationSubmitter.SubmitAggregates(ctx, []datamodel.Link{msg})
	})

	pieceAggregator := NewPieceAggregator(inProgressWorkspace, aggregateStore, linkQueue)
	pieceQueue.Register("piece_aggregate", func(ctx context.Context, msg piece.PieceLink) error {
		return pieceAggregator.AggregatePieces(ctx, []piece.PieceLink{msg})
	})

	return &LocalAggregator{
		pieceQueue: pieceQueue,
		linkQueue:  linkQueue,
	}, nil
}
