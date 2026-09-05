package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func postBody(t *testing.T, router http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	return rec
}

func putBody(t *testing.T, router http.Handler, key, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/"+key, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	return rec
}

func TestHandleCreateSuccess(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := postBody(t, router, `{"key":"alpha","enabled":true,"description":"first","rollout_percent":75}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body %q)", rec.Code, rec.Body.String())
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
	if flag.RolloutPercent != 75 {
		t.Fatalf("rollout_percent = %d, want 75", flag.RolloutPercent)
	}
}

func TestHandleCreateDefaultsRolloutPercent(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := postBody(t, router, `{"key":"alpha","enabled":true}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body %q)", rec.Code, rec.Body.String())
	}
	var flag Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flag); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if flag.RolloutPercent != 100 {
		t.Fatalf("rollout_percent = %d, want 100", flag.RolloutPercent)
	}
	if flag.Description != "" {
		t.Fatalf("description = %q, want empty", flag.Description)
	}
}

func TestHandleCreateDuplicate(t *testing.T) {
	store := NewStore()
	if err := store.Create(Flag{Key: "alpha", Enabled: true, RolloutPercent: 100}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	router := NewRouter(NewServer(store))
	rec := postBody(t, router, `{"key":"alpha","enabled":true}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatal("error field missing from 409 body")
	}
}

func TestHandleCreateInvalidKey(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := postBody(t, router, `{"key":"not valid!","enabled":true}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleCreateEmptyKey(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := postBody(t, router, `{"key":"","enabled":true}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleCreateTooLongKey(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := postBody(t, router, `{"key":"`+strings.Repeat("a", 129)+`","enabled":true}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleCreateMissingEnabled(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := postBody(t, router, `{"key":"alpha"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleCreateRolloutPercentOutOfRange(t *testing.T) {
	for _, rp := range []int{-1, 101} {
		router := NewRouter(NewServer(NewStore()))
		rec := postBody(t, router, jsonRollout(rp))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent %d: status = %d, want 400", rp, rec.Code)
		}
	}
}

func TestHandleCreateBodyTooLarge(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	// A body well over 1 MiB that is still valid JSON-ish padding.
	huge := `{"key":"alpha","enabled":true,"description":"` + strings.Repeat("x", maxBodyBytes) + `"}`
	rec := postBody(t, router, huge)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleCreateInvalidJSON(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := postBody(t, router, `{not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleUpdateSuccess(t *testing.T) {
	store := NewStore()
	if err := store.Create(Flag{Key: "alpha", Enabled: true, Description: "old", RolloutPercent: 10}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	router := NewRouter(NewServer(store))
	rec := putBody(t, router, "alpha", `{"enabled":false,"description":"new","rollout_percent":50}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}
	var flag Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flag); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if flag.Key != "alpha" {
		t.Fatalf("key = %q, want alpha (key must not change)", flag.Key)
	}
	if flag.Enabled {
		t.Fatal("enabled = true, want false")
	}
	if flag.Description != "new" {
		t.Fatalf("description = %q, want new", flag.Description)
	}
	if flag.RolloutPercent != 50 {
		t.Fatalf("rollout_percent = %d, want 50", flag.RolloutPercent)
	}
}

func TestHandleUpdatePartialKeepsExisting(t *testing.T) {
	store := NewStore()
	if err := store.Create(Flag{Key: "alpha", Enabled: true, Description: "keep", RolloutPercent: 42}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	router := NewRouter(NewServer(store))
	rec := putBody(t, router, "alpha", `{"enabled":false}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var flag Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flag); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if flag.Enabled {
		t.Fatal("enabled = true, want false")
	}
	if flag.Description != "keep" {
		t.Fatalf("description = %q, want keep (omitted field must be preserved)", flag.Description)
	}
	if flag.RolloutPercent != 42 {
		t.Fatalf("rollout_percent = %d, want 42 (omitted field must be preserved)", flag.RolloutPercent)
	}
}

func TestHandleUpdateUnknownKey(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := putBody(t, router, "missing", `{"enabled":true}`)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestHandleUpdateInvalidPathKey(t *testing.T) {
	router := NewRouter(NewServer(NewStore()))
	rec := putBody(t, router, "invalid!", `{"enabled":true}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleUpdateMissingEnabled(t *testing.T) {
	store := NewStore()
	if err := store.Create(Flag{Key: "alpha", Enabled: true, RolloutPercent: 100}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	router := NewRouter(NewServer(store))
	rec := putBody(t, router, "alpha", `{"description":"x"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleUpdateRolloutPercentOutOfRange(t *testing.T) {
	store := NewStore()
	if err := store.Create(Flag{Key: "alpha", Enabled: true, RolloutPercent: 100}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	router := NewRouter(NewServer(store))
	rec := putBody(t, router, "alpha", `{"enabled":true,"rollout_percent":200}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleUpdateBodyTooLarge(t *testing.T) {
	store := NewStore()
	if err := store.Create(Flag{Key: "alpha", Enabled: true, RolloutPercent: 100}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	router := NewRouter(NewServer(store))
	huge := `{"enabled":true,"description":"` + strings.Repeat("x", maxBodyBytes) + `"}`
	rec := putBody(t, router, "alpha", huge)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

// jsonRollout builds a request body with the given rollout_percent value.
func jsonRollout(rp int) string {
	return `{"key":"alpha","enabled":true,"rollout_percent":` + strconv.Itoa(rp) + `}`
}
