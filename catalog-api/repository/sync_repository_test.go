package repository

import (
	"database/sql"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockSyncRepo(t *testing.T) (*SyncRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewSyncRepository(db), mock
}

var syncEndpointColumns = []string{
	"id", "user_id", "name", "type", "url", "username", "password", "sync_direction",
	"local_path", "remote_path", "sync_settings", "status", "created_at", "updated_at", "last_sync_at",
}

var syncSessionColumns = []string{
	"id", "endpoint_id", "user_id", "status", "sync_type", "started_at", "completed_at",
	"duration", "total_files", "synced_files", "failed_files", "skipped_files", "error_message",
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestSyncRepository_Constructor(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	repo := NewSyncRepository(db)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// CreateEndpoint
// ---------------------------------------------------------------------------

func TestSyncRepository_CreateEndpoint(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		ep      *models.SyncEndpoint
		setup   func(mock sqlmock.Sqlmock)
		wantID  int
		wantErr bool
	}{
		{
			name: "success",
			ep: &models.SyncEndpoint{
				UserID:        1,
				Name:          "My WebDAV",
				Type:          "webdav",
				URL:           "https://dav.example.com",
				Username:      "user",
				Password:      "pass",
				SyncDirection: "bidirectional",
				LocalPath:     "/local/media",
				RemotePath:    "/remote/media",
				Status:        "active",
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO sync_endpoints").
					WithArgs(1, "My WebDAV", "webdav", "https://dav.example.com",
						"user", "pass", "bidirectional", "/local/media", "/remote/media",
						nil, "active", now, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantID: 1,
		},
		{
			name: "database error",
			ep: &models.SyncEndpoint{
				UserID:    1,
				CreatedAt: now,
				UpdatedAt: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO sync_endpoints").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo(t)
			tt.setup(mock)

			id, err := repo.CreateEndpoint(tt.ep)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetEndpoint
// ---------------------------------------------------------------------------

func TestSyncRepository_GetEndpoint(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, ep *models.SyncEndpoint)
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(syncEndpointColumns).
					AddRow(1, 1, "WebDAV", "webdav", "https://dav.example.com",
						"user", "pass", "upload", "/local", "/remote",
						nil, "active", now, now, nil)
				mock.ExpectQuery("SELECT .+ FROM sync_endpoints WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, ep *models.SyncEndpoint) {
				assert.Equal(t, 1, ep.ID)
				assert.Equal(t, "WebDAV", ep.Name)
				assert.Equal(t, "webdav", ep.Type)
				assert.Nil(t, ep.LastSyncAt)
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_endpoints WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo(t)
			tt.setup(mock)

			ep, err := repo.GetEndpoint(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ep)
			tt.check(t, ep)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateEndpoint
// ---------------------------------------------------------------------------

func TestSyncRepository_UpdateEndpoint(t *testing.T) {
	now := time.Now()

	repo, mock := newMockSyncRepo(t)
	mock.ExpectExec("UPDATE sync_endpoints").
		WithArgs("Updated", "webdav", "https://dav.example.com", "user", "pass",
			"upload", "/local", "/remote", nil, "active", now, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateEndpoint(&models.SyncEndpoint{
		ID:            1,
		Name:          "Updated",
		Type:          "webdav",
		URL:           "https://dav.example.com",
		Username:      "user",
		Password:      "pass",
		SyncDirection: "upload",
		LocalPath:     "/local",
		RemotePath:    "/remote",
		Status:        "active",
		UpdatedAt:     now,
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// DeleteEndpoint
// ---------------------------------------------------------------------------

func TestSyncRepository_DeleteEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM sync_endpoints WHERE id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM sync_endpoints WHERE id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo(t)
			tt.setup(mock)

			err := repo.DeleteEndpoint(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// CreateSession
// ---------------------------------------------------------------------------

func TestSyncRepository_CreateSession(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		session *models.SyncSession
		setup   func(mock sqlmock.Sqlmock)
		wantID  int
		wantErr bool
	}{
		{
			name: "success",
			session: &models.SyncSession{
				EndpointID: 1,
				UserID:     1,
				Status:     "running",
				SyncType:   "full",
				StartedAt:  now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO sync_sessions").
					WithArgs(1, 1, "running", "full", now, 0, 0, 0, 0).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantID: 1,
		},
		{
			name: "database error",
			session: &models.SyncSession{
				EndpointID: 1,
				UserID:     1,
				StartedAt:  now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO sync_sessions").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo(t)
			tt.setup(mock)

			id, err := repo.CreateSession(tt.session)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetSession
// ---------------------------------------------------------------------------

func TestSyncRepository_GetSession(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, session *models.SyncSession)
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(syncSessionColumns).
					AddRow(1, 1, 1, "completed", "full", now, now,
						int64(120), 100, 95, 3, 2, nil)
				mock.ExpectQuery("SELECT .+ FROM sync_sessions WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, session *models.SyncSession) {
				assert.Equal(t, 1, session.ID)
				assert.Equal(t, "completed", session.Status)
				assert.Equal(t, 100, session.TotalFiles)
				assert.NotNil(t, session.Duration)
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_sessions WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo(t)
			tt.setup(mock)

			session, err := repo.GetSession(tt.id)
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

// ---------------------------------------------------------------------------
// CreateSchedule
// ---------------------------------------------------------------------------

func TestSyncRepository_CreateSchedule(t *testing.T) {
	now := time.Now()

	repo, mock := newMockSyncRepo(t)
	mock.ExpectExec("INSERT INTO sync_schedules").
		WithArgs(1, 1, "daily", true, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := repo.CreateSchedule(&models.SyncSchedule{
		EndpointID: 1,
		UserID:     1,
		Frequency:  "daily",
		IsActive:   true,
		CreatedAt:  now,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetActiveSchedules
// ---------------------------------------------------------------------------

func TestSyncRepository_GetActiveSchedules(t *testing.T) {
	now := time.Now()

	repo, mock := newMockSyncRepo(t)
	rows := sqlmock.NewRows([]string{"id", "endpoint_id", "user_id", "frequency", "last_run", "next_run", "is_active", "created_at"}).
		AddRow(1, 1, 1, "daily", nil, now.Add(24*time.Hour), true, now)
	mock.ExpectQuery("SELECT .+ FROM sync_schedules WHERE is_active = 1").
		WillReturnRows(rows)

	schedules, err := repo.GetActiveSchedules()
	require.NoError(t, err)
	assert.Len(t, schedules, 1)
	assert.Equal(t, "daily", schedules[0].Frequency)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// CleanupSessions
// ---------------------------------------------------------------------------

func TestSyncRepository_CleanupSessions(t *testing.T) {
	olderThan := time.Now().Add(-30 * 24 * time.Hour)

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM sync_sessions").
					WithArgs(olderThan).
					WillReturnResult(sqlmock.NewResult(0, 5))
			},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM sync_sessions").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo(t)
			tt.setup(mock)

			err := repo.CleanupSessions(olderThan)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserEndpoints
// ---------------------------------------------------------------------------

func TestSyncRepository_GetUserEndpoints(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		userID  int
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:   "returns endpoints",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(syncEndpointColumns).
					AddRow(1, 1, "WebDAV", "webdav", "https://dav.example.com",
						"user", "pass", "upload", "/local", "/remote",
						nil, "active", now, now, nil)
				mock.ExpectQuery("SELECT .+ FROM sync_endpoints WHERE user_id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:   "empty",
			userID: 2,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_endpoints WHERE user_id").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows(syncEndpointColumns))
			},
			want: 0,
		},
		{
			name:   "database error",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM sync_endpoints WHERE user_id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockSyncRepo(t)
			tt.setup(mock)

			eps, err := repo.GetUserEndpoints(tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, eps, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
