package services

import (
	"testing"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorReportingService_FilterSensitiveData(t *testing.T) {
	svc := &ErrorReportingService{
		config: &ErrorReportingConfig{
			FilterSensitiveData: true,
		},
	}

	tests := []struct {
		name           string
		report         *models.ErrorReport
		checkMessage   string
		checkStack     string
		checkContext   map[string]interface{}
	}{
		{
			name: "redacts password in message",
			report: &models.ErrorReport{
				Message:    "failed to validate password for user",
				StackTrace: "at auth.go:50",
				Context:    nil,
			},
			checkMessage: "failed to validate [redacted] for user",
			checkStack:   "at [redacted].go:50", // "auth" contains "auth" which is a sensitive pattern
		},
		{
			name: "redacts sensitive context keys",
			report: &models.ErrorReport{
				Message:    "connection failed",
				StackTrace: "at db.go:100",
				Context: map[string]interface{}{
					"password":    "secret123",
					"auth_token":  "abc123",
					"request_url": "https://example.com",
					"user_id":     42,
				},
			},
			checkContext: map[string]interface{}{
				"password":    "[REDACTED]",
				"auth_token":  "[REDACTED]",
				"request_url": "https://example.com",
				"user_id":     42,
			},
		},
		{
			name: "nil context is handled",
			report: &models.ErrorReport{
				Message:    "simple error",
				StackTrace: "at main.go:1",
				Context:    nil,
			},
		},
		{
			name: "empty context is handled",
			report: &models.ErrorReport{
				Message:    "simple error",
				StackTrace: "at main.go:1",
				Context:    map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.filterSensitiveData(tt.report)
			require.NotNil(t, result)

			if tt.checkMessage != "" {
				assert.Equal(t, tt.checkMessage, result.Message)
			}

			if tt.checkStack != "" {
				assert.Equal(t, tt.checkStack, result.StackTrace)
			}

			if tt.checkContext != nil {
				for key, expectedVal := range tt.checkContext {
					assert.Equal(t, expectedVal, result.Context[key], "context key %s", key)
				}
			}
		})
	}
}

func TestErrorReportingService_GenerateFingerprint(t *testing.T) {
	svc := &ErrorReportingService{}

	tests := []struct {
		name    string
		report1 *models.ErrorReport
		report2 *models.ErrorReport
		same    bool
	}{
		{
			name: "same error produces same fingerprint",
			report1: &models.ErrorReport{
				Level:     "error",
				Component: "auth",
				ErrorCode: "AUTH_001",
			},
			report2: &models.ErrorReport{
				Level:     "error",
				Component: "auth",
				ErrorCode: "AUTH_001",
			},
			same: true,
		},
		{
			name: "different errors produce different fingerprints",
			report1: &models.ErrorReport{
				Level:     "error",
				Component: "auth",
				ErrorCode: "AUTH_001",
			},
			report2: &models.ErrorReport{
				Level:     "warning",
				Component: "db",
				ErrorCode: "DB_002",
			},
			same: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp1 := svc.generateFingerprint(tt.report1)
			fp2 := svc.generateFingerprint(tt.report2)

			assert.NotEmpty(t, fp1)
			assert.NotEmpty(t, fp2)
			assert.Len(t, fp1, 16)
			assert.Len(t, fp2, 16)

			if tt.same {
				assert.Equal(t, fp1, fp2)
			} else {
				assert.NotEqual(t, fp1, fp2)
			}
		})
	}
}

func TestErrorReportingService_GenerateCrashFingerprint(t *testing.T) {
	svc := &ErrorReportingService{}

	tests := []struct {
		name    string
		report1 *models.CrashReport
		report2 *models.CrashReport
		same    bool
	}{
		{
			name: "same crash produces same fingerprint",
			report1: &models.CrashReport{
				Signal:  "SIGSEGV",
				Message: "segmentation fault",
			},
			report2: &models.CrashReport{
				Signal:  "SIGSEGV",
				Message: "segmentation fault",
			},
			same: true,
		},
		{
			name: "different crashes produce different fingerprints",
			report1: &models.CrashReport{
				Signal:  "SIGSEGV",
				Message: "segmentation fault",
			},
			report2: &models.CrashReport{
				Signal:  "SIGABRT",
				Message: "abort",
			},
			same: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp1 := svc.generateCrashFingerprint(tt.report1)
			fp2 := svc.generateCrashFingerprint(tt.report2)

			assert.NotEmpty(t, fp1)
			assert.NotEmpty(t, fp2)
			assert.Len(t, fp1, 16)
			assert.Len(t, fp2, 16)

			if tt.same {
				assert.Equal(t, fp1, fp2)
			} else {
				assert.NotEqual(t, fp1, fp2)
			}
		})
	}
}

func TestErrorReportingService_GetColorForLevel(t *testing.T) {
	svc := &ErrorReportingService{}

	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{
			name:     "error level",
			level:    "error",
			expected: "danger",
		},
		{
			name:     "fatal level",
			level:    "fatal",
			expected: "danger",
		},
		{
			name:     "ERROR uppercase",
			level:    "ERROR",
			expected: "danger",
		},
		{
			name:     "warning level",
			level:    "warning",
			expected: "warning",
		},
		{
			name:     "warn level",
			level:    "warn",
			expected: "warning",
		},
		{
			name:     "info level",
			level:    "info",
			expected: "good",
		},
		{
			name:     "debug level (default)",
			level:    "debug",
			expected: "#36a64f",
		},
		{
			name:     "unknown level",
			level:    "unknown",
			expected: "#36a64f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.getColorForLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorReportingService_CollectSystemInfo(t *testing.T) {
	svc := &ErrorReportingService{
		config: &ErrorReportingConfig{
			IncludeSystemInfo: true,
		},
	}

	info := svc.collectSystemInfo()

	assert.NotNil(t, info)
	assert.Contains(t, info, "os")
	assert.Contains(t, info, "arch")
	assert.Contains(t, info, "go_version")
	assert.Contains(t, info, "num_cpu")
	assert.Contains(t, info, "num_goroutine")
	assert.Contains(t, info, "memory_alloc")
	assert.Contains(t, info, "memory_total_alloc")
	assert.Contains(t, info, "memory_sys")
	assert.Contains(t, info, "hostname")
	assert.Contains(t, info, "working_directory")
}

func TestErrorReportingService_Configuration(t *testing.T) {
	svc := NewErrorReportingService(nil, nil)

	// Test default configuration
	config := svc.GetConfiguration()
	require.NotNil(t, config)
	assert.False(t, config.CrashlyticsEnabled)
	assert.True(t, config.EmailNotifications)
	assert.True(t, config.AutoReporting)
	assert.Equal(t, 100, config.MaxErrorsPerHour)
	assert.Equal(t, 30, config.RetentionDays)
	assert.True(t, config.IncludeStackTrace)
	assert.True(t, config.IncludeSystemInfo)
	assert.True(t, config.FilterSensitiveData)

	// Test updating configuration
	newConfig := &ErrorReportingConfig{
		CrashlyticsEnabled: true,
		MaxErrorsPerHour:   50,
		RetentionDays:      60,
	}

	err := svc.UpdateConfiguration(newConfig)
	assert.NoError(t, err)

	updatedConfig := svc.GetConfiguration()
	assert.True(t, updatedConfig.CrashlyticsEnabled)
	assert.Equal(t, 50, updatedConfig.MaxErrorsPerHour)
	assert.Equal(t, 60, updatedConfig.RetentionDays)
}

func TestErrorReportingService_EnabledState(t *testing.T) {
	svc := NewErrorReportingService(nil, nil)

	// Service should be enabled by default
	assert.True(t, svc.enabled)
}

func TestErrorReportingService_ReportError_Disabled(t *testing.T) {
	svc := &ErrorReportingService{
		enabled: false,
		config:  &ErrorReportingConfig{},
	}

	report, err := svc.ReportError(1, &models.ErrorReportRequest{
		Level:   "error",
		Message: "test error",
	})

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "error reporting is disabled")
}

func TestErrorReportingService_ReportCrash_Disabled(t *testing.T) {
	svc := &ErrorReportingService{
		enabled: false,
		config:  &ErrorReportingConfig{},
	}

	report, err := svc.ReportCrash(1, &models.CrashReportRequest{
		Signal:  "SIGSEGV",
		Message: "test crash",
	})

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "crash reporting is disabled")
}

func TestErrorReportingService_SendSlackNotification_EmptyWebhook(t *testing.T) {
	svc := &ErrorReportingService{
		config: &ErrorReportingConfig{
			SlackWebhookURL: "",
		},
	}

	err := svc.sendSlackNotification(&models.ErrorReport{
		Level:   "error",
		Message: "test",
	})
	assert.NoError(t, err) // Should return nil when webhook is empty

	err = svc.sendSlackCrashNotification(&models.CrashReport{
		Signal:  "SIGSEGV",
		Message: "test crash",
	})
	assert.NoError(t, err) // Should return nil when webhook is empty
}
