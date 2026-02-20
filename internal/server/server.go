package server

import (
	"context"
	cfg "gophkeeper/internal/config/server"
	"gophkeeper/internal/infra/postgres"
	"gophkeeper/internal/server/grpc"
	"gophkeeper/internal/server/http"
	"gophkeeper/internal/service/jwtmanager"
	"gophkeeper/internal/service/server"
)

type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
}

func New(cfg cfg.Config) (*Server, error) {
	storage, err := postgres.New(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	JWTManager := jwtmanager.New(cfg.Key)
	service := server.New(&storage, JWTManager)
	grpcServer := grpc.New(cfg.GrpcPort, cfg.Logger, service)
	httpServer := http.New(cfg.HttpPort, cfg.Logger, service)

	return &Server{grpcServer, httpServer}, nil
}

func (s *Server) Start(ctx context.Context) error {
	var (
		chErr = make(chan error, 1)
		err   error
	)

	go func() {
		if err := s.grpcServer.Start(ctx); err != nil {
			chErr <- err
		}
	}()

	go func() {
		if err := s.httpServer.Start(ctx); err != nil {
			chErr <- err
		}
	}()

	select {
	case err := <-chErr:
		return err
	case <-ctx.Done():
		s.Stop(ctx)
	}
	err = <-chErr

	return err
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.grpcServer.Stop(ctx); err != nil {
		return err
	}
	if err := s.httpServer.Stop(ctx); err != nil {
		return err
	}
	return nil
}
