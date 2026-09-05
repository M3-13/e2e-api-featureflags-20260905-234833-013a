package main

import (
	"log"
	"net/http"
	"time"
)

// statusWriter captures the status code written by the wrapped handler so the
// logging middleware can report it.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware wraps next and logs each request's method, path (without
// the query string), resulting status and duration.
func LoggingMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		logger.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}
