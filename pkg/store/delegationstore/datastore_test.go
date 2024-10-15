package delegationstore

import (
	"context"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/storage/pkg/capability"
	"github.com/storacha/storage/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestDsDelegationStore(t *testing.T) {
	t.Run("roundtrip", func(t *testing.T) {
		store, err := NewDsDelegationStore(datastore.NewMapDatastore())
		require.NoError(t, err)

		dlg, err := delegation.Delegate[capability.Unit](
			testutil.RandomSigner(),
			testutil.RandomDID(),
			[]ucan.Capability[capability.Unit]{
				ucan.NewCapability("test/test", testutil.RandomDID().String(), capability.Unit{}),
			},
		)
		require.NoError(t, err)

		err = store.Put(context.Background(), dlg)
		require.NoError(t, err)

		res, err := store.Get(context.Background(), dlg.Link())
		require.NoError(t, err)
		testutil.RequireEqualDelegation(t, dlg, res)
	})
}
