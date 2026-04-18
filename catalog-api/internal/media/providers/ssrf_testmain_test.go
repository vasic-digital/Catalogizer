package providers

import (
	"os"
	"testing"

	"catalogizer/internal/services"
)

// TestMain relaxes the SSRF guard in internal/services for this
// package's test binary so the many httptest servers (127.0.0.1)
// stay reachable. The flag is reverted after the suite.
// Production binaries never set it.
func TestMain(m *testing.M) {
	restore := services.SetTestAllowPrivateNetworks(true)
	code := m.Run()
	restore()
	os.Exit(code)
}
