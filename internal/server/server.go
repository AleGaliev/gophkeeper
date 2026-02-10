package server

import (
	"context"
	"gophkeeper/internal/infra/postgres"
	"gophkeeper/internal/log"
	"gophkeeper/internal/server/grpc"
	"gophkeeper/internal/server/http"
	"gophkeeper/internal/service"
	"gophkeeper/internal/service/jwtmanager"
)

type Server struct {
	httpServer *http.Server
	grpcServer *grpc.Server
}

func New(cfg Config) (*Server, error) {
	postgresStorage, err := postgres.New(cfg.DatabaseDSN)
	if err != nil {
		return &Server{}, err
	}
	JWTManager := jwtmanager.New(cfg.Key)

	service := service.New(&postgresStorage, JWTManager)
	logger := log.New(cfg.LogLevel)
	return &Server{
		httpServer: http.New(cfg.HttpPort, logger, service),
		grpcServer: grpc.New(cfg.GrpcPort, logger, service),
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	var chErr = make(chan error, 1)

	go func() {
		if err := s.httpServer.Start(ctx); err != nil {
			chErr <- err
		}
	}()

	go func() {
		if err := s.grpcServer.Start(ctx); err != nil {
			chErr <- err
		}
	}()

	select {
	case err := <-chErr:
		return err
	case <-ctx.Done():
		if err := s.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.httpServer.Stop(ctx); err != nil {
		return err
	}
	if err := s.grpcServer.Stop(ctx); err != nil {
		return err
	}
	return nil
}
