package replicator

import (
	"context"
	"net/url"

	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/did"
)

type Task struct {
	// bucket to associate with blob
	Space did.DID
	// the blob in question
	Blob blob.Blob
	// the location to replicate the blob from
	Source url.URL
	// the location to replicate the blob to
	Sink url.URL
	// invocation responsible for spawning this replication
	// should be a replica/transfer invocation
	Invocation invocation.Invocation
}

type Replicator interface {
	Replicate(context.Context, *Task) error
}
