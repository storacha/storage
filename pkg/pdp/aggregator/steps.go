package aggregator

import (
	"context"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-capabilities/pkg/types"
	"github.com/storacha/go-piece/pkg/piece"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/internal/ipldstore"
	"github.com/storacha/storage/pkg/pdp/aggregator/aggregate"
	"github.com/storacha/storage/pkg/pdp/aggregator/fns"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

type QueuePieceAggregationFn func(context.Context, piece.PieceLink) error

// Step 1: Generate aggregates from pieces

type InProgressWorkspace interface {
	GetBuffer(context.Context) (fns.Buffer, error)
	PutBuffer(context.Context, fns.Buffer) error
}

type bufferKey struct{}

func (bufferKey) String() string { return "buffer" }

type inProgressWorkSpace struct {
	store ipldstore.KVStore[bufferKey, fns.Buffer]
}

func (i *inProgressWorkSpace) GetBuffer(ctx context.Context) (fns.Buffer, error) {
	buf, err := i.store.Get(ctx, bufferKey{})
	if store.IsNotFound(err) {
		err := i.store.Put(ctx, bufferKey{}, fns.Buffer{})
		return fns.Buffer{}, err
	}
	return buf, err
}

func (i *inProgressWorkSpace) PutBuffer(ctx context.Context, buffer fns.Buffer) error {
	return i.store.Put(ctx, bufferKey{}, buffer)
}

func NewInProgressWorkspace(store store.Store) InProgressWorkspace {
	return &inProgressWorkSpace{
		ipldstore.IPLDStore[bufferKey, fns.Buffer](store, fns.BufferType(), types.Converters...),
	}
}

type AggregateStore ipldstore.KVStore[datamodel.Link, aggregate.Aggregate]
type QueueSubmissionFn func(ctx context.Context, aggregateLink datamodel.Link) error

type PieceAggregator struct {
	workspace       InProgressWorkspace
	store           AggregateStore
	queueSubmission QueueSubmissionFn
}

func NewPieceAggregator(workspace InProgressWorkspace, store AggregateStore, queueSubmission QueueSubmissionFn) *PieceAggregator {
	return &PieceAggregator{
		workspace:       workspace,
		store:           store,
		queueSubmission: queueSubmission,
	}
}

func (pa *PieceAggregator) AggregatePieces(ctx context.Context, pieces []piece.PieceLink) error {
	buffer, err := pa.workspace.GetBuffer(ctx)
	if err != nil {
		return fmt.Errorf("reading in progress pieces from work space: %w", err)
	}
	buffer, aggregates, err := fns.AggregatePieces(buffer, pieces)
	if err != nil {
		return fmt.Errorf("calculating aggegates: %w", err)
	}
	if err := pa.workspace.PutBuffer(ctx, buffer); err != nil {
		return fmt.Errorf("updating work space: %w", err)
	}
	for _, aggregate := range aggregates {
		err := pa.store.Put(ctx, aggregate.Root.Link(), aggregate)
		if err != nil {
			return fmt.Errorf("storing aggregate: %w", err)
		}
		if err := pa.queueSubmission(ctx, aggregate.Root.Link()); err != nil {
			return fmt.Errorf("queueing aggregates for submission: %w", err)
		}
	}
	return nil
}

// Step 2: Submit to curio

type QueuePieceAcceptFn func(ctx context.Context, aggregateLink datamodel.Link) error

type AggregateSubmitter struct {
	proofSet         uint64
	store            AggregateStore
	client           *curio.Client
	queuePieceAccept QueuePieceAcceptFn
}

func NewAggregateSubmitteer(proofSet uint64, store AggregateStore, client *curio.Client, queuePieceAccept QueuePieceAcceptFn) *AggregateSubmitter {
	return &AggregateSubmitter{
		store:            store,
		client:           client,
		queuePieceAccept: queuePieceAccept,
	}
}

func (as *AggregateSubmitter) SubmitAggregates(ctx context.Context, aggregateLinks []datamodel.Link) error {
	aggregates := make([]aggregate.Aggregate, 0, len(aggregateLinks))
	for _, aggregateLink := range aggregateLinks {
		aggregate, err := as.store.Get(ctx, aggregateLink)
		if err != nil {
			return fmt.Errorf("reading aggregates: %w", err)
		}
		aggregates = append(aggregates, aggregate)
	}
	if err := fns.SubmitAggregates(ctx, as.client, as.proofSet, aggregates); err != nil {
		return fmt.Errorf("submitting aggregates to Curio: %w", err)
	}
	for _, aggregateLink := range aggregateLinks {
		err := as.queuePieceAccept(ctx, aggregateLink)
		if err != nil {
			return fmt.Errorf("queuing piece acceptance: %w", err)
		}
	}
	return nil
}

// Step 3: generate receipts for piece accept

type PieceAccepter struct {
	issuer         ucan.Signer
	aggregateStore AggregateStore
	receiptStore   receiptstore.ReceiptStore
}

func NewPieceAccepter(issuer ucan.Signer, aggregateStore AggregateStore, receiptStore receiptstore.ReceiptStore) *PieceAccepter {
	return &PieceAccepter{
		issuer:         issuer,
		aggregateStore: aggregateStore,
		receiptStore:   receiptStore,
	}
}

func (pa *PieceAccepter) AcceptPieces(ctx context.Context, aggregateLinks []datamodel.Link) error {
	aggregates := make([]aggregate.Aggregate, 0, len(aggregateLinks))
	for _, aggregateLink := range aggregateLinks {
		aggregate, err := pa.aggregateStore.Get(ctx, aggregateLink)
		if err != nil {
			return fmt.Errorf("reading aggregates: %w", err)
		}
		aggregates = append(aggregates, aggregate)
	}
	// TODO: Should we actually send a piece accept invocation? It seems unneccesary it's all the same machine
	receipts, err := fns.GenerateReceiptsForAggregates(pa.issuer, aggregates)
	if err != nil {
		return fmt.Errorf("generating receipts: %w", err)
	}
	for _, receipt := range receipts {
		if err := pa.receiptStore.Put(ctx, receipt); err != nil {
			return err
		}
	}
	return nil
}
