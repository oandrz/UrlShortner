package store

import (
	"UrlShortner/internal/codec"
	"UrlShortner/internal/errorhandling"
	"sync"
)

type Store interface {
	Save(url string) string
	Get(code string) (string, error)
}

type MemStore struct {
	mu      sync.RWMutex
	counter int
	storage map[string]string
}

func (s *MemStore) Save(url string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	decodedCode := codec.EncodeBase62(uint64(s.counter))
	s.storage[decodedCode] = url

	return decodedCode
}

func (s *MemStore) Get(code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.storage[code]
	if !ok {
		return "", errorhandling.ErrNotFound
	}
	return url, nil
}

func NewMemStore() *MemStore {
	return &MemStore{
		counter: 0,
		storage: make(map[string]string),
	}
}
