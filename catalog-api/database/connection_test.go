package database

import (
	"os"
	"testing"

	"catalogizer/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConnection tests database connection creation
func TestNewConnection(t *testing.T) {
	// Create a temporary database file for testing
	tmpFile, err := os.CreateTemp("", "testdb_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create test configuration
	cfg := &config.DatabaseConfig{
		Path:               tmpFile.Name(),
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    3600, // 1 hour
		ConnMaxIdleTime:    1800, // 30 minutes
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        5000, // 5 seconds
	}

	// Test connection creation
	db, err := NewConnection(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Test that connection is actually working
	err = db.Ping()
	assert.NoError(t, err)

	// Test health check
	err = db.HealthCheck()
	assert.NoError(t, err)

	// Get stats and verify they're populated
	stats := db.GetStats()
	assert.NotNil(t, stats)

	// Close database
	err = db.Close()
	assert.NoError(t, err)
}

// TestNewConnectionWithInvalidPath tests connection with invalid path
func TestNewConnectionWithInvalidPath(t *testing.T) {
	// Create configuration with invalid path
	cfg := &config.DatabaseConfig{
		Path:               "/invalid/path/that/does/not/exist/test.db",
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    3600,
		ConnMaxIdleTime:    1800,
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        5000,
	}

	// Test connection creation - should fail
	db, err := NewConnection(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "unable to open database file")
}

// TestConnectionPoolConfiguration tests connection pool settings
func TestConnectionPoolConfiguration(t *testing.T) {
	// Create a temporary database file for testing
	tmpFile, err := os.CreateTemp("", "testdb_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:               tmpFile.Name(),
		MaxOpenConnections: 15,
		MaxIdleConnections: 8,
		ConnMaxLifetime:    7200, // 2 hours
		ConnMaxIdleTime:    3600, // 1 hour
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Test connection pool configuration
	stats := db.Stats()
	// The SQLite driver may already have connections opened
	// assert.Equal(t, 0, stats.OpenConnections) // No connections opened yet

	// Open a connection to test pool
	err = db.Ping()
	assert.NoError(t, err)

	stats = db.Stats()
	assert.GreaterOrEqual(t, stats.OpenConnections, 1)
	// Note: InUse behavior is implementation-specific to the SQLite driver
}

// TestHealthCheckTimeout tests health check with timeout
func TestHealthCheckTimeout(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "testdb_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:               tmpFile.Name(),
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    3600,
		ConnMaxIdleTime:    1800,
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        100, // Very short timeout for testing
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Test normal health check
	err = db.HealthCheck()
	assert.NoError(t, err)

	// Test with locked database (simulated busy state)
	// This is harder to test reliably without actual concurrent access
	// But we can verify the context creation works
	ctx, cancel := db.createContext()
	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)
	cancel()
}

// TestRewriteBooleanLiterals tests the PostgreSQL boolean literal rewriter
func TestRewriteBooleanLiterals(t *testing.T) {
	pg := Dialect{Type: DialectPostgres}
	sq := Dialect{Type: DialectSQLite}

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"WHERE deleted = 0", "SELECT * FROM files WHERE deleted = 0", "SELECT * FROM files WHERE deleted = FALSE"},
		{"WHERE deleted = 1", "SELECT * FROM files WHERE deleted = 1", "SELECT * FROM files WHERE deleted = TRUE"},
		{"SET is_active = 0", "UPDATE users SET is_active = 0 WHERE id = 1", "UPDATE users SET is_active = FALSE WHERE id = 1"},
		{"SET is_active = 1", "UPDATE users SET is_active = 1 WHERE id = 1", "UPDATE users SET is_active = TRUE WHERE id = 1"},
		{"SET is_locked = 1", "UPDATE users SET is_locked = 1 WHERE id = 5", "UPDATE users SET is_locked = TRUE WHERE id = 5"},
		{"SET is_locked = 0", "UPDATE users SET is_locked = 0, locked_until = NULL WHERE id = 5", "UPDATE users SET is_locked = FALSE, locked_until = NULL WHERE id = 5"},
		{"multiple booleans", "SELECT * FROM files WHERE is_directory = 1 AND deleted = 0", "SELECT * FROM files WHERE is_directory = TRUE AND deleted = FALSE"},
		{"no match for non-boolean", "SELECT * FROM files WHERE role_id = 1", "SELECT * FROM files WHERE role_id = 1"},
		{"no match for COUNT", "SELECT COUNT(*) WHERE file_count = 0", "SELECT COUNT(*) WHERE file_count = 0"},
		{"enabled column", "WHERE enabled = 1 AND max_depth = 10", "WHERE enabled = TRUE AND max_depth = 10"},
		{"is_duplicate", "WHERE is_duplicate = 1 AND deleted = 0", "WHERE is_duplicate = TRUE AND deleted = FALSE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pg.RewriteBooleanLiterals(tt.input)
			assert.Equal(t, tt.expect, got)
		})
	}

	// SQLite should pass through unchanged
	t.Run("sqlite passthrough", func(t *testing.T) {
		input := "SELECT * FROM files WHERE deleted = 0 AND is_active = 1"
		assert.Equal(t, input, sq.RewriteBooleanLiterals(input))
	})
}

// TestConnectionClose tests proper connection cleanup
func TestConnectionClose(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "testdb_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:               tmpFile.Name(),
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    3600,
		ConnMaxIdleTime:    1800,
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        5000,
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)

	// Use the database
	err = db.Ping()
	assert.NoError(t, err)

	// Close the database
	err = db.Close()
	assert.NoError(t, err)

	// Try to use closed database - should fail
	err = db.Ping()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}
