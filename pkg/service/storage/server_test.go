package storage

import (
	"bytes"
	"context"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/client"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/storage/pkg/capability"
	"github.com/storacha/storage/pkg/capability/assert"
	"github.com/storacha/storage/pkg/capability/blob"
	bdm "github.com/storacha/storage/pkg/capability/blob/datamodel"
	"github.com/storacha/storage/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

var allocateReceiptSchema = []byte(`
	type Result union {
		| AllocateOk "ok"
		| Any "error"
	} representation keyed

	type AllocateOk struct {
		size Int
		address optional Address
	}

	type Address struct {
		url String
		headers {String:String}
		expires Int
	}
`)

var acceptReceiptSchema = []byte(`
	type Result union {
		| AcceptOk "ok"
		| Any "error"
	} representation keyed

	type AcceptOk struct {
		site Link
	}
`)

func TestServer(t *testing.T) {
	svc, err := New(WithIdentity(testutil.Alice), WithLogLevel("*", "info"))
	require.NoError(t, err)
	t.Cleanup(func() { svc.Close() })

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
						ucan.CaveatBuilder(capability.Unit{}),
					),
					ucan.NewCapability(
						blob.AcceptAbility,
						testutil.Alice.DID().String(),
						ucan.CaveatBuilder(capability.Unit{}),
					),
				},
			),
		)(t),
	)

	t.Run("blob/allocate", func(t *testing.T) {
		space := testutil.RandomDID()
		digest := testutil.RandomMultihash()
		size := uint64(rand.IntN(32) + 1)
		cause := testutil.RandomCID()

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

		reader := testutil.Must(receipt.NewReceiptReader[bdm.AllocateOkModel, ipld.Node](allocateReceiptSchema))(t)
		rcpt := testutil.Must(reader.Read(rcptlnk, resp.Blocks()))(t)

		result.MatchResultR0(rcpt.Out(), func(ok bdm.AllocateOkModel) {
			fmt.Printf("%+v\n", ok)
			require.Equal(t, size, uint64(ok.Size))

			allocs, err := svc.Blobs().Allocations().List(context.Background(), digest)
			require.NoError(t, err)

			require.Len(t, allocs, 1)
			require.Equal(t, digest, allocs[0].Blob.Digest)
			require.Equal(t, size, allocs[0].Blob.Size)
			require.Equal(t, space, allocs[0].Space)
			require.Equal(t, inv.Link(), allocs[0].Cause)
		}, func(x ipld.Node) {
			f := testutil.BindFailure(t, x)
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})
	})

	t.Run("blob/accept", func(t *testing.T) {
		space := testutil.RandomDID()
		size := uint64(rand.IntN(32) + 1)
		data := testutil.RandomBytes(int(size))
		digest := testutil.Must(multihash.Sum(data, multihash.SHA2_256, -1))(t)
		cause := testutil.RandomCID()
		exp := uint64(time.Now().Add(time.Second * 30).Unix())

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
			Expires: exp,
			Put: blob.Promise{
				UcanAwait: blob.Result{
					Selector: ".out.ok",
					Link:     testutil.RandomCID(),
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

		reader := testutil.Must(receipt.NewReceiptReader[bdm.AcceptOkModel, ipld.Node](acceptReceiptSchema))(t)
		rcpt := testutil.Must(reader.Read(rcptlnk, resp.Blocks()))(t)

		result.MatchResultR0(rcpt.Out(), func(ok bdm.AcceptOkModel) {
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
			require.Equal(t, digest, nb.Content)
			require.Equal(t, loc.String(), nb.Location[0].String())

			// TODO: assert IPNI advert published
		}, func(x ipld.Node) {
			f := testutil.BindFailure(t, x)
			fmt.Println(f.Message)
			fmt.Println(*f.Stack)
			require.Nil(t, f)
		})
	})
}
