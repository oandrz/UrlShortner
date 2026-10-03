package store

import (
	"UrlShortner/internal/codec"
	"UrlShortner/internal/errorhandling"
	"context"
	"sync"
)

type Store interface {
	Save(ctx context.Context, url string) (string, error)
	Get(ctx context.Context, code string) (string, error)
}

type MemStore struct {
	mu      sync.RWMutex
	counter int
	storage map[string]string
}

func (s *MemStore) Save(ctx context.Context, url string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	decodedCode := codec.EncodeBase62(uint64(s.counter))
	s.storage[decodedCode] = url

	return decodedCode, nil
}

func (s *MemStore) Get(ctx context.Context, code string) (string, error) {
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
