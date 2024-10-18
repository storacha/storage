package delegationstore

import (
	"context"
	"fmt"
	"io"

	"github.com/ipfs/go-datastore"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/ucan"
)

type DsDelegationStore struct {
	data datastore.Datastore
}

func (d *DsDelegationStore) Put(ctx context.Context, dlg delegation.Delegation) error {
	k := encodeKey(dlg)
	b, err := io.ReadAll(dlg.Archive())
	if err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}

	err = d.data.Put(ctx, k, b)
	if err != nil {
		return fmt.Errorf("writing to datastore: %w", err)
	}

	return nil
}

func (d *DsDelegationStore) Get(ctx context.Context, root ucan.Link) (delegation.Delegation, error) {
	k := datastore.NewKey(root.String())

	data, err := d.data.Get(ctx, k)
	if err != nil {
		return nil, fmt.Errorf("getting from datastore: %w", err)
	}

	dlg, err := delegation.Extract(data)
	if err != nil {
		return nil, fmt.Errorf("extracting delegation: %w", err)
	}

	return dlg, nil
}

var _ DelegationStore = (*DsDelegationStore)(nil)

// NewDsDelegationStore creates a [DelegationStore] backed by an IPFS datastore.
func NewDsDelegationStore(ds datastore.Datastore) (*DsDelegationStore, error) {
	return &DsDelegationStore{ds}, nil
}

func encodeKey(d delegation.Delegation) datastore.Key {
	return datastore.NewKey(d.Link().String())
}
