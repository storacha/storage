package delegationstore

import (
	"context"
	"fmt"
	"io"

	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/ipni-publisher/pkg/store"
)

type delegationStore struct {
	data store.Store
}

func (d *delegationStore) Put(ctx context.Context, dlg delegation.Delegation) error {
	err := d.data.Put(ctx, dlg.Link().String(), dlg.Archive())
	if err != nil {
		return fmt.Errorf("writing to datastore: %w", err)
	}
	return nil
}

func (d *delegationStore) Get(ctx context.Context, root ucan.Link) (delegation.Delegation, error) {

	r, err := d.data.Get(ctx, root.String())
	if err != nil {
		return nil, fmt.Errorf("getting from datastore: %w", err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading delegation data: %w", err)
	}
	dlg, err := delegation.Extract(data)
	if err != nil {
		return nil, fmt.Errorf("extracting delegation: %w", err)
	}

	return dlg, nil
}

// NewDelegationStore creates a [DelegationStore] backed by a simple store interface
func NewDelegationStore(store store.Store) (DelegationStore, error) {
	return &delegationStore{store}, nil
}
