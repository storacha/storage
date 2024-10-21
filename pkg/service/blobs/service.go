package blobs

import (
	"fmt"
	"net/url"

	"github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/storage/pkg/access"
	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/presigner"
	"github.com/storacha/storage/pkg/store/allocationstore"
	"github.com/storacha/storage/pkg/store/blobstore"
)

type BlobService struct {
	access     access.Access
	allocStore allocationstore.AllocationStore
	blobStore  blobstore.Blobstore
	presigner  presigner.RequestPresigner
}

func (b *BlobService) Access() access.Access {
	return b.access
}

func (b *BlobService) Allocations() allocationstore.AllocationStore {
	return b.allocStore
}

func (b *BlobService) Presigner() presigner.RequestPresigner {
	return b.presigner
}

func (b *BlobService) Store() blobstore.Blobstore {
	return b.blobStore
}

var _ Blobs = (*BlobService)(nil)

func New(id principal.Signer, blobStore blobstore.Blobstore, allocsDatastore datastore.Datastore, publicURL url.URL) (*BlobService, error) {
	allocStore, err := allocationstore.NewDsAllocationStore(allocsDatastore)
	if err != nil {
		return nil, err
	}

	accessURL := publicURL
	accessURL.Path = "/blob"
	access, err := access.NewPatternAccess(fmt.Sprintf("%s/{blob}", accessURL.String()))
	if err != nil {
		return nil, err
	}

	accessKeyID := id.DID().String()
	idDigest, _ := multihash.Sum(id.Encode(), multihash.SHA2_256, -1)
	secretAccessKey := digestutil.Format(idDigest)
	presigner, err := presigner.NewS3RequestPresigner(accessKeyID, secretAccessKey, publicURL, "blob")
	if err != nil {
		return nil, err
	}

	return &BlobService{access, allocStore, blobStore, presigner}, nil
}
