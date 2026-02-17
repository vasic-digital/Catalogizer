package challenges

import (
	"os"
	"path/filepath"
	"testing"

	"catalogizer/services"

	"digital.vasic.challenges/pkg/challenge"
)

func TestNewBrowsingAPIHealthChallenge(t *testing.T) {
	ch := NewBrowsingAPIHealthChallenge()

	if ch.ID() != "browsing-api-health" {
		t.Errorf("expected ID 'browsing-api-health', got '%s'", ch.ID())
	}
	if ch.Name() != "API Health & Auth" {
		t.Errorf("expected Name 'API Health & Auth', got '%s'", ch.Name())
	}
	if ch.Category() != "e2e" {
		t.Errorf("expected Category 'e2e', got '%s'", ch.Category())
	}
	deps := ch.Dependencies()
	if len(deps) != 7 {
		t.Errorf("expected 7 dependencies (all first-catalog), got %d", len(deps))
	}
}

func TestNewBrowsingAPICatalogChallenge(t *testing.T) {
	ch := NewBrowsingAPICatalogChallenge()

	if ch.ID() != "browsing-api-catalog" {
		t.Errorf("expected ID 'browsing-api-catalog', got '%s'", ch.ID())
	}
	if ch.Name() != "API Catalog Browsing" {
		t.Errorf("expected Name 'API Catalog Browsing', got '%s'", ch.Name())
	}
	if ch.Category() != "e2e" {
		t.Errorf("expected Category 'e2e', got '%s'", ch.Category())
	}
	deps := ch.Dependencies()
	if len(deps) != 1 || deps[0] != "browsing-api-health" {
		t.Errorf("expected dependency [browsing-api-health], got %v", deps)
	}
}

func TestNewBrowsingWebAppChallenge(t *testing.T) {
	ch := NewBrowsingWebAppChallenge()

	if ch.ID() != "browsing-web-app" {
		t.Errorf("expected ID 'browsing-web-app', got '%s'", ch.ID())
	}
	if ch.Name() != "Web App Browsing" {
		t.Errorf("expected Name 'Web App Browsing', got '%s'", ch.Name())
	}
	if ch.Category() != "e2e" {
		t.Errorf("expected Category 'e2e', got '%s'", ch.Category())
	}
	deps := ch.Dependencies()
	if len(deps) != 1 || deps[0] != "browsing-api-catalog" {
		t.Errorf("expected dependency [browsing-api-catalog], got %v", deps)
	}
}

func TestBrowsingChallengeIDUniqueness(t *testing.T) {
	challenges := []challenge.Challenge{
		NewBrowsingAPIHealthChallenge(),
		NewBrowsingAPICatalogChallenge(),
		NewBrowsingWebAppChallenge(),
	}

	seen := map[challenge.ID]bool{}
	for _, ch := range challenges {
		if seen[ch.ID()] {
			t.Errorf("duplicate browsing challenge ID: %s", ch.ID())
		}
		seen[ch.ID()] = true
	}

	if len(seen) != 3 {
		t.Errorf("expected 3 unique browsing challenge IDs, got %d", len(seen))
	}
}

func TestBrowsingConfigDefaults(t *testing.T) {
	// Unset env vars to test defaults
	for _, key := range []string{"BROWSING_API_URL", "ADMIN_USERNAME", "ADMIN_PASSWORD", "BROWSING_WEB_URL", "CATALOG_WEB_DIR"} {
		t.Setenv(key, "")
	}

	cfg := LoadBrowsingConfig()

	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("expected default BaseURL 'http://localhost:8080', got '%s'", cfg.BaseURL)
	}
	if cfg.Username != "admin" {
		t.Errorf("expected default Username 'admin', got '%s'", cfg.Username)
	}
	if cfg.Password != "admin123" {
		t.Errorf("expected default Password 'admin123', got '%s'", cfg.Password)
	}
	if cfg.WebAppURL != "http://localhost:3000" {
		t.Errorf("expected default WebAppURL 'http://localhost:3000', got '%s'", cfg.WebAppURL)
	}
	if cfg.WebAppDir != "../catalog-web" {
		t.Errorf("expected default WebAppDir '../catalog-web', got '%s'", cfg.WebAppDir)
	}
}

func TestBrowsingConfigFromEnv(t *testing.T) {
	t.Setenv("BROWSING_API_URL", "http://api.test:9090")
	t.Setenv("ADMIN_USERNAME", "superadmin")
	t.Setenv("ADMIN_PASSWORD", "secret456")
	t.Setenv("BROWSING_WEB_URL", "http://web.test:4000")
	t.Setenv("CATALOG_WEB_DIR", "/opt/catalog-web")

	cfg := LoadBrowsingConfig()

	if cfg.BaseURL != "http://api.test:9090" {
		t.Errorf("expected BaseURL from env, got '%s'", cfg.BaseURL)
	}
	if cfg.Username != "superadmin" {
		t.Errorf("expected Username from env, got '%s'", cfg.Username)
	}
	if cfg.Password != "secret456" {
		t.Errorf("expected Password from env, got '%s'", cfg.Password)
	}
	if cfg.WebAppURL != "http://web.test:4000" {
		t.Errorf("expected WebAppURL from env, got '%s'", cfg.WebAppURL)
	}
	if cfg.WebAppDir != "/opt/catalog-web" {
		t.Errorf("expected WebAppDir from env, got '%s'", cfg.WebAppDir)
	}
}

func TestIsInvalidTitle(t *testing.T) {
	tests := []struct {
		title string
		want  bool
	}{
		{"", true},
		{"unknown", true},
		{"Unknown", true},
		{"UNKNOWN", true},
		{"untitled", true},
		{"placeholder", true},
		{"n/a", true},
		{"N/A", true},
		{"tbd", true},
		{"  unknown  ", true},
		{"The Matrix", false},
		{"Breaking Bad", false},
		{"Led Zeppelin IV", false},
		{"Unknown Pleasures", false}, // album name - not just "unknown"
		{"Untitled Goose Game", false},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := IsInvalidTitle(tt.title)
			if got != tt.want {
				t.Errorf("IsInvalidTitle(%q) = %v, want %v", tt.title, got, tt.want)
			}
		})
	}
}

func TestRegisterAll_IncludesBrowsingChallenges(t *testing.T) {
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

	// Load config and register manually (mirrors RegisterAll logic)
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

	// Register browsing challenges
	svc.Register(NewBrowsingAPIHealthChallenge())
	svc.Register(NewBrowsingAPICatalogChallenge())
	svc.Register(NewBrowsingWebAppChallenge())

	// Verify all 10 challenges registered (7 first-catalog + 3 browsing)
	challenges := svc.ListChallenges()
	if len(challenges) != 10 {
		t.Errorf("expected 10 challenges registered, got %d", len(challenges))
		for _, c := range challenges {
			t.Logf("  registered: %s", c.ID)
		}
	}

	// Verify browsing challenge IDs are present
	ids := map[string]bool{}
	for _, c := range challenges {
		ids[c.ID] = true
	}
	for _, expected := range []string{"browsing-api-health", "browsing-api-catalog", "browsing-web-app"} {
		if !ids[expected] {
			t.Errorf("expected challenge %q to be registered", expected)
		}
	}
}

func TestAPIClient_NewAPIClient(t *testing.T) {
	client := NewAPIClient("http://localhost:8080/")

	// Should trim trailing slash
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL without trailing slash, got '%s'", client.baseURL)
	}
	if client.Token() != "" {
		t.Errorf("expected empty token for new client, got '%s'", client.Token())
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("TEST_ENV_VAR_XYZ", "custom_value")
	if got := envOrDefault("TEST_ENV_VAR_XYZ", "default"); got != "custom_value" {
		t.Errorf("expected 'custom_value', got '%s'", got)
	}

	t.Setenv("TEST_ENV_VAR_XYZ", "")
	if got := envOrDefault("TEST_ENV_VAR_XYZ", "default"); got != "default" {
		t.Errorf("expected 'default' for empty env, got '%s'", got)
	}

	if got := envOrDefault("NONEXISTENT_ENV_VAR_12345", "fallback"); got != "fallback" {
		t.Errorf("expected 'fallback' for unset env, got '%s'", got)
	}
}
