package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"

	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/store"
)

type BlobAcceptRequest struct {
	Space did.DID
	Blob  blob.Blob
	Put   blob.Promise
}

type BlobAcceptResponse struct {
	Claim delegation.Delegation
	// only present when using PDP
	Piece *piece.PieceLink
}

func blobAccept(ctx context.Context, service Service, req *BlobAcceptRequest) (*BlobAcceptResponse, error) {
	log := log.With("blob", digestutil.Format(req.Blob.Digest))
	log.Infof("%s %s", blob.AcceptAbility, req.Space)

	var (
		err      error
		loc      url.URL
		pdpPiece piece.PieceLink
		resp     = new(BlobAcceptResponse)
	)
	if service.PDP() == nil {
		_, err = service.Blobs().Store().Get(ctx, req.Blob.Digest)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return nil, NewAllocatedMemoryNotWrittenError()
			}
			log.Errorw("getting blob", "error", err)
			return nil, fmt.Errorf("getting blob: %w", err)
		}

		loc, err = service.Blobs().Access().GetDownloadURL(req.Blob.Digest)
		if err != nil {
			log.Errorw("creating retrieval URL for blob", "error", err)
			return nil, fmt.Errorf("creating retrieval URL for blob: %w", err)
		}
	} else {
		// locate the piece from the pdp service
		pdpPiece, err = service.PDP().PieceFinder().FindPiece(ctx, req.Blob.Digest, req.Blob.Size)
		if err != nil {
			log.Errorw("finding piece for blob", "error", err)
			return nil, fmt.Errorf("finding piece for blob: %w", err)
		}
		// get a download url
		loc = service.PDP().PieceFinder().URLForPiece(pdpPiece)
		// submit the piece for aggregation
		err = service.PDP().Aggregator().AggregatePiece(ctx, pdpPiece)
		if err != nil {
			log.Errorw("submitting piece for aggregation", "error", err)
			return nil, fmt.Errorf("submitting piece for aggregation: %w", err)
		}
		resp.Piece = &pdpPiece
	}

	claim, err := assert.Location.Delegate(
		service.ID(),
		req.Space,
		service.ID().DID().String(),
		assert.LocationCaveats{
			Space:    req.Space,
			Content:  types.FromHash(req.Blob.Digest),
			Location: []url.URL{loc},
		},
		delegation.WithNoExpiration(),
	)
	if err != nil {
		log.Errorw("creating location commitment", "error", err)
		return nil, fmt.Errorf("creating location commitment: %w", err)
	}

	err = service.Claims().Store().Put(ctx, claim)
	if err != nil {
		log.Errorw("putting location claim for blob", "error", err)
		return nil, fmt.Errorf("putting location claim for blob: %w", err)
	}

	err = service.Claims().Publisher().Publish(ctx, claim)
	if err != nil {
		log.Errorw("publishing location commitment", "error", err)
		return nil, fmt.Errorf("publishing location commitment: %w", err)
	}

	resp.Claim = claim
	return resp, nil
}
