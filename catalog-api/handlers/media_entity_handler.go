package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"catalogizer/internal/media/models"
	"catalogizer/repository"
	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

// MediaEntityHandler handles entity-level media browsing endpoints.
type MediaEntityHandler struct {
	itemRepo    *repository.MediaItemRepository
	fileRepo    *repository.MediaFileRepository
	extMetaRepo *repository.ExternalMetadataRepository
	userMetaRepo *repository.UserMetadataRepository
}

// NewMediaEntityHandler creates a new media entity handler.
func NewMediaEntityHandler(
	itemRepo *repository.MediaItemRepository,
	fileRepo *repository.MediaFileRepository,
	extMetaRepo *repository.ExternalMetadataRepository,
	userMetaRepo *repository.UserMetadataRepository,
) *MediaEntityHandler {
	return &MediaEntityHandler{
		itemRepo:     itemRepo,
		fileRepo:     fileRepo,
		extMetaRepo:  extMetaRepo,
		userMetaRepo: userMetaRepo,
	}
}

// ListEntities handles GET /api/v1/entities — list entities with filters and pagination.
func (h *MediaEntityHandler) ListEntities(c *gin.Context) {
	ctx := c.Request.Context()

	query := c.Query("query")
	mediaType := c.Query("type")
	limitStr := c.DefaultQuery("limit", "24")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 || limit > 200 {
		limit = 24
	}
	if offset < 0 {
		offset = 0
	}

	var mediaTypeIDs []int64
	if mediaType != "" {
		_, typeID, err := h.itemRepo.GetMediaTypeByName(ctx, mediaType)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid media type", err)
			return
		}
		mediaTypeIDs = append(mediaTypeIDs, typeID)
	}

	if query == "" {
		query = "%"
	}

	items, total, err := h.itemRepo.Search(ctx, query, mediaTypeIDs, limit, offset)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to list entities", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  itemsToJSON(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEntity handles GET /api/v1/entities/:id — get entity with details.
func (h *MediaEntityHandler) GetEntity(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	item, err := h.itemRepo.GetByID(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Entity not found", err)
		return
	}

	// Get file count
	fileCount, _ := h.fileRepo.CountByItem(ctx, id)

	// Get children count
	children, _ := h.itemRepo.GetChildren(ctx, id)
	childrenCount := len(children)

	// Get external metadata
	extMeta, _ := h.extMetaRepo.GetByItem(ctx, id)

	// Get media type name
	types, _ := h.itemRepo.GetMediaTypes(ctx)
	typeName := ""
	for _, mt := range types {
		if mt.ID == item.MediaTypeID {
			typeName = mt.Name
			break
		}
	}

	result := entityDetailJSON(item, typeName, fileCount, int64(childrenCount), extMeta)
	c.JSON(http.StatusOK, result)
}

// GetEntityChildren handles GET /api/v1/entities/:id/children.
func (h *MediaEntityHandler) GetEntityChildren(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.itemRepo.GetByParent(ctx, id, limit, offset)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get children", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  itemsToJSON(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEntityFiles handles GET /api/v1/entities/:id/files.
func (h *MediaEntityHandler) GetEntityFiles(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	files, err := h.fileRepo.GetFilesByItem(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get files", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"total": len(files),
	})
}

// GetEntityMetadata handles GET /api/v1/entities/:id/metadata.
func (h *MediaEntityHandler) GetEntityMetadata(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	metadata, err := h.extMetaRepo.GetByItem(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get metadata", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metadata": metadata,
	})
}

// GetEntityDuplicates handles GET /api/v1/entities/:id/duplicates.
func (h *MediaEntityHandler) GetEntityDuplicates(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	item, err := h.itemRepo.GetByID(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Entity not found", err)
		return
	}

	dups, err := h.itemRepo.GetDuplicates(ctx, item.Title, item.MediaTypeID, item.Year)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to find duplicates", err)
		return
	}

	// Exclude self
	var filtered []*models.MediaItem
	for _, d := range dups {
		if d.ID != id {
			filtered = append(filtered, d)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"duplicates": itemsToJSON(filtered),
		"total":      len(filtered),
	})
}

// GetEntityTypes handles GET /api/v1/entities/types — list media types with counts.
func (h *MediaEntityHandler) GetEntityTypes(c *gin.Context) {
	ctx := c.Request.Context()

	types, err := h.itemRepo.GetMediaTypes(ctx)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get media types", err)
		return
	}

	counts, _ := h.itemRepo.CountByType(ctx)

	var result []gin.H
	for _, mt := range types {
		count := int64(0)
		if c, ok := counts[mt.Name]; ok {
			count = c
		}
		result = append(result, gin.H{
			"id":          mt.ID,
			"name":        mt.Name,
			"description": mt.Description,
			"count":       count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"types": result,
	})
}

// BrowseByType handles GET /api/v1/entities/browse/:type.
func (h *MediaEntityHandler) BrowseByType(c *gin.Context) {
	ctx := c.Request.Context()

	typeName := c.Param("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "24"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 200 {
		limit = 24
	}
	if offset < 0 {
		offset = 0
	}

	_, typeID, err := h.itemRepo.GetMediaTypeByName(ctx, typeName)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid media type", err)
		return
	}

	// Only return top-level items (no children)
	items, total, err := h.itemRepo.GetByParent(ctx, typeID, limit, offset)
	if err != nil {
		// Fallback: GetByType returns all items of that type
		items, total, err = h.itemRepo.GetByType(ctx, typeID, limit, offset)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to browse type", err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  itemsToJSON(items),
		"total":  total,
		"type":   typeName,
		"limit":  limit,
		"offset": offset,
	})
}

// RefreshEntityMetadata handles POST /api/v1/entities/:id/metadata/refresh.
func (h *MediaEntityHandler) RefreshEntityMetadata(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "Metadata refresh queued",
		"entity_id": id,
	})
}

// UpdateUserMetadata handles PUT /api/v1/entities/:id/user-metadata.
func (h *MediaEntityHandler) UpdateUserMetadata(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	var req struct {
		UserRating    *float64 `json:"user_rating"`
		WatchedStatus *string  `json:"watched_status"`
		Favorite      *bool    `json:"favorite"`
		PersonalNotes *string  `json:"personal_notes"`
		Tags          []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Extract user ID from JWT context (default to 1 for now)
	userID := int64(1)
	if uid, exists := c.Get("user_id"); exists {
		if uidInt, ok := uid.(int64); ok {
			userID = uidInt
		}
	}

	um := &models.UserMetadata{
		MediaItemID:   id,
		UserID:        userID,
		UserRating:    req.UserRating,
		WatchedStatus: req.WatchedStatus,
		PersonalNotes: req.PersonalNotes,
		Tags:          req.Tags,
	}
	if req.Favorite != nil {
		um.Favorite = *req.Favorite
	}

	if err := h.userMetaRepo.Upsert(ctx, um); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update user metadata", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User metadata updated",
	})
}

// GetEntityStats handles GET /api/v1/entities/stats.
func (h *MediaEntityHandler) GetEntityStats(c *gin.Context) {
	ctx := c.Request.Context()

	totalCount, _ := h.itemRepo.Count(ctx)
	countByType, _ := h.itemRepo.CountByType(ctx)

	c.JSON(http.StatusOK, gin.H{
		"total_entities": totalCount,
		"by_type":        countByType,
	})
}

// StreamEntity handles GET /api/v1/entities/:id/stream — returns streaming info for primary file.
func (h *MediaEntityHandler) StreamEntity(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	files, err := h.fileRepo.GetFilesByItem(ctx, id)
	if err != nil || len(files) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "No files available for streaming", err)
		return
	}

	// Find primary file or use first one
	primary := files[0]
	for _, f := range files {
		if f.IsPrimary {
			primary = f
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":  id,
		"file_id":    primary.FileID,
		"stream_url": fmt.Sprintf("/api/v1/download/file/%d", primary.FileID),
	})
}

// DownloadEntity handles GET /api/v1/entities/:id/download — returns download info.
func (h *MediaEntityHandler) DownloadEntity(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	fileIDStr := c.Query("file_id")

	files, err := h.fileRepo.GetFilesByItem(ctx, id)
	if err != nil || len(files) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "No files available for download", err)
		return
	}

	target := files[0]
	if fileIDStr != "" {
		fileID, _ := strconv.ParseInt(fileIDStr, 10, 64)
		for _, f := range files {
			if f.FileID == fileID {
				target = f
				break
			}
		}
	} else {
		for _, f := range files {
			if f.IsPrimary {
				target = f
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":    id,
		"file_id":      target.FileID,
		"download_url": fmt.Sprintf("/api/v1/download/file/%d", target.FileID),
		"total_files":  len(files),
	})
}

// GetInstallInfo handles GET /api/v1/entities/:id/install-info — software installation details.
func (h *MediaEntityHandler) GetInstallInfo(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	item, err := h.itemRepo.GetByID(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Entity not found", err)
		return
	}

	// Verify this is a software entity
	types, _ := h.itemRepo.GetMediaTypes(ctx)
	typeName := ""
	for _, mt := range types {
		if mt.ID == item.MediaTypeID {
			typeName = mt.Name
			break
		}
	}
	if typeName != "software" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Install info is only available for software entities", nil)
		return
	}

	files, err := h.fileRepo.GetFilesByItem(ctx, id)
	if err != nil || len(files) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "No files available", err)
		return
	}

	primary := files[0]
	for _, f := range files {
		if f.IsPrimary {
			primary = f
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":    id,
		"title":        item.Title,
		"file_id":      primary.FileID,
		"download_url": fmt.Sprintf("/api/v1/entities/%d/download", id),
		"total_files":  len(files),
	})
}

// ListDuplicateGroups handles GET /api/v1/entities/duplicates — global duplicate listing.
func (h *MediaEntityHandler) ListDuplicateGroups(c *gin.Context) {
	ctx := c.Request.Context()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	groups, total, err := h.itemRepo.ListDuplicateGroups(ctx, limit, offset)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to list duplicates", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// --- JSON helpers ---

func itemsToJSON(items []*models.MediaItem) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, itemToJSON(item))
	}
	return result
}

func itemToJSON(item *models.MediaItem) gin.H {
	h := gin.H{
		"id":             item.ID,
		"media_type_id":  item.MediaTypeID,
		"title":          item.Title,
		"status":         item.Status,
		"first_detected": item.FirstDetected,
		"last_updated":   item.LastUpdated,
	}
	if item.OriginalTitle != nil {
		h["original_title"] = *item.OriginalTitle
	}
	if item.Year != nil {
		h["year"] = *item.Year
	}
	if item.Description != nil {
		h["description"] = *item.Description
	}
	if len(item.Genre) > 0 {
		h["genre"] = item.Genre
	}
	if item.Director != nil {
		h["director"] = *item.Director
	}
	if item.Rating != nil {
		h["rating"] = *item.Rating
	}
	if item.Runtime != nil {
		h["runtime"] = *item.Runtime
	}
	if item.Language != nil {
		h["language"] = *item.Language
	}
	if item.ParentID != nil {
		h["parent_id"] = *item.ParentID
	}
	if item.SeasonNumber != nil {
		h["season_number"] = *item.SeasonNumber
	}
	if item.EpisodeNumber != nil {
		h["episode_number"] = *item.EpisodeNumber
	}
	if item.TrackNumber != nil {
		h["track_number"] = *item.TrackNumber
	}
	return h
}

func entityDetailJSON(item *models.MediaItem, typeName string, fileCount, childrenCount int64, extMeta []*models.ExternalMetadata) gin.H {
	h := itemToJSON(item)
	h["media_type"] = typeName
	h["file_count"] = fileCount
	h["children_count"] = childrenCount
	if extMeta != nil {
		h["external_metadata"] = extMeta
	} else {
		h["external_metadata"] = []interface{}{}
	}
	return h
}
