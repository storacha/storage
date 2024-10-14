package storage

import (
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/go-ucanto/server"
)

func NewServer(id principal.Signer) (server.ServerView, error) {
	service := NewService()
	var opts []server.Option
	for ability, method := range service {
		opts = append(opts, server.WithServiceMethod(ability, method))
	}
	return server.NewServer(id, opts...)
}
