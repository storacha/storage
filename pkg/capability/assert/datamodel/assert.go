package datamodel

import (
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
)

//go:embed assert.ipldsch
var assert []byte

var assertTS *schema.TypeSystem

func init() {
	ts, err := ipld.LoadSchemaBytes(assert)
	if err != nil {
		panic(fmt.Errorf("loading assert schema: %w", err))
	}
	assertTS = ts
}

func LocationCaveatsType() schema.Type {
	return assertTS.TypeByName("LocationCaveats")
}

type RangeModel struct {
	Offset uint64
	Length *uint64
}

type LocationCaveatsModel struct {
	Content  []byte
	Location []string
	Range    *RangeModel
	Space    []byte
}
