package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to write a JSON config file and return its path
func writeConfigFile(t *testing.T, dir string, cfg interface{}) string {
	t.Helper()
	path := filepath.Join(dir, "config.json")
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0644))
	return path
}

// helper to build a minimal valid config map
func minimalConfigMap() map[string]interface{} {
	return map[string]interface{}{
		"server":   map[string]interface{}{},
		"database": map[string]interface{}{},
		"smb":      map[string]interface{}{},
		"auth":     map[string]interface{}{},
		"logging":  map[string]interface{}{},
		"catalog":  map[string]interface{}{},
	}
}

// --- LoadFromFile tests ---

func TestLoadFromFile_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgMap := map[string]interface{}{
		"server": map[string]interface{}{
			"host":          "0.0.0.0",
			"port":          "9090",
			"read_timeout":  15,
			"write_timeout": 20,
			"idle_timeout":  45,
			"enable_cors":   true,
		},
		"database": map[string]interface{}{
			"driver":   "sqlite3",
			"database": "test.db",
		},
		"smb": map[string]interface{}{
			"timeout":    60,
			"chunk_size": 2048,
			"hosts": []map[string]interface{}{
				{"name": "nas1", "host": "192.168.1.10", "port": 445, "share": "media"},
			},
		},
		"auth": map[string]interface{}{
			"enable_auth": true,
			"jwt_secret":  "secret123",
		},
		"logging": map[string]interface{}{
			"level":  "debug",
			"format": "json",
		},
		"catalog": map[string]interface{}{
			"temp_dir":            "/var/tmp",
			"max_archive_size":    5368709120,
			"download_chunk_size": 4096,
		},
	}

	path := writeConfigFile(t, dir, cfgMap)
	cfg, err := LoadFromFile(path)

	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, 15, cfg.Server.ReadTimeout)
	assert.Equal(t, 20, cfg.Server.WriteTimeout)
	assert.Equal(t, 45, cfg.Server.IdleTimeout)
	assert.True(t, cfg.Server.EnableCORS)
	assert.Equal(t, "sqlite3", cfg.Database.Driver)
	assert.Equal(t, "test.db", cfg.Database.Database)
	assert.Equal(t, 60, cfg.SMB.Timeout)
	assert.Equal(t, 2048, cfg.SMB.ChunkSize)
	assert.Len(t, cfg.SMB.Hosts, 1)
	assert.Equal(t, "nas1", cfg.SMB.Hosts[0].Name)
	assert.True(t, cfg.Auth.EnableAuth)
	assert.Equal(t, "secret123", cfg.Auth.JWTSecret)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "/var/tmp", cfg.Catalog.TempDir)
	assert.Equal(t, int64(5368709120), cfg.Catalog.MaxArchiveSize)
	assert.Equal(t, 4096, cfg.Catalog.DownloadChunkSize)
}

func TestLoadFromFile_MissingFile(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open config file")
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte("{invalid json!!!"), 0644))

	_, err := LoadFromFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode config")
}

func TestLoadFromFile_EmptyJSON(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, map[string]interface{}{})

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	// validate() should fill all defaults
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "8080", cfg.Server.Port)
}

func TestLoadFromFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte(""), 0644))

	_, err := LoadFromFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode config")
}

// --- Load() tests (env var CATALOG_CONFIG_PATH) ---

func TestLoad_UsesEnvConfigPath(t *testing.T) {
	dir := t.TempDir()
	cfgMap := minimalConfigMap()
	cfgMap["server"] = map[string]interface{}{"host": "envhost", "port": "7777"}
	path := writeConfigFile(t, dir, cfgMap)

	original, hadOriginal := os.LookupEnv("CATALOG_CONFIG_PATH")
	os.Setenv("CATALOG_CONFIG_PATH", path)
	defer func() {
		if hadOriginal {
			os.Setenv("CATALOG_CONFIG_PATH", original)
		} else {
			os.Unsetenv("CATALOG_CONFIG_PATH")
		}
	}()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "envhost", cfg.Server.Host)
	assert.Equal(t, "7777", cfg.Server.Port)
}

func TestLoad_DefaultPathFallback(t *testing.T) {
	// When CATALOG_CONFIG_PATH is not set, Load() tries "config.json" in CWD.
	// That file likely doesn't exist in the test CWD, so we expect an error.
	original, hadOriginal := os.LookupEnv("CATALOG_CONFIG_PATH")
	os.Unsetenv("CATALOG_CONFIG_PATH")
	defer func() {
		if hadOriginal {
			os.Setenv("CATALOG_CONFIG_PATH", original)
		}
	}()

	_, err := Load()
	// Could succeed or fail depending on whether config.json exists in CWD.
	// We just verify it doesn't panic. If error, it should mention the config file.
	if err != nil {
		assert.Contains(t, err.Error(), "config file")
	}
}

func TestLoad_EnvPathMissingFile(t *testing.T) {
	original, hadOriginal := os.LookupEnv("CATALOG_CONFIG_PATH")
	os.Setenv("CATALOG_CONFIG_PATH", "/tmp/nonexistent_catalog_test_config.json")
	defer func() {
		if hadOriginal {
			os.Setenv("CATALOG_CONFIG_PATH", original)
		} else {
			os.Unsetenv("CATALOG_CONFIG_PATH")
		}
	}()

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open config file")
}

// --- validate() default assignment tests ---

func TestValidate_DefaultValues(t *testing.T) {
	tests := []struct {
		name    string
		input   Config
		checkFn func(t *testing.T, cfg *Config)
	}{
		{
			name:  "empty config gets all defaults",
			input: Config{},
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "localhost", cfg.Server.Host)
				assert.Equal(t, "8080", cfg.Server.Port)
				assert.Equal(t, 30, cfg.Server.ReadTimeout)
				assert.Equal(t, 30, cfg.Server.WriteTimeout)
				assert.Equal(t, 60, cfg.Server.IdleTimeout)
				assert.Equal(t, "/tmp", cfg.Catalog.TempDir)
				assert.Equal(t, int64(1024*1024*1024), cfg.Catalog.MaxArchiveSize)
				assert.Equal(t, 1024*1024, cfg.Catalog.DownloadChunkSize)
				assert.Equal(t, 30, cfg.SMB.Timeout)
				assert.Equal(t, 1024*1024, cfg.SMB.ChunkSize)
			},
		},
		{
			name: "provided host is preserved",
			input: Config{
				Server: ServerConfig{Host: "192.168.1.1"},
			},
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "192.168.1.1", cfg.Server.Host)
				assert.Equal(t, "8080", cfg.Server.Port) // default
			},
		},
		{
			name: "provided port is preserved",
			input: Config{
				Server: ServerConfig{Port: "3000"},
			},
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "localhost", cfg.Server.Host) // default
				assert.Equal(t, "3000", cfg.Server.Port)
			},
		},
		{
			name: "provided timeouts are preserved",
			input: Config{
				Server: ServerConfig{
					ReadTimeout:  10,
					WriteTimeout: 15,
					IdleTimeout:  120,
				},
			},
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 10, cfg.Server.ReadTimeout)
				assert.Equal(t, 15, cfg.Server.WriteTimeout)
				assert.Equal(t, 120, cfg.Server.IdleTimeout)
			},
		},
		{
			name: "provided catalog values are preserved",
			input: Config{
				Catalog: CatalogConfig{
					TempDir:           "/custom/tmp",
					MaxArchiveSize:    999,
					DownloadChunkSize: 512,
				},
			},
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "/custom/tmp", cfg.Catalog.TempDir)
				assert.Equal(t, int64(999), cfg.Catalog.MaxArchiveSize)
				assert.Equal(t, 512, cfg.Catalog.DownloadChunkSize)
			},
		},
		{
			name: "provided SMB values are preserved",
			input: Config{
				SMB: SMBConfig{
					Timeout:   120,
					ChunkSize: 4096,
				},
			},
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 120, cfg.SMB.Timeout)
				assert.Equal(t, 4096, cfg.SMB.ChunkSize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.input
			err := cfg.validate()
			require.NoError(t, err)
			tt.checkFn(t, &cfg)
		})
	}
}

func TestValidate_ReturnsNoError(t *testing.T) {
	// validate() currently always returns nil, but test the contract
	cfg := Config{}
	err := cfg.validate()
	assert.NoError(t, err)
}

// --- GetServerAddress tests ---

func TestGetServerAddress(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		expected string
	}{
		{
			name:     "localhost with default port",
			host:     "localhost",
			port:     "8080",
			expected: "localhost:8080",
		},
		{
			name:     "IP address with custom port",
			host:     "192.168.1.100",
			port:     "3000",
			expected: "192.168.1.100:3000",
		},
		{
			name:     "bind all interfaces",
			host:     "0.0.0.0",
			port:     "443",
			expected: "0.0.0.0:443",
		},
		{
			name:     "IPv6 address",
			host:     "::1",
			port:     "8080",
			expected: "::1:8080",
		},
		{
			name:     "empty host and port",
			host:     "",
			port:     "",
			expected: ":",
		},
		{
			name:     "hostname with high port",
			host:     "myserver.local",
			port:     "65535",
			expected: "myserver.local:65535",
		},
		{
			name:     "empty host with port",
			host:     "",
			port:     "8080",
			expected: ":8080",
		},
		{
			name:     "host with empty port",
			host:     "localhost",
			port:     "",
			expected: "localhost:",
		},
		{
			name:     "non-numeric port string",
			host:     "localhost",
			port:     "notaport",
			expected: "localhost:notaport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Host: tt.host,
					Port: tt.port,
				},
			}
			assert.Equal(t, tt.expected, cfg.GetServerAddress())
		})
	}
}

// --- JSON parsing edge cases ---

func TestLoadFromFile_PartialConfig(t *testing.T) {
	// Only server section provided; everything else gets defaults
	dir := t.TempDir()
	cfgMap := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "partialhost",
			"port": "1234",
		},
	}
	path := writeConfigFile(t, dir, cfgMap)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, "partialhost", cfg.Server.Host)
	assert.Equal(t, "1234", cfg.Server.Port)
	// Defaults applied
	assert.Equal(t, 30, cfg.Server.ReadTimeout)
	assert.Equal(t, "/tmp", cfg.Catalog.TempDir)
	assert.Equal(t, 30, cfg.SMB.Timeout)
}

func TestLoadFromFile_ExtraFieldsIgnored(t *testing.T) {
	dir := t.TempDir()
	cfgMap := map[string]interface{}{
		"server":        map[string]interface{}{"host": "testhost"},
		"unknown_field": "should be ignored",
		"another":       123,
	}
	path := writeConfigFile(t, dir, cfgMap)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, "testhost", cfg.Server.Host)
}

func TestLoadFromFile_NullValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"server": {"host": null, "port": null}}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	// null decodes to zero values, then validate fills defaults
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "8080", cfg.Server.Port)
}

func TestLoadFromFile_WrongTypes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	// port is a string field, giving it an int should cause a decode error
	content := `{"server": {"port": 8080}}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	_, err := LoadFromFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode config")
}

func TestLoadFromFile_SMBHostsArray(t *testing.T) {
	dir := t.TempDir()
	cfgMap := map[string]interface{}{
		"smb": map[string]interface{}{
			"hosts": []map[string]interface{}{
				{
					"name":     "server1",
					"host":     "10.0.0.1",
					"port":     445,
					"share":    "share1",
					"username": "user1",
					"password": "pass1",
					"domain":   "WORKGROUP",
				},
				{
					"name":     "server2",
					"host":     "10.0.0.2",
					"port":     139,
					"share":    "share2",
					"username": "user2",
					"password": "pass2",
					"domain":   "DOMAIN",
				},
			},
		},
	}
	path := writeConfigFile(t, dir, cfgMap)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	require.Len(t, cfg.SMB.Hosts, 2)
	assert.Equal(t, "server1", cfg.SMB.Hosts[0].Name)
	assert.Equal(t, "10.0.0.1", cfg.SMB.Hosts[0].Host)
	assert.Equal(t, 445, cfg.SMB.Hosts[0].Port)
	assert.Equal(t, "share1", cfg.SMB.Hosts[0].Share)
	assert.Equal(t, "user1", cfg.SMB.Hosts[0].Username)
	assert.Equal(t, "pass1", cfg.SMB.Hosts[0].Password)
	assert.Equal(t, "WORKGROUP", cfg.SMB.Hosts[0].Domain)
	assert.Equal(t, "server2", cfg.SMB.Hosts[1].Name)
	assert.Equal(t, "DOMAIN", cfg.SMB.Hosts[1].Domain)
}

func TestLoadFromFile_EmptySMBHosts(t *testing.T) {
	dir := t.TempDir()
	cfgMap := map[string]interface{}{
		"smb": map[string]interface{}{
			"hosts": []map[string]interface{}{},
		},
	}
	path := writeConfigFile(t, dir, cfgMap)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.Empty(t, cfg.SMB.Hosts)
}

func TestLoadFromFile_HTTPSConfig(t *testing.T) {
	dir := t.TempDir()
	cfgMap := map[string]interface{}{
		"server": map[string]interface{}{
			"enable_https": true,
			"cert_file":    "/etc/ssl/cert.pem",
			"key_file":     "/etc/ssl/key.pem",
		},
	}
	path := writeConfigFile(t, dir, cfgMap)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.True(t, cfg.Server.EnableHTTPS)
	assert.Equal(t, "/etc/ssl/cert.pem", cfg.Server.CertFile)
	assert.Equal(t, "/etc/ssl/key.pem", cfg.Server.KeyFile)
}

func TestLoadFromFile_DatabaseConfig(t *testing.T) {
	dir := t.TempDir()
	cfgMap := map[string]interface{}{
		"database": map[string]interface{}{
			"driver":   "postgres",
			"host":     "db.example.com",
			"port":     5432,
			"database": "catalogdb",
			"username": "admin",
			"password": "s3cret",
			"ssl_mode": "require",
		},
	}
	path := writeConfigFile(t, dir, cfgMap)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, "postgres", cfg.Database.Driver)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "catalogdb", cfg.Database.Database)
	assert.Equal(t, "admin", cfg.Database.Username)
	assert.Equal(t, "s3cret", cfg.Database.Password)
	assert.Equal(t, "require", cfg.Database.SSLMode)
}

// --- Validate interaction with LoadFromFile ---

func TestLoadFromFile_DefaultsAppliedAfterLoad(t *testing.T) {
	// Load a config with zero-value fields and verify all defaults are set
	dir := t.TempDir()
	cfgMap := map[string]interface{}{
		"server":  map[string]interface{}{},
		"catalog": map[string]interface{}{},
		"smb":     map[string]interface{}{},
	}
	path := writeConfigFile(t, dir, cfgMap)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)

	// Server defaults
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "8080", cfg.Server.Port)
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

func TestLoadFromFile_BooleanDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgMap := map[string]interface{}{}
	path := writeConfigFile(t, dir, cfgMap)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	// Booleans default to false (Go zero value)
	assert.False(t, cfg.Server.EnableCORS)
	assert.False(t, cfg.Server.EnableHTTPS)
	assert.False(t, cfg.Auth.EnableAuth)
}

func TestLoadFromFile_PermissionDenied(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0000))
	defer os.Chmod(path, 0644) // cleanup

	_, err := LoadFromFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open config file")
}
