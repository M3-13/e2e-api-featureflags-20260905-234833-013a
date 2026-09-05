package main

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes v as a JSON response with the given status, setting the
// Content-Type and the nosniff security header.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error object {"error": msg} with the given status.
// A 500 status is always masked to "internal server error" so no internal
// detail leaks to the client.
func writeError(w http.ResponseWriter, status int, msg string) {
	if status == http.StatusInternalServerError {
		msg = "internal server error"
	}
	writeJSON(w, status, map[string]string{"error": msg})
}
