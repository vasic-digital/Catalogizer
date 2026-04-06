package services

import (
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ConfigurationService — DB-backed tests for 66-75% functions
// ---------------------------------------------------------------------------

func TestConfigurationService_GetConfiguration_LoadsDefault(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	configRepo := repository.NewConfigurationRepository(db)
	svc := NewConfigurationService(configRepo, t.TempDir()+"/config.json")

	config, err := svc.GetConfiguration()
	require.NoError(t, err)
	assert.NotNil(t, config)
}

func TestConfigurationService_CompleteWizard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	configRepo := repository.NewConfigurationRepository(db)
	svc := NewConfigurationService(configRepo, t.TempDir()+"/config.json")

	finalData := map[string]interface{}{
		"db_type":    "sqlite",
		"db_host":    "localhost",
		"db_port":    5432,
		"db_name":    "catalogizer",
		"jwt_secret": "test-secret-key",
	}

	config, err := svc.CompleteWizard(1, finalData)
	require.NoError(t, err)
	assert.NotNil(t, config)
}

func TestConfigurationService_SaveAndLoadConfiguration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	configRepo := repository.NewConfigurationRepository(db)
	svc := NewConfigurationService(configRepo, t.TempDir()+"/config.json")

	// Get default config
	config, err := svc.GetConfiguration()
	require.NoError(t, err)

	// Save it
	err = svc.SaveConfiguration(config)
	require.NoError(t, err)

	// Load it again
	config2, err := svc.GetConfiguration()
	require.NoError(t, err)
	assert.NotNil(t, config2)
}

func TestConfigurationService_ExportConfiguration_Boost(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	configRepo := repository.NewConfigurationRepository(db)
	svc := NewConfigurationService(configRepo, t.TempDir()+"/config.json")

	data, err := svc.ExportConfiguration()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

// ---------------------------------------------------------------------------
// AnalyticsService — GenerateReport with DB (36.4% -> higher)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GenerateReport_WithDBRepo(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create session_data table
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

	tests := []struct {
		name       string
		reportType string
		wantErr    bool
	}{
		{name: "user_activity", reportType: "user_activity", wantErr: false},
		{name: "system_overview", reportType: "system_overview", wantErr: false},
		{name: "unsupported", reportType: "nonexistent", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &models.ReportRequest{
				ReportType: tt.reportType,
				Params: map[string]interface{}{
					"start_date": time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
					"end_date":   time.Now().Format("2006-01-02"),
					"user_id":    1,
				},
			}

			report, err := svc.GenerateReport(1, request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, report)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AnalyticsService — CreateReport with DB (72.7% functions -> higher)
// ---------------------------------------------------------------------------

func TestAnalyticsService_CreateReport_UnsupportedType_Boost(t *testing.T) {
	svc := NewAnalyticsService(nil)

	_, err := svc.CreateReport("unknown_type", map[string]interface{}{
		"start_date": time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
		"end_date":   time.Now().Format("2006-01-02"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported report type")
}

func TestAnalyticsService_CreateReport_MissingDateRange(t *testing.T) {
	svc := NewAnalyticsService(nil)

	_, err := svc.CreateReport("user_activity", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start_date")
}

func TestAnalyticsService_CreateReport_WithDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create session_data table
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

	params := map[string]interface{}{
		"start_date": time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
		"end_date":   time.Now().Format("2006-01-02"),
	}

	reportTypes := []string{"user_activity", "system_overview"}
	for _, rt := range reportTypes {
		t.Run(rt, func(t *testing.T) {
			report, err := svc.CreateReport(rt, params)
			require.NoError(t, err)
			assert.NotNil(t, report)
			assert.Equal(t, "completed", report.Status)
		})
	}
}

// ---------------------------------------------------------------------------
// AnalyticsService — GetUserAnalytics with DB (75% -> higher)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GetUserAnalytics_WithDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	analyticsRepo := repository.NewAnalyticsRepository(db)
	svc := NewAnalyticsService(analyticsRepo)

	analytics, err := svc.GetUserAnalytics(1,
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)
	require.NoError(t, err)
	assert.NotNil(t, analytics)
	assert.Equal(t, 1, analytics.UserID)
}

// ---------------------------------------------------------------------------
// AnalyticsService — GetSystemAnalytics with DB (70% -> higher)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GetSystemAnalytics_WithDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create session_data table
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

	analytics, err := svc.GetSystemAnalytics(
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)
	require.NoError(t, err)
	assert.NotNil(t, analytics)
}

// ---------------------------------------------------------------------------
// AnalyticsService — GetMediaAnalytics with DB (83.3% -> higher)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GetMediaAnalytics_WithDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	analyticsRepo := repository.NewAnalyticsRepository(db)
	svc := NewAnalyticsService(analyticsRepo)

	analytics, err := svc.GetMediaAnalytics(1,
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)
	require.NoError(t, err)
	assert.NotNil(t, analytics)
	assert.Equal(t, 1, analytics.MediaID)
}

// ---------------------------------------------------------------------------
// ErrorReportingService — sendToCrashlytics with API key (10% -> higher)
// ---------------------------------------------------------------------------

func TestErrorReportingService_SendToCrashlytics_WithAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	errorRepo := repository.NewErrorReportingRepository(db)
	crashRepo := repository.NewCrashReportingRepository(db)
	svc := NewErrorReportingService(errorRepo, crashRepo)

	// Set a fake API key but no HTTP client -> returns nil since httpClient is nil
	svc.config.CrashlyticsAPIKey = "fake-key"
	svc.httpClient = nil

	report := &models.CrashReport{
		Signal:     "SIGSEGV",
		Message:    "test crash",
		ReportedAt: time.Now(),
	}

	err := svc.sendToCrashlytics(report)
	assert.NoError(t, err) // Returns nil when httpClient is nil
}

// ---------------------------------------------------------------------------
// SyncService — performLocalSync covers local sync path
// ---------------------------------------------------------------------------

func TestSyncService_StartSync_LocalType(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	localDir := t.TempDir()
	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "LocalSync",
		Type:          models.SyncTypeLocal,
		URL:           "file://" + localDir,
		SyncDirection: models.SyncDirectionBidirectional,
		LocalPath:     localDir,
		RemotePath:    t.TempDir(),
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	session, err := service.StartSync(id, 1)
	require.NoError(t, err)
	assert.NotNil(t, session)

	// Give goroutine a moment to run
	time.Sleep(100 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// SyncService — performCloudSync covers cloud sync paths
// ---------------------------------------------------------------------------

func TestSyncService_PerformCloudSync_S3Type(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	endpoint := &models.SyncEndpoint{Type: "s3", SyncSettings: nil}

	err := service.performCloudSync(session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "S3 bucket not specified")
}

func TestSyncService_PerformCloudSync_GoogleDriveType(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{}
	endpoint := &models.SyncEndpoint{Type: "google_drive", SyncSettings: nil}

	err := service.performCloudSync(session, endpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GCS bucket not specified")
}

// ---------------------------------------------------------------------------
// LogManagementService — StreamLogs
// ---------------------------------------------------------------------------

func TestLogManagementService_StreamLogs_DisabledBoost(t *testing.T) {
	svc := NewLogManagementService(nil)
	svc.config.RealTimeLogging = false

	_, _, err := svc.StreamLogs(1, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "real-time logging is disabled")
}

func TestLogManagementService_StreamLogs_EnabledBoost(t *testing.T) {
	svc := NewLogManagementService(nil)
	svc.config.RealTimeLogging = true

	ch, _, err := svc.StreamLogs(1, &models.LogStreamFilters{})
	require.NoError(t, err)
	assert.NotNil(t, ch)
}

// ---------------------------------------------------------------------------
// AuthService — Login edge cases with DB
// ---------------------------------------------------------------------------

func TestAuthService_Login_InactiveUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	authSvc := NewAuthService(nil, "test-jwt-secret")
	salt, _ := authSvc.generateSalt()
	hash, _ := authSvc.hashPassword("Password123!", salt)

	// Set user as inactive
	_, err := db.Exec(`UPDATE users SET password_hash = ?, salt = ?, is_active = 0 WHERE id = 1`, hash, salt)
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(db)
	svc := NewAuthService(userRepo, "test-jwt-secret")

	_, err = svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "127.0.0.1", "TestAgent/1.0")
	assert.Error(t, err)
}

func TestAuthService_Login_LockedUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	authSvc := NewAuthService(nil, "test-jwt-secret")
	salt, _ := authSvc.generateSalt()
	hash, _ := authSvc.hashPassword("Password123!", salt)

	// Set user as locked until future
	_, err := db.Exec(`UPDATE users SET password_hash = ?, salt = ?, is_locked = 1, locked_until = datetime('now', '+1 hour') WHERE id = 1`, hash, salt)
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(db)
	svc := NewAuthService(userRepo, "test-jwt-secret")

	_, err = svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "127.0.0.1", "TestAgent/1.0")
	assert.Error(t, err)
}
