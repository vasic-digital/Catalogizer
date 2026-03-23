package services

import (
	"context"
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// AuthService — DB-backed tests for auth functions at 55-80%
// ---------------------------------------------------------------------------

func newTestAuthServiceWithDB(t *testing.T) (*AuthService, *repository.UserRepository) {
	t.Helper()
	db := setupTestDB(t)

	// Seed a test user with known password
	authSvc := NewAuthService(nil, "test-jwt-secret-key-for-auth-service")
	salt, _ := authSvc.generateSalt()
	hash, _ := authSvc.hashPassword("Password123!", salt)

	_, err := db.Exec(
		`UPDATE users SET password_hash = ?, salt = ?, role_id = 2 WHERE id = 1`,
		hash, salt,
	)
	require.NoError(t, err)

	// Also create user_sessions table if needed (already in setupTestDB)
	userRepo := repository.NewUserRepository(db)
	svc := NewAuthService(userRepo, "test-jwt-secret-key-for-auth-service")
	return svc, userRepo
}

func TestAuthService_LoginAndLogout(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	// Login
	result, err := svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "127.0.0.1", "TestAgent/1.0")
	require.NoError(t, err)
	assert.NotEmpty(t, result.SessionToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotNil(t, result.User)

	// Logout using session token
	err = svc.Logout(result.SessionToken)
	assert.NoError(t, err)
}

func TestAuthService_LoginBadPassword(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	_, err := svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}, "127.0.0.1", "TestAgent/1.0")
	assert.Error(t, err)
}

func TestAuthService_LoginNonexistentUser(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	_, err := svc.Login(models.LoginRequest{
		Username: "nonexistent",
		Password: "whatever",
	}, "127.0.0.1", "TestAgent/1.0")
	assert.Error(t, err)
}

func TestAuthService_RefreshToken(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	// Login first
	result, err := svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "127.0.0.1", "TestAgent/1.0")
	require.NoError(t, err)

	// Refresh
	refreshed, err := svc.RefreshToken(result.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshed.SessionToken)
	assert.NotEmpty(t, refreshed.RefreshToken)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	_, err := svc.RefreshToken("invalid-refresh-token")
	assert.Error(t, err)
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	result, err := svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "127.0.0.1", "TestAgent/1.0")
	require.NoError(t, err)

	user, err := svc.GetCurrentUser(result.SessionToken)
	require.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.NotNil(t, user.Role)
}

func TestAuthService_GetCurrentUser_InvalidToken(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	_, err := svc.GetCurrentUser("invalid-token")
	assert.Error(t, err)
}

func TestAuthService_ChangePassword(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	err := svc.ChangePassword(1, "Password123!", "NewPassword456!")
	require.NoError(t, err)

	// Old password should no longer work
	_, err = svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "127.0.0.1", "TestAgent/1.0")
	assert.Error(t, err)

	// New password should work
	_, err = svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "NewPassword456!",
	}, "127.0.0.1", "TestAgent/1.0")
	require.NoError(t, err)
}

func TestAuthService_ChangePassword_WrongCurrent(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	err := svc.ChangePassword(1, "WrongPassword", "NewPassword456!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "current password is incorrect")
}

func TestAuthService_ResetPassword(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	err := svc.ResetPassword(1, "ResetPassword789!")
	require.NoError(t, err)

	// Login with new password
	_, err = svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "ResetPassword789!",
	}, "127.0.0.1", "TestAgent/1.0")
	require.NoError(t, err)
}

func TestAuthService_CheckPermission(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	// User ID 1 has role_id 2 (admin) with ["read","write","admin",...]
	hasPerm, err := svc.CheckPermission(1, "read")
	require.NoError(t, err)
	assert.True(t, hasPerm)

	// Check a permission that doesn't exist
	hasPerm, err = svc.CheckPermission(1, "nonexistent_perm")
	require.NoError(t, err)
	assert.False(t, hasPerm)
}

func TestAuthService_CheckPermission_NonexistentUser(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	_, err := svc.CheckPermission(9999, "read")
	assert.Error(t, err)
}

func TestAuthService_CheckAccountLockout_NotLocked(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	// User has 0 failed attempts, should not lock
	err := svc.CheckAccountLockout(1)
	assert.NoError(t, err)
}

func TestAuthService_CheckAccountLockout_TooManyAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set failed_login_attempts to 5 to trigger lockout
	_, err := db.Exec(`UPDATE users SET failed_login_attempts = 5 WHERE id = 1`)
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(db)
	svc := NewAuthService(userRepo, "test-secret")

	err = svc.CheckAccountLockout(1)
	assert.NoError(t, err) // LockAccount returns nil on success

	// Verify account is now locked
	user, err := userRepo.GetByID(1)
	require.NoError(t, err)
	assert.True(t, user.IsLocked)
}

func TestAuthService_CheckAccountLockout_NonexistentUser(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	err := svc.CheckAccountLockout(9999)
	assert.Error(t, err)
}

func TestAuthService_LogoutAll(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	// Create a few sessions
	_, err := svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "127.0.0.1", "TestAgent/1.0")
	require.NoError(t, err)

	_, err = svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "192.168.1.1", "TestAgent/2.0")
	require.NoError(t, err)

	// Logout all
	err = svc.LogoutAll(1)
	assert.NoError(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	result, err := svc.Login(models.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}, "127.0.0.1", "TestAgent/1.0")
	require.NoError(t, err)

	claims, err := svc.ValidateToken(result.SessionToken)
	require.NoError(t, err)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestAuthService_ValidateToken_Invalid_Boost(t *testing.T) {
	svc, _ := newTestAuthServiceWithDB(t)

	_, err := svc.ValidateToken("invalid.token.here")
	assert.Error(t, err)
}

func TestAuthService_HashPasswordForUser_Boost(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	hash, salt, err := svc.HashPasswordForUser("TestPassword123!")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEmpty(t, salt)
}

func TestAuthService_GenerateSecureToken_Boost(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	token, err := svc.GenerateSecureToken(32)
	require.NoError(t, err)
	assert.Len(t, token, 64) // hex encoding doubles length
}

func TestAuthService_ValidatePassword_Variations(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "valid password", password: "SecurePass123!", wantErr: false},
		{name: "too short", password: "Ab1!", wantErr: true},
		{name: "no uppercase", password: "lowercaseonly123!", wantErr: true},
		{name: "no lowercase", password: "UPPERCASEONLY123!", wantErr: true},
		{name: "no digit", password: "NoDigitsHere!", wantErr: true},
		{name: "empty", password: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ChallengeService — RunByCategory (21.4%)
// ---------------------------------------------------------------------------

func TestChallengeService_RunByCategory_EmptyCategory(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	_, err := svc.RunByCategory(context.Background(), "nonexistent-category")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no challenges found")
}

// ---------------------------------------------------------------------------
// AnalyticsService — GenerateReport deeper coverage (36.4% -> higher)
// ---------------------------------------------------------------------------

func TestAnalyticsService_GenerateReport_AllReportTypes(t *testing.T) {
	svc := NewAnalyticsService(nil)

	reportTypes := []string{
		"user_activity",
		"media_popularity",
		"system_overview",
		"geographic_analysis",
		"custom_report",
	}

	for _, rt := range reportTypes {
		t.Run(rt, func(t *testing.T) {
			request := &models.ReportRequest{
				ReportType: rt,
				Params: map[string]interface{}{
					"start_date": time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
					"end_date":   time.Now().Format("2006-01-02"),
				},
			}

			report, err := svc.GenerateReport(1, request)
			// All should succeed via fallback path since analyticsRepo is nil
			require.NoError(t, err)
			assert.NotNil(t, report)
			assert.Equal(t, "completed", report.Status)
			assert.NotEmpty(t, report.Data)
		})
	}
}

func TestAnalyticsService_GenerateReport_EmptyParams(t *testing.T) {
	svc := NewAnalyticsService(nil)

	request := &models.ReportRequest{
		ReportType: "summary",
	}

	report, err := svc.GenerateReport(1, request)
	require.NoError(t, err)
	assert.NotNil(t, report)
}

// ---------------------------------------------------------------------------
// ErrorReportingService — UpdateErrorStatus / UpdateCrashStatus (90% -> 100%)
// ---------------------------------------------------------------------------

func TestErrorReportingService_UpdateErrorStatus_Boost(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	errorRepo := repository.NewErrorReportingRepository(db)
	crashRepo := repository.NewCrashReportingRepository(db)
	svc := NewErrorReportingService(errorRepo, crashRepo)

	// Create a report first
	req := &models.ErrorReportRequest{
		Level:     "error",
		Message:   "test error for update",
		Component: "test",
	}

	created, err := svc.ReportError(1, req)
	require.NoError(t, err)

	// Update status
	err = svc.UpdateErrorStatus(created.ID, 1, "resolved")
	assert.NoError(t, err)
}

func TestErrorReportingService_UpdateCrashStatus_Boost(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	errorRepo := repository.NewErrorReportingRepository(db)
	crashRepo := repository.NewCrashReportingRepository(db)
	svc := NewErrorReportingService(errorRepo, crashRepo)

	// Create a crash report first
	req := &models.CrashReportRequest{
		Signal:  "SIGSEGV",
		Message: "test crash for update",
	}

	created, err := svc.ReportCrash(1, req)
	require.NoError(t, err)

	err = svc.UpdateCrashStatus(created.ID, 1, "investigated")
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// LogManagementService — CollectLogs (50% -> higher)
// ---------------------------------------------------------------------------

func TestLogManagementService_CollectLogs_DB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := repository.NewLogManagementRepository(db)
	svc := NewLogManagementService(logRepo)

	request := &models.LogCollectionRequest{
		Name:        "Test Collection",
		Description: "test description",
		Components:  []string{"api"},
		LogLevel:    "info",
	}

	collection, err := svc.CollectLogs(1, request)
	require.NoError(t, err)
	assert.NotNil(t, collection)
	assert.Equal(t, "Test Collection", collection.Name)
}

// ---------------------------------------------------------------------------
// ConversionService — handleConversionSuccess (80% -> higher)
// ---------------------------------------------------------------------------

func TestConversionService_HandleConversionSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	convRepo := repository.NewConversionRepository(db)
	svc := NewConversionService(convRepo, nil, nil)

	// Create a job
	job := &models.ConversionJob{
		UserID:         1,
		SourcePath:     "/tmp/source.mp4",
		TargetPath:     "/tmp/target.avi",
		SourceFormat:   "mp4",
		TargetFormat:   "avi",
		ConversionType: "video",
		Status:         models.ConversionStatusRunning,
	}

	id, err := convRepo.CreateJob(job)
	require.NoError(t, err)
	job.ID = id

	svc.handleConversionSuccess(job)
	// Verify the job status was updated
	updated, err := convRepo.GetJob(id)
	require.NoError(t, err)
	assert.Equal(t, models.ConversionStatusCompleted, updated.Status)
}

// ---------------------------------------------------------------------------
// SyncService — GetUserEndpoints / GetUserSessions / GetSyncStatistics
// ---------------------------------------------------------------------------

func TestSyncService_GetUserEndpoints(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	svc := NewSyncService(repo, nil, nil)

	// Create a few endpoints
	for i := 0; i < 3; i++ {
		_, err := repo.CreateEndpoint(&models.SyncEndpoint{
			UserID:        1,
			Name:          "Endpoint",
			Type:          models.SyncTypeLocal,
			URL:           "file:///tmp",
			SyncDirection: models.SyncDirectionUpload,
			LocalPath:     "/tmp",
			Status:        models.SyncStatusActive,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		require.NoError(t, err)
	}

	endpoints, err := svc.GetUserEndpoints(1)
	require.NoError(t, err)
	assert.Len(t, endpoints, 3)

	// Different user should get empty
	endpoints2, err := svc.GetUserEndpoints(999)
	require.NoError(t, err)
	assert.Len(t, endpoints2, 0)
}

func TestSyncService_GetSyncStatistics_Boost(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	svc := NewSyncService(repo, nil, nil)

	userID := 1
	stats, err := svc.GetSyncStatistics(&userID, time.Now().Add(-24*time.Hour), time.Now())
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestSyncService_GetUserSessions_Boost(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	svc := NewSyncService(repo, nil, nil)

	sessions, err := svc.GetUserSessions(1, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, sessions)
}
