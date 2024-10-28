package publisher

import (
	"fmt"
	"net/http"

	"github.com/storacha/ipni-publisher/pkg/server"
	"github.com/storacha/ipni-publisher/pkg/store"
)

type Server struct {
	server *server.Server
}

func NewServer(store store.EncodeableStore) (*Server, error) {
	server, err := server.NewServer(store)
	if err != nil {
		return nil, err
	}
	return &Server{server}, nil
}

func (srv *Server) Serve(mux *http.ServeMux) {
	mux.HandleFunc(fmt.Sprintf("GET %s/{ad}", server.IPNIPath), srv.server.ServeHTTP)
}
