package pieceadder

import (
	"context"
	"encoding/hex"
	"net/url"

	"github.com/multiformats/go-multihash"
	"github.com/storacha/piri/pkg/pdp/curio"
)

type PieceAdder interface {
	AddPiece(ctx context.Context, digest multihash.Multihash, size uint64) (*url.URL, error)
}

// Generates URLs by interacting with Curio
type CurioAdder struct {
	client curio.PDPClient
}

var _ PieceAdder = (*CurioAdder)(nil)

func (p *CurioAdder) AddPiece(ctx context.Context, digest multihash.Multihash, size uint64) (*url.URL, error) {
	decoded, err := multihash.Decode(digest)
	if err != nil {
		return nil, err
	}
	ref, err := p.client.AddPiece(ctx, curio.AddPiece{
		Check: curio.PieceHash{
			Name: decoded.Name,
			Size: int64(size),
			Hash: hex.EncodeToString(decoded.Digest),
		},
	})
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, nil
	}
	refURL, err := url.Parse(ref.URL)
	if err != nil {
		return nil, err
	}
	return refURL, nil
}

func NewCurioAdder(client curio.PDPClient) PieceAdder {
	return &CurioAdder{client}
}
