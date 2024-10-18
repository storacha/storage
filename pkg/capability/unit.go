package capability

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

// Unit is a success type that can be used when there is no data to return from
// a capability handler.
type Unit struct{}

func (u Unit) ToIPLD() (datamodel.Node, error) {
	np := basicnode.Prototype.Any
	nb := np.NewBuilder()
	ma, err := nb.BeginMap(0)
	if err != nil {
		return nil, err
	}
	ma.Finish()
	return nb.Build(), nil
}
