package middleware

import (
	"context"
	"net/http"
	"time"
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
