package publisher

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/capability/assert"
	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/internal/testutil"
	"github.com/storacha/storage/pkg/metadata"
	"github.com/storacha/storage/pkg/service/publisher/advertisement"
	"github.com/stretchr/testify/require"
)

func TestPublisherService(t *testing.T) {
	signer, err := signer.Generate()
	require.NoError(t, err)

	priv, err := crypto.UnmarshalEd25519PrivateKey(signer.Raw())
	require.NoError(t, err)

	addr, err := multiaddr.NewMultiaddr("/dns4/localhost/tcp/3000/http")
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("publishes location commitments", func(t *testing.T) {
		dstore := dssync.MutexWrap(datastore.NewMapDatastore())
		publisherStore := store.FromDatastore(dstore)

		svc, err := New(priv, publisherStore, addr)
		require.NoError(t, err)

		space := testutil.RandomDID()
		shard := testutil.RandomMultihash()
		location := testutil.Must(url.Parse(fmt.Sprintf("http://localhost:3000/blob/%s", digestutil.Format(shard))))(t)

		claim, err := assert.Location.Delegate(
			signer,
			space,
			signer.DID().String(),
			assert.LocationCaveats{
				Space:    space,
				Content:  shard,
				Location: []url.URL{*location},
			},
			delegation.WithNoExpiration(),
		)
		require.NoError(t, err)

		err = svc.Publish(ctx, claim)
		require.NoError(t, err)

		hd, err := publisherStore.Head(ctx)
		require.NoError(t, err)

		ad, err := publisherStore.Advert(ctx, hd.Head)
		require.NoError(t, err)

		require.Equal(
			t,
			testutil.Must(advertisement.EncodeContextID(space, shard))(t),
			ad.ContextID,
		)

		meta := metadata.MetadataContext.New()
		err = meta.UnmarshalBinary(ad.Metadata)
		require.NoError(t, err)

		protocol := meta.Get(metadata.LocationCommitmentID)
		require.NotNil(t, protocol)

		lcmeta, ok := protocol.(*metadata.LocationCommitmentMetadata)
		require.True(t, ok)

		require.Equal(t, claim.Link(), lcmeta.Claim)

		var ents []multihash.Multihash
		for digest, err := range publisherStore.Entries(ctx, ad.Entries) {
			require.NoError(t, err)
			ents = append(ents, digest)
		}
		require.Len(t, ents, 1)
		require.Equal(t, shard, ents[0])
	})
}
