package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestRouter() http.Handler {
	return NewRouter(NewServer(NewStore()))
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want ok", body["status"])
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/healthz", nil)
	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow == "" {
		t.Fatal("Allow header missing on 405")
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatal("error field missing from 405 body")
	}
}

func TestRegisteredRoutesReachable(t *testing.T) {
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
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(rt.method, rt.path, nil)
		newTestRouter().ServeHTTP(rec, req)
		// A registered route is handled by our handlers, which always answer
		// with JSON (writeJSON/writeError): either an application/json
		// content-type, or an error object {"error":...}. A truly missing route
		// is answered by the mux as a plain-text 404 with neither. So a JSON
		// error body (or an application/json content-type) marks the route as
		// registered, even when it legitimately answers 404 for an unknown key.
		if !isRegisteredResponse(rec) {
			t.Fatalf("%s %s answered without a JSON error body or application/json content-type, want it registered", rt.method, rt.path)
		}
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
