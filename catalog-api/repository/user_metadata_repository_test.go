package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/internal/media/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockUserMetadataRepo creates a UserMetadataRepository backed by sqlmock.
func newMockUserMetadataRepo(t *testing.T) (*UserMetadataRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewUserMetadataRepository(db), mock
}

// userMetadataColumns is the standard column set for user_metadata queries.
var userMetadataColumns = []string{
	"id", "media_item_id", "user_id", "user_rating", "watched_status",
	"watched_date", "personal_notes", "tags", "favorite",
	"created_at", "updated_at",
}

func sampleUserMetadataRow(now time.Time) []driver.Value {
	rating := 8.5
	status := "watched"
	return []driver.Value{
		int64(1), int64(10), int64(1), &rating, &status,
		nil, nil, `["action","scifi"]`, true,
		now, now,
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestUserMetadataRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		wantID  int64
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO user_metadata").
					WillReturnResult(sqlmock.NewResult(3, 1))
			},
			wantID: 3,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO user_metadata").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockUserMetadataRepo(t)
			tt.setup(mock)

			rating := 8.5
			um := &models.UserMetadata{
				MediaItemID: 10,
				UserID:      1,
				UserRating:  &rating,
				Favorite:    true,
				Tags:        []string{"action"},
			}
			id, err := repo.Create(context.Background(), um)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantID, um.ID)
			assert.False(t, um.CreatedAt.IsZero())
			assert.False(t, um.UpdatedAt.IsZero())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetByItemAndUser
// ---------------------------------------------------------------------------

func TestUserMetadataRepository_GetByItemAndUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		itemID  int64
		userID  int64
		setup   func(mock sqlmock.Sqlmock)
		wantNil bool
		wantErr bool
		check   func(t *testing.T, um *models.UserMetadata)
	}{
		{
			name:   "found",
			itemID: 10,
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE media_item_id").
					WithArgs(int64(10), int64(1)).
					WillReturnRows(sqlmock.NewRows(userMetadataColumns).
						AddRow(sampleUserMetadataRow(now)...))
			},
			check: func(t *testing.T, um *models.UserMetadata) {
				assert.Equal(t, int64(10), um.MediaItemID)
				assert.Equal(t, int64(1), um.UserID)
				assert.True(t, um.Favorite)
				require.NotNil(t, um.UserRating)
				assert.Equal(t, 8.5, *um.UserRating)
				assert.Contains(t, um.Tags, "action")
				assert.Contains(t, um.Tags, "scifi")
			},
		},
		{
			name:   "not found returns nil",
			itemID: 99,
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE media_item_id").
					WithArgs(int64(99), int64(1)).
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name:   "database error",
			itemID: 10,
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE media_item_id").
					WithArgs(int64(10), int64(1)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockUserMetadataRepo(t)
			tt.setup(mock)

			um, err := repo.GetByItemAndUser(context.Background(), tt.itemID, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, um)
			} else {
				require.NotNil(t, um)
				tt.check(t, um)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUserMetadataRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE user_metadata SET").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE user_metadata SET").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockUserMetadataRepo(t)
			tt.setup(mock)

			um := &models.UserMetadata{
				ID:       1,
				Favorite: true,
				Tags:     []string{"updated"},
			}
			err := repo.Update(context.Background(), um)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.False(t, um.UpdatedAt.IsZero())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetFavorites
// ---------------------------------------------------------------------------

func TestUserMetadataRepository_GetFavorites(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		userID    int64
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "returns favorites",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				row1 := sampleUserMetadataRow(now)
				row2 := sampleUserMetadataRow(now)
				row2[0] = int64(2)
				row2[1] = int64(20)
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE user_id").
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows(userMetadataColumns).
						AddRow(row1...).
						AddRow(row2...))
			},
			wantCount: 2,
		},
		{
			name:   "empty result",
			userID: 99,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE user_id").
					WithArgs(int64(99)).
					WillReturnRows(sqlmock.NewRows(userMetadataColumns))
			},
			wantCount: 0,
		},
		{
			name:   "database error",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE user_id").
					WithArgs(int64(1)).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockUserMetadataRepo(t)
			tt.setup(mock)

			items, err := repo.GetFavorites(context.Background(), tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, items, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetByWatchedStatus
// ---------------------------------------------------------------------------

func TestUserMetadataRepository_GetByWatchedStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		userID    int64
		status    string
		setup     func(mock sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "returns results",
			userID: 1,
			status: "watched",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE user_id").
					WithArgs(int64(1), "watched").
					WillReturnRows(sqlmock.NewRows(userMetadataColumns).
						AddRow(sampleUserMetadataRow(now)...))
			},
			wantCount: 1,
		},
		{
			name:   "empty result",
			userID: 1,
			status: "plan_to_watch",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM user_metadata WHERE user_id").
					WithArgs(int64(1), "plan_to_watch").
					WillReturnRows(sqlmock.NewRows(userMetadataColumns))
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockUserMetadataRepo(t)
			tt.setup(mock)

			items, err := repo.GetByWatchedStatus(context.Background(), tt.userID, tt.status)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, items, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
