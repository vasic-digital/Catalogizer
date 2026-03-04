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

// ---------------------------------------------------------------------------
// GenerateReport
// ---------------------------------------------------------------------------

func TestReportingService_GenerateReport_UnsupportedType(t *testing.T) {
	service := NewReportingService(nil, nil)

	report, err := service.GenerateReport("unsupported_type", "json", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported report type")
	assert.Nil(t, report)
}

func TestReportingService_GenerateReport_UserAnalytics_MissingUserID(t *testing.T) {
	service := NewReportingService(nil, nil)

	report, err := service.GenerateReport("user_analytics", "json", map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-01-31",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id parameter required")
	assert.Nil(t, report)
}

func TestReportingService_GenerateReport_UserAnalytics_MissingDateRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	report, err := service.GenerateReport("user_analytics", "json", map[string]interface{}{
		"user_id": 1,
	})
	assert.Error(t, err)
	assert.Nil(t, report)
}

func TestReportingService_GenerateReport_SystemOverview_MissingDateRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	report, err := service.GenerateReport("system_overview", "json", map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, report)
}

func TestReportingService_GenerateReport_MediaAnalytics_MissingDateRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	report, err := service.GenerateReport("media_analytics", "json", map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, report)
}

func TestReportingService_GenerateReport_SecurityAudit_MissingDateRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	report, err := service.GenerateReport("security_audit", "json", map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, report)
}

func TestReportingService_GenerateReport_PerformanceMetrics_MissingDateRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	report, err := service.GenerateReport("performance_metrics", "json", map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, report)
}

// ---------------------------------------------------------------------------
// formatReport
// ---------------------------------------------------------------------------

func TestReportingService_FormatReport_UnsupportedFormat(t *testing.T) {
	service := NewReportingService(nil, nil)

	content, err := service.formatReport(map[string]string{"test": "data"}, "invalid_format", "test_report")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
	assert.Nil(t, content)
}

// ---------------------------------------------------------------------------
// calculateSystemHealth
// ---------------------------------------------------------------------------

func TestReportingService_CalculateSystemHealth_AllScenarios(t *testing.T) {
	service := NewReportingService(nil, nil)

	tests := []struct {
		name          string
		totalUsers    int
		activeUsers   int
		mediaAccesses int
		expectHealthy bool
	}{
		{
			name:          "healthy system",
			totalUsers:    100,
			activeUsers:   80,
			mediaAccesses: 5000,
			expectHealthy: true,
		},
		{
			name:          "warning system - low activity",
			totalUsers:    100,
			activeUsers:   20,
			mediaAccesses: 100,
			expectHealthy: false,
		},
		{
			name:          "critical system - no users",
			totalUsers:    0,
			activeUsers:   0,
			mediaAccesses: 0,
			expectHealthy: false,
		},
		{
			name:          "single active user",
			totalUsers:    1,
			activeUsers:   1,
			mediaAccesses: 10,
			expectHealthy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := service.calculateSystemHealth(tt.totalUsers, tt.activeUsers, tt.mediaAccesses)
			assert.NotNil(t, health)
			assert.NotEmpty(t, health.Status)
		})
	}
}

// ---------------------------------------------------------------------------
// extractDateRange additional tests
// ---------------------------------------------------------------------------

func TestReportingService_ExtractDateRange_ValidRange(t *testing.T) {
	service := NewReportingService(nil, nil)

	start, end, err := service.extractDateRange(map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-01-31",
	})
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)
	assert.True(t, start.Before(end) || start.Equal(end))
}

// ---------------------------------------------------------------------------
// Wrapper function tests
// ---------------------------------------------------------------------------

func TestReportingService_AnalyzeUserTimePatterns(t *testing.T) {
	service := NewReportingService(nil, nil)

	logs := []models.MediaAccessLog{
		{AccessTime: time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)},
		{AccessTime: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
	}

	result := service.analyzeUserTimePatterns(logs)
	assert.NotNil(t, result)

	hourly, ok := result["hourly"].(map[int]int)
	assert.True(t, ok)
	assert.Equal(t, 1, hourly[9])
	assert.Equal(t, 1, hourly[14])
}

func TestReportingService_AnalyzeAccessPatterns(t *testing.T) {
	service := NewReportingService(nil, nil)

	logs := []models.MediaAccessLog{
		{AccessTime: time.Date(2025, 1, 6, 9, 0, 0, 0, time.UTC)}, // Monday
	}

	result := service.analyzeAccessPatterns(logs)
	assert.NotNil(t, result)

	hourly, ok := result["hourly"].(map[int]int)
	assert.True(t, ok)
	assert.Equal(t, 1, hourly[9])
}

func TestReportingService_AnalyzeGeographicDistribution(t *testing.T) {
	service := NewReportingService(nil, nil)

	logs := []models.MediaAccessLog{
		{Location: &models.Location{Latitude: 40.71, Longitude: -74.01}},
		{Location: &models.Location{Latitude: 51.51, Longitude: -0.13}},
	}

	result := service.analyzeGeographicDistribution(logs)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result["40.71,-74.01"])
	assert.Equal(t, 1, result["51.51,-0.13"])
}

func TestReportingService_AnalyzeGeographicDistribution_Empty(t *testing.T) {
	service := NewReportingService(nil, nil)

	result := service.analyzeGeographicDistribution([]models.MediaAccessLog{})
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestReportingService_AnalyzeDeviceDistribution(t *testing.T) {
	service := NewReportingService(nil, nil)

	android := "Android"
	pixel := "Pixel 7"
	ios := "iOS"
	iphone := "iPhone 15"

	logs := []models.MediaAccessLog{
		{DeviceInfo: &models.DeviceInfo{Platform: &android, DeviceModel: &pixel}},
		{DeviceInfo: &models.DeviceInfo{Platform: &ios, DeviceModel: &iphone}},
	}

	result := service.analyzeDeviceDistribution(logs)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result["Android Pixel 7"])
	assert.Equal(t, 1, result["iOS iPhone 15"])
}

func TestReportingService_AnalyzeDeviceDistribution_Empty(t *testing.T) {
	service := NewReportingService(nil, nil)

	result := service.analyzeDeviceDistribution([]models.MediaAccessLog{})
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

// ---------------------------------------------------------------------------
// formatAsMarkdown tests
// ---------------------------------------------------------------------------

func TestReportingService_FormatAsMarkdown_UserAnalytics(t *testing.T) {
	service := NewReportingService(nil, nil)

	displayName := "Test User"
	report := &models.UserAnalyticsReport{
		User: &models.User{
			ID:          1,
			Username:    "testuser",
			DisplayName: &displayName,
			Email:       "test@example.com",
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		TotalMediaAccesses: 150,
		TotalEvents:        42,
	}

	content, err := service.formatAsMarkdown(report, "user_analytics")
	assert.NoError(t, err)
	assert.NotNil(t, content)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# User Analytics Report")
	assert.Contains(t, contentStr, "Test User")
	assert.Contains(t, contentStr, "testuser")
	assert.Contains(t, contentStr, "Total Media Accesses: 150")
	assert.Contains(t, contentStr, "Total Events: 42")
	assert.Contains(t, contentStr, "2025-01-01")
}

func TestReportingService_FormatAsMarkdown_UserAnalytics_NilDisplayName(t *testing.T) {
	service := NewReportingService(nil, nil)

	report := &models.UserAnalyticsReport{
		User: &models.User{
			ID:          1,
			Username:    "testuser",
			DisplayName: nil,
			Email:       "test@example.com",
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		TotalMediaAccesses: 0,
		TotalEvents:        0,
	}

	content, err := service.formatAsMarkdown(report, "user_analytics")
	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Contains(t, string(content), "# User Analytics Report")
}

func TestReportingService_FormatAsMarkdown_SystemOverview(t *testing.T) {
	service := NewReportingService(nil, nil)

	report := &models.SystemOverviewReport{
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		TotalUsers:         100,
		ActiveUsers:        80,
		TotalMediaAccesses: 5000,
		TotalEvents:        200,
	}

	content, err := service.formatAsMarkdown(report, "system_overview")
	assert.NoError(t, err)
	assert.NotNil(t, content)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# System Overview Report")
	assert.Contains(t, contentStr, "Total Users: 100")
	assert.Contains(t, contentStr, "Active Users: 80")
	assert.Contains(t, contentStr, "Total Media Accesses: 5000")
	assert.Contains(t, contentStr, "Total Events: 200")
}

func TestReportingService_FormatAsMarkdown_DefaultType(t *testing.T) {
	service := NewReportingService(nil, nil)

	data := map[string]string{"key": "value"}

	content, err := service.formatAsMarkdown(data, "unknown_type")
	assert.NoError(t, err)
	assert.NotNil(t, content)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# unknown_type Report")
	assert.Contains(t, contentStr, "```json")
}

// ---------------------------------------------------------------------------
// formatAsHTML tests
// ---------------------------------------------------------------------------

func TestReportingService_FormatAsHTML_UserAnalytics(t *testing.T) {
	service := NewReportingService(nil, nil)

	displayName := "Test User"
	report := &models.UserAnalyticsReport{
		User: &models.User{
			ID:          1,
			Username:    "testuser",
			DisplayName: &displayName,
			Email:       "test@example.com",
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		TotalMediaAccesses: 150,
		TotalEvents:        42,
	}

	content, err := service.formatAsHTML(report, "user_analytics")
	assert.NoError(t, err)
	assert.NotNil(t, content)

	contentStr := string(content)
	assert.Contains(t, contentStr, "<!DOCTYPE html>")
	assert.Contains(t, contentStr, "testuser")
	assert.Contains(t, contentStr, "Test User")
	assert.Contains(t, contentStr, "test@example.com")
	assert.Contains(t, contentStr, "Total Media Accesses: 150")
	assert.Contains(t, contentStr, "Total Events: 42")
}

func TestReportingService_FormatAsHTML_UserAnalytics_NilDisplayName(t *testing.T) {
	service := NewReportingService(nil, nil)

	report := &models.UserAnalyticsReport{
		User: &models.User{
			ID:          1,
			Username:    "testuser",
			DisplayName: nil,
			Email:       "test@example.com",
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		TotalMediaAccesses: 0,
		TotalEvents:        0,
	}

	content, err := service.formatAsHTML(report, "user_analytics")
	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Contains(t, string(content), "<!DOCTYPE html>")
}

func TestReportingService_FormatAsHTML_SystemOverview(t *testing.T) {
	service := NewReportingService(nil, nil)

	report := &models.SystemOverviewReport{
		StartDate:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		TotalUsers:         100,
		ActiveUsers:        80,
		TotalMediaAccesses: 5000,
		TotalEvents:        200,
	}

	content, err := service.formatAsHTML(report, "system_overview")
	assert.NoError(t, err)
	assert.NotNil(t, content)

	contentStr := string(content)
	assert.Contains(t, contentStr, "<!DOCTYPE html>")
	assert.Contains(t, contentStr, "Total Users: 100")
	assert.Contains(t, contentStr, "Active Users: 80")
	assert.Contains(t, contentStr, "Total Media Accesses: 5000")
	assert.Contains(t, contentStr, "Total Events: 200")
}

func TestReportingService_FormatAsHTML_DefaultType(t *testing.T) {
	service := NewReportingService(nil, nil)

	data := map[string]string{"key": "value"}

	content, err := service.formatAsHTML(data, "unknown_type")
	assert.NoError(t, err)
	assert.NotNil(t, content)

	contentStr := string(content)
	assert.Contains(t, contentStr, "<!DOCTYPE html>")
	assert.Contains(t, contentStr, "<pre>")
}

// ---------------------------------------------------------------------------
// formatReport with markdown and html
// ---------------------------------------------------------------------------

func TestReportingService_FormatReport_Markdown(t *testing.T) {
	service := NewReportingService(nil, nil)

	data := map[string]string{"test": "data"}
	content, err := service.formatReport(data, "markdown", "generic_report")
	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Contains(t, string(content), "# generic_report Report")
}

func TestReportingService_FormatReport_HTML(t *testing.T) {
	service := NewReportingService(nil, nil)

	data := map[string]string{"test": "data"}
	content, err := service.formatReport(data, "html", "generic_report")
	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Contains(t, string(content), "<!DOCTYPE html>")
}

// ---------------------------------------------------------------------------
// Additional helper method tests
// ---------------------------------------------------------------------------

func TestReportingService_GenerateActivitySummary_SingleActivity(t *testing.T) {
	service := NewReportingService(nil, nil)

	activities := []models.UserActivitySummary{
		{TotalAccesses: 50},
	}

	summary := service.generateActivitySummary(activities)
	assert.Equal(t, 1, summary.TotalUsers)
	assert.Equal(t, 50, summary.TotalAccesses)
	assert.Equal(t, 50.0, summary.AverageAccesses)
}

func TestReportingService_GetMostActiveHour_SingleAccess(t *testing.T) {
	service := NewReportingService(nil, nil)

	logs := []models.MediaAccessLog{
		{AccessTime: time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC)},
	}

	hour := service.getMostActiveHour(logs)
	assert.Equal(t, 15, hour)
}

func TestReportingService_AnalyzeTimeDistribution_AllSlots(t *testing.T) {
	service := NewReportingService(nil, nil)

	logs := []models.MediaAccessLog{
		{AccessTime: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)},  // morning
		{AccessTime: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)}, // afternoon
		{AccessTime: time.Date(2025, 1, 1, 19, 0, 0, 0, time.UTC)}, // evening
		{AccessTime: time.Date(2025, 1, 1, 2, 0, 0, 0, time.UTC)},  // night
	}

	dist := service.analyzeTimeDistribution(logs)
	assert.Equal(t, 1, dist["morning"])
	assert.Equal(t, 1, dist["afternoon"])
	assert.Equal(t, 1, dist["evening"])
	assert.Equal(t, 1, dist["night"])
}

// ===========================================================================
// Additional formatReport tests
// ===========================================================================

func TestReportingService_FormatReport_JSON_Complex(t *testing.T) {
	service := NewReportingService(nil, nil)

	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"key": "value",
		},
		"array": []int{1, 2, 3},
	}

	result, err := service.formatReport(data, "json", "default")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, string(result), "nested")
	assert.Contains(t, string(result), "key")
}

func TestReportingService_FormatReport_PDF_ReturnsError(t *testing.T) {
	service := NewReportingService(nil, nil)

	data := map[string]interface{}{"test": "data"}

	// PDF generation fails without unipdf license
	_, err := service.formatReport(data, "pdf", "default")
	assert.Error(t, err)
}
