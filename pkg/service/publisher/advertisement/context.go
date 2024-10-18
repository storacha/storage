package advertisement

import (
	"bytes"

	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/did"
)

// Encode canonically encodes ContextID data.
func EncodeContextID(space did.DID, digest mh.Multihash) ([]byte, error) {
	return mh.Sum(bytes.Join([][]byte{space.Bytes(), digest}, nil), mh.SHA2_256, -1)
}
