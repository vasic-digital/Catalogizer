package services

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/models"
	"catalogizer/repository"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSyncTestDB creates an in-memory SQLite database with sync tables for testing.
func newSyncTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	rawDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	db := database.WrapDB(rawDB, database.DialectSQLite)
	require.NotNil(t, db)

	// Create sync tables
	schema := `
	CREATE TABLE IF NOT EXISTS sync_endpoints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		url TEXT NOT NULL,
		username TEXT DEFAULT '',
		password TEXT DEFAULT '',
		sync_direction TEXT NOT NULL,
		local_path TEXT NOT NULL,
		remote_path TEXT DEFAULT '',
		sync_settings TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_sync_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS sync_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		endpoint_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		sync_type TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		duration INTEGER,
		total_files INTEGER DEFAULT 0,
		synced_files INTEGER DEFAULT 0,
		failed_files INTEGER DEFAULT 0,
		skipped_files INTEGER DEFAULT 0,
		error_message TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id)
	);

	CREATE TABLE IF NOT EXISTS sync_schedules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		endpoint_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		frequency TEXT NOT NULL,
		last_run DATETIME,
		next_run DATETIME,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (endpoint_id) REFERENCES sync_endpoints(id)
	);
	`
	_, err = rawDB.Exec(schema)
	require.NoError(t, err)

	cleanup := func() {
		rawDB.Close()
	}

	return db, cleanup
}

// ---------------------------------------------------------------------------
// SyncRepository — CreateEndpoint / GetEndpoint / UpdateEndpoint / DeleteEndpoint
// ---------------------------------------------------------------------------

func TestSyncRepository_CreateAndGetEndpoint(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Test WebDAV",
		Type:          models.SyncTypeWebDAV,
		URL:           "https://example.com/webdav",
		Username:      "testuser",
		Password:      "testpass",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/data/local",
		RemotePath:    "/data/remote",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)
	assert.Greater(t, id, 0)

	// Retrieve
	got, err := repo.GetEndpoint(id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "Test WebDAV", got.Name)
	assert.Equal(t, models.SyncTypeWebDAV, got.Type)
	assert.Equal(t, models.SyncDirectionUpload, got.SyncDirection)
	assert.Equal(t, "/data/local", got.LocalPath)
	assert.Equal(t, "/data/remote", got.RemotePath)
	assert.Equal(t, models.SyncStatusActive, got.Status)
}

func TestSyncRepository_GetEndpoint_NotFound(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	_, err := repo.GetEndpoint(9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint not found")
}

func TestSyncRepository_UpdateEndpoint(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Original Name",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp/sync",
		SyncDirection: models.SyncDirectionBidirectional,
		LocalPath:     "/original/path",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	// Update
	endpoint.ID = id
	endpoint.Name = "Updated Name"
	endpoint.URL = "file:///tmp/updated"
	endpoint.UpdatedAt = time.Now()
	now := time.Now()
	endpoint.LastSyncAt = &now

	err = repo.UpdateEndpoint(endpoint)
	require.NoError(t, err)

	// Verify
	got, err := repo.GetEndpoint(id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", got.Name)
	assert.Equal(t, "file:///tmp/updated", got.URL)
	assert.NotNil(t, got.LastSyncAt)
}

func TestSyncRepository_DeleteEndpoint(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "To Delete",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp/del",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/del/path",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	err = repo.DeleteEndpoint(id)
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetEndpoint(id)
	assert.Error(t, err)
}

func TestSyncRepository_GetUserEndpoints(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	for i := 0; i < 3; i++ {
		ep := &models.SyncEndpoint{
			UserID:        42,
			Name:          "EP_" + string(rune('A'+i)),
			Type:          models.SyncTypeLocal,
			URL:           "file:///tmp",
			SyncDirection: models.SyncDirectionUpload,
			LocalPath:     "/data",
			Status:        models.SyncStatusActive,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_, err := repo.CreateEndpoint(ep)
		require.NoError(t, err)
	}

	// Different user
	ep := &models.SyncEndpoint{
		UserID:        99,
		Name:          "Other User",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/data",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err := repo.CreateEndpoint(ep)
	require.NoError(t, err)

	endpoints, err := repo.GetUserEndpoints(42)
	require.NoError(t, err)
	assert.Equal(t, 3, len(endpoints))

	endpoints99, err := repo.GetUserEndpoints(99)
	require.NoError(t, err)
	assert.Equal(t, 1, len(endpoints99))
}

// ---------------------------------------------------------------------------
// SyncRepository — CreateSession / GetSession / UpdateSession / GetUserSessions
// ---------------------------------------------------------------------------

func TestSyncRepository_CreateAndGetSession(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	session := &models.SyncSession{
		EndpointID:  1,
		UserID:      1,
		Status:      models.SyncSessionStatusRunning,
		SyncType:    models.SyncTypeManual,
		StartedAt:   time.Now(),
		TotalFiles:  100,
		SyncedFiles: 0,
	}

	id, err := repo.CreateSession(session)
	require.NoError(t, err)
	assert.Greater(t, id, 0)

	got, err := repo.GetSession(id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, models.SyncSessionStatusRunning, got.Status)
	assert.Equal(t, models.SyncTypeManual, got.SyncType)
	assert.Equal(t, 100, got.TotalFiles)
}

func TestSyncRepository_GetSession_NotFound(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	_, err := repo.GetSession(9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestSyncRepository_UpdateSession(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	session := &models.SyncSession{
		EndpointID: 1,
		UserID:     1,
		Status:     models.SyncSessionStatusRunning,
		SyncType:   models.SyncTypeManual,
		StartedAt:  time.Now(),
	}

	id, err := repo.CreateSession(session)
	require.NoError(t, err)

	// Update to completed
	session.ID = id
	session.Status = models.SyncSessionStatusCompleted
	now := time.Now()
	session.CompletedAt = &now
	dur := 5 * time.Minute
	session.Duration = &dur
	session.TotalFiles = 50
	session.SyncedFiles = 45
	session.FailedFiles = 3
	session.SkippedFiles = 2

	err = repo.UpdateSession(session)
	require.NoError(t, err)

	got, err := repo.GetSession(id)
	require.NoError(t, err)
	assert.Equal(t, models.SyncSessionStatusCompleted, got.Status)
	assert.NotNil(t, got.CompletedAt)
	assert.NotNil(t, got.Duration)
	assert.Equal(t, 50, got.TotalFiles)
	assert.Equal(t, 45, got.SyncedFiles)
	assert.Equal(t, 3, got.FailedFiles)
	assert.Equal(t, 2, got.SkippedFiles)
}

func TestSyncRepository_UpdateSession_WithError(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	session := &models.SyncSession{
		EndpointID: 1,
		UserID:     1,
		Status:     models.SyncSessionStatusRunning,
		SyncType:   models.SyncTypeManual,
		StartedAt:  time.Now(),
	}

	id, err := repo.CreateSession(session)
	require.NoError(t, err)

	session.ID = id
	session.Status = models.SyncSessionStatusFailed
	now := time.Now()
	session.CompletedAt = &now
	errMsg := "connection timeout"
	session.ErrorMessage = &errMsg

	err = repo.UpdateSession(session)
	require.NoError(t, err)

	got, err := repo.GetSession(id)
	require.NoError(t, err)
	assert.Equal(t, models.SyncSessionStatusFailed, got.Status)
	require.NotNil(t, got.ErrorMessage)
	assert.Equal(t, "connection timeout", *got.ErrorMessage)
}

func TestSyncRepository_GetUserSessions(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	for i := 0; i < 5; i++ {
		session := &models.SyncSession{
			EndpointID: 1,
			UserID:     42,
			Status:     models.SyncSessionStatusCompleted,
			SyncType:   models.SyncTypeManual,
			StartedAt:  time.Now().Add(-time.Duration(i) * time.Hour),
		}
		_, err := repo.CreateSession(session)
		require.NoError(t, err)
	}

	// Get first page
	sessions, err := repo.GetUserSessions(42, 3, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, len(sessions))

	// Get second page
	sessions, err = repo.GetUserSessions(42, 3, 3)
	require.NoError(t, err)
	assert.Equal(t, 2, len(sessions))
}

// ---------------------------------------------------------------------------
// SyncRepository — CreateSchedule / GetActiveSchedules
// ---------------------------------------------------------------------------

func TestSyncRepository_CreateSchedule(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	schedule := &models.SyncSchedule{
		EndpointID: 1,
		UserID:     1,
		Frequency:  models.SyncFrequencyHourly,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	id, err := repo.CreateSchedule(schedule)
	require.NoError(t, err)
	assert.Greater(t, id, 0)
}

func TestSyncRepository_GetActiveSchedules(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	// Create active and inactive schedules
	active := &models.SyncSchedule{
		EndpointID: 1,
		UserID:     1,
		Frequency:  models.SyncFrequencyDaily,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}
	_, err := repo.CreateSchedule(active)
	require.NoError(t, err)

	inactive := &models.SyncSchedule{
		EndpointID: 2,
		UserID:     1,
		Frequency:  models.SyncFrequencyWeekly,
		IsActive:   false,
		CreatedAt:  time.Now(),
	}
	_, err = repo.CreateSchedule(inactive)
	require.NoError(t, err)

	schedules, err := repo.GetActiveSchedules()
	require.NoError(t, err)
	assert.Equal(t, 1, len(schedules))
	assert.Equal(t, models.SyncFrequencyDaily, schedules[0].Frequency)
}

// ---------------------------------------------------------------------------
// SyncRepository — GetStatistics
// ---------------------------------------------------------------------------

func TestSyncRepository_GetStatistics(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(time.Hour)

	// Create sessions with different statuses
	for _, status := range []string{
		models.SyncSessionStatusCompleted,
		models.SyncSessionStatusCompleted,
		models.SyncSessionStatusFailed,
	} {
		session := &models.SyncSession{
			EndpointID:  1,
			UserID:      1,
			Status:      status,
			SyncType:    models.SyncTypeManual,
			StartedAt:   time.Now(),
			SyncedFiles: 10,
			FailedFiles: 2,
		}
		id, err := repo.CreateSession(session)
		require.NoError(t, err)

		// Mark completed with duration
		session.ID = id
		now := time.Now()
		session.CompletedAt = &now
		dur := 30 * time.Second
		session.Duration = &dur
		err = repo.UpdateSession(session)
		require.NoError(t, err)
	}

	stats, err := repo.GetStatistics(nil, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalSessions)
	assert.Equal(t, 2, stats.ByStatus[models.SyncSessionStatusCompleted])
	assert.Equal(t, 1, stats.ByStatus[models.SyncSessionStatusFailed])

	// Filter by user
	userID := 1
	stats, err = repo.GetStatistics(&userID, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalSessions)
}

// ---------------------------------------------------------------------------
// SyncRepository — CleanupSessions
// ---------------------------------------------------------------------------

func TestSyncRepository_CleanupSessions(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	// Create old completed session
	session := &models.SyncSession{
		EndpointID: 1,
		UserID:     1,
		Status:     models.SyncSessionStatusCompleted,
		SyncType:   models.SyncTypeManual,
		StartedAt:  time.Now().Add(-48 * time.Hour),
	}
	id, err := repo.CreateSession(session)
	require.NoError(t, err)

	// Complete it with old date
	session.ID = id
	oldTime := time.Now().Add(-47 * time.Hour)
	session.CompletedAt = &oldTime
	err = repo.UpdateSession(session)
	require.NoError(t, err)

	// Create recent session
	recentSession := &models.SyncSession{
		EndpointID: 1,
		UserID:     1,
		Status:     models.SyncSessionStatusCompleted,
		SyncType:   models.SyncTypeManual,
		StartedAt:  time.Now(),
	}
	recentID, err := repo.CreateSession(recentSession)
	require.NoError(t, err)

	recentSession.ID = recentID
	now := time.Now()
	recentSession.CompletedAt = &now
	err = repo.UpdateSession(recentSession)
	require.NoError(t, err)

	// Cleanup old sessions (older than 24 hours)
	err = repo.CleanupSessions(time.Now().Add(-24 * time.Hour))
	require.NoError(t, err)

	// Old session should be deleted
	_, err = repo.GetSession(id)
	assert.Error(t, err)

	// Recent session should still exist
	got, err := repo.GetSession(recentID)
	require.NoError(t, err)
	assert.Equal(t, recentID, got.ID)
}

// ---------------------------------------------------------------------------
// SyncService — copyFile
// ---------------------------------------------------------------------------

func TestSyncService_CopyFile(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(srcFile, []byte("hello sync world"), 0644)
	require.NoError(t, err)

	// Copy to a new subdirectory
	dstFile := filepath.Join(tmpDir, "sub", "dest.txt")
	err = service.copyFile(srcFile, dstFile, 0644)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, "hello sync world", string(content))
}

func TestSyncService_CopyFile_NonexistentSource(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tmpDir := t.TempDir()
	err := service.copyFile("/nonexistent/source.txt", filepath.Join(tmpDir, "dest.txt"), 0644)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// SyncService — shouldSkipFile additional patterns
// ---------------------------------------------------------------------------

func TestSyncService_ShouldSkipFile_TempExtension(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	assert.True(t, service.shouldSkipFile("file.temp", &models.SyncEndpoint{}))
	assert.True(t, service.shouldSkipFile("file.tmp", &models.SyncEndpoint{}))
	assert.True(t, service.shouldSkipFile(".DS_Store", &models.SyncEndpoint{}))
	assert.False(t, service.shouldSkipFile("document.pdf", &models.SyncEndpoint{}))
	assert.False(t, service.shouldSkipFile("video.mp4", &models.SyncEndpoint{}))
}

// ---------------------------------------------------------------------------
// SyncService — handleSyncSuccess / handleSyncError (pure logic)
// ---------------------------------------------------------------------------

func TestSyncService_HandleSyncSuccess(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	session := &models.SyncSession{
		EndpointID: 1,
		UserID:     1,
		Status:     models.SyncSessionStatusRunning,
		SyncType:   models.SyncTypeManual,
		StartedAt:  time.Now().Add(-5 * time.Minute),
	}

	id, err := repo.CreateSession(session)
	require.NoError(t, err)
	session.ID = id

	service.handleSyncSuccess(session)

	assert.Equal(t, models.SyncSessionStatusCompleted, session.Status)
	assert.NotNil(t, session.CompletedAt)
	assert.NotNil(t, session.Duration)
}

func TestSyncService_HandleSyncError(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	session := &models.SyncSession{
		EndpointID: 1,
		UserID:     1,
		Status:     models.SyncSessionStatusRunning,
		SyncType:   models.SyncTypeManual,
		StartedAt:  time.Now().Add(-2 * time.Minute),
	}

	id, err := repo.CreateSession(session)
	require.NoError(t, err)
	session.ID = id

	service.handleSyncError(session, assert.AnError)

	assert.Equal(t, models.SyncSessionStatusFailed, session.Status)
	assert.NotNil(t, session.CompletedAt)
	assert.NotNil(t, session.ErrorMessage)
	assert.NotNil(t, session.Duration)
}

// ---------------------------------------------------------------------------
// SyncRepository — GetEndpointsByType
// ---------------------------------------------------------------------------

func TestSyncRepository_GetEndpointsByType(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)

	// Create endpoints of different types
	types := []string{models.SyncTypeWebDAV, models.SyncTypeWebDAV, models.SyncTypeLocal}
	for i, typ := range types {
		ep := &models.SyncEndpoint{
			UserID:        1,
			Name:          "EP_" + string(rune('A'+i)),
			Type:          typ,
			URL:           "test://url",
			SyncDirection: models.SyncDirectionUpload,
			LocalPath:     "/data",
			Status:        models.SyncStatusActive,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_, err := repo.CreateEndpoint(ep)
		require.NoError(t, err)
	}

	webdavEPs, err := repo.GetEndpointsByType(models.SyncTypeWebDAV)
	require.NoError(t, err)
	assert.Equal(t, 2, len(webdavEPs))

	localEPs, err := repo.GetEndpointsByType(models.SyncTypeLocal)
	require.NoError(t, err)
	assert.Equal(t, 1, len(localEPs))

	cloudEPs, err := repo.GetEndpointsByType(models.SyncTypeCloudStorage)
	require.NoError(t, err)
	assert.Equal(t, 0, len(cloudEPs))
}

// ---------------------------------------------------------------------------
// SyncService — UpdateEndpoint (0% coverage)
// ---------------------------------------------------------------------------

func TestSyncService_UpdateEndpoint(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	// Create a local-type endpoint (avoids network in testConnection)
	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Original",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp/sync",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/data/local",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	t.Run("owner can update own endpoint", func(t *testing.T) {
		updated, err := service.UpdateEndpoint(id, 1, &models.UpdateSyncEndpointRequest{
			Name: "Updated Name",
			URL:  "file:///tmp/new-sync",
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "file:///tmp/new-sync", updated.URL)
	})

	t.Run("update with multiple fields", func(t *testing.T) {
		isActive := true
		updated, err := service.UpdateEndpoint(id, 1, &models.UpdateSyncEndpointRequest{
			Username:      "newuser",
			Password:      "newpass",
			SyncDirection: models.SyncDirectionDownload,
			LocalPath:     "/new/local",
			IsActive:      &isActive,
		})
		require.NoError(t, err)
		assert.Equal(t, "newuser", updated.Username)
		assert.Equal(t, models.SyncDirectionDownload, updated.SyncDirection)
		assert.Equal(t, "/new/local", updated.LocalPath)
	})

	t.Run("nonexistent endpoint returns error", func(t *testing.T) {
		_, err := service.UpdateEndpoint(9999, 1, &models.UpdateSyncEndpointRequest{
			Name: "Ghost",
		})
		assert.Error(t, err)
	})

	t.Run("set inactive via IsActive pointer", func(t *testing.T) {
		isActive := false
		updated, err := service.UpdateEndpoint(id, 1, &models.UpdateSyncEndpointRequest{
			IsActive: &isActive,
		})
		require.NoError(t, err)
		assert.Equal(t, models.SyncStatusInactive, updated.Status)
	})

	t.Run("set sync settings", func(t *testing.T) {
		settings := `{"key":"value"}`
		updated, err := service.UpdateEndpoint(id, 1, &models.UpdateSyncEndpointRequest{
			SyncSettings: &settings,
		})
		require.NoError(t, err)
		require.NotNil(t, updated.SyncSettings)
		assert.Equal(t, settings, *updated.SyncSettings)
	})
}

// ---------------------------------------------------------------------------
// SyncService — GetSession (37.5% coverage)
// ---------------------------------------------------------------------------

func TestSyncService_GetSession(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	session := &models.SyncSession{
		EndpointID: 1,
		UserID:     1,
		Status:     models.SyncSessionStatusRunning,
		SyncType:   models.SyncTypeManual,
		StartedAt:  time.Now(),
	}

	id, err := repo.CreateSession(session)
	require.NoError(t, err)

	t.Run("owner can view own session", func(t *testing.T) {
		got, err := service.GetSession(id, 1)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("nonexistent session returns error", func(t *testing.T) {
		_, err := service.GetSession(9999, 1)
		assert.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// SyncService — ScheduleSync (0% coverage)
// ---------------------------------------------------------------------------

func TestSyncService_ScheduleSync(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	// Create an endpoint first
	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Schedule Test",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/data",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	epID, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	t.Run("owner can schedule sync", func(t *testing.T) {
		schedule, err := service.ScheduleSync(epID, 1, &models.SyncSchedule{
			Frequency: models.SyncFrequencyDaily,
		})
		require.NoError(t, err)
		assert.Equal(t, epID, schedule.EndpointID)
		assert.Equal(t, 1, schedule.UserID)
		assert.True(t, schedule.IsActive)
		assert.Equal(t, models.SyncFrequencyDaily, schedule.Frequency)
		assert.Greater(t, schedule.ID, 0)
	})

	t.Run("nonexistent endpoint returns error", func(t *testing.T) {
		_, err := service.ScheduleSync(9999, 1, &models.SyncSchedule{
			Frequency: models.SyncFrequencyWeekly,
		})
		assert.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// SyncService — GetEndpoint (37.5% coverage)
// ---------------------------------------------------------------------------

func TestSyncService_GetEndpoint(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	endpoint := &models.SyncEndpoint{
		UserID:        1,
		Name:          "Get Test",
		Type:          models.SyncTypeLocal,
		URL:           "file:///tmp",
		SyncDirection: models.SyncDirectionUpload,
		LocalPath:     "/data",
		Status:        models.SyncStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	epID, err := repo.CreateEndpoint(endpoint)
	require.NoError(t, err)

	t.Run("owner can get own endpoint", func(t *testing.T) {
		got, err := service.GetEndpoint(epID, 1)
		require.NoError(t, err)
		assert.Equal(t, "Get Test", got.Name)
	})

	t.Run("nonexistent endpoint returns error", func(t *testing.T) {
		_, err := service.GetEndpoint(9999, 1)
		assert.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// SyncService — logSyncError / updateSyncProgress / notifyUser (0% coverage)
// ---------------------------------------------------------------------------

func TestSyncService_LogSyncError(t *testing.T) {
	service := NewSyncService(nil, nil, nil)
	session := &models.SyncSession{ID: 42, UserID: 1}

	// These are simple fmt.Printf wrappers — just verify they don't panic
	service.logSyncError(session, "test error message")
	service.updateSyncProgress(session, "50% complete")
	service.notifyUser(session, "sync completed")
}

// ---------------------------------------------------------------------------
// SyncService — GetSyncStatistics
// ---------------------------------------------------------------------------

func TestSyncService_GetSyncStatistics(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	stats, err := service.GetSyncStatistics(nil, time.Now().Add(-24*time.Hour), time.Now())
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalSessions)
}

// ---------------------------------------------------------------------------
// SyncService — GetUserSessions
// ---------------------------------------------------------------------------

func TestSyncService_GetUserSessions(t *testing.T) {
	db, cleanup := newSyncTestDB(t)
	defer cleanup()

	repo := repository.NewSyncRepository(db)
	service := NewSyncService(repo, nil, nil)

	sessions, err := service.GetUserSessions(1, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, sessions)
}
