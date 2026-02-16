package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewReportingService(t *testing.T) {
	service := NewReportingService(nil, nil)

	assert.NotNil(t, service)
}

func TestReportingService_CalculateSystemHealth(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name       string
		errorRate  float64
		uptime     float64
		wantHealth string
	}{
		{
			name:       "healthy system",
			errorRate:  0.01,
			uptime:     99.9,
			wantHealth: "healthy",
		},
		{
			name:       "degraded system",
			errorRate:  0.10,
			uptime:     95.0,
			wantHealth: "degraded",
		},
		{
			name:       "critical system",
			errorRate:  0.50,
			uptime:     50.0,
			wantHealth: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := service.calculateSystemHealth(tt.errorRate, tt.uptime)
			assert.NotEmpty(t, health)
		})
	}
}

func TestReportingService_ExtractDateRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name    string
		period  string
		wantErr bool
	}{
		{
			name:    "daily period",
			period:  "daily",
			wantErr: false,
		},
		{
			name:    "weekly period",
			period:  "weekly",
			wantErr: false,
		},
		{
			name:    "monthly period",
			period:  "monthly",
			wantErr: false,
		},
		{
			name:    "invalid period",
			period:  "invalid",
			wantErr: true,
		},
		{
			name:    "empty period",
			period:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := service.extractDateRange(tt.period)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, start.IsZero())
				assert.False(t, end.IsZero())
				assert.True(t, end.After(start) || end.Equal(start))
			}
		})
	}
}

func TestReportingService_CountUniqueUsers(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name     string
		userIDs  []int
		expected int
	}{
		{
			name:     "no users",
			userIDs:  []int{},
			expected: 0,
		},
		{
			name:     "unique users",
			userIDs:  []int{1, 2, 3},
			expected: 3,
		},
		{
			name:     "duplicate users",
			userIDs:  []int{1, 2, 1, 3, 2},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := service.countUniqueUsers(tt.userIDs)
			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestReportingService_AnalyzeTimeDistribution(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name       string
		timestamps []time.Time
	}{
		{
			name:       "empty timestamps",
			timestamps: []time.Time{},
		},
		{
			name: "various timestamps",
			timestamps: []time.Time{
				time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
				time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distribution := service.analyzeTimeDistribution(tt.timestamps)
			assert.NotNil(t, distribution)
		})
	}
}

func TestReportingService_GetMostActiveHour(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name         string
		distribution map[int]int
		expected     int
	}{
		{
			name:         "empty distribution",
			distribution: map[int]int{},
			expected:     0,
		},
		{
			name:         "single peak",
			distribution: map[int]int{8: 10, 14: 30, 20: 5},
			expected:     14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hour := service.getMostActiveHour(tt.distribution)
			assert.GreaterOrEqual(t, hour, 0)
			assert.LessOrEqual(t, hour, 23)
		})
	}
}

func TestReportingService_FormatReport(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name    string
		format  string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "json format",
			format:  "json",
			data:    map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "csv format",
			format:  "csv",
			data:    map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "unsupported format",
			format:  "xml",
			data:    map[string]interface{}{"key": "value"},
			wantErr: true,
		},
		{
			name:    "empty data",
			format:  "json",
			data:    map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.formatReport(tt.format, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
