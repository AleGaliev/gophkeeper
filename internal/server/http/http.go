package http

import (
	"context"
	"gophkeeper/internal/delivery/http/handlers"
	"gophkeeper/internal/dto/log"
	"gophkeeper/internal/dto/service"
	"net/http"
	"time"
)

type Server struct {
	Server    *http.Server
	LogServer log.Logger
}

func New(httpPort string, logServer log.Logger, service service.Service) *Server {

	server := &http.Server{
		Addr:    httpPort,
		Handler: handlers.New(service, logServer),
	}

	return &Server{
		Server:    server,
		LogServer: logServer,
	}
}

func (s *Server) Start(ctx context.Context) error {
	var (
		chErr = make(chan error, 1)
		err   error
	)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil {
			chErr <- err
		}
	}()
	s.LogServer.Info(ctx, "Starting http server", "server adrr", s.Server.Addr)

	err = <-chErr

	return err
}

func (s *Server) Stop(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.LogServer.Info(ctx, "Shutdowning http server")

	err := s.Server.Shutdown(shutdownCtx)

	return err
}
