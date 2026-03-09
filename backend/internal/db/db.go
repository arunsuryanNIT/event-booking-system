// Package db provides database connectivity and migration utilities for PostgreSQL.
package db

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq" // Postgres driver registration
)

// Connect opens a PostgreSQL connection pool and verifies connectivity with a ping.
// It configures pool limits suitable for moderate concurrency: 25 open, 5 idle, 5 min lifetime.
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
