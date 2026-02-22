package handlers

import (
	"net/http"
	"strconv"

	"catalogizer/internal/media/models"
	"catalogizer/repository"
	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

// CollectionHandler handles media collection CRUD endpoints.
type CollectionHandler struct {
	repo *repository.MediaCollectionRepository
}

// NewCollectionHandler creates a new collection handler.
func NewCollectionHandler(repo *repository.MediaCollectionRepository) *CollectionHandler {
	return &CollectionHandler{repo: repo}
}

// ListCollections handles GET /api/v1/collections.
func (h *CollectionHandler) ListCollections(c *gin.Context) {
	ctx := c.Request.Context()

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

	collections, total, err := h.repo.List(ctx, limit, offset)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to list collections", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  collectionsToJSON(collections),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetCollection handles GET /api/v1/collections/:id.
func (h *CollectionHandler) GetCollection(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid collection ID", err)
		return
	}

	coll, err := h.repo.GetByID(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Collection not found", err)
		return
	}

	c.JSON(http.StatusOK, coll)
}

// CreateCollection handles POST /api/v1/collections.
func (h *CollectionHandler) CreateCollection(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Name           string            `json:"name" binding:"required"`
		CollectionType string            `json:"collection_type,omitempty" binding:"-"`
		Description    *string           `json:"description,omitempty"`
		IsPublic       *bool             `json:"is_public,omitempty"`
		IsSmart        *bool             `json:"is_smart,omitempty"`
		ExternalIDs    map[string]string `json:"external_ids,omitempty"`
		CoverURL       *string           `json:"cover_url,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set default collection type if empty
	collectionType := req.CollectionType
	if collectionType == "" {
		collectionType = "custom"
	}
	// Map request to model
	coll := &models.MediaCollection{
		Name:           req.Name,
		CollectionType: collectionType,
		Description:    req.Description,
		ExternalIDs:    req.ExternalIDs,
		CoverURL:       req.CoverURL,
		TotalItems:     0,
	}
	// is_public and is_smart are not stored in current schema; ignore for now

	id, err := h.repo.Create(ctx, coll)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create collection", err)
		return
	}

	coll.ID = id
	c.JSON(http.StatusCreated, coll)
}

// UpdateCollection handles PUT /api/v1/collections/:id.
func (h *CollectionHandler) UpdateCollection(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid collection ID", err)
		return
	}

	var req struct {
		Name           *string           `json:"name,omitempty"`
		CollectionType *string           `json:"collection_type,omitempty"`
		Description    *string           `json:"description,omitempty"`
		IsPublic       *bool             `json:"is_public,omitempty"`
		IsSmart        *bool             `json:"is_smart,omitempty"`
		ExternalIDs    map[string]string `json:"external_ids,omitempty"`
		CoverURL       *string           `json:"cover_url,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Fetch existing collection
	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Collection not found", err)
		return
	}

	// Apply updates
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.CollectionType != nil {
		existing.CollectionType = *req.CollectionType
	}
	if req.Description != nil {
		existing.Description = req.Description
	}
	if req.ExternalIDs != nil {
		existing.ExternalIDs = req.ExternalIDs
	}
	if req.CoverURL != nil {
		existing.CoverURL = req.CoverURL
	}
	// is_public and is_smart ignored

	if err := h.repo.Update(ctx, existing); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update collection", err)
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteCollection handles DELETE /api/v1/collections/:id.
func (h *CollectionHandler) DeleteCollection(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid collection ID", err)
		return
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete collection", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collection deleted",
	})
}

// collectionsToJSON converts slice of MediaCollection to JSON-friendly slice.
func collectionsToJSON(collections []*models.MediaCollection) []gin.H {
	result := make([]gin.H, 0, len(collections))
	for _, coll := range collections {
		result = append(result, gin.H{
			"id":              coll.ID,
			"name":            coll.Name,
			"collection_type": coll.CollectionType,
			"description":     coll.Description,
			"total_items":     coll.TotalItems,
			"external_ids":    coll.ExternalIDs,
			"cover_url":       coll.CoverURL,
			"created_at":      coll.CreatedAt,
			"updated_at":      coll.UpdatedAt,
		})
	}
	return result
}
