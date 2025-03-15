package jobqueue

import (
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-ucanto/core/ipld/codec/cbor"
)

type Serializer[T any] interface {
	Serialize(val T) ([]byte, error)
	Deserialize(data []byte) (T, error)
}

type IPLDSerializerCBOR[T any] struct {
	Typ  schema.Type
	Opts []bindnode.Option
}

func (i *IPLDSerializerCBOR[T]) Serialize(val T) ([]byte, error) {
	return cbor.Encode(&val, i.Typ, i.Opts...)
}

func (i *IPLDSerializerCBOR[T]) Deserialize(data []byte) (T, error) {
	var out T
	if err := cbor.Decode(data, &out, i.Typ, i.Opts...); err != nil {
		return out, err
	}
	return out, nil
}
