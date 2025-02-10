package main

import (
	"net/http"
	"strings"

	"github.com/storacha/storage/cmd/lambda"
	"github.com/storacha/storage/pkg/aws"
)

func main() {
	lambda.StartHTTPHandler(makeHandler)
}

func makeHandler(cfg aws.Config) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		blobStr := parts[len(parts)-1]
		http.Redirect(w, r, cfg.BlobsPublicURL+"/"+cfg.BlobStorePrefix+blobStr, http.StatusTemporaryRedirect)
	}), nil
}
