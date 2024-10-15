package blobstore

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/multiformats/go-multihash"
)

// ErrDataInconsistent is returned when the data being written does not hash to
// the expected value.
var ErrDataInconsistent = errors.New("data consistency check failed")

// GetOption is an option configuring byte retrieval from a blobstore.
type GetOption func(cfg *options) error

type Range struct {
	Offset uint64
	Length *uint64
}

type options struct {
	byteRange Range
}

// WithRange configures a byte range to extract.
func WithRange(byteRange Range) GetOption {
	return func(opts *options) error {
		opts.byteRange = byteRange
		return nil
	}
}

type Object interface {
	// Size returns the total size of the object in bytes.
	Size() int64
	Body() io.Reader
}

type Blobstore interface {
	// Put stores the bytes to the store and ensures it hashes to the passed
	// digest.
	Put(context.Context, multihash.Multihash, io.Reader) error
	// Get retrieves the object identified by the passed digest. Returns nil and
	// [ErrNotFound] if the object does not exist.
	//
	// Note: data is not hashed on read.
	Get(context.Context, multihash.Multihash, ...GetOption) (Object, error)
}

// FileSystemer exposes the filesystem interface for reading blobs.
type FileSystemer interface {
	// FileSystem returns a filesystem interface for reading blobs.
	FileSystem() http.FileSystem
	// EncodePath converts a digest to a filesystem path.
	EncodePath(digest multihash.Multihash) string
}
