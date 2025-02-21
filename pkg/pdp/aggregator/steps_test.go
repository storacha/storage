package aggregator_test

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/piece/digest"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/storage/internal/mocks"
	"github.com/storacha/storage/pkg/pdp/aggregator"
	"github.com/storacha/storage/pkg/pdp/aggregator/aggregate"
	"github.com/storacha/storage/pkg/pdp/aggregator/fns"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupPieceAggregatorDependencies(
	ctrl *gomock.Controller,
	submittedLinks *[]datamodel.Link,
) (
	*mocks.MockInProgressWorkspace,
	*mocks.MockKVStore[datamodel.Link, aggregate.Aggregate],
	aggregator.QueueSubmissionFn,
	*mocks.MockBufferedAggregator,
) {
	workspaceMock := mocks.NewMockInProgressWorkspace(ctrl)
	storeMock := mocks.NewMockKVStore[datamodel.Link, aggregate.Aggregate](ctrl)
	baMock := mocks.NewMockBufferedAggregator(ctrl)
	queueSubmissionMock := func(ctx context.Context, aggregateLink datamodel.Link) error {
		*submittedLinks = append(*submittedLinks, aggregateLink)
		return nil
	}

	return workspaceMock, storeMock, queueSubmissionMock, baMock
}

func TestPieceAggregator_StoreAndSubmit(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var submittedLinks []datamodel.Link
	workspaceMock, storeMock, queueSubMock, baMock := setupPieceAggregatorDependencies(ctrl, &submittedLinks)

	pa := aggregator.NewPieceAggregator(workspaceMock, storeMock, queueSubMock, aggregator.WithAggregator(baMock))
	// the below makes assertion that when three aggregates are returned by the aggregator of the piece-aggregator
	// three writes are made to the aggregate-store and three submissions are made too the queue-submission function.
	p1 := createPiece(t, MB)
	p2 := createPiece(t, MB)
	p3 := createPiece(t, MB)
	expectedPieces := []piece.PieceLink{p1, p2, p3}
	expectedAggregates := []aggregate.Aggregate{
		{
			Root: p1,
		},
		{
			Root: p2,
		},
		{
			Root: p3,
		},
	}
	expectedBuffer := fns.Buffer{}

	workspaceMock.EXPECT().GetBuffer(ctx).Return(expectedBuffer, nil)
	baMock.EXPECT().AggregatePieces(expectedBuffer, expectedPieces).Return(expectedBuffer, expectedAggregates, nil)
	workspaceMock.EXPECT().PutBuffer(ctx, expectedBuffer).Return(nil)
	storeMock.EXPECT().Put(ctx, gomock.Any(), gomock.Any()).Return(nil)
	storeMock.EXPECT().Put(ctx, gomock.Any(), gomock.Any()).Return(nil)
	storeMock.EXPECT().Put(ctx, gomock.Any(), gomock.Any()).Return(nil)

	err := pa.AggregatePieces(ctx, expectedPieces)
	require.NoError(t, err)

	require.Len(t, submittedLinks, len(expectedAggregates))
}

func TestPieceAggregator_GetBufferError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var submittedLinks []datamodel.Link
	workspaceMock, storeMock, queueSubMock, baMock := setupPieceAggregatorDependencies(ctrl, &submittedLinks)

	pa := aggregator.NewPieceAggregator(workspaceMock, storeMock, queueSubMock, aggregator.WithAggregator(baMock))
	p1 := createPiece(t, MB)
	p2 := createPiece(t, MB)
	p3 := createPiece(t, MB)
	expectedPieces := []piece.PieceLink{p1, p2, p3}
	expectedBuffer := fns.Buffer{}

	workspaceMock.EXPECT().GetBuffer(ctx).Return(expectedBuffer, fmt.Errorf("get buffer error"))

	err := pa.AggregatePieces(ctx, expectedPieces)
	require.Error(t, err)

	require.Len(t, submittedLinks, 0)
}

func TestPieceAggregator_PutBufferError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var submittedLinks []datamodel.Link
	workspaceMock, storeMock, queueSubMock, baMock := setupPieceAggregatorDependencies(ctrl, &submittedLinks)

	pa := aggregator.NewPieceAggregator(workspaceMock, storeMock, queueSubMock, aggregator.WithAggregator(baMock))
	// the below makes assertion that when three aggregates are returned by the aggregator of the piece-aggregator
	// three writes are made to the aggregate-store and three submissions are made too the queue-submission function.
	p1 := createPiece(t, MB)
	p2 := createPiece(t, MB)
	p3 := createPiece(t, MB)
	expectedPieces := []piece.PieceLink{p1, p2, p3}
	expectedAggregates := []aggregate.Aggregate{
		{
			Root: p1,
		},
		{
			Root: p2,
		},
		{
			Root: p3,
		},
	}
	expectedBuffer := fns.Buffer{}

	workspaceMock.EXPECT().GetBuffer(ctx).Return(expectedBuffer, nil)
	baMock.EXPECT().AggregatePieces(expectedBuffer, expectedPieces).Return(expectedBuffer, expectedAggregates, nil)
	workspaceMock.EXPECT().PutBuffer(ctx, expectedBuffer).Return(fmt.Errorf("put buffer error"))

	err := pa.AggregatePieces(ctx, expectedPieces)
	require.Error(t, err)

	require.Len(t, submittedLinks, 0)
}

func TestPieceAggregator_AggregatePieceError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var submittedLinks []datamodel.Link
	workspaceMock, storeMock, queueSubMock, baMock := setupPieceAggregatorDependencies(ctrl, &submittedLinks)

	pa := aggregator.NewPieceAggregator(workspaceMock, storeMock, queueSubMock, aggregator.WithAggregator(baMock))
	// the below makes assertion that when three aggregates are returned by the aggregator of the piece-aggregator
	// three writes are made to the aggregate-store and three submissions are made too the queue-submission function.
	p1 := createPiece(t, MB)
	p2 := createPiece(t, MB)
	p3 := createPiece(t, MB)
	expectedPieces := []piece.PieceLink{p1, p2, p3}
	expectedBuffer := fns.Buffer{}

	workspaceMock.EXPECT().GetBuffer(ctx).Return(expectedBuffer, nil)
	baMock.EXPECT().AggregatePieces(expectedBuffer, expectedPieces).Return(expectedBuffer, nil, fmt.Errorf("aggregate piece error"))

	err := pa.AggregatePieces(ctx, expectedPieces)
	require.Error(t, err)

	require.Len(t, submittedLinks, 0)
}

const (
	MB = 1 << 20
)

// TODO(forrest): move this to a shared location with nfs pkg.

// createPiece is a helper that produces a piece with the given unpadded size,
// using random data so we don't rely on any pre-computed fixtures.
func createPiece(t *testing.T, unpaddedSize int64) piece.PieceLink {
	t.Helper()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	dataReader := io.LimitReader(r, unpaddedSize)

	calc := &commp.Calc{}
	n, err := io.Copy(calc, dataReader)
	require.NoError(t, err, "failed copying data into commp.Calc")
	require.Equal(t, unpaddedSize, n)

	commP, paddedSize, err := calc.Digest()
	require.NoError(t, err, "failed to compute commP")

	pieceDigest, err := digest.FromCommitmentAndSize(commP, uint64(unpaddedSize))
	require.NoError(t, err, "failed building piece digest from commP")

	p := piece.FromPieceDigest(pieceDigest)
	// Ensure our piece’s PaddedSize matches the commp library’s reported paddedSize.
	require.Equal(t, paddedSize, p.PaddedSize())

	t.Logf("Created test piece: %s from unpadded size: %d", pieceLinkString(p), unpaddedSize)
	return p
}

// pieceLinkString is a helper to display piece metadata in logs.
func pieceLinkString(p piece.PieceLink) string {
	return fmt.Sprintf("Piece: %s, Height: %d, Padding: %d, PaddedSize: %d",
		p.Link(), p.Height(), p.Padding(), p.PaddedSize())
}
