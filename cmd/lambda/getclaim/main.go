package main

import (
	"net/http"

	"github.com/storacha/storage/cmd/lambda"
	"github.com/storacha/storage/pkg/aws"
	"github.com/storacha/storage/pkg/service/claims"
)

func main() {
	lambda.StartHTTPHandler(makeHandler)
}

func makeHandler(cfg aws.Config) (http.Handler, error) {
	service, err := aws.Construct(cfg)
	if err != nil {
		return nil, err
	}

	return claims.NewHandler(service.Claims().Store()), nil
}
