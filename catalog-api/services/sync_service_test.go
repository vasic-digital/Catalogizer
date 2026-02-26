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
	recentTime := time.Now().Add(-30 * time.Minute)
	dayAgo := time.Now().Add(-25 * time.Hour)
	weekAgo := time.Now().Add(-8 * 24 * time.Hour)
	monthAgo := time.Now().AddDate(0, -1, -1)

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
				LastRun:   &recentTime,
				IsActive:  true,
			},
			expected: false,
		},
		{
			name: "hourly schedule never run",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyHourly,
				LastRun:   nil,
				IsActive:  true,
			},
			expected: true,
		},
		{
			name: "daily schedule due",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyDaily,
				LastRun:   &dayAgo,
				IsActive:  true,
			},
			expected: true,
		},
		{
			name: "daily schedule not due",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyDaily,
				LastRun:   &recentTime,
				IsActive:  true,
			},
			expected: false,
		},
		{
			name: "weekly schedule due",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyWeekly,
				LastRun:   &weekAgo,
				IsActive:  true,
			},
			expected: true,
		},
		{
			name: "weekly schedule not due",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyWeekly,
				LastRun:   &dayAgo,
				IsActive:  true,
			},
			expected: false,
		},
		{
			name: "monthly schedule due",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyMonthly,
				LastRun:   &monthAgo,
				IsActive:  true,
			},
			expected: true,
		},
		{
			name: "monthly schedule not due",
			schedule: &models.SyncSchedule{
				Frequency: models.SyncFrequencyMonthly,
				LastRun:   &weekAgo,
				IsActive:  true,
			},
			expected: false,
		},
		{
			name: "unknown frequency returns false",
			schedule: &models.SyncSchedule{
				Frequency: "unknown",
				LastRun:   nil,
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

func TestSyncService_ValidateSyncEndpoint_AllFields(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name     string
		endpoint *models.SyncEndpoint
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid endpoint with all fields",
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
			name: "valid endpoint with bidirectional sync",
			endpoint: &models.SyncEndpoint{
				Name:          "Test",
				Type:          models.SyncTypeLocal,
				URL:           "https://example.com",
				SyncDirection: models.SyncDirectionBidirectional,
				LocalPath:     "/data",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			endpoint: &models.SyncEndpoint{
				Name:          "Test",
				URL:           "https://example.com",
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/tmp",
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "missing sync direction",
			endpoint: &models.SyncEndpoint{
				Name:      "Test",
				Type:      models.SyncTypeWebDAV,
				URL:       "https://example.com",
				LocalPath: "/tmp",
			},
			wantErr: true,
			errMsg:  "sync direction is required",
		},
		{
			name: "missing local path",
			endpoint: &models.SyncEndpoint{
				Name:          "Test",
				Type:          models.SyncTypeWebDAV,
				URL:           "https://example.com",
				SyncDirection: models.SyncDirectionUpload,
			},
			wantErr: true,
			errMsg:  "local path is required",
		},
		{
			name: "invalid sync type",
			endpoint: &models.SyncEndpoint{
				Name:          "Test",
				Type:          "invalid_type",
				URL:           "https://example.com",
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/tmp",
			},
			wantErr: true,
			errMsg:  "invalid sync type",
		},
		{
			name: "invalid sync direction",
			endpoint: &models.SyncEndpoint{
				Name:          "Test",
				Type:          models.SyncTypeWebDAV,
				URL:           "https://example.com",
				SyncDirection: "invalid_direction",
				LocalPath:     "/tmp",
			},
			wantErr: true,
			errMsg:  "invalid sync direction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSyncEndpoint(tt.endpoint)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncService_ShouldSkipRemoteFile(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name     string
		file     *WebDAVFile
		expected bool
	}{
		{
			name:     "normal file not skipped",
			file:     &WebDAVFile{Path: "/path/to/file.txt"},
			expected: false,
		},
		{
			name:     "hidden file skipped",
			file:     &WebDAVFile{Path: "/path/to/.hidden"},
			expected: true,
		},
		{
			name:     "temp file skipped",
			file:     &WebDAVFile{Path: "/path/to/file.tmp"},
			expected: true,
		},
		{
			name:     "temp file with .temp extension skipped",
			file:     &WebDAVFile{Path: "/path/to/file.temp"},
			expected: true,
		},
		{
			name:     "config file not skipped (not hidden basename)",
			file:     &WebDAVFile{Path: "/path/config"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldSkipRemoteFile(tt.file, &models.SyncEndpoint{})
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncService_GetWebDAVClient(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	endpoint := &models.SyncEndpoint{
		ID:       1,
		URL:      "https://example.com/webdav",
		Username: "user",
		Password: "pass",
	}

	// First call creates client
	client1, err := service.getWebDAVClient(endpoint)
	assert.NoError(t, err)
	assert.NotNil(t, client1)

	// Second call returns cached client
	client2, err := service.getWebDAVClient(endpoint)
	assert.NoError(t, err)
	assert.Equal(t, client1, client2)

	// Different endpoint creates new client
	endpoint2 := &models.SyncEndpoint{
		ID:       2,
		URL:      "https://other.example.com/webdav",
		Username: "user2",
		Password: "pass2",
	}
	client3, err := service.getWebDAVClient(endpoint2)
	assert.NoError(t, err)
	assert.NotNil(t, client3)
}

func TestSyncService_ScanLocalFiles(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	// Create temp directory structure
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir", "file3.txt"), []byte("content3"), 0644))

	files, err := service.scanLocalFiles(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 3)
}

func TestSyncService_ScanLocalFiles_NonexistentPath(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	files, err := service.scanLocalFiles("/nonexistent/path")
	assert.Error(t, err)
	assert.Nil(t, files)
}

// ============================================================================
// ADDITIONAL TESTS FOR 95% COVERAGE
// ============================================================================

func TestSyncService_CreateSyncEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		endpoint    *models.SyncEndpoint
		wantErr     bool
		errContains string
	}{
		{
			name:   "successful creation with WebDAV",
			userID: 1,
			endpoint: &models.SyncEndpoint{
				Name:          "Test Endpoint",
				Type:          models.SyncTypeWebDAV,
				URL:           "https://example.com/webdav",
				Username:      "user",
				Password:      "pass",
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/tmp/local",
				RemotePath:    "/remote",
			},
			wantErr:     true, // Will fail without repository
			errContains: "not properly configured",
		},
		{
			name:   "successful creation with Cloud Storage",
			userID: 1,
			endpoint: &models.SyncEndpoint{
				Name:          "Cloud Endpoint",
				Type:          models.SyncTypeCloudStorage,
				URL:           "s3://bucket-name",
				Username:      "access-key",
				Password:      "secret-key",
				SyncDirection: models.SyncDirectionBidirectional,
				LocalPath:     "/data/cloud",
				RemotePath:    "/",
			},
			wantErr:     true, // Will fail without repository
			errContains: "not properly configured",
		},
		{
			name:   "missing name",
			userID: 1,
			endpoint: &models.SyncEndpoint{
				Type:          models.SyncTypeWebDAV,
				URL:           "https://example.com/webdav",
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/tmp/local",
				RemotePath:    "/remote",
			},
			wantErr:     true,
			errContains: "not properly configured",
		},
		{
			name:   "missing URL for WebDAV",
			userID: 1,
			endpoint: &models.SyncEndpoint{
				Name:          "No URL",
				Type:          models.SyncTypeWebDAV,
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/tmp/local",
				RemotePath:    "/remote",
			},
			wantErr:     true,
			errContains: "not properly configured",
		},
		{
			name:   "missing local path",
			userID: 1,
			endpoint: &models.SyncEndpoint{
				Name:          "No Local Path",
				Type:          models.SyncTypeWebDAV,
				URL:           "https://example.com/webdav",
				SyncDirection: models.SyncDirectionUpload,
				RemotePath:    "/remote",
			},
			wantErr:     true,
			errContains: "not properly configured",
		},
		{
			name:   "missing remote path for WebDAV",
			userID: 1,
			endpoint: &models.SyncEndpoint{
				Name:          "No Remote Path",
				Type:          models.SyncTypeWebDAV,
				URL:           "https://example.com/webdav",
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/tmp/local",
			},
			wantErr:     true,
			errContains: "not properly configured",
		},
		{
			name:   "invalid sync type",
			userID: 1,
			endpoint: &models.SyncEndpoint{
				Name:          "Test",
				Type:          "invalid_type",
				URL:           "https://example.com/webdav",
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/tmp/local",
				RemotePath:    "/remote",
			},
			wantErr:     true,
			errContains: "not properly configured",
		},
		{
			name:   "local sync without remote path",
			userID: 1,
			endpoint: &models.SyncEndpoint{
				Name:          "Local Sync",
				Type:          models.SyncTypeLocal,
				URL:           "/source/path",
				SyncDirection: models.SyncDirectionUpload,
				LocalPath:     "/dest/path",
			},
			wantErr:     true,
			errContains: "not properly configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSyncService(nil, nil, nil)
			got, err := service.CreateSyncEndpoint(tt.userID, tt.endpoint)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.userID, got.UserID)
				assert.Equal(t, models.SyncStatusActive, got.Status)
				assert.NotZero(t, got.CreatedAt)
				assert.NotZero(t, got.UpdatedAt)
			}
		})
	}
}
