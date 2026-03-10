// Package middleware provides HTTP middleware for the API server.
package middleware

import (
	"net/http"
	"time"

	"github.com/arunsuryan/event-booking-system/backend/internal/logger"
)

// responseRecorder wraps http.ResponseWriter to capture the status code
// written by downstream handlers, so the logging middleware can report it.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before delegating to the underlying writer.
func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// Logging returns middleware that logs every HTTP request with method, path,
// status code, and duration in structured JSON.
func Logging(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rr, r)

			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rr.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
