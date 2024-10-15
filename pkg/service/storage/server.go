package storage

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/go-ucanto/server"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/storage/pkg/capability/blob"
)

var log = logging.Logger("storage")

func NewServer(id principal.Signer, storageService Service) (server.ServerView, error) {
	return server.NewServer(
		id,
		server.WithServiceMethod(
			blob.AllocateAbility,
			server.Provide(
				blob.Allocate,
				func(cap ucan.Capability[blob.AllocateCaveats], inv invocation.Invocation, ctx server.InvocationContext) (blob.AllocateSuccess, receipt.Effects, error) {
					log.Fatal("not implemented")
					return blob.AllocateSuccess{}, nil, nil
				},
			),
		),
		server.WithServiceMethod(
			blob.AcceptAbility,
			server.Provide(
				blob.Accept,
				func(cap ucan.Capability[blob.AcceptCaveats], inv invocation.Invocation, ctx server.InvocationContext) (blob.AcceptSuccess, receipt.Effects, error) {
					log.Fatal("not implemented")
					return blob.AcceptSuccess{}, nil, nil
				},
			),
		),
	)
}
