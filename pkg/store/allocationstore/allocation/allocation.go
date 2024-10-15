package allocation

import (
	"bytes"
	"fmt"

	"github.com/ipld/go-ipld-prime/codec"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	adm "github.com/storacha/storage/pkg/store/allocationstore/datamodel"
)

type Allocation struct {
	// Space is the DID of the space this data was allocated for.
	Space did.DID
	// Digest is the hash of the data.
	Digest multihash.Multihash
	// Size of the data in bytes.
	Size uint64
	// Expiration is the time (in seconds since unix epoch) at which the
	// allocation becomes invalid and can no longer be accepted.
	Expiration uint64
	// Cause is a link to the UCAN that requested the allocation.
	Cause ucan.Link
}

func (al Allocation) ToIPLD() (datamodel.Node, error) {
	md := &adm.AllocationModel{
		Space:      al.Space.Bytes(),
		Digest:     al.Digest,
		Size:       int64(al.Size),
		Expiration: int64(al.Expiration),
		Cause:      al.Cause,
	}
	return ipld.WrapWithRecovery(md, adm.AllocationType())
}

func Encode(alloc Allocation, enc codec.Encoder) ([]byte, error) {
	n, err := alloc.ToIPLD()
	if err != nil {
		return nil, fmt.Errorf("encoding to IPLD: %w", err)
	}

	if enc == nil {
		enc = dagcbor.Encode
	}

	buf := bytes.NewBuffer([]byte{})
	err = enc(n, buf)
	if err != nil {
		return nil, fmt.Errorf("encoding to data format: %w", err)
	}

	return buf.Bytes(), nil
}

func Decode(data []byte, dec codec.Decoder) (Allocation, error) {
	if dec == nil {
		dec = dagcbor.Decode
	}

	nb := bindnode.Prototype((*adm.AllocationModel)(nil), adm.AllocationType()).NewBuilder()

	err := dec(nb, bytes.NewBuffer(data))
	if err != nil {
		return Allocation{}, fmt.Errorf("decoding from data format: %w", err)
	}

	nd := nb.Build()
	model := bindnode.Unwrap(nd).(*adm.AllocationModel)

	space, err := did.Decode(model.Space)
	if err != nil {
		return Allocation{}, fmt.Errorf("decoding space DID: %w", err)
	}

	digest, err := multihash.Cast(model.Digest)
	if err != nil {
		return Allocation{}, fmt.Errorf("decoding digest: %w", err)
	}

	return Allocation{
		Space:      space,
		Digest:     digest,
		Size:       uint64(model.Size),
		Expiration: uint64(model.Expiration),
		Cause:      model.Cause,
	}, nil
}
