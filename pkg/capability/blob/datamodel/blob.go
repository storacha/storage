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

func AllocateOkType() schema.Type {
	return blobTS.TypeByName("AllocateOk")
}

func AcceptCaveatsType() schema.Type {
	return blobTS.TypeByName("AcceptCaveats")
}

func AcceptOkType() schema.Type {
	return blobTS.TypeByName("AcceptOk")
}

type BlobModel struct {
	Digest []byte
	Size   int64
}

type AllocateCaveatsModel struct {
	Space []byte
	Blob  BlobModel
	Cause ipld.Link
}

type HeadersModel struct {
	Keys   []string
	Values map[string]string
}

type AddressModel struct {
	Url     string
	Headers HeadersModel
	Expires int64
}

type AllocateOkModel struct {
	Size    int64
	Address *AddressModel
}

type ResultModel struct {
	Selector string
	Link     ipld.Link
}

type PromiseModel struct {
	UcanAwait ResultModel
}

type AcceptCaveatsModel struct {
	Space   []byte
	Blob    BlobModel
	Expires int64
	Put     PromiseModel
}

type AcceptOkModel struct {
	Site ipld.Link
}
