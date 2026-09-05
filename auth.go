package main

import "net/http"

// withAdminAuth wraps next and protects mutating methods (POST, PUT, DELETE)
// behind an API key. When adminKey is empty the service degrades safely: those
// methods answer 503 rather than staying open. When it is set, the caller must
// present X-API-Key: <adminKey> or receive 401. Read-only and safe methods
// (GET, OPTIONS, HEAD) pass through untouched so /healthz and evaluation stay
// open.
func withAdminAuth(next http.Handler, adminKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete:
			if adminKey == "" {
				writeError(w, http.StatusServiceUnavailable, "authentication not configured")
				return
			}
			if r.Header.Get("X-API-Key") != adminKey {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
