package infra

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	clog "digital.vasic.containers/pkg/logging"
	"digital.vasic.containers/pkg/remote"
)

// mapEnv returns a getenv-style func backed by a map, for deterministic
// config-parsing tests with no process-environment dependency.
func mapEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadConfig_Defaults(t *testing.T) {
	cfg := LoadConfig(mapEnv(nil))

	if cfg.Enabled {
		t.Errorf("Enabled default = true, want false (must be off by default)")
	}
	if cfg.Required {
		t.Errorf("Required default = true, want false")
	}
	if cfg.TargetHost != "" {
		t.Errorf("TargetHost default = %q, want empty", cfg.TargetHost)
	}
	if cfg.ComposeFile != defaultComposeFile {
		t.Errorf("ComposeFile default = %q, want %q", cfg.ComposeFile, defaultComposeFile)
	}
	if cfg.PostgresPort != defaultPostgresPort {
		t.Errorf("PostgresPort default = %q, want %q", cfg.PostgresPort, defaultPostgresPort)
	}
	if cfg.RedisPort != defaultRedisPort {
		t.Errorf("RedisPort default = %q, want %q", cfg.RedisPort, defaultRedisPort)
	}
	if cfg.MinioPort != defaultMinioPort {
		t.Errorf("MinioPort default = %q, want %q", cfg.MinioPort, defaultMinioPort)
	}
	if cfg.HealthTimeout != defaultHealthTimeout {
		t.Errorf("HealthTimeout default = %v, want %v", cfg.HealthTimeout, defaultHealthTimeout)
	}
	if cfg.DiscoveryTimeout != defaultDiscoveryTimeout {
		t.Errorf("DiscoveryTimeout default = %v, want %v", cfg.DiscoveryTimeout, defaultDiscoveryTimeout)
	}
	if cfg.isRemote() {
		t.Errorf("isRemote() = true with empty TargetHost, want false")
	}
}

func TestLoadConfig_EnabledRemote(t *testing.T) {
	cfg := LoadConfig(mapEnv(map[string]string{
		"INFRA_PROVISION_ENABLED":         "true",
		"INFRA_PROVISION_REQUIRED":        "true",
		"INFRA_TARGET_HOST":               "thinker.local",
		"INFRA_COMPOSE_FILE":              "deploy/infra-compose.yml",
		"INFRA_STAGE_COMPOSE":             "true",
		"INFRA_POSTGRES_PORT":             "25432",
		"INFRA_HEALTH_TIMEOUT_SECONDS":    "9",
		"INFRA_DISCOVERY_TIMEOUT_SECONDS": "7",
		"INFRA_BOOT_TIMEOUT_SECONDS":      "300",
	}))

	if !cfg.Enabled {
		t.Errorf("Enabled = false, want true")
	}
	if !cfg.Required {
		t.Errorf("Required = false, want true")
	}
	if cfg.TargetHost != "thinker.local" {
		t.Errorf("TargetHost = %q, want thinker.local", cfg.TargetHost)
	}
	if !cfg.StageCompose {
		t.Errorf("StageCompose = false, want true")
	}
	if cfg.HealthTimeout != 9*time.Second {
		t.Errorf("HealthTimeout = %v, want 9s", cfg.HealthTimeout)
	}
	if cfg.DiscoveryTimeout != 7*time.Second {
		t.Errorf("DiscoveryTimeout = %v, want 7s", cfg.DiscoveryTimeout)
	}
	if cfg.BootTimeout != 300*time.Second {
		t.Errorf("BootTimeout = %v, want 300s", cfg.BootTimeout)
	}
	if !cfg.isRemote() {
		t.Errorf("isRemote() = false for thinker.local, want true")
	}
}

func TestLoadConfig_InvalidBoolFallsBackToDefault(t *testing.T) {
	cfg := LoadConfig(mapEnv(map[string]string{
		"INFRA_PROVISION_ENABLED": "not-a-bool",
	}))
	if cfg.Enabled {
		t.Errorf("Enabled = true for invalid bool, want false (fallback to default)")
	}
}

func TestIsRemote_LoopbackIsLocal(t *testing.T) {
	for _, h := range []string{"", "localhost", "127.0.0.1", "::1"} {
		cfg := Config{TargetHost: h}
		if cfg.isRemote() {
			t.Errorf("isRemote() = true for %q, want false", h)
		}
	}
}

func TestBuildEndpoints_Remote(t *testing.T) {
	cfg := LoadConfig(mapEnv(map[string]string{
		"INFRA_PROVISION_ENABLED": "true",
		"INFRA_TARGET_HOST":       "thinker.local",
	}))

	eps, err := buildEndpoints(cfg)
	if err != nil {
		t.Fatalf("buildEndpoints error: %v", err)
	}
	if len(eps) != 3 {
		t.Fatalf("got %d endpoints, want 3", len(eps))
	}

	wantPort := map[string]string{
		"postgres": defaultPostgresPort,
		"redis":    defaultRedisPort,
		"minio":    defaultMinioPort,
	}

	// All endpoints must share one compose file so BootManager groups them
	// into a single compose-up call.
	var sharedCompose string
	for name, ep := range eps {
		if ep.Host != "thinker.local" {
			t.Errorf("%s Host = %q, want thinker.local", name, ep.Host)
		}
		if !ep.Remote {
			t.Errorf("%s Remote = false, want true", name)
		}
		if !ep.Required {
			t.Errorf("%s Required = false, want true", name)
		}
		if !ep.Enabled {
			t.Errorf("%s Enabled = false, want true", name)
		}
		if ep.HealthType != "tcp" {
			t.Errorf("%s HealthType = %q, want tcp", name, ep.HealthType)
		}
		if !ep.DiscoveryEnabled {
			t.Errorf("%s DiscoveryEnabled = false, want true", name)
		}
		if ep.DiscoveryMethod != "tcp" {
			t.Errorf("%s DiscoveryMethod = %q, want tcp", name, ep.DiscoveryMethod)
		}
		if ep.ServiceName != name {
			t.Errorf("%s ServiceName = %q, want %q", name, ep.ServiceName, name)
		}
		if ep.Port != wantPort[name] {
			t.Errorf("%s Port = %q, want %q", name, ep.Port, wantPort[name])
		}
		if ep.ComposeFile == "" {
			t.Errorf("%s ComposeFile is empty", name)
		}
		if sharedCompose == "" {
			sharedCompose = ep.ComposeFile
		} else if ep.ComposeFile != sharedCompose {
			t.Errorf("%s ComposeFile = %q, want shared %q", name, ep.ComposeFile, sharedCompose)
		}
		if ep.DiscoveryTimeout != defaultDiscoveryTimeout {
			t.Errorf("%s DiscoveryTimeout = %v, want %v", name, ep.DiscoveryTimeout, defaultDiscoveryTimeout)
		}
		if ep.Timeout != defaultHealthTimeout {
			t.Errorf("%s health Timeout = %v, want %v", name, ep.Timeout, defaultHealthTimeout)
		}
	}

	// Remote compose path: derived default <project-dir>/<basename>.
	wantRemoteCompose := filepath.ToSlash(filepath.Join(defaultRemoteProjectDir, "infra-compose.yml"))
	if sharedCompose != wantRemoteCompose {
		t.Errorf("remote ComposeFile = %q, want %q", sharedCompose, wantRemoteCompose)
	}
}

func TestBuildEndpoints_CustomPortsAndRemoteCompose(t *testing.T) {
	cfg := LoadConfig(mapEnv(map[string]string{
		"INFRA_PROVISION_ENABLED":   "true",
		"INFRA_TARGET_HOST":         "thinker.local",
		"INFRA_POSTGRES_PORT":       "15432",
		"INFRA_REDIS_PORT":          "16379",
		"INFRA_MINIO_PORT":          "19000",
		"INFRA_REMOTE_COMPOSE_FILE": "/srv/catalogizer-infra/infra-compose.yml",
	}))

	eps, err := buildEndpoints(cfg)
	if err != nil {
		t.Fatalf("buildEndpoints error: %v", err)
	}
	if eps["postgres"].Port != "15432" {
		t.Errorf("postgres Port = %q, want 15432", eps["postgres"].Port)
	}
	if eps["redis"].Port != "16379" {
		t.Errorf("redis Port = %q, want 16379", eps["redis"].Port)
	}
	if eps["minio"].Port != "19000" {
		t.Errorf("minio Port = %q, want 19000", eps["minio"].Port)
	}
	if eps["postgres"].ComposeFile != "/srv/catalogizer-infra/infra-compose.yml" {
		t.Errorf("ComposeFile = %q, want explicit remote path", eps["postgres"].ComposeFile)
	}
}

func TestBuildEndpoints_LocalMode(t *testing.T) {
	cfg := LoadConfig(mapEnv(map[string]string{
		"INFRA_PROVISION_ENABLED": "true",
		// No TargetHost => local mode.
		"INFRA_COMPOSE_FILE": "deploy/infra-compose.yml",
	}))

	eps, err := buildEndpoints(cfg)
	if err != nil {
		t.Fatalf("buildEndpoints error: %v", err)
	}
	for name, ep := range eps {
		if ep.Remote {
			t.Errorf("%s Remote = true in local mode, want false", name)
		}
		if ep.Host != "localhost" {
			t.Errorf("%s Host = %q, want localhost", name, ep.Host)
		}
		if !filepath.IsAbs(ep.ComposeFile) {
			t.Errorf("%s ComposeFile = %q, want absolute local path", name, ep.ComposeFile)
		}
	}
}

func TestProvision_DisabledIsNoop(t *testing.T) {
	// With provisioning disabled, provision must return nil and perform no
	// network I/O regardless of any remote target.
	cfg := Config{Enabled: false, TargetHost: "thinker.local"}
	if err := provision(context.Background(), cfg, clog.NopLogger{}); err != nil {
		t.Errorf("provision(disabled) = %v, want nil", err)
	}
}

func TestSelectHost_ByName(t *testing.T) {
	hosts := []remote.RemoteHost{
		{Name: "alpha", Address: "alpha.local"},
		{Name: "thinker", Address: "thinker.local"},
	}
	got, err := selectHost(hosts, Config{RemoteHostName: "thinker", TargetHost: "thinker.local"})
	if err != nil {
		t.Fatalf("selectHost error: %v", err)
	}
	if got.Name != "thinker" {
		t.Errorf("selected %q, want thinker", got.Name)
	}
}

func TestSelectHost_ByAddressMatch(t *testing.T) {
	hosts := []remote.RemoteHost{
		{Name: "alpha", Address: "alpha.local"},
		{Name: "node2", Address: "thinker.local"},
	}
	got, err := selectHost(hosts, Config{TargetHost: "thinker.local"})
	if err != nil {
		t.Fatalf("selectHost error: %v", err)
	}
	if got.Address != "thinker.local" {
		t.Errorf("selected address %q, want thinker.local", got.Address)
	}
}

func TestSelectHost_SingleFallback(t *testing.T) {
	hosts := []remote.RemoteHost{{Name: "only", Address: "only.local"}}
	got, err := selectHost(hosts, Config{TargetHost: "thinker.local"})
	if err != nil {
		t.Fatalf("selectHost error: %v", err)
	}
	if got.Name != "only" {
		t.Errorf("selected %q, want only", got.Name)
	}
}

func TestSelectHost_AmbiguousErrors(t *testing.T) {
	hosts := []remote.RemoteHost{
		{Name: "a", Address: "a.local"},
		{Name: "b", Address: "b.local"},
	}
	if _, err := selectHost(hosts, Config{TargetHost: "thinker.local"}); err == nil {
		t.Errorf("selectHost ambiguous = nil error, want error")
	}
}

func TestSelectHost_NamedNotFoundErrors(t *testing.T) {
	hosts := []remote.RemoteHost{{Name: "only", Address: "only.local"}}
	if _, err := selectHost(hosts, Config{RemoteHostName: "missing"}); err == nil {
		t.Errorf("selectHost named-not-found = nil error, want error")
	}
}
