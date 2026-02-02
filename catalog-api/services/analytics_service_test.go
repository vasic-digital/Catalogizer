package services

import (
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
)

func newTestAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

func TestAnalyticsService_CountUniqueMedia(t *testing.T) {
	svc := newTestAnalyticsService()

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
			name:     "nil logs",
			logs:     nil,
			expected: 0,
		},
		{
			name: "single media accessed once",
			logs: []models.MediaAccessLog{
				{MediaID: 1},
			},
			expected: 1,
		},
		{
			name: "single media accessed multiple times",
			logs: []models.MediaAccessLog{
				{MediaID: 1},
				{MediaID: 1},
				{MediaID: 1},
			},
			expected: 1,
		},
		{
			name: "multiple unique media",
			logs: []models.MediaAccessLog{
				{MediaID: 1},
				{MediaID: 2},
				{MediaID: 3},
			},
			expected: 3,
		},
		{
			name: "mixed unique and duplicate media",
			logs: []models.MediaAccessLog{
				{MediaID: 1},
				{MediaID: 2},
				{MediaID: 1},
				{MediaID: 3},
				{MediaID: 2},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.countUniqueMedia(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_CountUniqueUsers(t *testing.T) {
	svc := newTestAnalyticsService()

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
			name: "single user",
			logs: []models.MediaAccessLog{
				{UserID: 1},
				{UserID: 1},
			},
			expected: 1,
		},
		{
			name: "multiple unique users",
			logs: []models.MediaAccessLog{
				{UserID: 1},
				{UserID: 2},
				{UserID: 3},
				{UserID: 1},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.countUniqueUsers(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_CalculateTotalPlaybackTime(t *testing.T) {
	svc := newTestAnalyticsService()

	dur1 := 10 * time.Minute
	dur2 := 20 * time.Minute
	dur3 := 30 * time.Second

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected time.Duration
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: 0,
		},
		{
			name: "all nil durations",
			logs: []models.MediaAccessLog{
				{PlaybackDuration: nil},
				{PlaybackDuration: nil},
			},
			expected: 0,
		},
		{
			name: "single duration",
			logs: []models.MediaAccessLog{
				{PlaybackDuration: &dur1},
			},
			expected: 10 * time.Minute,
		},
		{
			name: "multiple durations",
			logs: []models.MediaAccessLog{
				{PlaybackDuration: &dur1},
				{PlaybackDuration: &dur2},
				{PlaybackDuration: &dur3},
			},
			expected: 30*time.Minute + 30*time.Second,
		},
		{
			name: "mixed nil and valid durations",
			logs: []models.MediaAccessLog{
				{PlaybackDuration: &dur1},
				{PlaybackDuration: nil},
				{PlaybackDuration: &dur2},
			},
			expected: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.calculateTotalPlaybackTime(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_CalculateAveragePlaybackTime(t *testing.T) {
	svc := newTestAnalyticsService()

	dur10 := 10 * time.Minute
	dur20 := 20 * time.Minute

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected time.Duration
	}{
		{
			name:     "empty logs returns zero",
			logs:     []models.MediaAccessLog{},
			expected: 0,
		},
		{
			name: "single log",
			logs: []models.MediaAccessLog{
				{PlaybackDuration: &dur10},
			},
			expected: 10 * time.Minute,
		},
		{
			name: "two logs average",
			logs: []models.MediaAccessLog{
				{PlaybackDuration: &dur10},
				{PlaybackDuration: &dur20},
			},
			expected: 15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.calculateAveragePlaybackTime(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_FindMostAccessedMedia(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name          string
		logs          []models.MediaAccessLog
		expectedCount int
	}{
		{
			name:          "empty logs",
			logs:          []models.MediaAccessLog{},
			expectedCount: 0,
		},
		{
			name: "single media",
			logs: []models.MediaAccessLog{
				{MediaID: 1},
				{MediaID: 1},
			},
			expectedCount: 1,
		},
		{
			name: "multiple media",
			logs: []models.MediaAccessLog{
				{MediaID: 1},
				{MediaID: 2},
				{MediaID: 1},
				{MediaID: 3},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.findMostAccessedMedia(tt.logs)
			assert.Len(t, result, tt.expectedCount)

			if tt.name == "multiple media" {
				// Verify counts are correct
				countMap := make(map[int]int)
				for _, r := range result {
					countMap[r.MediaID] = r.AccessCount
				}
				assert.Equal(t, 2, countMap[1])
				assert.Equal(t, 1, countMap[2])
				assert.Equal(t, 1, countMap[3])
			}
		})
	}
}

func TestAnalyticsService_AnalyzeAccessTimes(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected map[string]int
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: map[string]int{},
		},
		{
			name: "accesses at different hours",
			logs: []models.MediaAccessLog{
				{AccessTime: time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)},
				{AccessTime: time.Date(2025, 1, 1, 9, 30, 0, 0, time.UTC)},
				{AccessTime: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
			},
			expected: map[string]int{
				"09": 2,
				"14": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.analyzeAccessTimes(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_AnalyzePopularTimeRanges(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected map[string]int
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: map[string]int{},
		},
		{
			name: "morning access",
			logs: []models.MediaAccessLog{
				{AccessTime: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)},
				{AccessTime: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
			},
			expected: map[string]int{
				"morning": 2,
			},
		},
		{
			name: "mixed time ranges",
			logs: []models.MediaAccessLog{
				{AccessTime: time.Date(2025, 1, 1, 7, 0, 0, 0, time.UTC)},  // morning
				{AccessTime: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)}, // afternoon
				{AccessTime: time.Date(2025, 1, 1, 19, 0, 0, 0, time.UTC)}, // evening
				{AccessTime: time.Date(2025, 1, 1, 23, 0, 0, 0, time.UTC)}, // night
				{AccessTime: time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC)},  // night
			},
			expected: map[string]int{
				"morning":   1,
				"afternoon": 1,
				"evening":   1,
				"night":     2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.analyzePopularTimeRanges(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_FilterLogsByDate(t *testing.T) {
	svc := newTestAnalyticsService()

	startDate := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		logs          []models.MediaAccessLog
		expectedCount int
	}{
		{
			name:          "empty logs",
			logs:          []models.MediaAccessLog{},
			expectedCount: 0,
		},
		{
			name: "all within range",
			logs: []models.MediaAccessLog{
				{AccessTime: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)},
				{AccessTime: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
			},
			expectedCount: 2,
		},
		{
			name: "none within range",
			logs: []models.MediaAccessLog{
				{AccessTime: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)},
				{AccessTime: time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC)},
			},
			expectedCount: 0,
		},
		{
			name: "mixed - some within range",
			logs: []models.MediaAccessLog{
				{AccessTime: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)},
				{AccessTime: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
				{AccessTime: time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC)},
			},
			expectedCount: 1,
		},
		{
			name: "boundary dates are excluded",
			logs: []models.MediaAccessLog{
				{AccessTime: startDate},
				{AccessTime: endDate},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.filterLogsByDate(tt.logs, startDate, endDate)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestAnalyticsService_CalculateUserRetention(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected float64
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: 0.0,
		},
		{
			name: "single log",
			logs: []models.MediaAccessLog{
				{UserID: 1, AccessTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
			expected: 0.0,
		},
		{
			name: "single user with multiple accesses over 10 days",
			logs: []models.MediaAccessLog{
				{UserID: 1, AccessTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{UserID: 1, AccessTime: time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)},
			},
			expected: 10.0,
		},
		{
			name: "two users with different retention",
			logs: []models.MediaAccessLog{
				{UserID: 1, AccessTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{UserID: 1, AccessTime: time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)},
				{UserID: 2, AccessTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{UserID: 2, AccessTime: time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)},
			},
			expected: 15.0, // (10 + 20) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.calculateUserRetention(tt.logs)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestAnalyticsService_AnalyzeDeviceUsage(t *testing.T) {
	svc := newTestAnalyticsService()

	android := "Android"
	pixel := "Pixel 7"
	ios := "iOS"
	iphone := "iPhone 15"

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected map[string]int
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: map[string]int{},
		},
		{
			name: "logs without device info",
			logs: []models.MediaAccessLog{
				{DeviceInfo: nil},
			},
			expected: map[string]int{},
		},
		{
			name: "logs with device info",
			logs: []models.MediaAccessLog{
				{DeviceInfo: &models.DeviceInfo{Platform: &android, DeviceModel: &pixel}},
				{DeviceInfo: &models.DeviceInfo{Platform: &android, DeviceModel: &pixel}},
				{DeviceInfo: &models.DeviceInfo{Platform: &ios, DeviceModel: &iphone}},
			},
			expected: map[string]int{
				"Android Pixel 7":  2,
				"iOS iPhone 15":    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.analyzeDeviceUsage(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_CalculateSystemHealthScore(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name      string
		analytics *models.SystemAnalytics
		expected  float64
	}{
		{
			name: "no users",
			analytics: &models.SystemAnalytics{
				TotalUsers:         0,
				ActiveUsers:        0,
				TotalMediaAccesses: 0,
				TotalEvents:        0,
			},
			expected: 0.0,
		},
		{
			name: "all users active with media and events",
			analytics: &models.SystemAnalytics{
				TotalUsers:             100,
				ActiveUsers:            100,
				TotalMediaAccesses:     500,
				TotalEvents:            200,
				AverageSessionDuration: 10 * time.Minute,
			},
			expected: 100.0, // 40 (active ratio) + 30 (accesses) + 20 (events) + 10 (session)
		},
		{
			name: "half users active, no events",
			analytics: &models.SystemAnalytics{
				TotalUsers:             100,
				ActiveUsers:            50,
				TotalMediaAccesses:     100,
				TotalEvents:            0,
				AverageSessionDuration: 1 * time.Minute,
			},
			expected: 50.0, // 20 (half active) + 30 (accesses) + 0 + 0 (session < 5min)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.calculateSystemHealthScore(tt.analytics)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestAnalyticsService_CalculateGrowthRate(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name       string
		growthData []models.UserGrowthPoint
		expected   float64
	}{
		{
			name:       "empty data",
			growthData: []models.UserGrowthPoint{},
			expected:   0.0,
		},
		{
			name: "single point",
			growthData: []models.UserGrowthPoint{
				{UserCount: 10},
			},
			expected: 0.0,
		},
		{
			name: "zero initial count",
			growthData: []models.UserGrowthPoint{
				{UserCount: 0},
				{UserCount: 50},
			},
			expected: 0.0,
		},
		{
			name: "100% growth",
			growthData: []models.UserGrowthPoint{
				{UserCount: 50},
				{UserCount: 100},
			},
			expected: 100.0,
		},
		{
			name: "50% growth",
			growthData: []models.UserGrowthPoint{
				{UserCount: 100},
				{UserCount: 120},
				{UserCount: 150},
			},
			expected: 50.0,
		},
		{
			name: "negative growth",
			growthData: []models.UserGrowthPoint{
				{UserCount: 100},
				{UserCount: 80},
			},
			expected: -20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.calculateGrowthRate(tt.growthData)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestAnalyticsService_CalculateEngagementLevel(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name      string
		analytics *models.SystemAnalytics
		expected  string
	}{
		{
			name: "no users",
			analytics: &models.SystemAnalytics{
				TotalUsers:         0,
				TotalMediaAccesses: 0,
			},
			expected: "low",
		},
		{
			name: "low engagement",
			analytics: &models.SystemAnalytics{
				TotalUsers:         100,
				TotalMediaAccesses: 500, // 5 per user
			},
			expected: "low",
		},
		{
			name: "medium engagement",
			analytics: &models.SystemAnalytics{
				TotalUsers:         100,
				TotalMediaAccesses: 3000, // 30 per user
			},
			expected: "medium",
		},
		{
			name: "high engagement",
			analytics: &models.SystemAnalytics{
				TotalUsers:         100,
				TotalMediaAccesses: 6000, // 60 per user
			},
			expected: "high",
		},
		{
			name: "boundary - exactly 20 per user",
			analytics: &models.SystemAnalytics{
				TotalUsers:         10,
				TotalMediaAccesses: 200,
			},
			expected: "medium",
		},
		{
			name: "boundary - exactly 50 per user",
			analytics: &models.SystemAnalytics{
				TotalUsers:         10,
				TotalMediaAccesses: 500,
			},
			expected: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.calculateEngagementLevel(tt.analytics)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_AnalyzeAccessPatterns(t *testing.T) {
	svc := newTestAnalyticsService()

	logs := []models.MediaAccessLog{
		{AccessTime: time.Date(2025, 1, 6, 9, 0, 0, 0, time.UTC)},  // Monday, 09:00
		{AccessTime: time.Date(2025, 1, 6, 14, 0, 0, 0, time.UTC)}, // Monday, 14:00
		{AccessTime: time.Date(2025, 1, 7, 9, 0, 0, 0, time.UTC)},  // Tuesday, 09:00
	}

	result := svc.analyzeAccessPatterns(logs)

	hourly, ok := result["hourly"].(map[string]int)
	assert.True(t, ok)
	assert.Equal(t, 2, hourly["09"])
	assert.Equal(t, 1, hourly["14"])

	daily, ok := result["daily"].(map[string]int)
	assert.True(t, ok)
	assert.Equal(t, 2, daily["Monday"])
	assert.Equal(t, 1, daily["Tuesday"])
}

func TestAnalyticsService_ExtractDateRange(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
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
			name: "missing start_date",
			params: map[string]interface{}{
				"end_date": "2025-01-31",
			},
			wantErr: true,
			errMsg:  "start_date parameter required",
		},
		{
			name: "missing end_date",
			params: map[string]interface{}{
				"start_date": "2025-01-01",
			},
			wantErr: true,
			errMsg:  "end_date parameter required",
		},
		{
			name: "invalid start_date format",
			params: map[string]interface{}{
				"start_date": "not-a-date",
				"end_date":   "2025-01-31",
			},
			wantErr: true,
			errMsg:  "invalid start_date format",
		},
		{
			name: "invalid end_date format",
			params: map[string]interface{}{
				"start_date": "2025-01-01",
				"end_date":   "not-a-date",
			},
			wantErr: true,
			errMsg:  "invalid end_date format",
		},
		{
			name:    "empty params",
			params:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate, endDate, err := svc.extractDateRange(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 2025, startDate.Year())
				assert.Equal(t, time.January, startDate.Month())
				assert.Equal(t, 1, startDate.Day())
				assert.Equal(t, 2025, endDate.Year())
				assert.Equal(t, time.January, endDate.Month())
				assert.Equal(t, 31, endDate.Day())
			}
		})
	}
}

func TestAnalyticsService_CreateReport_UnsupportedType(t *testing.T) {
	svc := newTestAnalyticsService()

	report, err := svc.CreateReport("unsupported_type", map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "unsupported report type")
}

func TestAnalyticsService_AnalyzeLocations(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name     string
		logs     []models.MediaAccessLog
		expected map[string]int
	}{
		{
			name:     "empty logs",
			logs:     []models.MediaAccessLog{},
			expected: map[string]int{},
		},
		{
			name: "no locations",
			logs: []models.MediaAccessLog{
				{Location: nil},
			},
			expected: map[string]int{},
		},
		{
			name: "with locations",
			logs: []models.MediaAccessLog{
				{Location: &models.Location{Latitude: 40.71, Longitude: -74.01}},
				{Location: &models.Location{Latitude: 40.71, Longitude: -74.01}},
				{Location: &models.Location{Latitude: 51.51, Longitude: -0.13}},
			},
			expected: map[string]int{
				"40.71,-74.01": 2,
				"51.51,-0.13":  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.analyzeLocations(tt.logs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyticsService_GetTopLocations(t *testing.T) {
	svc := newTestAnalyticsService()

	tests := []struct {
		name          string
		data          map[string]interface{}
		limit         int
		expectedCount int
	}{
		{
			name:          "empty data",
			data:          map[string]interface{}{},
			limit:         10,
			expectedCount: 0,
		},
		{
			name: "no locations key",
			data: map[string]interface{}{
				"countries": map[string]int{"US": 100},
			},
			limit:         10,
			expectedCount: 0,
		},
		{
			name: "with locations within limit",
			data: map[string]interface{}{
				"locations": []map[string]interface{}{
					{"city": "New York"},
					{"city": "London"},
				},
			},
			limit:         10,
			expectedCount: 2,
		},
		{
			name: "locations exceeding limit",
			data: map[string]interface{}{
				"locations": []map[string]interface{}{
					{"city": "New York"},
					{"city": "London"},
					{"city": "Tokyo"},
				},
			},
			limit:         2,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.getTopLocations(tt.data, tt.limit)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}
