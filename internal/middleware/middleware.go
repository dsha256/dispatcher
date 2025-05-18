package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddleware logs the request details.
func LoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Info("Request started", "method", r.Method, "url", r.URL.String())
		next.ServeHTTP(w, r)
		logger.Info("Request completed", "method", r.Method, "url", r.URL.String(), "duration", time.Since(start).String())
	})
}

// RecoveryMiddleware recovers from panics and logs them.
func RecoveryMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Recovery from panic", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
