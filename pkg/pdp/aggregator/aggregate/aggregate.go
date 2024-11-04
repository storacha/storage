package aggregate

import (
	// for go:embed
	_ "embed"
	"errors"
	"fmt"
	"slices"

	"github.com/filecoin-project/go-data-segment/merkletree"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-piece/pkg/piece"
	"github.com/storacha/go-ucanto/core/iterable"
	"github.com/storacha/storage/pkg/pdp/curio"
)

//go:embed aggregate.ipldsch
var aggregateSchema []byte

var aggregateTS *schema.TypeSystem

func init() {
	ts, err := ipld.LoadSchemaBytes(aggregateSchema)
	if err != nil {
		panic(fmt.Errorf("loading blob schema: %w", err))
	}
	aggregateTS = ts
}

func AggregateType() schema.Type {
	return aggregateTS.TypeByName("Aggregate")
}

var ErrIncorrectTree = errors.New("tree leave does not match piece link")

type AggregatePiece struct {
	Link           piece.PieceLink
	InclusionProof merkletree.ProofData
}

type Aggregate struct {
	Root   piece.PieceLink
	Pieces []AggregatePiece
}

func (aggregate Aggregate) ToCurioAddRoot() curio.AddRoot {
	return curio.AddRoot{
		RootCID: aggregate.Root.V1Link().String(),
		Subroots: slices.Collect(iterable.Map(func(aggregatePiece AggregatePiece) curio.SubrootEntry {
			return curio.SubrootEntry{SubrootCID: aggregatePiece.Link.V1Link().String()}
		}, slices.Values(aggregate.Pieces))),
	}
}
