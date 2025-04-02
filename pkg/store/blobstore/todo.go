package blobstore

import (
	"context"
	"fmt"
	"io"

	"github.com/multiformats/go-multihash"

	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/store"
)

func NewFakeMapBlobstore() *FakeMapBlobstore {
	data := map[string][]byte{}
	return &FakeMapBlobstore{data}
}

type FakeMapBlobstore struct {
	data map[string][]byte
}

func (mb *FakeMapBlobstore) Get(ctx context.Context, digest multihash.Multihash, opts ...GetOption) (Object, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	k := digestutil.Format(digest)
	b, ok := mb.data[k]
	if !ok {
		return nil, store.ErrNotFound
	}

	obj := MapObject{bytes: b, byteRange: o.byteRange}
	return obj, nil
}

func (mb *FakeMapBlobstore) Put(ctx context.Context, digest multihash.Multihash, size uint64, body io.Reader) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	if len(b) > int(size) {
		return ErrTooLarge
	}
	if len(b) < int(size) {
		return ErrTooSmall
	}

	k := digestutil.Format(digest)
	mb.data[k] = b

	return nil
}
