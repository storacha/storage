package claimstore

import (
	"context"

	"github.com/multiformats/go-multihash"
)

type ClaimStore interface {
	Get(context.Context, multihash.Multihash) ([]byte, error)
	Put(context.Context, multihash.Multihash) error
}
