package pdp

import (
	"github.com/storacha/storage/pkg/pdp/aggregator"
	"github.com/storacha/storage/pkg/pdp/pieceadder"
	"github.com/storacha/storage/pkg/pdp/piecefinder"
)

type PDP interface {
	PieceAdder() pieceadder.PieceAdder
	PieceFinder() piecefinder.PieceFinder
	Aggregator() aggregator.Aggregator
}
