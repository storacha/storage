package publisher

import (
	"context"

	"github.com/storacha/go-ucanto/core/delegation"
)

type Publisher interface {
	// Publish advertises content claims/commitments found on this node to the
	// storacha network.
	Publish(context.Context, delegation.Delegation) error
}
