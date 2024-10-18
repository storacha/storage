package storage

import (
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/storage/pkg/access"
	"github.com/storacha/storage/pkg/presigner"
	"github.com/storacha/storage/pkg/service/publisher"
	"github.com/storacha/storage/pkg/store/allocationstore"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/claimstore"
)

type Service interface {
	// ID is the storage service identity, used to sign UCAN invocations and receipts.
	ID() principal.Signer
	// Allocations is a store for received blob allocations.
	Allocations() allocationstore.AllocationStore
	// Blobs is the storage interface for blobs.
	Blobs() blobstore.Blobstore
	// Claims is the storage for location claims.
	Claims() claimstore.ClaimStore
	// Presigner provides an interface to allow signed request access to uplaod blobs.
	Presigner() presigner.RequestPresigner
	// Access provides an interface to allowing public access to download blobs.
	Access() access.Access
	// Publisher advertises content claims/commitments found on this node to the
	// storacha network.
	Publisher() publisher.Publisher
}
