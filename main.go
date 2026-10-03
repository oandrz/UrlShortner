package main

import (
	"UrlShortner/data"
	"UrlShortner/internal/errorhandling"
	"UrlShortner/internal/store"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

/*
*

	Step 1 — server that boots. main() creates a ServeMux, passes it to http.ListenAndServe(":8080", mux). Zero routes registered. Verify: curl -i localhost:8080/anything returns 404 from the mux itself. Now you know the process is listening and routing works before any logic exists.

	Step 2 — redirect path, hardcoded. Seed the package-level map with one entry by hand, say "abc" to https://go.dev. Register GET /{code}, look up r.PathValue("code"), redirect on hit, 404 on miss. Verify both: curl -i localhost:8080/abc and curl -i localhost:8080/nope. Two of your three acceptance checks pass now, with no JSON written
	yet.

	Step 3 — POST handler. Decode the body into shortenRequest, generate a dumb code (counter is fine), store, encode shortenResponse. Verify with the POST curl.

	Step 4 — full loop. POST, take the returned code, GET it, land on the real URL.
*/

var storage store.Store = store.NewMemStore()

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		url, err := storage.Get(code)
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
	})
	mux.HandleFunc("POST /shorten", func(writer http.ResponseWriter, request *http.Request) {
		shortenerRequest := data.ShortenerRequest{}
		err := json.NewDecoder(request.Body).Decode(&shortenerRequest)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Printf("error decoding shortener request: %v\n", err)
			return
		}
		if shortenerRequest.Url == "" {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Println("url cannot be empty")
			return
		}
		fmt.Printf("shortener request: %v\n", shortenerRequest)

		decodedCode := storage.Save(shortenerRequest.Url)
		response := data.ShortenerResponse{
			Url: decodedCode,
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(writer).Encode(response)
		if err != nil {
			fmt.Printf("error encoding shortener response: %v\n", err)
			return
		}
	})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println(err)
		return
	}
}
