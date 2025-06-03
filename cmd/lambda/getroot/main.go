package main

import (
	"net/http"

	"github.com/storacha/piri/cmd/lambda"
	"github.com/storacha/piri/pkg/aws"
	"github.com/storacha/piri/pkg/server"
)

func main() {
	lambda.StartHTTPHandler(makeHandler)
}

func makeHandler(cfg aws.Config) (http.Handler, error) {
	return server.NewHandler(cfg.Signer), nil
}
