package capability

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

type Failure struct {
	Name    string
	Message string
}

func (f Failure) Error() string {
	return f.Message
}

func (f Failure) ToIPLD() (datamodel.Node, error) {
	np := basicnode.Prototype.Any
	nb := np.NewBuilder()
	ma, err := nb.BeginMap(2)
	if err != nil {
		return nil, err
	}
	ma.AssembleKey().AssignString("name")
	ma.AssembleValue().AssignString(f.Name)
	ma.AssembleKey().AssignString("message")
	ma.AssembleValue().AssignString(f.Message)
	ma.Finish()
	return nb.Build(), nil
}
