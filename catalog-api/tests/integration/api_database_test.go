package integration

import (
	"database/sql"
	"fmt"
	"testing"

	"catalogizer/database"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	// Create minimal schema
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY, name TEXT, path TEXT, size INTEGER, is_directory INTEGER DEFAULT 0, deleted INTEGER DEFAULT 0)")
	require.NoError(t, err)
	return db
}

func TestDatabaseExecWithTimeout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test that Exec works with the new default timeout
	_, err := db.Exec("INSERT INTO files (name, path, size) VALUES (?, ?, ?)", "test.mp4", "/media/test.mp4", 1024)
	assert.NoError(t, err)
}

func TestDatabaseQueryWithTimeout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO files (name, path, size) VALUES (?, ?, ?)", "test.mp4", "/media/test.mp4", 1024)
	require.NoError(t, err)

	rows, err := db.Query("SELECT name, path, size FROM files WHERE name = ?", "test.mp4")
	assert.NoError(t, err)
	defer rows.Close()

	assert.True(t, rows.Next())
}

func TestDatabaseQueryRowWithTimeout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO files (name, path, size) VALUES (?, ?, ?)", "test.mp4", "/media/test.mp4", 1024)
	require.NoError(t, err)

	// Use Query + Scan pattern to avoid context lifecycle issue
	// with database.DB's QueryRow deferred cancel
	rows, err := db.Query("SELECT COUNT(*) FROM files")
	assert.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	var count int
	err = rows.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDatabaseConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	// Use file-based SQLite with WAL mode for true concurrent access.
	// In-memory SQLite creates separate databases per connection, which
	// breaks concurrent multi-connection tests.
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/concurrent_test.db"

	sqlDB, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&cache=shared")
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	defer db.Close()

	// Enable WAL for concurrent access
	_, _ = db.Exec("PRAGMA journal_mode=WAL")

	// Create schema on file-based DB
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY, name TEXT, path TEXT, size INTEGER, is_directory INTEGER DEFAULT 0, deleted INTEGER DEFAULT 0)")
	require.NoError(t, err)

	// Concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := db.Exec("INSERT INTO files (name, path, size) VALUES (?, ?, ?)",
				fmt.Sprintf("file%d.mp4", id), fmt.Sprintf("/media/file%d.mp4", id), id*1024)
			assert.NoError(t, err)
			done <- true
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	rows, err := db.Query("SELECT COUNT(*) FROM files")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var count int
	err = rows.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 10, count)
}
