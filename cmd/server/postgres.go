package main

import (
	"context"
	"database/sql"
	"errors"
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

// postgresStore is a Store backed by Postgres. It holds the connection pool;
// each method runs one SQL statement against it.
type postgresStore struct {
	db *sql.DB
}

func newPostgresStore(db *sql.DB) *postgresStore {
	return &postgresStore{db: db}
}

// save inserts a new code→url row. The count column defaults to 0 in the schema,
// so we don't set it here — mirroring the struct's zero value.
func (s *postgresStore) save(url string) (string, error) {
	code := generateCode()
	// $1, $2 are bound to code, url safely by the driver — never interpolate.
	_, err := s.db.Exec(`INSERT INTO urls (code, url) VALUES ($1, $2)`, code, url)
	if err != nil {
		return "", fmt.Errorf("save: %w", err)
	}
	return code, nil
}

// get looks up a code and increments its count in one statement. It returns the URL if found, or ok=false if not found. The count is incremented by the SQL, so
// this method is safe to call concurrently for the same code.
func (s *postgresStore) get(code string) (string, bool, error) {
	var url string
	// UPDATE ... RETURNING lets us do both steps in one statement safely.
	err := s.db.QueryRow(`UPDATE urls SET count = count + 1 WHERE code = $1 RETURNING url`, code).Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil // not found is not an error
	}
	if err != nil {
		return "", false, fmt.Errorf("get: %w", err)
	}
	return url, true, nil
}

// stats looks up a code and returns its count without incrementing it. It returns
// ok=false if the code is not found.
func (s *postgresStore) stats(code string) (int, bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT count FROM urls WHERE code = $1`, code).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil // not found is not an error
	}
	if err != nil {
		return 0, false, fmt.Errorf("stats: %w", err)
	}
	return count, true, nil
}

// Compile-time check that *postgresStore satisfies Store — same guarantee the
// in-memory *store has, so both stay swappable behind the interface.
var _ Store = (*postgresStore)(nil)
