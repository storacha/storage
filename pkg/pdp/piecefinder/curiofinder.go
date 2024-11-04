package piecefinder

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	multihash "github.com/multiformats/go-multihash"
	"github.com/storacha/go-piece/pkg/piece"
	"github.com/storacha/storage/pkg/pdp/curio"
)

var ErrNotFound = errors.New("piece not found after maximum retries")

type CurioFinder struct {
	client *curio.Client
}

const maxAttempts = 10
const retryDelay = 5 * time.Second

// GetDownloadURL implements access.Access.
func (a *CurioFinder) FindPiece(ctx context.Context, digest multihash.Multihash, size uint64) (piece.PieceLink, error) {
	decoded, err := multihash.Decode(digest)
	if err != nil {
		return nil, err
	}

	// TODO: improve this. @magik6k says curio will have piece ready for processing
	// in seconds, but we're not sure how long that will be. We need to iterate on this
	// till we have a better solution
	attempts := 0
	for {
		result, err := a.client.FindPiece(ctx, curio.PieceHash{
			Hash: hex.EncodeToString(decoded.Digest),
			Name: decoded.Name,
			Size: int64(size),
		})
		if err == nil {
			pieceCID, err := cid.Decode(result.PieceCID)
			if err != nil {
				return nil, err
			}
			return piece.FromV1LinkAndSize(cidlink.Link{pieceCID}, size)
		}
		var errFailedResponse curio.ErrFailedResponse
		if !errors.As(err, &errFailedResponse) {
			return nil, err
		}
		if errFailedResponse.StatusCode != http.StatusNotFound {
			return nil, err
		}
		// piece not found, try again
		attempts++
		if attempts >= maxAttempts {
			return nil, ErrNotFound
		}
		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (a *CurioFinder) URLForPiece(piece piece.PieceLink) url.URL {
	return a.client.GetPieceURL(piece.V1Link().String())
}

func NewCurioFinder(client *curio.Client) PieceFinder {
	return &CurioFinder{client}
}
