package publisher

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/delegation"
	ipnipub "github.com/storacha/ipni-publisher/pkg/publisher"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/capability/assert"
	"github.com/storacha/storage/pkg/metadata"
	"github.com/storacha/storage/pkg/service/publisher/advertisement"
)

const claimPattern = "{claim}"

var log = logging.Logger("publisher")

type PublisherService struct {
	store            store.PublisherStore
	publisher        ipnipub.Publisher
	peerInfo         peer.AddrInfo
	claimPathPattern string
}

func (pub *PublisherService) Store() store.PublisherStore {
	return pub.store
}

func (pub *PublisherService) Publish(ctx context.Context, claim delegation.Delegation) error {
	ability := claim.Capabilities()[0].Can()
	switch ability {
	case assert.LocationAbility:
		return PublishLocationCommitment(ctx, pub.publisher, pub.peerInfo, pub.claimPathPattern, claim)
	default:
		return fmt.Errorf("unknown claim: %s", ability)
	}
}

func PublishLocationCommitment(ctx context.Context, publisher ipnipub.Publisher, peerInfo peer.AddrInfo, pathPattern string, claim delegation.Delegation) error {
	provider := peer.AddrInfo{ID: peerInfo.ID}
	suffix, err := multiaddr.NewMultiaddr("/http-path/" + url.PathEscape(pathPattern))
	if err != nil {
		return fmt.Errorf("building http-path suffix: %w", err)
	}
	for _, addr := range peerInfo.Addrs {
		provider.Addrs = append(provider.Addrs, multiaddr.Join(addr, suffix))
	}

	cap := claim.Capabilities()[0]
	nb, rerr := assert.LocationCaveatsReader.Read(cap.Nb())
	if rerr != nil {
		return fmt.Errorf("reading location commitment data: %w", rerr)
	}

	digests := []multihash.Multihash{nb.Content}
	contextid, err := advertisement.EncodeContextID(nb.Space, nb.Content)
	if err != nil {
		return fmt.Errorf("encoding advertisement context ID: %w", err)
	}

	var exp int
	if claim.Expiration() != nil {
		exp = *claim.Expiration()
	}

	meta := metadata.MetadataContext.New(
		&metadata.LocationCommitmentMetadata{
			Claim:      claim.Link(),
			Expiration: int64(exp),
		},
	)

	adlink, err := publisher.Publish(ctx, provider, string(contextid), digests, meta)
	if err != nil {
		return fmt.Errorf("publishing claim: %w", err)
	}

	log.Infof("published advertisement for location commitment: %s", adlink)

	// TODO: cache in indexing-service
	return nil
}

var _ Publisher = (*PublisherService)(nil)

// New creates a [Publisher] that publishes content claims/commitments to IPNI
// and caches them with the indexing service.
//
// The publicAddr parameter is the base public address where adverts and claims
// can be read from. When publishing, the address is suffixed with a
// /http-path/<path> multiaddr, where "path" is the URI encoded version of the
// configured claim path.
//
// Note: publicAddr address must be HTTP(S).
func New(id crypto.PrivKey, publisherStore store.PublisherStore, publicAddr multiaddr.Multiaddr, opts ...Option) (*PublisherService, error) {
	o := &options{}
	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}

	publisher, err := ipnipub.New(
		id,
		publisherStore,
		ipnipub.WithDirectAnnounce("https://cid.contact/announce"),
		ipnipub.WithAnnounceAddrs(publicAddr.String()),
	)
	if err != nil {
		return nil, err
	}

	found := false
	for _, p := range publicAddr.Protocols() {
		if p.Code == multiaddr.P_HTTPS || p.Code == multiaddr.P_HTTP {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("IPNI publisher address is not HTTP(S): %s", publicAddr)
	}
	claimPath := o.claimPath
	if claimPath == "" {
		claimPath = fmt.Sprintf("claim/%s", claimPattern)
	}
	if !strings.Contains(claimPath, claimPattern) {
		return nil, fmt.Errorf(`path string does not contain required pattern: "%s"`, claimPattern)
	}
	claimPath = strings.TrimPrefix(claimPath, "/")

	peerid, err := peer.IDFromPrivateKey(id)
	if err != nil {
		return nil, err
	}
	peerInfo := peer.AddrInfo{
		ID:    peerid,
		Addrs: []multiaddr.Multiaddr{publicAddr},
	}
	return &PublisherService{publisherStore, publisher, peerInfo, claimPath}, nil
}
