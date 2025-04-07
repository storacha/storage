package capabilities

import (
	"context"

	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
)

type Capabilities interface {
	BlobAccept(context.Context, *BlobAcceptRequest) (*BlobAcceptResponse, error)
	BlobAllocate(context.Context, *BlobAllocateRequest) (*BlobAllocateResponse, error)
}

type BlobAllocateRequest struct {
	Space did.DID
	Blob  blob.Blob
	Cause ucan.Link
}

type BlobAllocateResponse struct {
	Size    uint64
	Address *blob.Address
}
type BlobAcceptRequest struct {
	Space did.DID
	Blob  blob.Blob
	Put   blob.Promise
}

type BlobAcceptResponse struct {
	Claim delegation.Delegation
	// only present when using PDP
	Piece *piece.PieceLink
}
