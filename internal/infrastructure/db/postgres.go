package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxPool is the actual type for PostgreSQL connection pool
type PgxPool = *pgxpool.Pool

// PGDsn is a tagged type to avoid string ambiguity in Wire
type PGDsn string

func NewPostgresPool(dsn PGDsn) (PgxPool, error) {
	pool, err := pgxpool.New(context.Background(), string(dsn))
	if err != nil {
		return nil, err
	}
	return pool, nil
}
