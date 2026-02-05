package http

import (
	"context"
	"gophkeeper/internal/delivery/handlers"
	"gophkeeper/internal/infra/postgres"
	"gophkeeper/internal/log"
	"gophkeeper/internal/server"
	"gophkeeper/internal/service"
	"gophkeeper/internal/service/jwtmanager"
	"net/http"
	"time"
)

type Server struct {
	Server    *http.Server
	LogServer log.Logger
}

func New(cfg server.Config) (*Server, error) {
	postgresStorage, err := postgres.New(cfg.DatabaseDSN)
	if err != nil {
		return &Server{}, err
	}
	JWTManager := jwtmanager.New(cfg.Key)

	service := service.New(&postgresStorage, JWTManager)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handlers.New(service, cfg.Logger),
	}

	return &Server{
		Server:    server,
		LogServer: cfg.Logger,
	}, err
}

func (s *Server) Start(ctx context.Context) error {
	var chErr = make(chan error, 1)

	s.LogServer.Info(ctx, "Starting server", "server adrr", s.Server.Addr)
	if err := s.Server.ListenAndServe(); err != nil {
		return err
	}
	go func() {
		if err := s.Server.ListenAndServe(); err != nil {
			chErr <- err
		}
	}()

	select {
	case err := <-chErr:
		return err
	case <-ctx.Done():
		s.LogServer.Info(ctx, "Shutting down server")
		if err := s.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Server.Shutdown(shutdownCtx)

	return err
}
