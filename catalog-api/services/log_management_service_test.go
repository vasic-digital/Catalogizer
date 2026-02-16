package services

import (
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
)

func TestNewLogManagementService(t *testing.T) {
	service := NewLogManagementService(nil)

	assert.NotNil(t, service)
}

func TestLogManagementService_IsLogLevelIncluded(t *testing.T) {
	service := NewLogManagementService(nil)

	tests := []struct {
		name     string
		level    string
		minLevel string
		expected bool
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
			name:     "warning includes warning",
			level:    "warning",
			minLevel: "warning",
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
			name:     "info excluded by warning minimum",
			level:    "info",
			minLevel: "warning",
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
		name    string
		message string
	}{
		{
			name:    "simple error message",
			message: "connection refused: dial tcp 127.0.0.1:5432",
		},
		{
			name:    "short message",
			message: "timeout",
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
}

func TestLogManagementService_MatchesFilters(t *testing.T) {
	service := NewLogManagementService(nil)

	tests := []struct {
		name     string
		entry    *models.LogEntry
		filters  map[string]interface{}
		expected bool
	}{
		{
			name: "matching message_contains filter",
			entry: &models.LogEntry{
				Message:   "error connecting to database",
				Level:     "error",
				Component: "api",
				Timestamp: time.Now(),
			},
			filters:  map[string]interface{}{"message_contains": "database"},
			expected: true,
		},
		{
			name: "no matching message_contains filter",
			entry: &models.LogEntry{
				Message:   "server started successfully",
				Level:     "info",
				Component: "api",
				Timestamp: time.Now(),
			},
			filters:  map[string]interface{}{"message_contains": "database"},
			expected: false,
		},
		{
			name: "empty filters match all",
			entry: &models.LogEntry{
				Message:   "any message",
				Level:     "info",
				Component: "api",
				Timestamp: time.Now(),
			},
			filters:  map[string]interface{}{},
			expected: true,
		},
		{
			name: "component filter match",
			entry: &models.LogEntry{
				Message:   "something happened",
				Level:     "info",
				Component: "auth",
				Timestamp: time.Now(),
			},
			filters:  map[string]interface{}{"component": "auth"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.matchesFilters(tt.entry, tt.filters)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogManagementService_GenerateInsights(t *testing.T) {
	service := NewLogManagementService(nil)

	tests := []struct {
		name     string
		entries  []*models.LogEntry
		analysis *models.LogAnalysis
	}{
		{
			name: "some entries with analysis",
			entries: []*models.LogEntry{
				{Level: "info", Component: "api", Message: "server started"},
			},
			analysis: &models.LogAnalysis{
				TotalEntries:       1,
				EntriesByLevel:     map[string]int{"info": 1},
				EntriesByComponent: map[string]int{"api": 1},
				ErrorPatterns:      map[string]int{},
			},
		},
		{
			name: "with some errors",
			entries: []*models.LogEntry{
				{Level: "error", Component: "api", Message: "test error"},
				{Level: "info", Component: "api", Message: "test info"},
			},
			analysis: &models.LogAnalysis{
				TotalEntries:       2,
				EntriesByLevel:     map[string]int{"error": 1, "info": 1},
				EntriesByComponent: map[string]int{"api": 2},
				ErrorPatterns:      map[string]int{"test error": 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insights := service.generateInsights(tt.entries, tt.analysis)
			assert.NotNil(t, insights)
		})
	}
}
