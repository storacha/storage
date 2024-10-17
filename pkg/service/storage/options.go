package storage

import (
	"net/url"

	"github.com/ipfs/go-datastore"
	"github.com/storacha/go-ucanto/principal"
)

type config struct {
	id                  principal.Signer
	publicURL           url.URL
	dataDir             string
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

// WithDataDir configures the path to the filesystem directory where uploads
// will be stored.
func WithDataDir(dir string) Option {
	return func(c *config) error {
		c.dataDir = dir
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
