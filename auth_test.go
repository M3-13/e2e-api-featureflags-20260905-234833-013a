package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// passthrough is a stub next handler that confirms the request reached the
// wrapped handler.
func passthrough(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithAdminAuthEmptyKeyPostServiceUnavailable(t *testing.T) {
	handler := withAdminAuth(http.HandlerFunc(passthrough), "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("POST with empty admin key: got %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestWithAdminAuthEmptyKeyGetHealthzOpen(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	handler := withAdminAuth(router, "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz with empty admin key: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestWithAdminAuthSetKeyPostMissingKeyUnauthorized(t *testing.T) {
	handler := withAdminAuth(http.HandlerFunc(passthrough), "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST without key: got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithAdminAuthSetKeyPostWrongKeyUnauthorized(t *testing.T) {
	handler := withAdminAuth(http.HandlerFunc(passthrough), "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	req.Header.Set("X-API-Key", "wrong")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST with wrong key: got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithAdminAuthSetKeyPostCorrectKeyPassesThrough(t *testing.T) {
	handler := withAdminAuth(http.HandlerFunc(passthrough), "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	req.Header.Set("X-API-Key", "secret")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST with correct key: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestWithAdminAuthSetKeyPutAndDeleteRequireKey(t *testing.T) {
	handler := withAdminAuth(http.HandlerFunc(passthrough), "secret")

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/flags/x", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s without key: got %d, want %d", method, rec.Code, http.StatusUnauthorized)
		}

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(method, "/flags/x", nil)
		req.Header.Set("X-API-Key", "secret")
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s with correct key: got %d, want %d", method, rec.Code, http.StatusOK)
		}
	}
}

func TestWithAdminAuthSetKeyGetRemainsOpenWithoutKey(t *testing.T) {
	handler := withAdminAuth(http.HandlerFunc(passthrough), "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET without key: got %d, want %d", rec.Code, http.StatusOK)
	}
}
