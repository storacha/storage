package aggregator

import (
	"context"

	"github.com/storacha/go-piece/pkg/piece"
)

// Aggregator is an interface for accessing an aggregation queue
type Aggregator interface {
	AggregatePiece(ctx context.Context, pieceLink piece.PieceLink) error
}
