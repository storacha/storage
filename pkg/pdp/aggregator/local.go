package aggregator

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-capabilities/pkg/types"
	"github.com/storacha/go-jobqueue"
	"github.com/storacha/go-piece/pkg/piece"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/internal/ipldstore"
	"github.com/storacha/storage/pkg/pdp/aggregator/aggregate"
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
	pieceAggregatorQueue    *jobqueue.JobQueue[piece.PieceLink]
	aggregateSubmitterQueue *jobqueue.JobQueue[datamodel.Link]
	pieceAccepterQueue      *jobqueue.JobQueue[datamodel.Link]
}

// Startup starts up aggregation queues
func (la *LocalAggregator) Startup() {
	la.pieceAggregatorQueue.Startup()
	la.aggregateSubmitterQueue.Startup()
	la.pieceAccepterQueue.Startup()
}

// Shutdown shuts down aggregation queues
func (la *LocalAggregator) Shutdown(ctx context.Context) {
	la.pieceAggregatorQueue.Shutdown(ctx)
	la.aggregateSubmitterQueue.Shutdown(ctx)
	la.pieceAccepterQueue.Shutdown(ctx)
}

// AggregatePiece is the frontend to aggregation
func (la *LocalAggregator) AggregatePiece(ctx context.Context, pieceLink piece.PieceLink) error {
	return la.pieceAggregatorQueue.Queue(ctx, pieceLink)
}

// NewLocal constructs an aggregator to run directly on a machine from a local datastore
func NewLocal(ds datastore.Datastore, client *curio.Client, proofSet uint64, issuer ucan.Signer, receiptStore receiptstore.ReceiptStore) *LocalAggregator {

	aggregateStore := ipldstore.IPLDStore[datamodel.Link, aggregate.Aggregate](
		store.SimpleStoreFromDatastore(namespace.Wrap(ds, datastore.NewKey(aggregatePrefix))),
		aggregate.AggregateType(), types.Converters...)
	inProgressWorkspace := NewInProgressWorkspace(store.SimpleStoreFromDatastore(namespace.Wrap(ds, datastore.NewKey(workspaceKey))))

	// construct queues -- somewhat frstratingly these have to be constructed backward for now
	pieceAccepter := NewPieceAccepter(issuer, aggregateStore, receiptStore)
	pieceAccepterQueue := jobqueue.NewJobQueue[datamodel.Link](
		jobqueue.MultiJobHandler(pieceAccepter.AcceptPieces),
		jobqueue.WithErrorHandler(handleError),
		jobqueue.WithBuffer(queueBuffer))

	aggregationSubmitter := NewAggregateSubmitteer(proofSet, aggregateStore, client, pieceAccepterQueue.Queue)
	aggregationSubmitterQueue := jobqueue.NewJobQueue[datamodel.Link](
		jobqueue.MultiJobHandler(aggregationSubmitter.SubmitAggregates),
		jobqueue.WithErrorHandler(handleError),
		jobqueue.WithBuffer(queueBuffer))

	pieceAggregator := NewPieceAggregator(inProgressWorkspace, aggregateStore, aggregationSubmitterQueue.Queue)
	pieceAggregatorQueue := jobqueue.NewJobQueue[piece.PieceLink](
		jobqueue.MultiJobHandler(pieceAggregator.AggregatePieces),
		jobqueue.WithErrorHandler(handleError),
		jobqueue.WithBuffer(queueBuffer),
	)

	return &LocalAggregator{
		pieceAggregatorQueue:    pieceAggregatorQueue,
		aggregateSubmitterQueue: aggregationSubmitterQueue,
		pieceAccepterQueue:      pieceAccepterQueue,
	}
}
