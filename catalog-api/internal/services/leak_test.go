package services

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	// Relax SSRF guard for the duration of this test package so
	// in-process httptest servers (always 127.0.0.1) stay reachable.
	// Production binaries never set this flag. SSRF tests themselves
	// revert the flag per-test via withStrictSSRFGuard.
	testAllowPrivateNetworks = true

	goleak.VerifyTestMain(m,
		// Known goroutines from third-party libraries
		goleak.IgnoreTopFunction("github.com/gin-gonic/gin.(*Engine).handleHTTPRequest"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
		goleak.IgnoreTopFunction("database/sql.(*DB).connectionOpener"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	)
}
