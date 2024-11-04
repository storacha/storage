package pieceadder

import (
	"context"
	"net/url"

	multihash "github.com/multiformats/go-multihash"
)

type PieceAdder interface {
	AddPiece(ctx context.Context, digest multihash.Multihash, size uint64) (url.URL, error)
}
