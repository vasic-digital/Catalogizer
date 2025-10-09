package tests

import (
	"testing"
	"time"

	"catalog-api/models"
)

func TestAnalyticsService_TrackEvent(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Create test user
	user := CreateTestUser(t, suite.DB.DB, 1)

	tests := []struct {
		name          string
		userID        int
		request       *models.AnalyticsEventRequest
		expectError   bool
		expectedType  string
	}{
		{
			name:   "Valid media view event",
			userID: user.ID,
			request: &models.AnalyticsEventRequest{
				EventType:  "media_view",
				EntityType: "media_item",
				EntityID:   1,
				Metadata: map[string]interface{}{
					"duration": 120,
					"quality":  "1080p",
				},
				SessionID: "session_123",
				IPAddress: "192.168.1.1",
				UserAgent: "Test Browser",
			},
			expectError:  false,
			expectedType: "media_view",
		},
		{
			name:   "Valid user login event",
			userID: user.ID,
			request: &models.AnalyticsEventRequest{
				EventType: "user_login",
				Metadata: map[string]interface{}{
					"login_method": "email",
				},
				SessionID: "session_456",
				IPAddress: "192.168.1.2",
				UserAgent: "Test Browser 2",
			},
			expectError:  false,
			expectedType: "user_login",
		},
		{
			name:   "Empty event type should fail",
			userID: user.ID,
			request: &models.AnalyticsEventRequest{
				EventType: "",
				SessionID: "session_789",
			},
			expectError: true,
		},
		{
			name:   "Invalid user ID should fail",
			userID: 99999,
			request: &models.AnalyticsEventRequest{
				EventType: "test_event",
				SessionID: "session_000",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := suite.AnalyticsService.TrackEvent(tt.userID, tt.request)

			if tt.expectError {
				AssertError(t, err, "Expected error for invalid input")
				AssertNil(t, event, "Event should be nil on error")
			} else {
				AssertNoError(t, err, "Should not error for valid input")
				AssertNotNil(t, event, "Event should not be nil")
				AssertEqual(t, tt.expectedType, event.EventType, "Event type should match")
				AssertEqual(t, tt.userID, event.UserID, "User ID should match")
			}
		})
	}
}

func TestAnalyticsService_GetEventsByUser(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Create test user
	user := CreateTestUser(t, suite.DB.DB, 1)

	// Create test events
	for i := 0; i < 5; i++ {
		_, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
			EventType:  "test_event",
			EntityType: "test_entity",
			EntityID:   i,
			SessionID:  "test_session",
		})
		AssertNoError(t, err, "Should create test events")
	}

	tests := []struct {
		name        string
		userID      int
		filters     *models.AnalyticsFilters
		expectCount int
		expectError bool
	}{
		{
			name:   "Get all events for user",
			userID: user.ID,
			filters: &models.AnalyticsFilters{
				Limit: 10,
			},
			expectCount: 5,
			expectError: false,
		},
		{
			name:   "Get events with limit",
			userID: user.ID,
			filters: &models.AnalyticsFilters{
				Limit: 3,
			},
			expectCount: 3,
			expectError: false,
		},
		{
			name:   "Get events by type",
			userID: user.ID,
			filters: &models.AnalyticsFilters{
				EventType: "test_event",
				Limit:     10,
			},
			expectCount: 5,
			expectError: false,
		},
		{
			name:   "Get events for non-existent user",
			userID: 99999,
			filters: &models.AnalyticsFilters{
				Limit: 10,
			},
			expectCount: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := suite.AnalyticsService.GetEventsByUser(tt.userID, tt.filters)

			if tt.expectError {
				AssertError(t, err, "Expected error")
			} else {
				AssertNoError(t, err, "Should not error")
				AssertEqual(t, tt.expectCount, len(events), "Event count should match")
			}
		})
	}
}

func TestAnalyticsService_GetAnalytics(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Create test user
	user := CreateTestUser(t, suite.DB.DB, 1)

	// Create test events with different types
	eventTypes := []string{"media_view", "media_download", "user_login", "user_logout"}
	for _, eventType := range eventTypes {
		for i := 0; i < 3; i++ {
			_, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
				EventType: eventType,
				SessionID: "test_session",
			})
			AssertNoError(t, err, "Should create test events")
		}
	}

	analytics, err := suite.AnalyticsService.GetAnalytics(user.ID, &models.AnalyticsFilters{})
	AssertNoError(t, err, "Should get analytics")
	AssertNotNil(t, analytics, "Analytics should not be nil")
	AssertEqual(t, 12, analytics.TotalEvents, "Total events should be 12")
	AssertEqual(t, 4, len(analytics.EventsByType), "Should have 4 event types")

	// Test each event type count
	for _, eventType := range eventTypes {
		count, exists := analytics.EventsByType[eventType]
		AssertEqual(t, true, exists, "Event type should exist in analytics")
		AssertEqual(t, 3, count, "Event type count should be 3")
	}
}

func TestAnalyticsService_GetDashboardMetrics(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Create test user
	user := CreateTestUser(t, suite.DB.DB, 1)

	// Create test events for different time periods
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	// Events for today
	for i := 0; i < 5; i++ {
		_, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
			EventType: "media_view",
			SessionID: "today_session",
		})
		AssertNoError(t, err, "Should create today's events")
	}

	metrics, err := suite.AnalyticsService.GetDashboardMetrics(user.ID)
	AssertNoError(t, err, "Should get dashboard metrics")
	AssertNotNil(t, metrics, "Metrics should not be nil")
	AssertEqual(t, user.ID, metrics.UserID, "User ID should match")
}

func TestAnalyticsService_GetRealtimeMetrics(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Create test user
	user := CreateTestUser(t, suite.DB.DB, 1)

	// Create recent events
	for i := 0; i < 3; i++ {
		_, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
			EventType: "media_view",
			SessionID: "realtime_session",
		})
		AssertNoError(t, err, "Should create realtime events")
	}

	metrics, err := suite.AnalyticsService.GetRealtimeMetrics(user.ID)
	AssertNoError(t, err, "Should get realtime metrics")
	AssertNotNil(t, metrics, "Metrics should not be nil")
}

func TestAnalyticsService_GenerateReport(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Create test user
	user := CreateTestUser(t, suite.DB.DB, 1)

	// Create test data
	for i := 0; i < 10; i++ {
		_, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
			EventType:  "media_view",
			EntityType: "media_item",
			EntityID:   i,
			SessionID:  "report_session",
		})
		AssertNoError(t, err, "Should create test events")
	}

	tests := []struct {
		name        string
		userID      int
		request     *models.ReportRequest
		expectError bool
	}{
		{
			name:   "Valid report request",
			userID: user.ID,
			request: &models.ReportRequest{
				ReportType: "user_activity",
				StartDate:  time.Now().Add(-24 * time.Hour),
				EndDate:    time.Now(),
				Format:     "json",
			},
			expectError: false,
		},
		{
			name:   "Invalid report type",
			userID: user.ID,
			request: &models.ReportRequest{
				ReportType: "invalid_type",
				StartDate:  time.Now().Add(-24 * time.Hour),
				EndDate:    time.Now(),
				Format:     "json",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := suite.AnalyticsService.GenerateReport(tt.userID, tt.request)

			if tt.expectError {
				AssertError(t, err, "Expected error for invalid input")
				AssertNil(t, report, "Report should be nil on error")
			} else {
				AssertNoError(t, err, "Should not error for valid input")
				AssertNotNil(t, report, "Report should not be nil")
				AssertEqual(t, tt.request.ReportType, report.ReportType, "Report type should match")
			}
		})
	}
}

func TestAnalyticsService_CleanupOldEvents(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Create test user
	user := CreateTestUser(t, suite.DB.DB, 1)

	// Create test events
	for i := 0; i < 5; i++ {
		_, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
			EventType: "test_event",
			SessionID: "cleanup_session",
		})
		AssertNoError(t, err, "Should create test events")
	}

	// Cleanup events older than 1 day (should not affect recent events)
	cutoff := time.Now().Add(-24 * time.Hour)
	err := suite.AnalyticsService.CleanupOldEvents(cutoff)
	AssertNoError(t, err, "Should cleanup old events")

	// Verify events still exist (they're recent)
	events, err := suite.AnalyticsService.GetEventsByUser(user.ID, &models.AnalyticsFilters{Limit: 10})
	AssertNoError(t, err, "Should get events after cleanup")
	AssertEqual(t, 5, len(events), "Recent events should not be cleaned up")
}

func TestAnalyticsService_EdgeCases(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	t.Run("Nil request should fail", func(t *testing.T) {
		event, err := suite.AnalyticsService.TrackEvent(1, nil)
		AssertError(t, err, "Nil request should cause error")
		AssertNil(t, event, "Event should be nil")
	})

	t.Run("Empty filters should work", func(t *testing.T) {
		events, err := suite.AnalyticsService.GetEventsByUser(1, nil)
		AssertNoError(t, err, "Nil filters should work")
		AssertNotNil(t, events, "Events should not be nil")
	})

	t.Run("Large metadata should work", func(t *testing.T) {
		user := CreateTestUser(t, suite.DB.DB, 2)
		largeMetadata := make(map[string]interface{})
		for i := 0; i < 100; i++ {
			largeMetadata[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
		}

		event, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
			EventType: "large_metadata_test",
			Metadata:  largeMetadata,
			SessionID: "large_session",
		})
		AssertNoError(t, err, "Large metadata should work")
		AssertNotNil(t, event, "Event should be created")
	})

	t.Run("Special characters in event type", func(t *testing.T) {
		user := CreateTestUser(t, suite.DB.DB, 3)
		event, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
			EventType: "test-event_with.special/chars",
			SessionID: "special_session",
		})
		AssertNoError(t, err, "Special characters should work")
		AssertNotNil(t, event, "Event should be created")
	})
}

// Benchmark tests
func BenchmarkAnalyticsService_TrackEvent(b *testing.B) {
	suite := BenchmarkSetup(b)
	defer suite.Cleanup()

	user := CreateTestUser(b, suite.DB.DB, 1)
	request := &models.AnalyticsEventRequest{
		EventType:  "benchmark_event",
		EntityType: "test_entity",
		EntityID:   1,
		SessionID:  "benchmark_session",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := suite.AnalyticsService.TrackEvent(user.ID, request)
			if err != nil {
				b.Errorf("TrackEvent failed: %v", err)
			}
		}
	})
}

func BenchmarkAnalyticsService_GetAnalytics(b *testing.B) {
	suite := BenchmarkSetup(b)
	defer suite.Cleanup()

	user := CreateTestUser(b, suite.DB.DB, 1)

	// Create test data
	for i := 0; i < 1000; i++ {
		_, err := suite.AnalyticsService.TrackEvent(user.ID, &models.AnalyticsEventRequest{
			EventType: "benchmark_event",
			SessionID: "benchmark_session",
		})
		if err != nil {
			b.Fatalf("Failed to create test data: %v", err)
		}
	}

	filters := &models.AnalyticsFilters{Limit: 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := suite.AnalyticsService.GetAnalytics(user.ID, filters)
		if err != nil {
			b.Errorf("GetAnalytics failed: %v", err)
		}
	}
}