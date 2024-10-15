package datamodel

import (
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
)

//go:embed allocation.ipldsch
var allocationSchema []byte

var allocationTS *schema.TypeSystem

func init() {
	ts, err := ipld.LoadSchemaBytes(allocationSchema)
	if err != nil {
		panic(fmt.Errorf("loading allocation schema: %w", err))
	}
	allocationTS = ts
}

func AllocationType() schema.Type {
	return allocationTS.TypeByName("Allocation")
}

type AllocationModel struct {
	Space  []byte
	Digest []byte
	Size   int
	Cause  ipld.Link
}
