package api

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	e *echo.Echo
}

func NewServer(p *PDP) *Server {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	RegisterEchoRoutes(e, p)

	return &Server{e: e}
}

func (s *Server) Start() error {
	go s.e.Start(":8080")
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
