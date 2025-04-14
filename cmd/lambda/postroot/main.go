package main

import (
	"net/http"

	ucanserver "github.com/storacha/go-ucanto/server"

	"github.com/storacha/storage/cmd/lambda"
	"github.com/storacha/storage/pkg/aws"
	"github.com/storacha/storage/pkg/principalresolver"
	"github.com/storacha/storage/pkg/service/storage"
)

func main() {
	lambda.StartHTTPHandler(makeHandler)
}

func makeHandler(cfg aws.Config) (http.Handler, error) {
	service, err := aws.Construct(cfg)
	if err != nil {
		return nil, err
	}

	presolv, err := principalresolver.New(cfg.PrincipalMapping)
	if err != nil {
		return nil, err
	}

	server, err := storage.NewUCANServer(service, ucanserver.WithPrincipalResolver(presolv.ResolveDIDKey))
	if err != nil {
		return nil, err
	}

	return storage.NewHandler(server), nil
}
