package storage

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
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
					//
					// UCAN Validation
					//

					// only service principal can perform an allocation
					if cap.With() != iCtx.ID().DID().String() {
						return blob.AllocateOk{}, nil, NewUnsupportedCapabilityError(cap)
					}

					// enforce max upload size requirements
					if cap.Nb().Blob.Size > maxUploadSize {
						return blob.AllocateOk{}, nil, NewBlobSizeLimitExceededError(cap.Nb().Blob.Size, maxUploadSize)
					}

					//
					// end UCAN Validation
					//

					// FIXME: use a real context, requires changes to server
					ctx := context.TODO()
					resp, err := blobAllocate(ctx, storageService, &BlobAllocateRequest{
						Space:           cap.Nb().Space,
						Blob:            cap.Nb().Blob,
						AllocationCause: cap.Nb().Cause,
						InvocationCause: inv.Link(),
					})
					if err != nil {
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					return blob.AllocateOk{
						Size:    resp.Size,
						Address: resp.Address,
					}, nil, nil
				},
			),
		),
		server.WithServiceMethod(
			blob.AcceptAbility,
			server.Provide(
				blob.Accept,
				func(cap ucan.Capability[blob.AcceptCaveats], inv invocation.Invocation, iCtx server.InvocationContext) (blob.AcceptOk, fx.Effects, error) {
					//
					// UCAN Validation
					//

					// only service principal can perform an allocation
					if cap.With() != iCtx.ID().DID().String() {
						return blob.AcceptOk{}, nil, NewUnsupportedCapabilityError(cap)
					}

					//
					// end UCAN Validation
					//

					// FIXME: use a real context, requires changes to server
					ctx := context.TODO()
					resp, err := blobAccept(ctx, storageService, &BlobAcceptRequest{
						Space: cap.Nb().Space,
						Blob:  cap.Nb().Blob,
						Put:   cap.Nb().Put,
					})
					if err != nil {
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}

					var forks []fx.Effect
					forks = append(forks, fx.FromInvocation(resp.Claim))

					var pdpLink *ucan.Link
					if resp.Piece != nil {
						// generate the invocation that will complete when aggregation is complete and the piece is accepted
						pieceAccept, err := pdp.Accept.Invoke(
							storageService.ID(),
							storageService.ID(),
							storageService.ID().DID().String(),
							pdp.AcceptCaveats{
								Piece: *resp.Piece,
							}, delegation.WithNoExpiration())
						if err != nil {
							log.Errorf("creating piece accept invocation: %w", err)
							return blob.AcceptOk{}, nil, failure.FromError(err)
						}
						pieceAcceptLink := pieceAccept.Link()
						pdpLink = &pieceAcceptLink
						forks = append(forks, fx.FromInvocation(pieceAccept))
					}

					return blob.AcceptOk{Site: resp.Claim.Link(), PDP: pdpLink}, fx.NewEffects(fx.WithFork(forks...)), nil
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
