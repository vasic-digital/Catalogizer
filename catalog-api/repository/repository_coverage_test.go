package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"catalogizer/database"
	mediamodels "catalogizer/internal/media/models"
	"catalogizer/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// UserRepository — GetSession, GetSessionByRefreshToken (0% coverage)
// ===========================================================================

func newMockUserRepo2(t *testing.T) (*UserRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewUserRepository(db), mock
}

var userSessionColumns = []string{
	"id", "user_id", "session_token", "refresh_token", "device_info",
	"ip_address", "user_agent", "is_active", "expires_at", "created_at",
	"last_activity_at",
}

func TestUserRepository_GetSession(t *testing.T) {
	now := time.Now()
	ipAddr := "127.0.0.1"
	userAgent := "TestAgent/1.0"

	tests := []struct {
		name    string
		id      string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		errMsg  string
		check   func(t *testing.T, s *models.UserSession)
	}{
		{
			name: "success",
			id:   "42",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_sessions WHERE id").
					WithArgs("42").
					WillReturnRows(sqlmock.NewRows(userSessionColumns).
						AddRow(42, 1, "tok123", "refresh456",
							`{"platform":"Linux","device_type":"desktop"}`,
							&ipAddr, &userAgent, true, now.Add(24*time.Hour), now, now))
			},
			check: func(t *testing.T, s *models.UserSession) {
				assert.Equal(t, 42, s.ID)
				assert.Equal(t, 1, s.UserID)
				assert.Equal(t, "tok123", s.SessionToken)
				assert.True(t, s.IsActive)
				assert.Equal(t, "Linux", *s.DeviceInfo.Platform)
			},
		},
		{
			name: "not found",
			id:   "999",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_sessions WHERE id").
					WithArgs("999").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "session not found",
		},
		{
			name: "database error",
			id:   "1",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_sessions WHERE id").
					WithArgs("1").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
			errMsg:  "failed to get session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockUserRepo2(t)
			tt.setup(mock)

			session, err := repo.GetSession(tt.id)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, session)
			tt.check(t, session)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetSessionByRefreshToken(t *testing.T) {
	now := time.Now()
	ipAddr := "10.0.0.1"
	userAgent := "Mozilla/5.0"

	tests := []struct {
		name    string
		token   string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, s *models.UserSession)
	}{
		{
			name:  "success",
			token: "refresh-tok-abc",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_sessions WHERE refresh_token").
					WithArgs("refresh-tok-abc").
					WillReturnRows(sqlmock.NewRows(userSessionColumns).
						AddRow(7, 2, "sess-tok-xyz", "refresh-tok-abc",
							`{"platform":"Windows","device_type":"desktop"}`,
							&ipAddr, &userAgent, true, now.Add(48*time.Hour), now, now))
			},
			check: func(t *testing.T, s *models.UserSession) {
				assert.Equal(t, 7, s.ID)
				assert.Equal(t, 2, s.UserID)
				assert.Equal(t, "Windows", *s.DeviceInfo.Platform)
			},
		},
		{
			name:  "not found",
			token: "nonexistent",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_sessions WHERE refresh_token").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name:  "database error",
			token: "refresh-tok",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_sessions WHERE refresh_token").
					WithArgs("refresh-tok").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockUserRepo2(t)
			tt.setup(mock)

			session, err := repo.GetSessionByRefreshToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, session)
			tt.check(t, session)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// SyncRepository — UpdateSession, GetUserSessions, GetStatistics,
//                   GetEndpointsByType, scanSessions (all 0%)
// ===========================================================================

func newMockSyncRepo2(t *testing.T) (*SyncRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewSyncRepository(db), mock
}

var syncSessionColumnsCov = []string{
	"id", "endpoint_id", "user_id", "status", "sync_type",
	"started_at", "completed_at", "duration", "total_files",
	"synced_files", "failed_files", "skipped_files", "error_message",
}

func TestSyncRepository_UpdateSession(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(10 * time.Minute)
	dur := 10 * time.Minute
	errMsg := "timeout"

	tests := []struct {
		name    string
		session *models.SyncSession
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success with all fields",
			session: &models.SyncSession{
				ID:           1,
				Status:       "completed",
				CompletedAt:  &completedAt,
				Duration:     &dur,
				TotalFiles:   100,
				SyncedFiles:  90,
				FailedFiles:  5,
				SkippedFiles: 5,
				ErrorMessage: nil,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sync_sessions").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "success with error message",
			session: &models.SyncSession{
				ID:           2,
				Status:       "failed",
				CompletedAt:  &completedAt,
				Duration:     &dur,
				ErrorMessage: &errMsg,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sync_sessions").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "success with nil optional fields",
			session: &models.SyncSession{
				ID:     3,
				Status: "running",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sync_sessions").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			session: &models.SyncSession{
				ID:     1,
				Status: "completed",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sync_sessions").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo2(t)
			tt.setup(mock)

			err := repo.UpdateSession(tt.session)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSyncRepository_GetUserSessions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		userID    int
		limit     int
		offset    int
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "returns sessions",
			userID: 1,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_sessions").
					WithArgs(1, 10, 0).
					WillReturnRows(sqlmock.NewRows(syncSessionColumnsCov).
						AddRow(1, 10, 1, "completed", "full", now, now, int64(300),
							100, 95, 3, 2, nil).
						AddRow(2, 10, 1, "failed", "incremental", now, nil, nil,
							50, 10, 40, 0, "connection lost"))
			},
			wantCount: 2,
		},
		{
			name:   "empty result",
			userID: 99,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_sessions").
					WithArgs(99, 10, 0).
					WillReturnRows(sqlmock.NewRows(syncSessionColumnsCov))
			},
			wantCount: 0,
		},
		{
			name:   "database error",
			userID: 1,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_sessions").
					WithArgs(1, 10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo2(t)
			tt.setup(mock)

			sessions, err := repo.GetUserSessions(tt.userID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, sessions, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSyncRepository_GetEndpointsByType(t *testing.T) {
	now := time.Now()

	syncEndpointColumns := []string{
		"id", "user_id", "name", "type", "url", "username", "password",
		"sync_direction", "local_path", "remote_path", "sync_settings",
		"status", "created_at", "updated_at", "last_sync_at",
	}

	tests := []struct {
		name      string
		syncType  string
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:     "returns endpoints",
			syncType: "webdav",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_endpoints").
					WithArgs("webdav").
					WillReturnRows(sqlmock.NewRows(syncEndpointColumns).
						AddRow(1, 1, "My WebDAV", "webdav", "https://dav.example.com",
							"user", "pass", "bidirectional", "/local", "/remote",
							nil, "active", now, now, nil))
			},
			wantCount: 1,
		},
		{
			name:     "empty result",
			syncType: "nonexistent",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_endpoints").
					WithArgs("nonexistent").
					WillReturnRows(sqlmock.NewRows(syncEndpointColumns))
			},
			wantCount: 0,
		},
		{
			name:     "database error",
			syncType: "ftp",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_endpoints").
					WithArgs("ftp").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo2(t)
			tt.setup(mock)

			endpoints, err := repo.GetEndpointsByType(tt.syncType)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, endpoints, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSyncRepository_GetStatistics(t *testing.T) {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	userID := 1

	tests := []struct {
		name    string
		userID  *int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, s *models.SyncStatistics)
	}{
		{
			name:   "success with user filter",
			userID: &userID,
			setup: func(mock sqlmock.Sqlmock) {
				// Status query
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
						AddRow("completed", 10).
						AddRow("failed", 2))
				// Type query
				mock.ExpectQuery("SELECT sync_type, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"sync_type", "count"}).
						AddRow("full", 8).
						AddRow("incremental", 4))
				// Files query
				mock.ExpectQuery("SELECT SUM").
					WillReturnRows(sqlmock.NewRows([]string{"total_synced", "total_failed"}).
						AddRow(500, 20))
				// Duration query
				mock.ExpectQuery("SELECT AVG").
					WillReturnRows(sqlmock.NewRows([]string{"avg_duration"}).
						AddRow(300.5))
			},
			check: func(t *testing.T, s *models.SyncStatistics) {
				assert.Equal(t, 12, s.TotalSessions)
				assert.Equal(t, 10, s.ByStatus["completed"])
				assert.Equal(t, 2, s.ByStatus["failed"])
				assert.Equal(t, 8, s.ByType["full"])
				assert.Equal(t, 500, s.TotalFilesSynced)
				assert.Equal(t, 20, s.TotalFilesFailed)
				assert.NotNil(t, s.AverageDuration)
				// Success rate: 10/(10+2) * 100 = 83.33%
				assert.InDelta(t, 83.33, s.SuccessRate, 0.1)
			},
		},
		{
			name:   "success without user filter",
			userID: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
						AddRow("completed", 5))
				mock.ExpectQuery("SELECT sync_type, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"sync_type", "count"}).
						AddRow("full", 5))
				mock.ExpectQuery("SELECT SUM").
					WillReturnRows(sqlmock.NewRows([]string{"total_synced", "total_failed"}).
						AddRow(nil, nil))
				mock.ExpectQuery("SELECT AVG").
					WillReturnRows(sqlmock.NewRows([]string{"avg_duration"}).
						AddRow(nil))
			},
			check: func(t *testing.T, s *models.SyncStatistics) {
				assert.Equal(t, 5, s.TotalSessions)
				assert.Equal(t, 0, s.TotalFilesSynced)
				assert.Nil(t, s.AverageDuration)
				assert.Equal(t, 100.0, s.SuccessRate) // Only completed, no failed
			},
		},
		{
			name:   "status query error",
			userID: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name:   "type query error",
			userID: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"status", "count"}))
				mock.ExpectQuery("SELECT sync_type, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo2(t)
			tt.setup(mock)

			stats, err := repo.GetStatistics(tt.userID, start, end)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, stats)
			tt.check(t, stats)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// ConversionRepository — GetStatistics (0% coverage)
// ===========================================================================

func newMockConversionRepo2(t *testing.T) (*ConversionRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewConversionRepository(db), mock
}

func TestConversionRepository_GetStatistics(t *testing.T) {
	start := time.Now().Add(-7 * 24 * time.Hour)
	end := time.Now()
	userID := 1

	tests := []struct {
		name    string
		userID  *int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, s *models.ConversionStatistics)
	}{
		{
			name:   "success with user filter",
			userID: &userID,
			setup: func(mock sqlmock.Sqlmock) {
				// Status query
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
						AddRow("completed", 15).
						AddRow("failed", 3))
				// Type query
				mock.ExpectQuery("SELECT conversion_type, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"conversion_type", "count"}).
						AddRow("video", 12).
						AddRow("audio", 6))
				// Format query
				mock.ExpectQuery("SELECT target_format, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"target_format", "count"}).
						AddRow("mp4", 10).
						AddRow("mkv", 5))
				// Duration query
				mock.ExpectQuery("SELECT AVG").
					WillReturnRows(sqlmock.NewRows([]string{"avg_duration"}).
						AddRow(120.0))
			},
			check: func(t *testing.T, s *models.ConversionStatistics) {
				assert.Equal(t, 18, s.TotalJobs)
				assert.Equal(t, 15, s.ByStatus["completed"])
				assert.Equal(t, 3, s.ByStatus["failed"])
				assert.Equal(t, 12, s.ByType["video"])
				assert.Equal(t, 10, s.ByFormat["mp4"])
				assert.NotNil(t, s.AverageDuration)
				// 15/(15+3) * 100 = 83.33%
				assert.InDelta(t, 83.33, s.SuccessRate, 0.1)
			},
		},
		{
			name:   "success without user filter — only completed",
			userID: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
						AddRow("completed", 5))
				mock.ExpectQuery("SELECT conversion_type, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"conversion_type", "count"}))
				mock.ExpectQuery("SELECT target_format, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"target_format", "count"}))
				mock.ExpectQuery("SELECT AVG").
					WillReturnRows(sqlmock.NewRows([]string{"avg_duration"}).AddRow(nil))
			},
			check: func(t *testing.T, s *models.ConversionStatistics) {
				assert.Equal(t, 5, s.TotalJobs)
				assert.Equal(t, 100.0, s.SuccessRate) // All completed, no failed
				assert.Nil(t, s.AverageDuration)
			},
		},
		{
			name:   "status query error",
			userID: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name:   "type query error",
			userID: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"status", "count"}))
				mock.ExpectQuery("SELECT conversion_type, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name:   "format query error",
			userID: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT status, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"status", "count"}))
				mock.ExpectQuery("SELECT conversion_type, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"conversion_type", "count"}))
				mock.ExpectQuery("SELECT target_format, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockConversionRepo2(t)
			tt.setup(mock)

			stats, err := repo.GetStatistics(tt.userID, start, end)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, stats)
			tt.check(t, stats)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// FileRepository — UpdateFilePath, UpdateDirectoryPaths,
//                   GetDirectoriesSortedBySize, GetDirectoriesSortedByDuplicates (all 0%)
// ===========================================================================

func newMockFileRepo2(t *testing.T) (*FileRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewFileRepository(db), mock
}

func TestFileRepository_UpdateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		fileID  int64
		newPath string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:    "success — subdirectory path",
			fileID:  1,
			newPath: "/docs/notes/readme.txt",
			setup: func(mock sqlmock.Sqlmock) {
				// Parent directory lookup
				mock.ExpectQuery("SELECT id FROM files WHERE path").
					WithArgs("/docs/notes").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(5)))
				// Update file record
				mock.ExpectExec("UPDATE files").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "success — root path",
			fileID:  2,
			newPath: "/readme.txt",
			setup: func(mock sqlmock.Sqlmock) {
				// No parent query for root path (dir is "/")
				mock.ExpectExec("UPDATE files").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "success — parent directory not found",
			fileID:  3,
			newPath: "/new-dir/file.txt",
			setup: func(mock sqlmock.Sqlmock) {
				// Parent lookup returns no rows
				mock.ExpectQuery("SELECT id FROM files WHERE path").
					WithArgs("/new-dir").
					WillReturnError(sql.ErrNoRows)
				// Update still succeeds (parentID is nil)
				mock.ExpectExec("UPDATE files").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "parent lookup error",
			fileID:  4,
			newPath: "/docs/file.txt",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM files WHERE path").
					WithArgs("/docs").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name:    "update error",
			fileID:  5,
			newPath: "/docs/file.txt",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM files WHERE path").
					WithArgs("/docs").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectExec("UPDATE files").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFileRepo2(t)
			tt.setup(mock)

			err := repo.UpdateFilePath(context.Background(), tt.fileID, tt.newPath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFileRepository_UpdateDirectoryPaths(t *testing.T) {
	tests := []struct {
		name    string
		oldPath string
		newPath string
		root    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:    "success — directory with children",
			oldPath: "/old-dir",
			newPath: "/new-dir",
			root:    "my-root",
			setup: func(mock sqlmock.Sqlmock) {
				// Query for files to update
				mock.ExpectQuery("SELECT id, path, is_directory").
					WithArgs("my-root", "/old-dir", "/old-dir/%").
					WillReturnRows(sqlmock.NewRows([]string{"id", "path", "is_directory"}).
						AddRow(int64(1), "/old-dir", true).
						AddRow(int64(2), "/old-dir/file.txt", false))
				// Update directory itself — root path
				mock.ExpectExec("UPDATE files").
					WillReturnResult(sqlmock.NewResult(0, 1))
				// Update child file — parent lookup + update
				mock.ExpectQuery("SELECT id FROM files WHERE path").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectExec("UPDATE files").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "success — empty directory (no files to update)",
			oldPath: "/empty",
			newPath: "/new-empty",
			root:    "my-root",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, path, is_directory").
					WithArgs("my-root", "/empty", "/empty/%").
					WillReturnRows(sqlmock.NewRows([]string{"id", "path", "is_directory"}))
			},
		},
		{
			name:    "query error",
			oldPath: "/old",
			newPath: "/new",
			root:    "my-root",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, path, is_directory").
					WithArgs("my-root", "/old", "/old/%").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFileRepo2(t)
			tt.setup(mock)

			err := repo.UpdateDirectoryPaths(context.Background(), tt.oldPath, tt.newPath, tt.root)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFileRepository_GetDirectoriesSortedBySize(t *testing.T) {
	now := time.Now()

	dirInfoColumns := []string{
		"path", "name", "storage_root_name", "file_count", "directory_count",
		"total_size", "duplicate_count", "modified_at",
	}

	tests := []struct {
		name      string
		root      string
		page      models.PaginationOptions
		ascending bool
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:      "returns directories descending",
			root:      "my-root",
			page:      models.PaginationOptions{Page: 1, Limit: 10},
			ascending: false,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT path, name, storage_root_name").
					WithArgs("my-root", 10, 0).
					WillReturnRows(sqlmock.NewRows(dirInfoColumns).
						AddRow("/videos", "videos", "my-root", int64(100), int64(5), int64(1073741824), int64(10), now).
						AddRow("/music", "music", "my-root", int64(500), int64(20), int64(536870912), int64(3), now))
			},
			wantCount: 2,
		},
		{
			name:      "returns directories ascending",
			root:      "my-root",
			page:      models.PaginationOptions{Page: 1, Limit: 5},
			ascending: true,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT path, name, storage_root_name").
					WithArgs("my-root", 5, 0).
					WillReturnRows(sqlmock.NewRows(dirInfoColumns))
			},
			wantCount: 0,
		},
		{
			name:      "database error",
			root:      "my-root",
			page:      models.PaginationOptions{Page: 1, Limit: 10},
			ascending: false,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT path, name, storage_root_name").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFileRepo2(t)
			tt.setup(mock)

			dirs, err := repo.GetDirectoriesSortedBySize(context.Background(), tt.root, tt.page, tt.ascending)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, dirs, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFileRepository_GetDirectoriesSortedByDuplicates(t *testing.T) {
	now := time.Now()

	dirInfoColumns := []string{
		"path", "name", "storage_root_name", "file_count", "directory_count",
		"total_size", "duplicate_count", "modified_at",
	}

	tests := []struct {
		name      string
		root      string
		page      models.PaginationOptions
		ascending bool
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:      "returns directories",
			root:      "my-root",
			page:      models.PaginationOptions{Page: 1, Limit: 10},
			ascending: false,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT path, name, storage_root_name").
					WithArgs("my-root", 10, 0).
					WillReturnRows(sqlmock.NewRows(dirInfoColumns).
						AddRow("/downloads", "downloads", "my-root", int64(200), int64(10), int64(1024000), int64(50), now))
			},
			wantCount: 1,
		},
		{
			name:      "ascending order",
			root:      "other-root",
			page:      models.PaginationOptions{Page: 2, Limit: 5},
			ascending: true,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT path, name, storage_root_name").
					WithArgs("other-root", 5, 5).
					WillReturnRows(sqlmock.NewRows(dirInfoColumns))
			},
			wantCount: 0,
		},
		{
			name:      "database error",
			root:      "my-root",
			page:      models.PaginationOptions{Page: 1, Limit: 10},
			ascending: false,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT path, name, storage_root_name").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFileRepo2(t)
			tt.setup(mock)

			dirs, err := repo.GetDirectoriesSortedByDuplicates(context.Background(), tt.root, tt.page, tt.ascending)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, dirs, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// LogManagementRepository — GetLogCollectionsByUser, GetRecentLogEntries,
//   CreateLogShare, GetLogShare, GetLogShareByToken, UpdateLogShare,
//   GetLogSharesByUser (all 0%)
// ===========================================================================

func newMockLogManagementRepo(t *testing.T) (*LogManagementRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewLogManagementRepository(db), mock
}

var logCollectionColumns = []string{
	"id", "user_id", "name", "description", "components", "log_level",
	"start_time", "end_time", "created_at", "completed_at", "status",
	"entry_count", "filters",
}

func TestLogManagementRepository_GetLogCollectionsByUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		userID    int
		limit     int
		offset    int
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "returns collections",
			userID: 1,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_collections").
					WithArgs(1, 10, 0).
					WillReturnRows(sqlmock.NewRows(logCollectionColumns).
						AddRow(1, 1, "Debug Logs", "API debug", `["api","web"]`, "debug",
							now, now, now, now, "completed", 100, `{"key":"value"}`))
			},
			wantCount: 1,
		},
		{
			name:   "empty result",
			userID: 99,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_collections").
					WithArgs(99, 10, 0).
					WillReturnRows(sqlmock.NewRows(logCollectionColumns))
			},
			wantCount: 0,
		},
		{
			name:   "database error",
			userID: 1,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_collections").
					WithArgs(1, 10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogManagementRepo(t)
			tt.setup(mock)

			collections, err := repo.GetLogCollectionsByUser(tt.userID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, collections, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

var logEntryColumns = []string{
	"id", "collection_id", "timestamp", "level", "component", "message", "context",
}

func TestLogManagementRepository_GetRecentLogEntries(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		component string
		limit     int
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:      "returns entries",
			component: "api",
			limit:     5,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_entries").
					WithArgs("api", 5).
					WillReturnRows(sqlmock.NewRows(logEntryColumns).
						AddRow(1, 10, now, "error", "api", "request failed", `{"path":"/api/v1/files"}`).
						AddRow(2, 10, now, "warn", "api", "slow query", `{}`))
			},
			wantCount: 2,
		},
		{
			name:      "empty result",
			component: "worker",
			limit:     10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_entries").
					WithArgs("worker", 10).
					WillReturnRows(sqlmock.NewRows(logEntryColumns))
			},
			wantCount: 0,
		},
		{
			name:      "database error",
			component: "api",
			limit:     5,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_entries").
					WithArgs("api", 5).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogManagementRepo(t)
			tt.setup(mock)

			entries, err := repo.GetRecentLogEntries(tt.component, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, entries, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

var logShareColumns = []string{
	"id", "collection_id", "user_id", "share_token", "share_type",
	"expires_at", "created_at", "accessed_at", "is_active",
	"permissions", "recipients",
}

func TestLogManagementRepository_CreateLogShare(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		wantID  int
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO log_shares").
					WillReturnResult(sqlmock.NewResult(42, 1))
			},
			wantID: 42,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO log_shares").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogManagementRepo(t)
			tt.setup(mock)

			share := &models.LogShare{
				CollectionID: 1,
				UserID:       1,
				ShareToken:   "tok-abc",
				ShareType:    "public",
				ExpiresAt:    now.Add(24 * time.Hour),
				CreatedAt:    now,
				IsActive:     true,
				Permissions:  []string{"read"},
				Recipients:   []string{"user@example.com"},
			}
			err := repo.CreateLogShare(share)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, share.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLogManagementRepository_GetLogShare(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, s *models.LogShare)
	}{
		{
			name: "success with accessed_at",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_shares WHERE id").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(logShareColumns).
						AddRow(1, 10, 1, "tok-xyz", "public", now.Add(24*time.Hour),
							now, now, true, `["read","download"]`, `["admin@test.com"]`))
			},
			check: func(t *testing.T, s *models.LogShare) {
				assert.Equal(t, 1, s.ID)
				assert.Equal(t, "tok-xyz", s.ShareToken)
				assert.NotNil(t, s.AccessedAt)
				assert.Contains(t, s.Permissions, "read")
				assert.Contains(t, s.Recipients, "admin@test.com")
			},
		},
		{
			name: "success without accessed_at",
			id:   2,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_shares WHERE id").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows(logShareColumns).
						AddRow(2, 10, 1, "tok-abc", "private", now.Add(24*time.Hour),
							now, nil, true, `["read"]`, `[]`))
			},
			check: func(t *testing.T, s *models.LogShare) {
				assert.Equal(t, 2, s.ID)
				assert.Nil(t, s.AccessedAt)
			},
		},
		{
			name: "database error",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_shares WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogManagementRepo(t)
			tt.setup(mock)

			share, err := repo.GetLogShare(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, share)
			tt.check(t, share)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLogManagementRepository_GetLogShareByToken(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		token   string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, s *models.LogShare)
	}{
		{
			name:  "success",
			token: "tok-share-abc",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_shares WHERE share_token").
					WithArgs("tok-share-abc").
					WillReturnRows(sqlmock.NewRows(logShareColumns).
						AddRow(5, 10, 1, "tok-share-abc", "public", now.Add(48*time.Hour),
							now, nil, true, `["read"]`, `[]`))
			},
			check: func(t *testing.T, s *models.LogShare) {
				assert.Equal(t, 5, s.ID)
				assert.Equal(t, "tok-share-abc", s.ShareToken)
			},
		},
		{
			name:  "not found",
			token: "nonexistent",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_shares WHERE share_token").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogManagementRepo(t)
			tt.setup(mock)

			share, err := repo.GetLogShareByToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, share)
			tt.check(t, share)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLogManagementRepository_UpdateLogShare(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		share   *models.LogShare
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			share: &models.LogShare{
				ID:          1,
				ShareType:   "private",
				ExpiresAt:   now.Add(72 * time.Hour),
				AccessedAt:  &now,
				IsActive:    true,
				Permissions: []string{"read", "download"},
				Recipients:  []string{"user@test.com"},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE log_shares SET").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			share: &models.LogShare{
				ID:          1,
				Permissions: []string{},
				Recipients:  []string{},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE log_shares SET").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogManagementRepo(t)
			tt.setup(mock)

			err := repo.UpdateLogShare(tt.share)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLogManagementRepository_GetLogSharesByUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		userID    int
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "returns shares",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_shares").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(logShareColumns).
						AddRow(1, 10, 1, "tok-1", "public", now.Add(24*time.Hour),
							now, nil, true, `["read"]`, `[]`).
						AddRow(2, 20, 1, "tok-2", "private", now.Add(48*time.Hour),
							now, now, false, `["read","write"]`, `["a@b.com"]`))
			},
			wantCount: 2,
		},
		{
			name:   "empty result",
			userID: 99,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_shares").
					WithArgs(99).
					WillReturnRows(sqlmock.NewRows(logShareColumns))
			},
			wantCount: 0,
		},
		{
			name:   "database error",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_shares").
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogManagementRepo(t)
			tt.setup(mock)

			shares, err := repo.GetLogSharesByUser(tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, shares, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// MediaFileRepository — GetDuplicateFiles (0% coverage)
// ===========================================================================

func newMockMediaFileRepo2(t *testing.T) (*MediaFileRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewMediaFileRepository(db), mock
}

func TestMediaFileRepository_GetDuplicateFiles(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
		check     func(t *testing.T, groups []DuplicateFileGroup)
	}{
		{
			name: "returns duplicate groups with item IDs",
			setup: func(mock sqlmock.Sqlmock) {
				// Main duplicate query
				mock.ExpectQuery("SELECT file_id, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"file_id", "item_count"}).
						AddRow(int64(100), int64(3)).
						AddRow(int64(200), int64(2)))
				// Item IDs for file 100
				mock.ExpectQuery("SELECT media_item_id").
					WithArgs(int64(100)).
					WillReturnRows(sqlmock.NewRows([]string{"media_item_id"}).
						AddRow(int64(1)).AddRow(int64(2)).AddRow(int64(3)))
				// Item IDs for file 200
				mock.ExpectQuery("SELECT media_item_id").
					WithArgs(int64(200)).
					WillReturnRows(sqlmock.NewRows([]string{"media_item_id"}).
						AddRow(int64(4)).AddRow(int64(5)))
			},
			wantCount: 2,
			check: func(t *testing.T, groups []DuplicateFileGroup) {
				assert.Equal(t, int64(100), groups[0].FileID)
				assert.Equal(t, int64(3), groups[0].ItemCount)
				assert.Len(t, groups[0].ItemIDs, 3)
				assert.Equal(t, int64(200), groups[1].FileID)
				assert.Len(t, groups[1].ItemIDs, 2)
			},
		},
		{
			name: "no duplicates",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT file_id, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"file_id", "item_count"}))
			},
			wantCount: 0,
		},
		{
			name: "main query error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT file_id, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name: "item IDs query error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT file_id, COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"file_id", "item_count"}).
						AddRow(int64(100), int64(2)))
				mock.ExpectQuery("SELECT media_item_id").
					WithArgs(int64(100)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaFileRepo2(t)
			tt.setup(mock)

			groups, err := repo.GetDuplicateFiles(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, groups, tt.wantCount)
			if tt.check != nil {
				tt.check(t, groups)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// UserMetadataRepository — Upsert (0% coverage)
// ===========================================================================

func newMockUserMetadataRepo2(t *testing.T) (*UserMetadataRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewUserMetadataRepository(db), mock
}

var userMetadataColumnsCov = []string{
	"id", "media_item_id", "user_id", "user_rating", "watched_status",
	"watched_date", "personal_notes", "tags", "favorite", "created_at", "updated_at",
}

func TestUserMetadataRepository_Upsert(t *testing.T) {
	now := time.Now()
	rating := 8.5
	status := "watched"

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "upsert — creates new when not existing",
			setup: func(mock sqlmock.Sqlmock) {
				// GetByItemAndUser returns no rows
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE media_item_id").
					WithArgs(int64(1), int64(2)).
					WillReturnError(sql.ErrNoRows)
				// Create via InsertReturningID
				mock.ExpectExec("INSERT INTO user_metadata").
					WillReturnResult(sqlmock.NewResult(10, 1))
			},
		},
		{
			name: "upsert — updates existing",
			setup: func(mock sqlmock.Sqlmock) {
				// GetByItemAndUser returns existing record
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE media_item_id").
					WithArgs(int64(1), int64(2)).
					WillReturnRows(sqlmock.NewRows(userMetadataColumnsCov).
						AddRow(int64(5), int64(1), int64(2), &rating, &status,
							nil, nil, `["action"]`, true, now, now))
				// Update
				mock.ExpectExec("UPDATE user_metadata SET").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "lookup error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE media_item_id").
					WithArgs(int64(1), int64(2)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockUserMetadataRepo2(t)
			tt.setup(mock)

			um := &mediamodels.UserMetadata{
				MediaItemID: 1,
				UserID:      2,
				UserRating:  &rating,
				Favorite:    true,
			}
			err := repo.Upsert(context.Background(), um)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// MediaItemRepository — additional coverage for Update, Search (edge cases),
//                        GetMediaTypes (error paths), SetPrimary
// ===========================================================================

func TestMediaItemRepository_Search_WithTypeFilter(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		query      string
		mediaTypes []int64
		limit      int
		offset     int
		setup      func(mock sqlmock.Sqlmock)
		wantErr    bool
		wantCount  int
		wantTotal  int64
	}{
		{
			name:       "search with type filter",
			query:      "action",
			mediaTypes: []int64{1, 2},
			limit:      10,
			offset:     0,
			setup: func(mock sqlmock.Sqlmock) {
				// Count query
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))
				// Data query
				row1 := sampleMediaItemRow(now)
				row1[0] = int64(1)
				row1[2] = "Action Movie"
				row2 := sampleMediaItemRow(now)
				row2[0] = int64(2)
				row2[2] = "Action Series"
				mock.ExpectQuery("SELECT .+ FROM media_items").
					WillReturnRows(sqlmock.NewRows(mediaItemColumns).
						AddRow(row1...).AddRow(row2...))
			},
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name:       "search count error",
			query:      "test",
			mediaTypes: nil,
			limit:      10,
			offset:     0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name:       "search data query error",
			query:      "test",
			mediaTypes: nil,
			limit:      10,
			offset:     0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
				mock.ExpectQuery("SELECT .+ FROM media_items").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			items, total, err := repo.Search(context.Background(), tt.query, tt.mediaTypes, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, items, tt.wantCount)
			assert.Equal(t, tt.wantTotal, total)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMediaItemRepository_GetMediaTypes_ErrorPaths(t *testing.T) {
	now := time.Now()

	mediaTypeColumns := []string{
		"id", "name", "description", "detection_patterns", "metadata_providers",
		"created_at", "updated_at",
	}

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, types []mediamodels.MediaType)
	}{
		{
			name: "success with null patterns",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_types").
					WillReturnRows(sqlmock.NewRows(mediaTypeColumns).
						AddRow(int64(1), "movie", "Movies", nil, nil, now, now).
						AddRow(int64(2), "tv_show", "TV Shows", `["*.mkv"]`, `["tmdb"]`, now, now))
			},
			check: func(t *testing.T, types []mediamodels.MediaType) {
				require.Len(t, types, 2)
				assert.Equal(t, "movie", types[0].Name)
				assert.Nil(t, types[0].DetectionPatterns)
				assert.Equal(t, "tv_show", types[1].Name)
				assert.Contains(t, types[1].DetectionPatterns, "*.mkv")
			},
		},
		{
			name: "success with invalid JSON patterns — gracefully nil",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_types").
					WillReturnRows(sqlmock.NewRows(mediaTypeColumns).
						AddRow(int64(1), "movie", "Movies", `{invalid`, `{invalid`, now, now))
			},
			check: func(t *testing.T, types []mediamodels.MediaType) {
				require.Len(t, types, 1)
				assert.Nil(t, types[0].DetectionPatterns)
				assert.Nil(t, types[0].MetadataProviders)
			},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_types").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaItemRepo(t)
			tt.setup(mock)

			types, err := repo.GetMediaTypes(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, types)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMediaItemRepository_GetMediaTypeByName_DBError(t *testing.T) {
	repo, mock := newMockMediaItemRepo(t)

	mock.ExpectQuery("SELECT .+ FROM media_types WHERE name").
		WithArgs("movie").
		WillReturnError(sql.ErrConnDone)

	mt, id, err := repo.GetMediaTypeByName(context.Background(), "movie")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get media type by name")
	assert.Nil(t, mt)
	assert.Equal(t, int64(0), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===========================================================================
// MediaFileRepository — SetPrimary (error paths)
// ===========================================================================

func TestMediaFileRepository_SetPrimary_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		mediaItemID int64
		fileID      int64
		setup       func(mock sqlmock.Sqlmock)
		wantErr     bool
		errIs       error
	}{
		{
			name:        "success",
			mediaItemID: 1,
			fileID:      10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WillReturnResult(sqlmock.NewResult(0, 3))
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:        "file not linked — no rows affected",
			mediaItemID: 1,
			fileID:      999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WillReturnResult(sqlmock.NewResult(0, 3))
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errIs:   sql.ErrNoRows,
		},
		{
			name:        "clear primary error",
			mediaItemID: 1,
			fileID:      10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name:        "set primary error",
			mediaItemID: 1,
			fileID:      10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WillReturnResult(sqlmock.NewResult(0, 3))
				mock.ExpectExec("UPDATE media_files SET is_primary").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockMediaFileRepo2(t)
			tt.setup(mock)

			err := repo.SetPrimary(context.Background(), tt.mediaItemID, tt.fileID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ===========================================================================
// MediaCollectionRepository — error paths (marshalJSONFieldString nil branch)
// ===========================================================================

func newMockMediaCollectionRepo2(t *testing.T) (*MediaCollectionRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewMediaCollectionRepository(db), mock
}

func TestMediaCollectionRepository_Create_DBError(t *testing.T) {
	repo, mock := newMockMediaCollectionRepo2(t)

	mock.ExpectExec("INSERT INTO media_collections").
		WillReturnError(sql.ErrConnDone)

	coll := &mediamodels.MediaCollection{
		Name:           "Test Collection",
		CollectionType: "manual",
	}

	_, err := repo.Create(context.Background(), coll)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create media collection")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaCollectionRepository_GetByID_NotFound(t *testing.T) {
	repo, mock := newMockMediaCollectionRepo2(t)

	mock.ExpectQuery("SELECT .+ FROM media_collections WHERE id").
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	coll, err := repo.GetByID(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "media collection not found")
	assert.Nil(t, coll)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaCollectionRepository_GetByID_DBError(t *testing.T) {
	repo, mock := newMockMediaCollectionRepo2(t)

	mock.ExpectQuery("SELECT .+ FROM media_collections WHERE id").
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	coll, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get media collection")
	assert.Nil(t, coll)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaCollectionRepository_Update_NotFound(t *testing.T) {
	repo, mock := newMockMediaCollectionRepo2(t)

	mock.ExpectExec("UPDATE media_collections SET").
		WillReturnResult(sqlmock.NewResult(0, 0))

	coll := &mediamodels.MediaCollection{
		ID:             999,
		Name:           "Updated",
		CollectionType: "manual",
	}
	err := repo.Update(context.Background(), coll)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "media collection not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaCollectionRepository_Update_DBError(t *testing.T) {
	repo, mock := newMockMediaCollectionRepo2(t)

	mock.ExpectExec("UPDATE media_collections SET").
		WillReturnError(sql.ErrConnDone)

	coll := &mediamodels.MediaCollection{
		ID:             1,
		Name:           "Updated",
		CollectionType: "manual",
	}
	err := repo.Update(context.Background(), coll)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update media collection")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaCollectionRepository_Delete_NotFound(t *testing.T) {
	repo, mock := newMockMediaCollectionRepo2(t)

	mock.ExpectExec("DELETE FROM media_collections WHERE id").
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "media collection not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaCollectionRepository_List_CountError(t *testing.T) {
	repo, mock := newMockMediaCollectionRepo2(t)

	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.List(context.Background(), 10, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count media collections")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaCollectionRepository_List_QueryError(t *testing.T) {
	repo, mock := newMockMediaCollectionRepo2(t)

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT .+ FROM media_collections").
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.List(context.Background(), 10, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list media collections")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===========================================================================
// StatsRepository — additional coverage for error paths
// ===========================================================================

func newMockStatsRepo2(t *testing.T) (*StatsRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewStatsRepository(db), mock
}

func TestStatsRepository_GetDuplicateStats_Error(t *testing.T) {
	repo, mock := newMockStatsRepo2(t)

	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrConnDone)

	stats, err := repo.GetDuplicateStats(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStatsRepository_GetDuplicateStats_WithStorageRoot(t *testing.T) {
	repo, mock := newMockStatsRepo2(t)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"total_duplicates", "duplicate_groups", "wasted_space",
			"largest_group", "average_group_size",
		}).AddRow(50, 10, int64(1073741824), 5, 3.5))

	stats, err := repo.GetDuplicateStats(context.Background(), "my-root")
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, int64(50), stats.TotalDuplicates)
	assert.Equal(t, int64(10), stats.DuplicateGroups)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===========================================================================
// MediaItemRepository — Create with genre and cast_crew JSON fields
// ===========================================================================

func TestMediaItemRepository_Create_WithGenreAndCastCrew(t *testing.T) {
	repo, mock := newMockMediaItemRepo(t)

	mock.ExpectExec("INSERT INTO media_items").
		WillReturnResult(sqlmock.NewResult(5, 1))

	director := "Director 1"
	item := &mediamodels.MediaItem{
		MediaTypeID: 1,
		Title:       "Movie With Metadata",
		Genre:       []string{"Action", "Sci-Fi"},
		CastCrew:    &mediamodels.CastCrew{Director: &director},
		Status:      "active",
	}
	id, err := item.ID, error(nil)
	id, err = repo.Create(context.Background(), item)
	require.NoError(t, err)
	assert.Equal(t, int64(5), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaItemRepository_Update_Success(t *testing.T) {
	repo, mock := newMockMediaItemRepo(t)

	mock.ExpectExec("UPDATE media_items SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	item := &mediamodels.MediaItem{
		ID:          1,
		MediaTypeID: 1,
		Title:       "Updated Movie",
		Genre:       []string{"Drama"},
		Status:      "active",
	}
	err := repo.Update(context.Background(), item)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaItemRepository_Update_Error(t *testing.T) {
	repo, mock := newMockMediaItemRepo(t)

	mock.ExpectExec("UPDATE media_items SET").
		WillReturnError(sql.ErrConnDone)

	item := &mediamodels.MediaItem{
		ID:     1,
		Title:  "Test",
		Status: "active",
	}
	err := repo.Update(context.Background(), item)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update media item")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===========================================================================
// MediaItemRepository — scanItem with cast_crew JSON parsing
// ===========================================================================

func TestMediaItemRepository_GetByID_WithCastCrewJSON(t *testing.T) {
	now := time.Now()
	repo, mock := newMockMediaItemRepo(t)

	row := []driver.Value{
		int64(1), int64(1), "Test Movie", nil, 2024, nil,
		`["Action"]`, nil, `{"director":"Nolan","writers":["Writer1"]}`, 9.0, 150, nil, nil,
		"active", nil, nil, nil, nil,
		now, now,
	}

	mock.ExpectQuery("SELECT .+ FROM media_items WHERE id").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(mediaItemColumns).AddRow(row...))

	item, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "Test Movie", item.Title)
	require.NotNil(t, item.Genre)
	assert.Contains(t, item.Genre, "Action")
	require.NotNil(t, item.CastCrew)
	require.NotNil(t, item.CastCrew.Director)
	assert.Equal(t, "Nolan", *item.CastCrew.Director)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===========================================================================
// MediaItemRepository — ListDuplicateGroups error path
// ===========================================================================

func TestMediaItemRepository_ListDuplicateGroups_CountError(t *testing.T) {
	repo, mock := newMockMediaItemRepo(t)

	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(sql.ErrConnDone)

	groups, total, err := repo.ListDuplicateGroups(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count duplicate groups")
	assert.Nil(t, groups)
	assert.Equal(t, int64(0), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaItemRepository_ListDuplicateGroups_QueryError(t *testing.T) {
	repo, mock := newMockMediaItemRepo(t)

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))
	mock.ExpectQuery("SELECT mi.title").
		WillReturnError(sql.ErrConnDone)

	groups, total, err := repo.ListDuplicateGroups(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list duplicate groups")
	assert.Nil(t, groups)
	assert.Equal(t, int64(0), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}
