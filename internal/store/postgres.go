package store

import (
	"UrlShortner/internal/codec"
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

func (pgStore *PostgresStore) Save(ctx context.Context, url string) (string, error) {
	tx, err := pgStore.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, "INSERT INTO links (url) VALUES ($1) returning id", url)

	urlId := 0
	err = row.Scan(&urlId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errorhandling.ErrNotFound
		}

		return "", err
	}

	decodedCode := codec.EncodeBase62(uint64(urlId))

	_, err = tx.ExecContext(ctx, "UPDATE links SET code = $1 WHERE id = $2", decodedCode, urlId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errorhandling.ErrNotFound
		}

		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return decodedCode, nil
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{
		db: db,
	}
}
