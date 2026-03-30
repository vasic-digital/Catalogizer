package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorReportingServiceClose(t *testing.T) {
	service := NewErrorReportingService(nil, nil)
	assert.NotNil(t, service)

	// Close should not panic even with no in-flight goroutines
	assert.NotPanics(t, func() {
		service.Close()
	})
}

func TestErrorReportingServiceClose_Idempotent(t *testing.T) {
	service := NewErrorReportingService(nil, nil)

	// Calling Close multiple times should not panic
	assert.NotPanics(t, func() {
		service.Close()
		service.Close()
	})
}
