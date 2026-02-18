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

func newMockFavoritesRepo(t *testing.T) (*FavoritesRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewFavoritesRepository(db), mock
}

var favoriteColumns = []string{
	"id", "user_id", "entity_type", "entity_id", "category",
	"notes", "tags", "is_public", "created_at", "updated_at",
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestFavoritesRepository_Constructor(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	repo := NewFavoritesRepository(db)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// CreateFavorite
// ---------------------------------------------------------------------------

func TestFavoritesRepository_CreateFavorite(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		fav     *models.Favorite
		setup   func(mock sqlmock.Sqlmock)
		wantID  int
		wantErr bool
	}{
		{
			name: "success without tags",
			fav: &models.Favorite{
				UserID:     1,
				EntityType: "media",
				EntityID:   42,
				IsPublic:   false,
				CreatedAt:  now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO favorites").
					WithArgs(1, "media", 42, nil, nil, nil, false, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantID: 1,
		},
		{
			name: "success with tags",
			fav: &models.Favorite{
				UserID:     1,
				EntityType: "media",
				EntityID:   42,
				Tags:       &[]string{"action", "favorite"},
				IsPublic:   true,
				CreatedAt:  now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO favorites").
					WithArgs(1, "media", 42, nil, nil, sqlmock.AnyArg(), true, now).
					WillReturnResult(sqlmock.NewResult(5, 1))
			},
			wantID: 5,
		},
		{
			name: "database error",
			fav: &models.Favorite{
				UserID:     1,
				EntityType: "media",
				EntityID:   42,
				CreatedAt:  now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO favorites").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFavoritesRepo(t)
			tt.setup(mock)

			id, err := repo.CreateFavorite(tt.fav)
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
// GetFavorite
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetFavorite(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		userID     int
		entityType string
		entityID   int
		setup      func(mock sqlmock.Sqlmock)
		wantNil    bool
		wantErr    bool
	}{
		{
			name:       "success",
			userID:     1,
			entityType: "media",
			entityID:   42,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(favoriteColumns).
					AddRow(1, 1, "media", 42, nil, nil, nil, false, now, nil)
				mock.ExpectQuery("SELECT .+ FROM favorites WHERE user_id").
					WithArgs(1, "media", 42).
					WillReturnRows(rows)
			},
		},
		{
			name:       "not found returns nil",
			userID:     1,
			entityType: "media",
			entityID:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM favorites WHERE user_id").
					WithArgs(1, "media", 999).
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name:       "database error",
			userID:     1,
			entityType: "media",
			entityID:   42,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM favorites WHERE user_id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFavoritesRepo(t)
			tt.setup(mock)

			fav, err := repo.GetFavorite(tt.userID, tt.entityType, tt.entityID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, fav)
			} else {
				assert.NotNil(t, fav)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetFavoriteByID
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetFavoriteByID(t *testing.T) {
	now := time.Now()

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
				rows := sqlmock.NewRows(favoriteColumns).
					AddRow(1, 1, "media", 42, nil, nil, nil, false, now, nil)
				mock.ExpectQuery("SELECT .+ FROM favorites WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM favorites WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFavoritesRepo(t)
			tt.setup(mock)

			fav, err := repo.GetFavoriteByID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, fav)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateFavorite
// ---------------------------------------------------------------------------

func TestFavoritesRepository_UpdateFavorite(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		fav     *models.Favorite
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			fav: &models.Favorite{
				ID:        1,
				IsPublic:  true,
				UpdatedAt: &now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE favorites").
					WithArgs(nil, nil, nil, true, &now, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			fav:  &models.Favorite{ID: 1},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE favorites").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFavoritesRepo(t)
			tt.setup(mock)

			err := repo.UpdateFavorite(tt.fav)
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
// DeleteFavorite
// ---------------------------------------------------------------------------

func TestFavoritesRepository_DeleteFavorite(t *testing.T) {
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
				mock.ExpectExec("DELETE FROM favorites WHERE id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM favorites WHERE id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFavoritesRepo(t)
			tt.setup(mock)

			err := repo.DeleteFavorite(tt.id)
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
// CountUserFavorites
// ---------------------------------------------------------------------------

func TestFavoritesRepository_CountUserFavorites(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		entityType *string
		setup      func(mock sqlmock.Sqlmock)
		wantCount  int
		wantErr    bool
	}{
		{
			name:   "count all favorites for user",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))
			},
			wantCount: 15,
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
			repo, mock := newMockFavoritesRepo(t)
			tt.setup(mock)

			count, err := repo.CountUserFavorites(tt.userID, tt.entityType)
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
// GetFavoritesCountByEntityType
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetFavoritesCountByEntityType(t *testing.T) {
	repo, mock := newMockFavoritesRepo(t)
	rows := sqlmock.NewRows([]string{"entity_type", "count"}).
		AddRow("media", 10).
		AddRow("share", 5)
	mock.ExpectQuery("SELECT entity_type, COUNT").
		WithArgs(1).
		WillReturnRows(rows)

	counts, err := repo.GetFavoritesCountByEntityType(1)
	require.NoError(t, err)
	assert.Equal(t, 10, counts["media"])
	assert.Equal(t, 5, counts["share"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetRecentFavorites
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetRecentFavorites(t *testing.T) {
	now := time.Now()

	repo, mock := newMockFavoritesRepo(t)
	rows := sqlmock.NewRows(favoriteColumns).
		AddRow(1, 1, "media", 42, nil, nil, nil, false, now, nil).
		AddRow(2, 1, "media", 43, nil, nil, nil, true, now, nil)
	mock.ExpectQuery("SELECT .+ FROM favorites WHERE user_id").
		WithArgs(1, 5).
		WillReturnRows(rows)

	favs, err := repo.GetRecentFavorites(1, 5)
	require.NoError(t, err)
	assert.Len(t, favs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// CreateFavoriteCategory
// ---------------------------------------------------------------------------

func TestFavoritesRepository_CreateFavoriteCategory(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		cat     *models.FavoriteCategory
		setup   func(mock sqlmock.Sqlmock)
		wantID  int
		wantErr bool
	}{
		{
			name: "success",
			cat: &models.FavoriteCategory{
				UserID:    1,
				Name:      "Movies",
				IsPublic:  true,
				CreatedAt: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO favorite_categories").
					WithArgs(1, "Movies", nil, nil, nil, true, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantID: 1,
		},
		{
			name: "database error",
			cat: &models.FavoriteCategory{
				UserID:    1,
				Name:      "Movies",
				CreatedAt: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO favorite_categories").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockFavoritesRepo(t)
			tt.setup(mock)

			id, err := repo.CreateFavoriteCategory(tt.cat)
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
// DeleteFavoriteCategory
// ---------------------------------------------------------------------------

func TestFavoritesRepository_DeleteFavoriteCategory(t *testing.T) {
	repo, mock := newMockFavoritesRepo(t)
	mock.ExpectExec("DELETE FROM favorite_categories WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteFavoriteCategory(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// RevokeFavoriteShare
// ---------------------------------------------------------------------------

func TestFavoritesRepository_RevokeFavoriteShare(t *testing.T) {
	repo, mock := newMockFavoritesRepo(t)
	mock.ExpectExec("UPDATE favorite_shares SET is_active = 0 WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RevokeFavoriteShare(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
