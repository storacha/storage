package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ipfs/go-datastore"
	"github.com/ipni/go-libipni/maurl"
	"github.com/storacha/go-metadata"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/pdp"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/service/blobs"
	"github.com/storacha/storage/pkg/service/claims"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/delegationstore"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

type StorageService struct {
	id           principal.Signer
	blobs        blobs.Blobs
	claims       claims.Claims
	pdp          pdp.PDP
	receiptStore receiptstore.ReceiptStore
	startFuncs   []func() error
	closeFuncs   []func(ctx context.Context) error
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

func (s *StorageService) PDP() pdp.PDP {
	return s.pdp
}

func (s *StorageService) Receipts() receiptstore.ReceiptStore {
	return s.receiptStore
}

func (s *StorageService) Startup() error {
	var err error
	for _, startFunc := range s.startFuncs {
		err = errors.Join(startFunc())
	}
	s.startFuncs = []func() error{}
	return err
}

func (s *StorageService) Close(ctx context.Context) error {
	var err error
	for _, close := range s.closeFuncs {
		err = errors.Join(close(ctx))
	}
	s.closeFuncs = []func(context.Context) error{}
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

	var closeFuncs []func(context.Context) error
	var startFuncs []func() error

	blobOpts := []blobs.Option{}

	if c.allocationStore == nil {
		allocDs := c.allocationDatastore
		if allocDs == nil {
			allocDs = datastore.NewMapDatastore()
			log.Warn("Allocation datastore not configured, using in-memory datastore")
		}
		closeFuncs = append(closeFuncs, func(context.Context) error { return allocDs.Close() })
		blobOpts = append(blobOpts, blobs.WithDSAllocationStore(allocDs))
	} else {
		blobOpts = append(blobOpts, blobs.WithAllocationStore(c.allocationStore))
	}

	claimStore := c.claimStore
	if claimStore == nil {
		claimDs := c.claimDatastore
		if claimDs == nil {
			claimDs = datastore.NewMapDatastore()
			log.Warn("Claim datastore not configured, using in-memory datastore")
		}
		closeFuncs = append(closeFuncs, func(context.Context) error { return claimDs.Close() })
		var err error
		claimStore, err = delegationstore.NewDsDelegationStore(claimDs)
		if err != nil {
			return nil, fmt.Errorf("creating claim store: %w", err)
		}
	}
	publisherStore := c.publisherStore
	if publisherStore == nil {
		publisherDs := c.publisherDatastore
		if publisherDs == nil {
			publisherDs = datastore.NewMapDatastore()
			log.Warn("Publisher datastore not configured, using in-memory datastore")
		}
		closeFuncs = append(closeFuncs, func(context.Context) error { return publisherDs.Close() })
		publisherStore = store.FromDatastore(publisherDs, store.WithMetadataContext(metadata.MetadataContext))
	}
	pubURL := c.publicURL
	if pubURL == (url.URL{}) {
		u, _ := url.Parse("http://localhost:3000")
		log.Warnf("Public URL not configured, using default: %s", u)
		pubURL = *u
	}

	receiptStore := c.receiptStore
	if receiptStore == nil {
		receiptDS := c.receiptDatastore
		if receiptDS == nil {
			receiptDS = datastore.NewMapDatastore()
			log.Warn("Receipt datastore not configured, using in-memory datastore")
		}
		closeFuncs = append(closeFuncs, func(context.Context) error { return receiptDS.Close() })
		var err error
		receiptStore, err = receiptstore.NewDsReceiptStore(receiptDS)
		if err != nil {
			return nil, fmt.Errorf("creating receipt store: %w", err)
		}
	}

	var pdpImpl pdp.PDP
	if c.pdp == nil {
		blobStore := c.blobStore
		if blobStore == nil {
			blobStore = blobstore.NewMapBlobstore()
			log.Warn("Blob store not configured, using in-memory store")
		}

		blobOpts = append(blobOpts, blobs.WithBlobstore(blobStore))
		if c.blobsPublicURL != (url.URL{}) {
			blobOpts = append(blobOpts, blobs.WithPublicURLAccess(c.blobsPublicURL))
		} else {
			blobOpts = append(blobOpts, blobs.WithPublicURLAccess(pubURL))
		}

		if c.blobsPresigner != nil {
			blobOpts = append(blobOpts, blobs.WithPresigner(c.blobsPresigner))
		} else {
			blobOpts = append(blobOpts, blobs.WithPublicURLPresigner(id, pubURL))
		}
	} else {
		curioAuth, err := curio.CreateCurioJWTAuthHeader("storacha", id)
		if err != nil {
			return nil, fmt.Errorf("generating curio JWT: %w", err)
		}
		pdpImpl = c.pdp.PDPService
		if pdpImpl == nil {
			client := curio.New(http.DefaultClient, c.pdp.CurioEndpoint, curioAuth)
			pdpService := pdp.NewLocal(c.pdp.PDPDatastore, client, c.pdp.ProofSet, id, receiptStore)
			closeFuncs = append(closeFuncs, pdpService.Shutdown)
			startFuncs = append(startFuncs, pdpService.Startup)
			pdpImpl = pdpService
		}
	}
	blobs, err := blobs.New(blobOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating blob service: %w", err)
	}

	peerAddr, err := maurl.FromURL(&pubURL)
	if err != nil {
		return nil, fmt.Errorf("parsing publisher url as multiaddr: %w", err)
	}

	claims, err := claims.New(
		id,
		claimStore,
		publisherStore,
		peerAddr,
		claims.WithPublisherDirectAnnounce(c.announceURLs...),
		claims.WithPublisherAnnounceAddress(c.publisherAnnouceAddr),
		claims.WithPublisherIndexingService(c.indexingService),
		claims.WithPublisherIndexingServiceProof(c.indexingServiceProofs...),
	)
	if err != nil {
		return nil, fmt.Errorf("creating claim service: %w", err)
	}

	return &StorageService{
		id:           c.id,
		blobs:        blobs,
		claims:       claims,
		closeFuncs:   closeFuncs,
		startFuncs:   startFuncs,
		receiptStore: receiptStore,
		pdp:          pdpImpl,
	}, nil
}
