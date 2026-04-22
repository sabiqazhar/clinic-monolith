package db

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx as database/sql driver
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

// NewPostgresSQLDB creates a *sql.DB for use with database/sql interfaces (e.g., OutboxRelay).
// Note: This creates a new connection pool separate from the pgxpool.
func NewPostgresSQLDB(dsn PGDsn) (*sql.DB, error) {
	db, err := sql.Open("pgx", string(dsn))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
