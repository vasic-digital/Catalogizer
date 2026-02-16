package repository

import (
	"database/sql"
	"testing"
	"time"

	"catalogizer/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockErrorRepo(t *testing.T) (*ErrorReportingRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return NewErrorReportingRepository(db), mock
}

var errorReportColumns = []string{
	"id", "user_id", "level", "message", "error_code", "component",
	"stack_trace", "context", "system_info", "user_agent", "url",
	"fingerprint", "status", "reported_at", "resolved_at",
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_Constructor(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewErrorReportingRepository(db)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// CreateErrorReport
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_CreateErrorReport(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		report  *models.ErrorReport
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			report: &models.ErrorReport{
				UserID:      1,
				Level:       "error",
				Message:     "null pointer",
				ErrorCode:   "NPE001",
				Component:   "auth",
				StackTrace:  "auth.go:42",
				Context:     map[string]interface{}{"request_id": "req-123"},
				SystemInfo:  map[string]interface{}{"go_version": "1.21"},
				UserAgent:   "Mozilla/5.0",
				URL:         "/api/v1/login",
				Fingerprint: "fp-error-1",
				Status:      "new",
				ReportedAt:  now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO error_reports").
					WithArgs(1, "error", "null pointer", "NPE001", "auth", "auth.go:42",
						sqlmock.AnyArg(), sqlmock.AnyArg(), "Mozilla/5.0", "/api/v1/login",
						"fp-error-1", "new", now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			report: &models.ErrorReport{
				UserID:     1,
				Level:      "error",
				Context:    map[string]interface{}{},
				SystemInfo: map[string]interface{}{},
				ReportedAt: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO error_reports").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockErrorRepo(t)
			tt.setup(mock)

			err := repo.CreateErrorReport(tt.report)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, 1, tt.report.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetErrorReport
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_GetErrorReport(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, report *models.ErrorReport)
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(errorReportColumns).
					AddRow(1, 1, "error", "null pointer", "NPE001", "auth",
						"auth.go:42", `{"request_id":"req-123"}`, `{"go_version":"1.21"}`,
						"Mozilla/5.0", "/api/v1/login", "fp-1", "new", now, nil)
				mock.ExpectQuery("SELECT .+ FROM error_reports WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, report *models.ErrorReport) {
				assert.Equal(t, 1, report.ID)
				assert.Equal(t, "error", report.Level)
				assert.Equal(t, "auth", report.Component)
				assert.Nil(t, report.ResolvedAt)
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM error_reports WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockErrorRepo(t)
			tt.setup(mock)

			report, err := repo.GetErrorReport(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, report)
			tt.check(t, report)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateErrorReport
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_UpdateErrorReport(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		report  *models.ErrorReport
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			report: &models.ErrorReport{
				ID:          1,
				Level:       "error",
				Message:     "updated",
				Context:     map[string]interface{}{},
				SystemInfo:  map[string]interface{}{},
				Fingerprint: "fp-1",
				Status:      "resolved",
				ResolvedAt:  &now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE error_reports SET").
					WithArgs("error", "updated", "", "", "", sqlmock.AnyArg(), sqlmock.AnyArg(),
						"", "", "fp-1", "resolved", &now, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			report: &models.ErrorReport{
				ID:         1,
				Context:    map[string]interface{}{},
				SystemInfo: map[string]interface{}{},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE error_reports SET").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockErrorRepo(t)
			tt.setup(mock)

			err := repo.UpdateErrorReport(tt.report)
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
// DeleteErrorReport
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_DeleteErrorReport(t *testing.T) {
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
				mock.ExpectExec("DELETE FROM error_reports WHERE id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM error_reports WHERE id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockErrorRepo(t)
			tt.setup(mock)

			err := repo.DeleteErrorReport(tt.id)
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
// GetErrorCountInLastHour
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_GetErrorCountInLastHour(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		setup     func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "returns count",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
			},
			wantCount: 3,
		},
		{
			name:   "database error",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockErrorRepo(t)
			tt.setup(mock)

			count, err := repo.GetErrorCountInLastHour(tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetRecentErrorCount
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_GetRecentErrorCount(t *testing.T) {
	repo, mock := newMockErrorRepo(t)
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))

	count, err := repo.GetRecentErrorCount(24 * time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 12, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetErrorReportsByUser
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_GetErrorReportsByUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		userID  int
		filters *models.ErrorReportFilters
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:    "without filters",
			userID:  1,
			filters: nil,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(errorReportColumns).
					AddRow(1, 1, "error", "msg", "ERR", "comp", "stack",
						`{}`, `{}`, "agent", "/url", "fp", "new", now, nil)
				mock.ExpectQuery("SELECT .+ FROM error_reports").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:   "with level filter",
			userID: 1,
			filters: &models.ErrorReportFilters{
				Level: "error",
				Limit: 10,
			},
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(errorReportColumns).
					AddRow(1, 1, "error", "msg", "ERR", "comp", "stack",
						`{}`, `{}`, "agent", "/url", "fp", "new", now, nil)
				mock.ExpectQuery("SELECT .+ FROM error_reports").
					WithArgs(1, "error", 10).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:    "database error",
			userID:  1,
			filters: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM error_reports").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockErrorRepo(t)
			tt.setup(mock)

			reports, err := repo.GetErrorReportsByUser(tt.userID, tt.filters)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, reports, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// CleanupOldReports
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_CleanupOldReports(t *testing.T) {
	olderThan := time.Now().Add(-90 * 24 * time.Hour)

	repo, mock := newMockErrorRepo(t)
	mock.ExpectExec("DELETE FROM error_reports WHERE reported_at").
		WithArgs(olderThan).
		WillReturnResult(sqlmock.NewResult(0, 20))

	err := repo.CleanupOldReports(olderThan)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetTopErrors
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_GetTopErrors(t *testing.T) {
	now := time.Now()

	repo, mock := newMockErrorRepo(t)
	rows := sqlmock.NewRows([]string{"fingerprint", "count", "last_seen", "first_seen", "message", "component", "level"}).
		AddRow("fp1", 20, now, now.Add(-24*time.Hour), "error 1", "auth", "error").
		AddRow("fp2", 10, now, now.Add(-48*time.Hour), "error 2", "api", "warning")
	mock.ExpectQuery("SELECT fingerprint, COUNT").
		WithArgs(1, sqlmock.AnyArg(), 10).
		WillReturnRows(rows)

	errors, err := repo.GetTopErrors(1, 10, 7*24*time.Hour)
	require.NoError(t, err)
	assert.Len(t, errors, 2)
	assert.Equal(t, "fp1", errors[0].Fingerprint)
	assert.Equal(t, 20, errors[0].Count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetErrorsByFingerprint
// ---------------------------------------------------------------------------

func TestErrorReportingRepository_GetErrorsByFingerprint(t *testing.T) {
	now := time.Now()

	repo, mock := newMockErrorRepo(t)
	rows := sqlmock.NewRows(errorReportColumns).
		AddRow(1, 1, "error", "msg", "ERR", "comp", "stack",
			`{}`, `{}`, "agent", "/url", "fp-abc", "new", now, nil)
	mock.ExpectQuery("SELECT .+ FROM error_reports WHERE fingerprint").
		WithArgs("fp-abc", 10).
		WillReturnRows(rows)

	reports, err := repo.GetErrorsByFingerprint("fp-abc", 10)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	assert.Equal(t, "fp-abc", reports[0].Fingerprint)
	assert.NoError(t, mock.ExpectationsWereMet())
}
