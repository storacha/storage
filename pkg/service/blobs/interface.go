package blobs

import (
	"github.com/storacha/storage/pkg/access"
	"github.com/storacha/storage/pkg/presigner"
	"github.com/storacha/storage/pkg/store/allocationstore"
	"github.com/storacha/storage/pkg/store/blobstore"
)

type Blobs interface {
	// Blobs is the storage interface for blobs.
	Store() blobstore.Blobstore
	// Allocations is a store for received blob allocations.
	Allocations() allocationstore.AllocationStore
	// Presigner provides an interface to allow signed request access to upload blobs.
	Presigner() presigner.RequestPresigner
	// Access provides an interface to allowing public access to download blobs.
	Access() access.Access
}
