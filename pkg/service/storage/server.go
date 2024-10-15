package storage

import (
	"context"
	"errors"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/go-ucanto/server"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/storage/pkg/capability/blob"
	"github.com/storacha/storage/pkg/store"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
)

var log = logging.Logger("storage")

func NewServer(id principal.Signer, storageService Service) (server.ServerView, error) {
	return server.NewServer(
		id,
		server.WithServiceMethod(
			blob.AllocateAbility,
			server.Provide(
				blob.Allocate,
				func(cap ucan.Capability[blob.AllocateCaveats], inv invocation.Invocation, ctx server.InvocationContext) (blob.AllocateOk, receipt.Effects, error) {
					log.Infof("%s z%s => %s", blob.AllocateAbility, cap.Nb().Blob.Digest.B58String(), cap.Nb().Space)

					_, err := storageService.Blobs().Get(context.Background(), cap.Nb().Blob.Digest)
					if err == nil {
						return blob.AllocateOk{Size: 0}, nil, nil
					}
					if !errors.Is(err, store.ErrNotFound) {
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					err = storageService.Allocations().Put(context.Background(), allocation.Allocation{
						Space:   cap.Nb().Space,
						Blob:    allocation.Blob(cap.Nb().Blob),
						Expires: 0,
						Cause:   inv.Link(),
					})
					if err != nil {
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					return blob.AllocateOk{}, nil, nil
				},
			),
		),
		server.WithServiceMethod(
			blob.AcceptAbility,
			server.Provide(
				blob.Accept,
				func(cap ucan.Capability[blob.AcceptCaveats], inv invocation.Invocation, ctx server.InvocationContext) (blob.AcceptOk, receipt.Effects, error) {
					log.Infof("%s z%s => %s", blob.AcceptAbility, cap.Nb().Blob.Digest.B58String(), cap.Nb().Space)
					return blob.AcceptOk{}, nil, nil
				},
			),
		),
	)
}
