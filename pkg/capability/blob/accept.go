package blob

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
	bdm "github.com/storacha/storage/pkg/capability/blob/datamodel"
)

const AcceptAbility = "blob/accept"

type Result struct {
	Selector string
	Link     ipld.Link
}

type Promise struct {
	UcanAwait Result
}

type AcceptCaveats struct {
	Space   did.DID
	Blob    Blob
	Expires uint64
	Put     Promise
}

func (ac AcceptCaveats) ToIPLD() (datamodel.Node, error) {
	md := &bdm.AcceptCaveatsModel{
		Space: ac.Space.Bytes(),
		Blob: bdm.BlobModel{
			Digest: ac.Blob.Digest,
			Size:   int64(ac.Blob.Size),
		},
		Expires: int64(ac.Expires),
		Put: bdm.PromiseModel{
			UcanAwait: bdm.ResultModel{
				Selector: ac.Put.UcanAwait.Selector,
				Link:     ac.Put.UcanAwait.Link,
			},
		},
	}
	return ipld.WrapWithRecovery(md, bdm.AcceptCaveatsType())
}

type AcceptOk struct {
	Site ucan.Link
}

func (ao AcceptOk) ToIPLD() (datamodel.Node, error) {
	md := &bdm.AcceptOkModel{Site: ao.Site}
	return ipld.WrapWithRecovery(md, bdm.AcceptOkType())
}

var Accept = validator.NewCapability(
	AcceptAbility,
	schema.DIDString(),
	schema.Mapped(schema.Struct[bdm.AcceptCaveatsModel](bdm.AcceptCaveatsType(), nil), func(model bdm.AcceptCaveatsModel) (AcceptCaveats, failure.Failure) {
		space, err := did.Decode(model.Space)
		if err != nil {
			return AcceptCaveats{}, failure.FromError(fmt.Errorf("decoding space DID: %w", err))
		}

		digest, err := multihash.Cast(model.Blob.Digest)
		if err != nil {
			return AcceptCaveats{}, failure.FromError(fmt.Errorf("decoding digest: %w", err))
		}

		return AcceptCaveats{
			Space: space,
			Blob: Blob{
				Digest: digest,
				Size:   uint64(model.Blob.Size),
			},
			Expires: uint64(model.Expires),
			Put: Promise{
				UcanAwait: Result{
					Selector: model.Put.UcanAwait.Selector,
					Link:     model.Put.UcanAwait.Link,
				},
			},
		}, nil
	}),
	validator.DefaultDerives,
)
