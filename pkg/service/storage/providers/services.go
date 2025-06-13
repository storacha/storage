package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ipfs/go-datastore"
	"github.com/ipni/go-libipni/maurl"
	"github.com/multiformats/go-multiaddr"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/storacha/go-ucanto/client"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal"
	ucanhttp "github.com/storacha/go-ucanto/transport/http"
	"go.uber.org/fx"

	"github.com/storacha/piri/pkg/access"
	"github.com/storacha/piri/pkg/config"
	"github.com/storacha/piri/pkg/pdp"
	"github.com/storacha/piri/pkg/pdp/curio"
	"github.com/storacha/piri/pkg/presets"
	"github.com/storacha/piri/pkg/presigner"
	"github.com/storacha/piri/pkg/service/blobs"
	"github.com/storacha/piri/pkg/service/claims"
	"github.com/storacha/piri/pkg/service/replicator"
	"github.com/storacha/piri/pkg/store/allocationstore"
	"github.com/storacha/piri/pkg/store/blobstore"
	"github.com/storacha/piri/pkg/store/claimstore"
	"github.com/storacha/piri/pkg/store/receiptstore"
)

// URLsParams for URL parsing
type URLsParams struct {
	fx.In
	Config config.UCANServer
}

// URLsResult provides parsed URLs
type URLsResult struct {
	fx.Out
	PublicURL *url.URL `name:"public"`
}

// NewURLs parses and provides URLs from configuration
func NewURLs(params URLsParams) (URLsResult, error) {
	result := URLsResult{}

	// Parse public URL
	var err error
	result.PublicURL, err = url.Parse(params.Config.PublicURL)
	if err != nil {
		return result, fmt.Errorf("parsing public URL: %w", err)
	}

	return result, nil
}

// AccessParams for creating access service
type AccessParams struct {
	fx.In
	Config config.UCANServer
}

// NewAccess creates blob access service
func NewAccess(params AccessParams) access.Access {
	// If custom access is needed, it would be configured here
	// For now, we return nil to let blobs service create its own
	return nil
}

// PresignerParams for creating presigner
type PresignerParams struct {
	fx.In
	Config   config.UCANServer
	Identity principal.Signer
}

// NewPresigner creates blob presigner service
func NewPresigner(params PresignerParams) presigner.RequestPresigner {
	// If custom presigner is needed, it would be configured here
	// For now, we return nil to let blobs service create its own
	return nil
}

// UploadServiceParams for upload service creation
type UploadServiceParams struct {
	fx.In
	Config config.UCANServer
}

// NewUploadService creates connection to upload service
func NewUploadService(params UploadServiceParams) (client.Connection, error) {
	// Use configured service if provided
	if params.Config.UploadServiceDID != "" && params.Config.UploadServiceURL != "" {
		usDID, err := did.Parse(params.Config.UploadServiceDID)
		if err != nil {
			return nil, fmt.Errorf("parsing upload service DID: %w", err)
		}

		serviceURL, err := url.Parse(params.Config.UploadServiceURL)
		if err != nil {
			return nil, fmt.Errorf("parsing upload service URL: %w", err)
		}

		channel := ucanhttp.NewHTTPChannel(serviceURL)
		return client.NewConnection(usDID, channel)
	}

	// Use default presets
	channel := ucanhttp.NewHTTPChannel(presets.UploadServiceURL)
	conn, err := client.NewConnection(presets.UploadServiceDID, channel)
	if err != nil {
		return nil, fmt.Errorf("creating upload service connection: %w", err)
	}
	return conn, nil
}

// PDPParams for PDP service creation
type PDPParams struct {
	fx.In
	Config       config.UCANServer
	Identity     principal.Signer
	ReceiptStore receiptstore.ReceiptStore
	PDPDatastore datastore.Datastore `name:"pdp" optional:"true"`
	Lifecycle    fx.Lifecycle
}

// NewPDPService creates PDP service if configured
func NewPDPService(params PDPParams) (pdp.PDP, error) {
	if params.Config.PDPServerURL == "" {
		return nil, nil
	}

	pdpURL, err := url.Parse(params.Config.PDPServerURL)
	if err != nil {
		return nil, fmt.Errorf("parsing PDP server URL: %w", err)
	}

	// Create Curio auth
	curioAuth, err := curio.CreateCurioJWTAuthHeader("storacha", params.Identity)
	if err != nil {
		return nil, fmt.Errorf("generating curio JWT: %w", err)
	}

	// Create Curio client
	curioClient := curio.New(http.DefaultClient, pdpURL, curioAuth)

	// Create PDP service
	pdpService, err := pdp.NewRemotePDPService(
		params.PDPDatastore,
		filepath.Join(params.Config.DataDir, "pdp"),
		curioClient,
		params.Config.ProofSet,
		params.Identity,
		params.ReceiptStore,
	)
	if err != nil {
		return nil, fmt.Errorf("creating PDP service: %w", err)
	}

	// Register lifecycle
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return pdpService.Startup(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return pdpService.Shutdown(ctx)
		},
	})

	return pdpService, nil
}

// BlobsParams for blobs service creation
type BlobsParams struct {
	fx.In
	Config          config.UCANServer
	Identity        principal.Signer
	BlobStore       blobstore.Blobstore
	AllocationStore allocationstore.AllocationStore `optional:"true"`
	AllocationDS    datastore.Datastore             `name:"allocation"`
	PublicURL       *url.URL                        `name:"public"`
	Access          access.Access                   `optional:"true"`
	Presigner       presigner.RequestPresigner      `optional:"true"`
	PDP             pdp.PDP                         `optional:"true"`
}

// NewBlobsService creates the blobs service
func NewBlobsService(params BlobsParams) (blobs.Blobs, error) {
	opts := []blobs.Option{}

	// Always use the allocation datastore
	opts = append(opts, blobs.WithDSAllocationStore(params.AllocationDS))

	// Only configure blob storage if PDP is not enabled
	if params.PDP == nil {
		opts = append(opts, blobs.WithBlobstore(params.BlobStore))

		// Configure access
		if params.Access != nil {
			opts = append(opts, blobs.WithAccess(params.Access))
		} else {
			opts = append(opts, blobs.WithPublicURLAccess(*params.PublicURL))
		}

		// Configure presigner
		if params.Presigner != nil {
			opts = append(opts, blobs.WithPresigner(params.Presigner))
		} else {
			opts = append(opts, blobs.WithPublicURLPresigner(params.Identity, *params.PublicURL))
		}
	}

	return blobs.New(opts...)
}

// ClaimsParams for claims service creation
type ClaimsParams struct {
	fx.In
	Config         config.UCANServer
	Identity       principal.Signer
	ClaimStore     claimstore.ClaimStore
	PublisherStore store.PublisherStore
	PublicURL      *url.URL `name:"public"`
}

// NewClaimsService creates the claims service
func NewClaimsService(params ClaimsParams) (claims.Claims, error) {
	// Parse multiaddr from public URL
	var peerAddr multiaddr.Multiaddr
	var err error
	if params.PublicURL.Host == "" {
		u, _ := url.Parse("http://localhost:3000")
		log.Warnf("Public URL not configured, using default: %s", u)
		peerAddr, err = maurl.FromURL(u)
		if err != nil {
			return nil, err
		}
	} else {
		peerAddr, err = maurl.FromURL(params.PublicURL)
		if err != nil {
			return nil, fmt.Errorf("parsing publisher URL as multiaddr: %w", err)
		}

	}

	opts := []claims.Option{}

	// Configure announce URLs
	if len(params.Config.IPNIAnnounceURLs) > 0 {
		var announceURLs []url.URL
		for _, announceURL := range params.Config.IPNIAnnounceURLs {
			u, err := url.Parse(announceURL)
			if err != nil {
				return nil, fmt.Errorf("parsing announce URL: %w", err)
			}
			announceURLs = append(announceURLs, *u)
		}
		opts = append(opts, claims.WithPublisherDirectAnnounce(announceURLs...))
	}

	if params.Config.PDPServerURL != "" {
		pdpServerURL, err := url.Parse(params.Config.PDPServerURL)
		if err != nil {
			return nil, fmt.Errorf("parsing curio URL: %w", err)
		}

		curioAddr, err := maurl.FromURL(pdpServerURL)
		if err != nil {
			return nil, err
		}
		pieceAddr, err := multiaddr.NewMultiaddr("/http-path/" + url.PathEscape("piece/{blobCID}"))
		if err != nil {
			return nil, err
		}
		blobAddr := multiaddr.Join(curioAddr, pieceAddr)
		opts = append(opts, claims.WithPublisherBlobAddress(blobAddr))
	}

	// Configure indexing service
	if params.Config.IndexingServiceDID != "" {
		// Parse peer ID from DID
		indexerDID, err := did.Parse(params.Config.IndexingServiceDID)
		if err != nil {
			return nil, fmt.Errorf("parsing indexing service DID: %w", err)
		}

		indexerURL, err := url.Parse(params.Config.IndexingServiceURL)
		if err != nil {
			return nil, fmt.Errorf("parsing indexing service URL: %w", err)
		}

		opts = append(opts, claims.WithPublisherIndexingServiceConfig(indexerDID, *indexerURL))
	}

	return claims.New(
		params.Identity,
		params.ClaimStore,
		params.PublisherStore,
		peerAddr,
		opts...,
	)
}

// ReplicatorParams for replicator service creation
type ReplicatorParams struct {
	fx.In
	Identity      principal.Signer
	PDP           pdp.PDP `optional:"true"`
	Blobs         blobs.Blobs
	Claims        claims.Claims
	ReceiptStore  receiptstore.ReceiptStore
	UploadService client.Connection
	Config        config.UCANServer
	Lifecycle     fx.Lifecycle
}

// NewReplicatorService creates the replicator service
func NewReplicatorService(params ReplicatorParams) (replicator.Replicator, error) {
	replicatorPath := filepath.Join(params.Config.DataDir, "replicator")
	if err := os.MkdirAll(replicatorPath, 0755); err != nil {
		return nil, fmt.Errorf("creating replicator directory: %w", err)
	}
	repl, err := replicator.New(
		params.Identity,
		params.PDP,
		params.Blobs,
		params.Claims,
		params.ReceiptStore,
		params.UploadService,
		replicatorPath,
	)
	if err != nil {
		return nil, fmt.Errorf("creating replicator service: %w", err)
	}

	// Register lifecycle
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// TODO this is stopping duing tests randomly
			return repl.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return repl.Stop(ctx)
		},
	})

	return repl, nil
}

// StorageServiceParams for main storage service creation
type StorageServiceParams struct {
	fx.In
	Identity      principal.Signer
	Blobs         blobs.Blobs
	Claims        claims.Claims
	PDP           pdp.PDP `optional:"true"`
	ReceiptStore  receiptstore.ReceiptStore
	Replicator    replicator.Replicator
	UploadService client.Connection
}

// ServicesModule provides all service dependencies
var ServicesModule = fx.Module("services",
	// URLs and config parsing
	fx.Provide(NewURLs),

	// External services
	fx.Provide(
		NewAccess,
		NewPresigner,
		NewUploadService,
		NewPDPService,
	),

	// Core services
	fx.Provide(
		NewBlobsService,
		NewClaimsService,
		NewReplicatorService,
	),
)
