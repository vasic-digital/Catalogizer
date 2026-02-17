package challenges

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// EndpointConfig holds all configured challenge endpoints.
type EndpointConfig struct {
	Endpoints []Endpoint `json:"endpoints"`
}

// Endpoint describes a single SMB endpoint for challenge validation.
type Endpoint struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Host        string      `json:"host"`
	Port        int         `json:"port"`
	Share       string      `json:"share"`
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	Domain      string      `json:"domain"`
	ReadOnly    bool        `json:"readonly"`
	Directories []Directory `json:"directories"`
}

// Directory describes a single directory within an endpoint.
type Directory struct {
	Path        string `json:"path"`
	ContentType string `json:"content_type"`
}

// LoadEndpointConfig reads and parses the endpoint configuration
// from the given JSON file path.
func LoadEndpointConfig(path string) (*EndpointConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read endpoint config %s: %w", path, err)
	}

	var cfg EndpointConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse endpoint config %s: %w", path, err)
	}

	return &cfg, nil
}

// DefaultConfigPath returns the default path to the endpoint
// configuration file. Resolution order:
//  1. CHALLENGE_CONFIG_PATH environment variable (for containers)
//  2. Relative to source file (works in dev / go run)
//  3. Relative to working directory (fallback)
func DefaultConfigPath() string {
	if envPath := os.Getenv("CHALLENGE_CONFIG_PATH"); envPath != "" {
		return envPath
	}
	// Try to resolve relative to the source file location first,
	// which works when running from any working directory.
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		p := filepath.Join(filepath.Dir(filename), "config", "endpoints.json")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join("challenges", "config", "endpoints.json")
}
