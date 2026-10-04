package main

import (
	"UrlShortner/internal/handler"
	"UrlShortner/internal/store"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

/*
*

	Step 1 — server that boots. main() creates a ServeMux, passes it to http.ListenAndServe(":8080", mux). Zero routes registered. Verify: curl -i localhost:8080/anything returns 404 from the mux itself. Now you know the process is listening and routing works before any logic exists.

	Step 2 — redirect path, hardcoded. Seed the package-level map with one entry by hand, say "abc" to https://go.dev. Register GET /{code}, look up r.PathValue("code"), redirect on hit, 404 on miss. Verify both: curl -i localhost:8080/abc and curl -i localhost:8080/nope. Two of your three acceptance checks pass now, with no JSON written
	yet.

	Step 3 — POST handler. Decode the body into shortenRequest, generate a dumb code (counter is fine), store, encode shortenResponse. Verify with the POST curl.

	Step 4 — full loop. POST, take the returned code, GET it, land on the real URL.
*/

func main() {

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	var storage store.Store = store.NewPostgresStore(db)
	var redirectHandler = handler.NewHandlerConfig(storage)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		redirectHandler.RedirectGetURLBasedCode(w, r)
	})
	mux.HandleFunc("POST /shorten", func(writer http.ResponseWriter, request *http.Request) {
		redirectHandler.RedirectShortenURLCompute(writer, request)
	})

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println(err)
		return
	}
}
