package keystore

import (
	"context"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
)

const (
	DatastorePrefix = "keystore/"
	DefaultKeyName  = "default"
)

// KeyInfo is used for storing keys in KeyStore
type KeyInfo struct {
	PrivateKey []byte
}

// KeyStore is used for storing secret keys
type KeyStore interface {
	// Get gets a key out of keystore and returns KeyInfo corresponding to named key
	Get(context.Context, string) (KeyInfo, error)
	// Put saves a key info under given name
	Put(context.Context, string, KeyInfo) error
}

type keyStore struct {
	ds datastore.Datastore
}

func NewKeyStore(ds datastore.Datastore) (KeyStore, error) {
	ks := namespace.Wrap(ds, datastore.NewKey(DatastorePrefix))
	return &keyStore{ds: ks}, nil
}

func (k *keyStore) Get(ctx context.Context, s string) (KeyInfo, error) {
	res, err := k.ds.Get(ctx, datastore.NewKey(s))
	if err != nil {
		return KeyInfo{}, fmt.Errorf("getting key (%s): %w", s, err)
	}
	return KeyInfo{PrivateKey: res}, nil
}

func (k *keyStore) Put(ctx context.Context, s string, info KeyInfo) error {
	if err := k.ds.Put(ctx, datastore.NewKey(s), info.PrivateKey); err != nil {
		return fmt.Errorf("putting key (%s): %w", s, err)
	}
	return nil
}
