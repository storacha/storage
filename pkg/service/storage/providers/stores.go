package providers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ipfs/go-datastore"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/storacha/go-libstoracha/metadata"
	"go.uber.org/fx"

	"github.com/storacha/piri/pkg/config"
	"github.com/storacha/piri/pkg/store/allocationstore"
	"github.com/storacha/piri/pkg/store/blobstore"
	"github.com/storacha/piri/pkg/store/claimstore"
	"github.com/storacha/piri/pkg/store/delegationstore"
	"github.com/storacha/piri/pkg/store/receiptstore"
)

// StoreParams contains dependencies for creating stores
type StoreParams struct {
	fx.In
	Config       config.UCANServer
	AllocationDS datastore.Datastore `name:"allocation"`
	ClaimDS      datastore.Datastore `name:"claim"`
	PublisherDS  datastore.Datastore `name:"publisher"`
	ReceiptDS    datastore.Datastore `name:"receipt"`
}

// NewAllocationStore creates an allocation store
// Note: Currently returns nil as the blobs service manages its own allocation store
// This is here for future extensibility
func NewAllocationStore(params StoreParams) (allocationstore.AllocationStore, error) {
	// The blobs service currently creates its own allocation store from the datastore
	// This provider is here for future use when we want to inject a custom allocation store
	return nil, nil
}

// NewClaimStore creates a claim store
func NewClaimStore(params StoreParams) (claimstore.ClaimStore, error) {
	store, err := delegationstore.NewDsDelegationStore(params.ClaimDS)
	if err != nil {
		return nil, fmt.Errorf("creating claim store: %w", err)
	}
	return store, nil
}

// NewPublisherStore creates a publisher store
func NewPublisherStore(params StoreParams) store.PublisherStore {
	return store.FromDatastore(params.PublisherDS, store.WithMetadataContext(metadata.MetadataContext))
}

// NewReceiptStore creates a receipt store
func NewReceiptStore(params StoreParams) (receiptstore.ReceiptStore, error) {
	store, err := receiptstore.NewDsReceiptStore(params.ReceiptDS)
	if err != nil {
		return nil, fmt.Errorf("creating receipt store: %w", err)
	}
	return store, nil
}

// BlobStoreParams for blob store creation
type BlobStoreParams struct {
	fx.In
	Config config.UCANServer
}

// NewBlobStore creates a blob store based on configuration
func NewBlobStore(params BlobStoreParams) (blobstore.Blobstore, error) {
	if params.Config.DataDir == "" {
		log.Warn("DataDir not configured, using in-memory blobstore store")
		return blobstore.NewMapBlobstore(), nil
	}

	// Create directory if it doesn't exist
	blobPath := filepath.Join(params.Config.DataDir, "blob")
	if err := os.MkdirAll(blobPath, 0755); err != nil {
		return nil, fmt.Errorf("creating blob store directory: %w", err)
	}
	tmpBlobPath := filepath.Join(params.Config.TempDir, "piri-blob")

	// Create file-based blob store
	bs, err := blobstore.NewFsBlobstore(blobPath, tmpBlobPath)
	if err != nil {
		return nil, fmt.Errorf("creating filesystem blob store: %w", err)
	}

	log.Infof("Created filesystem blob store at %s", blobPath)
	return bs, nil
}

// StoresModule provides all store dependencies
var StoresModule = fx.Module("stores",
	fx.Provide(
		NewAllocationStore,
		NewClaimStore,
		NewPublisherStore,
		NewReceiptStore,
		NewBlobStore,
	),
)
