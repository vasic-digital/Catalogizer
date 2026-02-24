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

func TestReportingService_FilterLogsByDateRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	startDate := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected int
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: 0,
		},
		{
			name: "all within range",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)},
				{UserID: 2, MediaID: 2, AccessTime: time.Date(2025, 1, 10, 14, 0, 0, 0, time.UTC)},
			},
			expected: 2,
		},
		{
			name: "some outside range",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
				{UserID: 2, MediaID: 2, AccessTime: time.Date(2025, 1, 10, 14, 0, 0, 0, time.UTC)},
				{UserID: 3, MediaID: 3, AccessTime: time.Date(2025, 1, 20, 18, 0, 0, 0, time.UTC)},
			},
			expected: 1,
		},
		{
			name: "all outside range",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
				{UserID: 2, MediaID: 2, AccessTime: time.Date(2025, 1, 20, 14, 0, 0, 0, time.UTC)},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.filterLogsByDateRange(tt.logs, startDate, endDate)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestReportingService_AnalyzeUserAccessPatterns(t *testing.T) {
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
			name: "logs with various hours and days",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Date(2025, 1, 6, 8, 0, 0, 0, time.UTC)},  // Monday
				{UserID: 1, MediaID: 2, AccessTime: time.Date(2025, 1, 6, 14, 0, 0, 0, time.UTC)}, // Monday
				{UserID: 1, MediaID: 3, AccessTime: time.Date(2025, 1, 7, 20, 0, 0, 0, time.UTC)}, // Tuesday
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := service.analyzeUserAccessPatterns(tt.logs)
			assert.NotNil(t, patterns)
			assert.Contains(t, patterns, "hourly")
			assert.Contains(t, patterns, "daily")
		})
	}
}

func TestReportingService_AnalyzeUserDeviceUsage(t *testing.T) {
	service := NewReportingService(nil, nil)

	platform := "iOS"
	deviceModel := "iPhone 14"

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected int
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: 0,
		},
		{
			name: "logs with device info",
			logs: []models.MediaAccessLog{
				{
					UserID: 1, MediaID: 1,
					AccessTime: time.Now(),
					DeviceInfo: &models.DeviceInfo{Platform: &platform, DeviceModel: &deviceModel},
				},
			},
			expected: 1,
		},
		{
			name: "logs without device info",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Now(), DeviceInfo: nil},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.analyzeUserDeviceUsage(tt.logs)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestReportingService_AnalyzeUserLocations(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected int
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: 0,
		},
		{
			name: "logs with locations",
			logs: []models.MediaAccessLog{
				{
					UserID: 1, MediaID: 1, AccessTime: time.Now(),
					Location: &models.Location{Latitude: 40.71, Longitude: -74.00},
				},
				{
					UserID: 2, MediaID: 2, AccessTime: time.Now(),
					Location: &models.Location{Latitude: 51.51, Longitude: -0.13},
				},
			},
			expected: 2,
		},
		{
			name: "logs without locations",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Now(), Location: nil},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.analyzeUserLocations(tt.logs)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestReportingService_AnalyzeUserPopularContent(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected int
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: 0,
		},
		{
			name: "multiple accesses same media",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Now()},
				{UserID: 1, MediaID: 1, AccessTime: time.Now()},
				{UserID: 1, MediaID: 2, AccessTime: time.Now()},
			},
			expected: 2, // 2 unique media items
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.analyzeUserPopularContent(tt.logs)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestReportingService_GetLastActivityTime(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected time.Time
	}{
		{
			name:     "empty logs returns zero time",
			logs:     []models.MediaAccessLog{},
			expected: time.Time{},
		},
		{
			name: "single log",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Date(2025, 1, 10, 14, 0, 0, 0, time.UTC)},
			},
			expected: time.Date(2025, 1, 10, 14, 0, 0, 0, time.UTC),
		},
		{
			name: "multiple logs returns latest",
			logs: []models.MediaAccessLog{
				{UserID: 1, MediaID: 1, AccessTime: time.Date(2025, 1, 5, 10, 0, 0, 0, time.UTC)},
				{UserID: 1, MediaID: 2, AccessTime: time.Date(2025, 1, 15, 20, 0, 0, 0, time.UTC)},
				{UserID: 1, MediaID: 3, AccessTime: time.Date(2025, 1, 10, 14, 0, 0, 0, time.UTC)},
			},
			expected: time.Date(2025, 1, 15, 20, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getLastActivityTime(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReportingService_GetPreferredDevices(t *testing.T) {
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
			name: "logs with device info",
			logs: func() []models.MediaAccessLog {
				platform := "Android"
				model := "Pixel 7"
				return []models.MediaAccessLog{
					{UserID: 1, MediaID: 1, AccessTime: time.Now(), DeviceInfo: &models.DeviceInfo{Platform: &platform, DeviceModel: &model}},
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getPreferredDevices(tt.logs)
			// Result may be nil for empty device info
			if len(tt.logs) > 0 && tt.logs[0].DeviceInfo != nil {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestReportingService_GetAccessedLocations(t *testing.T) {
	service := NewReportingService(nil, nil)

	logs := []models.MediaAccessLog{
		{UserID: 1, MediaID: 1, AccessTime: time.Now(), Location: &models.Location{Latitude: 40.71, Longitude: -74.00}},
	}

	result := service.getAccessedLocations(logs)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result))
}

func TestReportingService_GenerateActivitySummary(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name       string
		activities []models.UserActivitySummary
	}{
		{
			name:       "empty activities",
			activities: []models.UserActivitySummary{},
		},
		{
			name: "multiple activities",
			activities: []models.UserActivitySummary{
				{TotalAccesses: 10},
				{TotalAccesses: 20},
				{TotalAccesses: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := service.generateActivitySummary(tt.activities)
			if len(tt.activities) == 0 {
				assert.Equal(t, 0, summary.TotalUsers)
			} else {
				assert.Equal(t, len(tt.activities), summary.TotalUsers)
				assert.Equal(t, 60, summary.TotalAccesses)
				assert.Equal(t, 20.0, summary.AverageAccesses)
			}
		})
	}
}

func TestReportingService_CalculateUsageStatistics(t *testing.T) {
	service := NewReportingService(nil, nil)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	stats := service.calculateUsageStatistics(startDate, endDate)
	assert.NotEmpty(t, stats.PeakHours)
	assert.Greater(t, stats.AverageDaily, 0)
}

func TestReportingService_CalculatePerformanceMetrics(t *testing.T) {
	service := NewReportingService(nil, nil)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	metrics := service.calculatePerformanceMetrics(startDate, endDate)
	assert.Greater(t, metrics.ResponseTime, 0.0)
	assert.Greater(t, metrics.Throughput, 0)
}

func TestReportingService_CalculateSecurityMetrics(t *testing.T) {
	service := NewReportingService(nil, nil)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	metrics := service.calculateSecurityMetrics(startDate, endDate)
	assert.NotEmpty(t, metrics.ThreatLevel)
	assert.Greater(t, metrics.SecurityScore, 0.0)
}

func TestReportingService_CalculateAverageSessionDuration(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name     string
		sessions []models.SessionData
		expected time.Duration
	}{
		{
			name:     "empty sessions",
			sessions: []models.SessionData{},
			expected: 0,
		},
		{
			name: "single session",
			sessions: []models.SessionData{
				{Duration: 10 * time.Minute},
			},
			expected: 10 * time.Minute,
		},
		{
			name: "multiple sessions",
			sessions: []models.SessionData{
				{Duration: 10 * time.Minute},
				{Duration: 20 * time.Minute},
				{Duration: 30 * time.Minute},
			},
			expected: 20 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateAverageSessionDuration(tt.sessions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReportingService_CalculateResponseTimes(t *testing.T) {
	service := NewReportingService(nil, nil)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	result := service.calculateResponseTimes(startDate, endDate)
	assert.Greater(t, result.Average, 0.0)
	assert.Greater(t, result.Max, result.Min)
	assert.Greater(t, result.P99, result.P95)
}

func TestReportingService_CalculateSystemLoad(t *testing.T) {
	service := NewReportingService(nil, nil)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	result := service.calculateSystemLoad(startDate, endDate)
	assert.Greater(t, result.CPU, 0.0)
	assert.Greater(t, result.Memory, 0.0)
	assert.Greater(t, result.Disk, 0.0)
}

func TestReportingService_CalculateErrorRates(t *testing.T) {
	service := NewReportingService(nil, nil)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	result := service.calculateErrorRates(startDate, endDate)
	assert.Greater(t, result.Total, 0.0)
}

func TestReportingService_AnalyzeUserEngagement(t *testing.T) {
	service := NewReportingService(nil, nil)

	logs := []models.MediaAccessLog{
		{UserID: 1, MediaID: 1, AccessTime: time.Now()},
	}

	engagement := service.analyzeUserEngagement(logs)
	assert.Greater(t, engagement.AverageSessionTime, 0.0)
	assert.Greater(t, engagement.ReturnRate, 0.0)
}

func TestReportingService_ExtractDateRange_InvalidFormats(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "invalid start_date format",
			params: map[string]interface{}{"start_date": "not-a-date", "end_date": "2025-01-31"},
		},
		{
			name:   "invalid end_date format",
			params: map[string]interface{}{"start_date": "2025-01-01", "end_date": "not-a-date"},
		},
		{
			name:   "non-string start_date",
			params: map[string]interface{}{"start_date": 12345, "end_date": "2025-01-31"},
		},
		{
			name:   "non-string end_date",
			params: map[string]interface{}{"start_date": "2025-01-01", "end_date": 12345},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.extractDateRange(tt.params)
			assert.Error(t, err)
		})
	}
}
