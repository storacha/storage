package publisher

import (
	"context"

	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/storacha/go-ucanto/core/delegation"
)

type Publisher interface {
	// Store is the storage interface for published advertisements.
	Store() store.PublisherStore
	// Publish advertises content claims/commitments found on this node to the
	// storacha network.
	Publish(context.Context, delegation.Delegation) error
}
