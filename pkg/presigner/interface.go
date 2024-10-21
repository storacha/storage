package presigner

import (
	"context"
	"net/http"
	"net/url"

	"github.com/multiformats/go-multihash"
)

type RequestPresigner interface {
	// SignUploadURL creates and signs a URL that allows a PUT request to upload
	// data for the given digest/size to the service.
	//
	// The ttl parameter determines the number of seconds the signed URL will be
	// valid for.
	//
	// It returns a signed URL that will accept a PUT request, and a set of HTTP
	// headers that should also be sent with the request.
	SignUploadURL(ctx context.Context, digest multihash.Multihash, size uint64, ttl uint64) (url.URL, http.Header, error)
	// VerifyUploadURL ensures the upload URL was signed by this service. It
	// returns the _signed_ URL and headers or error if the signature is invalid.
	VerifyUploadURL(ctx context.Context, url url.URL, headers http.Header) (url.URL, http.Header, error)
}
