package server

import (
	"errors"
	"fmt"
	"net/http"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/storage/pkg/service/storage"
)

var log = logging.Logger("server")

type config struct {
	id      principal.Signer
	service storage.Service
}

type Option func(*config)

// WithIdentity specifies the server DID.
func WithIdentity(s principal.Signer) Option {
	return func(c *config) {
		c.id = s
	}
}

// WithService configures the storage service the server should use.
func WithService(service storage.Service) Option {
	return func(c *config) {
		c.service = service
	}
}

// ListenAndServe creates a new indexing service HTTP server, and starts it up.
func ListenAndServe(addr string, opts ...Option) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: NewServer(opts...),
	}
	log.Infof("Listening on %s", addr)
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// NewServer creates a new storage node server.
func NewServer(opts ...Option) *http.ServeMux {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}

	if c.id == nil {
		log.Warn("Generating a server identity as one has not been set!")
		id, err := ed25519.Generate()
		if err != nil {
			panic(err)
		}
		c.id = id
	}
	log.Infof("Server ID: %s", c.id.DID())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", getRootHandler(c.id))
	mux.HandleFunc("POST /", postRootHandler(c.id, c.service))
	mux.HandleFunc("GET /claim/{hash}", getClaimHandler(c.id, c.service))
	mux.HandleFunc("GET /blob/{hash}", getBlobHandler(c.id, c.service))
	mux.HandleFunc("PUT /blob/{hash}", putBlobHandler(c.id, c.service))
	return mux
}

// getHandler displays version info when a GET request is sent to "/".
func getRootHandler(id principal.Signer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🔥 storage v0.0.0\n"))
		w.Write([]byte("- https://github.com/storacha/storage\n"))
		w.Write([]byte(fmt.Sprintf("- %s", id.DID())))
	}
}

func postRootHandler(id principal.Signer, service storage.Service) func(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}

func getClaimHandler(id principal.Signer, service storage.Service) func(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}

func getBlobHandler(id principal.Signer, service storage.Service) func(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}

func putBlobHandler(id principal.Signer, service storage.Service) func(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}
