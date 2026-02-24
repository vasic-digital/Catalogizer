package tests

import (
	"fmt"
	"io"
	"log"
	"testing"

	"go.uber.org/goleak"
)

// TestMain is the entry point for all tests with goroutine leak detection.
func TestMain(m *testing.M) {
	// Setup
	fmt.Println("Starting Catalogizer v3.0 Test Suite...")

	// Disable logging during tests for cleaner output
	log.SetOutput(io.Discard)

	// VerifyTestMain runs m.Run() internally and calls os.Exit with the result.
	// It also checks for leaked goroutines after all tests complete.
	goleak.VerifyTestMain(m,
		// Known goroutines from third-party libraries
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
		goleak.IgnoreTopFunction("database/sql.(*DB).connectionOpener"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	)
}
