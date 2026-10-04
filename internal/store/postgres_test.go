package store

import (
	"UrlShortner/internal/codec"
	"UrlShortner/internal/errorhandling"
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// openTestDB connects to the Postgres named by DATABASE_URL. Tests that need a
// real database skip themselves when it is not set, so `go test ./...` still
// passes without Docker running.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres test")
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping database: %v", err)
	}
	return db
}

// TestPostgresStoreGet inserts a row directly, so it checks Get without
// depending on Save. The code contains "-", which base62 never produces, so it
// cannot collide with a generated code.
func TestPostgresStoreGet(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	const code, want = "test-get", "https://example.com/get"
	cleanup := func() { db.ExecContext(ctx, `DELETE FROM links WHERE code = $1`, code) }
	cleanup()
	t.Cleanup(cleanup)

	if _, err := db.ExecContext(ctx, `INSERT INTO links (code, url) VALUES ($1, $2)`, code, want); err != nil {
		t.Fatalf("insert test row: %v", err)
	}

	store := NewPostgresStore(db)

	got, err := store.Get(ctx, code)
	if err != nil {
		t.Errorf("Get(%q) error: %v, want %q", code, err, want)
	} else if got != want {
		t.Errorf("Get(%q) = %q, want %q", code, got, want)
	}

	_, err = store.Get(ctx, "no-such-code")
	if !errors.Is(err, errorhandling.ErrNotFound) {
		t.Errorf("Get(missing) error = %v, want ErrNotFound", err)
	}
}

// TestPostgresStoreSave checks that the code is the base62 of the row's id,
// that Get finds what Save stored, and that no row is left without a code.
func TestPostgresStoreSave(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	const first, second = "https://example.com/save-1", "https://example.com/save-2"
	cleanup := func() { db.ExecContext(ctx, `DELETE FROM links WHERE url IN ($1, $2)`, first, second) }
	cleanup()
	t.Cleanup(cleanup)

	store := NewPostgresStore(db)

	code, err := store.Save(ctx, first)
	if err != nil {
		t.Fatalf("Save(%q) error: %v", first, err)
	}
	if code == "" {
		t.Fatalf("Save(%q) returned an empty code", first)
	}

	var id uint64
	if err := db.QueryRowContext(ctx, `SELECT id FROM links WHERE code = $1`, code).Scan(&id); err != nil {
		t.Fatalf("no row with code %q after Save: %v", code, err)
	}
	if want := codec.EncodeBase62(id); code != want {
		t.Errorf("Save returned code %q for id %d, want %q", code, id, want)
	}

	got, err := store.Get(ctx, code)
	if err != nil {
		t.Errorf("Get(%q) error: %v, want %q", code, err, first)
	} else if got != first {
		t.Errorf("Get(%q) = %q, want %q", code, got, first)
	}

	otherCode, err := store.Save(ctx, second)
	if err != nil {
		t.Fatalf("Save(%q) error: %v", second, err)
	}
	if otherCode == code {
		t.Errorf("two Saves returned the same code %q", code)
	}

	var missing int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM links WHERE code IS NULL AND url IN ($1, $2)`, first, second).Scan(&missing); err != nil {
		t.Fatalf("count rows without a code: %v", err)
	}
	if missing != 0 {
		t.Errorf("%d saved rows have no code", missing)
	}
}

// TestPostgresStoreSaveRunsOnTx gives the store a pool of exactly one
// connection. The transaction takes that connection, so a statement run on db
// instead of tx waits for a second one that never comes, and Save times out.
func TestPostgresStoreSaveRunsOnTx(t *testing.T) {
	admin := openTestDB(t)

	const url = "https://example.com/save-one-conn"
	cleanup := func() { admin.ExecContext(context.Background(), `DELETE FROM links WHERE url = $1`, url) }
	cleanup()
	t.Cleanup(cleanup)

	db := openTestDB(t)
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := NewPostgresStore(db).Save(ctx, url); err != nil {
		t.Fatalf("Save with a one-connection pool: %v (is a statement running on db instead of tx?)", err)
	}
}

// TestPostgresStoreSaveRollsBackOnError makes the INSERT fail (Postgres rejects
// a zero byte in TEXT) and checks that the transaction gave its connection
// back. A transaction that is never rolled back holds its connection forever.
func TestPostgresStoreSaveRollsBackOnError(t *testing.T) {
	db := openTestDB(t)

	if _, err := NewPostgresStore(db).Save(context.Background(), "https://example.com/\x00"); err == nil {
		t.Fatal("Save with a zero byte in the URL succeeded, want an error")
	}

	if inUse := db.Stats().InUse; inUse != 0 {
		t.Errorf("%d connection(s) still in use after a failed Save, want 0 (missing tx.Rollback?)", inUse)
	}
}
