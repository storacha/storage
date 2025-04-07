package storage

import (
	"context"
	"fmt"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/pdp"
	"github.com/storacha/go-libstoracha/capabilities/replica"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/dag/blockstore"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/receipt/fx"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/core/result/failure"
	fdm "github.com/storacha/go-ucanto/core/result/failure/datamodel"
	"github.com/storacha/go-ucanto/server"
	"github.com/storacha/go-ucanto/ucan"

	"github.com/storacha/storage/pkg/service/capabilities"
	"github.com/storacha/storage/pkg/service/replicator"
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

					resp, err := storageService.Capabilities().BlobAllocate(ctx, &capabilities.BlobAllocateRequest{
						Space: cap.Nb().Space,
						Blob:  cap.Nb().Blob,
						Cause: inv.Link(),
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
					resp, err := storageService.Capabilities().BlobAccept(ctx, &capabilities.BlobAcceptRequest{
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
							log.Error("creating piece accept invocation", "error", err)
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
		server.WithServiceMethod(
			replica.AllocateAbility,
			server.Provide(
				replica.Allocate,
				func(cap ucan.Capability[replica.AllocateCaveats], inv invocation.Invocation, iCtx server.InvocationContext) (replica.AllocateOk, fx.Effects, error) {
					//
					// UCAN Validation
					//

					// only service principal can perform an allocation
					if cap.With() != iCtx.ID().DID().String() {
						return replica.AllocateOk{}, nil, NewUnsupportedCapabilityError(cap)
					}

					//
					// end UCAN Validation
					//

					// create the transfer invocation: an fx of the allocate invocation receipt.
					trnsfInv, err := replica.Transfer.Invoke(
						storageService.ID(),
						storageService.ID(),
						storageService.ID().DID().GoString(),
						replica.TransferCaveats{
							Space:    cap.Nb().Space,
							Blob:     cap.Nb().Blob,
							Location: cap.Nb().Location,
							Cause:    inv.Link(),
						},
					)
					if err != nil {
						return replica.AllocateOk{}, nil, failure.FromError(err)
					}

					// read the location claim from this invocation to obtain the DID of the URL
					// to replicate from on the primary storage node.
					br, err := blockstore.NewBlockReader(blockstore.WithBlocksIterator(inv.Blocks()))
					if err != nil {
						return replica.AllocateOk{}, nil, failure.FromError(err)
					}
					claim, err := delegation.NewDelegationView(cap.Nb().Location, br)
					if err != nil {
						return replica.AllocateOk{}, nil, failure.FromError(err)
					}

					// TODO since there is a slice of capabilities here we need to validate the 0th is the correct one
					// unsure what `With()` should be compared with for a capability.
					lc, err := assert.LocationCaveatsReader.Read(claim.Capabilities()[0].Nb())
					if err != nil {
						return replica.AllocateOk{}, nil, failure.FromError(err)
					}

					if len(lc.Location) < 1 {
						return replica.AllocateOk{}, nil, failure.FromError(fmt.Errorf("location missing from location claim"))
					}

					// TODO: which one do we pick if > 1?
					replicaAddress := lc.Location[0]

					// FIXME: use a real context, requires changes to server
					ctx := context.TODO()
					allocateResp, err := storageService.Capabilities().BlobAllocate(ctx, &capabilities.BlobAllocateRequest{
						Space: cap.Nb().Space,
						Blob:  cap.Nb().Blob,
						Cause: inv.Link(),
					})
					if err != nil {
						return replica.AllocateOk{}, nil, failure.FromError(err)
					}

					// will run replication async, sending the receipt of the transfer invocation
					// to the upload service.
					if err := storageService.Replicator().Replicate(ctx, &replicator.Task{
						Space:      cap.Nb().Space,
						Blob:       cap.Nb().Blob,
						Source:     replicaAddress,
						Sink:       allocateResp.Address.URL,
						Invocation: trnsfInv,
					}); err != nil {
						return replica.AllocateOk{}, nil, failure.FromError(fmt.Errorf("failed to enqueue replication task: %w", err))
					}

					return replica.AllocateOk{Size: allocateResp.Size}, fx.NewEffects(fx.WithFork(fx.FromInvocation(trnsfInv))), nil
				},
			),
		),
	)

	return server.NewServer(storageService.ID(), options...)
}
