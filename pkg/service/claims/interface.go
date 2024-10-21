package claims

import (
	"github.com/storacha/storage/pkg/service/publisher"
	"github.com/storacha/storage/pkg/store/claimstore"
)

type Claims interface {
	// Store is the storage for location claims.
	Store() claimstore.ClaimStore
	// Publisher advertises content claims/commitments found on this node to the
	// storacha network.
	Publisher() publisher.Publisher
}
