package storage

import (
	"fmt"
	"io"
	"net/url"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/storage/pkg/service/blobs"
	"github.com/storacha/storage/pkg/service/claims"
	"github.com/storacha/storage/pkg/store/blobstore"
)

type StorageService struct {
	id         principal.Signer
	blobs      blobs.Blobs
	claims     claims.Claims
	closeFuncs []func() error
	io.Closer
}

func (s *StorageService) Blobs() blobs.Blobs {
	return s.blobs
}

func (s *StorageService) Claims() claims.Claims {
	return s.claims
}

func (s *StorageService) ID() principal.Signer {
	return s.id
}

func (s *StorageService) Close() error {
	var err error
	for _, close := range s.closeFuncs {
		err = close()
	}
	s.closeFuncs = []func() error{}
	return err
}

var _ Service = (*StorageService)(nil)

func New(opts ...Option) (*StorageService, error) {
	c := &config{}
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	id := c.id
	if id == nil {
		log.Warn("Generating a server identity as one has not been configured!")
		signer, err := ed25519.Generate()
		if err != nil {
			return nil, err
		}
		id = signer
	}
	log.Infof("Server ID: %s", id.DID())

	priv, err := crypto.UnmarshalEd25519PrivateKey(id.Raw())
	if err != nil {
		return nil, fmt.Errorf("unmarshaling private key: %w", err)
	}

	var closeFuncs []func() error

	blobStore := c.blobStore
	if blobStore == nil {
		ds, err := blobstore.NewMapBlobstore()
		if err != nil {
			return nil, err
		}
		blobStore = ds
		log.Warn("Blob store not configured, using in-memory store")
	}

	allocDs := c.allocationDatastore
	if allocDs == nil {
		allocDs = datastore.NewMapDatastore()
		log.Warn("Allocation datastore not configured, using in-memory datastore")
	}
	closeFuncs = append(closeFuncs, allocDs.Close)

	claimDs := c.claimDatastore
	if claimDs == nil {
		claimDs = datastore.NewMapDatastore()
		log.Warn("Claim datastore not configured, using in-memory datastore")
	}
	closeFuncs = append(closeFuncs, claimDs.Close)

	publisherDs := c.publisherDatastore
	if publisherDs == nil {
		publisherDs = datastore.NewMapDatastore()
		log.Warn("Publisher datastore not configured, using in-memory datastore")
	}
	closeFuncs = append(closeFuncs, publisherDs.Close)

	pubURL := c.publicURL
	if pubURL == (url.URL{}) {
		u, _ := url.Parse("http://localhost:3000")
		log.Warnf("Public URL not configured, using default: %s", u)
		pubURL = *u
	}

	blobs, err := blobs.New(id, blobStore, allocDs, pubURL)
	if err != nil {
		return nil, fmt.Errorf("creating blob service: %w", err)
	}

	claims, err := claims.New(priv, claimDs, publisherDs, pubURL)
	if err != nil {
		return nil, fmt.Errorf("creating claim service: %w", err)
	}

	return &StorageService{
		id:         c.id,
		blobs:      blobs,
		claims:     claims,
		closeFuncs: closeFuncs,
	}, nil
}
