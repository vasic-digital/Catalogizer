package tests

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

// Default addresses for test infrastructure services.
// These match the ports defined in docker-compose.test-infra.yml.
// Override via environment variables if services run on non-default addresses.
const (
	DefaultFTPAddr    = "localhost:2121"
	DefaultSMBAddr    = "localhost:1445"
	DefaultWebDAVAddr = "localhost:8081"
	DefaultNFSAddr    = "localhost:2049"

	// connectionTimeout is the maximum time to wait when probing a service.
	// Kept short (1 second) so tests skip quickly when infrastructure is absent.
	connectionTimeout = 1 * time.Second
)

// InfraStatus holds the reachability state of all test infrastructure services.
type InfraStatus struct {
	FTP    bool
	SMB    bool
	WebDAV bool
	NFS    bool
}

// TestInfraCredentials provides default credentials for test infrastructure services.
// These match the credentials configured in docker-compose.test-infra.yml.
type TestInfraCredentials struct {
	Username string
	Password string
}

// FTPCredentials returns the default FTP test credentials.
func FTPCredentials() TestInfraCredentials {
	return TestInfraCredentials{
		Username: envOrDefault("FTP_TEST_USER", "testuser"),
		Password: envOrDefault("FTP_TEST_PASS", "testpass"),
	}
}

// SMBCredentials returns the default SMB test credentials.
// The test share ("testshare") also allows guest access.
func SMBCredentials() TestInfraCredentials {
	return TestInfraCredentials{
		Username: envOrDefault("SMB_TEST_USER", "testuser"),
		Password: envOrDefault("SMB_TEST_PASS", "testpass"),
	}
}

// WebDAVCredentials returns the default WebDAV test credentials.
func WebDAVCredentials() TestInfraCredentials {
	return TestInfraCredentials{
		Username: envOrDefault("WEBDAV_TEST_USER", "testuser"),
		Password: envOrDefault("WEBDAV_TEST_PASS", "testpass"),
	}
}

// SMBShareName returns the default SMB test share name.
func SMBShareName() string {
	return envOrDefault("SMB_TEST_SHARE", "testshare")
}

// ---------------------------------------------------------------------------
// Require* functions — call these at the start of integration tests.
// They skip the test with t.Skip() if the corresponding service is not
// reachable, rather than failing. This allows the full test suite to run
// even when test infrastructure is partially or fully absent.
// ---------------------------------------------------------------------------

// RequireFTP skips the test if the FTP test server is not reachable.
// Returns the FTP server address for use in test configuration.
func RequireFTP(t *testing.T) string {
	t.Helper()
	addr := ftpAddr()
	if !isReachable(addr) {
		t.Skipf("FTP test server not reachable at %s — start test infrastructure with: podman-compose -f docker-compose.test-infra.yml up -d", addr)
	}
	return addr
}

// RequireSMB skips the test if the SMB test server is not reachable.
// Returns the SMB server address for use in test configuration.
func RequireSMB(t *testing.T) string {
	t.Helper()
	addr := smbAddr()
	if !isReachable(addr) {
		t.Skipf("SMB test server not reachable at %s — start test infrastructure with: podman-compose -f docker-compose.test-infra.yml up -d", addr)
	}
	return addr
}

// RequireWebDAV skips the test if the WebDAV test server is not reachable.
// Returns the WebDAV server address for use in test configuration.
func RequireWebDAV(t *testing.T) string {
	t.Helper()
	addr := webdavAddr()
	if !isReachable(addr) {
		t.Skipf("WebDAV test server not reachable at %s — start test infrastructure with: podman-compose -f docker-compose.test-infra.yml up -d", addr)
	}
	return addr
}

// RequireNFS skips the test if the NFS test server is not reachable.
// Returns the NFS server address for use in test configuration.
func RequireNFS(t *testing.T) string {
	t.Helper()
	addr := nfsAddr()
	if !isReachable(addr) {
		t.Skipf("NFS test server not reachable at %s — start test infrastructure with: podman-compose -f docker-compose.test-infra.yml up -d", addr)
	}
	return addr
}

// RequireAnyInfra skips the test if none of the test infrastructure services
// are reachable. Useful for tests that can work with any available protocol.
func RequireAnyInfra(t *testing.T) InfraStatus {
	t.Helper()
	status := ProbeInfra()
	if !status.FTP && !status.SMB && !status.WebDAV && !status.NFS {
		t.Skip("No test infrastructure services reachable — start test infrastructure with: podman-compose -f docker-compose.test-infra.yml up -d")
	}
	return status
}

// RequireAllInfra skips the test if any test infrastructure service is not
// reachable. Useful for comprehensive protocol comparison tests.
func RequireAllInfra(t *testing.T) InfraStatus {
	t.Helper()
	status := ProbeInfra()
	missing := []string{}
	if !status.FTP {
		missing = append(missing, "FTP")
	}
	if !status.SMB {
		missing = append(missing, "SMB")
	}
	if !status.WebDAV {
		missing = append(missing, "WebDAV")
	}
	if !status.NFS {
		missing = append(missing, "NFS")
	}
	if len(missing) > 0 {
		t.Skipf("Test infrastructure services not reachable: %v — start test infrastructure with: podman-compose -f docker-compose.test-infra.yml up -d", missing)
	}
	return status
}

// ProbeInfra checks which test infrastructure services are currently reachable.
// Does not call t.Skip; returns the status for the caller to decide.
func ProbeInfra() InfraStatus {
	return InfraStatus{
		FTP:    isReachable(ftpAddr()),
		SMB:    isReachable(smbAddr()),
		WebDAV: isReachable(webdavAddr()),
		NFS:    isReachable(nfsAddr()),
	}
}

// String returns a human-readable summary of the infrastructure status.
func (s InfraStatus) String() string {
	status := func(ok bool) string {
		if ok {
			return "UP"
		}
		return "DOWN"
	}
	return fmt.Sprintf("FTP=%s SMB=%s WebDAV=%s NFS=%s",
		status(s.FTP), status(s.SMB), status(s.WebDAV), status(s.NFS))
}

// AvailableCount returns how many infrastructure services are reachable.
func (s InfraStatus) AvailableCount() int {
	count := 0
	if s.FTP {
		count++
	}
	if s.SMB {
		count++
	}
	if s.WebDAV {
		count++
	}
	if s.NFS {
		count++
	}
	return count
}

// ---------------------------------------------------------------------------
// Address resolution — env var overrides for non-default setups
// ---------------------------------------------------------------------------

func ftpAddr() string {
	return envOrDefault("FTP_TEST_SERVER", DefaultFTPAddr)
}

func smbAddr() string {
	return envOrDefault("SMB_TEST_SERVER", DefaultSMBAddr)
}

func webdavAddr() string {
	return envOrDefault("WEBDAV_TEST_SERVER", DefaultWebDAVAddr)
}

func nfsAddr() string {
	return envOrDefault("NFS_TEST_SERVER", DefaultNFSAddr)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// isReachable attempts a TCP connection to the given address with a short
// timeout. Returns true if the connection succeeds, false otherwise.
// This is a lightweight probe — it does not perform protocol handshakes.
func isReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, connectionTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// envOrDefault reads an environment variable, returning the fallback if unset or empty.
func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
