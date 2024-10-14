package blobstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
)

type FileObject struct {
	name      string
	size      int64
	byteRange Range
}

func (o FileObject) Size() int64 {
	return o.size
}

func (o FileObject) Body() io.Reader {
	r, w := io.Pipe()
	f, err := os.Open(o.name)
	if err != nil {
		r.CloseWithError(err)
		return r
	}

	if o.byteRange.Offset > 0 {
		f.Seek(int64(o.byteRange.Offset), io.SeekStart)
	}

	go func() {
		var err error
		if o.byteRange.Length != nil {
			_, err = io.CopyN(w, f, int64(*o.byteRange.Length))
		} else {
			_, err = io.Copy(w, f)
		}
		f.Close()
		w.CloseWithError(err)
	}()

	return r
}

func toPath(digest multihash.Multihash) string {
	str, _ := multibase.Encode(multibase.Base58BTC, digest)
	var parts []string
	for i := 0; i < len(str); i += 2 {
		end := i + 2
		if end > len(str) {
			end = len(str)
		}
		parts = append(parts, str[i:end])
	}
	return path.Join(parts...)
}

type FsBlobstore struct {
	rootdir string
}

func (b *FsBlobstore) Get(ctx context.Context, digest multihash.Multihash, opts ...GetOption) (Object, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	n := path.Join(b.rootdir, toPath(digest))
	f, err := os.Open(n)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	inf, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	return FileObject{name: n, size: inf.Size(), byteRange: o.byteRange}, nil
}

func (b *FsBlobstore) Put(ctx context.Context, digest multihash.Multihash, body io.Reader) error {
	info, err := multihash.Decode(digest)
	if err != nil {
		return fmt.Errorf("decoding digest: %w", err)
	}
	if info.Code != multihash.SHA2_256 {
		return fmt.Errorf("unsupported digest: 0x%x", info.Code)
	}

	n := path.Join(b.rootdir, toPath(digest))
	err = os.MkdirAll(path.Dir(n), 0755)
	if err != nil {
		return fmt.Errorf("creating intermediate directories: %w", err)
	}

	f, err := os.Create(n)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	hash := sha256.New()
	tee := io.TeeReader(body, hash)

	_, err = io.Copy(f, tee)
	if err != nil {
		os.Remove(n) // remove any bytes written
		return fmt.Errorf("writing file: %w", err)
	}

	if !bytes.Equal(hash.Sum(nil), info.Digest) {
		os.Remove(n)
		return ErrDataInconsistent
	}

	return nil
}

var _ Blobstore = (*FsBlobstore)(nil)

func NewFsBlobstore(rootdir string) (*FsBlobstore, error) {
	err := os.MkdirAll(rootdir, 0755)
	if err != nil {
		return nil, fmt.Errorf("root directory not writable: %w", err)
	}
	return &FsBlobstore{rootdir}, nil
}
