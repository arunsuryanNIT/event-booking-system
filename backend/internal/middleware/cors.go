package middleware

import (
	"net/http"
)

// CORS returns middleware that sets permissive Cross-Origin Resource Sharing headers.
// In production behind Nginx, the frontend and API share the same origin so CORS
// isn't triggered. This is needed for local development where the Vite dev server
// (port 3000) makes requests to the Go backend (port 8080).
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
