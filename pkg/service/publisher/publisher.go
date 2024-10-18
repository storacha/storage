package publisher

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/delegation"
	ipni "github.com/storacha/ipni-publisher/pkg/publisher"
	"github.com/storacha/storage/pkg/capability/assert"
	"github.com/storacha/storage/pkg/metadata"
	"github.com/storacha/storage/pkg/service/publisher/advertisement"
)

const claimPattern = "{claim}"

var log = logging.Logger("publisher")

type PublisherService struct {
	publisher        ipni.Publisher
	peerInfo         peer.AddrInfo
	claimPathPattern string
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

func PublishLocationCommitment(ctx context.Context, publisher ipni.Publisher, peerInfo peer.AddrInfo, pathPattern string, claim delegation.Delegation) error {
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
// The peerInfo parameter is the base public peer information. When publishing,
// all addresses are suffixed with a /http-path/<path> multiaddr, where "path"
// is the URI encoded version of claimPathPattern.
//
// Note: peerInfo addresses must be HTTP(S).
//
// The claimPathPattern parameter MUST include the string "{claim}", which upon
// retrieval is replaced with the CID of a content claim. That is to say, the
// combination of each address in peerInfo + claimPathPattern (with "{claim}"
// replaced with a real CID) should be the URL allowing to GET that content
// claim on this node.
//
// e.g. If peerInfo contains a multiaddr /dns4/n0.storacha.network/tcp/443/https
// with pathPattern: "claim/{claim}", then a claim should be retrievable at URL:
// https://n0.storacha.network/claim/bafyreidn6rkycfi2wvn6zbzgd2jnpi362opytoyprt5e27g44whrnh453a
func New(publisher ipni.Publisher, peerInfo peer.AddrInfo, claimPathPattern string) (*PublisherService, error) {
	for _, addr := range peerInfo.Addrs {
		found := false
		for _, p := range addr.Protocols() {
			if p.Code == multiaddr.P_HTTPS || p.Code == multiaddr.P_HTTP {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("IPNI publisher address is not HTTP(S): %s", addr)
		}
	}
	if !strings.Contains(claimPathPattern, claimPattern) {
		return nil, fmt.Errorf(`path string does not contain required pattern: "%s"`, claimPattern)
	}
	claimPathPattern = strings.TrimPrefix(claimPathPattern, "/")
	return &PublisherService{publisher, peerInfo, claimPathPattern}, nil
}
