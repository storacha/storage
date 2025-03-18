package pdp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/storage/pkg/pdp/aggregator"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/pdp/pieceadder"
	"github.com/storacha/storage/pkg/pdp/piecefinder"
	"github.com/storacha/storage/pkg/store/receiptstore"
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

func NewLocal(
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
