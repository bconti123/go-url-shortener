package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver with database/sql
)

// openDB opens a connection pool to Postgres and verifies it is reachable.
// It returns a *sql.DB — a concurrency-safe pool meant to live for the whole
// program and be shared, not opened per request.
func openDB(dsn string) (*sql.DB, error) {
	// sql.Open does NOT connect; it only validates the driver name and DSN
	// and prepares the pool. The first real connection is lazy.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Ping forces a real connection now, so a misconfigured or down database
	// fails at startup instead of on the first user request. The context caps
	// how long we wait before giving up.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}
