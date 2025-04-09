package storage

import (
	"bytes"
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/replica"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/client"
	"github.com/storacha/go-ucanto/core/car"
	"github.com/storacha/go-ucanto/core/dag/blockstore"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/message"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/core/result/failure"
	fdm "github.com/storacha/go-ucanto/core/result/failure/datamodel"
	"github.com/storacha/go-ucanto/core/result/ok"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"

	"github.com/storacha/storage/pkg/internal/testutil"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	svc, err := New(WithIdentity(testutil.Alice), WithLogLevel("*", "warn"))
	require.NoError(t, err)
	err = svc.Startup()
	require.NoError(t, err)
	t.Cleanup(func() {
		svc.Close(ctx)
	})

	srv, err := NewUCANServer(svc)
	require.NoError(t, err)

	conn := testutil.Must(client.NewConnection(testutil.Service, srv))(t)

	prf := delegation.FromDelegation(
		testutil.Must(
			delegation.Delegate(
				testutil.Alice,
				testutil.Service,
				[]ucan.Capability[ucan.CaveatBuilder]{
					ucan.NewCapability(
						blob.AllocateAbility,
						testutil.Alice.DID().String(),
						ucan.CaveatBuilder(ok.Unit{}),
					),
					ucan.NewCapability(
						blob.AcceptAbility,
						testutil.Alice.DID().String(),
						ucan.CaveatBuilder(ok.Unit{}),
					),
				},
			),
		)(t),
	)

	t.Run("blob/allocate", func(t *testing.T) {
		space := testutil.RandomDID(t)
		digest := testutil.RandomMultihash(t)
		size := uint64(rand.IntN(32) + 1)
		cause := testutil.RandomCID(t)

		nb := blob.AllocateCaveats{
			Space: space,
			Blob: blob.Blob{
				Digest: digest,
				Size:   size,
			},
			Cause: cause,
		}
		cap := blob.Allocate.New(testutil.Alice.DID().String(), nb)
		inv, err := invocation.Invoke(testutil.Service, testutil.Alice, cap, delegation.WithProof(prf))
		require.NoError(t, err)

		resp, err := client.Execute([]invocation.Invocation{inv}, conn)
		require.NoError(t, err)

		// get the receipt link for the invocation from the response
		rcptlnk, ok := resp.Get(inv.Link())
		require.True(t, ok, "missing receipt for invocation: %s", inv.Link())

		reader := testutil.Must(receipt.NewReceiptReaderFromTypes[blob.AllocateOk, fdm.FailureModel](blob.AllocateOkType(), fdm.FailureType(), types.Converters...))(t)
		rcpt := testutil.Must(reader.Read(rcptlnk, resp.Blocks()))(t)

		result.MatchResultR0(rcpt.Out(), func(ok blob.AllocateOk) {
			fmt.Printf("%+v\n", ok)
			require.Equal(t, size, uint64(ok.Size))

			allocs, err := svc.Blobs().Allocations().List(context.Background(), digest)
			require.NoError(t, err)

			require.Len(t, allocs, 1)
			require.Equal(t, digest, allocs[0].Blob.Digest)
			require.Equal(t, size, allocs[0].Blob.Size)
			require.Equal(t, space, allocs[0].Space)
			require.Equal(t, inv.Link(), allocs[0].Cause)
		}, func(f fdm.FailureModel) {
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})
	})

	t.Run("repeat blob/allocate for same blob", func(t *testing.T) {
		space := testutil.RandomDID(t)
		size := uint64(rand.IntN(32) + 1)
		data := testutil.RandomBytes(t, int(size))
		digest := testutil.Must(multihash.Sum(data, multihash.SHA2_256, -1))(t)
		cause := testutil.RandomCID(t)

		nb := blob.AllocateCaveats{
			Space: space,
			Blob: blob.Blob{
				Digest: digest,
				Size:   size,
			},
			Cause: cause,
		}
		cap := blob.Allocate.New(testutil.Alice.DID().String(), nb)

		invokeBlobAllocate := func() result.Result[blob.AllocateOk, fdm.FailureModel] {
			inv, err := invocation.Invoke(testutil.Service, testutil.Alice, cap, delegation.WithProof(prf))
			require.NoError(t, err)

			resp, err := client.Execute([]invocation.Invocation{inv}, conn)
			require.NoError(t, err)

			rcptlnk, ok := resp.Get(inv.Link())
			require.True(t, ok, "missing receipt for invocation: %s", inv.Link())

			reader := testutil.Must(receipt.NewReceiptReaderFromTypes[blob.AllocateOk, fdm.FailureModel](blob.AllocateOkType(), fdm.FailureType(), types.Converters...))(t)
			rcpt := testutil.Must(reader.Read(rcptlnk, resp.Blocks()))(t)
			return rcpt.Out()
		}

		result.MatchResultR0(invokeBlobAllocate(), func(ok blob.AllocateOk) {
			fmt.Printf("%+v\n", ok)
			require.Equal(t, size, uint64(ok.Size))
			require.NotNil(t, ok.Address)
		}, func(f fdm.FailureModel) {
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})

		// now again without upload
		result.MatchResultR0(invokeBlobAllocate(), func(ok blob.AllocateOk) {
			fmt.Printf("%+v\n", ok)
			require.Equal(t, uint64(0), ok.Size)
			require.NotNil(t, ok.Address)
		}, func(f fdm.FailureModel) {
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})

		// simulate a blob upload
		err = svc.Blobs().Store().Put(context.Background(), digest, size, bytes.NewReader(data))
		require.NoError(t, err)

		// now again after upload
		result.MatchResultR0(invokeBlobAllocate(), func(ok blob.AllocateOk) {
			fmt.Printf("%+v\n", ok)
			require.Equal(t, uint64(0), ok.Size)
			require.Nil(t, ok.Address)
		}, func(f fdm.FailureModel) {
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})
	})

	t.Run("repeat blob/allocate for same blob in different space", func(t *testing.T) {
		space0 := testutil.RandomDID(t)
		space1 := testutil.RandomDID(t)
		size := uint64(rand.IntN(32) + 1)
		data := testutil.RandomBytes(t, int(size))
		digest := testutil.Must(multihash.Sum(data, multihash.SHA2_256, -1))(t)
		cause := testutil.RandomCID(t)

		invokeBlobAllocate := func(space did.DID) result.Result[blob.AllocateOk, fdm.FailureModel] {
			nb := blob.AllocateCaveats{
				Space: space,
				Blob: blob.Blob{
					Digest: digest,
					Size:   size,
				},
				Cause: cause,
			}
			cap := blob.Allocate.New(testutil.Alice.DID().String(), nb)

			inv, err := invocation.Invoke(testutil.Service, testutil.Alice, cap, delegation.WithProof(prf))
			require.NoError(t, err)

			resp, err := client.Execute([]invocation.Invocation{inv}, conn)
			require.NoError(t, err)

			rcptlnk, ok := resp.Get(inv.Link())
			require.True(t, ok, "missing receipt for invocation: %s", inv.Link())

			reader := testutil.Must(receipt.NewReceiptReaderFromTypes[blob.AllocateOk, fdm.FailureModel](blob.AllocateOkType(), fdm.FailureType(), types.Converters...))(t)
			rcpt := testutil.Must(reader.Read(rcptlnk, resp.Blocks()))(t)
			return rcpt.Out()
		}

		result.MatchResultR0(invokeBlobAllocate(space0), func(ok blob.AllocateOk) {
			fmt.Printf("%+v\n", ok)
			require.Equal(t, size, uint64(ok.Size))
			require.NotNil(t, ok.Address)
		}, func(f fdm.FailureModel) {
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})

		// simulate a blob upload
		err = svc.Blobs().Store().Put(context.Background(), digest, size, bytes.NewReader(data))
		require.NoError(t, err)

		// now again after upload, but in different space
		result.MatchResultR0(invokeBlobAllocate(space1), func(ok blob.AllocateOk) {
			fmt.Printf("%+v\n", ok)
			require.Equal(t, size, uint64(ok.Size))
			require.Nil(t, ok.Address)
		}, func(f fdm.FailureModel) {
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})
	})

	t.Run("blob/accept", func(t *testing.T) {
		space := testutil.RandomDID(t)
		size := uint64(rand.IntN(32) + 1)
		data := testutil.RandomBytes(t, int(size))
		digest := testutil.Must(multihash.Sum(data, multihash.SHA2_256, -1))(t)
		cause := testutil.RandomCID(t)

		allocNb := blob.AllocateCaveats{
			Space: space,
			Blob: blob.Blob{
				Digest: digest,
				Size:   size,
			},
			Cause: cause,
		}
		allocCap := blob.Allocate.New(testutil.Alice.DID().String(), allocNb)
		allocInv, err := invocation.Invoke(testutil.Service, testutil.Alice, allocCap, delegation.WithProof(prf))
		require.NoError(t, err)

		_, err = client.Execute([]invocation.Invocation{allocInv}, conn)
		require.NoError(t, err)

		// simulate a blob upload
		err = svc.Blobs().Store().Put(context.Background(), digest, size, bytes.NewReader(data))
		require.NoError(t, err)
		// get the expected download URL
		loc, err := svc.Blobs().Access().GetDownloadURL(digest)
		require.NoError(t, err)

		// eventually service will invoke blob/accept
		acceptNb := blob.AcceptCaveats{
			Space: space,
			Blob: blob.Blob{
				Digest: digest,
				Size:   size,
			},
			Put: blob.Promise{
				UcanAwait: blob.Await{
					Selector: ".out.ok",
					Link:     testutil.RandomCID(t),
				},
			},
		}
		// fmt.Println(printer.Sprint(testutil.Must(acceptNb.ToIPLD())(t)))
		acceptCap := blob.Accept.New(testutil.Alice.DID().String(), acceptNb)
		acceptInv, err := invocation.Invoke(testutil.Service, testutil.Alice, acceptCap, delegation.WithProof(prf))
		require.NoError(t, err)

		resp, err := client.Execute([]invocation.Invocation{acceptInv}, conn)
		require.NoError(t, err)

		// get the receipt link for the invocation from the response
		rcptlnk, ok := resp.Get(acceptInv.Link())
		require.True(t, ok, "missing receipt for invocation: %s", acceptInv.Link())

		reader := testutil.Must(receipt.NewReceiptReaderFromTypes[blob.AcceptOk, fdm.FailureModel](blob.AcceptOkType(), fdm.FailureType(), types.Converters...))(t)
		rcpt := testutil.Must(reader.Read(rcptlnk, resp.Blocks()))(t)

		result.MatchResultR0(rcpt.Out(), func(ok blob.AcceptOk) {
			fmt.Printf("%+v\n", ok)

			claim, err := svc.Claims().Store().Get(context.Background(), ok.Site)
			require.NoError(t, err)

			require.Equal(t, testutil.Alice.DID(), claim.Issuer())
			require.Equal(t, space, claim.Audience().DID())
			require.Equal(t, assert.LocationAbility, claim.Capabilities()[0].Can())
			require.Equal(t, testutil.Alice.DID().String(), claim.Capabilities()[0].With())

			nb, err := assert.LocationCaveatsReader.Read(claim.Capabilities()[0].Nb())
			require.NoError(t, err)

			require.Equal(t, space, nb.Space)
			require.Equal(t, digest, nb.Content.Hash())
			require.Equal(t, loc.String(), nb.Location[0].String())

			// TODO: assert IPNI advert published
		}, func(f fdm.FailureModel) {
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})

		require.NotEmpty(t, rcpt.Fx().Fork())
		effect := rcpt.Fx().Fork()[0]
		claim, ok := effect.Invocation()
		require.True(t, ok)
		require.Equal(t, assert.LocationAbility, claim.Capabilities()[0].Can())
	})
}

// This test verifies that the UCAN server correctly constructs, signs, and executes the replica
// allocation process, and that the simulated endpoints correctly emulate the expected interactions.
// It simulates an HTTP Server with the following properties:
//    - A lightweight HTTP server is spun up on port 8080 to simulate external endpoints.
//    - The `/get` endpoint simulates the source node by returning the original blob data: The node data is being replicated from
//    - The `/put` endpoint fakes the replica node, accepting data and storing it via the service. Essentially, "this" node.
//    - The `/upload-service` endpoint emulates the upload service by decoding a CAR payload and
//      generating a transfer receipt message, mimicking post-upload processing.
//
func TestReplicaAllocateTransfer(t *testing.T) {
	// Test setup parameters.
	expectedSpace := testutil.RandomDID(t)
	expectedSize := uint64(rand.IntN(32) + 1)
	expectedData := testutil.RandomBytes(t, int(expectedSize))
	expectedDigest := testutil.Must(multihash.Sum(expectedData, multihash.SHA2_256, -1))(t)
	replicas := 8
	serverAddr := ":8080"
	sourcePath, sinkPath, uploadServicePath := "get", "put", "upload-service"

	// Helper to create URLs.
	makeURL := func(path string) *url.URL {
		return testutil.Must(url.Parse(fmt.Sprintf("http://127.0.0.1%s/%s", serverAddr, path)))(t)
	}
	locationURL := makeURL(sourcePath)
	uploadServiceURL := makeURL(uploadServicePath)
	presignedURL := makeURL(sinkPath)
	fakeBlobPresigner := &FakePresigned{uploadURL: *presignedURL}

	// Set up service.
	svc, err := New(
		WithIdentity(testutil.Alice),
		WithLogLevel("*", "warn"),
		WithBlobsPresigner(fakeBlobPresigner),
		WithUploadServiceConfig(testutil.Alice, *uploadServiceURL),
	)
	require.NoError(t, err)
	require.NoError(t, svc.Startup())

	// Create a cancellable context and start the fake HTTP server.
	// If this context times out before the final assertion, we fail the test.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fakeServer, transferOkChan := startTestHTTPServer(ctx, t, expectedDigest, expectedData, svc, serverAddr, sourcePath, sinkPath, uploadServicePath)
	t.Cleanup(func() {
		fakeServer.Close()
		svc.Close(ctx)
	})

	srv, err := NewUCANServer(svc)
	require.NoError(t, err)
	conn := testutil.Must(client.NewConnection(testutil.Service, srv))(t)

	// Build UCAN delegation for required capabilities.
	caps := []ucan.Capability[ucan.CaveatBuilder]{
		ucan.NewCapability(replica.AllocateAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
		// these are required to fulfill the replica Allocate Ability.
		ucan.NewCapability(blob.AllocateAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
		ucan.NewCapability(blob.AcceptAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
	}
	prf := delegation.FromDelegation(testutil.Must(delegation.Delegate(testutil.Alice, testutil.Service, caps))(t))

	// A location commitment indicating where the blob MUST be fetched from.
	// The locationURL points at our TestHTTPServer
	expectedLocationClaimCaveats := assert.LocationCaveats{
		Space:    expectedSpace,
		Content:  types.FromHash(expectedDigest),
		Location: []url.URL{*locationURL},
		Range: &assert.Range{
			Offset: 1,
			Length: &expectedSize,
		},
	}
	lcd, err := assert.Location.Delegate(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedLocationClaimCaveats,
		delegation.WithProof(prf),
	)
	require.NoError(t, err)

	// Invoke blob replication.
	expectedReplicaCaveats := blob.ReplicateCaveats{
		Blob: blob.Blob{
			Digest: expectedDigest,
			Size:   expectedSize,
		},
		Replicas: replicas,
		Location: lcd.Root().Link(),
	}
	bri, err := blob.Replicate.Invoke(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedReplicaCaveats,
	)
	require.NoError(t, err)
	// attach the location claim to the blob replicate invocation
	for block, err := range lcd.Blocks() {
		require.NoError(t, err)
		require.NoError(t, bri.Attach(block))
	}

	// Invoke replica allocation - what we are testing(!!!)
	expectedAllocateCaveats := replica.AllocateCaveats{
		Space:    expectedSpace,
		Blob:     blob.Blob{Digest: expectedDigest, Size: expectedSize},
		Location: lcd.Root().Link(),
		Cause:    bri.Root().Link(),
	}
	rbi, err := replica.Allocate.Invoke(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedAllocateCaveats,
	)
	require.NoError(t, err)
	// now attach the blob replicate invocation, and its corresponding location claim
	for block, err := range bri.Blocks() {
		require.NoError(t, err)
		require.NoError(t, rbi.Attach(block))
	}

	// Execute invocation
	res, err := client.Execute([]invocation.Invocation{rbi}, conn)
	require.NoError(t, err)

	// assert the size of the allocation matches our expected size.
	reader, err := receipt.NewReceiptReaderFromTypes[replica.AllocateOk, fdm.FailureModel](
		replica.AllocateOkType(), fdm.FailureType(), types.Converters...,
	)
	require.NoError(t, err)
	rcptLink, ok := res.Get(rbi.Link())
	require.True(t, ok)
	rcpt, err := reader.Read(rcptLink, res.Blocks())
	require.NoError(t, err)
	alloc, err := result.Unwrap(result.MapError(rcpt.Out(), failure.FromFailureModel))
	require.NoError(t, err)
	require.Equal(t, expectedSize, alloc.Size)

	// Wait for transfer receipt message, we wait at most 10 seconds (context timeout), or fail.
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err(), "test did not produce transfer receipt in time")
	case transferOkMsg := <-transferOkChan:
		// sanity
		require.NotNil(t, transferOkMsg)

		// expect one invocation and one receipt
		require.Len(t, transferOkMsg.Invocations(), 1)
		require.Len(t, transferOkMsg.Receipts(), 1)

		transferInvocationCid := testutil.Must(cid.Parse(transferOkMsg.Invocations()[0].String()))(t)
		reader := testutil.Must(blockstore.NewBlockReader(blockstore.WithBlocksIterator(transferOkMsg.Blocks())))(t)

		// read the transfer invocation
		transferCav := mustGetInvocationCaveats[replica.TransferCaveats](t, reader, cidlink.Link{Cid: transferInvocationCid}, replica.TransferCaveatsReader.Read)
		// assert on transfer fields
		require.Equal(t, expectedSize, transferCav.Blob.Size)
		require.Equal(t, expectedDigest, transferCav.Blob.Digest)
		require.Equal(t, expectedSpace, transferCav.Space)

		// transfer location is the initial location from blob replicate request
		locationCav := mustGetInvocationCaveats[assert.LocationCaveats](t, reader, transferCav.Location, assert.LocationCaveatsReader.Read)
		require.Equal(t, expectedLocationClaimCaveats, locationCav)

		// transfer cause is the replica allocate cause
		replicaAllocateCav := mustGetInvocationCaveats[replica.AllocateCaveats](t, reader, transferCav.Cause, replica.AllocateCaveatsReader.Read)
		require.Equal(t, expectedAllocateCaveats, replicaAllocateCav)

		// replica allocate caused by blob replicate
		blobReplicateInv := mustGetInvocationCaveats[blob.ReplicateCaveats](t, reader, replicaAllocateCav.Cause, blob.ReplicateCaveatsReader.Read)
		require.Equal(t, expectedReplicaCaveats, blobReplicateInv)

		// read the receipt of the transfer invocation asserting the location caveats of Site contain expected values.
		transferReceiptReader := testutil.Must(receipt.NewReceiptReaderFromTypes[replica.TransferOk, fdm.FailureModel](replica.TransferOkType(), fdm.FailureType(), types.Converters...))(t)
		transferReceiptCid := testutil.Must(cid.Parse(transferOkMsg.Receipts()[0].String()))(t)
		transferReceipt := testutil.Must(transferReceiptReader.Read(cidlink.Link{Cid: transferReceiptCid}, reader.Iterator()))(t)
		transferOk := testutil.Must(result.Unwrap(result.MapError(transferReceipt.Out(), failure.FromFailureModel)))(t)
		require.Nil(t, transferOk.PDP)
		locationCavRct := mustGetInvocationCaveats[assert.LocationCaveats](t, reader, transferOk.Site, assert.LocationCaveatsReader.Read)
		require.Equal(t, expectedSpace, locationCavRct.Space)
		require.Equal(t, expectedDigest, locationCavRct.Content.Hash())
		require.Len(t, locationCavRct.Location, 1)
		require.Equal(t, fmt.Sprintf("/blob/z%s", expectedDigest.B58String()), locationCavRct.Location[0].Path)

	}
}

func TestReplicaAllocateTransferWithExistingAllocation(t *testing.T) {
	// Test setup parameters.
	expectedSpace := testutil.RandomDID(t)
	expectedSize := uint64(rand.IntN(32) + 1)
	expectedTransferSize := uint64(0)
	expectedData := testutil.RandomBytes(t, int(expectedSize))
	expectedDigest := testutil.Must(multihash.Sum(expectedData, multihash.SHA2_256, -1))(t)
	replicas := 8
	serverAddr := ":8080"
	sourcePath, sinkPath, uploadServicePath := "get", "put", "upload-service"

	// Helper to create URLs.
	makeURL := func(path string) *url.URL {
		return testutil.Must(url.Parse(fmt.Sprintf("http://127.0.0.1%s/%s", serverAddr, path)))(t)
	}
	locationURL := makeURL(sourcePath)
	uploadServiceURL := makeURL(uploadServicePath)
	presignedURL := makeURL(sinkPath)
	fakeBlobPresigner := &FakePresigned{uploadURL: *presignedURL}

	// Set up service.
	svc, err := New(
		WithIdentity(testutil.Alice),
		WithLogLevel("*", "warn"),
		WithBlobsPresigner(fakeBlobPresigner),
		WithUploadServiceConfig(testutil.Alice, *uploadServiceURL),
	)
	require.NoError(t, err)
	require.NoError(t, svc.Startup())

	// Create a cancellable context and start the fake HTTP server.
	// If this context times out before the final assertion, we fail the test.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fakeServer, transferOkChan := startTestHTTPServer(ctx, t, expectedDigest, expectedData, svc, serverAddr, sourcePath, sinkPath, uploadServicePath)
	t.Cleanup(func() {
		fakeServer.Close()
		svc.Close(ctx)
	})

	srv, err := NewUCANServer(svc)
	require.NoError(t, err)
	conn := testutil.Must(client.NewConnection(testutil.Service, srv))(t)

	// Build UCAN delegation for required capabilities.
	caps := []ucan.Capability[ucan.CaveatBuilder]{
		ucan.NewCapability(replica.AllocateAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
		// these are required to fulfill the replica Allocate Ability.
		ucan.NewCapability(blob.AllocateAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
		ucan.NewCapability(blob.AcceptAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
	}
	prf := delegation.FromDelegation(testutil.Must(delegation.Delegate(testutil.Alice, testutil.Service, caps))(t))

	// A location commitment indicating where the blob MUST be fetched from.
	// The locationURL points at our TestHTTPServer
	expectedLocationClaimCaveats := assert.LocationCaveats{
		Space:    expectedSpace,
		Content:  types.FromHash(expectedDigest),
		Location: []url.URL{*locationURL},
		Range: &assert.Range{
			Offset: 1,
			Length: &expectedSize,
		},
	}
	lcd, err := assert.Location.Delegate(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedLocationClaimCaveats,
		delegation.WithProof(prf),
	)
	require.NoError(t, err)

	// Invoke blob replication.
	expectedReplicaCaveats := blob.ReplicateCaveats{
		Blob: blob.Blob{
			Digest: expectedDigest,
			Size:   expectedSize,
		},
		Replicas: replicas,
		Location: lcd.Root().Link(),
	}
	bri, err := blob.Replicate.Invoke(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedReplicaCaveats,
	)
	require.NoError(t, err)
	// attach the location claim to the blob replicate invocation
	for block, err := range lcd.Blocks() {
		require.NoError(t, err)
		require.NoError(t, bri.Attach(block))
	}

	// Add an allocation for this data indicating space has already been allocated
	// for it, but a transfer is still required.
	err = svc.Blobs().Allocations().Put(ctx, allocation.Allocation{
		Space: expectedSpace,
		Blob: allocation.Blob{
			Digest: expectedDigest,
			Size:   expectedSize,
		},
		Expires: uint64(time.Now().Add(time.Hour).UTC().Unix()),
		Cause:   bri.Link(),
	})
	require.NoError(t, err)

	// Invoke replica allocation - what we are testing(!!!)
	expectedAllocateCaveats := replica.AllocateCaveats{
		Space:    expectedSpace,
		Blob:     blob.Blob{Digest: expectedDigest, Size: expectedSize},
		Location: lcd.Root().Link(),
		Cause:    bri.Root().Link(),
	}
	rbi, err := replica.Allocate.Invoke(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedAllocateCaveats,
	)
	require.NoError(t, err)
	// now attach the blob replicate invocation, and its corresponding location claim
	for block, err := range bri.Blocks() {
		require.NoError(t, err)
		require.NoError(t, rbi.Attach(block))
	}

	// Execute invocation
	res, err := client.Execute([]invocation.Invocation{rbi}, conn)
	require.NoError(t, err)

	// assert the size of the allocation matches our expected size.
	reader, err := receipt.NewReceiptReaderFromTypes[replica.AllocateOk, fdm.FailureModel](
		replica.AllocateOkType(), fdm.FailureType(), types.Converters...,
	)
	require.NoError(t, err)
	rcptLink, ok := res.Get(rbi.Link())
	require.True(t, ok)
	rcpt, err := reader.Read(rcptLink, res.Blocks())
	require.NoError(t, err)
	alloc, err := result.Unwrap(result.MapError(rcpt.Out(), failure.FromFailureModel))
	require.NoError(t, err)
	require.EqualValues(t, expectedTransferSize, alloc.Size)

	// Wait for transfer receipt message, we wait at most 10 seconds (context timeout), or fail.
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err(), "test did not produce transfer receipt in time")
	case transferOkMsg := <-transferOkChan:
		// sanity
		require.NotNil(t, transferOkMsg)

		// expect one invocation and one receipt
		require.Len(t, transferOkMsg.Invocations(), 1)
		require.Len(t, transferOkMsg.Receipts(), 1)

		transferInvocationCid := testutil.Must(cid.Parse(transferOkMsg.Invocations()[0].String()))(t)
		reader := testutil.Must(blockstore.NewBlockReader(blockstore.WithBlocksIterator(transferOkMsg.Blocks())))(t)

		// read the transfer invocation
		transferCav := mustGetInvocationCaveats[replica.TransferCaveats](t, reader, cidlink.Link{Cid: transferInvocationCid}, replica.TransferCaveatsReader.Read)
		// assert on transfer fields
		require.EqualValues(t, expectedTransferSize, transferCav.Blob.Size)
		require.Equal(t, expectedDigest, transferCav.Blob.Digest)
		require.Equal(t, expectedSpace, transferCav.Space)

		// transfer location is the initial location from blob replicate request
		locationCav := mustGetInvocationCaveats[assert.LocationCaveats](t, reader, transferCav.Location, assert.LocationCaveatsReader.Read)
		require.Equal(t, expectedLocationClaimCaveats, locationCav)

		// transfer cause is the replica allocate cause
		replicaAllocateCav := mustGetInvocationCaveats[replica.AllocateCaveats](t, reader, transferCav.Cause, replica.AllocateCaveatsReader.Read)
		require.Equal(t, expectedAllocateCaveats, replicaAllocateCav)

		// replica allocate caused by blob replicate
		blobReplicateInv := mustGetInvocationCaveats[blob.ReplicateCaveats](t, reader, replicaAllocateCav.Cause, blob.ReplicateCaveatsReader.Read)
		require.Equal(t, expectedReplicaCaveats, blobReplicateInv)

		// read the receipt of the transfer invocation asserting the location caveats of Site contain expected values.
		transferReceiptReader := testutil.Must(receipt.NewReceiptReaderFromTypes[replica.TransferOk, fdm.FailureModel](replica.TransferOkType(), fdm.FailureType(), types.Converters...))(t)
		transferReceiptCid := testutil.Must(cid.Parse(transferOkMsg.Receipts()[0].String()))(t)
		transferReceipt := testutil.Must(transferReceiptReader.Read(cidlink.Link{Cid: transferReceiptCid}, reader.Iterator()))(t)
		transferOk := testutil.Must(result.Unwrap(result.MapError(transferReceipt.Out(), failure.FromFailureModel)))(t)
		require.Nil(t, transferOk.PDP)
		locationCavRct := mustGetInvocationCaveats[assert.LocationCaveats](t, reader, transferOk.Site, assert.LocationCaveatsReader.Read)
		require.Equal(t, expectedSpace, locationCavRct.Space)
		require.Equal(t, expectedDigest, locationCavRct.Content.Hash())
		require.Len(t, locationCavRct.Location, 1)
		require.Equal(t, fmt.Sprintf("/blob/z%s", expectedDigest.B58String()), locationCavRct.Location[0].Path)

	}
}

func TestReplicaAllocateWithExistingAllocationAndReceieved(t *testing.T) {
	// Test setup parameters.
	expectedSpace := testutil.RandomDID(t)
	expectedSize := uint64(rand.IntN(32) + 1)
	expectedTransferSize := uint64(0)
	expectedData := testutil.RandomBytes(t, int(expectedSize))
	expectedDigest := testutil.Must(multihash.Sum(expectedData, multihash.SHA2_256, -1))(t)
	replicas := 8
	serverAddr := ":8080"
	sourcePath, sinkPath, uploadServicePath := "get", "put", "upload-service"

	// Helper to create URLs.
	makeURL := func(path string) *url.URL {
		return testutil.Must(url.Parse(fmt.Sprintf("http://127.0.0.1%s/%s", serverAddr, path)))(t)
	}
	locationURL := makeURL(sourcePath)
	uploadServiceURL := makeURL(uploadServicePath)
	presignedURL := makeURL(sinkPath)
	fakeBlobPresigner := &FakePresigned{uploadURL: *presignedURL}

	// Set up service.
	svc, err := New(
		WithIdentity(testutil.Alice),
		WithLogLevel("*", "warn"),
		WithBlobsPresigner(fakeBlobPresigner),
		WithUploadServiceConfig(testutil.Alice, *uploadServiceURL),
	)
	require.NoError(t, err)
	require.NoError(t, svc.Startup())

	// Create a cancellable context and start the fake HTTP server.
	// If this context times out before the final assertion, we fail the test.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Hour)
	defer cancel()
	fakeServer, transferOkChan := startTestHTTPServer(ctx, t, expectedDigest, expectedData, svc, serverAddr, sourcePath, sinkPath, uploadServicePath)
	t.Cleanup(func() {
		fakeServer.Close()
		svc.Close(ctx)
	})

	srv, err := NewUCANServer(svc)
	require.NoError(t, err)
	conn := testutil.Must(client.NewConnection(testutil.Service, srv))(t)

	// Build UCAN delegation for required capabilities.
	caps := []ucan.Capability[ucan.CaveatBuilder]{
		ucan.NewCapability(replica.AllocateAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
		// these are required to fulfill the replica Allocate Ability.
		ucan.NewCapability(blob.AllocateAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
		ucan.NewCapability(blob.AcceptAbility, testutil.Alice.DID().String(), ucan.CaveatBuilder(ok.Unit{})),
	}
	prf := delegation.FromDelegation(testutil.Must(delegation.Delegate(testutil.Alice, testutil.Service, caps))(t))

	// A location commitment indicating where the blob MUST be fetched from.
	// The locationURL points at our TestHTTPServer
	expectedLocationClaimCaveats := assert.LocationCaveats{
		Space:    expectedSpace,
		Content:  types.FromHash(expectedDigest),
		Location: []url.URL{*locationURL},
		Range: &assert.Range{
			Offset: 1,
			Length: &expectedSize,
		},
	}
	lcd, err := assert.Location.Delegate(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedLocationClaimCaveats,
		delegation.WithProof(prf),
	)
	require.NoError(t, err)

	// Invoke blob replication.
	expectedReplicaCaveats := blob.ReplicateCaveats{
		Blob: blob.Blob{
			Digest: expectedDigest,
			Size:   expectedSize,
		},
		Replicas: replicas,
		Location: lcd.Root().Link(),
	}
	bri, err := blob.Replicate.Invoke(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedReplicaCaveats,
	)
	require.NoError(t, err)
	// attach the location claim to the blob replicate invocation
	for block, err := range lcd.Blocks() {
		require.NoError(t, err)
		require.NoError(t, bri.Attach(block))
	}

	// Add an allocation for this data indicating space has already been allocated
	// for it, but a transfer is still required.
	err = svc.Blobs().Allocations().Put(ctx, allocation.Allocation{
		Space: expectedSpace,
		Blob: allocation.Blob{
			Digest: expectedDigest,
			Size:   expectedSize,
		},
		Expires: uint64(time.Now().Add(time.Hour).UTC().Unix()),
		Cause:   bri.Link(),
	})
	require.NoError(t, err)
	// additionally, include the piece in the store to indicate it's been receieved
	err = svc.blobs.Store().Put(ctx, expectedDigest, expectedSize, bytes.NewReader(expectedData))
	require.NoError(t, err)

	// Invoke replica allocation - what we are testing(!!!)
	expectedAllocateCaveats := replica.AllocateCaveats{
		Space:    expectedSpace,
		Blob:     blob.Blob{Digest: expectedDigest, Size: expectedSize},
		Location: lcd.Root().Link(),
		Cause:    bri.Root().Link(),
	}
	rbi, err := replica.Allocate.Invoke(
		testutil.Alice,
		testutil.Alice.DID(),
		testutil.Alice.DID().String(),
		expectedAllocateCaveats,
	)
	require.NoError(t, err)
	// now attach the blob replicate invocation, and its corresponding location claim
	for block, err := range bri.Blocks() {
		require.NoError(t, err)
		require.NoError(t, rbi.Attach(block))
	}

	// Execute invocation
	res, err := client.Execute([]invocation.Invocation{rbi}, conn)
	require.NoError(t, err)

	// assert the size of the allocation matches our expected size.
	reader, err := receipt.NewReceiptReaderFromTypes[replica.AllocateOk, fdm.FailureModel](
		replica.AllocateOkType(), fdm.FailureType(), types.Converters...,
	)
	require.NoError(t, err)
	rcptLink, ok := res.Get(rbi.Link())
	require.True(t, ok)
	rcpt, err := reader.Read(rcptLink, res.Blocks())
	require.NoError(t, err)
	alloc, err := result.Unwrap(result.MapError(rcpt.Out(), failure.FromFailureModel))
	require.NoError(t, err)
	require.EqualValues(t, expectedTransferSize, alloc.Size)

	// Wait for transfer receipt message, we wait at most 10 seconds (context timeout), or fail.
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err(), "test did not produce transfer receipt in time")
	case transferOkMsg := <-transferOkChan:
		// sanity
		require.NotNil(t, transferOkMsg)

		// expect one invocation and one receipt
		require.Len(t, transferOkMsg.Invocations(), 1)
		require.Len(t, transferOkMsg.Receipts(), 1)

		transferInvocationCid := testutil.Must(cid.Parse(transferOkMsg.Invocations()[0].String()))(t)
		reader := testutil.Must(blockstore.NewBlockReader(blockstore.WithBlocksIterator(transferOkMsg.Blocks())))(t)

		// read the transfer invocation
		transferCav := mustGetInvocationCaveats[replica.TransferCaveats](t, reader, cidlink.Link{Cid: transferInvocationCid}, replica.TransferCaveatsReader.Read)
		// assert on transfer fields
		require.EqualValues(t, expectedTransferSize, transferCav.Blob.Size)
		require.Equal(t, expectedDigest, transferCav.Blob.Digest)
		require.Equal(t, expectedSpace, transferCav.Space)

		// transfer location is the initial location from blob replicate request
		locationCav := mustGetInvocationCaveats[assert.LocationCaveats](t, reader, transferCav.Location, assert.LocationCaveatsReader.Read)
		require.Equal(t, expectedLocationClaimCaveats, locationCav)

		// transfer cause is the replica allocate cause
		replicaAllocateCav := mustGetInvocationCaveats[replica.AllocateCaveats](t, reader, transferCav.Cause, replica.AllocateCaveatsReader.Read)
		require.Equal(t, expectedAllocateCaveats, replicaAllocateCav)

		// replica allocate caused by blob replicate
		blobReplicateInv := mustGetInvocationCaveats[blob.ReplicateCaveats](t, reader, replicaAllocateCav.Cause, blob.ReplicateCaveatsReader.Read)
		require.Equal(t, expectedReplicaCaveats, blobReplicateInv)

		// read the receipt of the transfer invocation asserting the location caveats of Site contain expected values.
		transferReceiptReader := testutil.Must(receipt.NewReceiptReaderFromTypes[replica.TransferOk, fdm.FailureModel](replica.TransferOkType(), fdm.FailureType(), types.Converters...))(t)
		transferReceiptCid := testutil.Must(cid.Parse(transferOkMsg.Receipts()[0].String()))(t)
		transferReceipt := testutil.Must(transferReceiptReader.Read(cidlink.Link{Cid: transferReceiptCid}, reader.Iterator()))(t)
		transferOk := testutil.Must(result.Unwrap(result.MapError(transferReceipt.Out(), failure.FromFailureModel)))(t)
		require.Nil(t, transferOk.PDP)
		locationCavRct := mustGetInvocationCaveats[assert.LocationCaveats](t, reader, transferOk.Site, assert.LocationCaveatsReader.Read)
		require.Equal(t, expectedSpace, locationCavRct.Space)
		require.Equal(t, expectedDigest, locationCavRct.Content.Hash())
		require.Len(t, locationCavRct.Location, 1)
		require.Equal(t, fmt.Sprintf("/blob/z%s", expectedDigest.B58String()), locationCavRct.Location[0].Path)

	}
}

func mustGetInvocationCaveats[T ipld.Builder](t *testing.T, reader blockstore.BlockReader, inv ucan.Link, invReader func(any) (T, failure.Failure)) T {
	view := testutil.Must(invocation.NewInvocationView(inv, reader))(t)
	invc := testutil.Must(invReader(view.Capabilities()[0].Nb()))(t)
	return invc
}

// startTestHTTPServer starts a simple HTTP server with configurable endpoints.
func startTestHTTPServer(
	ctx context.Context,
	t *testing.T,
	digest multihash.Multihash,
	serveData []byte,
	svc Service,
	addr, sourcePath, sinkPath, uploadServicePath string,
) (*http.Server, <-chan message.AgentMessage) {
	agentCh := make(chan message.AgentMessage, 1)
	mux := http.NewServeMux()

	// Endpoint to serve data.
	mux.HandleFunc(fmt.Sprintf("/%s", sourcePath), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(serveData)
	})
	// Endpoint to store data on the replica.
	mux.HandleFunc(fmt.Sprintf("/%s", sinkPath), func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, svc.Blobs().Store().Put(ctx, digest, uint64(len(serveData)), bytes.NewReader(serveData)))
		_, _ = w.Write(serveData)
	})
	// Endpoint to simulate the upload service.
	mux.HandleFunc(fmt.Sprintf("/%s", uploadServicePath), func(w http.ResponseWriter, r *http.Request) {
		roots, blocks, err := car.Decode(r.Body)
		require.NoError(t, err)
		bstore, err := blockstore.NewBlockReader(blockstore.WithBlocksIterator(blocks))
		require.NoError(t, err)
		agentMessage, err := message.NewMessage(roots, bstore)
		require.NoError(t, err)
		agentCh <- agentMessage
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatalf("HTTP server ListenAndServe failed: %v", err)
		}
	}()
	time.Sleep(50 * time.Millisecond)
	t.Cleanup(func() {
		require.NoError(t, server.Close())
	})
	return server, agentCh
}

// FakePresigned is a stub for upload URL presigning.
type FakePresigned struct {
	uploadURL url.URL
}

func (f *FakePresigned) SignUploadURL(ctx context.Context, digest multihash.Multihash, size, ttl uint64) (url.URL, http.Header, error) {
	return f.uploadURL, nil, nil
}

func (f *FakePresigned) VerifyUploadURL(ctx context.Context, url url.URL, headers http.Header) (url.URL, http.Header, error) {
	// TODO: implement when needed.
	panic("implement me")
}
