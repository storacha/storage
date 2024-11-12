package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/storacha/storage/pkg/aws"
	"github.com/storacha/storage/pkg/service/claims"
)

func main() {
	config := aws.FromEnv(context.Background())
	service, err := aws.Construct(config)
	if err != nil {
		panic(err)
	}
	handler := claims.NewHandler(service.Claims().Store())
	lambda.Start(httpadapter.NewV2(http.HandlerFunc(handler)).ProxyWithContext)
}
