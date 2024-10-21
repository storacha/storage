package claims

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/storage/pkg/store"
	"github.com/storacha/storage/pkg/store/claimstore"
)

var log = logging.Logger("blobs")

type Server struct {
	claims claimstore.ClaimStore
}

func NewServer(claims claimstore.ClaimStore) (*Server, error) {
	return &Server{claims}, nil
}

func (srv *Server) Serve(mux *http.ServeMux) {
	mux.HandleFunc("GET /claim/{claim}", NewHandler(srv.claims))
}

func NewHandler(claims claimstore.ClaimStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := cid.Parse(r.PathValue("claim"))
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
