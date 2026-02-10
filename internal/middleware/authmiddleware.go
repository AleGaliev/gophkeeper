package middleware

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JWTManager interface {
	GetLoginFromToken(tokenString string) (string, error)
}

func AuthMiddleware(jwtManager JWTManager) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("Authorization")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := jwtManager.GetLoginFromToken(c.Value)

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-User-Login", user)

			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func AuthInterceptor(jwtManager JWTManager, skipAuthMethod []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		for _, metod := range skipAuthMethod {
			if metod == info.FullMethod {
				return handler(ctx, req)
			}
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied,
				"metadata not found in context")
		}
		tokens := md.Get("Authorization")
		if len(tokens) == 0 {
			return nil, status.Errorf(codes.PermissionDenied,
				"authorization token is not provided")
		}
		for _, t := range tokens {
			user, err := jwtManager.GetLoginFromToken(t)
			if err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "not authorized")
			}
			ctx = context.WithValue(ctx, "user", user)
		}

		resp, err := handler(ctx, req)

		return resp, err
	}
}
