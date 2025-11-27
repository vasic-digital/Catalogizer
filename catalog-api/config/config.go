package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the API configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	Catalog  CatalogConfig  `json:"catalog"`
	Storage  StorageConfig  `json:"storage"`
	Logging  LoggingConfig  `json:"logging"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
	EnableCORS   bool   `json:"enable_cors"`
	EnableHTTPS  bool   `json:"enable_https"`
	CertFile     string `json:"cert_file,omitempty"`
	KeyFile      string `json:"key_file,omitempty"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Path               string `json:"path"`
	MaxOpenConnections int    `json:"max_open_connections"`
	MaxIdleConnections int    `json:"max_idle_connections"`
	ConnMaxLifetime    int    `json:"conn_max_lifetime"`
	ConnMaxIdleTime    int    `json:"conn_max_idle_time"`
	EnableWAL          bool   `json:"enable_wal"`
	CacheSize          int    `json:"cache_size"`
	BusyTimeout        int    `json:"busy_timeout"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	JWTSecret          string `json:"jwt_secret"`
	JWTExpirationHours int    `json:"jwt_expiration_hours"`
	EnableAuth         bool   `json:"enable_auth"`
	AdminUsername      string `json:"admin_username"`
	AdminPassword      string `json:"admin_password"`
}

// CatalogConfig contains catalog-specific configuration
type CatalogConfig struct {
	DefaultPageSize      int      `json:"default_page_size"`
	MaxPageSize          int      `json:"max_page_size"`
	EnableCache          bool     `json:"enable_cache"`
	CacheTTLMinutes      int      `json:"cache_ttl_minutes"`
	MaxConcurrentScans   int      `json:"max_concurrent_scans"`
	DownloadChunkSize    int      `json:"download_chunk_size"`
	MaxArchiveSize       int64    `json:"max_archive_size"`
	AllowedDownloadTypes []string `json:"allowed_download_types"`
	TempDir              string   `json:"temp_dir"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

// StorageConfig contains storage configuration for multiple protocols
type StorageConfig struct {
	Roots []StorageRootConfig `json:"roots"`
}

// StorageRootConfig represents configuration for a single storage root
type StorageRootConfig struct {
	ID                       string                 `json:"id"`
	Name                     string                 `json:"name"`
	Protocol                 string                 `json:"protocol"` // smb, ftp, nfs, webdav, local
	Enabled                  bool                   `json:"enabled"`
	MaxDepth                 int                    `json:"max_depth"`
	EnableDuplicateDetection bool                   `json:"enable_duplicate_detection"`
	EnableMetadataExtraction bool                   `json:"enable_metadata_extraction"`
	IncludePatterns          []string               `json:"include_patterns,omitempty"`
	ExcludePatterns          []string               `json:"exclude_patterns,omitempty"`
	Settings                 map[string]interface{} `json:"settings"` // Protocol-specific settings
}

// LoadConfig loads configuration from file or creates default
func LoadConfig(configPath string) (*Config, error) {
	config := getDefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		if err := saveConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// getDefaultConfig returns default configuration
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  120,
			EnableCORS:   true,
			EnableHTTPS:  true, // Enable HTTPS by default for security
		},
		Database: DatabaseConfig{
			Path:               "./catalog.db",
			MaxOpenConnections: 25,
			MaxIdleConnections: 5,
			ConnMaxLifetime:    300,
			ConnMaxIdleTime:    60,
			EnableWAL:          true,
			CacheSize:          -2000,
			BusyTimeout:        5000,
		},
		Auth: AuthConfig{
			JWTSecret:          "", // Must be set via environment variable
			JWTExpirationHours: 24,
			EnableAuth:         true, // Enable auth by default for security
			AdminUsername:      "", // Must be set via environment variable
			AdminPassword:      "", // Must be set via environment variable
		},
		Catalog: CatalogConfig{
			DefaultPageSize:      100,
			MaxPageSize:          1000,
			EnableCache:          true,
			CacheTTLMinutes:      15,
			MaxConcurrentScans:   3,
			DownloadChunkSize:    1024 * 1024,            // 1MB
			MaxArchiveSize:       1024 * 1024 * 1024 * 5, // 5GB
			AllowedDownloadTypes: []string{"*"},
			TempDir:              os.TempDir() + "/catalog-api", // Use system temp directory
		},
		Storage: StorageConfig{
			Roots: []StorageRootConfig{
				{
					ID:                       "local-example",
					Name:                     "Local Files",
					Protocol:                 "local",
					Enabled:                  true,
					MaxDepth:                 10,
					EnableDuplicateDetection: true,
					EnableMetadataExtraction: true,
					Settings: map[string]interface{}{
						"base_path": "/tmp/catalog-data",
					},
				},
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	if config.Auth.EnableAuth {
		// Check for environment variables first
		if config.Auth.JWTSecret == "" {
			config.Auth.JWTSecret = os.Getenv("JWT_SECRET")
		}
		if config.Auth.AdminUsername == "" {
			config.Auth.AdminUsername = os.Getenv("ADMIN_USERNAME")
		}
		if config.Auth.AdminPassword == "" {
			config.Auth.AdminPassword = os.Getenv("ADMIN_PASSWORD")
		}
		
		// Validate required security settings
		if config.Auth.JWTSecret == "" {
			return fmt.Errorf("JWT secret must be set via JWT_SECRET environment variable or config")
		}
		if len(config.Auth.JWTSecret) < 32 {
			return fmt.Errorf("JWT secret must be at least 32 characters long")
		}
		if config.Auth.AdminUsername == "" || config.Auth.AdminPassword == "" {
			return fmt.Errorf("admin credentials must be set via ADMIN_USERNAME and ADMIN_PASSWORD environment variables")
		}
	}

	if config.Catalog.DefaultPageSize <= 0 {
		return fmt.Errorf("default page size must be positive")
	}

	if config.Catalog.MaxPageSize < config.Catalog.DefaultPageSize {
		return fmt.Errorf("max page size must be >= default page size")
	}

	return nil
}

// saveConfig saves configuration to file
func saveConfig(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDatabaseURL returns the database connection URL
func (c *Config) GetDatabaseURL() string {
	params := "?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1"
	if c.Database.EnableWAL {
		params += "&_wal_autocheckpoint=1000"
	}
	if c.Database.CacheSize != 0 {
		params += fmt.Sprintf("&_cache_size=%d", c.Database.CacheSize)
	}
	return c.Database.Path + params
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
