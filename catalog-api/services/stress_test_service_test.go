package services

import (
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
)

func TestNewStressTestService(t *testing.T) {
	service := NewStressTestService(nil, nil)

	assert.NotNil(t, service)
}

func TestStressTestService_ValidateTestConfiguration(t *testing.T) {
	service := NewStressTestService(nil, nil)

	tests := []struct {
		name    string
		test    *models.StressTest
		wantErr bool
	}{
		{
			name: "empty name",
			test: &models.StressTest{
				Name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateTestConfiguration(tt.test)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStressTestService_SelectRandomScenario(t *testing.T) {
	service := NewStressTestService(nil, nil)

	scenarios := []models.StressTestScenario{
		{Name: "read", URL: "/api/v1/test", Method: "GET"},
		{Name: "write", URL: "/api/v1/test", Method: "POST"},
		{Name: "mixed", URL: "/api/v1/test", Method: "PUT"},
		{Name: "search", URL: "/api/v1/search", Method: "GET"},
	}

	for i := 0; i < 10; i++ {
		scenario := service.selectRandomScenario(scenarios)
		assert.NotNil(t, scenario)
	}
}

func TestStressTestService_SelectRandomScenario_EmptyList(t *testing.T) {
	service := NewStressTestService(nil, nil)

	scenario := service.selectRandomScenario([]models.StressTestScenario{})
	assert.Nil(t, scenario)
}

func TestStressTestService_GenerateRecommendations(t *testing.T) {
	service := NewStressTestService(nil, nil)

	tests := []struct {
		name   string
		result *models.StressTestResult
	}{
		{
			name: "good performance",
			result: &models.StressTestResult{
				ErrorRate:       0.01,
				AvgResponseTime: 50.0,
				TotalRequests:   1000,
				SuccessfulReqs:  990,
			},
		},
		{
			name: "poor performance",
			result: &models.StressTestResult{
				ErrorRate:       25.0,
				AvgResponseTime: 2000.0,
				TotalRequests:   1000,
				SuccessfulReqs:  750,
			},
		},
		{
			name: "zero values",
			result: &models.StressTestResult{
				ErrorRate:       0.0,
				AvgResponseTime: 0.0,
				TotalRequests:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := service.generateRecommendations(tt.result)
			assert.NotNil(t, recommendations)
		})
	}
}

func TestStressTestService_GenerateReportSummary(t *testing.T) {
	service := NewStressTestService(nil, nil)

	tests := []struct {
		name   string
		test   *models.StressTest
		result *models.StressTestResult
	}{
		{
			name: "normal run",
			test: &models.StressTest{
				Name:            "Load Test",
				ConcurrentUsers: 10,
				DurationSeconds: 60,
			},
			result: &models.StressTestResult{
				TotalRequests:   1000,
				SuccessfulReqs:  995,
				FailedRequests:  5,
				AvgResponseTime: 100.0,
				Duration:        60 * time.Second,
			},
		},
		{
			name: "zero requests",
			test: &models.StressTest{
				Name:            "Empty Test",
				ConcurrentUsers: 0,
				DurationSeconds: 0,
			},
			result: &models.StressTestResult{
				TotalRequests:   0,
				SuccessfulReqs:  0,
				FailedRequests:  0,
				AvgResponseTime: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := service.generateReportSummary(tt.test, tt.result)
			assert.NotEmpty(t, summary)
		})
	}
}

func TestStressTestService_CreateTestMetrics(t *testing.T) {
	service := NewStressTestService(nil, nil)

	metrics := service.createTestMetrics()

	assert.NotNil(t, metrics)
}

func TestStressTestService_GetSystemLoad(t *testing.T) {
	service := NewStressTestService(nil, nil)

	load, err := service.GetSystemLoad()

	assert.NoError(t, err)
	assert.NotNil(t, load)
}
