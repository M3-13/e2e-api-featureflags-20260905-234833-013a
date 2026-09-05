package main

import "net/http"

// handleEvaluate answers GET /flags/{key}/evaluate. It is registered but not
// yet implemented: the evaluation ticket fills this in.
func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "flag evaluation not implemented")
}
