package storage

import (
	"context"

	"github.com/storacha/go-ucanto/principal"

	"github.com/storacha/storage/pkg/pdp"
	"github.com/storacha/storage/pkg/pdp/replicator"
	"github.com/storacha/storage/pkg/service/blobs"
	"github.com/storacha/storage/pkg/service/claims"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

type Service interface {
	// ID is the storage service identity, used to sign UCAN invocations and receipts.
	ID() principal.Signer
	// PDP handles PDP aggregation
	PDP() pdp.PDP
	// Blobs provides access to the blobs service.
	Blobs() blobs.Blobs
	// Claims provides access to the claims service.
	Claims() claims.Claims
	// Receipts provides access to receipts
	Receipts() receiptstore.ReceiptStore
}

type Replicator interface {
	Enqueue(ctx context.Context, task *replicator.Task) error
}
