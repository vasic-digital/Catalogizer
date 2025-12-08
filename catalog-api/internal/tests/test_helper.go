package tests

import (
	"database/sql"
	"testing"

	// Import SQLite driver once for all tests
	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run basic schema setup if needed
	// This can be expanded to include migrations if needed

	return db
}
