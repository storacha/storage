package claims

import (
	"github.com/multiformats/go-multiaddr"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/service/publisher"
	"github.com/storacha/storage/pkg/store/claimstore"
)

type ClaimService struct {
	store     claimstore.ClaimStore
	publisher publisher.Publisher
}

func (c *ClaimService) Publisher() publisher.Publisher {
	return c.publisher
}

func (c *ClaimService) Store() claimstore.ClaimStore {
	return c.store
}

var _ Claims = (*ClaimService)(nil)

func New(id principal.Signer, claimStore claimstore.ClaimStore, publisherStore store.PublisherStore, addr multiaddr.Multiaddr, opts ...Option) (*ClaimService, error) {
	o := &options{}
	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}

	publisher, err := publisher.New(
		id,
		publisherStore,
		addr,
		publisher.WithDirectAnnounce(o.announceURLs...),
		publisher.WithIndexingService(o.indexingService),
		publisher.WithIndexingServiceProof(o.indexingServiceProofs...),
	)
	if err != nil {
		return nil, err
	}

	return &ClaimService{claimStore, publisher}, nil
}
