package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ipfs/go-datastore"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/storage/pkg/pdp"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/service/blobs"
	"github.com/storacha/storage/pkg/service/claims"
	"github.com/storacha/storage/pkg/store/blobstore"
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
	allocDs := c.allocationDatastore
	if allocDs == nil {
		allocDs = datastore.NewMapDatastore()
		log.Warn("Allocation datastore not configured, using in-memory datastore")
	}
	closeFuncs = append(closeFuncs, func(context.Context) error { return allocDs.Close() })

	claimDs := c.claimDatastore
	if claimDs == nil {
		claimDs = datastore.NewMapDatastore()
		log.Warn("Claim datastore not configured, using in-memory datastore")
	}
	closeFuncs = append(closeFuncs, func(context.Context) error { return claimDs.Close() })

	publisherDs := c.publisherDatastore
	if publisherDs == nil {
		publisherDs = datastore.NewMapDatastore()
		log.Warn("Publisher datastore not configured, using in-memory datastore")
	}
	closeFuncs = append(closeFuncs, func(context.Context) error { return publisherDs.Close() })

	pubURL := c.publicURL
	if pubURL == (url.URL{}) {
		u, _ := url.Parse("http://localhost:3000")
		log.Warnf("Public URL not configured, using default: %s", u)
		pubURL = *u
	}

	receiptDS := c.receiptDatastore
	if receiptDS == nil {
		receiptDS = datastore.NewMapDatastore()
		log.Warn("Receipt datastore not configured, using in-memory datastore")
	}
	closeFuncs = append(closeFuncs, func(context.Context) error { return receiptDS.Close() })

	receiptStore, err := receiptstore.NewDsReceiptStore(receiptDS)
	if err != nil {
		return nil, fmt.Errorf("creating receipt store: %w", err)
	}

	blobOpts := []blobs.Option{blobs.WithDSAllocationStore(allocDs)}

	var pdpImpl pdp.PDP
	if c.pdp == nil {
		blobStore := c.blobStore
		if blobStore == nil {
			blobStore = blobstore.NewMapBlobstore()
			log.Warn("Blob store not configured, using in-memory store")
		}
		blobOpts = append(blobOpts, blobs.WithBlobstore(blobStore))
		blobOpts = append(blobOpts, blobs.WithPublicURLAccess(pubURL))
		blobOpts = append(blobOpts, blobs.WithPublicURLPresigner(id, pubURL))
	} else {
		client := curio.New(http.DefaultClient, c.pdp.CurioEndpoint, c.pdp.CurioAuthHeader)
		pdpService := pdp.NewLocal(c.pdp.PDPDatastore, client, c.pdp.ProofSet, id, receiptStore)
		closeFuncs = append(closeFuncs, pdpService.Shutdown)
		startFuncs = append(startFuncs, pdpService.Startup)
		pdpImpl = pdpService
	}
	blobs, err := blobs.New(blobOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating blob service: %w", err)
	}

	claims, err := claims.New(
		id,
		claimDs,
		publisherDs,
		pubURL,
		claims.WithPublisherDirectAnnounce(c.announceURLs...),
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
