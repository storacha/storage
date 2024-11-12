package pdp

import (
	"context"
	"errors"

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
	startFuncs  []func() error
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

func (p *PDPService) Startup() error {
	var err error
	for _, startFunc := range p.startFuncs {
		err = errors.Join(startFunc())
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

func NewLocal(ds datastore.Datastore, client *curio.Client, proofSet uint64, issuer ucan.Signer, receiptStore receiptstore.ReceiptStore) *PDPService {
	aggregator := aggregator.NewLocal(ds, client, proofSet, issuer, receiptStore)
	return &PDPService{
		aggregator:  aggregator,
		pieceFinder: piecefinder.NewCurioFinder(client),
		pieceAdder:  pieceadder.NewCurioAdder(client),
		startFuncs: []func() error{
			func() error { aggregator.Startup(); return nil },
		},
		closeFuncs: []func(context.Context) error{
			func(ctx context.Context) error { aggregator.Shutdown(ctx); return nil },
		},
	}
}
