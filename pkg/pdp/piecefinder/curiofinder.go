package piecefinder

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/piece/piece"

	"github.com/storacha/piri/pkg/pdp/curio"
	"github.com/storacha/piri/pkg/store"
)

type PieceFinder interface {
	FindPiece(ctx context.Context, digest multihash.Multihash, size uint64) (piece.PieceLink, error)
	URLForPiece(piece.PieceLink) url.URL
}

var _ PieceFinder = (*CurioFinder)(nil)

type CurioFinder struct {
	client      curio.PDPClient
	maxAttempts int
	retryDelay  time.Duration
}

type Option func(cf *CurioFinder)

func WithRetryDelay(d time.Duration) Option {
	return func(cf *CurioFinder) {
		cf.retryDelay = d
	}
}

func WithMaxAttempts(n int) Option {
	return func(cf *CurioFinder) {
		cf.maxAttempts = n
	}
}

const defaultMaxAttempts = 10
const defaultRetryDelay = 5 * time.Second

func NewCurioFinder(client curio.PDPClient, opts ...Option) PieceFinder {
	cf := &CurioFinder{
		client:      client,
		maxAttempts: defaultMaxAttempts,
		retryDelay:  defaultRetryDelay,
	}

	for _, opt := range opts {
		opt(cf)
	}
	return cf
}

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
			return piece.FromV1LinkAndSize(cidlink.Link{Cid: pieceCID}, size)
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
		if attempts >= a.maxAttempts {
			return nil, fmt.Errorf("maximum retries exceeded: %w", store.ErrNotFound)
		}
		timer := time.NewTimer(a.retryDelay)
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
