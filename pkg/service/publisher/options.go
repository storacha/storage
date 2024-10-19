package publisher

import (
	"net/url"

	"github.com/multiformats/go-multiaddr"
)

type options struct {
	pubHTTPAnnounceAddrs []multiaddr.Multiaddr
	announceURLs         []*url.URL
	claimPath            string
}

type Option func(*options) error

// WithDirectAnnounce sets indexer URLs to send direct HTTP announcements to.
func WithDirectAnnounce(announceURLs ...string) Option {
	return func(o *options) error {
		for _, urlStr := range announceURLs {
			u, err := url.Parse(urlStr)
			if err != nil {
				return err
			}
			o.announceURLs = append(o.announceURLs, u)
		}
		return nil
	}
}

// WithAnnounceAddrs configures the multiaddrs that are put into announce
// messages to tell indexers the addresses to fetch advertisements from.
func WithAnnounceAddrs(addrs ...string) Option {
	return func(opts *options) error {
		for _, addr := range addrs {
			if addr != "" {
				maddr, err := multiaddr.NewMultiaddr(addr)
				if err != nil {
					return err
				}
				opts.pubHTTPAnnounceAddrs = append(opts.pubHTTPAnnounceAddrs, maddr)
			}
		}
		return nil
	}
}

// WithClaimPath configures the claim path, the path from which claims are
// served from this server. If not set, it defauts to "claim/{claim}".
//
// The claim path MUST include the string "{claim}", which upon retrieval is
// replaced with the CID of a content claim. That is to say, the combination of
// each address in peerInfo + claimPathPattern (with "{claim}" replaced with a
// real CID) should be the URL allowing to GET that content claim on this node.
//
// e.g. If peerInfo contains a multiaddr /dns4/n0.storacha.network/tcp/443/https
// with pathPattern: "claim/{claim}", then a claim should be retrievable at URL:
// https://n0.storacha.network/claim/bafyreidn6rkycfi2wvn6zbzgd2jnpi362opytoyprt5e27g44whrnh453a
func WithClaimPath(path string) Option {
	return func(opts *options) error {
		opts.claimPath = path
		return nil
	}
}
