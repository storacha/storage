package storage

import (
	"net/url"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/storage/pkg/store/blobstore"
)

type config struct {
	id                  principal.Signer
	publicURL           url.URL
	blobStore           blobstore.Blobstore
	allocationDatastore datastore.Datastore
	claimDatastore      datastore.Datastore
	publisherDatastore  datastore.Datastore
}

type Option func(*config) error

// WithIdentity configures the storage service identity, used to sign UCAN
// invocations and receipts.
func WithIdentity(signer principal.Signer) Option {
	return func(c *config) error {
		c.id = signer
		return nil
	}
}

// WithPublicURL configures the URL this storage node will be publically
// accessible from.
func WithPublicURL(url url.URL) Option {
	return func(c *config) error {
		c.publicURL = url
		return nil
	}
}

// WithBlobstore configures the blob storage to use.
func WithBlobstore(blobStore blobstore.Blobstore) Option {
	return func(c *config) error {
		c.blobStore = blobStore
		return nil
	}
}

// WithAllocationDatastore configures the underlying datastore to use for
// storing allocation records. Note: the datastore MUST have efficient support
// for prefix queries.
func WithAllocationDatastore(dstore datastore.Datastore) Option {
	return func(c *config) error {
		c.allocationDatastore = dstore
		return nil
	}
}

// WithClaimDatastore configures the underlying datastore to use for storing
// content claims made by this node.
func WithClaimDatastore(dstore datastore.Datastore) Option {
	return func(c *config) error {
		c.claimDatastore = dstore
		return nil
	}
}

// WithPublisherDatastore configures the underlying datastore to use for storing
// IPNI advertisements and their entries.
func WithPublisherDatastore(dstore datastore.Datastore) Option {
	return func(c *config) error {
		c.publisherDatastore = dstore
		return nil
	}
}

// WithLogLevel changes the log level of a specific subsystem name=="*" changes
// all subsystems.
func WithLogLevel(name string, level string) Option {
	return func(c *config) error {
		logging.SetLogLevel(name, level)
		return nil
	}
}
