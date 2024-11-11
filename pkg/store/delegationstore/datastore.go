package delegationstore

import (
	"github.com/ipfs/go-datastore"
	"github.com/storacha/ipni-publisher/pkg/store"
)

// NewDsDelegationStore creates a [DelegationStore] backed by an IPFS datastore.
func NewDsDelegationStore(ds datastore.Datastore) (DelegationStore, error) {
	return NewDelegationStore(store.SimpleStoreFromDatastore(ds))
}
