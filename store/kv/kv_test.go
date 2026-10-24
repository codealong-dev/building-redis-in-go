package kv

import (
	"codealong/store"
	"errors"
	"reflect"
	"testing"
)

func TestKeys_ReturnsAllKeysSorted(t *testing.T) {
	s := NewStore(10)
	s.Set("charlie", "3")
	s.Set("alpha", "1")
	s.Set("bravo", "2")

	got := s.Keys()
	want := []string{"alpha", "bravo", "charlie"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestKeys_EmptyStore(t *testing.T) {
	s := NewStore(10)

	got := s.Keys()
	if len(got) != 0 {
		t.Errorf("Keys() on empty store = %v, wnat empty slice", got)
	}
}

func TestSetGet_EmptyKeys(t *testing.T) {
	s := NewStore(10)
	if _, err := s.Get(""); err == nil || !errors.Is(err, store.ErrEmptyKey) {
		t.Error("Get() failed")
	}
}

func TestSetGet_RoundTrip(t *testing.T) {
	s := NewStore(10)
	s.Set("hello", "world")

	if val, err := s.Get("hello"); err != nil || val != "world" {
		t.Error("Get() failed")
	}

	if val, err := s.Get("missing"); err == nil || val != "" {
		t.Error("Get() failed")
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name      string
		setup     map[string]string
		deleteKey string
		wantLen   int
	}{
		{"deletes existing key", map[string]string{"a": "1", "b": "2"}, "a", 1},
		{"no-op on missing key", map[string]string{"a": "1"}, "x", 1},
		{"no-op on empty store", map[string]string{}, "x", 0},
		{"deletes last remaining key", map[string]string{"only": "value"}, "only", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := NewStore(10)
			for k, v := range tc.setup {
				store.Set(k, v)
			}

			store.Delete(tc.deleteKey)

			if got := store.Len(); got != tc.wantLen {
				t.Errorf("after Delete(%q): Len() = %d, want %d", tc.deleteKey, got, tc.wantLen)
			}

			if _, err := store.Get(tc.deleteKey); err == nil {
				t.Errorf("after Delete(%q): key still present", tc.deleteKey)
			}
		})
	}
}