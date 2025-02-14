package blobstore

import (
	"context"
	"io"

	multihash "github.com/multiformats/go-multihash"
)

type noopBlobstore struct{}

func NewNoopBlobstore() Blobstore {
	return &noopBlobstore{}
}

func (b *noopBlobstore) Put(ctx context.Context, digest multihash.Multihash, size uint64, body io.Reader) error {
	return nil
}

func (b *noopBlobstore) Get(ctx context.Context, digest multihash.Multihash, opts ...GetOption) (Object, error) {
	return nil, nil
}
