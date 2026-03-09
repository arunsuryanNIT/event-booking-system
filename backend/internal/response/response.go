// Package response provides standardised JSON response helpers for HTTP handlers.
// Every API endpoint returns the same envelope structure so the frontend can
// parse responses uniformly.
//
//	Success: {"success": true,  "data": ..., "message": "..."}
//	Error:   {"success": false, "error": "...", "message": "..."}
package response

import (
	"encoding/json"
	"net/http"
)

// APIResponse is the standard envelope for all JSON responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message"`
}

// Success writes a JSON success response with the given HTTP status code, payload, and message.
func Success(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// Error writes a JSON error response with the given HTTP status code, error detail, and message.
func Error(w http.ResponseWriter, statusCode int, err string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   err,
		Message: message,
	})
}
