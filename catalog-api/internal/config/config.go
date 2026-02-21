package config

import (
	"fmt"
	"os"

	digitalconfig "digital.vasic.config/pkg/config"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	SMB      SMBConfig      `json:"smb"`
	Auth     AuthConfig     `json:"auth"`
	Logging  LoggingConfig  `json:"logging"`
	Catalog  CatalogConfig  `json:"catalog"`
}

type ServerConfig struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
	EnableCORS   bool   `json:"enable_cors"`
	EnableHTTPS  bool   `json:"enable_https"`
	CertFile     string `json:"cert_file"`
	KeyFile      string `json:"key_file"`
}

type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

type SMBConfig struct {
	Hosts     []SMBHost `json:"hosts"`
	Timeout   int       `json:"timeout"`
	ChunkSize int       `json:"chunk_size"`
}

type SMBHost struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Share    string `json:"share"`
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
}

type AuthConfig struct {
	EnableAuth bool   `json:"enable_auth"`
	JWTSecret  string `json:"jwt_secret"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type CatalogConfig struct {
	TempDir           string `json:"temp_dir"`
	MaxArchiveSize    int64  `json:"max_archive_size"`
	DownloadChunkSize int    `json:"download_chunk_size"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CATALOG_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	return LoadFromFile(configPath)
}

func LoadFromFile(path string) (*Config, error) {
	var config Config
	if err := digitalconfig.LoadFile(path, &config); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func (c *Config) validate() error {
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if c.Server.Host == "" {
		c.Server.Host = "localhost"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 60
	}

	if c.Catalog.TempDir == "" {
		c.Catalog.TempDir = "/tmp"
	}
	if c.Catalog.MaxArchiveSize == 0 {
		c.Catalog.MaxArchiveSize = 1024 * 1024 * 1024 // 1GB
	}
	if c.Catalog.DownloadChunkSize == 0 {
		c.Catalog.DownloadChunkSize = 1024 * 1024 // 1MB
	}

	if c.SMB.Timeout == 0 {
		c.SMB.Timeout = 30
	}
	if c.SMB.ChunkSize == 0 {
		c.SMB.ChunkSize = 1024 * 1024 // 1MB
	}

	return nil
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}
