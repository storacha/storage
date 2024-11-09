package storage

import (
	"fmt"
	"io"
	"net/http"

	"github.com/storacha/go-ucanto/server"
	ucanhttp "github.com/storacha/go-ucanto/transport/http"
)

type Server struct {
	ucanServer server.ServerView
}

func NewServer(service Service, options ...server.Option) (*Server, error) {
	ucanSrv, err := NewUCANServer(service, options...)
	if err != nil {
		return nil, fmt.Errorf("creating UCAN server: %w", err)
	}

	return &Server{ucanSrv}, nil
}

func (srv *Server) Serve(mux *http.ServeMux) {
	mux.HandleFunc("POST /", NewHandler(srv.ucanServer))
}

func NewHandler(server server.ServerView) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := server.Request(ucanhttp.NewHTTPRequest(r.Body, r.Header))
		if err != nil {
			log.Errorf("handling UCAN request: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for key, vals := range res.Headers() {
			for _, v := range vals {
				w.Header().Add(key, v)
			}
		}

		if res.Status() != 0 {
			w.WriteHeader(res.Status())
		}

		_, err = io.Copy(w, res.Body())
		if err != nil {
			log.Errorf("sending UCAN response: %w", err)
		}
	}
}
