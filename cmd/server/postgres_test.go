package main

import (
	"os"
	"testing"
)

// TestOpenDB is an integration test: it talks to a real Postgres. It runs only
// when DATABASE_URL is set, so the default `go test ./...` stays fast and needs
// no database. t.Skip marks the test skipped (not failed) when that's the case.
func TestOpenDB(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres integration test")
	}

	db, err := openDB(dsn)
	if err != nil {
		t.Fatalf("openDB(%q): %v", dsn, err)
	}
	defer db.Close()
}
