package allocationstore

import (
	"context"

	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
)

type Allocation struct {
	// Space is the DID of the space this data was allocated for.
	Space did.DID
	// Digest is the hash of the data.
	Digest multihash.Multihash
	// Size of the data in bytes.
	Size uint64
	// Cause is a link to the UCAN that requested the allocation.
	Cause ucan.Link
}

// AllocationStore tracks the items that have been, or will soon be stored on
// the storage node.
type AllocationStore interface {
	// List retrieves allocations by the digest of the data allocated.
	List(context.Context, multihash.Multihash) ([]Allocation, error)
	// Put adds or replaces allocation data in the store.
	Put(context.Context, Allocation) error
}
