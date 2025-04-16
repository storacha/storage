package pdp

import (
	"context"
	"database/sql"
	"errors"
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
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/go-ucanto/ucan"
	"gorm.io/driver/postgres"

	"github.com/storacha/storage/pkg/pdp/aggregator"
	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue"
	"github.com/storacha/storage/pkg/pdp/api"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/pdp/pieceadder"
	"github.com/storacha/storage/pkg/pdp/piecefinder"
	"github.com/storacha/storage/pkg/pdp/service"
	"github.com/storacha/storage/pkg/pdp/store"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/keystore"
	"github.com/storacha/storage/pkg/store/receiptstore"
	"github.com/storacha/storage/pkg/wallet"
)

type PDPService struct {
	aggregator  aggregator.Aggregator
	pieceFinder piecefinder.PieceFinder
	pieceAdder  pieceadder.PieceAdder
	startFuncs  []func(ctx context.Context) error
	closeFuncs  []func(ctx context.Context) error
}

func (p *PDPService) Aggregator() aggregator.Aggregator {
	return p.aggregator
}

func (p *PDPService) PieceAdder() pieceadder.PieceAdder {
	return p.pieceAdder
}

func (p *PDPService) PieceFinder() piecefinder.PieceFinder {
	return p.pieceFinder
}

func (p *PDPService) Startup(ctx context.Context) error {
	var err error
	for _, startFunc := range p.startFuncs {
		err = errors.Join(startFunc(ctx))
	}
	return err
}

func (p *PDPService) Shutdown(ctx context.Context) error {
	var err error
	for _, closeFunc := range p.closeFuncs {
		err = errors.Join(closeFunc(ctx))
	}
	return err
}

var _ PDP = (*PDPService)(nil)

func NewLocalPDPService(
	ctx context.Context,
	dataDir string,
	lotusClientAddr string,
	ethClientAddr string,
	dbConfig string,
	proofSet uint64,
	address common.Address,
	issuer principal.Signer,
	receiptStore receiptstore.ReceiptStore,
) (*PDPService, error) {
	ds, err := leveldb.NewDatastore(dataDir, nil)
	if err != nil {
		return nil, err
	}
	jobqueueDB, err := jobqueue.NewInMemoryDB()
	if err != nil {
		return nil, err
	}
	blobStore := blobstore.NewDsBlobstore(namespace.Wrap(ds, datastore.NewKey("blobs")))
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
	// TODO our current in process endpoint, later create a client without http stuffs.
	localEndpoint, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint URL: %w", err)
	}
	// NB: Auth not required
	localPDPClient := curio.New(http.DefaultClient, localEndpoint, "")
	agg, err := aggregator.NewLocal(ds, jobqueueDB, localPDPClient, proofSet, issuer, receiptStore)
	if err != nil {
		return nil, err
	}
	lotusURL, err := url.Parse(lotusClientAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing lotus client address: %w", err)
	}
	if lotusURL.Scheme != "ws" {
		return nil, fmt.Errorf("lotus client address must be 'ws'")
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
	return &PDPService{
		aggregator:  agg,
		pieceFinder: piecefinder.NewCurioFinder(localPDPClient),
		pieceAdder:  pieceadder.NewCurioAdder(localPDPClient),
		startFuncs: []func(ctx context.Context) error{
			func(ctx context.Context) error {
				if err := svr.Start(fmt.Sprintf(":%s", localEndpoint.Port())); err != nil {
					return fmt.Errorf("starting local pdp server: %w", err)
				}
				if err := agg.Startup(ctx); err != nil {
					return fmt.Errorf("failed to start aggregator: %w", err)
				}
				if err := pdpService.Start(ctx); err != nil {
					return fmt.Errorf("starting pdp service: %w", err)
				}
				return nil
			},
		},
		closeFuncs: []func(context.Context) error{
			func(ctx context.Context) error {
				var errs error
				if err := pdpService.Stop(ctx); err != nil {
					errs = multierror.Append(errs, err)
				}
				if err := svr.Shutdown(ctx); err != nil {
					errs = multierror.Append(errs, err)
				}
				agg.Shutdown(ctx)
				chainClientCloser()
				ethClient.Close()
				return errs
			},
		},
	}, nil
}

func NewRemotePDPService(
	ds datastore.Datastore,
	db *sql.DB,
	client *curio.Client,
	proofSet uint64,
	issuer ucan.Signer,
	receiptStore receiptstore.ReceiptStore,
) (*PDPService, error) {
	aggregator, err := aggregator.NewLocal(ds, db, client, proofSet, issuer, receiptStore)
	if err != nil {
		return nil, fmt.Errorf("creating local aggregator: %w", err)
	}
	return &PDPService{
		aggregator:  aggregator,
		pieceFinder: piecefinder.NewCurioFinder(client),
		pieceAdder:  pieceadder.NewCurioAdder(client),
		startFuncs: []func(ctx context.Context) error{
			func(ctx context.Context) error {
				return aggregator.Startup(ctx)
			},
		},
		closeFuncs: []func(context.Context) error{
			func(ctx context.Context) error { aggregator.Shutdown(ctx); return nil },
		},
	}, nil
}
