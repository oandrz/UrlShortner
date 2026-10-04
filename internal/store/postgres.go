package store

import (
	"UrlShortner/internal/errorhandling"
	"context"
	"database/sql"
	"errors"
)

type PostgresStore struct {
	db *sql.DB
}

func (pgStore *PostgresStore) Get(ctx context.Context, code string) (string, error) {
	row := pgStore.db.QueryRowContext(ctx, "SELECT url from links WHERE code = $1", code)
	url := ""
	err := row.Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errorhandling.ErrNotFound
		}

		return "", err
	}

	return url, nil
}

func NewPostgresStore(db *sql.DB) PostgresStore {
	return PostgresStore{
		db: db,
	}
}
