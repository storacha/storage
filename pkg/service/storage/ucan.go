package storage

import (
	"context"
	"errors"
	"net/url"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-capabilities/pkg/assert"
	"github.com/storacha/go-capabilities/pkg/blob"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/receipt/fx"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/server"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/store"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
)

var log = logging.Logger("storage")

const maxUploadSize = 127 * (1 << 25)

func NewUCANServer(storageService Service) (server.ServerView, error) {
	return server.NewServer(
		storageService.ID(),
		server.WithServiceMethod(
			blob.AllocateAbility,
			server.Provide(
				blob.Allocate,
				func(cap ucan.Capability[blob.AllocateCaveats], inv invocation.Invocation, iCtx server.InvocationContext) (blob.AllocateOk, fx.Effects, error) {
					ctx := context.TODO()
					digest := cap.Nb().Blob.Digest
					log := log.With("blob", digestutil.Format(digest))
					log.Infof("%s space: %s", blob.AllocateAbility, cap.Nb().Space)

					// only service principal can perform an allocation
					if cap.With() != iCtx.ID().DID().String() {
						return blob.AllocateOk{}, nil, NewUnsupportedCapabilityError(cap)
					}

					// check if we already have an allcoation for the blob in this space
					allocs, err := storageService.Blobs().Allocations().List(ctx, digest)
					if err != nil {
						log.Errorf("getting allocations: %w", err)
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					for _, a := range allocs {
						// if we find an allocation, check if we have the blob.
						if a.Space == cap.Nb().Space {
							_, err := storageService.Blobs().Store().Get(ctx, digest)
							if err == nil {
								// if we have it, it does not need upload
								return blob.AllocateOk{Size: 0}, nil, nil
							}
							if !errors.Is(err, store.ErrNotFound) {
								log.Errorf("getting blob: %w", err)
								return blob.AllocateOk{}, nil, failure.FromError(err)
							}
						}
					}

					if cap.Nb().Blob.Size > maxUploadSize {
						return blob.AllocateOk{}, nil, NewBlobSizeLimitExceededError(cap.Nb().Blob.Size, maxUploadSize)
					}

					expiresIn := uint64(60 * 60 * 24) // 1 day
					expiresAt := uint64(time.Now().Unix()) + expiresIn
					url, headers, err := storageService.Blobs().Presigner().SignUploadURL(ctx, digest, cap.Nb().Blob.Size, expiresIn)
					if err != nil {
						log.Errorf("signing upload URL: %w", err)
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					err = storageService.Blobs().Allocations().Put(ctx, allocation.Allocation{
						Space:   cap.Nb().Space,
						Blob:    allocation.Blob(cap.Nb().Blob),
						Expires: expiresAt,
						Cause:   inv.Link(),
					})
					if err != nil {
						log.Errorf("putting allocation: %w", err)
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					return blob.AllocateOk{
						Size: cap.Nb().Blob.Size,
						Address: &blob.Address{
							URL:     url,
							Headers: headers,
							Expires: expiresAt,
						},
					}, nil, nil
				},
			),
		),
		server.WithServiceMethod(
			blob.AcceptAbility,
			server.Provide(
				blob.Accept,
				func(cap ucan.Capability[blob.AcceptCaveats], inv invocation.Invocation, iCtx server.InvocationContext) (blob.AcceptOk, fx.Effects, error) {
					ctx := context.TODO()
					digest := cap.Nb().Blob.Digest
					log := log.With("blob", digestutil.Format(digest))
					log.Infof("%s %s", blob.AcceptAbility, cap.Nb().Space)

					// only service principal can perform an allocation
					if cap.With() != iCtx.ID().DID().String() {
						return blob.AcceptOk{}, nil, NewUnsupportedCapabilityError(cap)
					}

					_, err := storageService.Blobs().Store().Get(ctx, digest)
					if err != nil {
						if errors.Is(err, store.ErrNotFound) {
							return blob.AcceptOk{}, nil, NewAllocatedMemoryNotWrittenError()
						}
						log.Errorf("getting blob: %w", err)
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}

					loc, err := storageService.Blobs().Access().GetDownloadURL(digest)
					if err != nil {
						log.Errorf("creating retrieval URL for blob: %w", err)
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}

					claim, err := assert.Location.Delegate(
						storageService.ID(),
						cap.Nb().Space,
						storageService.ID().DID().String(),
						assert.LocationCaveats{
							Space:    cap.Nb().Space,
							Content:  assert.FromHash(digest),
							Location: []url.URL{loc},
						},
						delegation.WithNoExpiration(),
					)
					if err != nil {
						log.Errorf("creating location commitment: %w", err)
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}

					err = storageService.Claims().Store().Put(ctx, claim)
					if err != nil {
						log.Errorf("putting location claim for blob: %w", err)
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}

					err = storageService.Claims().Publisher().Publish(ctx, claim)
					if err != nil {
						log.Errorf("publishing location commitment: %w", err)
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}

					return blob.AcceptOk{Site: claim.Link()}, fx.NewEffects(fx.WithFork(fx.FromInvocation(claim))), nil
				},
			),
		),
	)
}
