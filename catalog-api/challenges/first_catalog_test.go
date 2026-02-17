package challenges

import (
	"os"
	"path/filepath"
	"testing"

	"catalogizer/services"

	"digital.vasic.challenges/pkg/challenge"
)

// testEndpoint returns a test endpoint config for unit tests.
func testEndpoint() *Endpoint {
	return &Endpoint{
		ID:       "test-nas",
		Name:     "Test NAS",
		Host:     "nas.test",
		Port:     445,
		Share:    "testshare",
		Username: "testuser",
		Password: "testpass",
		Domain:   "",
		ReadOnly: true,
		Directories: []Directory{
			{Path: "Music", ContentType: "music"},
			{Path: "Series", ContentType: "tv_show"},
			{Path: "Movies", ContentType: "movie"},
			{Path: "Software", ContentType: "software"},
			{Path: "Comics", ContentType: "comic"},
		},
	}
}

func TestNewSMBConnectivityChallenge(t *testing.T) {
	ep := testEndpoint()
	ch := NewSMBConnectivityChallenge(ep)

	if ch.ID() != "first-catalog-smb-connect" {
		t.Errorf("expected ID 'first-catalog-smb-connect', got '%s'", ch.ID())
	}
	if ch.Name() != "SMB Connectivity" {
		t.Errorf("expected Name 'SMB Connectivity', got '%s'", ch.Name())
	}
	if ch.Category() != "integration" {
		t.Errorf("expected Category 'integration', got '%s'", ch.Category())
	}
	if len(ch.Dependencies()) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(ch.Dependencies()))
	}
}

func TestNewDirectoryDiscoveryChallenge(t *testing.T) {
	ep := testEndpoint()
	ch := NewDirectoryDiscoveryChallenge(ep)

	if ch.ID() != "first-catalog-dir-discovery" {
		t.Errorf("expected ID 'first-catalog-dir-discovery', got '%s'", ch.ID())
	}
	if ch.Name() != "Directory Discovery" {
		t.Errorf("expected Name 'Directory Discovery', got '%s'", ch.Name())
	}
	if ch.Category() != "integration" {
		t.Errorf("expected Category 'integration', got '%s'", ch.Category())
	}
	deps := ch.Dependencies()
	if len(deps) != 1 || deps[0] != "first-catalog-smb-connect" {
		t.Errorf("expected dependency [first-catalog-smb-connect], got %v", deps)
	}
}

func TestNewMusicScanChallenge(t *testing.T) {
	ep := testEndpoint()
	dir := Directory{Path: "Music", ContentType: "music"}
	ch := NewMusicScanChallenge(ep, dir)

	if ch.ID() != "first-catalog-music-scan" {
		t.Errorf("expected ID 'first-catalog-music-scan', got '%s'", ch.ID())
	}
	if ch.Category() != "e2e" {
		t.Errorf("expected Category 'e2e', got '%s'", ch.Category())
	}
	deps := ch.Dependencies()
	if len(deps) != 1 || deps[0] != "first-catalog-dir-discovery" {
		t.Errorf("expected dependency [first-catalog-dir-discovery], got %v", deps)
	}
}

func TestNewSeriesScanChallenge(t *testing.T) {
	ep := testEndpoint()
	dir := Directory{Path: "Series", ContentType: "tv_show"}
	ch := NewSeriesScanChallenge(ep, dir)

	if ch.ID() != "first-catalog-series-scan" {
		t.Errorf("expected ID 'first-catalog-series-scan', got '%s'", ch.ID())
	}
	if ch.Category() != "e2e" {
		t.Errorf("expected Category 'e2e', got '%s'", ch.Category())
	}
}

func TestNewMoviesScanChallenge(t *testing.T) {
	ep := testEndpoint()
	dir := Directory{Path: "Movies", ContentType: "movie"}
	ch := NewMoviesScanChallenge(ep, dir)

	if ch.ID() != "first-catalog-movies-scan" {
		t.Errorf("expected ID 'first-catalog-movies-scan', got '%s'", ch.ID())
	}
	if ch.Category() != "e2e" {
		t.Errorf("expected Category 'e2e', got '%s'", ch.Category())
	}
}

func TestNewSoftwareScanChallenge(t *testing.T) {
	ep := testEndpoint()
	dir := Directory{Path: "Software", ContentType: "software"}
	ch := NewSoftwareScanChallenge(ep, dir)

	if ch.ID() != "first-catalog-software-scan" {
		t.Errorf("expected ID 'first-catalog-software-scan', got '%s'", ch.ID())
	}
	if ch.Category() != "e2e" {
		t.Errorf("expected Category 'e2e', got '%s'", ch.Category())
	}
}

func TestNewComicsScanChallenge(t *testing.T) {
	ep := testEndpoint()
	dir := Directory{Path: "Comics", ContentType: "comic"}
	ch := NewComicsScanChallenge(ep, dir)

	if ch.ID() != "first-catalog-comics-scan" {
		t.Errorf("expected ID 'first-catalog-comics-scan', got '%s'", ch.ID())
	}
	if ch.Category() != "e2e" {
		t.Errorf("expected Category 'e2e', got '%s'", ch.Category())
	}
}

func TestRegisterAll_WithValidConfig(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	content := `{
		"endpoints": [
			{
				"id": "test-endpoint",
				"name": "Test NAS",
				"host": "nas.test",
				"port": 445,
				"share": "data",
				"username": "user",
				"password": "pass",
				"domain": "",
				"readonly": true,
				"directories": [
					{"path": "Music", "content_type": "music"},
					{"path": "Series", "content_type": "tv_show"},
					{"path": "Movies", "content_type": "movie"},
					{"path": "Software", "content_type": "software"},
					{"path": "Comics", "content_type": "comic"}
				]
			}
		]
	}`

	configPath := filepath.Join(configDir, "endpoints.json")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load config and register manually to test the logic
	cfg, err := LoadEndpointConfig(configPath)
	if err != nil {
		t.Fatalf("LoadEndpointConfig failed: %v", err)
	}

	svc := services.NewChallengeService(filepath.Join(dir, "results"))

	for _, ep := range cfg.Endpoints {
		endpoint := ep
		svc.Register(NewSMBConnectivityChallenge(&endpoint))
		svc.Register(NewDirectoryDiscoveryChallenge(&endpoint))
		for _, d := range endpoint.Directories {
			dd := d
			switch dd.ContentType {
			case "music":
				svc.Register(NewMusicScanChallenge(&endpoint, dd))
			case "tv_show":
				svc.Register(NewSeriesScanChallenge(&endpoint, dd))
			case "movie":
				svc.Register(NewMoviesScanChallenge(&endpoint, dd))
			case "software":
				svc.Register(NewSoftwareScanChallenge(&endpoint, dd))
			case "comic":
				svc.Register(NewComicsScanChallenge(&endpoint, dd))
			}
		}
	}

	// Verify all 7 challenges registered
	challenges := svc.ListChallenges()
	if len(challenges) != 7 {
		t.Errorf("expected 7 challenges registered, got %d", len(challenges))
		for _, c := range challenges {
			t.Logf("  registered: %s", c.ID)
		}
	}
}

func TestLoadEndpointConfig_MissingFile_NoRegistration(t *testing.T) {
	// Verify that LoadEndpointConfig returns an error for a
	// missing file, which RegisterAll handles gracefully.
	_, err := LoadEndpointConfig("/nonexistent/challenges/config/endpoints.json")
	if err == nil {
		t.Fatal("expected error for missing config file, got nil")
	}

	// Simulate what RegisterAll does: on error, register nothing
	svc := services.NewChallengeService(t.TempDir())
	challenges := svc.ListChallenges()
	if len(challenges) != 0 {
		t.Errorf("expected 0 challenges without config, got %d", len(challenges))
	}
}

func TestChallengeIDUniqueness(t *testing.T) {
	ep := testEndpoint()

	challenges := []challenge.Challenge{
		NewSMBConnectivityChallenge(ep),
		NewDirectoryDiscoveryChallenge(ep),
		NewMusicScanChallenge(ep, Directory{Path: "Music", ContentType: "music"}),
		NewSeriesScanChallenge(ep, Directory{Path: "Series", ContentType: "tv_show"}),
		NewMoviesScanChallenge(ep, Directory{Path: "Movies", ContentType: "movie"}),
		NewSoftwareScanChallenge(ep, Directory{Path: "Software", ContentType: "software"}),
		NewComicsScanChallenge(ep, Directory{Path: "Comics", ContentType: "comic"}),
	}

	seen := map[challenge.ID]bool{}
	for _, ch := range challenges {
		if seen[ch.ID()] {
			t.Errorf("duplicate challenge ID: %s", ch.ID())
		}
		seen[ch.ID()] = true
	}

	if len(seen) != 7 {
		t.Errorf("expected 7 unique challenge IDs, got %d", len(seen))
	}
}
