package storage

import (
	"net/url"

	"github.com/multiformats/go-multihash"
	"github.com/storacha/storage/pkg/store/allocationstore"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/claimstore"
)

type Service interface {
	Allocations() allocationstore.AllocationStore
	Blobs() blobstore.Blobstore
	Claims() claimstore.ClaimStore
	SignURL(digest multihash.Multihash, size uint64, ttl uint64) (url.URL, map[string]string, error)
}
