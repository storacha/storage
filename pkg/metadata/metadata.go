// Extracted location commitment from:
// https://github.com/storacha/indexing-service/blob/88f309355262c687adeb1715738b951eb277b882/pkg/metadata/metadata.go
// TODO: metadata as seperate package? import from indexing-servce?
package metadata

import (
	"bytes"
	// for import
	_ "embed"
	"fmt"
	"io"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	ipnimd "github.com/ipni/go-libipni/metadata"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-varint"
)

var (
	_ ipnimd.Protocol = (*LocationCommitmentMetadata)(nil)

	//go:embed metadata.ipldsch
	schemaBytes                []byte
	locationCommitmentMetadata schema.TypedPrototype
)

// metadata identifiers
// currently we just use experimental codecs for now

// LocationCommitmentID is the multicodec for location commitments
const LocationCommitmentID = 0x3E0002

var nodePrototypes = map[multicodec.Code]schema.TypedPrototype{}

var MetadataContext ipnimd.MetadataContext

func init() {
	typeSystem, err := ipld.LoadSchemaBytes(schemaBytes)
	if err != nil {
		panic(fmt.Errorf("failed to load schema: %w", err))
	}
	locationCommitmentMetadata = bindnode.Prototype((*LocationCommitmentMetadata)(nil), typeSystem.TypeByName("LocationCommitmentMetadata"))
	nodePrototypes[LocationCommitmentID] = locationCommitmentMetadata
}

func init() {
	mdctx := ipnimd.Default
	mdctx = mdctx.WithProtocol(LocationCommitmentID, func() ipnimd.Protocol { return &LocationCommitmentMetadata{} })
	MetadataContext = mdctx
}

type HasClaim interface {
	GetClaim() ipld.Link
}

type Range struct {
	Offset uint64
	Length *uint64
}

// LocationCommitmentMetadata represents metadata for an location commitment.
type LocationCommitmentMetadata struct {
	// Claim indicates the CID of the claim - the claim should be fetchable by
	// combining the HTTP multiaddr of the provider with the claim CID.
	Claim ipld.Link
	// Expiration as unix epoch in seconds.
	Expiration int64
	// Shard is an optional alternate CID to use to lookup this location --
	// if the looked up shard is part of a larger shard
	Shard *ipld.Link
	// Range is an optional byte range within a shard.
	Range *Range
}

func (l *LocationCommitmentMetadata) ID() multicodec.Code {
	return LocationCommitmentID
}
func (l *LocationCommitmentMetadata) MarshalBinary() ([]byte, error) {
	return marshalBinary(l)
}
func (l *LocationCommitmentMetadata) UnmarshalBinary(data []byte) error {
	return unmarshalBinary(l, data)
}
func (l *LocationCommitmentMetadata) ReadFrom(r io.Reader) (n int64, err error) {
	return readFrom(l, r)
}
func (l *LocationCommitmentMetadata) GetClaim() ipld.Link {
	return l.Claim
}

type hasID[T any] interface {
	*T
	ID() multicodec.Code
}

func marshalBinary(metadata ipnimd.Protocol) ([]byte, error) {
	buf := bytes.NewBuffer(varint.ToUvarint(uint64(metadata.ID())))
	nd := bindnode.Wrap(metadata, nodePrototypes[metadata.ID()].Type())
	if err := dagcbor.Encode(nd, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshalBinary[PT hasID[T], T any](val PT, data []byte) error {
	r := bytes.NewReader(data)
	_, err := readFrom(val, r)
	return err
}

func readFrom[PT hasID[T], T any](val PT, r io.Reader) (int64, error) {
	cr := &countingReader{r: r}
	v, err := varint.ReadUvarint(cr)
	if err != nil {
		return cr.readCount, err
	}
	id := multicodec.Code(v)
	if id != val.ID() {
		return cr.readCount, fmt.Errorf("transport ID does not match %s: %s", val.ID(), id)
	}

	nb := nodePrototypes[val.ID()].NewBuilder()
	err = dagcbor.Decode(nb, cr)
	if err != nil {
		return cr.readCount, err
	}
	nd := nb.Build()
	read := bindnode.Unwrap(nd).(PT)
	*val = *read
	return cr.readCount, nil
}

// copied from go-libipni
var (
	_ io.Reader     = (*countingReader)(nil)
	_ io.ByteReader = (*countingReader)(nil)
)

type countingReader struct {
	readCount int64
	r         io.Reader
}

func (c *countingReader) ReadByte() (byte, error) {
	b := []byte{0}
	_, err := c.Read(b)
	return b[0], err
}

func (c *countingReader) Read(b []byte) (n int, err error) {
	read, err := c.r.Read(b)
	c.readCount += int64(read)
	return read, err
}
