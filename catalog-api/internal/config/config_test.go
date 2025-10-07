package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	originalEnv map[string]string
}

func (suite *ConfigTestSuite) SetupTest() {
	// Save original environment
	suite.originalEnv = make(map[string]string)
	envVars := []string{
		"DB_PATH", "DB_ENCRYPTION_KEY", "PORT", "HOST", "GIN_MODE",
		"JWT_SECRET", "JWT_EXPIRY_HOURS", "SMB_SOURCES", "SMB_USERNAME",
		"SMB_PASSWORD", "TMDB_API_KEY", "LOG_LEVEL",
	}

	for _, envVar := range envVars {
		if value, exists := os.LookupEnv(envVar); exists {
			suite.originalEnv[envVar] = value
		}
		os.Unsetenv(envVar)
	}
}

func (suite *ConfigTestSuite) TearDownTest() {
	// Restore original environment
	for key, value := range suite.originalEnv {
		os.Setenv(key, value)
	}
}

func (suite *ConfigTestSuite) TestLoadDefaultConfig() {
	cfg, err := Load()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg)

	// Test default values
	assert.Equal(suite.T(), "./data/catalogizer.db", cfg.Database.Path)
	assert.Equal(suite.T(), "8080", cfg.Server.Port)
	assert.Equal(suite.T(), "127.0.0.1", cfg.Server.Host)
	assert.Equal(suite.T(), "info", cfg.Logging.Level)
}

func (suite *ConfigTestSuite) TestLoadConfigWithEnvVars() {
	// Set environment variables
	os.Setenv("DB_PATH", "/custom/path/db.sqlite")
	os.Setenv("DB_ENCRYPTION_KEY", "test-encryption-key-32-chars-long")
	os.Setenv("PORT", "9090")
	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("GIN_MODE", "release")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key")
	os.Setenv("JWT_EXPIRY_HOURS", "48")
	os.Setenv("SMB_SOURCES", "smb://server1/media,smb://server2/backup")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")
	os.Setenv("TMDB_API_KEY", "test-tmdb-key")
	os.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg)

	// Test environment variable values
	assert.Equal(suite.T(), "/custom/path/db.sqlite", cfg.Database.Path)
	assert.Equal(suite.T(), "test-encryption-key-32-chars-long", cfg.Database.EncryptionKey)
	assert.Equal(suite.T(), "9090", cfg.Server.Port)
	assert.Equal(suite.T(), "0.0.0.0", cfg.Server.Host)
	assert.Equal(suite.T(), "release", cfg.Server.GinMode)
	assert.Equal(suite.T(), "test-jwt-secret-key", cfg.Auth.JWTSecret)
	assert.Equal(suite.T(), 48, cfg.Auth.JWTExpiryHours)
	assert.Equal(suite.T(), []string{"smb://server1/media", "smb://server2/backup"}, cfg.SMB.Sources)
	assert.Equal(suite.T(), "testuser", cfg.SMB.Username)
	assert.Equal(suite.T(), "testpass", cfg.SMB.Password)
	assert.Equal(suite.T(), "test-tmdb-key", cfg.External.TMDBAPIKey)
	assert.Equal(suite.T(), "debug", cfg.Logging.Level)
}

func (suite *ConfigTestSuite) TestConfigValidation() {
	// Test missing required fields
	os.Setenv("JWT_SECRET", "") // Empty JWT secret

	cfg, err := Load()

	// Should still load but with warnings
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cfg)
}

func (suite *ConfigTestSuite) TestSMBSourceParsing() {
	os.Setenv("SMB_SOURCES", "smb://server1/media, smb://server2/videos ,smb://server3/music")

	cfg, err := Load()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{
		"smb://server1/media",
		"smb://server2/videos",
		"smb://server3/music",
	}, cfg.SMB.Sources)
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
		os.Setenv("HOST", test.host)
		os.Setenv("PORT", test.port)

		cfg, err := Load()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), test.expected, cfg.GetServerAddress())
	}
}

func (suite *ConfigTestSuite) TestDatabaseConfig() {
	os.Setenv("DB_PATH", "/data/catalogizer.db")
	os.Setenv("DB_ENCRYPTION_KEY", "32-character-encryption-key-here")

	cfg, err := Load()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "/data/catalogizer.db", cfg.Database.Path)
	assert.Equal(suite.T(), "32-character-encryption-key-here", cfg.Database.EncryptionKey)
}

func (suite *ConfigTestSuite) TestAuthConfig() {
	os.Setenv("JWT_SECRET", "super-secret-jwt-key")
	os.Setenv("JWT_EXPIRY_HOURS", "72")

	cfg, err := Load()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "super-secret-jwt-key", cfg.Auth.JWTSecret)
	assert.Equal(suite.T(), 72, cfg.Auth.JWTExpiryHours)
}

func (suite *ConfigTestSuite) TestSMBConfig() {
	os.Setenv("SMB_SOURCES", "smb://media-server/movies")
	os.Setenv("SMB_USERNAME", "catalogizer")
	os.Setenv("SMB_PASSWORD", "secure-password")
	os.Setenv("SMB_DOMAIN", "COMPANY")

	cfg, err := Load()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"smb://media-server/movies"}, cfg.SMB.Sources)
	assert.Equal(suite.T(), "catalogizer", cfg.SMB.Username)
	assert.Equal(suite.T(), "secure-password", cfg.SMB.Password)
	assert.Equal(suite.T(), "COMPANY", cfg.SMB.Domain)
}

func (suite *ConfigTestSuite) TestExternalAPIConfig() {
	os.Setenv("TMDB_API_KEY", "tmdb-api-key-123")
	os.Setenv("SPOTIFY_CLIENT_ID", "spotify-client-id")
	os.Setenv("SPOTIFY_CLIENT_SECRET", "spotify-client-secret")
	os.Setenv("STEAM_API_KEY", "steam-api-key")

	cfg, err := Load()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "tmdb-api-key-123", cfg.External.TMDBAPIKey)
	assert.Equal(suite.T(), "spotify-client-id", cfg.External.SpotifyClientID)
	assert.Equal(suite.T(), "spotify-client-secret", cfg.External.SpotifyClientSecret)
	assert.Equal(suite.T(), "steam-api-key", cfg.External.SteamAPIKey)
}

func (suite *ConfigTestSuite) TestLoggingConfig() {
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("LOG_FILE", "/var/log/catalogizer.log")

	cfg, err := Load()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", cfg.Logging.Level)
	assert.Equal(suite.T(), "/var/log/catalogizer.log", cfg.Logging.File)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}