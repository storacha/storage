package aggregator

import (
	"context"

	"github.com/storacha/go-libstoracha/piece/piece"
)

// Aggregator is an interface for accessing an aggregation queue
type Aggregator interface {
	AggregatePiece(ctx context.Context, pieceLink piece.PieceLink) error
}
