package server

import (
	"errors"
	"fmt"
	"net/http"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/build"
	"github.com/storacha/storage/pkg/service/blobs"
	"github.com/storacha/storage/pkg/service/claims"
	"github.com/storacha/storage/pkg/service/publisher"
	"github.com/storacha/storage/pkg/service/storage"
)

var log = logging.Logger("server")

// ListenAndServe creates a new indexing service HTTP server, and starts it up.
func ListenAndServe(addr string, service storage.Service) error {
	srvMux, err := NewServer(service)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:    addr,
		Handler: srvMux,
	}
	log.Infof("Listening on %s", addr)
	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// NewServer creates a new storage node server.
func NewServer(service storage.Service) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", NewHandler(service.ID()))

	httpUcanSrv, err := storage.NewServer(service)
	if err != nil {
		return nil, fmt.Errorf("creating UCAN server: %w", err)
	}
	httpUcanSrv.Serve(mux)

	httpClaimsSrv, err := claims.NewServer(service.Claims().Store())
	if err != nil {
		return nil, fmt.Errorf("creating claims server: %w", err)
	}
	httpClaimsSrv.Serve(mux)

	httpBlobsSrv, err := blobs.NewServer(service.Blobs().Presigner(), service.Blobs().Allocations(), service.Blobs().Store())
	if err != nil {
		return nil, fmt.Errorf("creating blobs server: %w", err)
	}
	httpBlobsSrv.Serve(mux)

	publisherStore := service.Claims().Publisher().Store()
	encodableStore, ok := publisherStore.(store.EncodeableStore)
	if !ok {
		return nil, errors.New("publisher store does not implement EncodableStore")
	}

	httpPublisherSrv, err := publisher.NewServer(encodableStore)
	if err != nil {
		return nil, fmt.Errorf("creating IPNI publisher server: %w", err)
	}
	httpPublisherSrv.Serve(mux)

	return mux, nil
}

// NewHandler displays version info.
func NewHandler(id principal.Signer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("🔥 storage %s\n", build.Version)))
		w.Write([]byte("- https://github.com/storacha/storage\n"))
		w.Write([]byte(fmt.Sprintf("- %s", id.DID())))
	}
}
