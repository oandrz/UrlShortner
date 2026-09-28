package data

import (
	"fmt"
	"sync"
	"testing"
)

// TestStoreConcurrentAccess hammers one MemStore from many goroutines at once:
// each iteration starts one writer (Save) and one reader (Get), so reads and
// writes overlap. Run it with `go test -race ./...`.
func TestStoreConcurrentAccess(t *testing.T) {
	const n = 1000

	store := NewMemStore()
	codes := make([]string, n) // each goroutine writes only its own index, so the test itself is race-free

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			codes[i] = store.Save(fmt.Sprintf("https://example.com/%d", i))
		}()

		go func() {
			defer wg.Done()
			store.Get(EncodeBase62(uint64(i + 1)))
		}()
	}
	wg.Wait()

	// Every Save must hand out its own code, and every code must lead back to its own URL.
	seen := make(map[string]int, n)
	for i, code := range codes {
		if prev, dup := seen[code]; dup {
			t.Errorf("Save #%d and Save #%d both got code %q", prev, i, code)
		}
		seen[code] = i

		want := fmt.Sprintf("https://example.com/%d", i)
		got, err := store.Get(code)
		if err != nil {
			t.Errorf("Get(%q) error: %v, want %q", code, err, want)
		} else if got != want {
			t.Errorf("Get(%q) = %q, want %q (link overwritten)", code, got, want)
		}
	}
}
