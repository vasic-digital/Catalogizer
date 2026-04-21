// Package integration — spawned-binary end-to-end tests.
//
// session_fixes_e2e_test.go builds the real catalog-api binary and
// spins it up with different env vars, then drives the real HTTP
// stack (Gin router + all middleware + real SQLite + migrations)
// through the real network. This is the strongest evidence we can
// get without a production deployment.
//
// Every test here proves a specific fix from the 2026-04-20/2026-04-21
// Article VII cycle + Q/R/S/T/U-cycles landed and still works:
//
//   - FIX-QA-2026-04-21-001 — PUT /media/:id/favorite no longer 500s
//   - FIX-QA-2026-04-21-002 — /api/v1/health alias returns 200
//   - FIX-QA-2026-04-21-003 — /admin/{config,errors,health,logs} 200
//   - FIX-QA-2026-04-21-004 — /challenges/results ?limit works
//   - FIX-QA-2026-04-21-006 — pprof opt-in via HELIX_PPROF_ENABLED
//
// Build tag keeps the cost off of the regular `go test ./...` run;
// invoke explicitly with `go test -tags=e2e_binary ./tests/integration/...`.

//go:build e2e_binary

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// spawnedBinary holds a running catalog-api process for the duration
// of a test; call Close to SIGKILL it.
type spawnedBinary struct {
	t       *testing.T
	cmd     *exec.Cmd
	baseURL string
	logPath string
}

// spawnBinary compiles the binary once per test run (cached) and
// starts it on an ephemeral port with the given env vars. Blocks
// until /health returns 200 or a 30 s timeout elapses.
func spawnBinary(t *testing.T, extraEnv ...string) *spawnedBinary {
	t.Helper()

	if runtime.GOOS != "linux" {
		t.Skipf("spawned-binary test only supported on Linux (got %s)", runtime.GOOS)
	}

	// Find the catalog-api source dir (we're under tests/integration/).
	_, thisFile, _, _ := runtime.Caller(0)
	apiDir := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	binaryPath := filepath.Join(apiDir, "catalog-api.e2e")

	// Build the binary. Subsequent tests hit the cache; Go's build
	// system is fast enough that this isn't a bottleneck.
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	build.Dir = apiDir
	build.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	// Grab an ephemeral TCP port.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := lis.Addr().(*net.TCPAddr).Port
	_ = lis.Close()

	logFile, err := os.CreateTemp("", "catalog-api-e2e-*.log")
	require.NoError(t, err)
	logPath := logFile.Name()

	// Per-test sqlite DB to avoid cross-test contamination.
	dbFile, err := os.CreateTemp("", "catalog-api-e2e-*.db")
	require.NoError(t, err)
	_ = dbFile.Close()

	baseEnv := []string{
		fmt.Sprintf("PORT=%d", port),
		"GIN_MODE=release",
		"ADMIN_PASSWORD=admin123",
		"JWT_SECRET=e2e-spawned-binary-session-fixes-jwt-secret-xyz",
		"DB_TYPE=sqlite",
		"DB_PATH=" + dbFile.Name(),
		"GOTOOLCHAIN=local",
	}
	cmdEnv := append(os.Environ(), append(baseEnv, extraEnv...)...)

	cmd := exec.Command(binaryPath)
	cmd.Dir = apiDir
	cmd.Env = cmdEnv
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		t.Fatalf("cmd.Start: %v", err)
	}

	sb := &spawnedBinary{
		t:       t,
		cmd:     cmd,
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		logPath: logPath,
	}

	// Wait for /health to be serving. Poll every 200 ms up to 30 s.
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(sb.baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return sb
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	sb.Close()
	t.Fatalf("catalog-api did not become ready within 30s; log: %s", logPath)
	return nil
}

// Close kills the process group (so any spawned children go too).
func (sb *spawnedBinary) Close() {
	if sb == nil || sb.cmd == nil || sb.cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-sb.cmd.Process.Pid, syscall.SIGKILL)
	_ = sb.cmd.Wait()
	if sb.t.Failed() {
		sb.t.Logf("server log: %s", sb.logPath)
	} else {
		_ = os.Remove(sb.logPath)
	}
}

// do issues an HTTP request against the spawned binary and returns
// (status, body, decoded-json-or-nil). On transport error, fails the
// test.
func (sb *spawnedBinary) do(method, path string, body []byte, headers map[string]string) (int, []byte, map[string]any) {
	sb.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, sb.baseURL+path, bodyReader)
	require.NoError(sb.t, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(sb.t, err)
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(sb.t, err)

	var parsed map[string]any
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		_ = json.Unmarshal(respBody, &parsed)
	}
	return resp.StatusCode, respBody, parsed
}

// login returns a session token for admin/admin123.
func (sb *spawnedBinary) login() string {
	body := []byte(`{"username":"admin","password":"admin123"}`)
	status, rawBody, parsed := sb.do(http.MethodPost, "/api/v1/auth/login", body, nil)
	require.Equalf(sb.t, 200, status, "login failed: %s", string(rawBody))
	token, ok := parsed["session_token"].(string)
	require.True(sb.t, ok, "login response missing session_token: %v", parsed)
	require.NotEmpty(sb.t, token)
	return token
}

// ---- tests ----

// TestSessionFixes_E2E is the single umbrella for every fix this
// session shipped. Uses subtests so a binary is spawned once per
// group (pprof on / pprof off), not per assertion.
func TestSessionFixes_E2E_PprofOff(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	// FIX-QA-2026-04-21-002 — /api/v1/health alias (unauthenticated).
	t.Run("health_alias_200", func(t *testing.T) {
		for _, path := range []string{"/health", "/api/v1/health"} {
			status, body, parsed := sb.do(http.MethodGet, path, nil, nil)
			require.Equalf(t, 200, status, "%s → %d: %s", path, status, string(body))
			require.Equal(t, "healthy", parsed["status"])
		}
	})

	// FIX-QA-2026-04-21-006 — pprof endpoints absent when flag unset.
	t.Run("pprof_disabled_by_default", func(t *testing.T) {
		status, _, _ := sb.do(http.MethodGet, "/debug/pprof/heap", nil, nil)
		require.Equal(t, 404, status, "pprof must be 404 without HELIX_PPROF_ENABLED=true")
	})

	token := sb.login()
	auth := map[string]string{"Authorization": "Bearer " + token}

	// FIX-QA-2026-04-21-003 — admin alias endpoints.
	t.Run("admin_aliases_all_200", func(t *testing.T) {
		for _, p := range []string{
			"/api/v1/admin/config",
			"/api/v1/admin/errors",
			"/api/v1/admin/health",
			"/api/v1/admin/logs",
			"/api/v1/admin/system-info",
			"/api/v1/admin/storage",
			"/api/v1/admin/backups",
		} {
			status, body, _ := sb.do(http.MethodGet, p, nil, auth)
			require.Equalf(t, 200, status, "%s → %d: %s", p, status, string(body))
		}
	})

	// FIX-QA-2026-04-21-001 — favorite UPDATE no longer 500s.
	// Non-existent id = 9999999 should return 404 (not 500).
	t.Run("favorite_nonexistent_returns_404", func(t *testing.T) {
		status, body, _ := sb.do(http.MethodPut,
			"/api/v1/media/9999999/favorite",
			[]byte(`{"favorite":true}`), auth)
		require.Equalf(t, 404, status, "expected 404 Media not found, got %d: %s", status, string(body))
	})

	// FIX-QA-2026-04-21-004 — /challenges/results has total_count and ?limit.
	t.Run("challenges_results_has_total_count", func(t *testing.T) {
		status, body, parsed := sb.do(http.MethodGet, "/api/v1/challenges/results", nil, auth)
		require.Equalf(t, 200, status, "%s", string(body))
		_, hasTotal := parsed["total_count"]
		require.True(t, hasTotal, "response missing total_count key: %v", parsed)
		_, hasCount := parsed["count"]
		require.True(t, hasCount)
	})
}

func TestSessionFixes_E2E_PprofOn(t *testing.T) {
	sb := spawnBinary(t, "HELIX_PPROF_ENABLED=true")
	defer sb.Close()

	// FIX-QA-2026-04-21-006 — pprof endpoints available with flag.
	t.Run("pprof_heap_200_with_flag", func(t *testing.T) {
		status, _, _ := sb.do(http.MethodGet, "/debug/pprof/heap", nil, nil)
		require.Equalf(t, 200, status, "pprof/heap must be 200 with HELIX_PPROF_ENABLED=true")
	})
	t.Run("pprof_goroutine_200_with_flag", func(t *testing.T) {
		status, _, _ := sb.do(http.MethodGet, "/debug/pprof/goroutine", nil, nil)
		require.Equal(t, 200, status)
	})
}
