package aggregate

import (
	// for go:embed
	_ "embed"
	"errors"
	"fmt"
	"slices"

	"github.com/filecoin-project/go-data-segment/merkletree"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/core/iterable"
	"go.uber.org/zap/zapcore"

	"github.com/storacha/piri/pkg/pdp/curio"
)

//go:embed aggregate.ipldsch
var aggregateSchema []byte

var aggregateTS *schema.TypeSystem

func init() {
	ts, err := types.LoadSchemaBytes(aggregateSchema)
	if err != nil {
		panic(fmt.Errorf("loading blob schema: %w", err))
	}
	aggregateTS = ts
}

func AggregateType() schema.Type {
	return aggregateTS.TypeByName("Aggregate")
}

func PieceLinkType() schema.Type {
	return aggregateTS.TypeByName("PieceLink")
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

// MarshalLogObject makes Aggregate implement the zapcore.ObjectMarshaler interface
func (a Aggregate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("root", a.Root.Link().String())

	// One approach is to encode the slice as a separate array:
	enc.AddArray("pieces", zapcore.ArrayMarshalerFunc(func(arr zapcore.ArrayEncoder) error {
		for _, p := range a.Pieces {
			// Log each piece as a string, or you can nest objects here.
			arr.AppendString(p.Link.Link().String())
		}
		return nil
	}))

	return nil
}

func (aggregate Aggregate) ToCurioAddRoot() curio.AddRootRequest {
	return curio.AddRootRequest{
		RootCID: aggregate.Root.V1Link().String(),
		Subroots: slices.Collect(iterable.Map(func(aggregatePiece AggregatePiece) curio.SubrootEntry {
			return curio.SubrootEntry{SubrootCID: aggregatePiece.Link.V1Link().String()}
		}, slices.Values(aggregate.Pieces))),
	}
}
