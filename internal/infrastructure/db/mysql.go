package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDB is the connection type
type MySQLDB = *sql.DB

// MySQLDsn is a tagged type to avoid string ambiguity in Wire
type MySQLDsn string

func NewMySQLDB(dsn MySQLDsn) (MySQLDB, error) {
	db, err := sql.Open("mysql", string(dsn))
	if err != nil {
		return nil, err
	}

	// pool config (standart config)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mysql ping failed: %w", err)
	}

	return db, nil
}
