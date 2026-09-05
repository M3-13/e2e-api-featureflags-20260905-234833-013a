package main

import (
	"strconv"
	"sync"
	"testing"
)

func TestStoreCreateAndGet(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "feature.x", Enabled: true, Description: "desc", RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, ok := s.Get("feature.x")
	if !ok {
		t.Fatal("Get: flag not found")
	}
	if got != f {
		t.Fatalf("Get = %+v, want %+v", got, f)
	}
}

func TestStoreCreateDuplicate(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "dup", Enabled: true}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Create(f); err != ErrFlagExists {
		t.Fatalf("second Create err = %v, want ErrFlagExists", err)
	}
}

func TestStoreList(t *testing.T) {
	s := NewStore()
	if got := s.List(); len(got) != 0 {
		t.Fatalf("empty List len = %d, want 0", len(got))
	}
	_ = s.Create(Flag{Key: "a", Enabled: true})
	_ = s.Create(Flag{Key: "b", Enabled: false})
	if got := s.List(); len(got) != 2 {
		t.Fatalf("List len = %d, want 2", len(got))
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "k", Enabled: true})
	updated, ok := s.Update("k", Flag{Key: "k", Enabled: false, RolloutPercent: 30})
	if !ok {
		t.Fatal("Update: flag not found")
	}
	if updated.Enabled || updated.RolloutPercent != 30 {
		t.Fatalf("Update = %+v", updated)
	}
	if _, ok := s.Update("missing", Flag{Key: "missing"}); ok {
		t.Fatal("Update on missing key should return false")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "k", Enabled: true})
	if !s.Delete("k") {
		t.Fatal("Delete: expected true")
	}
	if _, ok := s.Get("k"); ok {
		t.Fatal("Get after Delete should be false")
	}
	if s.Delete("k") {
		t.Fatal("Delete on missing key should return false")
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(i%20)
			switch i % 4 {
			case 0:
				_ = s.Create(Flag{Key: key, Enabled: i%2 == 0})
			case 1:
				_, _ = s.Get(key)
			case 2:
				_ = s.List()
			case 3:
				s.Update(key, Flag{Key: key, Enabled: true})
				s.Delete(key)
			}
		}(i)
	}
	wg.Wait()
}
