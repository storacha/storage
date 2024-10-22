package claims

import (
	"net/url"

	"github.com/ipfs/go-datastore"
	"github.com/ipni/go-libipni/maurl"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/service/publisher"
	"github.com/storacha/storage/pkg/store/claimstore"
	"github.com/storacha/storage/pkg/store/delegationstore"
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

func New(id principal.Signer, claimsDatastore datastore.Datastore, publisherDatastore datastore.Datastore, publicURL url.URL, opts ...Option) (*ClaimService, error) {
	o := &options{}
	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}

	claimStore, err := delegationstore.NewDsDelegationStore(claimsDatastore)
	if err != nil {
		return nil, err
	}

	addr, err := maurl.FromURL(&publicURL)
	if err != nil {
		return nil, err
	}

	publisherStore := store.FromDatastore(publisherDatastore)

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
