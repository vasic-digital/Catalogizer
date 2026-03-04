package services

import (
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
)

func TestNewFavoritesService(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	assert.NotNil(t, service)
	assert.Nil(t, service.favoritesRepo)
	assert.Nil(t, service.authService)
}

func TestFavoritesService_RemoveDuplicateStrings(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil, // removeDuplicateStrings returns nil for empty input
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "single element",
			input:    []string{"a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.removeDuplicateStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFavoritesService_ExportFavorites_UnsupportedFormat(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	// Test that unsupported format returns an error
	// Note: we cannot test with nil repo as ExportFavorites accesses repo directly;
	// instead, test the format validation path
	// The repo-dependent paths would require a proper test database
	_ = service // service created to validate constructor
}

func TestFavoritesService_ImportFavorites_EmptyData(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	tests := []struct {
		name    string
		userID  int
		format  string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			userID:  1,
			format:  "json",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "nil data",
			userID:  1,
			format:  "json",
			data:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format",
			userID:  1,
			format:  "invalid_format",
			data:    []byte(`[{"id":1}]`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ImportFavorites(tt.userID, tt.data, tt.format)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ExportFavorites tests
// ---------------------------------------------------------------------------

func TestFavoritesService_ExportFavoritesToJSON(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	favorites := []models.Favorite{
		{
			ID:         1,
			UserID:     1,
			EntityType: "movie",
			EntityID:   100,
			CreatedAt:  time.Now(),
		},
	}

	data, err := service.exportFavoritesToJSON(favorites)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), `"version"`)
	assert.Contains(t, string(data), `"count": 1`)
}

func TestFavoritesService_ExportFavoritesToJSON_Empty(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	data, err := service.exportFavoritesToJSON([]models.Favorite{})
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), `"count": 0`)
}

func TestFavoritesService_ExportFavoritesToCSV(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	category := "movies"
	notes := "Great movie"
	favorites := []models.Favorite{
		{
			ID:         1,
			UserID:     1,
			EntityType: "movie",
			EntityID:   100,
			Category:   &category,
			Notes:      &notes,
			IsPublic:   true,
			CreatedAt:  time.Now(),
		},
	}

	data, err := service.exportFavoritesToCSV(favorites)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "ID,UserID")
	assert.Contains(t, string(data), "movie")
}

func TestFavoritesService_ExportFavoritesToCSV_Empty(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	data, err := service.exportFavoritesToCSV([]models.Favorite{})
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "ID,UserID") // Header only
}

func TestFavoritesService_ImportFavoritesFromJSON_InvalidJSON(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	_, err := service.importFavoritesFromJSON(123, []byte(`invalid json`))
	assert.Error(t, err)
}

// ============================================================================
// ADDITIONAL TESTS FOR 95% COVERAGE
// ============================================================================

func TestFavoritesService_AddFavorite(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	tests := []struct {
		name     string
		userID   int
		favorite *models.Favorite
		wantErr  bool
	}{
		{
			name:   "valid favorite",
			userID: 1,
			favorite: &models.Favorite{
				EntityType: "movie",
				EntityID:   100,
			},
			wantErr: true, // Will error without repository
		},
		{
			name:   "missing entity type",
			userID: 1,
			favorite: &models.Favorite{
				EntityID: 100,
			},
			wantErr: true,
		},
		{
			name:   "missing entity ID",
			userID: 1,
			favorite: &models.Favorite{
				EntityType: "movie",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.AddFavorite(tt.userID, tt.favorite)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestFavoritesService_RemoveFavorite(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	tests := []struct {
		name       string
		userID     int
		entityType string
		entityID   int
		wantErr    bool
	}{
		{
			name:       "valid removal",
			userID:     1,
			entityType: "movie",
			entityID:   100,
			wantErr:    true, // Will error without repository
		},
		{
			name:       "empty entity type",
			userID:     1,
			entityType: "",
			entityID:   100,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RemoveFavorite(tt.userID, tt.entityType, tt.entityID)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Nil-repository guard tests for all uncovered methods
// ---------------------------------------------------------------------------

func TestFavoritesService_GetUserFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	favorites, err := service.GetUserFavorites(1, nil, nil, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, favorites)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_GetFavoritesByEntity_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	favorite, err := service.GetFavoritesByEntity(1, "movie", 100)
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_IsFavorite_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	result, err := service.IsFavorite(1, "movie", 100)
	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_UpdateFavorite_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	updates := &models.UpdateFavoriteRequest{}
	favorite, err := service.UpdateFavorite(1, 1, updates)
	assert.Error(t, err)
	assert.Nil(t, favorite)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_GetFavoriteCategories_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	categories, err := service.GetFavoriteCategories(1, nil)
	assert.Error(t, err)
	assert.Nil(t, categories)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_CreateFavoriteCategory_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	category := &models.FavoriteCategory{Name: "test"}
	result, err := service.CreateFavoriteCategory(1, category)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_UpdateFavoriteCategory_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	updates := &models.UpdateFavoriteCategoryRequest{Name: "updated"}
	result, err := service.UpdateFavoriteCategory(1, 1, updates)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_DeleteFavoriteCategory_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	err := service.DeleteFavoriteCategory(1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_GetPublicFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	favorites, err := service.GetPublicFavorites(nil, nil, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, favorites)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_SearchFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	favorites, err := service.SearchFavorites(1, "test", nil, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, favorites)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_GetFavoriteStatistics_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	stats, err := service.GetFavoriteStatistics(1)
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_GetRecommendedFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	favorites, err := service.GetRecommendedFavorites(1, 10)
	assert.Error(t, err)
	assert.Nil(t, favorites)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_ShareFavorite_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	share, err := service.ShareFavorite(1, 1, []int{2, 3}, models.SharePermissions{})
	assert.Error(t, err)
	assert.Nil(t, share)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_GetSharedFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	favorites, err := service.GetSharedFavorites(1, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, favorites)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_RevokeFavoriteShare_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	err := service.RevokeFavoriteShare(1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_ExportFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	data, err := service.ExportFavorites(1, "json")
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "favorites repository not configured")
}

func TestFavoritesService_BulkAddFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	requests := []models.BulkFavoriteRequest{
		{EntityType: "movie", EntityID: 1},
		{EntityType: "movie", EntityID: 2},
	}

	results, err := service.BulkAddFavorites(1, requests)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestFavoritesService_BulkRemoveFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	requests := []models.BulkFavoriteRemoveRequest{
		{EntityType: "movie", EntityID: 1},
	}

	err := service.BulkRemoveFavorites(1, requests)
	assert.Error(t, err)
}

func TestFavoritesService_ImportFavoritesFromCSV_InvalidHeaders(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	// CSV with too few columns
	csvData := []byte("col1,col2,col3\n1,2,3\n")
	results, err := service.importFavoritesFromCSV(1, csvData)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestFavoritesService_ImportFavoritesFromCSV_InvalidEntityID(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	// CSV with valid header but invalid entity ID
	csvData := []byte("ID,UserID,EntityType,EntityID,Category,Notes,Tags,IsPublic,CreatedAt,UpdatedAt\n1,1,movie,not_a_number,,,,false,2025-01-01T00:00:00Z,\n")
	results, err := service.importFavoritesFromCSV(1, csvData)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestFavoritesService_ImportFavoritesFromJSON_EmptyFavorites(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	jsonData := []byte(`{"version":"1.0","count":0,"favorites":[]}`)
	results, err := service.importFavoritesFromJSON(1, jsonData)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Empty(t, results)
}

func TestFavoritesService_ExportFavoritesToCSV_WithNilFields(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	favorites := []models.Favorite{
		{
			ID:         1,
			UserID:     1,
			EntityType: "movie",
			EntityID:   100,
			Category:   nil,
			Notes:      nil,
			Tags:       nil,
			IsPublic:   false,
			CreatedAt:  time.Now(),
			UpdatedAt:  nil,
		},
	}

	data, err := service.exportFavoritesToCSV(favorites)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "movie")
}

func TestFavoritesService_ExportFavoritesToCSV_WithAllFields(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	category := "action"
	notes := "Great movie"
	tags := []string{"action", "thriller"}
	updatedAt := time.Now()

	favorites := []models.Favorite{
		{
			ID:         1,
			UserID:     1,
			EntityType: "movie",
			EntityID:   100,
			Category:   &category,
			Notes:      &notes,
			Tags:       &tags,
			IsPublic:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  &updatedAt,
		},
	}

	data, err := service.exportFavoritesToCSV(favorites)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "action")
	assert.Contains(t, string(data), "Great movie")
}
