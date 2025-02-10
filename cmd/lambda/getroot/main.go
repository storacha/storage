package main

import (
	"net/http"

	"github.com/storacha/storage/cmd/lambda"
	"github.com/storacha/storage/pkg/aws"
	"github.com/storacha/storage/pkg/server"
)

func main() {
	lambda.StartHTTPHandler(makeHandler)
}

func makeHandler(cfg aws.Config) (http.Handler, error) {
	return server.NewHandler(cfg.Signer), nil
}
