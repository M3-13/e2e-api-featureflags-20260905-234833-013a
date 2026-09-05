package main

import (
	"hash/fnv"
	"net/http"
)

// maxUserLength is the maximum allowed length of the user query parameter.
const maxUserLength = 128

// separator joins key and user before hashing, so ("a","bc") and ("ab","c")
// do not collide into the same bucket.
const separator = ":"

// handleEvaluate answers GET /flags/{key}/evaluate?user={id}. The user id is
// mapped onto a stable 0-99 bucket via an FNV-1a-64 hash over key + separator
// + user, so repeated calls for the same key and user return the same result.
func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" || len(user) > maxUserLength {
		writeError(w, http.StatusBadRequest, "user must be 1-128 characters")
		return
	}

	key := r.PathValue("key")
	flag, ok := s.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"enabled": evaluateFlag(flag, key, user)})
}

// evaluateFlag decides whether the flag is enabled for the given user.
func evaluateFlag(f Flag, key, user string) bool {
	if !f.Enabled {
		return false
	}
	if f.RolloutPercent >= 100 {
		return true
	}
	if f.RolloutPercent <= 0 {
		return false
	}
	return bucket(key, user) < uint64(f.RolloutPercent)
}

// bucket returns a stable 0-99 value for the key/user pair.
func bucket(key, user string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte(separator))
	_, _ = h.Write([]byte(user))
	return h.Sum64() % 100
}
