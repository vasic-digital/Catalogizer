package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSyncService(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	assert.NotNil(t, service)
}

func TestSyncService_ValidateSyncEndpoint(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name     string
		endpoint *models.SyncEndpoint
		wantErr  bool
	}{
		{
			name: "valid endpoint",
			endpoint: &models.SyncEndpoint{
				Name:          "Test Endpoint",
				Type:          models.SyncTypeWebDAV,
				URL:           "https://example.com/api",
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/tmp/sync",
				RemotePath:    "/remote/sync",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			endpoint: &models.SyncEndpoint{
				Name: "",
				URL:  "https://example.com/api",
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			endpoint: &models.SyncEndpoint{
				Name: "Test",
				URL:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSyncEndpoint(tt.endpoint)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncService_IsValidType(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	validTypes := []string{models.SyncTypeWebDAV, models.SyncTypeCloudStorage, models.SyncTypeLocal}

	tests := []struct {
		name     string
		syncType string
		expected bool
	}{
		{
			name:     "valid webdav sync type",
			syncType: models.SyncTypeWebDAV,
			expected: true,
		},
		{
			name:     "valid cloud_storage sync type",
			syncType: models.SyncTypeCloudStorage,
			expected: true,
		},
		{
			name:     "valid local sync type",
			syncType: models.SyncTypeLocal,
			expected: true,
		},
		{
			name:     "invalid sync type",
			syncType: "unknown",
			expected: false,
		},
		{
			name:     "empty sync type",
			syncType: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isValidType(tt.syncType, validTypes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncService_ShouldSkipFile(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name     string
		filename string
		endpoint *models.SyncEndpoint
		expected bool
	}{
		{
			name:     "hidden file is skipped",
			filename: ".gitignore",
			endpoint: &models.SyncEndpoint{},
			expected: true,
		},
		{
			name:     "temp file is skipped",
			filename: "data.tmp",
			endpoint: &models.SyncEndpoint{},
			expected: true,
		},
		{
			name:     "normal file is not skipped",
			filename: "test.txt",
			endpoint: &models.SyncEndpoint{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldSkipFile(tt.filename, tt.endpoint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncService_ShouldRunSchedule(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	pastTime := time.Now().Add(-2 * time.Hour)

	tests := []struct {
		name     string
		schedule *models.SyncSchedule
		expected bool
	}{
		{
			name: "hourly schedule due",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyHourly,
				LastRun:   &pastTime,
				IsActive:  true,
			},
			expected: true,
		},
		{
			name: "hourly schedule not due",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyHourly,
				LastRun:   func() *time.Time { t := time.Now(); return &t }(),
				IsActive:  true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldRunSchedule(tt.schedule)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncService_CalculateChecksum(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	// Create a temporary file for checksum testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	checksum, err := service.calculateChecksum(tmpFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, checksum)

	// Same file should produce same checksum
	checksum2, err := service.calculateChecksum(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, checksum, checksum2)

	// Non-existent file should return error
	_, err = service.calculateChecksum("/nonexistent/file.txt")
	assert.Error(t, err)
}
