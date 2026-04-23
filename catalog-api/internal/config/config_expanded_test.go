package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Load() with CATALOG_CONFIG_PATH env var
// ---------------------------------------------------------------------------

func TestLoad_UsesEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "10.0.0.1",
			"port": "7777",
		},
	}
	data, err := json.MarshalIndent(configData, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	origVal, origExists := os.LookupEnv("CATALOG_CONFIG_PATH")
	os.Setenv("CATALOG_CONFIG_PATH", configPath)
	defer func() {
		if origExists {
			os.Setenv("CATALOG_CONFIG_PATH", origVal)
		} else {
			os.Unsetenv("CATALOG_CONFIG_PATH")
		}
	}()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", cfg.Server.Host)
	assert.Equal(t, "7777", cfg.Server.Port)
}

func TestLoad_DefaultsToConfigJSON(t *testing.T) {
	origVal, origExists := os.LookupEnv("CATALOG_CONFIG_PATH")
	os.Unsetenv("CATALOG_CONFIG_PATH")
	defer func() {
		if origExists {
			os.Setenv("CATALOG_CONFIG_PATH", origVal)
		}
	}()

	// This will try to load "config.json" from the current directory.
	// It may or may not exist, but the function should not panic.
	_, _ = Load()
}

// ---------------------------------------------------------------------------
// LoadFromFile — error handling
// ---------------------------------------------------------------------------

func TestLoadFromFile_NonexistentFile(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load config")
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

	_, err := LoadFromFile(path)
	assert.Error(t, err)
}

func TestLoadFromFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0644))

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Defaults should be populated by validate()
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "28080", cfg.Server.Port)
}

// ---------------------------------------------------------------------------
// validate — default filling
// ---------------------------------------------------------------------------

func TestValidate_FillsAllDefaults(t *testing.T) {
	cfg := &Config{}
	err := cfg.validate()
	require.NoError(t, err)

	// Server defaults
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "28080", cfg.Server.Port)
	assert.Equal(t, 30, cfg.Server.ReadTimeout)
	assert.Equal(t, 30, cfg.Server.WriteTimeout)
	assert.Equal(t, 60, cfg.Server.IdleTimeout)

	// Catalog defaults
	assert.Equal(t, "/tmp", cfg.Catalog.TempDir)
	assert.Equal(t, int64(1024*1024*1024), cfg.Catalog.MaxArchiveSize)
	assert.Equal(t, 1024*1024, cfg.Catalog.DownloadChunkSize)

	// SMB defaults
	assert.Equal(t, 30, cfg.SMB.Timeout)
	assert.Equal(t, 1024*1024, cfg.SMB.ChunkSize)
}

func TestValidate_PreservesExistingValues(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         "9090",
			ReadTimeout:  60,
			WriteTimeout: 120,
			IdleTimeout:  180,
		},
		Catalog: CatalogConfig{
			TempDir:           "/custom/tmp",
			MaxArchiveSize:    2048,
			DownloadChunkSize: 512,
		},
		SMB: SMBConfig{
			Timeout:   45,
			ChunkSize: 2048,
		},
	}

	err := cfg.validate()
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, 60, cfg.Server.ReadTimeout)
	assert.Equal(t, 120, cfg.Server.WriteTimeout)
	assert.Equal(t, 180, cfg.Server.IdleTimeout)
	assert.Equal(t, "/custom/tmp", cfg.Catalog.TempDir)
	assert.Equal(t, int64(2048), cfg.Catalog.MaxArchiveSize)
	assert.Equal(t, 512, cfg.Catalog.DownloadChunkSize)
	assert.Equal(t, 45, cfg.SMB.Timeout)
	assert.Equal(t, 2048, cfg.SMB.ChunkSize)
}

// ---------------------------------------------------------------------------
// GetServerAddress — edge cases
// ---------------------------------------------------------------------------

func TestGetServerAddress_IPv6(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "::1",
			Port: "28080",
		},
	}
	assert.Equal(t, "::1:28080", cfg.GetServerAddress())
}

func TestGetServerAddress_EmptyHostPort(t *testing.T) {
	cfg := &Config{}
	_ = cfg.validate()
	assert.Equal(t, "localhost:28080", cfg.GetServerAddress())
}

// ---------------------------------------------------------------------------
// Full config round-trip
// ---------------------------------------------------------------------------

func TestFullConfigRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	original := map[string]interface{}{
		"server": map[string]interface{}{
			"host":          "api.example.com",
			"port":          "443",
			"read_timeout":  10,
			"write_timeout": 20,
			"idle_timeout":  30,
			"enable_cors":   true,
			"enable_https":  true,
			"cert_file":     "/etc/ssl/cert.pem",
			"key_file":      "/etc/ssl/key.pem",
		},
		"database": map[string]interface{}{
			"driver":   "postgres",
			"host":     "db.example.com",
			"port":     5432,
			"database": "catalogizer",
			"username": "admin",
			"password": "secret",
			"ssl_mode": "require",
		},
		"smb": map[string]interface{}{
			"hosts": []map[string]interface{}{
				{
					"name":     "nas01",
					"host":     "192.168.1.100",
					"port":     445,
					"share":    "media",
					"username": "user",
					"password": "pass",
					"domain":   "CORP",
				},
			},
			"timeout":    60,
			"chunk_size": 4194304,
		},
		"auth": map[string]interface{}{
			"enable_auth": true,
			"jwt_secret":  "super-secret-key",
		},
		"logging": map[string]interface{}{
			"level":  "debug",
			"format": "json",
		},
		"catalog": map[string]interface{}{
			"temp_dir":            "/var/tmp",
			"max_archive_size":    2147483648,
			"download_chunk_size": 2097152,
		},
	}

	data, err := json.MarshalIndent(original, "", "  ")
	require.NoError(t, err)
	path := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(path, data, 0644))

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)

	// Server
	assert.Equal(t, "api.example.com", cfg.Server.Host)
	assert.Equal(t, "443", cfg.Server.Port)
	assert.Equal(t, 10, cfg.Server.ReadTimeout)
	assert.Equal(t, 20, cfg.Server.WriteTimeout)
	assert.Equal(t, 30, cfg.Server.IdleTimeout)
	assert.True(t, cfg.Server.EnableCORS)
	assert.True(t, cfg.Server.EnableHTTPS)
	assert.Equal(t, "/etc/ssl/cert.pem", cfg.Server.CertFile)
	assert.Equal(t, "/etc/ssl/key.pem", cfg.Server.KeyFile)

	// Database
	assert.Equal(t, "postgres", cfg.Database.Driver)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "catalogizer", cfg.Database.Database)
	assert.Equal(t, "require", cfg.Database.SSLMode)

	// SMB
	assert.Len(t, cfg.SMB.Hosts, 1)
	assert.Equal(t, "CORP", cfg.SMB.Hosts[0].Domain)
	assert.Equal(t, 60, cfg.SMB.Timeout)
	assert.Equal(t, 4194304, cfg.SMB.ChunkSize)

	// Auth
	assert.True(t, cfg.Auth.EnableAuth)
	assert.Equal(t, "super-secret-key", cfg.Auth.JWTSecret)

	// Logging
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)

	// Catalog
	assert.Equal(t, "/var/tmp", cfg.Catalog.TempDir)
	assert.Equal(t, int64(2147483648), cfg.Catalog.MaxArchiveSize)
	assert.Equal(t, 2097152, cfg.Catalog.DownloadChunkSize)

	// GetServerAddress
	assert.Equal(t, "api.example.com:443", cfg.GetServerAddress())
}

// ---------------------------------------------------------------------------
// Config struct JSON serialization
// ---------------------------------------------------------------------------

func TestConfig_JSONMarshalUnmarshal(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: "28080",
		},
		Auth: AuthConfig{
			EnableAuth: true,
			JWTSecret:  "test-secret",
		},
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var restored Config
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, cfg.Server.Host, restored.Server.Host)
	assert.Equal(t, cfg.Auth.JWTSecret, restored.Auth.JWTSecret)
}
