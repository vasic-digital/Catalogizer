package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// SyncService — DeleteEndpoint (0% coverage)
// ---------------------------------------------------------------------------

func TestSyncService_DeleteEndpoint_OwnerDeletes(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Delete me",
		Type:          models.SyncTypeWebDAV,
		URL:           "https://example.com/webdav",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/tmp/test",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	// Owner can delete their own endpoint
	err = service.DeleteEndpoint(id, 1)
	assert.NoError(t, err)

	// Verify it was deleted
	_, err = repo.GetEndpoint(id)
	assert.Error(t, err)
}

func TestSyncService_DeleteEndpoint_NonOwnerNoPermission(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	// Create users table and seed a user for auth check
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT,
		salt TEXT,
		role_id INTEGER NOT NULL DEFAULT 1,
		first_name TEXT, last_name TEXT, display_name TEXT, avatar_url TEXT,
		time_zone TEXT, language TEXT, settings TEXT DEFAULT '{}',
		is_active BOOLEAN DEFAULT 1, is_locked BOOLEAN DEFAULT 0,
		locked_until DATETIME, failed_login_attempts INTEGER DEFAULT 0,
		last_login_at DATETIME, last_login_ip TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE, description TEXT,
		permissions TEXT DEFAULT '{}', is_system BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system) VALUES (1, 'user', 'Regular user', '["read","write"]', 1)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT OR IGNORE INTO users (id, username, email, password_hash, salt, role_id, is_active) VALUES (999, 'otheruser', 'other@example.com', '', '', 1, 1)`)
	require.NoError(t, err)

	repo := repository.NewSyncRepository(db)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")
	service := NewSyncService(repo, nil, authService)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Protected",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/tmp/test",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	// Different user without delete_shares permission should fail
	err = service.DeleteEndpoint(id, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestSyncService_DeleteEndpoint_NotFound(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	err := service.DeleteEndpoint(99999, 1)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// SyncService — StartSync (0% coverage)
// ---------------------------------------------------------------------------

func TestSyncService_StartSync_EndpointNotFound(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	_, err := service.StartSync(9999, 1)
	assert.Error(t, err)
}

func TestSyncService_StartSync_NotActive(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Inactive",
		Type:          models.SyncTypeWebDAV,
		URL:           "https://example.com",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/tmp",
		Status:        models.SyncStatusInactive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	_, err = service.StartSync(id, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

func TestSyncService_StartSync_NonOwnerNoPermission(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	// Create users + roles tables for auth check
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL, email TEXT UNIQUE NOT NULL,
		password_hash TEXT, salt TEXT, role_id INTEGER NOT NULL DEFAULT 1,
		first_name TEXT, last_name TEXT, display_name TEXT, avatar_url TEXT,
		time_zone TEXT, language TEXT, settings TEXT DEFAULT '{}',
		is_active BOOLEAN DEFAULT 1, is_locked BOOLEAN DEFAULT 0,
		locked_until DATETIME, failed_login_attempts INTEGER DEFAULT 0,
		last_login_at DATETIME, last_login_ip TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE, description TEXT,
		permissions TEXT DEFAULT '{}', is_system BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system) VALUES (1, 'user', 'Regular user', '["read","write"]', 1)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT OR IGNORE INTO users (id, username, email, password_hash, salt, role_id, is_active) VALUES (999, 'otheruser', 'other@example.com', '', '', 1, 1)`)
	require.NoError(t, err)

	repo := repository.NewSyncRepository(db)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")
	service := NewSyncService(repo, nil, authService)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Private",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/tmp",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	_, err = service.StartSync(id, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestSyncService_StartSync_OwnerSuccess(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Active",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/tmp/nonexistent-path-test",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	session, err := service.StartSync(id, 1)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, models.SyncSessionStatusRunning, session.Status)
	assert.Equal(t, models.SyncTypeManual, session.SyncType)

	// Give goroutine a moment to run
	time.Sleep(50 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// SyncService — performSync (0% coverage, indirectly via StartSync)
// ---------------------------------------------------------------------------

func TestSyncService_PerformSync_UnsupportedType(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "BadType",
		Type:          "unknown_type",
		URL:           "file:///tmp",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/tmp",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	session, err := service.StartSync(id, 1)
	require.NoError(t, err)
	assert.NotNil(t, session)

	// Wait for the goroutine to complete (it will fail and call handleSyncError)
	time.Sleep(100 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// SyncService — performCloudSync / performS3Sync / performGCS (0% coverage)
// ---------------------------------------------------------------------------

func TestSyncService_PerformCloudSync_UnsupportedCloudType(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	endpoint := &models.SyncEndpoint{Type: "unknown_cloud"}

	err := service.performCloudSync(session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported cloud storage type")
}

func TestSyncService_PerformS3Sync_NilSettings(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	endpoint := &models.SyncEndpoint{SyncSettings: nil}

	err := service.performS3Sync(nil, session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "S3 bucket not specified")
}

func TestSyncService_PerformS3Sync_MissingBucket(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	settings := `{"region": "us-east-1"}`
	endpoint := &models.SyncEndpoint{SyncSettings: &settings}

	err := service.performS3Sync(nil, session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "S3 bucket not specified")
}

func TestSyncService_PerformS3Sync_MissingAccessKey(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	settings := `{"bucket": "my-bucket", "region": "us-east-1"}`
	endpoint := &models.SyncEndpoint{SyncSettings: &settings}

	err := service.performS3Sync(nil, session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "S3 access key not specified")
}

func TestSyncService_PerformS3Sync_MissingSecretKey(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	settings := `{"bucket": "my-bucket", "access_key": "AKIA..."}`
	endpoint := &models.SyncEndpoint{SyncSettings: &settings}

	err := service.performS3Sync(nil, session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "S3 secret key not specified")
}

func TestSyncService_PerformS3Sync_InvalidJSON(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	settings := `{invalid json}`
	endpoint := &models.SyncEndpoint{SyncSettings: &settings}

	err := service.performS3Sync(nil, session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse S3 config")
}

func TestSyncService_PerformGCS_NilSettings(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	endpoint := &models.SyncEndpoint{SyncSettings: nil}

	err := service.performGoogleCloudStorageSync(nil, session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GCS bucket not specified")
}

func TestSyncService_PerformGCS_InvalidJSON(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	settings := `not valid json`
	endpoint := &models.SyncEndpoint{SyncSettings: &settings}

	err := service.performGoogleCloudStorageSync(nil, session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse Google Cloud Storage config")
}

func TestSyncService_PerformGCS_MissingBucket(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	settings := `{"credentials_file": "/tmp/creds.json"}`
	endpoint := &models.SyncEndpoint{SyncSettings: &settings}

	err := service.performGoogleCloudStorageSync(nil, session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GCS bucket not specified")
}

// ---------------------------------------------------------------------------
// SyncService — performWebDAVSync / uploadToWebDAV / downloadFromWebDAV (0%)
// ---------------------------------------------------------------------------

func TestSyncService_PerformWebDAVSync_UnsupportedDirection(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	endpoint := &models.SyncEndpoint{
		ID:            1,
		URL:           "https://example.com/webdav",
		Username:      "user",
		Password:      "pass",
		SyncDirection: "invalid_direction",
		LocalPath:     "/tmp",
		RemotePath:    "/remote",
	}

	session := &models.SyncSession{}
	err := service.performWebDAVSync(session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported sync direction")
}

func TestSyncService_PerformWebDAVSync_UploadScanFailure(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	endpoint := &models.SyncEndpoint{
		ID:            2,
		URL:           "https://example.com/webdav",
		Username:      "user",
		Password:      "pass",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/nonexistent/path/for/test",
		RemotePath:    "/remote",
	}

	session := &models.SyncSession{}
	err := service.performWebDAVSync(session, endpoint)
	// Upload tries to scan local files from a nonexistent path
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// SyncService — performBidirectionalSync (60% -> higher)
// ---------------------------------------------------------------------------

func TestSyncService_PerformBidirectionalSync_EmptyDirs(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	// Create temp dirs for bidirectional sync
	localDir := t.TempDir()
	remoteDir := t.TempDir()

	session := &models.SyncSession{Status: models.SyncSessionStatusRunning}

	err := service.performBidirectionalSync(localDir, remoteDir, session)
	assert.NoError(t, err)
}

func TestSyncService_PerformBidirectionalSync_WithFiles(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	localDir := t.TempDir()
	remoteDir := t.TempDir()

	// Create test files in both dirs
	err := os.WriteFile(filepath.Join(localDir, "local.txt"), []byte("local"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(remoteDir, "remote.txt"), []byte("remote"), 0644)
	require.NoError(t, err)

	session := &models.SyncSession{Status: models.SyncSessionStatusRunning}

	err = service.performBidirectionalSync(localDir, remoteDir, session)
	assert.NoError(t, err)
}

func TestSyncService_PerformBidirectionalSync_NonexistentSource(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	remoteDir := t.TempDir()

	session := &models.SyncSession{Status: models.SyncSessionStatusRunning}

	err := service.performBidirectionalSync("/nonexistent/path", remoteDir, session)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// AnalyticsService — GenerateReport (36.4% -> higher)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GenerateReport_NilRepo(t *testing.T) {
	svc := NewAnalyticsService(nil)

	tests := []struct {
		name       string
		reportType string
	}{
		{name: "user_activity type", reportType: "user_activity"},
		{name: "media_popularity type", reportType: "media_popularity"},
		{name: "system_overview type", reportType: "system_overview"},
		{name: "geographic_analysis type", reportType: "geographic_analysis"},
		{name: "unknown report type", reportType: "unknown_type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &models.ReportRequest{
				ReportType: tt.reportType,
				Params:     map[string]interface{}{"user_id": 1},
			}

			report, err := svc.GenerateReport(1, request)
			// With nil repo, CreateReport fails, but GenerateReport falls back to summary
			if tt.reportType == "unknown_type" {
				// Fallback for unsupported type in CreateReport
				require.NoError(t, err)
				assert.NotNil(t, report)
			} else {
				// Fallback because analyticsRepo is nil
				require.NoError(t, err)
				assert.NotNil(t, report)
				assert.Equal(t, "completed", report.Status)
			}
		})
	}
}

func TestAnalyticsService_GenerateReport_WithParams(t *testing.T) {
	svc := NewAnalyticsService(nil)

	request := &models.ReportRequest{
		ReportType: "custom",
		Params: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	report, err := svc.GenerateReport(1, request)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "custom", report.Type)
	assert.Contains(t, report.Data, "custom")
}

func TestAnalyticsService_GenerateReport_NilParams(t *testing.T) {
	svc := NewAnalyticsService(nil)

	request := &models.ReportRequest{
		ReportType: "summary",
		Params:     nil,
	}

	report, err := svc.GenerateReport(1, request)
	require.NoError(t, err)
	assert.NotNil(t, report)
}

// ---------------------------------------------------------------------------
// AnalyticsService — calculateAverageSessionDuration (42.9% -> higher)
// Tested indirectly via DB-backed tests below since it calls repo directly.
// ---------------------------------------------------------------------------

func TestAnalyticsService_CalculateAverageSessionDuration_ViaDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create session_data table for the analytics repo
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS session_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		duration INTEGER NOT NULL DEFAULT 0,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ended_at DATETIME
	)`)
	require.NoError(t, err)

	analyticsRepo := repository.NewAnalyticsRepository(db)
	svc := NewAnalyticsService(analyticsRepo)

	// No sessions => returns 0
	result := svc.calculateAverageSessionDuration(
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)
	assert.Equal(t, time.Duration(0), result)
}

// ---------------------------------------------------------------------------
// AnalyticsService — analyzePeakUsageHours, analyzePopularFileTypes,
// analyzeGeographicDistribution (75% -> higher)
// ---------------------------------------------------------------------------

func TestAnalyticsService_AnalyzePeakUsageHours_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	analyticsRepo := repository.NewAnalyticsRepository(db)
	svc := NewAnalyticsService(analyticsRepo)

	result := svc.analyzePeakUsageHours(
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)
	assert.NotNil(t, result)
}

func TestAnalyticsService_AnalyzePopularFileTypes_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	analyticsRepo := repository.NewAnalyticsRepository(db)
	svc := NewAnalyticsService(analyticsRepo)

	result := svc.analyzePopularFileTypes(
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)
	assert.NotNil(t, result)
}

func TestAnalyticsService_AnalyzeGeographicDistribution_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	analyticsRepo := repository.NewAnalyticsRepository(db)
	svc := NewAnalyticsService(analyticsRepo)

	result := svc.analyzeGeographicDistribution(
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)
	assert.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// ReportingService — calculateSecurityMetrics (70% -> higher)
// ---------------------------------------------------------------------------

func TestReportingService_CalculateSecurityMetrics_Boost(t *testing.T) {
	svc := NewReportingService(nil, nil)

	tests := []struct {
		name            string
		startDate       time.Time
		endDate         time.Time
		expectedThreat  string
		minScore        float64
		maxScore        float64
	}{
		{
			name:           "short period high score",
			startDate:      time.Now().Add(-24 * time.Hour),
			endDate:        time.Now(),
			expectedThreat: "low",
			minScore:       90.0,
			maxScore:       100.0,
		},
		{
			name:           "medium period",
			startDate:      time.Now().Add(-180 * 24 * time.Hour),
			endDate:        time.Now(),
			expectedThreat: "low",
			minScore:       70.0,
			maxScore:       100.0,
		},
		{
			name:           "long period medium threat",
			startDate:      time.Now().Add(-365 * 24 * time.Hour),
			endDate:        time.Now(),
			expectedThreat: "medium",
			minScore:       50.0,
			maxScore:       70.0,
		},
		{
			name:           "very long period caps at 50",
			startDate:      time.Now().Add(-1000 * 24 * time.Hour),
			endDate:        time.Now(),
			expectedThreat: "medium",
			minScore:       50.0,
			maxScore:       50.0,
		},
		{
			name:           "same day",
			startDate:      time.Now(),
			endDate:        time.Now(),
			expectedThreat: "low",
			minScore:       99.0,
			maxScore:       100.0,
		},
		{
			name:           "end before start (negative days capped to 1)",
			startDate:      time.Now(),
			endDate:        time.Now().Add(-24 * time.Hour),
			expectedThreat: "low",
			minScore:       99.0,
			maxScore:       100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := svc.calculateSecurityMetrics(tt.startDate, tt.endDate)
			assert.Equal(t, tt.expectedThreat, metrics.ThreatLevel)
			assert.GreaterOrEqual(t, metrics.SecurityScore, tt.minScore)
			assert.LessOrEqual(t, metrics.SecurityScore, tt.maxScore)
			assert.Equal(t, 0, metrics.VulnerabilityCount)
		})
	}
}

// ---------------------------------------------------------------------------
// ReportingService — calculateAverageSessionDuration
// ---------------------------------------------------------------------------

func TestReportingService_CalculateAverageSessionDuration_Boost(t *testing.T) {
	svc := NewReportingService(nil, nil)

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
			result := svc.calculateAverageSessionDuration(tt.sessions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// LogManagementService — compressLogFile (0% coverage)
// ---------------------------------------------------------------------------

func TestLogManagementService_CompressLogFile(t *testing.T) {
	svc := NewLogManagementService(nil)

	// Create a temp file to compress
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	err := os.WriteFile(logFile, []byte("test log content\nline2\nline3"), 0644)
	require.NoError(t, err)

	err = svc.compressLogFile(logFile)
	require.NoError(t, err)

	// Verify the .gz file was created
	_, err = os.Stat(logFile + ".gz")
	assert.NoError(t, err)

	// Verify original was removed
	_, err = os.Stat(logFile)
	assert.True(t, os.IsNotExist(err))
}

func TestLogManagementService_CompressLogFile_NotFound(t *testing.T) {
	svc := NewLogManagementService(nil)

	err := svc.compressLogFile("/nonexistent/path/test.log")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// LogManagementService — DatabaseLogCollector.CollectLogs (0% coverage)
// Uses logRepo which is nil in our test => should return error
// ---------------------------------------------------------------------------

func TestDatabaseLogCollector_Methods(t *testing.T) {
	collector := &DatabaseLogCollector{
		logRepo:       nil,
		componentName: "test-component",
	}

	assert.Equal(t, "database", collector.GetLogPath())
	assert.Equal(t, "test-component", collector.GetComponentName())
}

// ---------------------------------------------------------------------------
// LogManagementService — RevokeLogShare (70% -> higher)
// ---------------------------------------------------------------------------

func TestLogManagementService_RevokeLogShare_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := repository.NewLogManagementRepository(db)
	svc := NewLogManagementService(logRepo)

	// Non-existent share should return error
	err := svc.RevokeLogShare(9999, 1)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ErrorReportingService — sendToCrashlytics (10% -> higher)
// ---------------------------------------------------------------------------

func TestErrorReportingService_SendToCrashlytics_EmptyAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	errorRepo := repository.NewErrorReportingRepository(db)
	crashRepo := repository.NewCrashReportingRepository(db)
	svc := NewErrorReportingService(errorRepo, crashRepo)

	// Default config has empty CrashlyticsAPIKey
	report := &models.CrashReport{
		Signal:     "SIGSEGV",
		Message:    "test crash",
		ReportedAt: time.Now(),
	}

	err := svc.sendToCrashlytics(report)
	assert.NoError(t, err) // returns nil when key is empty
}

// ---------------------------------------------------------------------------
// ConversionService — convertPDFWithImageMagick (0% coverage)
// ---------------------------------------------------------------------------

func TestConversionService_ConvertPDFWithImageMagick_MissingConvert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	svc := NewConversionService(nil, nil, nil)

	job := &models.ConversionJob{
		SourcePath: "/tmp/nonexistent.pdf",
		TargetPath: "/tmp/output.png",
		Settings:   nil,
	}

	// This will fail because either convert is not found or the source doesn't exist
	err := svc.convertPDFWithImageMagick(job, "png")
	assert.Error(t, err)
}

func TestConversionService_ConvertPDFWithImageMagick_WithSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	svc := NewConversionService(nil, nil, nil)

	settings := `{"dpi": 300, "quality": 90}`
	job := &models.ConversionJob{
		SourcePath: "/tmp/nonexistent.pdf",
		TargetPath: "/tmp/output.png",
		Settings:   &settings,
	}

	err := svc.convertPDFWithImageMagick(job, "png")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ConversionService — ProcessJobQueue (70% -> higher)
// ---------------------------------------------------------------------------

func TestConversionService_ProcessJobQueue_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	conversionRepo := repository.NewConversionRepository(db)
	svc := NewConversionService(conversionRepo, nil, nil)

	// Empty queue should succeed with no jobs to process
	err := svc.ProcessJobQueue()
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// FavoritesService — GetRecommendedFavorites (10% -> higher)
// ---------------------------------------------------------------------------

func TestFavoritesService_GetRecommendedFavorites_NilRepoBoost(t *testing.T) {
	svc := NewFavoritesService(nil, nil)

	_, err := svc.GetRecommendedFavorites(1, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

// ---------------------------------------------------------------------------
// FavoritesService — exportFavoritesToJSON (80% -> higher)
// ---------------------------------------------------------------------------

func TestFavoritesService_ExportFavoritesToJSON_EmptyList(t *testing.T) {
	svc := NewFavoritesService(nil, nil)

	data, err := svc.exportFavoritesToJSON([]models.Favorite{})
	require.NoError(t, err)
	assert.Contains(t, string(data), `"count": 0`)
	assert.Contains(t, string(data), `"version": "1.0"`)
}

func TestFavoritesService_ExportFavoritesToJSON_WithFavorites(t *testing.T) {
	svc := NewFavoritesService(nil, nil)

	favorites := []models.Favorite{
		{ID: 1, UserID: 1, EntityType: "movie", EntityID: 100},
		{ID: 2, UserID: 1, EntityType: "song", EntityID: 200},
	}

	data, err := svc.exportFavoritesToJSON(favorites)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"count": 2`)
	assert.Contains(t, string(data), "movie")
	assert.Contains(t, string(data), "song")
}

// ---------------------------------------------------------------------------
// ErrorReportingService — CleanupOldReports (60% -> higher)
// ---------------------------------------------------------------------------

func TestErrorReportingService_CleanupOldReports(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	errorRepo := repository.NewErrorReportingRepository(db)
	crashRepo := repository.NewCrashReportingRepository(db)
	svc := NewErrorReportingService(errorRepo, crashRepo)

	// Cleanup with a past date should succeed (nothing to cleanup)
	err := svc.CleanupOldReports(time.Now().Add(-365 * 24 * time.Hour))
	assert.NoError(t, err)

	// Cleanup with future date should also succeed
	err = svc.CleanupOldReports(time.Now().Add(24 * time.Hour))
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// LogManagementService — CollectLogs (50% -> higher)
// ---------------------------------------------------------------------------

func TestFileLogCollector_CollectLogs_FileNotFound(t *testing.T) {
	collector := &FileLogCollector{
		logPath:       "/nonexistent/path.log",
		componentName: "test",
	}

	_, err := collector.CollectLogs()
	assert.Error(t, err)
}

func TestFileLogCollector_CollectLogs_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	err := os.WriteFile(logFile, []byte("log line 1\nlog line 2\n"), 0644)
	require.NoError(t, err)

	collector := &FileLogCollector{
		logPath:       logFile,
		componentName: "test",
	}

	entries, err := collector.CollectLogs()
	require.NoError(t, err)
	// The simplified implementation returns an empty (nil) slice
	assert.Empty(t, entries)
}

func TestFileLogCollector_Accessors(t *testing.T) {
	collector := &FileLogCollector{
		logPath:       "/var/log/test.log",
		componentName: "api",
	}

	assert.Equal(t, "/var/log/test.log", collector.GetLogPath())
	assert.Equal(t, "api", collector.GetComponentName())
}
