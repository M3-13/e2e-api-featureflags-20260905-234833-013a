package main

import "net/http"

// handleCreate answers POST /flags. It is registered but not yet implemented:
// the creation ticket fills this in.
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "flag creation not implemented")
}

// handleUpdate answers PUT /flags/{key}. It is registered but not yet
// implemented: the update ticket fills this in.
func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "flag update not implemented")
}
