package handler_test

import (
	"UrlShortner/internal/handler"
	"UrlShortner/internal/store"
)

// Compile-time checks: the build fails here if either store stops satisfying
// the interface the handlers need. Nothing runs; the blank identifier throws
// the values away.
var (
	_ handler.Store = (*store.MemStore)(nil)
	_ handler.Store = (*store.PostgresStore)(nil)
)
