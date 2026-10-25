package handler

import "net/http"

func NewRouter(h *RedirectHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		h.RedirectGetURLBasedCode(w, r)
	})
	mux.HandleFunc("POST /shorten", func(writer http.ResponseWriter, request *http.Request) {
		h.RedirectShortenURLCompute(writer, request)
	})

	return mux
}
