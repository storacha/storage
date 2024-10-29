package digest

import (
	"fmt"

	"github.com/multiformats/go-multihash"
	"github.com/multiformats/go-varint"
	"github.com/storacha/storage/pkg/piece/size"
)

const FR32_SHA256_TRUNC254_PADDED_BINARY_TREE_CODE = 0x1011
const name = "fr32-sha2-256-trunc254-padded-binary-tree"

func init() {
	multihash.Codes[FR32_SHA256_TRUNC254_PADDED_BINARY_TREE_CODE] = name
	multihash.Names[name] = FR32_SHA256_TRUNC254_PADDED_BINARY_TREE_CODE
}

var ErrWrongCode = fmt.Errorf("multihash code must be 0x%x", FR32_SHA256_TRUNC254_PADDED_BINARY_TREE_CODE)
var ErrWrongName = fmt.Errorf("multihash name must be %s", name)

type PieceDigest interface {
	Bytes() []byte
	Digest() []byte
	Name() string
	Code() uint64
	Size() int
	Padding() uint64
	Height() uint8
	DataCommitment() []byte
}

type pieceDigest []byte

func (p pieceDigest) Bytes() []byte {
	return p
}

// Code implements PieceDigest.
func (p pieceDigest) Code() uint64 {
	return FR32_SHA256_TRUNC254_PADDED_BINARY_TREE_CODE
}

// DataCommitment implements PieceDigest.
func (p pieceDigest) DataCommitment() []byte {
	dc, _ := DataCommitment(p)
	return dc
}

// Digest implements PieceDigest.
func (p pieceDigest) Digest() []byte {
	d, _ := Digest(p)
	return d
}

// Height implements PieceDigest.
func (p pieceDigest) Height() uint8 {
	h, _ := Height(p)
	return h
}

// Name implements PieceDigest.
func (p pieceDigest) Name() string {
	return name
}

// Padding implements PieceDigest.
func (p pieceDigest) Padding() uint64 {
	pd, _ := Padding(p)
	return pd
}

// Size implements PieceDigest.
func (p pieceDigest) Size() int {
	d, _ := Digest(p)
	return len(d)
}

func NewPieceDigest(mh multihash.Multihash) (PieceDigest, error) {
	// verify we pass all checks
	_, err := Padding(mh)
	if err != nil {
		return nil, err
	}
	return pieceDigest(mh), nil
}

func Digest(mh []byte) ([]byte, error) {
	decodedMh, err := multihash.Decode(mh)
	if err != nil {
		return nil, err
	}
	if decodedMh.Code != FR32_SHA256_TRUNC254_PADDED_BINARY_TREE_CODE {
		return nil, ErrWrongCode
	}
	if decodedMh.Name != name {
		return nil, ErrWrongCode
	}
	return decodedMh.Digest, nil
}

func Padding(pd []byte) (uint64, error) {
	d, err := Digest(pd)
	if err != nil {
		return 0, err
	}
	padding, _, err := varint.FromUvarint(d)
	return padding, err
}

func Height(pd []byte) (uint8, error) {
	d, err := Digest(pd)
	if err != nil {
		return 0, err
	}
	_, read, err := varint.FromUvarint(d)
	if err != nil {
		return 0, fmt.Errorf("reading padding: %w", err)
	}
	return d[read], nil
}

func DataCommitment(pd []byte) ([]byte, error) {
	d, err := Digest(pd)
	if err != nil {
		return nil, err
	}
	_, read, err := varint.FromUvarint(d)
	if err != nil {
		return nil, fmt.Errorf("reading padding: %w", err)
	}
	read++
	return d[read:], nil
}

// ToDigest converts a raw data commitment and unpadded length to a v2 piece multihash
// (i.e. log_2(padded data size in bytes) - 5, because 2^5 is 32 bytes which is the leaf node size) to a CID
// by adding:
// - hash type: multihash.SHA2_256_TRUNC254_PADDED_BINARY_TREE
//
// The helpers UnpaddedSizeToV1TreeHeight and Fr32PaddedSizeToV1TreeHeight may help in computing tree height
func FromCommitmentAndSize(commD []byte, unpaddedDataSize uint64) (PieceDigest, error) {
	if len(commD) != 32 {
		return nil, fmt.Errorf("commitments must be 32 bytes long")
	}

	if unpaddedDataSize < 127 {
		return nil, fmt.Errorf("unpadded data size must be at least 127, but was %d", unpaddedDataSize)
	}

	height, padding, err := size.UnpaddedSizeToV1TreeHeightAndPadding(unpaddedDataSize)
	if err != nil {
		return nil, err
	}

	if padding > varint.MaxValueUvarint63 {
		return nil, fmt.Errorf("padded data size must be less than 2^63-1, but was %d", padding)
	}

	mh := FR32_SHA256_TRUNC254_PADDED_BINARY_TREE_CODE
	paddingSize := varint.UvarintSize(padding)
	digestSize := len(commD) + 1 + paddingSize

	mhBuf := make(
		[]byte,
		varint.UvarintSize(uint64(mh))+varint.UvarintSize(uint64(digestSize))+digestSize,
	)

	pos := varint.PutUvarint(mhBuf, uint64(mh))
	pos += varint.PutUvarint(mhBuf[pos:], uint64(digestSize))
	pos += varint.PutUvarint(mhBuf[pos:], padding)
	mhBuf[pos] = height
	pos++
	copy(mhBuf[pos:], commD)
	return pieceDigest(mhBuf), nil
}
