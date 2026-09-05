package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutesReachable(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/flags"},
		{http.MethodGet, "/flags"},
		{http.MethodGet, "/flags/some-key"},
		{http.MethodPut, "/flags/some-key"},
		{http.MethodDelete, "/flags/some-key"},
		{http.MethodGet, "/flags/some-key/evaluate"},
	}
	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(rt.method, rt.path, nil)
			newTestRouter().ServeHTTP(rec, req)
			// A registered route is handled by our handlers, which always
			// answer with JSON (writeJSON/writeError): either an
			// application/json content-type, or an error object
			// {"error":...}. A truly missing route is answered by the mux as a
			// plain-text 404 with neither. So a JSON error body (or an
			// application/json content-type) marks the route as registered,
			// even when it legitimately answers 404 for an unknown key
			// (AC-05/AC-07).
			if !isRegisteredResponse(rec) {
				t.Fatalf("%s %s answered without a JSON error body or application/json content-type, want it registered", rt.method, rt.path)
			}
		})
	}
}

// isRegisteredResponse reports whether the response came from one of our JSON
// handlers rather than the mux's default plain-text 404.
func isRegisteredResponse(rec *httptest.ResponseRecorder) bool {
	if strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		return true
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err == nil {
		if _, ok := body["error"]; ok {
			return true
		}
	}
	return false
}
