package kv

import (
	"codealong/store"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Reads massively outnumber writes in a KV store, so a plain Mutex serializes
// traffic that doesn't actually conflict. RWMutex lets unlimited readers run
// in parallel, and only blocks them when a writer wants in.
//
// Note: we stop embedding the lock and name it `mu`. Embedding would leak
// RLock/RUnlock onto *Store, which is noisy and easy to misuse.
type Store struct {
	mu      sync.RWMutex
	data    map[string]string // maps are not concurrent-safe
	maxSize int               // 0 = unlimited
}

func NewStore(maxSize int) *Store {
	return &Store{
		data:    make(map[string]string),
		maxSize: maxSize,
	}
}

func (s *Store) Clone() *Store {
	cp := &Store{
		data:    make(map[string]string, len(s.data)),
		maxSize: s.maxSize,
	}

	for key, value := range s.data {
		cp.data[key] = value
	}

	return cp
}

// Len is read-only — RLock lets concurrent readers in.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data)
}

// Keys returns every key currently in the store, sorted alphabetically.
// Read-only path: RLock.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// Get is the hot path in a real KV store — under RLock, thousands of
// goroutines can read at the same time without blocking each other.
func (s *Store) Get(key string) (string, error) {
	if key == "" {
		return "", store.ErrEmptyKey
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]
	if !ok {
		return "", fmt.Errorf("key %s does not exists", key)
	}

	return val, nil
}

// Set mutates — exclusive Lock. All readers wait until we're done.
func (s *Store) Set(key, value string) error {
	if key == "" {
		return store.ErrEmptyKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Call len(s.data) inline — calling s.Len() here would try to acquire
	// the lock a second time and deadlock. A held lock is not reentrant.
	_, exists := s.data[key]
	if s.maxSize > 0 && len(s.data) >= s.maxSize && !exists {
		return fmt.Errorf("Set(%q): %w", key, store.ErrStoreFull)
	}

	s.data[key] = value
	return nil
}

// Delete mutates — exclusive Lock.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
}

// Incr increments the integer value stored at key by 1.
// If the key doesn't exist, it's initialized to 0 before incrementing.
// Mirrors Redis' INCR command.
func (s *Store) Incr(key string) (int, error) {
	if key == "" {
		return 0, store.ErrEmptyKey
	}

	// Incr is read-modify-write. It's tempting to RLock the read half,
	// but two goroutines holding RLock would both read 5, both write 6,
	// and you'd lose increments. Any path that writes takes a full Lock.
	s.mu.Lock()
	defer s.mu.Unlock()

	// READ
	current := 0
	if raw, ok := s.data[key]; ok {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return 0, fmt.Errorf("INCR %q: %w", key, store.ErrNotInteger)
		}
		current = n
	}

	// MODIFY
	current++
	time.Sleep(time.Microsecond) // simulate some work

	// WRITE
	s.data[key] = strconv.Itoa(current)
	return current, nil
}
