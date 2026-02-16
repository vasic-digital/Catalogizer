package services

import (
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
)

func TestNewReportingService(t *testing.T) {
	service := NewReportingService(nil, nil)

	assert.NotNil(t, service)
}

func TestReportingService_CalculateSystemHealth(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name          string
		totalUsers    int
		activeUsers   int
		mediaAccesses int
	}{
		{
			name:          "healthy system with active users",
			totalUsers:    100,
			activeUsers:   80,
			mediaAccesses: 5000,
		},
		{
			name:          "system with low activity",
			totalUsers:    100,
			activeUsers:   5,
			mediaAccesses: 10,
		},
		{
			name:          "empty system",
			totalUsers:    0,
			activeUsers:   0,
			mediaAccesses: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := service.calculateSystemHealth(tt.totalUsers, tt.activeUsers, tt.mediaAccesses)
			assert.NotEmpty(t, health.Status)
		})
	}
}

func TestReportingService_ExtractDateRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid date range",
			params: map[string]interface{}{
				"start_date": "2025-01-01",
				"end_date":   "2025-01-31",
			},
			wantErr: false,
		},
		{
			name:    "missing start_date",
			params:  map[string]interface{}{"end_date": "2025-01-31"},
			wantErr: true,
		},
		{
			name:    "missing end_date",
			params:  map[string]interface{}{"start_date": "2025-01-01"},
			wantErr: true,
		},
		{
			name:    "empty params",
			params:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := service.extractDateRange(tt.params)
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
		logs     []models.MediaAccessLog
		expected int
	}{
		{
			name:     "no logs",
			logs:     []models.MediaAccessLog{},
			expected: 0,
		},
		{
			name: "unique users",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Now()},
				{UserID: 2, MediaID: 1, AccessTime: time.Now()},
				{UserID: 3, MediaID: 1, AccessTime: time.Now()},
			},
			expected: 3,
		},
		{
			name: "duplicate users",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Now()},
				{UserID: 2, MediaID: 1, AccessTime: time.Now()},
				{UserID: 1, MediaID: 2, AccessTime: time.Now()},
				{UserID: 3, MediaID: 1, AccessTime: time.Now()},
				{UserID: 2, MediaID: 3, AccessTime: time.Now()},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := service.countUniqueUsers(tt.logs)
			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestReportingService_AnalyzeTimeDistribution(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name string
		logs []models.MediaAccessLog
	}{
		{
			name: "empty logs",
			logs: []models.MediaAccessLog{},
		},
		{
			name: "various access times",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)},
				{UserID: 1, MediaID: 2, AccessTime: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
				{UserID: 1, MediaID: 3, AccessTime: time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distribution := service.analyzeTimeDistribution(tt.logs)
			assert.NotNil(t, distribution)
		})
	}
}

func TestReportingService_GetMostActiveHour(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name string
		logs []models.MediaAccessLog
	}{
		{
			name: "empty logs",
			logs: []models.MediaAccessLog{},
		},
		{
			name: "peak at 14:00",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)},
				{UserID: 1, MediaID: 2, AccessTime: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
				{UserID: 2, MediaID: 3, AccessTime: time.Date(2025, 1, 1, 14, 30, 0, 0, time.UTC)},
				{UserID: 3, MediaID: 4, AccessTime: time.Date(2025, 1, 1, 14, 45, 0, 0, time.UTC)},
				{UserID: 1, MediaID: 5, AccessTime: time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hour := service.getMostActiveHour(tt.logs)
			assert.GreaterOrEqual(t, hour, 0)
			assert.LessOrEqual(t, hour, 23)
		})
	}
}

func TestReportingService_FormatReport(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name       string
		format     string
		reportType string
		data       interface{}
		wantErr    bool
	}{
		{
			name:       "json format",
			format:     "json",
			reportType: "generic",
			data:       map[string]interface{}{"key": "value"},
			wantErr:    false,
		},
		{
			name:       "unsupported format",
			format:     "xml",
			reportType: "generic",
			data:       map[string]interface{}{"key": "value"},
			wantErr:    true,
		},
		{
			name:       "empty data json",
			format:     "json",
			reportType: "generic",
			data:       map[string]interface{}{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.formatReport(tt.data, tt.format, tt.reportType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
