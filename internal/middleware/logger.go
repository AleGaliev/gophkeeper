package middleware

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
}

func MiddlewareHandlerLogger(logger Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			h.ServeHTTP(w, r)
			logger.Info(r.Context(), "request processed", "request url", r.RequestURI, "request metod", r.Method, "duration", start)
		}
		return http.HandlerFunc(fn)
	}
}

func LoggingInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		start := time.Now()

		resp, err := handler(ctx, req)
		logger.Info(ctx, "request processed", "request method", info.FullMethod, "duration", start)

		return resp, err
	}
}
