package allocationstore

import (
	"context"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/multiformats/go-multibase"
	multihash "github.com/multiformats/go-multihash"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
)

type DsAllocationStore struct {
	data datastore.Datastore
}

func (d *DsAllocationStore) List(ctx context.Context, digest multihash.Multihash) ([]allocation.Allocation, error) {
	pfx, _ := multibase.Encode(multibase.Base58BTC, digest)
	results, err := d.data.Query(ctx, query.Query{Prefix: pfx})
	if err != nil {
		return nil, fmt.Errorf("querying datastore: %w", err)
	}

	var allocs []allocation.Allocation
	for entry := range results.Next() {
		a, err := allocation.Decode(entry.Value, dagcbor.Decode)
		if err != nil {
			return nil, fmt.Errorf("decoding data: %w", err)
		}
		allocs = append(allocs, a)
	}
	return allocs, nil
}

func (d *DsAllocationStore) Put(ctx context.Context, alloc allocation.Allocation) error {
	k := encodeKey(alloc)
	b, err := allocation.Encode(alloc, dagcbor.Encode)
	if err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}

	err = d.data.Put(ctx, k, b)
	if err != nil {
		return fmt.Errorf("writing to datastore: %w", err)
	}

	return nil
}

var _ AllocationStore = (*DsAllocationStore)(nil)

// NewDsAllocationStore creates an [AllocationStore] backed by an IPFS datastore.
func NewDsAllocationStore(ds datastore.Datastore) (*DsAllocationStore, error) {
	return &DsAllocationStore{ds}, nil
}

func encodeKey(a allocation.Allocation) datastore.Key {
	str, _ := multibase.Encode(multibase.Base58BTC, a.Digest)
	return datastore.NewKey(fmt.Sprintf("%s/%s", str, a.Cause.String()))
}