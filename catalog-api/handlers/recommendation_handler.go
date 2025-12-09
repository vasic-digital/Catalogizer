// This is the new recommendation handler for Gin
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"catalogizer/internal/services"
	"catalogizer/models"
	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

type RecommendationHandler struct {
	recommendationService *services.RecommendationService
}

func NewRecommendationHandler(
	recommendationService *services.RecommendationService,
) *RecommendationHandler {
	return &RecommendationHandler{
		recommendationService: recommendationService,
	}
}

// GetSimilarItems retrieves similar items for a given media
// @Summary Get similar media items
// @Description Retrieves similar media items based on content similarity and external APIs
// @Tags recommendations
// @Accept json
// @Produce json
// @Param media_id path int true "Media ID"
// @Param max_local_items query int false "Maximum local items to return (default: 10)"
// @Param max_external_items query int false "Maximum external items to return (default: 5)"
// @Param include_external query bool false "Include external recommendations (default: false)"
// @Param similarity_threshold query number false "Minimum similarity threshold (default: 0.3)"
// @Success 200 {object} services.SimilarItemsResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/recommendations/similar/{media_id} [get]
func (h *RecommendationHandler) GetSimilarItems(c *gin.Context) {
	ctx := c.Request.Context()

	// Get media ID from URL
	mediaIDStr := c.Param("media_id")
	mediaID, err := strconv.ParseInt(mediaIDStr, 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid media ID", err)
		return
	}

	// Parse query parameters
	maxLocalItems, _ := strconv.Atoi(c.DefaultQuery("max_local_items", "10"))
	maxExternalItems, _ := strconv.Atoi(c.DefaultQuery("max_external_items", "5"))
	includeExternal := c.DefaultQuery("include_external", "false") == "true"
	similarityThreshold, _ := strconv.ParseFloat(c.DefaultQuery("similarity_threshold", "0.3"), 64)

	// Get media metadata from database
	mediaMetadata, err := h.getMediaMetadata(ctx, mediaID)
	if err != nil {
		if err.Error() == "media not found" {
			utils.SendErrorResponse(c, http.StatusNotFound, "Media not found", nil)
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get media metadata", err)
		}
		return
	}

	// Create recommendation request
	req := &services.SimilarItemsRequest{
		MediaID:             mediaIDStr,
		MediaMetadata:       mediaMetadata,
		MaxLocalItems:       maxLocalItems,
		MaxExternalItems:    maxExternalItems,
		IncludeExternal:     includeExternal,
		SimilarityThreshold: similarityThreshold,
	}

	// Get recommendations
	response, err := h.recommendationService.GetSimilarItems(ctx, req)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get recommendations", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTrendingItems retrieves trending media items
// @Summary Get trending media
// @Description Retrieves trending media items based on recent activity and ratings
// @Tags recommendations
// @Accept json
// @Produce json
// @Param media_type query string false "Filter by media type"
// @Param limit query int false "Maximum items to return (default: 20)"
// @Param time_range query string false "Time range: day, week, month, year (default: week)"
// @Success 200 {object} TrendingResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/recommendations/trending [get]
func (h *RecommendationHandler) GetTrendingItems(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	mediaType := c.Query("media_type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	timeRange := c.DefaultQuery("time_range", "week")

	// Get trending items
	trending, err := h.getTrendingItems(ctx, mediaType, limit, timeRange)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get trending items", err)
		return
	}

	response := TrendingResponse{
		Items:      trending,
		MediaType:  mediaType,
		TimeRange:  timeRange,
		GeneratedAt: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// PersonalizedRecommendations retrieves personalized recommendations for a user
// @Summary Get personalized recommendations
// @Description Retrieves personalized recommendations based on user's viewing history and preferences
// @Tags recommendations
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param limit query int false "Maximum items to return (default: 20)"
// @Success 200 {object} PersonalizedResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/recommendations/personalized/{user_id} [get]
func (h *RecommendationHandler) GetPersonalizedRecommendations(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from URL
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Get personalized recommendations
	recommendations, err := h.getPersonalizedRecommendations(ctx, userID, limit)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get personalized recommendations", err)
		return
	}

	response := PersonalizedResponse{
		UserID:         userID,
		Items:          recommendations,
		GeneratedAt:    time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// getMediaMetadata retrieves media metadata from database
func (h *RecommendationHandler) getMediaMetadata(ctx context.Context, mediaID int64) (*models.MediaMetadata, error) {
	// Query the database for media item
	query := `
		SELECT 
			id, title, media_type, year, description, rating, 
			duration, language, country, director, producer, 
			cast, resolution, file_size, created_at, updated_at
		FROM media_items 
		WHERE id = $1
	`

	var metadata models.MediaMetadata
	var castStr string
	err := h.recommendationService.GetDB().QueryRowContext(ctx, query, mediaID).Scan(
		&metadata.ID, &metadata.Title, &metadata.MediaType, &metadata.Year, 
		&metadata.Description, &metadata.Rating, &metadata.Duration, 
		&metadata.Language, &metadata.Country, &metadata.Director, 
		&metadata.Producer, &castStr, &metadata.Resolution, 
		&metadata.FileSize, &metadata.CreatedAt, &metadata.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse cast array if present
	if castStr != "" {
		// Simple parsing - in production, use proper JSON parsing
		metadata.Cast = []string{castStr}
	}

	return &metadata, nil
}

// getTrendingItems retrieves trending items based on recent activity
func (h *RecommendationHandler) getTrendingItems(ctx context.Context, mediaType string, limit int, timeRange string) ([]*models.MediaCatalogItem, error) {
	// Calculate time threshold
	var timeThreshold time.Time
	switch timeRange {
	case "day":
		timeThreshold = time.Now().AddDate(0, 0, -1)
	case "week":
		timeThreshold = time.Now().AddDate(0, 0, -7)
	case "month":
		timeThreshold = time.Now().AddDate(0, -1, 0)
	case "year":
		timeThreshold = time.Now().AddDate(-1, 0, 0)
	default:
		timeThreshold = time.Now().AddDate(0, 0, -7) // Default to week
	}

	// Build query
	query := `
		SELECT 
			id, title, media_type, year, description, cover_image, 
			rating, quality, file_size, duration, directory_path, 
			created_at, updated_at, is_favorite, watch_progress, 
			last_watched, is_downloaded
		FROM media_items
		WHERE updated_at >= $1
	`

	args := []interface{}{timeThreshold}
	argIndex := 2

	// Add media type filter if specified
	if mediaType != "" {
		query += " AND media_type = $" + strconv.Itoa(argIndex)
		args = append(args, mediaType)
		argIndex++
	}

	// Order by watch progress, rating, and recent activity
	query += `
		ORDER BY 
			CASE WHEN watch_progress > 0 THEN watch_progress ELSE 0 END DESC,
			COALESCE(rating, 0) DESC,
			last_watched DESC NULLS LAST,
			updated_at DESC
		LIMIT $` + strconv.Itoa(argIndex)
	args = append(args, limit)

	rows, err := h.recommendationService.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.MediaCatalogItem
	for rows.Next() {
		var item models.MediaCatalogItem
		err := rows.Scan(
			&item.ID, &item.Title, &item.MediaType, &item.Year, 
			&item.Description, &item.CoverImage, &item.Rating, 
			&item.Quality, &item.FileSize, &item.Duration, 
			&item.DirectoryPath, &item.CreatedAt, &item.UpdatedAt, 
			&item.IsFavorite, &item.WatchProgress, &item.LastWatched, 
			&item.IsDownloaded,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, nil
}

// getPersonalizedRecommendations retrieves personalized recommendations for a user
func (h *RecommendationHandler) getPersonalizedRecommendations(ctx context.Context, userID int64, limit int) ([]*models.MediaCatalogItem, error) {
	// Get user's viewing history and preferences
	query := `
		SELECT DISTINCT
			mi.id, mi.title, mi.media_type, mi.year, mi.description, 
			mi.cover_image, mi.rating, mi.quality, mi.file_size, 
			mi.duration, mi.directory_path, mi.created_at, mi.updated_at, 
			mi.is_favorite, mi.watch_progress, mi.last_watched, mi.is_downloaded
		FROM media_items mi
		LEFT JOIN user_watch_history uwh ON mi.id = uwh.media_id
		WHERE uwh.user_id = $1 OR mi.is_favorite = true
		ORDER BY 
			uwh.watched_at DESC NULLS LAST,
			mi.last_watched DESC NULLS LAST,
			mi.rating DESC NULLS LAST
		LIMIT $2
	`

	rows, err := h.recommendationService.GetDB().QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.MediaCatalogItem
	for rows.Next() {
		var item models.MediaCatalogItem
		err := rows.Scan(
			&item.ID, &item.Title, &item.MediaType, &item.Year, 
			&item.Description, &item.CoverImage, &item.Rating, 
			&item.Quality, &item.FileSize, &item.Duration, 
			&item.DirectoryPath, &item.CreatedAt, &item.UpdatedAt, 
			&item.IsFavorite, &item.WatchProgress, &item.LastWatched, 
			&item.IsDownloaded,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, nil
}

// TrendingResponse represents the response for trending items
type TrendingResponse struct {
	Items      []*models.MediaCatalogItem `json:"items"`
	MediaType  string                     `json:"media_type"`
	TimeRange  string                     `json:"time_range"`
	GeneratedAt time.Time                  `json:"generated_at"`
}

// PersonalizedResponse represents the response for personalized recommendations
type PersonalizedResponse struct {
	UserID      int64                      `json:"user_id"`
	Items       []*models.MediaCatalogItem  `json:"items"`
	GeneratedAt time.Time                  `json:"generated_at"`
}