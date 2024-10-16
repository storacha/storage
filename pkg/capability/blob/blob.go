package blob

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"

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

const AllocateAbility = "blob/allocate"

type Blob struct {
	Digest multihash.Multihash
	Size   uint64
}

type AllocateCaveats struct {
	Space did.DID
	Blob  Blob
	Cause ucan.Link
}

func (ac AllocateCaveats) ToIPLD() (datamodel.Node, error) {
	md := &bdm.AllocateCaveatsModel{
		Space: ac.Space.Bytes(),
		Blob: bdm.BlobModel{
			Digest: ac.Blob.Digest,
			Size:   int64(ac.Blob.Size),
		},
		Cause: ac.Cause,
	}
	return ipld.WrapWithRecovery(md, bdm.AllocateCaveatsType())
}

type Address struct {
	URL     url.URL
	Headers http.Header
	Expires uint64
}

type AllocateOk struct {
	Size    uint64
	Address *Address
}

func headersToMap(h http.Header) (map[string]string, error) {
	headers := map[string]string{}
	for k, v := range h {
		if len(v) > 1 {
			return nil, fmt.Errorf("unsupported multiple values in header: %s", k)
		}
		headers[k] = v[0]
	}
	return headers, nil
}

func (ao AllocateOk) ToIPLD() (datamodel.Node, error) {
	md := &bdm.AllocateOkModel{Size: int64(ao.Size)}
	if ao.Address != nil {
		keys := slices.Collect(maps.Keys(ao.Address.Headers))
		slices.Sort(keys)

		headers, err := headersToMap(ao.Address.Headers)
		if err != nil {
			return nil, err
		}

		md.Address = &bdm.AddressModel{
			URL: ao.Address.URL.String(),
			Headers: bdm.HeadersModel{
				Keys:   keys,
				Values: headers,
			},
			Expires: int64(ao.Address.Expires),
		}
	}
	return ipld.WrapWithRecovery(md, bdm.AllocateOkType())
}

var Allocate = validator.NewCapability(
	AllocateAbility,
	schema.DIDString(),
	schema.Mapped(schema.Struct[bdm.AllocateCaveatsModel](bdm.AllocateCaveatsType(), nil), func(model bdm.AllocateCaveatsModel) (AllocateCaveats, failure.Failure) {
		space, err := did.Decode(model.Space)
		if err != nil {
			return AllocateCaveats{}, failure.FromError(fmt.Errorf("decoding space DID: %w", err))
		}

		digest, err := multihash.Cast(model.Blob.Digest)
		if err != nil {
			return AllocateCaveats{}, failure.FromError(fmt.Errorf("decoding digest: %w", err))
		}

		return AllocateCaveats{
			Space: space,
			Blob: Blob{
				Digest: digest,
				Size:   uint64(model.Blob.Size),
			},
			Cause: model.Cause,
		}, nil
	}),
	validator.DefaultDerives,
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
	schema.Mapped(schema.Struct[bdm.AcceptCaveatsModel](bdm.AllocateCaveatsType(), nil), func(model bdm.AcceptCaveatsModel) (AcceptCaveats, failure.Failure) {
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
