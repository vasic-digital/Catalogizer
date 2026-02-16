package services

import (
	"testing"

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
		config  interface{}
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == nil {
				err := service.validateTestConfiguration(nil)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestStressTestService_SelectRandomScenario(t *testing.T) {
	service := NewStressTestService(nil, nil)

	scenarios := []string{"read", "write", "mixed", "search"}

	for i := 0; i < 10; i++ {
		scenario := service.selectRandomScenario(scenarios)
		assert.Contains(t, scenarios, scenario)
	}
}

func TestStressTestService_SelectRandomScenario_EmptyList(t *testing.T) {
	service := NewStressTestService(nil, nil)

	scenario := service.selectRandomScenario([]string{})
	assert.Empty(t, scenario)
}

func TestStressTestService_GenerateRecommendations(t *testing.T) {
	service := NewStressTestService(nil, nil)

	tests := []struct {
		name       string
		errorRate  float64
		avgLatency float64
	}{
		{
			name:       "good performance",
			errorRate:  0.01,
			avgLatency: 50.0,
		},
		{
			name:       "poor performance",
			errorRate:  0.25,
			avgLatency: 2000.0,
		},
		{
			name:       "zero values",
			errorRate:  0.0,
			avgLatency: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := service.generateRecommendations(tt.errorRate, tt.avgLatency)
			assert.NotNil(t, recommendations)
		})
	}
}

func TestStressTestService_GenerateReportSummary(t *testing.T) {
	service := NewStressTestService(nil, nil)

	tests := []struct {
		name          string
		totalRequests int
		totalErrors   int
		avgLatency    float64
	}{
		{
			name:          "normal run",
			totalRequests: 1000,
			totalErrors:   5,
			avgLatency:    100.0,
		},
		{
			name:          "zero requests",
			totalRequests: 0,
			totalErrors:   0,
			avgLatency:    0.0,
		},
		{
			name:          "all errors",
			totalRequests: 100,
			totalErrors:   100,
			avgLatency:    5000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := service.generateReportSummary(tt.totalRequests, tt.totalErrors, tt.avgLatency)
			assert.NotNil(t, summary)
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

	load := service.GetSystemLoad()

	assert.NotNil(t, load)
}
