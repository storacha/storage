package storage

/*
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/capabilities/blob"
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
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/server"
	"github.com/storacha/go-ucanto/ucan"

	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
)

type ReplicaAllocateRequest struct {
	Space    did.DID
	Blob     blob.Blob
	Location cid.Cid
	Cause    ucan.Link
}

type ReplicaAllocateResponse struct {
	Size uint64
}

func replicaAllocate(
	ctx context.Context,
	service Service,
	req *ReplicaAllocateRequest,
) (*ReplicaAllocateResponse, error) {

	resp, err := blobAllocate(ctx, service)

}

func ReplicaAllocate(
	ctx context.Context,
	storageService Service,
	cap ucan.Capability[replica.AllocateCaveats],
	inv invocation.Invocation,
	iCtx server.InvocationContext,
) (replica.AllocateOK, fx.Effects, error) {
	// NB: this method is essentially a wrapper around BlobAllocate with a different set of caveats and effects

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
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}

	allocateCaveats := blob.AllocateCaveats{
		Space: cap.Nb().Space,
		Blob:  cap.Nb().Blob,
		Cause: cap.Nb().Cause,
	}
	allocateReceipt, allocateFx, err := BlobAllocate(ctx, storageService, allocateCaveats, inv, iCtx)
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}
	// TODO: We should separate the UCAN bits our from BlobAllocate to simplify things a bit?
	// push the UCAN parts to the uncan.go file and have the calls wrap these which are ucan-less
	// NB(forrest): BlobAllocate doesn't return any effects. So we ignore tese
	_ = allocateFx

	go func() {
		transferReceipt, transferFs, err := ReplicaTransfer(ctx, storageService, nil)
	}()

	// name make fx for the transfer, then asynchronously invoke and run the transfer
	// meaning we need a jobqueue :3
	return replica.AllocateOK{Size: allocateReceipt.Size}, fx.NewEffects(fx.WithFork(fx.FromInvocation(trnsfInv))), nil
}

func ReplicaTransfer(
	ctx context.Context,
	storageService Service,
	cap ucan.Capability[replica.TransferCaveats],
	inv invocation.Invocation,
	iCtx server.InvocationContext,
) (replica.TransferOK, fx.Effects, error) {
	// read the location claim from this invocation.
	br, err := blockstore.NewBlockReader(blockstore.WithBlocksIterator(inv.Blocks()))
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}

	claim, err := delegation.NewDelegationView(cidlink.Link{Cid: cap.Nb().Location}, br)
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}

	lc, err := assert.LocationCaveatsReader.Read(claim.Capabilities()[0].Nb())
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}

	if len(lc.Location) < 1 {
		return replica.TransferOK{}, nil, failure.FromError(fmt.Errorf("location missing from location claim"))
	}

	// TODO is there a better way to pick when more than one is present?
	replicaAddress := lc.Location[0]

	// pull the data from the primary node.
	resp, err := http.Get(replicaAddress.String())
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}

	// we need to read the 'cause' of this invocation to get back the address
	// we created to upload the data to.
	// TODO: ZERO idea if this is the correct way to do this. Probably not.
	reader, err := receipt.NewReceiptReaderFromTypes[replica.AllocateOK, fdm.FailureModel](replica.AllocateOkType(), fdm.FailureType(), types.Converters...)
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}
	rcpt, err := reader.Read(inv.Link(), inv.Blocks())
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(fmt.Errorf("failed to read receipt: %w", err))
	}
	alloc, err := result.Unwrap(result.MapError(rcpt.Out(), failure.FromFailureModel))
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}

	log.Info("now uploading to: %s", alloc.Address)
	// stream the replica response to PDP service
	req, err := http.NewRequest(http.MethodPut, alloc.Address.URL.String(), resp.Body)
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}
	req.Header = alloc.Address.Headers
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		resData, err := io.ReadAll(res.Body)
		if err != nil {
			return replica.TransferOK{}, nil, failure.FromError(err)
		}
		err = fmt.Errorf("unsuccessful put, status: %s, message: %s", res.Status, string(resData))
		return replica.TransferOK{}, nil, failure.FromError(err)
	}

	acceptCaveats := blob.AcceptCaveats{
		Space: nil, // TODO: we need a space, one is given in the initial allocate invocation, should read from there?,
		Blob:  cap.Nb().Blob,
		Put: blob.Promise{
			UcanAwait: blob.Await{
				Selector: ".out.ok",
				Link:     inv.Link(),
			},
		},
	}

	acceptReceipt, acceptFx, err := BlobAccept(ctx, storageService, acceptCaveats, inv, iCtx)
	if err != nil {
		return replica.TransferOK{}, nil, failure.FromError(err)
	}

	return replica.TransferOK{
		Site: acceptReceipt.Site,
		PDP:  acceptReceipt.PDP,
	}, acceptFx, nil

}

func ReplicaAllocate(
	ctx context.Context,
	storageService Service,
	cap ucan.Capability[replica.AllocateCaveats],
	inv invocation.Invocation,
	iCtx server.InvocationContext,
) (replica.AllocateOK, fx.Effects, error) {
	//
	//	The upload service MUST select storage node(s) and allocate replication space when the space/blob/replicate invocation is received.
	//	The upload service allocates replication space on storage nodes by issuing a blob/replica/allocate invocation.
	//	The blob/replica/allocate task receipt includes an async task that will be performed by the storage node - blob/replica/transfer.
	//	The blob/replica/transfer task is completed when the storage node has transferred the blob from its location to the storage node.
	//

	// NB(forrest): this is similar to blob allocate where we first check if an allocation already exists
	// for the blob we are instructed to replicate, in the event an allocation does exist, this should be a NOOP.
	digest := cap.Nb().Blob.Digest
	log := log.With("blob", digestutil.Format(digest))
	log.Infof("%s space: %s", replica.AllocateAbility, cap.Nb().Space)

	// only service principal can perform a replica allocation
	// TODO validate what this is doing is right.
	if cap.With() != iCtx.ID().DID().String() {
		return replica.AllocateOK{}, nil, NewUnsupportedCapabilityError(cap)
	}


	// check if we already have an allocation for the blob in this space
	allocs, err := storageService.Blobs().Allocations().List(ctx, digest)
	if err != nil {
		log.Errorf("getting allocations: %w", err)
		return replica.AllocateOK{}, nil, failure.FromError(err)
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
			return replica.AllocateOK{}, nil, failure.FromError(err)
		}
	}

	// the size reported in the receipt is the number of bytes allocated
	// in the space - if a previous allocation already exists, this has
	// already been done, so the allocation size is 0
	size := cap.Nb().Blob.Size
	if allocated {
		size = 0
	}

	// TODO we have already allocated and received the data, how do we signal this?
	// for now this should be rare, so ignore it.
	if allocated && received {
		return replica.AllocateOK{}, nil, nil
	}

	// We haven't received the blob to be replicated yet, in this case we MUST:
	//	- issue a blob/replica/transfer invocation to the original storage node
	//		- TODO: how do we know the "original storage node" to issue the request to?
	//	- receive the data
	//		- based on the location commitment in the request we know where to pull the data from.
	// 		  This step should be a matter of an HTTP GET to the URL(s) in the location commitment of the request
	if !received {
		// this is roughly what you will get in the location commitment of the request
		//
		//	claim, err := assert.Location.Delegate(
		//		storageService.ID(),
		//		cap.Nb().Space,
		//		storageService.ID().DID().String(),
		//		assert.LocationCaveats{
		//			Space:    cap.Nb().Space,
		//			Content:  types.FromHash(digest),
		//			Location: []url.URL{loc},
		//		},
		//		delegation.WithNoExpiration(),
		//	)
		//
		// read the location claim from this invocation to obtain the DID of the storage
		// node we are replicating _from_.
		br, err := blockstore.NewBlockReader(blockstore.WithBlocksIterator(inv.Blocks()))
		if err != nil {
			return replica.AllocateOK{}, nil, failure.FromError(err)
		}
		claim, err := delegation.NewDelegationView(cidlink.Link{Cid: cap.Nb().Location}, br)
		if err != nil {
			return replica.AllocateOK{}, nil, failure.FromError(err)
		}
		theOriginalStorageNode := claim.Issuer().DID()
		if len(claim.Capabilities()) != 1 || claim.Capabilities()[0].Can() != assert.LocationAbility {
			return replica.AllocateOK{}, nil, failure.FromError(fmt.Errorf("TODO: invalid capabilites in location claim"))
		}

		lc, err := assert.LocationCaveatsReader.Read(claim.Capabilities()[0].Nb())
		if err != nil {
			return replica.AllocateOK{}, nil, failure.FromError(err)
		}

		if len(lc.Location) < 1 {
			return replica.AllocateOK{}, nil, failure.FromError(fmt.Errorf("location missing from location claim"))
		}

		// TODO maybe more validation here, need to pick a good one from the slice
		whereToSendTransferinvocation := fmt.Sprintf("%s://%s", lc.Location[0].Scheme, lc.Location[0].Host)

		transferInvocation, err := replica.Transfer.Invoke(
			storageService.ID(),
			theOriginalStorageNode,
			storageService.ID().DID().GoString(),
			replica.TransferCaveats{
				Blob:     cap.Nb().Blob,
				Location: cap.Nb().Location,
				Cause:    inv.Link(),
			},
		)
		if err != nil {
			return replica.AllocateOK{}, nil, failure.FromError(err)
		}
		// TODO this is a hack until we get the retrival flows working and spec'd
		// kick off a go routine that will fetch the data
		// TODO make this reliable, node can be offline, request can fail, this node can die.
		go func() {
			// TODO: the tasks you need to do:
			// - get the data
			// - verify it with the content matching
			// - issue a location commitment
			// - submit for PDP agg
			// 		- essentially follow the blob accept flow
			// - generate a receipt for the transfer allocation
			// 		- you will need to call receipt.Issue and invoke ucan conclude.
			// 		- the receipt from this needs to be sent back to the upload service, all async!
			// 		- meaning you need the upload-service did and url somewhere to send this receipt to
			//		  where we are acting as a client to the upload-service in this async operation.
			// TODO we can put the transfer invocation in the body of the request POST to get
			// the data as proof that we have permission to get the data.

			// TODO here is where we would use a retrival protocol, but for now we'll just fetch it
			// from the location we know.
			// replica := http.Client{}.Get(lc.Location[0], lc.Range)
			// TODO use lc.Content to verify the received data is what we expect.

			//  the rest here is a hybrid of allocate and accept where we put the data
			// into pdp and then accept it
			// TODO you will probably want to call go-ucanto/core/receipt/receipt.go.Issue
			// or something like this from the aggregator:
			//
			//	func GenerateReceipts(issuer ucan.Signer, aggregate aggregate.Aggregate) ([]receipt.AnyReceipt, error) {
			//		receipts := make([]receipt.AnyReceipt, 0, len(aggregate.Pieces))
			//		for _, aggregatePiece := range aggregate.Pieces {
			//			inv, err := pdp.Accept.Invoke(issuer, issuer, issuer.DID().String(), pdp.AcceptCaveats{
			//				Piece: aggregatePiece.Link,
			//			})
			//			if err != nil {
			//				return nil, fmt.Errorf("generating invocation: %w", err)
			//			}
			//			ok := result.Ok[pdp.AcceptOk, ipld.Builder](pdp.AcceptOk{
			//				Aggregate:      aggregate.Root,
			//				InclusionProof: aggregatePiece.InclusionProof,
			//				Piece:          aggregatePiece.Link,
			//			})
			//			rcpt, err := receipt.Issue(issuer, ok, ran.FromInvocation(inv))
			//			if err != nil {
			//				return nil, fmt.Errorf("issuing receipt: %w", err)
			//			}
			//			receipts = append(receipts, rcpt)
			//		}
			//		return receipts, nil
			//	}
			//
			// fetch the data
			resp, err := http.Get(lc.Location[0].String())
			if err != nil {
				// TODO
			}
			// TODO verify the content is matching - will punt this to the PDP service
			// TODO it would be neat if we could just call the existing blob allocate and blob accept methods on the server
			// wonder if we can do that instead eventually, or break their logic out into callable methods to avoid the
			// transport overhead, that's probably a better choice
			var url url.URL
			headers := http.Header{}
			expiresIn := uint64(60 * 60 * 24) // 1 day
			expiresAt := uint64(time.Now().Unix()) + expiresIn
			if storageService.PDP() == nil {
				// use standard blob upload
				url, headers, err = storageService.Blobs().Presigner().SignUploadURL(ctx, digest, cap.Nb().Blob.Size, expiresIn)
				if err != nil {
					log.Errorf("signing upload URL: %w", err)
					// TODO
				}
			} else {
				// use pdp service upload
				urlP, err := storageService.PDP().PieceAdder().AddPiece(ctx, digest, cap.Nb().Blob.Size)
				if err != nil {
					log.Errorf("adding to pdp service: %w", err)
					// TODO
				}
				url = *urlP
			}
			uploadAddress := &blob.Address{
				URL:     url,
				Headers: headers,
				Expires: expiresAt,
			}

			// even if a previous allocation was made in this space, we create
			// another for the new invocation.
			err = storageService.Blobs().Allocations().Put(ctx, allocation.Allocation{
				Space:   cap.Nb().Space,
				Blob:    allocation.Blob(cap.Nb().Blob),
				Expires: expiresAt,
				// TODO ensure this is the right cause of the invocation
				Cause: inv.Link(),
			})
			if err != nil {
				log.Errorf("putting allocation: %w", err)
				// TODO
			}

			a, err := storageService.Blobs().Allocations().List(ctx, cap.Nb().Blob.Digest)
			if err != nil {
				// TODO
			}
			if len(a) < 1 {
				// TODO
			}
			log.Info("successfully read allocation after write")

			fmt.Printf("now uploading to: %s\n", uploadAddress.URL.String())

			// we are going to stream the response from the original storage node
			req, err := http.NewRequest(http.MethodPut, uploadAddress.URL.String(), resp.Body)
			if err != nil {
				// TODO
			}
			req.Header = uploadAddress.Headers
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				// TODO
			}
			if res.StatusCode >= 300 || res.StatusCode < 200 {
				_, err := io.ReadAll(res.Body)
				if err != nil {
					// TODO
				}
				// TODO
			}

		}()
		// TODO the allocation put as is done in  latter half of blob.AllocateAbility
		return replica.AllocateOK{
			Size: size,
		}, fx.NewEffects(fx.WithFork(fx.FromInvocation(transferInvocation))), nil

	} //else
	// we have already received the data, no need to fetch it again
	// 	- TODO: what receipt to we issue in this event?
	//  	- how do we allow the "original storage node" to conclude this invocation chain?

}

	// read the location claim from this invocation to obtain the DID of the storage
	// node we are replicating _from_.
	br, err := blockstore.NewBlockReader(blockstore.WithBlocksIterator(inv.Blocks()))
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}
	claim, err := delegation.NewDelegationView(cidlink.Link{Cid: cap.Nb().Location}, br)
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}
	lc, err := assert.LocationCaveatsReader.Read(claim.Capabilities()[0].Nb())
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}

	if len(lc.Location) < 1 {
		return replica.AllocateOK{}, nil, failure.FromError(fmt.Errorf("location missing from location claim"))
	}

	// TODO maybe more validation here, need to pick a good one from the slice
	whereToSendTransferinvocation := fmt.Sprintf("%s://%s", lc.Location[0].Scheme, lc.Location[0].Host)
	transferFromNodeDID := claim.Issuer().DID()
	if len(claim.Capabilities()) != 1 || claim.Capabilities()[0].Can() != assert.LocationAbility {
		return replica.AllocateOK{}, nil, failure.FromError(fmt.Errorf("TODO: invalid capabilites in location claim"))
	}

	transferInvocation, err := replica.Transfer.Invoke(
		storageService.ID(),
		transferFromNodeDID,
		storageService.ID().DID().GoString(),
		replica.TransferCaveats{
			Blob: cap.Nb().Blob,
			// TODO: Thinking it might make
			Location: cap.Nb().Location,
			Cause:    inv.Link(),
		},
	)
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}

	transferFromNodeURL, err := url.Parse(whereToSendTransferinvocation)
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}

	ch := uhttp.NewHTTPChannel(transferFromNodeURL)
	conn, err := client.NewConnection(transferFromNodeDID, ch)
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}

	client.Execute([]invocation.Invocation{transferInvocation}, conn)

	// fetch the data to replicate
	resp, err := http.Get(lc.Location[0].String())
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}

	log.Info("now uploading to: %s\n", allocateReceipt.Address.URL.String())
	// stream the replica response to PDP service
	req, err := http.NewRequest(http.MethodPut, allocateReceipt.Address.URL.String(), resp.Body)
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}
	req.Header = allocateReceipt.Address.Headers
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		resData, err := io.ReadAll(res.Body)
		if err != nil {
			return replica.AllocateOK{}, nil, failure.FromError(err)
		}
		err = fmt.Errorf("unsuccessful put, status: %s, message: %s", res.Status, string(resData))
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}

	acceptCaveats := blob.AcceptCaveats{
		Space: nil,
		Blob:  nil,
		Put:   nil,
	}
	acceptReceipt, acceptFx, err := BlobAccept(ctx, storageService, acceptCaveats, inv, iCtx)
	if err != nil {
		return replica.AllocateOK{}, nil, failure.FromError(err)
	}

	return replica.AllocateOK{
		Size: allocateReceipt.Size,
		Address: blob.Address{
			URL:     url.URL{},
			Headers: nil,
			Expires: 0,
		},
	}, nil, nil
*/
