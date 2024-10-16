package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	ucanhttp "github.com/storacha/go-ucanto/transport/http"
	"github.com/storacha/storage/pkg/presigner"
	"github.com/storacha/storage/pkg/service/storage"
	"github.com/storacha/storage/pkg/store"
	"github.com/storacha/storage/pkg/store/allocationstore"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/claimstore"
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
	mux.HandleFunc("GET /claim/{cid}", getClaimHandler(c.service.Claims()))
	mux.HandleFunc("GET /blob/{digest}", getBlobHandler(c.service.Blobs()))
	mux.HandleFunc("PUT /blob/{digest}", putBlobHandler(c.service.Presigner(), c.service.Allocations(), c.service.Blobs()))
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
	server, err := storage.NewServer(id, service)
	if err != nil {
		log.Fatalf("creating ucanto server: %s", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		res, _ := server.Request(ucanhttp.NewHTTPRequest(r.Body, r.Header))

		for key, vals := range res.Headers() {
			for _, v := range vals {
				w.Header().Add(key, v)
			}
		}

		if res.Status() != 0 {
			w.WriteHeader(res.Status())
		}

		_, err := io.Copy(w, res.Body())
		if err != nil {
			log.Errorf("sending UCAN response: %w", err)
		}
	}
}

func getClaimHandler(claims claimstore.ClaimStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := cid.Parse(r.PathValue("cid"))
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid CID: %s", err), http.StatusBadRequest)
			return
		}

		dlg, err := claims.Get(r.Context(), cidlink.Link{Cid: c})
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				http.Error(w, fmt.Sprintf("not found: %s", c), http.StatusNotFound)
				return
			}
			log.Errorf("getting claim: %w", err)
			http.Error(w, "failed to get claim", http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(w, dlg.Archive())
		if err != nil {
			log.Warnf("serving claim: %s: %w", c, err)
		}
	}
}

func getBlobHandler(blobs blobstore.Blobstore) func(http.ResponseWriter, *http.Request) {
	if fsblobs, ok := blobs.(blobstore.FileSystemer); ok {
		serveHTTP := http.FileServer(fsblobs.FileSystem()).ServeHTTP
		return func(w http.ResponseWriter, r *http.Request) {
			// trim base58btc multibase prefix if it was added
			digest, err := mh.FromB58String(strings.TrimLeft(r.PathValue("digest"), "z"))
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid multihash digest: %s", err), http.StatusBadRequest)
				return
			}
			r.URL.Path = fsblobs.EncodePath(digest)
			serveHTTP(w, r)
		}
	}

	log.Error("blobstore does not support filesystem access")
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not supported", http.StatusInternalServerError)
	}
}

func putBlobHandler(presigner presigner.RequestPresigner, allocs allocationstore.AllocationStore, blobs blobstore.Blobstore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, sHeaders, err := presigner.VerifyUploadURL(r.Context(), *r.URL, r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// trim base58btc multibase prefix if it was added
		digest, err := mh.FromB58String(strings.TrimLeft(r.PathValue("digest"), "z"))
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid multihash digest: %s", err), http.StatusBadRequest)
			return
		}

		results, err := allocs.List(r.Context(), digest)
		if err != nil {
			log.Errorf("listing allocations: %w", err)
			http.Error(w, "list allocations failed", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			log.Warnf("missing allocation for write to: z%s", digest.B58String())
			http.Error(w, "missing allocation", http.StatusForbidden)
			return
		}

		expired := true
		for _, a := range results {
			exp := a.Expires
			if exp > uint64(time.Now().Unix()) {
				expired = false
				break
			}
		}
		if expired {
			http.Error(w, "expired allocation", http.StatusForbidden)
			return
		}

		log.Infof("found %d allocations for write to: z%s", len(results), digest.B58String())

		_, err = blobs.Get(r.Context(), digest)
		if err == nil {
			log.Warnf("repeated write to: z%s", digest.B58String())
			http.Error(w, "object exists", http.StatusConflict)
			return
		}
		if !errors.Is(err, store.ErrNotFound) {
			log.Errorf("getting exising blob: %w", err)
			http.Error(w, "read failed", http.StatusInternalServerError)
			return
		}

		// ensure the size comes from a signed header
		contentLength, err := strconv.ParseInt(sHeaders.Get("Content-Length"), 10, 64)
		if err != nil {
			log.Warnf("parsing signed Content-Length header: %w", err)
			http.Error(w, "invalid size", http.StatusInternalServerError)
			return
		}

		err = blobs.Put(r.Context(), digest, uint64(contentLength), r.Body)
		if err == nil {
			log.Errorf("writing to: z%s: %w", digest.B58String(), err)
			if errors.Is(err, blobstore.ErrDataInconsistent) {
				http.Error(w, "data consistency check failed", http.StatusBadRequest)
				return
			}
			http.Error(w, "write failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
