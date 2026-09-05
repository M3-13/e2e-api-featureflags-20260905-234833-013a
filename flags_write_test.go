package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newWriteTestServer() *Server {
	return NewServer(NewStore())
}

func TestHandleCreateSuccess(t *testing.T) {
	s := newWriteTestServer()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"feature.x","enabled":true}`))
	s.handleCreate(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	var f Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &f); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if f.Key != "feature.x" || !f.Enabled {
		t.Fatalf("flag = %+v, want key=feature.x enabled=true", f)
	}
	if f.RolloutPercent != 100 {
		t.Fatalf("rollout_percent default = %d, want 100", f.RolloutPercent)
	}
	if f.Description != "" {
		t.Fatalf("description default = %q, want empty", f.Description)
	}

	stored, ok := s.store.Get("feature.x")
	if !ok {
		t.Fatal("flag not stored")
	}
	if stored != f {
		t.Fatalf("stored = %+v, want %+v", stored, f)
	}
}

func TestHandleCreateRolloutDefaultAndProvided(t *testing.T) {
	s := newWriteTestServer()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"a","enabled":true,"description":"d","rollout_percent":30}`))
	s.handleCreate(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	var f Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &f); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if f.Description != "d" || f.RolloutPercent != 30 {
		t.Fatalf("flag = %+v, want description=d rollout=30", f)
	}
}

func TestHandleCreateDuplicate(t *testing.T) {
	s := newWriteTestServer()
	body := `{"key":"dup","enabled":true}`

	rec := httptest.NewRecorder()
	s.handleCreate(rec, httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, want 201", rec.Code)
	}

	rec = httptest.NewRecorder()
	s.handleCreate(rec, httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want 409", rec.Code)
	}
	var e map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("error body not JSON: %v", err)
	}
	if e["error"] == "" {
		t.Fatal("error field missing")
	}
}

func TestHandleCreateInvalidKey(t *testing.T) {
	s := newWriteTestServer()
	cases := []string{
		`{"key":"bad key!","enabled":true}`,
		`{"key":"","enabled":true}`,
		`{"key":"` + strings.Repeat("a", 129) + `","enabled":true}`,
		`{"key":"slash/key","enabled":true}`,
	}
	for _, body := range cases {
		rec := httptest.NewRecorder()
		s.handleCreate(rec, httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body)))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %q status = %d, want 400", body, rec.Code)
		}
	}
}

func TestHandleCreateMissingEnabled(t *testing.T) {
	s := newWriteTestServer()
	rec := httptest.NewRecorder()
	s.handleCreate(rec, httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"k"}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleCreateRolloutOutOfRange(t *testing.T) {
	s := newWriteTestServer()
	cases := []string{
		`{"key":"k","enabled":true,"rollout_percent":101}`,
		`{"key":"k","enabled":true,"rollout_percent":-1}`,
	}
	for _, body := range cases {
		rec := httptest.NewRecorder()
		s.handleCreate(rec, httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body)))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %q status = %d, want 400", body, rec.Code)
		}
	}
}

func TestHandleCreateBodyOver1MiB(t *testing.T) {
	s := newWriteTestServer()
	big := `{"key":"big","enabled":true,"description":"` + strings.Repeat("x", maxBodyBytes) + `"}`
	rec := httptest.NewRecorder()
	s.handleCreate(rec, httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(big)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleCreateInvalidJSON(t *testing.T) {
	s := newWriteTestServer()
	rec := httptest.NewRecorder()
	s.handleCreate(rec, httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`not json`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleUpdateSuccess(t *testing.T) {
	s := newWriteTestServer()
	s.store.Create(Flag{Key: "k", Enabled: true, Description: "old", RolloutPercent: 10})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/k", strings.NewReader(`{"enabled":false,"description":"new","rollout_percent":80}`))
	req.SetPathValue("key", "k")
	s.handleUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var f Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &f); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if f.Key != "k" || f.Enabled || f.Description != "new" || f.RolloutPercent != 80 {
		t.Fatalf("flag = %+v, want key=k enabled=false description=new rollout=80", f)
	}
}

func TestHandleUpdateUnknownKey(t *testing.T) {
	s := newWriteTestServer()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/missing", strings.NewReader(`{"enabled":true}`))
	req.SetPathValue("key", "missing")
	s.handleUpdate(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestHandleUpdateInvalid(t *testing.T) {
	s := newWriteTestServer()
	s.store.Create(Flag{Key: "k", Enabled: true})

	cases := []struct {
		body string
	}{
		{`{"enabled":true,"rollout_percent":101}`},
		{`{"enabled":true,"rollout_percent":-1}`},
		{`{"rollout_percent":50}`}, // missing enabled
		{`not json`},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/flags/k", strings.NewReader(c.body))
		req.SetPathValue("key", "k")
		s.handleUpdate(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %q status = %d, want 400", c.body, rec.Code)
		}
	}
}

func TestHandleUpdateBodyOver1MiB(t *testing.T) {
	s := newWriteTestServer()
	s.store.Create(Flag{Key: "k", Enabled: true})

	big := `{"enabled":true,"description":"` + strings.Repeat("x", maxBodyBytes) + `"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/k", strings.NewReader(big))
	req.SetPathValue("key", "k")
	s.handleUpdate(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleUpdatePartial(t *testing.T) {
	s := newWriteTestServer()
	s.store.Create(Flag{Key: "k", Enabled: true, Description: "keep", RolloutPercent: 40})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/k", strings.NewReader(`{"enabled":false}`))
	req.SetPathValue("key", "k")
	s.handleUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var f Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &f); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if f.Enabled || f.Description != "keep" || f.RolloutPercent != 40 {
		t.Fatalf("flag = %+v, want enabled=false description=keep rollout=40", f)
	}
}
