package challenges

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEndpointConfig(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	path := filepath.Join(dir, "endpoints.json")

	content := `{
		"endpoints": [
			{
				"id": "test-endpoint",
				"name": "Test NAS",
				"host": "nas.local",
				"port": 445,
				"share": "data",
				"username": "user",
				"password": "pass",
				"domain": "WORKGROUP",
				"readonly": true,
				"directories": [
					{"path": "Music", "content_type": "music"},
					{"path": "Movies", "content_type": "movie"}
				]
			}
		]
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadEndpointConfig(path)
	if err != nil {
		t.Fatalf("LoadEndpointConfig returned error: %v", err)
	}

	if len(cfg.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(cfg.Endpoints))
	}

	ep := cfg.Endpoints[0]
	if ep.ID != "test-endpoint" {
		t.Errorf("expected ID 'test-endpoint', got '%s'", ep.ID)
	}
	if ep.Host != "nas.local" {
		t.Errorf("expected Host 'nas.local', got '%s'", ep.Host)
	}
	if ep.Port != 445 {
		t.Errorf("expected Port 445, got %d", ep.Port)
	}
	if ep.Share != "data" {
		t.Errorf("expected Share 'data', got '%s'", ep.Share)
	}
	if ep.Username != "user" {
		t.Errorf("expected Username 'user', got '%s'", ep.Username)
	}
	if !ep.ReadOnly {
		t.Error("expected ReadOnly to be true")
	}
	if len(ep.Directories) != 2 {
		t.Fatalf("expected 2 directories, got %d", len(ep.Directories))
	}
	if ep.Directories[0].Path != "Music" {
		t.Errorf("expected first dir path 'Music', got '%s'", ep.Directories[0].Path)
	}
	if ep.Directories[0].ContentType != "music" {
		t.Errorf("expected first dir content_type 'music', got '%s'", ep.Directories[0].ContentType)
	}
}

func TestLoadEndpointConfig_FileNotFound(t *testing.T) {
	_, err := LoadEndpointConfig("/nonexistent/path/endpoints.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadEndpointConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(path, []byte("not valid json{{{"), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadEndpointConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Fatal("DefaultConfigPath returned empty string")
	}
	if filepath.Base(path) != "endpoints.json" {
		t.Errorf("expected filename 'endpoints.json', got '%s'", filepath.Base(path))
	}
}
