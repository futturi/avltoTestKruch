package server

import (
	"context"
	"net/http"
)

type Server struct {
	Server *http.Server
}

func (s *Server) InitServer(port string, handler http.Handler) error {

	s.Server = &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	return s.Server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
