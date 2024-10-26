package blobs

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/storage/pkg/presigner"
	"github.com/storacha/storage/pkg/store/allocationstore"
	"github.com/storacha/storage/pkg/store/blobstore"
)

var log = logging.Logger("blobs")

type Server struct {
	blobs     blobstore.Blobstore
	presigner presigner.RequestPresigner
	allocs    allocationstore.AllocationStore
}

func NewServer(presigner presigner.RequestPresigner, allocs allocationstore.AllocationStore, blobs blobstore.Blobstore) (*Server, error) {
	return &Server{blobs, presigner, allocs}, nil
}

func (srv *Server) Serve(mux *http.ServeMux) {
	mux.HandleFunc("GET /blob/{blob}", NewBlobGetHandler(srv.blobs))
	mux.HandleFunc("PUT /blob/{blob}", NewBlobPutHandler(srv.presigner, srv.allocs, srv.blobs))
}

func NewBlobGetHandler(blobs blobstore.Blobstore) func(http.ResponseWriter, *http.Request) {
	if fsblobs, ok := blobs.(blobstore.FileSystemer); ok {
		serveHTTP := http.FileServer(fsblobs.FileSystem()).ServeHTTP
		return func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = r.URL.Path[len("/blob"):]
			serveHTTP(w, r)
		}
	}

	log.Error("blobstore does not support filesystem access")
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not supported", http.StatusInternalServerError)
	}
}

func NewBlobPutHandler(presigner presigner.RequestPresigner, allocs allocationstore.AllocationStore, blobs blobstore.Blobstore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, sHeaders, err := presigner.VerifyUploadURL(r.Context(), *r.URL, r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		_, bytes, err := multibase.Decode(r.PathValue("blob"))
		if err != nil {
			http.Error(w, fmt.Sprintf("decoding multibase encoded digest: %s", err), http.StatusBadRequest)
			return
		}

		digest, err := multihash.Cast(bytes)
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

		log.Infof("Found %d allocations for write to: z%s", len(results), digest.B58String())

		// ensure the size comes from a signed header
		contentLength, err := strconv.ParseInt(sHeaders.Get("Content-Length"), 10, 64)
		if err != nil {
			log.Warnf("parsing signed Content-Length header: %w", err)
			http.Error(w, "invalid size", http.StatusInternalServerError)
			return
		}

		err = blobs.Put(r.Context(), digest, uint64(contentLength), r.Body)
		if err != nil {
			log.Errorf("writing to: z%s: %w", digest.B58String(), err)
			if errors.Is(err, blobstore.ErrDataInconsistent) {
				http.Error(w, "data consistency check failed", http.StatusConflict)
				return
			}
			http.Error(w, "write failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
