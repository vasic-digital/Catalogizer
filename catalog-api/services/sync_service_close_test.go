package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncServiceClose(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	assert.NotNil(t, service)

	// Close should not panic even with no active goroutines
	assert.NotPanics(t, func() {
		service.Close()
	})
}

func TestSyncServiceClose_Idempotent(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	// Calling Close multiple times should not panic
	assert.NotPanics(t, func() {
		service.Close()
		service.Close()
	})
}
