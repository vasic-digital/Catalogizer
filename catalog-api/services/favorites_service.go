package services

import (
	"fmt"
	"time"

	"catalogizer/models"
	"catalogizer/repository"
)

type FavoritesService struct {
	favoritesRepo *repository.FavoritesRepository
	authService   *AuthService
}

func NewFavoritesService(favoritesRepo *repository.FavoritesRepository, authService *AuthService) *FavoritesService {
	return &FavoritesService{
		favoritesRepo: favoritesRepo,
		authService:   authService,
	}
}

func (s *FavoritesService) AddFavorite(userID int, favorite *models.Favorite) (*models.Favorite, error) {
	existing, err := s.favoritesRepo.GetFavorite(userID, favorite.EntityType, favorite.EntityID)
	if err == nil && existing != nil {
		return existing, fmt.Errorf("item already in favorites")
	}

	favorite.UserID = userID
	favorite.CreatedAt = time.Now()

	id, err := s.favoritesRepo.CreateFavorite(favorite)
	if err != nil {
		return nil, fmt.Errorf("failed to add favorite: %w", err)
	}

	favorite.ID = id
	return favorite, nil
}

func (s *FavoritesService) RemoveFavorite(userID int, entityType string, entityID int) error {
	favorite, err := s.favoritesRepo.GetFavorite(userID, entityType, entityID)
	if err != nil {
		return fmt.Errorf("favorite not found: %w", err)
	}

	if favorite.UserID != userID {
		return fmt.Errorf("unauthorized to remove this favorite")
	}

	return s.favoritesRepo.DeleteFavorite(favorite.ID)
}

func (s *FavoritesService) GetUserFavorites(userID int, entityType *string, category *string, limit, offset int) ([]models.Favorite, error) {
	return s.favoritesRepo.GetUserFavorites(userID, entityType, category, limit, offset)
}

func (s *FavoritesService) GetFavoritesByEntity(userID int, entityType string, entityID int) (*models.Favorite, error) {
	return s.favoritesRepo.GetFavorite(userID, entityType, entityID)
}

func (s *FavoritesService) IsFavorite(userID int, entityType string, entityID int) (bool, error) {
	favorite, err := s.favoritesRepo.GetFavorite(userID, entityType, entityID)
	if err != nil {
		return false, nil
	}
	return favorite != nil, nil
}

func (s *FavoritesService) UpdateFavorite(userID int, favoriteID int, updates *models.UpdateFavoriteRequest) (*models.Favorite, error) {
	favorite, err := s.favoritesRepo.GetFavoriteByID(favoriteID)
	if err != nil {
		return nil, fmt.Errorf("favorite not found: %w", err)
	}

	if favorite.UserID != userID {
		return nil, fmt.Errorf("unauthorized to update this favorite")
	}

	if updates.Category != nil {
		favorite.Category = updates.Category
	}

	if updates.Notes != nil {
		favorite.Notes = updates.Notes
	}

	if updates.Tags != nil {
		favorite.Tags = updates.Tags
	}

	if updates.IsPublic != nil {
		favorite.IsPublic = *updates.IsPublic
	}

	favorite.UpdatedAt = &time.Time{}
	*favorite.UpdatedAt = time.Now()

	err = s.favoritesRepo.UpdateFavorite(favorite)
	if err != nil {
		return nil, fmt.Errorf("failed to update favorite: %w", err)
	}

	return favorite, nil
}

func (s *FavoritesService) GetFavoriteCategories(userID int, entityType *string) ([]models.FavoriteCategory, error) {
	return s.favoritesRepo.GetFavoriteCategories(userID, entityType)
}

func (s *FavoritesService) CreateFavoriteCategory(userID int, category *models.FavoriteCategory) (*models.FavoriteCategory, error) {
	category.UserID = userID
	category.CreatedAt = time.Now()

	id, err := s.favoritesRepo.CreateFavoriteCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	category.ID = id
	return category, nil
}

func (s *FavoritesService) UpdateFavoriteCategory(userID int, categoryID int, updates *models.UpdateFavoriteCategoryRequest) (*models.FavoriteCategory, error) {
	category, err := s.favoritesRepo.GetFavoriteCategoryByID(categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	if category.UserID != userID {
		return nil, fmt.Errorf("unauthorized to update this category")
	}

	if updates.Name != "" {
		category.Name = updates.Name
	}

	if updates.Description != nil {
		category.Description = updates.Description
	}

	if updates.Color != nil {
		category.Color = updates.Color
	}

	if updates.Icon != nil {
		category.Icon = updates.Icon
	}

	if updates.IsPublic != nil {
		category.IsPublic = *updates.IsPublic
	}

	category.UpdatedAt = &time.Time{}
	*category.UpdatedAt = time.Now()

	err = s.favoritesRepo.UpdateFavoriteCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

func (s *FavoritesService) DeleteFavoriteCategory(userID int, categoryID int) error {
	category, err := s.favoritesRepo.GetFavoriteCategoryByID(categoryID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	if category.UserID != userID {
		return fmt.Errorf("unauthorized to delete this category")
	}

	favoritesCount, err := s.favoritesRepo.CountFavoritesByCategory(categoryID)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}

	if favoritesCount > 0 {
		return fmt.Errorf("cannot delete category with existing favorites")
	}

	return s.favoritesRepo.DeleteFavoriteCategory(categoryID)
}

func (s *FavoritesService) GetPublicFavorites(entityType *string, category *string, limit, offset int) ([]models.Favorite, error) {
	return s.favoritesRepo.GetPublicFavorites(entityType, category, limit, offset)
}

func (s *FavoritesService) SearchFavorites(userID int, query string, entityType *string, limit, offset int) ([]models.Favorite, error) {
	return s.favoritesRepo.SearchFavorites(userID, query, entityType, limit, offset)
}

func (s *FavoritesService) GetFavoriteStatistics(userID int) (*models.FavoriteStatistics, error) {
	stats := &models.FavoriteStatistics{
		UserID: userID,
	}

	totalCount, err := s.favoritesRepo.CountUserFavorites(userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total favorites count: %w", err)
	}
	stats.TotalFavorites = totalCount

	entityTypeCounts, err := s.favoritesRepo.GetFavoritesCountByEntityType(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity type counts: %w", err)
	}
	stats.FavoritesByEntityType = entityTypeCounts

	categoryCounts, err := s.favoritesRepo.GetFavoritesCountByCategory(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}
	stats.FavoritesByCategory = categoryCounts

	recentFavorites, err := s.favoritesRepo.GetRecentFavorites(userID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent favorites: %w", err)
	}
	stats.RecentFavorites = recentFavorites

	return stats, nil
}

func (s *FavoritesService) GetRecommendedFavorites(userID int, limit int) ([]models.RecommendedFavorite, error) {
	userFavorites, err := s.favoritesRepo.GetUserFavorites(userID, nil, nil, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user favorites: %w", err)
	}

	var entityTypes []string
	var categories []string

	for _, favorite := range userFavorites {
		entityTypes = append(entityTypes, favorite.EntityType)
		if favorite.Category != nil {
			categories = append(categories, *favorite.Category)
		}
	}

	entityTypes = s.removeDuplicateStrings(entityTypes)
	categories = s.removeDuplicateStrings(categories)

	var recommendations []models.RecommendedFavorite

	for _, entityType := range entityTypes {
		similarFavorites, err := s.favoritesRepo.GetSimilarFavorites(userID, entityType, limit/len(entityTypes))
		if err != nil {
			continue
		}

		for _, favorite := range similarFavorites {
			recommendation := models.RecommendedFavorite{
				Favorite:        favorite,
				RecommendReason: fmt.Sprintf("Based on your interest in %s", entityType),
				RecommendScore:  0.8,
				RecommendedAt:   time.Now(),
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	return recommendations, nil
}

func (s *FavoritesService) ShareFavorite(userID int, favoriteID int, shareWith []int, permissions models.SharePermissions) (*models.FavoriteShare, error) {
	favorite, err := s.favoritesRepo.GetFavoriteByID(favoriteID)
	if err != nil {
		return nil, fmt.Errorf("favorite not found: %w", err)
	}

	if favorite.UserID != userID {
		return nil, fmt.Errorf("unauthorized to share this favorite")
	}

	share := &models.FavoriteShare{
		FavoriteID:   favoriteID,
		SharedByUser: userID,
		SharedWith:   shareWith,
		Permissions:  permissions,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}

	id, err := s.favoritesRepo.CreateFavoriteShare(share)
	if err != nil {
		return nil, fmt.Errorf("failed to create favorite share: %w", err)
	}

	share.ID = id
	return share, nil
}

func (s *FavoritesService) GetSharedFavorites(userID int, limit, offset int) ([]models.Favorite, error) {
	return s.favoritesRepo.GetSharedFavorites(userID, limit, offset)
}

func (s *FavoritesService) RevokeFavoriteShare(userID int, shareID int) error {
	share, err := s.favoritesRepo.GetFavoriteShareByID(shareID)
	if err != nil {
		return fmt.Errorf("share not found: %w", err)
	}

	if share.SharedByUser != userID {
		return fmt.Errorf("unauthorized to revoke this share")
	}

	return s.favoritesRepo.RevokeFavoriteShare(shareID)
}

func (s *FavoritesService) BulkAddFavorites(userID int, favorites []models.BulkFavoriteRequest) ([]models.Favorite, error) {
	var results []models.Favorite
	var errors []error

	for _, req := range favorites {
		favorite := &models.Favorite{
			EntityType: req.EntityType,
			EntityID:   req.EntityID,
			Category:   req.Category,
			Notes:      req.Notes,
			Tags:       req.Tags,
			IsPublic:   req.IsPublic,
		}

		result, err := s.AddFavorite(userID, favorite)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		results = append(results, *result)
	}

	if len(errors) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("failed to add any favorites: %v", errors)
	}

	return results, nil
}

func (s *FavoritesService) BulkRemoveFavorites(userID int, favorites []models.BulkFavoriteRemoveRequest) error {
	var errors []error

	for _, req := range favorites {
		err := s.RemoveFavorite(userID, req.EntityType, req.EntityID)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove some favorites: %v", errors)
	}

	return nil
}

func (s *FavoritesService) ExportFavorites(userID int, format string) ([]byte, error) {
	favorites, err := s.favoritesRepo.GetUserFavorites(userID, nil, nil, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user favorites: %w", err)
	}

	switch format {
	case "json":
		return s.exportFavoritesToJSON(favorites)
	case "csv":
		return s.exportFavoritesToCSV(favorites)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (s *FavoritesService) ImportFavorites(userID int, data []byte, format string) ([]models.Favorite, error) {
	switch format {
	case "json":
		return s.importFavoritesFromJSON(userID, data)
	case "csv":
		return s.importFavoritesFromCSV(userID, data)
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}
}

func (s *FavoritesService) removeDuplicateStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (s *FavoritesService) exportFavoritesToJSON(favorites []models.Favorite) ([]byte, error) {
	// Implementation would use JSON marshal
	return nil, fmt.Errorf("JSON export not yet implemented")
}

func (s *FavoritesService) exportFavoritesToCSV(favorites []models.Favorite) ([]byte, error) {
	// Implementation would use CSV writer
	return nil, fmt.Errorf("CSV export not yet implemented")
}

func (s *FavoritesService) importFavoritesFromJSON(userID int, data []byte) ([]models.Favorite, error) {
	// Implementation would use JSON unmarshal
	return nil, fmt.Errorf("JSON import not yet implemented")
}

func (s *FavoritesService) importFavoritesFromCSV(userID int, data []byte) ([]models.Favorite, error) {
	// Implementation would use CSV reader
	return nil, fmt.Errorf("CSV import not yet implemented")
}
