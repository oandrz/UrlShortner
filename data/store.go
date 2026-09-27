package data

import "sync"

type Store struct {
	mu      sync.RWMutex
	counter int
	storage map[string]string
}

func (s *Store) Save(url string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	decodedCode := EncodeBase62(uint64(s.counter))
	s.storage[decodedCode] = url

	return decodedCode
}

func (s *Store) Get(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.storage[code]
	return url, ok
}

func NewStore() *Store {
	return &Store{
		counter: 0,
		storage: make(map[string]string),
	}
}
