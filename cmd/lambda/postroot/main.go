package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	ucanserver "github.com/storacha/go-ucanto/server"
	"github.com/storacha/storage/pkg/aws"
	"github.com/storacha/storage/pkg/presets"
	"github.com/storacha/storage/pkg/principalresolver"
	"github.com/storacha/storage/pkg/service/storage"
)

func main() {
	config := aws.FromEnv(context.Background())
	service, err := aws.Construct(config)
	if err != nil {
		panic(err)
	}
	presolv, err := principalresolver.New(presets.PrincipalMapping)
	if err != nil {
		panic(err)
	}
	server, err := storage.NewUCANServer(service, ucanserver.WithPrincipalResolver(presolv.ResolveDIDKey))
	if err != nil {
		panic(err)
	}
	handler := storage.NewHandler(server)
	lambda.Start(httpadapter.NewV2(http.HandlerFunc(handler)).ProxyWithContext)
}
