package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/go-ucanto/server"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/storage/pkg/capability"
	"github.com/storacha/storage/pkg/capability/blob"
	"github.com/storacha/storage/pkg/store"
	"github.com/storacha/storage/pkg/store/allocationstore/allocation"
)

var log = logging.Logger("storage")

const maxUploadSize = 127 * (1 << 25)

func NewServer(id principal.Signer, storageService Service) (server.ServerView, error) {
	return server.NewServer(
		id,
		server.WithServiceMethod(
			blob.AllocateAbility,
			server.Provide(
				blob.Allocate,
				func(cap ucan.Capability[blob.AllocateCaveats], inv invocation.Invocation, iCtx server.InvocationContext) (blob.AllocateOk, receipt.Effects, error) {
					log.Infof("%s z%s => %s", blob.AllocateAbility, cap.Nb().Blob.Digest.B58String(), cap.Nb().Space)
					ctx := context.TODO()

					// TODO: restrict invocation to authority (storacha service)

					// check if we already have an allcoation for the blob in this space
					allocs, err := storageService.Allocations().List(ctx, cap.Nb().Blob.Digest)
					if err != nil {
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					for _, a := range allocs {
						// if we find an allocation, check if we have the blob.
						if a.Space == cap.Nb().Space {
							_, err := storageService.Blobs().Get(ctx, cap.Nb().Blob.Digest)
							if err == nil {
								// if we have it, it does not need upload
								return blob.AllocateOk{Size: 0}, nil, nil
							}
							if !errors.Is(err, store.ErrNotFound) {
								return blob.AllocateOk{}, nil, failure.FromError(err)
							}
						}
					}

					if cap.Nb().Blob.Size > maxUploadSize {
						return blob.AllocateOk{}, nil, capability.Failure{
							Name:    "BlobSizeOutsideOfSupportedRange",
							Message: fmt.Sprintf("Blob of %d bytes, exceeds size limit of %d bytes", cap.Nb().Blob.Size, maxUploadSize),
						}
					}

					expiresIn := uint64(60 * 60 * 24) // 1 day
					expiresAt := uint64(time.Now().Unix()) + expiresIn
					url, headers, err := storageService.Presigner().SignUploadURL(ctx, cap.Nb().Blob.Digest, cap.Nb().Blob.Size, expiresIn)
					if err != nil {
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					err = storageService.Allocations().Put(ctx, allocation.Allocation{
						Space:   cap.Nb().Space,
						Blob:    allocation.Blob(cap.Nb().Blob),
						Expires: expiresAt,
						Cause:   inv.Link(),
					})
					if err != nil {
						return blob.AllocateOk{}, nil, failure.FromError(err)
					}

					return blob.AllocateOk{
						Size: cap.Nb().Blob.Size,
						Address: &blob.Address{
							URL:     url,
							Headers: headers,
							Expires: expiresAt,
						},
					}, nil, nil
				},
			),
		),
		server.WithServiceMethod(
			blob.AcceptAbility,
			server.Provide(
				blob.Accept,
				func(cap ucan.Capability[blob.AcceptCaveats], inv invocation.Invocation, iCtx server.InvocationContext) (blob.AcceptOk, receipt.Effects, error) {
					log.Infof("%s z%s => %s", blob.AcceptAbility, cap.Nb().Blob.Digest.B58String(), cap.Nb().Space)
					ctx := context.TODO()

					// TODO: restrict invocation to authority (storacha service)

					_, err := storageService.Blobs().Get(ctx, cap.Nb().Blob.Digest)
					if err != nil {
						if errors.Is(err, store.ErrNotFound) {
							return blob.AcceptOk{}, nil, capability.Failure{
								Name:    "AllocatedMemoryHadNotBeenWrittenTo",
								Message: "Blob not found",
							}
						}
						return blob.AcceptOk{}, nil, failure.FromError(err)
					}

					return blob.AcceptOk{}, nil, nil
				},
			),
		),
	)
}
