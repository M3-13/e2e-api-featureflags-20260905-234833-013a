package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func seedFlag(t *testing.T, s *Store, f Flag) {
	t.Helper()
	if err := s.Create(f); err != nil {
		t.Fatalf("seed flag %q: %v", f.Key, err)
	}
}

func TestEvaluateStableAcrossRepeats(t *testing.T) {
	s := NewStore()
	seedFlag(t, s, Flag{Key: "feature", Enabled: true, RolloutPercent: 50})
	router := NewRouter(NewServer(s))

	var first bool
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user=alice", nil)
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("iteration %d: status %d", i, rec.Code)
		}
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if err := decodeBody(rec, &body); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if i == 0 {
			first = body.Enabled
		} else if body.Enabled != first {
			t.Fatalf("iteration %d: enabled=%v, want stable %v", i, body.Enabled, first)
		}
	}
}

func TestEvaluateDisabledAlwaysFalse(t *testing.T) {
	s := NewStore()
	seedFlag(t, s, Flag{Key: "off", Enabled: false, RolloutPercent: 100})
	router := NewRouter(NewServer(s))

	for _, user := range []string{"a", "b", "c"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/flags/off/evaluate?user="+user, nil)
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("user %q: status %d", user, rec.Code)
		}
		if enabled := bodyEnabled(t, rec); enabled {
			t.Fatalf("user %q: enabled=true, want false for disabled flag", user)
		}
	}
}

func TestEvaluateRolloutZeroAlwaysFalse(t *testing.T) {
	s := NewStore()
	seedFlag(t, s, Flag{Key: "zero", Enabled: true, RolloutPercent: 0})
	router := NewRouter(NewServer(s))

	for i := 0; i < 20; i++ {
		user := "user" + string(rune('a'+i))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/flags/zero/evaluate?user="+user, nil)
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("user %q: status %d", user, rec.Code)
		}
		if enabled := bodyEnabled(t, rec); enabled {
			t.Fatalf("user %q: enabled=true, want false for rollout 0", user)
		}
	}
}

func TestEvaluateRolloutHundredAlwaysTrue(t *testing.T) {
	s := NewStore()
	seedFlag(t, s, Flag{Key: "full", Enabled: true, RolloutPercent: 100})
	router := NewRouter(NewServer(s))

	for i := 0; i < 20; i++ {
		user := "user" + string(rune('a'+i))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/flags/full/evaluate?user="+user, nil)
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("user %q: status %d", user, rec.Code)
		}
		if enabled := bodyEnabled(t, rec); !enabled {
			t.Fatalf("user %q: enabled=false, want true for rollout 100", user)
		}
	}
}

func TestEvaluateMissingUser400(t *testing.T) {
	s := NewStore()
	seedFlag(t, s, Flag{Key: "feature", Enabled: true, RolloutPercent: 50})
	router := NewRouter(NewServer(s))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestEvaluateEmptyUser400(t *testing.T) {
	s := NewStore()
	seedFlag(t, s, Flag{Key: "feature", Enabled: true, RolloutPercent: 50})
	router := NewRouter(NewServer(s))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user=", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestEvaluateTooLongUser400(t *testing.T) {
	s := NewStore()
	seedFlag(t, s, Flag{Key: "feature", Enabled: true, RolloutPercent: 50})
	router := NewRouter(NewServer(s))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user="+strings.Repeat("x", 129), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestEvaluateMaxLengthUserAccepted(t *testing.T) {
	s := NewStore()
	seedFlag(t, s, Flag{Key: "feature", Enabled: true, RolloutPercent: 100})
	router := NewRouter(NewServer(s))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user="+strings.Repeat("x", 128), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}

func TestEvaluateUnknownKey404(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/nope/evaluate?user=alice", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rec.Code)
	}
}

func decodeBody(rec *httptest.ResponseRecorder, v any) error {
	return json.Unmarshal(rec.Body.Bytes(), v)
}

func bodyEnabled(t *testing.T, rec *httptest.ResponseRecorder) bool {
	t.Helper()
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := decodeBody(rec, &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body.Enabled
}
