package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"notif/config"

	"github.com/labstack/echo/v4"
)

type server struct {
	e *echo.Echo
}

func NewServer() *server {
	e := echo.New()
	return &server{
		e: e,
	}
}

func (s *server) Serve(ctx context.Context) {
	cnf := config.C

	go func() {
		addr := fmt.Sprintf(":%d", cnf.Server.Port)
		if err := s.e.Start(addr); err != nil && err != http.ErrServerClosed {
			s.e.Logger.Fatal("shutting down the server")
		}
	}()

	go func() {
		<-ctx.Done()
		s.e.Logger.Info("Shutdown signal received")

		// Create timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.e.Shutdown(ctx); err != nil {
			s.e.Logger.Fatal(err)
		}
	}()

	s.e.POST("/send", SendSMS)
	s.e.GET("/orgs/:org/count", GetOrgCount)
}
