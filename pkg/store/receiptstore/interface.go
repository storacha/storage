package receiptstore

import (
	"context"

	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/ucan"
)

// ReceiptStore stores UCAN invocation receipts.
type ReceiptStore interface {
	// Get retrieves a receipt by it's CID.
	Get(context.Context, ucan.Link) (receipt.AnyReceipt, error)
	// Get retrieves a receipt by "ran" CID.
	GetByRan(context.Context, ucan.Link) (receipt.AnyReceipt, error)
	// Put adds or replaces a receipt in the store.
	Put(context.Context, receipt.AnyReceipt) error
}
