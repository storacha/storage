package receiptstore

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
)

const receiptsPrefix = "receipts/"
const ranLinkIndexPrefix = "ranLinkIndex/"

func NewDsReceiptStore(ds datastore.Datastore) (ReceiptStore, error) {
	receipts := namespace.Wrap(ds, datastore.NewKey(receiptsPrefix))
	store := store.SimpleStoreFromDatastore(receipts)
	ranLinkIndex := namespace.Wrap(ds, datastore.NewKey(ranLinkIndexPrefix))
	return NewReceiptStore(store, &dsRanLinkIndex{ranLinkIndex})
}

type dsRanLinkIndex struct {
	ds datastore.Datastore
}

// Get implements RanLinkIndex.
func (d *dsRanLinkIndex) Get(ctx context.Context, ran datamodel.Link) (datamodel.Link, error) {
	data, err := d.ds.Get(ctx, datastore.NewKey(ran.String()))
	if err != nil {
		return nil, err
	}
	c, err := cid.Cast(data)
	if err != nil {
		return nil, err
	}
	return cidlink.Link{Cid: c}, nil
}

// Put implements RanLinkIndex.
func (d *dsRanLinkIndex) Put(ctx context.Context, ran datamodel.Link, lnk datamodel.Link) error {
	return d.ds.Put(ctx, datastore.NewKey(ran.String()), []byte(lnk.Binary()))
}

var _ RanLinkIndex = (*dsRanLinkIndex)(nil)
