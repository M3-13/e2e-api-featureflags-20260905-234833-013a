package main

import (
	"errors"
	"sync"
)

// ErrFlagExists is returned by Store.Create when a flag with the same key
// already exists.
var ErrFlagExists = errors.New("flag already exists")

// Flag is a single feature flag.
type Flag struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

// Store is a thread-safe in-memory feature-flag store.
type Store struct {
	mu    sync.Mutex
	flags map[string]Flag
}

// NewStore returns an empty, ready-to-use Store.
func NewStore() *Store {
	return &Store{flags: make(map[string]Flag)}
}

// Create adds a flag, returning ErrFlagExists if the key is already present.
func (s *Store) Create(f Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[f.Key]; ok {
		return ErrFlagExists
	}
	s.flags[f.Key] = f
	return nil
}

// List returns all flags. The order is not defined.
func (s *Store) List() []Flag {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		result = append(result, f)
	}
	return result
}

// Get returns the flag with the given key and whether it exists.
func (s *Store) Get(key string) (Flag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.flags[key]
	return f, ok
}

// Update replaces the flag with the given key, returning the stored flag and
// whether the key existed.
func (s *Store) Update(key string, f Flag) (Flag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return Flag{}, false
	}
	s.flags[key] = f
	return f, true
}

// Delete removes the flag with the given key, returning whether it existed.
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return false
	}
	delete(s.flags, key)
	return true
}
