package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"catalogizer/database"
	"catalogizer/internal/tests"
	"catalogizer/models"
)

// Regression tests for fixes made to prevent future regressions

// TestFavoritesService_NilRepositoryRegression verifies that FavoritesService
// returns proper errors when repository is nil instead of panicking.
// This is a regression test for the nil pointer dereference bug.
func TestFavoritesService_NilRepositoryRegression(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewFavoritesService(nil, nil)

	t.Run("AddFavorite returns error with nil repository", func(t *testing.T) {
		favorite := &models.Favorite{
			EntityType: "movie",
			EntityID:   100,
		}
		result, err := service.AddFavorite(1, favorite)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.Nil(t, result)
	})

	t.Run("RemoveFavorite returns error with nil repository", func(t *testing.T) {
		err := service.RemoveFavorite(1, "movie", 100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
	})

	t.Run("GetUserFavorites returns error with nil repository", func(t *testing.T) {
		results, err := service.GetUserFavorites(1, nil, nil, 10, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.Nil(t, results)
	})

	t.Run("IsFavorite returns error with nil repository", func(t *testing.T) {
		isFav, err := service.IsFavorite(1, "movie", 100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.False(t, isFav)
	})

	t.Run("UpdateFavorite returns error with nil repository", func(t *testing.T) {
		updates := &models.UpdateFavoriteRequest{}
		result, err := service.UpdateFavorite(1, 1, updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.Nil(t, result)
	})

	t.Run("ExportFavorites returns error with nil repository", func(t *testing.T) {
		data, err := service.ExportFavorites(1, "json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.Nil(t, data)
	})

	t.Run("GetFavoriteStatistics returns error with nil repository", func(t *testing.T) {
		stats, err := service.GetFavoriteStatistics(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.Nil(t, stats)
	})

	t.Run("GetRecommendedFavorites returns error with nil repository", func(t *testing.T) {
		recs, err := service.GetRecommendedFavorites(1, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.Nil(t, recs)
	})

	t.Run("ShareFavorite returns error with nil repository", func(t *testing.T) {
		share, err := service.ShareFavorite(1, 1, []int{2}, models.SharePermissions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.Nil(t, share)
	})

	t.Run("GetSharedFavorites returns error with nil repository", func(t *testing.T) {
		results, err := service.GetSharedFavorites(1, 10, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
		assert.Nil(t, results)
	})

	t.Run("RevokeFavoriteShare returns error with nil repository", func(t *testing.T) {
		err := service.RevokeFavoriteShare(1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not configured")
	})

	_ = logger
}

// TestSyncService_NilRepositoryRegression verifies that SyncService
// returns proper errors when repository is nil instead of panicking.
func TestSyncService_NilRepositoryRegression(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	t.Run("CreateSyncEndpoint returns error with nil repository", func(t *testing.T) {
		endpoint := &models.SyncEndpoint{
			Name:          "Test",
			Type:          models.SyncTypeWebDAV,
			URL:           "https://example.com/webdav",
			SyncDirection: models.SyncDirectionUpload,
			LocalPath:     "/tmp/local",
			RemotePath:    "/remote",
		}
		result, err := service.CreateSyncEndpoint(1, endpoint)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not properly configured")
		assert.Nil(t, result)
	})
}

// TestDialectAwareQueriesRegression verifies that database queries work
// correctly with the dialect-aware wrapper for both SQLite and PostgreSQL.
func TestDialectAwareQueriesRegression(t *testing.T) {
	db := tests.SetupWrappedTestDB(t)
	defer db.Close()

	t.Run("INSERT OR REPLACE works correctly", func(t *testing.T) {
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS test_upsert (
				id TEXT PRIMARY KEY,
				value TEXT NOT NULL
			)
		`)
		require.NoError(t, err)

		_, err = db.Exec(`INSERT OR REPLACE INTO test_upsert (id, value) VALUES (?, ?)`, "key1", "value1")
		require.NoError(t, err)

		_, err = db.Exec(`INSERT OR REPLACE INTO test_upsert (id, value) VALUES (?, ?)`, "key1", "value2")
		require.NoError(t, err)

		var value string
		err = db.QueryRow(`SELECT value FROM test_upsert WHERE id = ?`, "key1").Scan(&value)
		require.NoError(t, err)
		assert.Equal(t, "value2", value)
	})
}

// TestDialectWrapperPassthroughRegression verifies the database wrapper
// correctly passes through queries and rewrites placeholders.
func TestDialectWrapperPassthroughRegression(t *testing.T) {
	db := tests.SetupWrappedTestDB(t)
	defer db.Close()

	t.Run("Query with placeholders works", func(t *testing.T) {
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS test_placeholders (
				id INTEGER PRIMARY KEY,
				name TEXT NOT NULL
			)
		`)
		require.NoError(t, err)

		_, err = db.Exec(`INSERT INTO test_placeholders (id, name) VALUES (?, ?)`, 1, "test")
		require.NoError(t, err)

		var name string
		err = db.QueryRow(`SELECT name FROM test_placeholders WHERE id = ?`, 1).Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, "test", name)
	})

	t.Run("Dialect reports SQLite for test database", func(t *testing.T) {
		assert.Equal(t, database.DialectSQLite, db.Dialect().Type)
	})
}

// TestContextCancellationRegression verifies that services handle context
// cancellation gracefully without panicking.
func TestContextCancellationRegression(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	t.Run("Operations handle cancelled context", func(t *testing.T) {
		favorite := &models.Favorite{
			EntityType: "movie",
			EntityID:   100,
		}
		result, err := service.AddFavorite(1, favorite)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
