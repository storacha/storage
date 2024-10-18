package delegationstore

import (
	"context"

	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/ucan"
)

// DelegationStore stores UCAN delegations.
type DelegationStore interface {
	// Get retrieves a delegation by it's root CID.
	Get(context.Context, ucan.Link) (delegation.Delegation, error)
	// Put adds or replaces a delegation in the store.
	Put(context.Context, delegation.Delegation) error
}
