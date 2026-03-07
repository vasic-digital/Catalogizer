package repository

import (
	"database/sql"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/models"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mutecomm/go-sqlcipher"
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

// ===========================================================================
// Real SQLite-backed tests for uncovered functions
// ===========================================================================

func newRealFavoritesRepo(t *testing.T) *FavoritesRepository {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	_, err = sqlDB.Exec(`
		CREATE TABLE favorites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER NOT NULL,
			category TEXT,
			notes TEXT,
			tags TEXT,
			is_public INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME
		)
	`)
	require.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE favorite_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			color TEXT,
			icon TEXT,
			entity_type TEXT,
			is_public INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME
		)
	`)
	require.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE favorite_shares (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			favorite_id INTEGER NOT NULL,
			shared_by_user INTEGER NOT NULL,
			shared_with TEXT NOT NULL,
			permissions TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		)
	`)
	require.NoError(t, err)

	return NewFavoritesRepository(db)
}

func seedFavorites(t *testing.T, repo *FavoritesRepository) {
	t.Helper()
	now := time.Now().Truncate(time.Second)
	tags := &[]string{"action", "thriller"}

	favorites := []models.Favorite{
		{UserID: 1, EntityType: "movie", EntityID: 100, Category: strPtr("watch"), Notes: strPtr("great movie"), Tags: tags, IsPublic: true, CreatedAt: now},
		{UserID: 1, EntityType: "movie", EntityID: 101, Category: strPtr("watch"), Notes: strPtr("good film"), IsPublic: false, CreatedAt: now.Add(-time.Hour)},
		{UserID: 1, EntityType: "music", EntityID: 200, Category: strPtr("listen"), Notes: strPtr("awesome song"), IsPublic: true, CreatedAt: now.Add(-2 * time.Hour)},
		{UserID: 2, EntityType: "movie", EntityID: 100, IsPublic: true, CreatedAt: now.Add(-3 * time.Hour)},
		{UserID: 2, EntityType: "music", EntityID: 201, IsPublic: false, CreatedAt: now.Add(-4 * time.Hour)},
	}

	for i := range favorites {
		_, err := repo.CreateFavorite(&favorites[i])
		require.NoError(t, err)
	}
}

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// GetUserFavorites
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetUserFavorites_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)
	seedFavorites(t, repo)

	t.Run("all favorites for user 1", func(t *testing.T) {
		favs, err := repo.GetUserFavorites(1, nil, nil, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 3)
	})

	t.Run("filter by entity type", func(t *testing.T) {
		et := "movie"
		favs, err := repo.GetUserFavorites(1, &et, nil, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 2)
		for _, f := range favs {
			assert.Equal(t, "movie", f.EntityType)
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		cat := "listen"
		favs, err := repo.GetUserFavorites(1, nil, &cat, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 1)
		assert.Equal(t, "music", favs[0].EntityType)
	})

	t.Run("pagination", func(t *testing.T) {
		favs, err := repo.GetUserFavorites(1, nil, nil, 2, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 2)

		favs2, err := repo.GetUserFavorites(1, nil, nil, 2, 2)
		require.NoError(t, err)
		assert.Len(t, favs2, 1)
	})

	t.Run("empty result", func(t *testing.T) {
		favs, err := repo.GetUserFavorites(999, nil, nil, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, favs)
	})
}

// ---------------------------------------------------------------------------
// GetPublicFavorites
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetPublicFavorites_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)
	seedFavorites(t, repo)

	t.Run("all public favorites", func(t *testing.T) {
		favs, err := repo.GetPublicFavorites(nil, nil, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 3)
		for _, f := range favs {
			assert.True(t, f.IsPublic)
		}
	})

	t.Run("filter by entity type", func(t *testing.T) {
		et := "movie"
		favs, err := repo.GetPublicFavorites(&et, nil, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 2)
	})

	t.Run("filter by category", func(t *testing.T) {
		cat := "watch"
		favs, err := repo.GetPublicFavorites(nil, &cat, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 1)
	})

	t.Run("pagination", func(t *testing.T) {
		favs, err := repo.GetPublicFavorites(nil, nil, 2, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 2)
	})
}

// ---------------------------------------------------------------------------
// SearchFavorites
// ---------------------------------------------------------------------------

func TestFavoritesRepository_SearchFavorites_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)
	seedFavorites(t, repo)

	t.Run("search by notes", func(t *testing.T) {
		favs, err := repo.SearchFavorites(1, "great", nil, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 1)
		assert.Equal(t, 100, favs[0].EntityID)
	})

	t.Run("search by tags", func(t *testing.T) {
		favs, err := repo.SearchFavorites(1, "action", nil, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 1)
	})

	t.Run("search with entity type filter", func(t *testing.T) {
		et := "movie"
		favs, err := repo.SearchFavorites(1, "good", &et, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 1)
	})

	t.Run("no results", func(t *testing.T) {
		favs, err := repo.SearchFavorites(1, "nonexistent", nil, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, favs)
	})
}

// ---------------------------------------------------------------------------
// GetFavoriteCategories
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetFavoriteCategories_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)

	now := time.Now().Truncate(time.Second)
	cats := []models.FavoriteCategory{
		{UserID: 1, Name: "Action", IsPublic: true, CreatedAt: now},
		{UserID: 1, Name: "Comedy", IsPublic: false, CreatedAt: now},
		{UserID: 2, Name: "Drama", IsPublic: true, CreatedAt: now},
	}
	for i := range cats {
		_, err := repo.CreateFavoriteCategory(&cats[i])
		require.NoError(t, err)
	}

	t.Run("get categories for user 1", func(t *testing.T) {
		result, err := repo.GetFavoriteCategories(1, nil)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		// ordered by name ASC
		assert.Equal(t, "Action", result[0].Name)
		assert.Equal(t, "Comedy", result[1].Name)
	})

	t.Run("empty result for non-existent user", func(t *testing.T) {
		result, err := repo.GetFavoriteCategories(999, nil)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// ---------------------------------------------------------------------------
// GetFavoritesCountByCategory
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetFavoritesCountByCategory_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)
	seedFavorites(t, repo)

	counts, err := repo.GetFavoritesCountByCategory(1)
	require.NoError(t, err)
	assert.Equal(t, 2, counts["watch"])
	assert.Equal(t, 1, counts["listen"])
}

// ---------------------------------------------------------------------------
// GetSimilarFavorites
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetSimilarFavorites_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)
	seedFavorites(t, repo)

	t.Run("find similar public movie favorites from other users", func(t *testing.T) {
		favs, err := repo.GetSimilarFavorites(1, "movie", 10)
		require.NoError(t, err)
		assert.Len(t, favs, 1)
		assert.Equal(t, 2, favs[0].UserID)
	})

	t.Run("no similar for music from user 2", func(t *testing.T) {
		// User 2's music fav is not public, user 1's music fav is public
		favs, err := repo.GetSimilarFavorites(2, "music", 10)
		require.NoError(t, err)
		assert.Len(t, favs, 1) // user 1's public music
	})
}

// ---------------------------------------------------------------------------
// GetFavoriteCategoryByID
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetFavoriteCategoryByID_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)

	now := time.Now().Truncate(time.Second)
	cat := &models.FavoriteCategory{UserID: 1, Name: "Favorites", IsPublic: true, CreatedAt: now}
	id, err := repo.CreateFavoriteCategory(cat)
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		result, err := repo.GetFavoriteCategoryByID(id)
		require.NoError(t, err)
		assert.Equal(t, "Favorites", result.Name)
		assert.Equal(t, 1, result.UserID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetFavoriteCategoryByID(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category not found")
	})
}

// ---------------------------------------------------------------------------
// UpdateFavoriteCategory
// ---------------------------------------------------------------------------

func TestFavoritesRepository_UpdateFavoriteCategory_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)

	now := time.Now().Truncate(time.Second)
	cat := &models.FavoriteCategory{UserID: 1, Name: "Old", IsPublic: false, CreatedAt: now}
	id, err := repo.CreateFavoriteCategory(cat)
	require.NoError(t, err)

	updated := time.Now().Truncate(time.Second)
	cat.ID = id
	cat.Name = "New"
	cat.IsPublic = true
	cat.UpdatedAt = &updated

	err = repo.UpdateFavoriteCategory(cat)
	require.NoError(t, err)

	result, err := repo.GetFavoriteCategoryByID(id)
	require.NoError(t, err)
	assert.Equal(t, "New", result.Name)
	assert.True(t, result.IsPublic)
}

// ---------------------------------------------------------------------------
// CountFavoritesByCategory
// ---------------------------------------------------------------------------

func TestFavoritesRepository_CountFavoritesByCategory_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)

	now := time.Now().Truncate(time.Second)
	cat := &models.FavoriteCategory{UserID: 1, Name: "watch", IsPublic: false, CreatedAt: now}
	catID, err := repo.CreateFavoriteCategory(cat)
	require.NoError(t, err)

	seedFavorites(t, repo)

	count, err := repo.CountFavoritesByCategory(catID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// ---------------------------------------------------------------------------
// CreateFavoriteShare
// ---------------------------------------------------------------------------

func TestFavoritesRepository_CreateFavoriteShare_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)
	seedFavorites(t, repo)

	now := time.Now().Truncate(time.Second)
	share := &models.FavoriteShare{
		FavoriteID:   1,
		SharedByUser: 1,
		SharedWith:   []int{2, 3},
		Permissions:  models.SharePermissions{CanView: true, CanEdit: false},
		CreatedAt:    now,
		IsActive:     true,
	}

	id, err := repo.CreateFavoriteShare(share)
	require.NoError(t, err)
	assert.Greater(t, id, 0)
}

// ---------------------------------------------------------------------------
// GetFavoriteShareByID
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetFavoriteShareByID_Real(t *testing.T) {
	repo := newRealFavoritesRepo(t)
	seedFavorites(t, repo)

	now := time.Now().Truncate(time.Second)
	share := &models.FavoriteShare{
		FavoriteID:   1,
		SharedByUser: 1,
		SharedWith:   []int{2, 3},
		Permissions:  models.SharePermissions{CanView: true, CanEdit: true},
		CreatedAt:    now,
		IsActive:     true,
	}

	id, err := repo.CreateFavoriteShare(share)
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		result, err := repo.GetFavoriteShareByID(id)
		require.NoError(t, err)
		assert.Equal(t, 1, result.FavoriteID)
		assert.Equal(t, 1, result.SharedByUser)
		assert.Equal(t, []int{2, 3}, result.SharedWith)
		assert.True(t, result.Permissions.CanView)
		assert.True(t, result.Permissions.CanEdit)
		assert.True(t, result.IsActive)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetFavoriteShareByID(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "share not found")
	})
}

// ---------------------------------------------------------------------------
// GetSharedFavorites (uses JSON_EXTRACT, requires sqlmock)
// ---------------------------------------------------------------------------

func TestFavoritesRepository_GetSharedFavorites_Mock(t *testing.T) {
	now := time.Now()

	t.Run("returns shared favorites", func(t *testing.T) {
		repo, mock := newMockFavoritesRepo(t)
		rows := sqlmock.NewRows(favoriteColumns).
			AddRow(1, 1, "movie", 100, nil, nil, nil, true, now, nil)
		mock.ExpectQuery("SELECT .+ FROM favorites f").
			WithArgs(sqlmock.AnyArg(), 10, 0).
			WillReturnRows(rows)

		favs, err := repo.GetSharedFavorites(2, 10, 0)
		require.NoError(t, err)
		assert.Len(t, favs, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock := newMockFavoritesRepo(t)
		mock.ExpectQuery("SELECT .+ FROM favorites f").
			WithArgs(sqlmock.AnyArg(), 10, 0).
			WillReturnRows(sqlmock.NewRows(favoriteColumns))

		favs, err := repo.GetSharedFavorites(3, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, favs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		repo, mock := newMockFavoritesRepo(t)
		mock.ExpectQuery("SELECT .+ FROM favorites f").
			WillReturnError(sql.ErrConnDone)

		_, err := repo.GetSharedFavorites(2, 10, 0)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
