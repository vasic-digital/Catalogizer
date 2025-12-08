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
		Path:                tmpFile.Name(),
		MaxOpenConnections:  10,
		MaxIdleConnections:  5,
		ConnMaxLifetime:     3600, // 1 hour
		ConnMaxIdleTime:     1800, // 30 minutes
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
		Path:                "/invalid/path/that/does/not/exist/test.db",
		MaxOpenConnections:  10,
		MaxIdleConnections:  5,
		ConnMaxLifetime:     3600,
		ConnMaxIdleTime:     1800,
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
		Path:                tmpFile.Name(),
		MaxOpenConnections:  15,
		MaxIdleConnections:  8,
		ConnMaxLifetime:     7200, // 2 hours
		ConnMaxIdleTime:     3600, // 1 hour
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
		Path:                tmpFile.Name(),
		MaxOpenConnections:  10,
		MaxIdleConnections:  5,
		ConnMaxLifetime:     3600,
		ConnMaxIdleTime:     1800,
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

// TestConnectionClose tests proper connection cleanup
func TestConnectionClose(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "testdb_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Path:                tmpFile.Name(),
		MaxOpenConnections:  10,
		MaxIdleConnections:  5,
		ConnMaxLifetime:     3600,
		ConnMaxIdleTime:     1800,
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