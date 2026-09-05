package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func doRequest(t *testing.T, router http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	router.ServeHTTP(rec, req)
	return rec
}

func seedStore(t *testing.T) *Store {
	t.Helper()
	s := NewStore()
	if err := s.Create(Flag{Key: "alpha", Enabled: true, Description: "first", RolloutPercent: 100}); err != nil {
		t.Fatalf("create alpha: %v", err)
	}
	if err := s.Create(Flag{Key: "beta", Enabled: false, Description: "second", RolloutPercent: 50}); err != nil {
		t.Fatalf("create beta: %v", err)
	}
	return s
}

func TestHandleListEmpty(t *testing.T) {
	rec := doRequest(t, NewRouter(NewServer(NewStore())), http.MethodGet, "/flags")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	var flags []Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flags); err != nil {
		t.Fatalf("body not a JSON array: %v", err)
	}
	if flags == nil {
		t.Fatal("body is null, want an empty array []")
	}
	if len(flags) != 0 {
		t.Fatalf("got %d flags, want 0", len(flags))
	}
}

func TestHandleListPopulated(t *testing.T) {
	router := NewRouter(NewServer(seedStore(t)))
	rec := doRequest(t, router, http.MethodGet, "/flags")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var flags []Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flags); err != nil {
		t.Fatalf("body not a JSON array: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("got %d flags, want 2", len(flags))
	}
	keys := map[string]bool{}
	for _, f := range flags {
		keys[f.Key] = true
	}
	if !keys["alpha"] || !keys["beta"] {
		t.Fatalf("flags %v missing alpha/beta", flags)
	}
}

func TestHandleGetFound(t *testing.T) {
	router := NewRouter(NewServer(seedStore(t)))
	rec := doRequest(t, router, http.MethodGet, "/flags/alpha")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	var flag Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flag); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if flag.Key != "alpha" {
		t.Fatalf("key = %q, want alpha", flag.Key)
	}
	if !flag.Enabled {
		t.Fatal("enabled = false, want true")
	}
	if flag.Description != "first" {
		t.Fatalf("description = %q, want first", flag.Description)
	}
	if flag.RolloutPercent != 100 {
		t.Fatalf("rollout_percent = %d, want 100", flag.RolloutPercent)
	}
}

func TestHandleGetNotFound(t *testing.T) {
	router := NewRouter(NewServer(seedStore(t)))
	rec := doRequest(t, router, http.MethodGet, "/flags/missing")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatal("error field missing from 404 body")
	}
}

func TestHandleDeleteRemovesFlag(t *testing.T) {
	store := seedStore(t)
	router := NewRouter(NewServer(store))

	rec := doRequest(t, router, http.MethodDelete, "/flags/alpha")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", rec.Body.String())
	}

	rec = doRequest(t, router, http.MethodGet, "/flags/alpha")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want 404", rec.Code)
	}
}

func TestHandleDeleteUnknownKey(t *testing.T) {
	router := NewRouter(NewServer(seedStore(t)))
	rec := doRequest(t, router, http.MethodDelete, "/flags/missing")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatal("error field missing from 404 body")
	}
}
