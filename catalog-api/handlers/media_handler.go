package handlers

import (
	"net/http"
	"strconv"

	"catalogizer/database"
	"catalogizer/models"
	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

// AndroidTVMediaHandler handles media-specific operations for Android TV
type AndroidTVMediaHandler struct {
	db *database.DB
}

// NewAndroidTVMediaHandler creates a new Android TV media handler
func NewAndroidTVMediaHandler(db *database.DB) *AndroidTVMediaHandler {
	return &AndroidTVMediaHandler{
		db: db,
	}
}

// GetMediaByID godoc
// @Summary Get media by ID
// @Description Retrieve media item details by ID
// @Tags media
// @Accept json
// @Produce json
// @Param id path int true "Media ID"
// @Success 200 {object} models.MediaCatalogItem
// @Failure 400 {object} utils.SendErrorResponse
// @Failure 404 {object} utils.SendErrorResponse
// @Failure 500 {object} utils.SendErrorResponse
// @Router /api/v1/media/{id} [get]
func (h *AndroidTVMediaHandler) GetMediaByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse media ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid media ID", err)
		return
	}

	// Query media from database
	query := `
		SELECT 
			id, title, media_type, year, description, cover_image, rating, quality,
			file_size, duration, directory_path, smb_path, created_at, updated_at,
			external_metadata, versions, is_favorite, watch_progress, last_watched, is_downloaded
		FROM media_items 
		WHERE id = $1
	`
	
	var mediaItem models.MediaCatalogItem
	err = h.db.QueryRowContext(ctx, query, id).Scan(
		&mediaItem.ID, &mediaItem.Title, &mediaItem.MediaType, &mediaItem.Year, 
		&mediaItem.Description, &mediaItem.CoverImage, &mediaItem.Rating, 
		&mediaItem.Quality, &mediaItem.FileSize, &mediaItem.Duration, 
		&mediaItem.DirectoryPath, &mediaItem.SMBPath, &mediaItem.CreatedAt, 
		&mediaItem.UpdatedAt, &mediaItem.ExternalMetadata, &mediaItem.Versions,
		&mediaItem.IsFavorite, &mediaItem.WatchProgress, &mediaItem.LastWatched,
		&mediaItem.IsDownloaded,
	)
	
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			utils.SendErrorResponse(c, http.StatusNotFound, "Media not found", err)
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve media", err)
		}
		return
	}

	c.JSON(http.StatusOK, mediaItem)
}

// UpdateWatchProgress godoc
// @Summary Update watch progress
// @Description Update the watch progress for a media item
// @Tags media
// @Accept json
// @Produce json
// @Param id path int true "Media ID"
// @Param body body map[string]float64 true "Progress value (0.0 to 1.0)"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.SendErrorResponse
// @Failure 404 {object} utils.SendErrorResponse
// @Failure 500 {object} utils.SendErrorResponse
// @Router /api/v1/media/{id}/progress [put]
func (h *AndroidTVMediaHandler) UpdateWatchProgress(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse media ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid media ID", err)
		return
	}

	// Parse request body
	var requestBody map[string]float64
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	progress, exists := requestBody["progress"]
	if !exists {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Progress field is required", nil)
		return
	}

	// Validate progress value
	if progress < 0.0 || progress > 1.0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Progress must be between 0.0 and 1.0", nil)
		return
	}

	// Update database
	query := `
		UPDATE media_items 
		SET watch_progress = $1, last_watched = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	
	result, err := h.db.ExecContext(ctx, query, progress, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update watch progress", err)
		return
	}

	// Check if media item exists
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to check update result", err)
		return
	}

	if rowsAffected == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "Media not found", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Watch progress updated successfully",
	})
}

// UpdateFavoriteStatus godoc
// @Summary Update favorite status
// @Description Update the favorite status for a media item
// @Tags media
// @Accept json
// @Produce json
// @Param id path int true "Media ID"
// @Param body body map[string]bool true "Favorite status"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.SendErrorResponse
// @Failure 404 {object} utils.SendErrorResponse
// @Failure 500 {object} utils.SendErrorResponse
// @Router /api/v1/media/{id}/favorite [put]
func (h *AndroidTVMediaHandler) UpdateFavoriteStatus(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse media ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid media ID", err)
		return
	}

	// Parse request body
	var requestBody map[string]bool
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	isFavorite, exists := requestBody["favorite"]
	if !exists {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Favorite field is required", nil)
		return
	}

	// Update database
	query := `
		UPDATE media_items 
		SET is_favorite = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	
	result, err := h.db.ExecContext(ctx, query, isFavorite, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update favorite status", err)
		return
	}

	// Check if media item exists
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to check update result", err)
		return
	}

	if rowsAffected == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "Media not found", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Favorite status updated successfully",
	})
}