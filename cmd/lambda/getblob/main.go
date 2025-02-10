package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/storacha/storage/pkg/aws"
)

func makeHandler(blobsPublicURL string, blobsKeyPrefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		blobStr := parts[len(parts)-1]
		http.Redirect(w, r, blobsPublicURL+"/"+blobsKeyPrefix+blobStr, http.StatusTemporaryRedirect)
	}
}
func main() {
	config := aws.FromEnv(context.Background())
	lambda.Start(httpadapter.NewV2(http.HandlerFunc(makeHandler(config.BlobsPublicURL, config.BlobStorePrefix))).ProxyWithContext)
}
