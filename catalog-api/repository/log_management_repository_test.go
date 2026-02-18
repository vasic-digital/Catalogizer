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

func newMockLogMgmtRepo(t *testing.T) (*LogManagementRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewLogManagementRepository(db), mock
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestLogManagementRepository_Constructor(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	repo := NewLogManagementRepository(db)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// CreateLogCollection
// ---------------------------------------------------------------------------

func TestLogManagementRepository_CreateLogCollection(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		coll    *models.LogCollection
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			coll: &models.LogCollection{
				UserID:      1,
				Name:        "debug-logs",
				Description: "Debug log collection",
				Components:  []string{"auth", "api"},
				LogLevel:    "debug",
				CreatedAt:   now,
				Status:      "pending",
				Filters:     map[string]interface{}{"key": "value"},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO log_collections").
					WithArgs(1, "debug-logs", "Debug log collection",
						sqlmock.AnyArg(), "debug", nil, nil, now, "pending", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			coll: &models.LogCollection{
				UserID:     1,
				Name:       "test",
				Components: []string{},
				Filters:    map[string]interface{}{},
				CreatedAt:  now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO log_collections").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogMgmtRepo(t)
			tt.setup(mock)

			err := repo.CreateLogCollection(tt.coll)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, 1, tt.coll.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetLogCollection
// ---------------------------------------------------------------------------

func TestLogManagementRepository_GetLogCollection(t *testing.T) {
	now := time.Now()

	logCollectionColumns := []string{
		"id", "user_id", "name", "description", "components", "log_level",
		"start_time", "end_time", "created_at", "completed_at", "status",
		"entry_count", "filters",
	}

	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, coll *models.LogCollection)
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(logCollectionColumns).
					AddRow(1, 1, "debug-logs", "desc", `["auth","api"]`, "debug",
						nil, nil, now, nil, "pending", 0, `{}`)
				mock.ExpectQuery("SELECT .+ FROM log_collections WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, coll *models.LogCollection) {
				assert.Equal(t, 1, coll.ID)
				assert.Equal(t, "debug-logs", coll.Name)
				assert.Contains(t, coll.Components, "auth")
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_collections WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogMgmtRepo(t)
			tt.setup(mock)

			coll, err := repo.GetLogCollection(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, coll)
			tt.check(t, coll)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateLogCollection
// ---------------------------------------------------------------------------

func TestLogManagementRepository_UpdateLogCollection(t *testing.T) {
	tests := []struct {
		name    string
		coll    *models.LogCollection
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			coll: &models.LogCollection{
				ID:         1,
				Name:       "updated",
				Components: []string{"auth"},
				Filters:    map[string]interface{}{},
				Status:     "completed",
				EntryCount: 100,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE log_collections SET").
					WithArgs("updated", "", sqlmock.AnyArg(), "",
						nil, nil, nil, "completed", 100, sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			coll: &models.LogCollection{
				ID:         1,
				Components: []string{},
				Filters:    map[string]interface{}{},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE log_collections SET").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogMgmtRepo(t)
			tt.setup(mock)

			err := repo.UpdateLogCollection(tt.coll)
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
// DeleteLogCollection
// ---------------------------------------------------------------------------

func TestLogManagementRepository_DeleteLogCollection(t *testing.T) {
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
				// First deletes entries
				mock.ExpectExec("DELETE FROM log_entries WHERE collection_id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 50))
				// Then deletes collection
				mock.ExpectExec("DELETE FROM log_collections WHERE id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "error deleting entries",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM log_entries WHERE collection_id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogMgmtRepo(t)
			tt.setup(mock)

			err := repo.DeleteLogCollection(tt.id)
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
// CreateLogEntry
// ---------------------------------------------------------------------------

func TestLogManagementRepository_CreateLogEntry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		entry   *models.LogEntry
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			entry: &models.LogEntry{
				CollectionID: 1,
				Timestamp:    now,
				Level:        "error",
				Component:    "auth",
				Message:      "login failed",
				Context:      map[string]interface{}{"user": "test"},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO log_entries").
					WithArgs(1, now, "error", "auth", "login failed", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			entry: &models.LogEntry{
				CollectionID: 1,
				Context:      map[string]interface{}{},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO log_entries").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogMgmtRepo(t)
			tt.setup(mock)

			err := repo.CreateLogEntry(tt.entry)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, 1, tt.entry.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetLogEntries
// ---------------------------------------------------------------------------

func TestLogManagementRepository_GetLogEntries(t *testing.T) {
	now := time.Now()
	logEntryColumns := []string{"id", "collection_id", "timestamp", "level", "component", "message", "context"}

	tests := []struct {
		name    string
		collID  int
		filters *models.LogEntryFilters
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:    "without filters",
			collID:  1,
			filters: nil,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(logEntryColumns).
					AddRow(1, 1, now, "error", "auth", "login failed", `{"user":"test"}`)
				mock.ExpectQuery("SELECT .+ FROM log_entries WHERE collection_id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:   "with level filter",
			collID: 1,
			filters: &models.LogEntryFilters{
				Level: "error",
				Limit: 10,
			},
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(logEntryColumns).
					AddRow(1, 1, now, "error", "auth", "login failed", `{}`)
				mock.ExpectQuery("SELECT .+ FROM log_entries WHERE collection_id").
					WithArgs(1, "error", 10).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:    "database error",
			collID:  1,
			filters: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM log_entries WHERE collection_id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogMgmtRepo(t)
			tt.setup(mock)

			entries, err := repo.GetLogEntries(tt.collID, tt.filters)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, entries, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteLogEntriesByCollection
// ---------------------------------------------------------------------------

func TestLogManagementRepository_DeleteLogEntriesByCollection(t *testing.T) {
	repo, mock := newMockLogMgmtRepo(t)
	mock.ExpectExec("DELETE FROM log_entries WHERE collection_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 50))

	err := repo.DeleteLogEntriesByCollection(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// CleanupOldCollections
// ---------------------------------------------------------------------------

func TestLogManagementRepository_CleanupOldCollections(t *testing.T) {
	olderThan := time.Now().Add(-30 * 24 * time.Hour)

	repo, mock := newMockLogMgmtRepo(t)
	mock.ExpectExec("DELETE FROM log_collections WHERE created_at").
		WithArgs(olderThan).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.CleanupOldCollections(olderThan)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// CleanupExpiredShares
// ---------------------------------------------------------------------------

func TestLogManagementRepository_CleanupExpiredShares(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE log_shares SET is_active = 0").
					WillReturnResult(sqlmock.NewResult(0, 2))
			},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE log_shares SET is_active = 0").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockLogMgmtRepo(t)
			tt.setup(mock)

			err := repo.CleanupExpiredShares()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
