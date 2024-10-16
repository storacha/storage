package blobstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/storage/pkg/store"
)

type MapObject struct {
	bytes     []byte
	byteRange Range
}

func (o MapObject) Size() int64 {
	return int64(len(o.bytes))
}

func (o MapObject) Body() io.Reader {
	b := o.bytes
	if o.byteRange.Offset > 0 {
		b = b[o.byteRange.Offset:]
	}
	if o.byteRange.Length != nil {
		b = b[0:*o.byteRange.Length]
	}
	return bytes.NewReader(b)
}

type MapBlobstore struct {
	data map[string][]byte
}

func (mb *MapBlobstore) Get(ctx context.Context, digest multihash.Multihash, opts ...GetOption) (Object, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	k, _ := multibase.Encode(multibase.Base58BTC, digest)
	b, ok := mb.data[k]
	if !ok {
		return nil, store.ErrNotFound
	}

	obj := MapObject{bytes: b, byteRange: o.byteRange}
	return obj, nil
}

func (mb *MapBlobstore) Put(ctx context.Context, digest multihash.Multihash, size uint64, body io.Reader) error {
	info, err := multihash.Decode(digest)
	if err != nil {
		return fmt.Errorf("decoding digest: %w", err)
	}
	if info.Code != multihash.SHA2_256 {
		return fmt.Errorf("unsupported digest: 0x%x", info.Code)
	}

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

	hash := sha256.New()
	hash.Write(b)

	if !bytes.Equal(hash.Sum(nil), info.Digest) {
		return ErrDataInconsistent
	}

	k, _ := multibase.Encode(multibase.Base58BTC, digest)
	mb.data[k] = b

	return nil
}

var _ Blobstore = (*MapBlobstore)(nil)

// NewMapBlobstore creates a [Blobstore] backed by an in-memory map.
func NewMapBlobstore() (*MapBlobstore, error) {
	data := map[string][]byte{}
	return &MapBlobstore{data}, nil
}
