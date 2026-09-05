package main

import "net/http"

// handleList answers GET /flags. It is registered but not yet implemented:
// the read ticket fills this in.
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "flag list not implemented")
}

// handleGet answers GET /flags/{key}. It is registered but not yet
// implemented: the read ticket fills this in.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "flag get not implemented")
}

// handleDelete answers DELETE /flags/{key}. It is registered but not yet
// implemented: the read ticket fills this in.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "flag delete not implemented")
}
