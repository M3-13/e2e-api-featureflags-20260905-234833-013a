package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

// maxBodyBytes bounds the request body read by the write handlers to 1 MiB so
// an oversized body is rejected before it is buffered.
const maxBodyBytes = 1 << 20

// keyPattern matches a flag key consisting only of [A-Za-z0-9._-].
var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// maxKeyLength is the maximum allowed length of a flag key.
const maxKeyLength = 128

// maxDescriptionLength bounds a flag description to 500 characters. An empty
// description is allowed; only a non-empty value longer than the limit is
// rejected.
const maxDescriptionLength = 500

// validDescription reports whether s is a valid flag description: empty is
// allowed, otherwise its length must not exceed maxDescriptionLength.
func validDescription(s string) bool {
	return len(s) <= maxDescriptionLength
}

// validJSONContentType reports whether the Content-Type header declares a JSON
// body. It accepts the plain type and the common charset-suffixed form.
func validJSONContentType(ct string) bool {
	switch ct {
	case "application/json", "application/json; charset=utf-8":
		return true
	}
	return false
}

// validKey reports whether key is 1-128 characters long and consists only of
// [A-Za-z0-9._-].
func validKey(key string) bool {
	if key == "" || len(key) > maxKeyLength {
		return false
	}
	return keyPattern.MatchString(key)
}

// createFlagRequest is the JSON body accepted by POST /flags. Pointer fields
// distinguish a missing field from its zero value.
type createFlagRequest struct {
	Key            *string `json:"key"`
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

// updateFlagRequest is the JSON body accepted by PUT /flags/{key}. The key is
// not part of the body: it comes from the path and cannot be changed.
type updateFlagRequest struct {
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

// handleCreate answers POST /flags. It validates the body, stores the flag and
// answers 201 with the created flag, 409 for a duplicate key, or 400 for an
// invalid body.
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if !validJSONContentType(r.Header.Get("Content-Type")) {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported media type")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req createFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusBadRequest, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Key == nil || !validKey(*req.Key) {
		writeError(w, http.StatusBadRequest, "key must be 1-128 characters of [A-Za-z0-9._-]")
		return
	}
	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}

	rolloutPercent := 100
	if req.RolloutPercent != nil {
		rolloutPercent = *req.RolloutPercent
	}
	if rolloutPercent < 0 || rolloutPercent > 100 {
		writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}
	if !validDescription(description) {
		writeError(w, http.StatusBadRequest, "description too long")
		return
	}

	flag := Flag{
		Key:            *req.Key,
		Enabled:        *req.Enabled,
		Description:    description,
		RolloutPercent: rolloutPercent,
	}

	if err := s.store.Create(flag); err != nil {
		if err == ErrFlagExists {
			writeError(w, http.StatusConflict, "flag already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, flag)
}

// handleUpdate answers PUT /flags/{key}. The key comes from the path and is not
// changeable. Optional fields that are omitted keep their existing values. It
// answers 200 with the updated flag, 404 for an unknown key, or 400 for an
// invalid body.
func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if !validKey(key) {
		writeError(w, http.StatusBadRequest, "key must be 1-128 characters of [A-Za-z0-9._-]")
		return
	}

	if !validJSONContentType(r.Header.Get("Content-Type")) {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported media type")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req updateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusBadRequest, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}
	if req.RolloutPercent != nil && (*req.RolloutPercent < 0 || *req.RolloutPercent > 100) {
		writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
		return
	}

	if req.Description != nil && !validDescription(*req.Description) {
		writeError(w, http.StatusBadRequest, "description too long")
		return
	}

	existing, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	existing.Enabled = *req.Enabled
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.RolloutPercent != nil {
		existing.RolloutPercent = *req.RolloutPercent
	}

	updated, ok := s.store.Update(key, existing)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}
