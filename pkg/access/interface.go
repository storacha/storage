package access

import (
	"net/url"

	"github.com/multiformats/go-multihash"
)

type Access interface {
	// GetDownloadURL constructs a public download URL for the given blob digest.
	// Note: it does not verify the blob exists.
	GetDownloadURL(digest multihash.Multihash) (url.URL, error)
}
