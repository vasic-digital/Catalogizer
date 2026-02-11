package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	originalEnv map[string]string
	tmpDir      string
}

func (suite *ConfigTestSuite) SetupTest() {
	suite.originalEnv = make(map[string]string)
	envVars := []string{
		"CATALOG_CONFIG_PATH", "DB_PATH", "DB_ENCRYPTION_KEY", "PORT", "HOST",
		"GIN_MODE", "JWT_SECRET", "JWT_EXPIRY_HOURS", "SMB_SOURCES",
		"SMB_USERNAME", "SMB_PASSWORD", "TMDB_API_KEY", "LOG_LEVEL",
	}

	for _, envVar := range envVars {
		if value, exists := os.LookupEnv(envVar); exists {
			suite.originalEnv[envVar] = value
		}
		os.Unsetenv(envVar)
	}

	suite.tmpDir = suite.T().TempDir()
}

func (suite *ConfigTestSuite) TearDownTest() {
	for key, value := range suite.originalEnv {
		os.Setenv(key, value)
	}
	// Clean up any env vars we set during tests
	os.Unsetenv("CATALOG_CONFIG_PATH")
}

// writeTestConfig writes a JSON config file and returns the path
func (suite *ConfigTestSuite) writeTestConfig(cfg interface{}) string {
	data, err := json.MarshalIndent(cfg, "", "  ")
	suite.Require().NoError(err)

	path := filepath.Join(suite.tmpDir, "config.json")
	err = os.WriteFile(path, data, 0644)
	suite.Require().NoError(err)

	return path
}

func (suite *ConfigTestSuite) TestLoadDefaultConfig() {
	// Write a minimal config - validate() will fill in defaults
	path := suite.writeTestConfig(map[string]interface{}{})

	cfg, err := LoadFromFile(path)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg)
	assert.Equal(suite.T(), "8080", cfg.Server.Port)
	assert.Equal(suite.T(), "localhost", cfg.Server.Host)
	assert.Equal(suite.T(), 30, cfg.Server.ReadTimeout)
	assert.Equal(suite.T(), 30, cfg.Server.WriteTimeout)
	assert.Equal(suite.T(), 60, cfg.Server.IdleTimeout)
	assert.Equal(suite.T(), "/tmp", cfg.Catalog.TempDir)
	assert.Equal(suite.T(), int64(1024*1024*1024), cfg.Catalog.MaxArchiveSize)
	assert.Equal(suite.T(), 1024*1024, cfg.Catalog.DownloadChunkSize)
	assert.Equal(suite.T(), 30, cfg.SMB.Timeout)
	assert.Equal(suite.T(), 1024*1024, cfg.SMB.ChunkSize)
}

func (suite *ConfigTestSuite) TestLoadConfigWithEnvVars() {
	// Write a config file, then use CATALOG_CONFIG_PATH env var to load it via Load()
	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "0.0.0.0",
			"port": "9090",
		},
		"auth": map[string]interface{}{
			"enable_auth": true,
			"jwt_secret":  "env-test-secret",
		},
	}
	path := suite.writeTestConfig(configData)

	os.Setenv("CATALOG_CONFIG_PATH", path)

	cfg, err := Load()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg)
	assert.Equal(suite.T(), "0.0.0.0", cfg.Server.Host)
	assert.Equal(suite.T(), "9090", cfg.Server.Port)
	assert.True(suite.T(), cfg.Auth.EnableAuth)
	assert.Equal(suite.T(), "env-test-secret", cfg.Auth.JWTSecret)
}

func (suite *ConfigTestSuite) TestConfigValidation() {
	// Test that validate() sets defaults for empty/zero fields
	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "",
			"port": "",
		},
	}
	path := suite.writeTestConfig(configData)

	cfg, err := LoadFromFile(path)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg)
	// validate() should have set defaults
	assert.Equal(suite.T(), "localhost", cfg.Server.Host)
	assert.Equal(suite.T(), "8080", cfg.Server.Port)
}

func (suite *ConfigTestSuite) TestSMBSourceParsing() {
	configData := map[string]interface{}{
		"smb": map[string]interface{}{
			"hosts": []map[string]interface{}{
				{
					"name":     "nas01",
					"host":     "192.168.1.100",
					"port":     445,
					"share":    "media",
					"username": "user1",
					"password": "pass1",
					"domain":   "WORKGROUP",
				},
				{
					"name":  "nas02",
					"host":  "192.168.1.101",
					"port":  445,
					"share": "backup",
				},
			},
			"timeout":    60,
			"chunk_size": 2097152,
		},
	}
	path := suite.writeTestConfig(configData)

	cfg, err := LoadFromFile(path)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg.SMB)
	assert.Len(suite.T(), cfg.SMB.Hosts, 2)
	assert.Equal(suite.T(), "nas01", cfg.SMB.Hosts[0].Name)
	assert.Equal(suite.T(), "192.168.1.100", cfg.SMB.Hosts[0].Host)
	assert.Equal(suite.T(), 445, cfg.SMB.Hosts[0].Port)
	assert.Equal(suite.T(), "media", cfg.SMB.Hosts[0].Share)
	assert.Equal(suite.T(), 60, cfg.SMB.Timeout)
	assert.Equal(suite.T(), 2097152, cfg.SMB.ChunkSize)
}

func (suite *ConfigTestSuite) TestGetServerAddress() {
	tests := []struct {
		host     string
		port     string
		expected string
	}{
		{"127.0.0.1", "8080", "127.0.0.1:8080"},
		{"0.0.0.0", "9090", "0.0.0.0:9090"},
		{"localhost", "3000", "localhost:3000"},
	}

	for _, test := range tests {
		configData := map[string]interface{}{
			"server": map[string]interface{}{
				"host": test.host,
				"port": test.port,
			},
		}
		path := suite.writeTestConfig(configData)

		cfg, err := LoadFromFile(path)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), test.expected, cfg.GetServerAddress())
	}
}

func (suite *ConfigTestSuite) TestDatabaseConfig() {
	configData := map[string]interface{}{
		"database": map[string]interface{}{
			"driver":   "postgres",
			"host":     "db.example.com",
			"port":     5432,
			"database": "testdb",
			"username": "dbuser",
			"password": "dbpass",
			"ssl_mode": "require",
		},
	}
	path := suite.writeTestConfig(configData)

	cfg, err := LoadFromFile(path)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "postgres", cfg.Database.Driver)
	assert.Equal(suite.T(), "db.example.com", cfg.Database.Host)
	assert.Equal(suite.T(), 5432, cfg.Database.Port)
	assert.Equal(suite.T(), "testdb", cfg.Database.Database)
	assert.Equal(suite.T(), "dbuser", cfg.Database.Username)
	assert.Equal(suite.T(), "dbpass", cfg.Database.Password)
	assert.Equal(suite.T(), "require", cfg.Database.SSLMode)
}

func (suite *ConfigTestSuite) TestAuthConfig() {
	configData := map[string]interface{}{
		"auth": map[string]interface{}{
			"enable_auth": true,
			"jwt_secret":  "super-secret-jwt-key",
		},
	}
	path := suite.writeTestConfig(configData)

	cfg, err := LoadFromFile(path)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), cfg.Auth.EnableAuth)
	assert.Equal(suite.T(), "super-secret-jwt-key", cfg.Auth.JWTSecret)
}

func (suite *ConfigTestSuite) TestSMBConfig() {
	configData := map[string]interface{}{
		"smb": map[string]interface{}{
			"hosts":      []map[string]interface{}{},
			"timeout":    45,
			"chunk_size": 4194304,
		},
	}
	path := suite.writeTestConfig(configData)

	cfg, err := LoadFromFile(path)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg.SMB)
	assert.NotNil(suite.T(), cfg.SMB.Hosts)
	assert.Equal(suite.T(), 45, cfg.SMB.Timeout)
	assert.Equal(suite.T(), 4194304, cfg.SMB.ChunkSize)
}

func (suite *ConfigTestSuite) TestExternalAPIConfig() {
	// Test that a config with all sections loads correctly
	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "0.0.0.0",
			"port": "8080",
		},
		"database": map[string]interface{}{
			"driver":   "sqlite3",
			"database": "test.db",
		},
		"auth": map[string]interface{}{
			"enable_auth": false,
		},
		"logging": map[string]interface{}{
			"level":  "debug",
			"format": "text",
		},
		"catalog": map[string]interface{}{
			"temp_dir": "/tmp/test",
		},
	}
	path := suite.writeTestConfig(configData)

	cfg, err := LoadFromFile(path)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg)
	assert.Equal(suite.T(), "0.0.0.0", cfg.Server.Host)
	assert.Equal(suite.T(), "sqlite3", cfg.Database.Driver)
	assert.False(suite.T(), cfg.Auth.EnableAuth)
	assert.Equal(suite.T(), "debug", cfg.Logging.Level)
	assert.Equal(suite.T(), "/tmp/test", cfg.Catalog.TempDir)
}

func (suite *ConfigTestSuite) TestLoggingConfig() {
	configData := map[string]interface{}{
		"logging": map[string]interface{}{
			"level":  "error",
			"format": "json",
		},
	}
	path := suite.writeTestConfig(configData)

	cfg, err := LoadFromFile(path)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", cfg.Logging.Level)
	assert.Equal(suite.T(), "json", cfg.Logging.Format)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
