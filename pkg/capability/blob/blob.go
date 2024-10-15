package blob

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
	bdm "github.com/storacha/storage/pkg/capability/blob/datamodel"
)

const AllocateAbility = "blob/allocate"

type AllocateCaveats struct{}

func (ac AllocateCaveats) ToIPLD() (datamodel.Node, error) {
	return nil, fmt.Errorf("not implemented")
}

type AllocateSuccess struct{}

func (as AllocateSuccess) ToIPLD() (datamodel.Node, error) {
	return nil, fmt.Errorf("not implemented")
}

var Allocate = validator.NewCapability[AllocateCaveats](
	AllocateAbility,
	schema.DIDString(),
	schema.Struct(bdm.AllocateCaveatsType(), nil),
	validator.DefaultDerives,
)

const AcceptAbility = "blob/accept"

type AcceptCaveats struct{}

func (ac AcceptCaveats) ToIPLD() (datamodel.Node, error) {
	return nil, fmt.Errorf("not implemented")
}

type AcceptSuccess struct{}

func (as AcceptSuccess) ToIPLD() (datamodel.Node, error) {
	return nil, fmt.Errorf("not implemented")
}

var Accept = validator.NewCapability[AcceptCaveats](
	AcceptAbility,
	schema.DIDString(),
	schema.Struct(bdm.AcceptCaveatsType(), nil),
	validator.DefaultDerives,
)
