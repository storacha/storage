package pdp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/hashicorp/go-multierror"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"gorm.io/driver/postgres"

	"github.com/storacha/storage/pkg/pdp/api"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/pdp/pieceadder"
	"github.com/storacha/storage/pkg/pdp/piecefinder"
	"github.com/storacha/storage/pkg/pdp/service"
	"github.com/storacha/storage/pkg/pdp/store"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/keystore"
	"github.com/storacha/storage/pkg/wallet"
)

type Server struct {
	pieceFinder piecefinder.PieceFinder
	pieceAdder  pieceadder.PieceAdder
	startFuncs  []func(ctx context.Context) error
	stopFuncs   []func(ctx context.Context) error
}

func (s *Server) Start(ctx context.Context) error {
	for _, startFunc := range s.startFuncs {
		if err := startFunc(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	var errs error
	for _, stopFunc := range s.stopFuncs {
		if err := stopFunc(ctx); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func NewServer(
	ctx context.Context,
	dataDir string,
	port int,
	lotusClientAddr string,
	ethClientAddr string,
	dbConfig string,
	address common.Address,
) (*Server, error) {
	ds, err := leveldb.NewDatastore(dataDir, nil)
	if err != nil {
		return nil, err
	}
	blobStore := blobstore.NewTODO_DsBlobstore(namespace.Wrap(ds, datastore.NewKey("blobs")))
	stashStore, err := store.NewStashStore(path.Join(dataDir, "stash"))
	if err != nil {
		return nil, err
	}
	keyStore, err := keystore.NewKeyStore(ds)
	if err != nil {
		return nil, err
	}
	wlt, err := wallet.NewWallet(keyStore)
	if err != nil {
		return nil, err
	}
	if has, err := wlt.Has(ctx, address); err != nil {
		return nil, fmt.Errorf("failed to read wallet for address %s: %w", address, err)
	} else if !has {
		return nil, fmt.Errorf("wallet for address %s not found", address)
	}
	// TODO our current in process endpoint, later create a client without http stuffs.
	localEndpoint, err := url.Parse(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint URL: %w", err)
	}
	// NB: Auth not required
	localPDPClient := curio.New(http.DefaultClient, localEndpoint, "")
	lotusURL, err := url.Parse(lotusClientAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing lotus client address: %w", err)
	}
	if lotusURL.Scheme != "ws" && lotusURL.Scheme != "wss" {
		return nil, fmt.Errorf("lotus client address must be 'ws' or 'wss', got %s", lotusURL.Scheme)
	}
	chainClient, chainClientCloser, err := client.NewFullNodeRPCV1(ctx, lotusURL.String(), nil)
	if err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(ethClientAddr)
	if err != nil {
		return nil, fmt.Errorf("connecting to eth client: %w", err)
	}
	dialector := postgres.Open(dbConfig)
	pdpService, err := service.NewPDPService(dialector, address, wlt, blobStore, stashStore, chainClient, ethClient)
	if err != nil {
		return nil, fmt.Errorf("creating pdp service: %w", err)
	}

	pdpAPI := &api.PDP{Service: pdpService}
	svr := api.NewServer(pdpAPI)
	return &Server{
		pieceFinder: piecefinder.NewCurioFinder(localPDPClient),
		pieceAdder:  pieceadder.NewCurioAdder(localPDPClient),
		startFuncs: []func(ctx context.Context) error{
			func(ctx context.Context) error {
				if err := svr.Start(fmt.Sprintf(":%s", localEndpoint.Port())); err != nil {
					return fmt.Errorf("starting local pdp server: %w", err)
				}
				if err := pdpService.Start(ctx); err != nil {
					return fmt.Errorf("starting pdp service: %w", err)
				}
				return nil
			},
		},
		stopFuncs: []func(context.Context) error{
			func(ctx context.Context) error {
				var errs error
				if err := svr.Shutdown(ctx); err != nil {
					errs = multierror.Append(errs, err)
				}
				if err := pdpService.Stop(ctx); err != nil {
					errs = multierror.Append(errs, err)
				}
				chainClientCloser()
				ethClient.Close()
				return errs
			},
		},
	}, nil

}
