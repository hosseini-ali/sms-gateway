package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"notif/config"
	"notif/internal/app"

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

func (s *server) Serve() {
	cnf := config.C

	go func() {
		addr := fmt.Sprintf(":%d", cnf.Server.Port)
		if err := s.e.Start(addr); err != nil && err != http.ErrServerClosed {
			s.e.Logger.Fatal("shutting down the server")
		}
	}()

	go func() {
		<-app.A.Ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.e.Shutdown(ctx); err != nil {
			s.e.Logger.Fatal(err)
		}
	}()

	s.e.POST("/send", SendSMS)
}
