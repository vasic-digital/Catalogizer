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

func newMockCrashRepo(t *testing.T) (*CrashReportingRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewCrashReportingRepository(db), mock
}

var crashReportColumns = []string{
	"id", "user_id", "signal", "message", "stack_trace", "context",
	"system_info", "fingerprint", "status", "reported_at", "resolved_at",
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_Constructor(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	repo := NewCrashReportingRepository(db)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// CreateCrashReport
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_CreateCrashReport(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		report  *models.CrashReport
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			report: &models.CrashReport{
				UserID:      1,
				Signal:      "SIGSEGV",
				Message:     "segmentation fault",
				StackTrace:  "main.go:42",
				Context:     map[string]interface{}{"action": "parse"},
				SystemInfo:  map[string]interface{}{"os": "linux"},
				Fingerprint: "abc123",
				Status:      "new",
				ReportedAt:  now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO crash_reports").
					WithArgs(1, "SIGSEGV", "segmentation fault", "main.go:42",
						sqlmock.AnyArg(), sqlmock.AnyArg(), "abc123", "new", now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			report: &models.CrashReport{
				UserID:     1,
				Signal:     "SIGSEGV",
				Context:    map[string]interface{}{},
				SystemInfo: map[string]interface{}{},
				ReportedAt: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO crash_reports").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockCrashRepo(t)
			tt.setup(mock)

			err := repo.CreateCrashReport(tt.report)
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
// GetCrashReport
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_GetCrashReport(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, report *models.CrashReport)
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(crashReportColumns).
					AddRow(1, 1, "SIGSEGV", "seg fault", "stack",
						`{"action":"parse"}`, `{"os":"linux"}`,
						"abc123", "new", now, nil)
				mock.ExpectQuery("SELECT .+ FROM crash_reports WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, report *models.CrashReport) {
				assert.Equal(t, 1, report.ID)
				assert.Equal(t, "SIGSEGV", report.Signal)
				assert.Nil(t, report.ResolvedAt)
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM crash_reports WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockCrashRepo(t)
			tt.setup(mock)

			report, err := repo.GetCrashReport(tt.id)
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
// UpdateCrashReport
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_UpdateCrashReport(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		report  *models.CrashReport
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			report: &models.CrashReport{
				ID:          1,
				Signal:      "SIGSEGV",
				Message:     "updated message",
				StackTrace:  "main.go:42",
				Context:     map[string]interface{}{"action": "parse"},
				SystemInfo:  map[string]interface{}{"os": "linux"},
				Fingerprint: "abc123",
				Status:      "resolved",
				ResolvedAt:  &now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE crash_reports SET").
					WithArgs("SIGSEGV", "updated message", "main.go:42",
						sqlmock.AnyArg(), sqlmock.AnyArg(), "abc123", "resolved", &now, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			report: &models.CrashReport{
				ID:         1,
				Context:    map[string]interface{}{},
				SystemInfo: map[string]interface{}{},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE crash_reports SET").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockCrashRepo(t)
			tt.setup(mock)

			err := repo.UpdateCrashReport(tt.report)
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
// DeleteCrashReport
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_DeleteCrashReport(t *testing.T) {
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
				mock.ExpectExec("DELETE FROM crash_reports WHERE id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM crash_reports WHERE id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockCrashRepo(t)
			tt.setup(mock)

			err := repo.DeleteCrashReport(tt.id)
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
// GetRecentCrashCount
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_GetRecentCrashCount(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		setup     func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "returns count",
			duration: 24 * time.Hour,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
			},
			wantCount: 5,
		},
		{
			name:     "database error",
			duration: 24 * time.Hour,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockCrashRepo(t)
			tt.setup(mock)

			count, err := repo.GetRecentCrashCount(tt.duration)
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
// GetCrashReportsByUser
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_GetCrashReportsByUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		userID  int
		filters *models.CrashReportFilters
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:    "without filters",
			userID:  1,
			filters: nil,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(crashReportColumns).
					AddRow(1, 1, "SIGSEGV", "seg fault", "stack",
						`{}`, `{}`, "abc", "new", now, nil)
				mock.ExpectQuery("SELECT .+ FROM crash_reports").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:   "with signal filter",
			userID: 1,
			filters: &models.CrashReportFilters{
				Signal: "SIGSEGV",
				Limit:  10,
			},
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(crashReportColumns).
					AddRow(1, 1, "SIGSEGV", "seg fault", "stack",
						`{}`, `{}`, "abc", "new", now, nil)
				mock.ExpectQuery("SELECT .+ FROM crash_reports").
					WithArgs(1, "SIGSEGV", 10).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:    "database error",
			userID:  1,
			filters: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM crash_reports").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockCrashRepo(t)
			tt.setup(mock)

			reports, err := repo.GetCrashReportsByUser(tt.userID, tt.filters)
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

func TestCrashReportingRepository_CleanupOldReports(t *testing.T) {
	olderThan := time.Now().Add(-90 * 24 * time.Hour)

	repo, mock := newMockCrashRepo(t)
	mock.ExpectExec("DELETE FROM crash_reports WHERE reported_at").
		WithArgs(olderThan).
		WillReturnResult(sqlmock.NewResult(0, 15))

	err := repo.CleanupOldReports(olderThan)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetCrashesByFingerprint
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_GetCrashesByFingerprint(t *testing.T) {
	now := time.Now()

	repo, mock := newMockCrashRepo(t)
	rows := sqlmock.NewRows(crashReportColumns).
		AddRow(1, 1, "SIGSEGV", "seg fault", "stack",
			`{}`, `{}`, "abc123", "new", now, nil)
	mock.ExpectQuery("SELECT .+ FROM crash_reports WHERE fingerprint").
		WithArgs("abc123", 10).
		WillReturnRows(rows)

	reports, err := repo.GetCrashesByFingerprint("abc123", 10)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	assert.Equal(t, "abc123", reports[0].Fingerprint)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetTopCrashes
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_GetTopCrashes(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name: "returns top crashes",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"fingerprint", "count", "last_seen", "first_seen", "message", "signal"}).
					AddRow("fp1", 10, now, now.Add(-24*time.Hour), "crash 1", "SIGSEGV").
					AddRow("fp2", 5, now, now.Add(-48*time.Hour), "crash 2", "SIGABRT")
				mock.ExpectQuery("SELECT fingerprint, COUNT").
					WithArgs(1, sqlmock.AnyArg(), 10).
					WillReturnRows(rows)
			},
			want: 2,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT fingerprint, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockCrashRepo(t)
			tt.setup(mock)

			crashes, err := repo.GetTopCrashes(1, 10, 7*24*time.Hour)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, crashes, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetCrashStatistics
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_GetCrashStatistics(t *testing.T) {
	userID := 1

	t.Run("success", func(t *testing.T) {
		repo, mock := newMockCrashRepo(t)

		// Total crashes
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crash_reports WHERE user_id = \\?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))

		// Crashes by signal
		signalRows := sqlmock.NewRows([]string{"signal", "count"}).
			AddRow("SIGSEGV", 8).
			AddRow("SIGABRT", 5).
			AddRow("SIGFPE", 2)
		mock.ExpectQuery("SELECT signal, COUNT").
			WithArgs(userID).
			WillReturnRows(signalRows)

		// Recent crashes (with time parameter)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crash_reports").
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		// Resolved crashes
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crash_reports WHERE user_id = \\? AND status").
			WithArgs(userID, models.CrashStatusResolved).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(9))

		// Average resolution time
		mock.ExpectQuery("SELECT AVG").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"avg"}).AddRow(3.5))

		// Crash rate (with time parameter)
		mock.ExpectQuery("SELECT CAST\\(COUNT\\(\\*\\) AS FLOAT\\)").
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"crash_rate"}).AddRow(0.5))

		stats, err := repo.GetCrashStatistics(userID)
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, 15, stats.TotalCrashes)
		assert.Equal(t, 3, stats.RecentCrashes)
		assert.Equal(t, 9, stats.ResolvedCrashes)
		assert.Equal(t, 3.5, stats.AvgResolutionTime)
		assert.Equal(t, 0.5, stats.CrashRate)
		assert.Equal(t, map[string]int{"SIGSEGV": 8, "SIGABRT": 5, "SIGFPE": 2}, stats.CrashesBySignal)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on total crashes", func(t *testing.T) {
		repo, mock := newMockCrashRepo(t)

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crash_reports WHERE user_id").
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		stats, err := repo.GetCrashStatistics(userID)
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ---------------------------------------------------------------------------
// GetCrashTrends
// ---------------------------------------------------------------------------

func TestCrashReportingRepository_GetCrashTrends(t *testing.T) {
	userID := 1
	days := 7

	t.Run("success", func(t *testing.T) {
		repo, mock := newMockCrashRepo(t)

		rows := sqlmock.NewRows([]string{"date", "count"}).
			AddRow("2026-02-20", 5).
			AddRow("2026-02-21", 3).
			AddRow("2026-02-22", 7)
		mock.ExpectQuery("SELECT DATE\\(reported_at\\)").
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnRows(rows)

		trends, err := repo.GetCrashTrends(userID, days)
		require.NoError(t, err)
		assert.Len(t, trends, 3)
		assert.Equal(t, 5, trends[0].Count)
		assert.Equal(t, 3, trends[1].Count)
		assert.Equal(t, 7, trends[2].Count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		repo, mock := newMockCrashRepo(t)

		mock.ExpectQuery("SELECT DATE\\(reported_at\\)").
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		trends, err := repo.GetCrashTrends(userID, days)
		assert.Error(t, err)
		assert.Nil(t, trends)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
