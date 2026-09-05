package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

// maxBodyBytes limits the request body of POST /flags and PUT /flags/{key} to
// 1 MiB, read before buffering so oversized requests are rejected outright.
const maxBodyBytes = 1 << 20

// keyPattern matches a valid flag key: 1–128 characters of [A-Za-z0-9._-].
var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// validKey reports whether key is 1–128 characters long and consists only of
// [A-Za-z0-9._-].
func validKey(key string) bool {
	if len(key) < 1 || len(key) > 128 {
		return false
	}
	return keyPattern.MatchString(key)
}

// flagCreateRequest is the JSON body of POST /flags. Pointer fields
// distinguish an omitted field from a zero value so that "enabled" can be
// required and "rollout_percent" can default to 100.
type flagCreateRequest struct {
	Key            string  `json:"key"`
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

// flagUpdateRequest is the JSON body of PUT /flags/{key}. The key is not part
// of the body; it is taken from the path and cannot be changed.
type flagUpdateRequest struct {
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

// handleCreate answers POST /flags. It validates the body, creates the flag
// and answers 201 with the stored flag, 409 for a duplicate key or 400 for an
// invalid body.
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req flagCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !validKey(req.Key) {
		writeError(w, http.StatusBadRequest, "invalid key")
		return
	}
	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	rollout := 100
	if req.RolloutPercent != nil {
		rollout = *req.RolloutPercent
		if rollout < 0 || rollout > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
	}

	flag := Flag{
		Key:            req.Key,
		Enabled:        *req.Enabled,
		Description:    description,
		RolloutPercent: rollout,
	}

	if err := s.store.Create(flag); err != nil {
		writeError(w, http.StatusConflict, "flag already exists")
		return
	}

	writeJSON(w, http.StatusCreated, flag)
}

// handleUpdate answers PUT /flags/{key}. The key is taken from the path and is
// immutable; enabled is required while description and rollout_percent are
// optional (an omitted field keeps its previous value). It answers 200 with the
// updated flag, 404 for an unknown key or 400 for an invalid body.
func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	existing, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req flagUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}

	description := existing.Description
	if req.Description != nil {
		description = *req.Description
	}

	rollout := existing.RolloutPercent
	if req.RolloutPercent != nil {
		rollout = *req.RolloutPercent
		if rollout < 0 || rollout > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
	}

	flag := Flag{
		Key:            key,
		Enabled:        *req.Enabled,
		Description:    description,
		RolloutPercent: rollout,
	}

	updated, ok := s.store.Update(key, flag)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}
