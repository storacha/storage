package allocationstore

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/storacha/storage/pkg/internal/testutil"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
	"github.com/stretchr/testify/require"
)

func TestDsAllocationStore(t *testing.T) {
	t.Run("roundtrip", func(t *testing.T) {
		store, err := NewDsAllocationStore(datastore.NewMapDatastore())
		require.NoError(t, err)

		alloc := allocation.Allocation{
			Space:  testutil.RandomDID(),
			Digest: testutil.RandomMultihash(),
			Size:   uint64(1 + rand.IntN(1000)),
			Cause:  testutil.RandomCID(),
		}

		err = store.Put(context.Background(), alloc)
		require.NoError(t, err)

		allocs, err := store.List(context.Background(), alloc.Digest)
		require.NoError(t, err)
		require.Len(t, allocs, 1)
		require.Equal(t, alloc, allocs[0])
	})

	t.Run("multiple", func(t *testing.T) {
		store, err := NewDsAllocationStore(datastore.NewMapDatastore())
		require.NoError(t, err)

		alloc0 := allocation.Allocation{
			Space:  testutil.RandomDID(),
			Digest: testutil.RandomMultihash(),
			Size:   uint64(1 + rand.IntN(1000)),
			Cause:  testutil.RandomCID(),
		}

		alloc1 := allocation.Allocation{
			Space:  testutil.RandomDID(),
			Digest: alloc0.Digest,
			Size:   alloc0.Size,
			Cause:  testutil.RandomCID(),
		}

		err = store.Put(context.Background(), alloc0)
		require.NoError(t, err)
		err = store.Put(context.Background(), alloc1)
		require.NoError(t, err)

		allocs, err := store.List(context.Background(), alloc0.Digest)
		require.NoError(t, err)
		require.Len(t, allocs, 2)

		require.Equal(t, []allocation.Allocation{alloc0, alloc1}, allocs)
	})
}
