// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"os"
	"testing"

	"catalogizer/internal/services"
)

// TestMain relaxes the SSRF guard for the entire integration test binary so
// that in-process httptest servers (always 127.0.0.1) remain reachable.
// Integration tests stand up full API servers, mock TMDB servers, and other
// local HTTP endpoints — all of which bind to loopback. The production binary
// never sets this flag; the guard stays strict in production.
func TestMain(m *testing.M) {
	restore := services.SetTestAllowPrivateNetworks(true)
	code := m.Run()
	restore()
	os.Exit(code)
}
