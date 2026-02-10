package grpc

import (
	"context"
	"google.golang.org/grpc"
	"gophkeeper/internal/delivery/grpc/handlers"
	"gophkeeper/internal/dto/log"
	"gophkeeper/internal/dto/service"
	"gophkeeper/internal/middleware"
	pb "gophkeeper/internal/pkg/proto"
	"net"
)

type GrpcServer struct {
	Server *grpc.Server
	Port   string
}

type Server struct {
	GrpcServer *GrpcServer
	LogServer  log.Logger
}

func New(grpcPort string, logServer log.Logger, service service.Service) *Server {

	listNotAuthMetod := handlers.GetListNotAuthMetod()

	serverGrpc := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.AuthInterceptor(service, listNotAuthMetod),
			middleware.LoggingInterceptor(logServer),
		),
	)

	pb.RegisterGophKeeperServer(serverGrpc, handlers.New(service, logServer))

	return &Server{
		GrpcServer: &GrpcServer{
			Server: serverGrpc,
			Port:   grpcPort,
		},
		LogServer: logServer,
	}
}

func (s *Server) Start(ctx context.Context) error {
	var (
		chErr = make(chan error, 1)
		err   error
	)

	go func() {
		listen, err := net.Listen("tcp", s.GrpcServer.Port)
		if err != nil {
			chErr <- err
		}
		if err = s.GrpcServer.Server.Serve(listen); err != nil {
			chErr <- err
		}
	}()
	s.LogServer.Info(ctx, "Starting grpc server", "server adrr", s.GrpcServer.Port)

	err = <-chErr

	return err
}

func (s *Server) Stop(ctx context.Context) error {
	s.LogServer.Info(ctx, "Shutdowning grpc server")
	s.GrpcServer.Server.GracefulStop()
	return nil
}
