package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogManagementServiceClose(t *testing.T) {
	service := NewLogManagementService(nil)
	assert.NotNil(t, service)

	// Close should not panic even with no in-flight goroutines
	assert.NotPanics(t, func() {
		service.Close()
	})
}

func TestLogManagementServiceClose_Idempotent(t *testing.T) {
	service := NewLogManagementService(nil)

	// Calling Close multiple times should not panic
	assert.NotPanics(t, func() {
		service.Close()
		service.Close()
	})
}

func TestLogManagementService_MaxLogEntriesCap(t *testing.T) {
	// Verify the maxLogEntries constant is set to a reasonable bound
	assert.Equal(t, 50000, maxLogEntries, "maxLogEntries should be 50000 to prevent unbounded memory growth")
}
