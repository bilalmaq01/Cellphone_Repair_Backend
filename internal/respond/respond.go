// Package respond provides small helpers for writing JSON HTTP responses.
package respond

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON writes v as a JSON response with the given status code.
// Order matters: set headers, then status, then body.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("respond: failed to encode JSON response: %v", err)
	}
}

// Error writes a JSON error response of the form {"error": message}.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
