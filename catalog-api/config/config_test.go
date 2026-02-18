package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultConfig(t *testing.T) {
	config := getDefaultConfig()

	assert.NotNil(t, config)

	// Server defaults
	assert.Equal(t, "localhost", config.Server.Host)
	assert.Equal(t, 8080, config.Server.Port)
	assert.True(t, config.Server.EnableCORS)
	assert.True(t, config.Server.EnableHTTPS)

	// Database defaults
	assert.Equal(t, "./catalog.db", config.Database.Path)
	assert.Equal(t, 25, config.Database.MaxOpenConnections)
	assert.True(t, config.Database.EnableWAL)

	// Auth defaults
	assert.Equal(t, 24, config.Auth.JWTExpirationHours)
	assert.True(t, config.Auth.EnableAuth)

	// Catalog defaults
	assert.Equal(t, 100, config.Catalog.DefaultPageSize)
	assert.Equal(t, 1000, config.Catalog.MaxPageSize)
	assert.True(t, config.Catalog.EnableCache)

	// Logging defaults
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	// Set required env vars
	os.Setenv("JWT_SECRET", "this-is-a-super-long-secret-key-for-testing")
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "password123")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ADMIN_USERNAME")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	config := getDefaultConfig()
	err := validateConfig(config)
	assert.NoError(t, err)
}

func TestValidateConfig_InvalidPort(t *testing.T) {
	config := getDefaultConfig()
	config.Auth.EnableAuth = false // Disable auth to isolate port validation

	tests := []struct {
		name string
		port int
	}{
		{"negative port", -1},
		{"zero port", 0},
		{"port too high", 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Server.Port = tt.port
			err := validateConfig(config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid server port")
		})
	}
}

func TestValidateConfig_EmptyDatabasePath(t *testing.T) {
	config := getDefaultConfig()
	config.Auth.EnableAuth = false
	config.Database.Type = "sqlite"
	config.Database.Path = ""

	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database path cannot be empty for sqlite")
}

func TestValidateConfig_AuthValidation(t *testing.T) {
	config := getDefaultConfig()
	config.Auth.EnableAuth = true

	// Clear env vars
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("ADMIN_USERNAME")
	os.Unsetenv("ADMIN_PASSWORD")

	// No JWT secret
	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret must be set")

	// JWT secret too short
	config.Auth.JWTSecret = "short"
	err = validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 32 characters")

	// Valid JWT but no admin credentials
	config.Auth.JWTSecret = "this-is-a-super-long-secret-key-for-testing"
	err = validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "admin credentials must be set")
}

func TestValidateConfig_PageSizeValidation(t *testing.T) {
	os.Setenv("JWT_SECRET", "this-is-a-super-long-secret-key-for-testing")
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "password123")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ADMIN_USERNAME")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	config := getDefaultConfig()

	// Invalid default page size
	config.Catalog.DefaultPageSize = 0
	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default page size must be positive")

	// Max page size less than default
	config.Catalog.DefaultPageSize = 100
	config.Catalog.MaxPageSize = 50
	err = validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max page size must be >= default page size")
}

func TestGetDatabaseURL_Postgres(t *testing.T) {
	config := getDefaultConfig()
	// Default is postgres
	url := config.GetDatabaseURL()
	assert.Contains(t, url, "postgres://")
	assert.Contains(t, url, "catalogizer")
	assert.Contains(t, url, "sslmode=disable")
}

func TestGetDatabaseURL_SQLite(t *testing.T) {
	config := getDefaultConfig()
	config.Database.Type = "sqlite"
	config.Database.Path = "./catalog.db"
	config.Database.BusyTimeout = 5000
	config.Database.EnableWAL = true

	url := config.GetDatabaseURL()
	assert.Contains(t, url, "./catalog.db")
	assert.Contains(t, url, "_busy_timeout=5000")
	assert.Contains(t, url, "_journal_mode=WAL")
	assert.Contains(t, url, "_foreign_keys=1")
	assert.Contains(t, url, "_wal_autocheckpoint=1000")
}

func TestGetDatabaseURL_WithCustomCacheSize(t *testing.T) {
	config := getDefaultConfig()
	config.Database.Type = "sqlite"
	config.Database.Path = "./catalog.db"
	config.Database.CacheSize = -4000

	url := config.GetDatabaseURL()
	assert.Contains(t, url, "_cache_size=-4000")
}

func TestGetDatabaseURL_DisabledWAL(t *testing.T) {
	config := getDefaultConfig()
	config.Database.Type = "sqlite"
	config.Database.Path = "./catalog.db"
	config.Database.EnableWAL = false

	url := config.GetDatabaseURL()
	assert.NotContains(t, url, "_wal_autocheckpoint")
}

func TestGetServerAddress(t *testing.T) {
	config := getDefaultConfig()

	addr := config.GetServerAddress()
	assert.Equal(t, "localhost:8080", addr)

	config.Server.Host = "0.0.0.0"
	config.Server.Port = 9000
	addr = config.GetServerAddress()
	assert.Equal(t, "0.0.0.0:9000", addr)
}

func TestLoadConfig_CreateDefault(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	// Disable auth for simpler testing
	os.Setenv("JWT_SECRET", "this-is-a-super-long-secret-key-for-testing")
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "password123")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ADMIN_USERNAME")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	// Load config (should create default)
	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// File should now exist
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("not valid json"), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.json")

	config := getDefaultConfig()
	err := saveConfig(config, configPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Verify permissions (should be 0600)
	info, _ := os.Stat(configPath)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}
