package pieceadder

import (
	"context"
	"encoding/hex"
	"net/url"

	multihash "github.com/multiformats/go-multihash"
	"github.com/storacha/storage/pkg/pdp/curio"
)

// Generates URLs by interacting with Curio
type CurioAdder struct {
	client *curio.Client
}

// SignUploadURL implements presigner.RequestPresigner.
func (p *CurioAdder) AddPiece(ctx context.Context, digest multihash.Multihash, size uint64) (url.URL, error) {
	decoded, err := multihash.Decode(digest)
	if err != nil {
		return url.URL{}, err
	}
	ref, err := p.client.AddPiece(ctx, curio.AddPiece{
		Check: curio.PieceHash{
			Name: decoded.Name,
			Size: int64(size),
			Hash: hex.EncodeToString(decoded.Digest),
		},
	})
	if err != nil {
		return url.URL{}, err
	}
	refURL, err := url.Parse(ref.URL)
	if err != nil {
		return url.URL{}, err
	}
	return *refURL, nil
}

func NewCurioAdder(client *curio.Client) PieceAdder {
	return &CurioAdder{client}
}