package digestutil

import (
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
)

func Format(digest multihash.Multihash) string {
	key, _ := multibase.Encode(multibase.Base58BTC, digest)
	return key
}
