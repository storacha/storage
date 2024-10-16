package storage

import (
	"github.com/storacha/storage/pkg/service/presigner"
	"github.com/storacha/storage/pkg/store/allocationstore"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/claimstore"
)

type Service interface {
	Allocations() allocationstore.AllocationStore
	Blobs() blobstore.Blobstore
	Claims() claimstore.ClaimStore
	Presigner() presigner.RequestPresigner
}
