package handler

import (
	"UrlShortner/internal/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectGetURLNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nope", nil)
	memStore := store.NewMemStore()

	handler := NewHandlerConfig(memStore)
	NewRouter(&handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want %d", rec.Code, http.StatusNotFound)
	}
	t.Logf("body: %q", rec.Body.String())
}

func TestRedirectGetExistingCode(t *testing.T) {
	ctx := t.Context()
	rec := httptest.NewRecorder()
	memStore := store.NewMemStore()

	code, err := memStore.Save(ctx, "https://go.dev")
	if err != nil {
		t.Fatalf("save data into memstore failed")
	}
	req := httptest.NewRequest("GET", "/"+code, nil)

	handler := NewHandlerConfig(memStore)
	NewRouter(&handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("got %d, want %d", rec.Code, http.StatusFound)
	}

	location := rec.Header().Get("Location")
	if location != "https://go.dev" {
		t.Errorf("got %v, want %v", location, "https://go.dev")
	}

	t.Logf("body: %q", rec.Body.String())
}
