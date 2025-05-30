package main

import (
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// Connects to a mock memory database
func newMockDatabase(t *testing.T) *sql.DB {
	// Creating a mock memory database
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatal("failed to create mock memory database", err)
	}
	// Applying migrations
	err = goose.Up(db, "migrations")
	if err != nil {
		t.Fatal("failed to apply mock memory database migrations", err)
	}
	return db
}

// Tests API endpoints
func TestHandlers(t *testing.T) {
	db := newMockDatabase(t)
	defer db.Close()
	// store := store.NewStore(db)
	// handlers := NewHandlers(store)
	t.Run("squanchy /squanchy", func(t *testing.T) {
		// To be implemented
	})
}
