package tests

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	// Setup
	fmt.Println("Starting Catalogizer v3.0 Test Suite...")

	// Disable logging during tests for cleaner output
	log.SetOutput(io.Discard)

	// Run tests
	code := m.Run()

	// Cleanup
	fmt.Println("Test Suite Completed")

	os.Exit(code)
}
