package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogManagementService(t *testing.T) {
	service := NewLogManagementService(nil)

	assert.NotNil(t, service)
}

func TestLogManagementService_IsLogLevelIncluded(t *testing.T) {
	service := NewLogManagementService(nil)

	tests := []struct {
		name       string
		level      string
		minLevel   string
		expected   bool
	}{
		{
			name:     "debug includes debug",
			level:    "debug",
			minLevel: "debug",
			expected: true,
		},
		{
			name:     "info includes info",
			level:    "info",
			minLevel: "info",
			expected: true,
		},
		{
			name:     "warn includes warn",
			level:    "warn",
			minLevel: "warn",
			expected: true,
		},
		{
			name:     "error includes error",
			level:    "error",
			minLevel: "error",
			expected: true,
		},
		{
			name:     "debug excluded by info minimum",
			level:    "debug",
			minLevel: "info",
			expected: false,
		},
		{
			name:     "info excluded by warn minimum",
			level:    "info",
			minLevel: "warn",
			expected: false,
		},
		{
			name:     "error included by debug minimum",
			level:    "error",
			minLevel: "debug",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isLogLevelIncluded(tt.level, tt.minLevel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogManagementService_ExtractErrorPattern(t *testing.T) {
	service := NewLogManagementService(nil)

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:    "simple error message",
			message: "connection refused: dial tcp 127.0.0.1:5432",
			expected: "connection refused",
		},
		{
			name:    "empty message",
			message: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractErrorPattern(tt.message)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLogManagementService_GenerateShareToken(t *testing.T) {
	service := NewLogManagementService(nil)

	token1 := service.generateShareToken()
	token2 := service.generateShareToken()

	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2)
}

func TestLogManagementService_MatchesFilters(t *testing.T) {
	service := NewLogManagementService(nil)

	tests := []struct {
		name     string
		message  string
		filters  []string
		expected bool
	}{
		{
			name:     "matching filter",
			message:  "error connecting to database",
			filters:  []string{"database"},
			expected: true,
		},
		{
			name:     "no matching filter",
			message:  "server started successfully",
			filters:  []string{"database", "error"},
			expected: false,
		},
		{
			name:     "empty filters",
			message:  "any message",
			filters:  []string{},
			expected: true,
		},
		{
			name:     "empty message",
			message:  "",
			filters:  []string{"test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.matchesFilters(tt.message, tt.filters)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogManagementService_GenerateInsights(t *testing.T) {
	service := NewLogManagementService(nil)

	tests := []struct {
		name       string
		errorCount int
		warnCount  int
	}{
		{
			name:       "no errors or warnings",
			errorCount: 0,
			warnCount:  0,
		},
		{
			name:       "some errors",
			errorCount: 10,
			warnCount:  5,
		},
		{
			name:       "many errors",
			errorCount: 100,
			warnCount:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insights := service.generateInsights(tt.errorCount, tt.warnCount)
			assert.NotNil(t, insights)
		})
	}
}
