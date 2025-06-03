package claims

import (
	"github.com/storacha/piri/pkg/service/publisher"
	"github.com/storacha/piri/pkg/store/claimstore"
)

type Claims interface {
	// Store is the storage for location claims.
	Store() claimstore.ClaimStore
	// Publisher advertises content claims/commitments found on this node to the
	// storacha network.
	Publisher() publisher.Publisher
}
