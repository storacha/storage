package jobqueue

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-ucanto/core/ipld/codec/cbor"
	"github.com/storacha/go-ucanto/core/ipld/codec/json"
)

type Serializer[T any] interface {
	Serialize(val T) ([]byte, error)
	Deserialize(data []byte) (T, error)
}

type IPLDSerializerCBOR[T any] struct {
	typ  schema.Type
	opts []bindnode.Option
}

func (i *IPLDSerializerCBOR[T]) Serialize(val T) ([]byte, error) {
	return cbor.Encode(val, i.typ, i.opts...)
}

func (i *IPLDSerializerCBOR[T]) Deserialize(data []byte) (T, error) {
	var out T
	if err := cbor.Decode(data, out, i.typ, i.opts...); err != nil {
		return out, err
	}
	return out, nil
}

type IPLDSerializerJSON[T any] struct {
	Typ  schema.Type
	Opts []bindnode.Option
}

func (i *IPLDSerializerJSON[T]) Serialize(val T) ([]byte, error) {
	return json.Encode(val, i.Typ, i.Opts...)
}

func (i *IPLDSerializerJSON[T]) Deserialize(data []byte) (T, error) {
	var out T
	if err := json.Decode(data, out, i.Typ, i.Opts...); err != nil {
		return out, err
	}
	return out, nil
}

type PassThroughSerializer[T any] struct{}

func (p PassThroughSerializer[T]) Serialize(val T) ([]byte, error) {
	b, ok := any(val).([]byte)
	if !ok {
		return nil, fmt.Errorf("PassThroughSerializer only supports []byte, got %T", val)
	}
	return b, nil
}

func (p PassThroughSerializer[T]) Deserialize(data []byte) (T, error) {
	var zero T
	// We cast []byte back to T, but T must be []byte or we return an error:
	if _, ok := any(zero).([]byte); !ok {
		return zero, fmt.Errorf("PassThroughSerializer only supports T = []byte")
	}
	return any(data).(T), nil
}
