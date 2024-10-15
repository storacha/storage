package datamodel

import (
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
)

//go:embed blob.ipldsch
var blobSchema []byte

var blobTS *schema.TypeSystem

func init() {
	ts, err := ipld.LoadSchemaBytes(blobSchema)
	if err != nil {
		panic(fmt.Errorf("loading blob schema: %w", err))
	}
	blobTS = ts
}

func AllocateCaveatsType() schema.Type {
	return blobTS.TypeByName("AllocateCaveats")
}

func AcceptCaveatsType() schema.Type {
	return blobTS.TypeByName("AcceptCaveats")
}

type AllocateCaveatsModel struct {
	Space []byte
}

type AcceptCaveatsModel struct {
	Space []byte
}
