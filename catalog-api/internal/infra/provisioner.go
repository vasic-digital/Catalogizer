// Package infra provisions the catalog-api's distributed runtime
// infrastructure (PostgreSQL, Redis, MinIO) at process startup using the
// digital.vasic.containers submodule (Constitution §11.4.76 — the sole
// container-orchestration layer; no ad-hoc docker/podman in the binary path).
//
// It replaces the manual `podman-compose up` stopgap recorded in
// deploy/MIGRATION_thinker_local.md with an on-demand boot driven by
// pkg/boot.BootManager. The boot is IDEMPOTENT by design: BootManager Phase 1
// TCP-discovers every endpoint and, when a service is already reachable, marks
// it "discovered" and SKIPS provisioning. Only when the infra is genuinely DOWN
// does Phase 2 run compose-up via the (remote) orchestrator. The target host is
// config-injected (§11.4.28) — never hardcoded — and the whole feature is
// OFF by default (INFRA_PROVISION_ENABLED), so the existing working setup is
// unchanged until an operator opts in.
package infra

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"digital.vasic.containers/pkg/boot"
	"digital.vasic.containers/pkg/compose"
	"digital.vasic.containers/pkg/discovery"
	"digital.vasic.containers/pkg/endpoint"
	"digital.vasic.containers/pkg/envconfig"
	"digital.vasic.containers/pkg/health"
	clog "digital.vasic.containers/pkg/logging"
	"digital.vasic.containers/pkg/remote"
	"digital.vasic.containers/pkg/runtime"
)

// Default configuration values. Every knob is overridable via the environment
// (§11.4.28 config injection) — these defaults mirror deploy/infra-compose.yml.
const (
	defaultComposeFile = "deploy/infra-compose.yml"
	// defaultRemoteProjectDir is the parent directory of the compose file on
	// the remote host. compose derives the project name (and therefore the
	// named-volume prefix) from this directory, so it MUST match the directory
	// the infra was originally stood up from to reattach the existing data
	// volumes (§9 data safety). It mirrors the deploy/MIGRATION_thinker_local.md
	// teardown recipe (`cd ~/catalogizer-infra && podman-compose down`).
	defaultRemoteProjectDir = "catalogizer-infra"

	defaultPostgresPort = "25432"
	defaultRedisPort    = "26379"
	defaultMinioPort    = "29000"

	defaultHealthTimeout    = 5 * time.Second
	defaultDiscoveryTimeout = 3 * time.Second
	defaultBootTimeout      = 600 * time.Second
)

// Config holds the resolved infra-provisioning configuration. All fields are
// populated from environment variables by LoadConfig.
type Config struct {
	// Enabled is the master switch (INFRA_PROVISION_ENABLED). When false,
	// Provision is a no-op and the existing setup is untouched.
	Enabled bool
	// Required (INFRA_PROVISION_REQUIRED) controls fail-fast vs fail-soft in
	// the caller. It is surfaced here so callers read a single source.
	Required bool
	// TargetHost (INFRA_TARGET_HOST) is the host the services listen on, used
	// for TCP discovery + health checks. Empty or localhost => local mode.
	TargetHost string
	// ComposeFile (INFRA_COMPOSE_FILE) is the LOCAL tracked compose file — the
	// single source of truth (regeneration mechanism per §11.4.77).
	ComposeFile string
	// RemoteComposeFile (INFRA_REMOTE_COMPOSE_FILE) is the compose path as seen
	// ON the remote host; it becomes the endpoint ComposeFile that the remote
	// orchestrator runs. Empty => derived deterministically (see
	// endpointComposeFile).
	RemoteComposeFile string
	// RemoteHostName (INFRA_REMOTE_HOST_NAME) selects which CONTAINERS_REMOTE_*
	// host to SSH to. Empty => match by address, then single-host fallback.
	RemoteHostName string
	// StageCompose (INFRA_STAGE_COMPOSE) copies the local tracked ComposeFile to
	// the remote RemoteComposeFile path (mkdir -p + scp) before booting. ON by
	// default: staging runs ONLY in the remote provision-WHEN-DOWN path (after the
	// idempotent-skip pre-flight has already returned for an up stack), and the
	// down-case REQUIRES the compose file to exist at the exact path the
	// orchestrator runs against — staged into the same project dir so the existing
	// named-volume prefix is preserved and data reattaches (§9). Set
	// INFRA_STAGE_COMPOSE=false to opt out (e.g. the remote file is managed
	// out-of-band). Staging a tracked compose file is benign + idempotent.
	StageCompose bool

	PostgresPort string // INFRA_POSTGRES_PORT
	RedisPort    string // INFRA_REDIS_PORT
	MinioPort    string // INFRA_MINIO_PORT

	HealthTimeout    time.Duration // INFRA_HEALTH_TIMEOUT_SECONDS
	DiscoveryTimeout time.Duration // INFRA_DISCOVERY_TIMEOUT_SECONDS
	BootTimeout      time.Duration // INFRA_BOOT_TIMEOUT_SECONDS
}

// LoadConfig reads the infra configuration from the supplied getenv function
// (inject os.Getenv in production; a map-backed func in tests). It performs no
// I/O and is safe to call repeatedly.
func LoadConfig(getenv func(string) string) Config {
	return Config{
		Enabled:           envBool(getenv, "INFRA_PROVISION_ENABLED", false),
		Required:          envBool(getenv, "INFRA_PROVISION_REQUIRED", false),
		TargetHost:        strings.TrimSpace(getenv("INFRA_TARGET_HOST")),
		ComposeFile:       envStr(getenv, "INFRA_COMPOSE_FILE", defaultComposeFile),
		RemoteComposeFile: strings.TrimSpace(getenv("INFRA_REMOTE_COMPOSE_FILE")),
		RemoteHostName:    strings.TrimSpace(getenv("INFRA_REMOTE_HOST_NAME")),
		StageCompose:      envBool(getenv, "INFRA_STAGE_COMPOSE", true),
		PostgresPort:      envStr(getenv, "INFRA_POSTGRES_PORT", defaultPostgresPort),
		RedisPort:         envStr(getenv, "INFRA_REDIS_PORT", defaultRedisPort),
		MinioPort:         envStr(getenv, "INFRA_MINIO_PORT", defaultMinioPort),
		HealthTimeout:     envDuration(getenv, "INFRA_HEALTH_TIMEOUT_SECONDS", defaultHealthTimeout),
		DiscoveryTimeout:  envDuration(getenv, "INFRA_DISCOVERY_TIMEOUT_SECONDS", defaultDiscoveryTimeout),
		BootTimeout:       envDuration(getenv, "INFRA_BOOT_TIMEOUT_SECONDS", defaultBootTimeout),
	}
}

// isRemote reports whether the target is a remote host (i.e. provisioning runs
// over SSH via the containers submodule). An empty/loopback target means the
// infra is local to this process.
func (c Config) isRemote() bool {
	h := strings.TrimSpace(c.TargetHost)
	return h != "" && h != "localhost" && h != "127.0.0.1" && h != "::1"
}

// serviceHost returns the host the services are reachable at for discovery +
// health checks.
func (c Config) serviceHost() string {
	if c.isRemote() {
		return c.TargetHost
	}
	return "localhost"
}

// resolveHostForDial returns an address Go's net.Dialer can actually reach.
//
// Forensic FACT (2026-06-29, §11.4.102): Go's pure-Go resolver cannot resolve
// mDNS ".local" names (avahi / nss-mdns), so a TCP dial to "thinker.local:25432"
// fails with "lookup thinker.local: i/o timeout" — even though ssh, ping and
// `getent` (which use the system NSS resolver, mDNS included) reach it fine.
// That broke BootManager's discovery phase (0 discovered → it fell through to
// provisioning instead of the idempotent skip). The fix: when Go's own
// LookupHost fails for a name, fall back to the system resolver via `getent
// hosts`. Returns the input unchanged when it is already an IP, when resolution
// succeeds via Go, or when no system fallback is available — in the last case
// the dial then fails HONESTLY rather than silently (§11.4.1).
//
// §11.4.81 cross-platform: `getent` is glibc/Linux; other platforms keep the
// hostname (honest best-effort) — a consuming deployment on a non-glibc host
// should set INFRA_TARGET_HOST to an IP.
func resolveHostForDial(host string) string {
	h := strings.TrimSpace(host)
	if h == "" || net.ParseIP(h) != nil {
		return host
	}
	if addrs, err := net.LookupHost(h); err == nil && len(addrs) > 0 {
		return addrs[0]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "getent", "hosts", h).Output()
	if err == nil {
		if fields := strings.Fields(string(out)); len(fields) > 0 && net.ParseIP(fields[0]) != nil {
			return fields[0]
		}
	}
	return host
}

// endpointComposeFile returns the compose-file path that the active
// orchestrator will use. For the remote orchestrator this is a path on the
// REMOTE host (relative paths resolve against the SSH login home); for the
// local orchestrator it is an absolute local path.
func (c Config) endpointComposeFile() (string, error) {
	if c.isRemote() {
		if c.RemoteComposeFile != "" {
			return c.RemoteComposeFile, nil
		}
		// Deterministic default: <project-dir>/<basename> under the remote
		// home so the compose project name (and named-volume prefix) is stable
		// and matches the original stand-up directory (§9 data reattachment).
		return path.Join(defaultRemoteProjectDir, filepath.Base(c.ComposeFile)), nil
	}
	abs, err := filepath.Abs(c.ComposeFile)
	if err != nil {
		return "", fmt.Errorf("infra: resolve compose file %q: %w", c.ComposeFile, err)
	}
	return abs, nil
}

// buildEndpoints constructs the three service endpoints (postgres/redis/minio)
// from the resolved config. All three share one ComposeFile so BootManager
// groups them into a single compose-up call. Pure function — no I/O.
func buildEndpoints(cfg Config) (map[string]endpoint.ServiceEndpoint, error) {
	composeFile, err := cfg.endpointComposeFile()
	if err != nil {
		return nil, err
	}
	host := cfg.serviceHost()
	isRemote := cfg.isRemote()

	specs := []struct {
		name string
		port string
	}{
		{"postgres", cfg.PostgresPort},
		{"redis", cfg.RedisPort},
		{"minio", cfg.MinioPort},
	}

	eps := make(map[string]endpoint.ServiceEndpoint, len(specs))
	for _, s := range specs {
		eps[s.name] = endpoint.ServiceEndpoint{
			Host:             host,
			Port:             s.port,
			Enabled:          true,
			Required:         true,
			Remote:           isRemote,
			HealthType:       "tcp",
			Timeout:          cfg.HealthTimeout,
			RetryCount:       3,
			ComposeFile:      composeFile,
			ServiceName:      s.name,
			DiscoveryEnabled: true,
			DiscoveryMethod:  "tcp",
			DiscoveryTimeout: cfg.DiscoveryTimeout,
		}
	}
	return eps, nil
}

// Provision provisions the distributed infra using the containers submodule.
// It is a no-op (returns nil) when INFRA_PROVISION_ENABLED is unset/false, so
// the default behaviour of the binary is unchanged. When enabled and the infra
// is already up, BootManager discovers it and skips provisioning (idempotent).
//
// The returned error is informational for the caller to decide fail-fast
// (INFRA_PROVISION_REQUIRED=true) vs fail-soft.
func Provision(ctx context.Context) error {
	return provision(ctx, LoadConfig(os.Getenv), clog.NewStdLogger("infra-provision"))
}

// unreachableEndpoints returns the names of endpoints that do NOT accept a TCP
// connection within a short probe timeout. An empty slice means the whole stack
// is already up (the idempotent skip path). Each endpoint's Host is already an
// IP resolved by buildEndpoints (resolveHostForDial), so this probe does not
// hit the Go-resolver mDNS gap.
func unreachableEndpoints(eps map[string]endpoint.ServiceEndpoint) []string {
	var down []string
	for name, ep := range eps {
		addr := net.JoinHostPort(resolveHostForDial(ep.Host), ep.Port)
		conn, derr := (&net.Dialer{Timeout: 2 * time.Second}).Dial("tcp", addr)
		if derr != nil {
			down = append(down, name)
			continue
		}
		_ = conn.Close()
	}
	return down
}

// provision is the testable core of Provision with an injected logger.
func provision(ctx context.Context, cfg Config, logger clog.Logger) error {
	if !cfg.Enabled {
		logger.Info("provisioning disabled (INFRA_PROVISION_ENABLED unset/false); skipping")
		return nil
	}

	eps, err := buildEndpoints(cfg)
	if err != nil {
		return err
	}

	// Idempotent pre-flight (§11.4.76 on-demand skip, made robust). If every
	// endpoint already accepts a TCP connection the infra is up, so return
	// WITHOUT invoking the heavier remote-SSH BootManager. This is also a
	// deliberate workaround for a submodule quirk: BootManager.BootAll's Phase-3
	// health check re-evaluates an already-"discovered" service and counts it as
	// "failed", turning a healthy steady state into a boot error. Probing here
	// keeps the common case a clean no-op; only a genuinely-DOWN stack falls
	// through to provisioning below.
	if down := unreachableEndpoints(eps); len(down) == 0 {
		logger.Info("infra already reachable (%d endpoints) — provisioning skipped", len(eps))
		return nil
	} else {
		logger.Info("infra endpoints needing provisioning: %v", down)
	}

	// Provisioning IS needed. Resolve each endpoint's Host to an IP so the
	// submodule's Go-resolver-based discovery + health phases can reach an mDNS
	// ".local" target (the resolveHostForDial gap). SSH to the remote host uses
	// the separately-configured CONTAINERS_REMOTE_* address, so this only
	// affects the discovery/health dials, not the compose transport.
	for name, ep := range eps {
		ep.Host = resolveHostForDial(ep.Host)
		eps[name] = ep
	}

	if cfg.BootTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.BootTimeout)
		defer cancel()
	}

	opts := []boot.BootManagerOption{
		boot.WithDiscoverer(discovery.NewTCPDiscoverer()),
		boot.WithHealthChecker(health.NewDefaultChecker()),
		boot.WithLogger(logger),
	}

	// Local runtime detection is best-effort: the remote compose path does not
	// need a local runtime, so a missing one must not abort provisioning.
	if rt, derr := runtime.AutoDetect(ctx); derr == nil {
		opts = append(opts, boot.WithRuntime(rt))
	} else {
		logger.Warn("no local container runtime detected (ok for remote compose): %v", derr)
	}

	if cfg.isRemote() {
		host, executor, herr := buildRemoteHost(cfg, logger)
		if herr != nil {
			return herr
		}
		defer func() { _ = executor.Close() }()

		hm := remote.NewHostManager(executor, logger)
		if aerr := hm.AddHost(host); aerr != nil {
			return fmt.Errorf("infra: register remote host %q: %w", host.Name, aerr)
		}

		if cfg.StageCompose {
			if serr := stageComposeFile(ctx, executor, host, cfg, logger); serr != nil {
				// Fail-soft: a staging failure must not abort the boot — the
				// remote file may already be in place from a prior deploy.
				logger.Warn("compose staging to %s failed (continuing): %v", host.Name, serr)
			}
		}

		orch := remote.NewRemoteComposeOrchestrator(host, executor, logger)
		opts = append(opts,
			boot.WithOrchestrator(orch),
			boot.WithHostManager(hm),
		)
	} else {
		wd, _ := os.Getwd()
		orch, oerr := compose.NewDefaultOrchestrator(wd, logger)
		if oerr != nil {
			return fmt.Errorf("infra: local compose orchestrator: %w", oerr)
		}
		opts = append(opts,
			boot.WithOrchestrator(orch),
			boot.WithProjectDir(wd),
		)
	}

	bm := boot.NewBootManager(eps, opts...)
	summary, berr := bm.BootAll(ctx)
	if summary != nil {
		logger.Info("infra %s", summary.String())
	}
	if berr != nil {
		return fmt.Errorf("infra: boot failed: %w", berr)
	}
	logger.Info(
		"infra provisioning complete (discovered=%d started=%d remote=%d skipped=%d failed=%d)",
		summary.Discovered, summary.Started, summary.Remote, summary.Skipped, summary.Failed,
	)
	return nil
}

// buildRemoteHost loads the CONTAINERS_REMOTE_* configuration, selects the
// target host, and builds an SSH executor for it.
func buildRemoteHost(cfg Config, logger clog.Logger) (remote.RemoteHost, *remote.SSHExecutor, error) {
	dc := envconfig.LoadFromEnv()
	hosts := dc.ToRemoteHosts()
	if len(hosts) == 0 {
		return remote.RemoteHost{}, nil, fmt.Errorf(
			"infra: remote target %q requires CONTAINERS_REMOTE_* host configuration (none found)",
			cfg.TargetHost,
		)
	}

	host, err := selectHost(hosts, cfg)
	if err != nil {
		return remote.RemoteHost{}, nil, err
	}

	executor, err := remote.NewSSHExecutor(logger,
		remote.WithConnectTimeout(time.Duration(dc.ConnectTimeout)*time.Second),
		remote.WithCommandTimeout(time.Duration(dc.CommandTimeout)*time.Second),
		remote.WithControlMaster(dc.ControlMasterEnabled),
		remote.WithControlPersist(time.Duration(dc.ControlPersist)*time.Second),
		remote.WithMaxConnections(dc.MaxConnections),
	)
	if err != nil {
		return remote.RemoteHost{}, nil, fmt.Errorf("infra: build ssh executor: %w", err)
	}
	return host, executor, nil
}

// selectHost chooses the remote host to provision on from the configured hosts.
// Resolution order (deterministic, no guessing per §11.4.6):
//  1. explicit INFRA_REMOTE_HOST_NAME match,
//  2. a host whose Address equals INFRA_TARGET_HOST,
//  3. the single configured host.
func selectHost(hosts []remote.RemoteHost, cfg Config) (remote.RemoteHost, error) {
	if cfg.RemoteHostName != "" {
		for _, h := range hosts {
			if h.Name == cfg.RemoteHostName {
				return h, nil
			}
		}
		return remote.RemoteHost{}, fmt.Errorf(
			"infra: no CONTAINERS_REMOTE host named %q", cfg.RemoteHostName)
	}
	for _, h := range hosts {
		if h.Address == cfg.TargetHost {
			return h, nil
		}
	}
	if len(hosts) == 1 {
		return hosts[0], nil
	}
	return remote.RemoteHost{}, fmt.Errorf(
		"infra: cannot select a remote host for target %q among %d configured hosts; "+
			"set INFRA_REMOTE_HOST_NAME", cfg.TargetHost, len(hosts))
}

// stageComposeFile copies the local tracked compose file to the remote host at
// the endpoint's compose path, creating the parent directory first. Uses the
// submodule's RemoteExecutor primitives (scp/ssh) — not an ad-hoc command.
func stageComposeFile(
	ctx context.Context,
	executor remote.RemoteExecutor,
	host remote.RemoteHost,
	cfg Config,
	logger clog.Logger,
) error {
	localAbs, err := filepath.Abs(cfg.ComposeFile)
	if err != nil {
		return fmt.Errorf("resolve local compose %q: %w", cfg.ComposeFile, err)
	}
	if _, statErr := os.Stat(localAbs); statErr != nil {
		return fmt.Errorf("local compose file %q not readable: %w", localAbs, statErr)
	}

	remotePath, err := cfg.endpointComposeFile()
	if err != nil {
		return err
	}

	if dir := path.Dir(remotePath); dir != "" && dir != "." && dir != "/" {
		if _, mkErr := executor.Execute(ctx, host, fmt.Sprintf("mkdir -p '%s'", dir)); mkErr != nil {
			return fmt.Errorf("mkdir -p %q on %s: %w", dir, host.Name, mkErr)
		}
	}

	logger.Info("staging compose file %s -> %s:%s", localAbs, host.Name, remotePath)
	return executor.CopyFile(ctx, host, localAbs, remotePath)
}

// --- small env helpers (no external deps) ---

func envStr(getenv func(string) string, key, fallback string) string {
	if v := strings.TrimSpace(getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(getenv func(string) string, key string, fallback bool) bool {
	v := strings.TrimSpace(getenv(key))
	if v == "" {
		return fallback
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return fallback
}

func envDuration(getenv func(string) string, key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(getenv(key))
	if v == "" {
		return fallback
	}
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return fallback
}
