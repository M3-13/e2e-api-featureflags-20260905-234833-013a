package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

// Server holds the dependencies shared by all HTTP handlers.
type Server struct {
	store *Store
}

// NewServer builds a Server around the given store.
func NewServer(s *Store) *Server {
	return &Server{store: s}
}

// NewRouter registers every route of the service and returns the handler,
// wrapped in the logging middleware. Unsupported methods on a known path
// answer 405 as JSON with an Allow header.
func NewRouter(s *Server) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("/healthz", methodNotAllowed("GET"))

	mux.HandleFunc("POST /flags", s.handleCreate)
	mux.HandleFunc("GET /flags", s.handleList)
	mux.HandleFunc("/flags", methodNotAllowed("GET, POST"))

	mux.HandleFunc("GET /flags/{key}", s.handleGet)
	mux.HandleFunc("PUT /flags/{key}", s.handleUpdate)
	mux.HandleFunc("DELETE /flags/{key}", s.handleDelete)
	mux.HandleFunc("/flags/{key}", methodNotAllowed("GET, PUT, DELETE"))

	mux.HandleFunc("GET /flags/{key}/evaluate", s.handleEvaluate)
	mux.HandleFunc("/flags/{key}/evaluate", methodNotAllowed("GET"))

	return LoggingMiddleware(mux, log.Default())
}

// methodNotAllowed builds a handler that answers 405 with a JSON error body
// and the given Allow header.
func methodNotAllowed(allow string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router := NewRouter(NewServer(NewStore()))
	srv := &http.Server{
		Addr:              "127.0.0.1:" + port,
		Handler:           withAdminAuth(router, os.Getenv("ADMIN_API_KEY")),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Fatal(srv.ListenAndServe())
}
