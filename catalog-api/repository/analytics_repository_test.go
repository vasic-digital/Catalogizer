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

func newMockAnalyticsRepo(t *testing.T) (*AnalyticsRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return NewAnalyticsRepository(db), mock
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_Constructor(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewAnalyticsRepository(db)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// LogMediaAccess
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_LogMediaAccess(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		access  *models.MediaAccessLog
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success without optional fields",
			access: &models.MediaAccessLog{
				UserID:     1,
				MediaID:    42,
				Action:     "play",
				AccessTime: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO media_access_logs").
					WithArgs(1, 42, "play", nil, nil, nil, nil, nil, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "success with device info and location",
			access: &models.MediaAccessLog{
				UserID:     1,
				MediaID:    42,
				Action:     "play",
				DeviceInfo: &models.DeviceInfo{},
				Location:   &models.Location{Latitude: 40.7, Longitude: -74.0},
				AccessTime: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO media_access_logs").
					WithArgs(1, 42, "play", sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, nil, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			access: &models.MediaAccessLog{
				UserID:     1,
				MediaID:    42,
				Action:     "play",
				AccessTime: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO media_access_logs").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockAnalyticsRepo(t)
			tt.setup(mock)

			err := repo.LogMediaAccess(tt.access)
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
// LogEvent
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_LogEvent(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		event   *models.AnalyticsEvent
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			event: &models.AnalyticsEvent{
				UserID:        1,
				EventType:     "media_access",
				EventCategory: "playback",
				Data:          `{"file_type":"video"}`,
				Timestamp:     now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO analytics_events").
					WithArgs(1, "media_access", "playback", `{"file_type":"video"}`, nil, nil, nil, nil, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			event: &models.AnalyticsEvent{
				UserID:    1,
				EventType: "test",
				Timestamp: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO analytics_events").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockAnalyticsRepo(t)
			tt.setup(mock)

			err := repo.LogEvent(tt.event)
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
// GetMediaAccessLogs
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_GetMediaAccessLogs(t *testing.T) {
	now := time.Now()
	accessLogColumns := []string{
		"id", "user_id", "media_id", "action", "device_info", "location",
		"ip_address", "user_agent", "playback_duration", "access_time",
	}

	tests := []struct {
		name    string
		userID  int
		mediaID *int
		limit   int
		offset  int
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:   "all logs with no filters",
			userID: 0,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(accessLogColumns).
					AddRow(1, 1, 42, "play", nil, nil, nil, nil, nil, now)
				mock.ExpectQuery("SELECT .+ FROM media_access_logs").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:   "filter by user ID",
			userID: 1,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(accessLogColumns).
					AddRow(1, 1, 42, "play", nil, nil, nil, nil, nil, now)
				mock.ExpectQuery("SELECT .+ FROM media_access_logs").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:   "database error",
			userID: 0,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM media_access_logs").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockAnalyticsRepo(t)
			tt.setup(mock)

			logs, err := repo.GetMediaAccessLogs(tt.userID, tt.mediaID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, logs, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetTotalUsers
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_GetTotalUsers(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name: "returns count",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))
			},
			wantCount: 25,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockAnalyticsRepo(t)
			tt.setup(mock)

			count, err := repo.GetTotalUsers()
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
// GetActiveUsers
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_GetActiveUsers(t *testing.T) {
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	tests := []struct {
		name      string
		setup     func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name: "returns active users count",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(DISTINCT user_id\\)").
					WithArgs(startDate, endDate).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			wantCount: 10,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(DISTINCT user_id\\)").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockAnalyticsRepo(t)
			tt.setup(mock)

			count, err := repo.GetActiveUsers(startDate, endDate)
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
// GetTotalMediaAccesses
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_GetTotalMediaAccesses(t *testing.T) {
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	repo, mock := newMockAnalyticsRepo(t)
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

	count, err := repo.GetTotalMediaAccesses(startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 100, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetTotalEvents
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_GetTotalEvents(t *testing.T) {
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	repo, mock := newMockAnalyticsRepo(t)
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))

	count, err := repo.GetTotalEvents(startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 50, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetTopAccessedMedia
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_GetTopAccessedMedia(t *testing.T) {
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name: "returns top accessed media",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"media_id", "access_count"}).
					AddRow(42, 100).
					AddRow(43, 50)
				mock.ExpectQuery("SELECT media_id, COUNT").
					WithArgs(startDate, endDate, 10).
					WillReturnRows(rows)
			},
			want: 2,
		},
		{
			name: "empty result",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT media_id, COUNT").
					WithArgs(startDate, endDate, 10).
					WillReturnRows(sqlmock.NewRows([]string{"media_id", "access_count"}))
			},
			want: 0,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT media_id, COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockAnalyticsRepo(t)
			tt.setup(mock)

			results, err := repo.GetTopAccessedMedia(startDate, endDate, 10)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, results, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetFileTypeData
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_GetFileTypeData(t *testing.T) {
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		want    map[string]int
		wantErr bool
	}{
		{
			name: "returns file type data",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"file_type", "count"}).
					AddRow("video", 50).
					AddRow("audio", 30)
				mock.ExpectQuery("SELECT").
					WithArgs(startDate, endDate).
					WillReturnRows(rows)
			},
			want: map[string]int{"video": 50, "audio": 30},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockAnalyticsRepo(t)
			tt.setup(mock)

			result, err := repo.GetFileTypeData(startDate, endDate)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetSessionData
// ---------------------------------------------------------------------------

func TestAnalyticsRepository_GetSessionData(t *testing.T) {
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()
	sessionStart := time.Now().Add(-2 * time.Hour)
	sessionEnd := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name: "returns session data",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"user_id", "session_start", "session_end", "duration_seconds"}).
					AddRow(1, sessionStart, sessionEnd, 3600.0)
				mock.ExpectQuery("SELECT user_id").
					WithArgs(startDate, endDate).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT user_id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockAnalyticsRepo(t)
			tt.setup(mock)

			results, err := repo.GetSessionData(startDate, endDate)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, results, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
