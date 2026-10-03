package handler

import (
	"UrlShortner/internal/errorhandling"
	"UrlShortner/internal/store"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ShortenerRequest struct {
	Url string `json:"url"`
}

type ShortenerResponse struct {
	Url string `json:"url"`
}

type RedirectHandler struct {
	storage store.Store
}

func (h *RedirectHandler) RedirectGetURLBasedCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	url, err := h.storage.Get(code)
	if err != nil {
		if errors.Is(err, errorhandling.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Println("URL Not Found")
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Something wrong with the server")
		return
	}

	http.Redirect(w, r, url, 302)
}

func (h *RedirectHandler) RedirectShortenURLCompute(w http.ResponseWriter, r *http.Request) {
	shortenerRequest := ShortenerRequest{}
	err := json.NewDecoder(r.Body).Decode(&shortenerRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("error decoding shortener request: %v\n", err)
		return
	}
	if shortenerRequest.Url == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Println("url cannot be empty")
		return
	}
	fmt.Printf("shortener request: %v\n", shortenerRequest)

	decodedCode := h.storage.Save(shortenerRequest.Url)
	response := ShortenerResponse{
		Url: decodedCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Printf("error encoding shortener response: %v\n", err)
		return
	}
}

func NewHandlerConfig(param store.Store) RedirectHandler {
	return RedirectHandler{
		storage: param,
	}
}
