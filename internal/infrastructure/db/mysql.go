package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDB is the connection type
type MySQLDB = *sql.DB

// MySQLDsn is a tagged type to avoid string ambiguity in Wire
type MySQLDsn string

// ConvertURLToDSN converts mysql:// URL format to native DSN format
func ConvertURLToDSN(mysqlURL string) string {
	if !strings.HasPrefix(mysqlURL, "mysql://") {
		return mysqlURL
	}

	if strings.Contains(mysqlURL, "@tcp(") {
		urlWithoutScheme := strings.TrimPrefix(mysqlURL, "mysql://")

		atIdx := strings.Index(urlWithoutScheme, "@tcp(")
		userPass := urlWithoutScheme[:atIdx]
		rest := urlWithoutScheme[atIdx+1:]

		tcpEnd := strings.Index(rest, ")")
		hostPort := rest[len("tcp("):tcpEnd]
		rest = rest[tcpEnd+1:]

		dbIdx := strings.Index(rest, "/")
		dbName := ""
		query := ""
		if dbIdx >= 0 {
			dbName = rest[dbIdx+1:]
			rest = rest[:dbIdx]
			if strings.HasPrefix(rest, "?") || strings.HasPrefix(rest, "&") {
				query = rest[1:]
			}
		}

		parts := strings.Split(userPass, ":")
		user := userPass
		pass := ""
		if len(parts) >= 2 {
			user = parts[0]
			pass = strings.Join(parts[1:], ":")
		}

		dsn := user + ":" + pass + "@tcp(" + hostPort + ")/" + dbName
		if query != "" {
			dsn += "?" + query
		}
		return dsn
	}

	return mysqlURL
}

func NewMySQLDB(dsn MySQLDsn) (MySQLDB, error) {
	nativeDSN := ConvertURLToDSN(string(dsn))
	db, err := sql.Open("mysql", nativeDSN)
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
