package capabilities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/principal"

	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/pdp"
	"github.com/storacha/storage/pkg/service/blobs"
	"github.com/storacha/storage/pkg/service/claims"
	"github.com/storacha/storage/pkg/store"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
)

var log = logging.Logger("capabilities")

var _ Capabilities = (*Service)(nil)

func New(id principal.Signer, blob blobs.Blobs, claim claims.Claims, p pdp.PDP) (*Service, error) {
	return &Service{
		id:    id,
		blob:  blob,
		claim: claim,
		pdp:   p,
	}, nil
}

type Service struct {
	id    principal.Signer
	blob  blobs.Blobs
	claim claims.Claims
	pdp   pdp.PDP
}

func (s *Service) BlobAllocate(ctx context.Context, req *BlobAllocateRequest) (*BlobAllocateResponse, error) {
	log := log.With("blob", digestutil.Format(req.Blob.Digest))
	log.Infof("%s space: %s", blob.AllocateAbility, req.Space)

	// check if we already have an allocation for the blob in this space
	allocs, err := s.blob.Allocations().List(ctx, req.Blob.Digest)
	if err != nil {
		log.Errorw("getting allocations", "error", err)
		return nil, fmt.Errorf("getting allocations: %w", err)
	}

	allocated := false
	for _, a := range allocs {
		if a.Space == req.Space {
			allocated = true
			break
		}
	}

	received := false
	// check if we received the blob (only possible if we have an allocation)
	if len(allocs) > 0 {
		if s.pdp != nil {
			_, err = s.pdp.PieceFinder().FindPiece(ctx, req.Blob.Digest, req.Blob.Size)
		} else {
			_, err = s.blob.Store().Get(ctx, req.Blob.Digest)
		}
		if err == nil {
			received = true
		}
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			log.Errorw("getting blob", "error", err)
			return nil, fmt.Errorf("getting blob: %w", err)
		}
	}

	// the size reported in the receipt is the number of bytes allocated
	// in the space - if a previous allocation already exists, this has
	// already been done, so the allocation size is 0
	size := req.Blob.Size
	if allocated {
		log.Info("blob allocation already exists")
		size = 0
	}

	// nothing to do
	if allocated && received {
		log.Info("blob already received")
		return &BlobAllocateResponse{
			Size: size,
			// NB: blob already receieved, therefor no address is needed for upload.
			Address: nil,
		}, nil
	}

	expiresIn := uint64(60 * 60 * 24) // 1 day
	expiresAt := uint64(time.Now().Unix()) + expiresIn

	var address *blob.Address
	// if not received yet, we need to generate a signed URL for the
	// upload, and include it in the receipt.
	if !received {
		var uploadURL url.URL
		headers := http.Header{}
		if s.pdp == nil {
			// use standard blob upload
			uploadURL, headers, err = s.blob.Presigner().SignUploadURL(ctx, req.Blob.Digest, req.Blob.Size, expiresIn)
			if err != nil {
				log.Errorw("signing upload URL", "error", err)
				return nil, fmt.Errorf("signing upload URL: %w", err)
			}
		} else {
			// use pdp service upload
			urlP, err := s.pdp.PieceAdder().AddPiece(ctx, req.Blob.Digest, req.Blob.Size)
			if err != nil {
				log.Errorw("adding to pdp service", "error", err)
				return nil, fmt.Errorf("adding to pdp service: %w", err)
			}
			uploadURL = *urlP
		}
		address = &blob.Address{
			URL:     uploadURL,
			Headers: headers,
			Expires: expiresAt,
		}
	}

	// even if a previous allocation was made in this space, we create
	// another for the new invocation.
	err = s.blob.Allocations().Put(ctx, allocation.Allocation{
		Space:   req.Space,
		Blob:    allocation.Blob(req.Blob),
		Expires: expiresAt,
		// REVIEW: is this the correct cause? The invocation link rather than the allocate caveats cause field?
		Cause: req.Cause,
	})
	if err != nil {
		log.Errorw("putting allocation", "error", err)
		return nil, fmt.Errorf("putting allocation: %w", err)
	}

	a, err := s.blob.Allocations().List(ctx, req.Blob.Digest)
	if err != nil {
		log.Errorw("listing allocation after put", "error", err)
		return nil, fmt.Errorf("listing allocation after put: %w", err)
	}
	if len(a) < 1 {
		log.Error("failed to read allocation after write")
		return nil, fmt.Errorf("failed to read allocation after write")
	}
	log.Info("successfully read allocation after write")

	return &BlobAllocateResponse{
		Size:    size,
		Address: address,
	}, nil

}

func (s *Service) BlobAccept(ctx context.Context, req *BlobAcceptRequest) (*BlobAcceptResponse, error) {
	log := log.With("blob", digestutil.Format(req.Blob.Digest))
	log.Infof("%s %s", blob.AcceptAbility, req.Space)

	var (
		err      error
		loc      url.URL
		pdpPiece piece.PieceLink
		resp     = new(BlobAcceptResponse)
	)
	if s.pdp == nil {
		_, err = s.blob.Store().Get(ctx, req.Blob.Digest)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return nil, fmt.Errorf("blob not found: %w", err)
			}
			log.Errorw("getting blob", "error", err)
			return nil, fmt.Errorf("getting blob: %w", err)
		}

		loc, err = s.blob.Access().GetDownloadURL(req.Blob.Digest)
		if err != nil {
			log.Errorw("creating retrieval URL for blob", "error", err)
			return nil, fmt.Errorf("creating retrieval URL for blob: %w", err)
		}
	} else {
		// locate the piece from the pdp service
		pdpPiece, err = s.pdp.PieceFinder().FindPiece(ctx, req.Blob.Digest, req.Blob.Size)
		if err != nil {
			log.Errorw("finding piece for blob", "error", err)
			return nil, fmt.Errorf("finding piece for blob: %w", err)
		}
		// get a download url
		loc = s.pdp.PieceFinder().URLForPiece(pdpPiece)
		// submit the piece for aggregation
		err = s.pdp.Aggregator().AggregatePiece(ctx, pdpPiece)
		if err != nil {
			log.Errorw("submitting piece for aggregation", "error", err)
			return nil, fmt.Errorf("submitting piece for aggregation: %w", err)
		}
		resp.Piece = &pdpPiece
	}

	claim, err := assert.Location.Delegate(
		s.id,
		req.Space,
		s.id.DID().String(),
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

	err = s.claim.Store().Put(ctx, claim)
	if err != nil {
		log.Errorw("putting location claim for blob", "error", err)
		return nil, fmt.Errorf("putting location claim for blob: %w", err)
	}

	err = s.claim.Publisher().Publish(ctx, claim)
	if err != nil {
		log.Errorw("publishing location commitment", "error", err)
		return nil, fmt.Errorf("publishing location commitment: %w", err)
	}

	resp.Claim = claim
	return resp, nil

}
