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

type BlobModel struct {
	Digest []byte
	Size   int64
}

type AllocationModel struct {
	Space   []byte
	Blob    BlobModel
	Expires int64
	Cause   ipld.Link
}
