package storage

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/pdp"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/receipt/fx"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/core/result/failure"
	fdm "github.com/storacha/go-ucanto/core/result/failure/datamodel"
	"github.com/storacha/go-ucanto/server"
	"github.com/storacha/go-ucanto/ucan"

	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/store"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
)

var log = logging.Logger("storage")

const maxUploadSize = 127 * (1 << 25)

func NewUCANServer(storageService Service, options ...server.Option) (server.ServerView, error) {
	options = append(
		options,
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
					// check if we already have an allocation for the blob in this space
					allocs, err := storageService.Blobs().Allocations().List(ctx, digest)
					if err != nil {
						log.Errorf("getting allocations: %w", err)
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					allocated := false
					for _, a := range allocs {
						if a.Space == cap.Nb().Space {
							allocated = true
							break
						}
					}

					received := false
					// check if we received the blob (only possible if we have an allocation)
					if len(allocs) > 0 {
						if storageService.PDP() != nil {
							_, err = storageService.PDP().PieceFinder().FindPiece(ctx, digest, cap.Nb().Blob.Size)
						} else {
							_, err = storageService.Blobs().Store().Get(ctx, digest)
						}
						if err == nil {
							received = true
						}
						if err != nil && !errors.Is(err, store.ErrNotFound) {
							log.Errorf("getting blob: %w", err)
							return blob.AllocateOk{}, nil, failure.FromError(err)
						}
					}

					// the size reported in the receipt is the number of bytes allocated
					// in the space - if a previous allocation already exists, this has
					// already been done, so the allocation size is 0
					size := cap.Nb().Blob.Size
					if allocated {
						size = 0
					}

					// nothing to do
					if allocated && received {
						return blob.AllocateOk{Size: size}, nil, nil
					}

					if cap.Nb().Blob.Size > maxUploadSize {
						return blob.AllocateOk{}, nil, NewBlobSizeLimitExceededError(cap.Nb().Blob.Size, maxUploadSize)
					}

					expiresIn := uint64(60 * 60 * 24) // 1 day
					expiresAt := uint64(time.Now().Unix()) + expiresIn

					var address *blob.Address
					// if not received yet, we need to generate a signed URL for the
					// upload, and include it in the receipt.
					if !received {
						var url url.URL
						headers := http.Header{}
						if storageService.PDP() == nil {
							// use standard blob upload
							url, headers, err = storageService.Blobs().Presigner().SignUploadURL(ctx, digest, cap.Nb().Blob.Size, expiresIn)
							if err != nil {
								log.Errorf("signing upload URL: %w", err)
								return blob.AllocateOk{}, nil, failure.FromError(err)
							}
						} else {
							// use pdp service upload
							urlP, err := storageService.PDP().PieceAdder().AddPiece(ctx, digest, cap.Nb().Blob.Size)
							if err != nil {
								log.Errorf("adding to pdp service: %w", err)
								return blob.AllocateOk{}, nil, failure.FromError(err)
							}
							url = *urlP
						}
						address = &blob.Address{
							URL:     url,
							Headers: headers,
							Expires: expiresAt,
						}
					}

					// even if a previous allocation was made in this space, we create
					// another for the new invocation.
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

					a, err := storageService.Blobs().Allocations().List(ctx, cap.Nb().Blob.Digest)
					if err != nil {
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}
					if len(a) < 1 {
						return blob.AllocateOk{}, nil, failure.FromError(errors.New("failed to read allocation after write"))
					}
					log.Info("successfully read allocation after write")

					return blob.AllocateOk{Size: size, Address: address}, nil, nil
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

					var loc url.URL

					var forks []fx.Effect
					var pdpLink *ucan.Link
					if storageService.PDP() == nil {
						_, err := storageService.Blobs().Store().Get(ctx, digest)
						if err != nil {
							if errors.Is(err, store.ErrNotFound) {
								return blob.AcceptOk{}, nil, NewAllocatedMemoryNotWrittenError()
							}
							log.Errorf("getting blob: %w", err)
							return blob.AcceptOk{}, nil, failure.FromError(err)
						}

						loc, err = storageService.Blobs().Access().GetDownloadURL(digest)
						if err != nil {
							log.Errorf("creating retrieval URL for blob: %w", err)
							return blob.AcceptOk{}, nil, failure.FromError(err)
						}
					} else {
						// locate the piece from the pdp service
						piece, err := storageService.PDP().PieceFinder().FindPiece(ctx, digest, cap.Nb().Blob.Size)
						if err != nil {
							log.Errorf("finding piece for blob: %w", err)
							return blob.AcceptOk{}, nil, failure.FromError(err)
						}
						// get a download url
						loc = storageService.PDP().PieceFinder().URLForPiece(piece)
						// submit the piece for aggregation
						err = storageService.PDP().Aggregator().AggregatePiece(ctx, piece)
						if err != nil {
							log.Errorf("submitting piece for aggregation: %w", err)
							return blob.AcceptOk{}, nil, failure.FromError(err)
						}
						// generate the invocation that will complete when aggregation is complete and the piece is accepted
						pieceAccept, err := pdp.Accept.Invoke(
							storageService.ID(),
							storageService.ID(),
							storageService.ID().DID().String(),
							pdp.AcceptCaveats{
								Piece: piece,
							}, delegation.WithNoExpiration())
						if err != nil {
							log.Errorf("creating piece accept invocation: %w", err)
							return blob.AcceptOk{}, nil, failure.FromError(err)
						}
						pieceAcceptLink := pieceAccept.Link()
						pdpLink = &pieceAcceptLink
						forks = append(forks, fx.FromInvocation(pieceAccept))
					}
					claim, err := assert.Location.Delegate(
						storageService.ID(),
						cap.Nb().Space,
						storageService.ID().DID().String(),
						assert.LocationCaveats{
							Space:    cap.Nb().Space,
							Content:  types.FromHash(digest),
							Location: []url.URL{loc},
						},
						delegation.WithNoExpiration(),
					)
					if err != nil {
						log.Errorf("creating location commitment: %w", err)
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}
					forks = append(forks, fx.FromInvocation(claim))

					err = storageService.Claims().Store().Put(ctx, claim)
					if err != nil {
						log.Errorf("putting location claim for blob: %w", err)
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}

					/*
						err = storageService.Claims().Publisher().Publish(ctx, claim)
						if err != nil {
							log.Errorf("publishing location commitment: %w", err)
							return blob.AcceptOk{}, nil, failure.FromError(err)
						}
					*/

					return blob.AcceptOk{Site: claim.Link(), PDP: pdpLink}, fx.NewEffects(fx.WithFork(forks...)), nil
				},
			),
		),
		server.WithServiceMethod(
			pdp.InfoAbility,
			server.Provide(
				pdp.Info,
				func(cap ucan.Capability[pdp.InfoCaveats], inv invocation.Invocation, iCtx server.InvocationContext) (pdp.InfoOk, fx.Effects, error) {
					ctx := context.TODO()
					// generate the invocation that would submit when this was first submitted
					pieceAccept, err := pdp.Accept.Invoke(
						storageService.ID(),
						storageService.ID(),
						storageService.ID().DID().GoString(),
						pdp.AcceptCaveats{
							Piece: cap.Nb().Piece,
						}, delegation.WithNoExpiration())
					if err != nil {
						log.Errorf("creating location commitment: %w", err)
						return pdp.InfoOk{}, nil, failure.FromError(err)
					}
					// look up the receipt for the accept invocation
					rcpt, err := storageService.Receipts().GetByRan(ctx, pieceAccept.Link())
					if err != nil {
						log.Errorf("looking up receipt: %w", err)
						return pdp.InfoOk{}, nil, failure.FromError(err)
					}
					// rebind the receipt to get the specific types for pdp/accept
					pieceAcceptReceipt, err := receipt.Rebind[pdp.AcceptOk, fdm.FailureModel](rcpt, pdp.AcceptOkType(), fdm.FailureType(), types.Converters...)
					if err != nil {
						log.Errorf("reading piece accept receipt: %w", err)
						return pdp.InfoOk{}, nil, failure.FromError(err)
					}
					// use the result from the accept receipt to generate the receipt for pdp/info
					return result.MatchResultR3(pieceAcceptReceipt.Out(),
						func(ok pdp.AcceptOk) (pdp.InfoOk, fx.Effects, error) {
							return pdp.InfoOk{
								Piece: cap.Nb().Piece,
								Aggregates: []pdp.InfoAcceptedAggregate{
									{
										Aggregate:      ok.Aggregate,
										InclusionProof: ok.InclusionProof,
									},
								},
							}, nil, nil
						},
						func(err fdm.FailureModel) (pdp.InfoOk, fx.Effects, error) {
							return pdp.InfoOk{}, nil, failure.FromFailureModel(err)
						},
					)
				},
			),
		),
	)

	return server.NewServer(storageService.ID(), options...)
}
