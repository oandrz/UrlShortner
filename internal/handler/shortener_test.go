package handler

import (
	"UrlShortner/internal/store"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type failingStore struct {
}

func (failingStore) Save(ctx context.Context, url string) (string, error) {
	return "", fmt.Errorf("dummy")
}

func (failingStore) Get(ctx context.Context, code string) (string, error) {
	return "", fmt.Errorf("dummy")
}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name       string
		storage    Store
		body       string
		wantStatus int
	}{
		{name: "valid", storage: store.NewMemStore(), body: "{ \"url\": \"https://go.dev\"}", wantStatus: 201},
		{name: "malformed json", storage: store.NewMemStore(), body: "{ \"url\": ", wantStatus: 400},
		{name: "empty url", storage: store.NewMemStore(), body: "{ \"url\": \"\"}", wantStatus: 422},
		{name: "javascript scheme", storage: store.NewMemStore(), body: "{ \"url\": \"javascript:alert(1)\"}", wantStatus: 422},
		{name: "ftp scheme", storage: store.NewMemStore(), body: "{ \"url\": \"ftp://x.com\"}", wantStatus: 422},
		{name: "no host scheme", storage: store.NewMemStore(), body: "{ \"url\": \"https://\"}", wantStatus: 422},
		{name: "not a url", storage: store.NewMemStore(), body: "{ \"url\": \"just a string\"}", wantStatus: 422},
		{name: "store fails", storage: failingStore{}, body: "{ \"url\": \"https://go.dev\"}", wantStatus: 500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler := NewHandlerConfig(tc.storage)
			req := httptest.NewRequest("POST", "/shorten", strings.NewReader(tc.body))

			NewRouter(&handler).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("got %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

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
