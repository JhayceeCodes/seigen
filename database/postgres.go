package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgres opens and verifies a PostgreSQL database connection.
func NewPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
