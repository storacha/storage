package piecefinder

import (
	"context"
	"net/url"

	multihash "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/piece/piece"
)

type PieceFinder interface {
	FindPiece(ctx context.Context, digest multihash.Multihash, size uint64) (piece.PieceLink, error)
	URLForPiece(piece.PieceLink) url.URL
}
