package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddlewareLogsMethodPathStatusDuration(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := LoggingMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		logger,
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	handler.ServeHTTP(rec, req)

	logLine := buf.String()
	if !strings.Contains(logLine, "POST") {
		t.Fatalf("log %q missing method POST", logLine)
	}
	if !strings.Contains(logLine, "/flags") {
		t.Fatalf("log %q missing path /flags", logLine)
	}
	if !strings.Contains(logLine, "201") {
		t.Fatalf("log %q missing status 201", logLine)
	}
	// The duration is the last token; verify it is present and non-empty.
	fields := strings.Fields(logLine)
	if len(fields) < 4 {
		t.Fatalf("log %q has %d fields, want at least 4 (method path status duration)", logLine, len(fields))
	}
	if fields[len(fields)-1] == "" {
		t.Fatalf("log %q missing duration", logLine)
	}
}

func TestLoggingMiddlewareOmitsQueryString(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := LoggingMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		logger,
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user=alice", nil)
	handler.ServeHTTP(rec, req)

	logLine := buf.String()
	if strings.Contains(logLine, "user=alice") {
		t.Fatalf("log %q leaked the query string", logLine)
	}
	if strings.Contains(logLine, "alice") {
		t.Fatalf("log %q leaked the user parameter", logLine)
	}
	if !strings.Contains(logLine, "/flags/feature/evaluate") {
		t.Fatalf("log %q missing path /flags/feature/evaluate", logLine)
	}
}
