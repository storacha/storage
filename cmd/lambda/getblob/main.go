package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/storacha/storage/pkg/aws"
)

func makeHandler(blobsPublicURL string, keyPattern string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		blobStr := parts[len(parts)-1]
		if keyPattern == "" {
			keyPattern = "blob/{blob}"
		}
		key := strings.ReplaceAll(keyPattern, "{blob}", blobStr)
		http.Redirect(w, r, blobsPublicURL+"/"+key, http.StatusTemporaryRedirect)
	}
}
func main() {
	config := aws.FromEnv(context.Background())
	lambda.Start(httpadapter.NewV2(http.HandlerFunc(makeHandler(config.BlobsPublicURL, config.BlobStoreBucketKeyPattern))).ProxyWithContext)
}
