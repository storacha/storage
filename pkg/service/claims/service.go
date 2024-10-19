package claims

import (
	"net/url"

	"github.com/ipfs/go-datastore"
	"github.com/ipni/go-libipni/maurl"
	"github.com/libp2p/go-libp2p/core/crypto"
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

func New(priv crypto.PrivKey, claimsDatastore datastore.Datastore, publisherDatastore datastore.Datastore, publicURL url.URL) (*ClaimService, error) {
	claimStore, err := delegationstore.NewDsDelegationStore(claimsDatastore)
	if err != nil {
		return nil, err
	}

	addr, err := maurl.FromURL(&publicURL)
	if err != nil {
		return nil, err
	}

	publisherStore := store.FromDatastore(publisherDatastore)

	publisher, err := publisher.New(priv, publisherStore, addr)
	if err != nil {
		return nil, err
	}

	return &ClaimService{claimStore, publisher}, nil
}
