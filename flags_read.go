package main

import "net/http"

// handleList answers GET /flags with every stored flag as a JSON array. An
// empty store yields an empty array, never null.
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

// handleGet answers GET /flags/{key} with the single flag, or 404 when the
// key is unknown.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	flag, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

// handleDelete answers DELETE /flags/{key} with 204 and no body on success,
// or 404 when the key is unknown.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if !s.store.Delete(key) {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
