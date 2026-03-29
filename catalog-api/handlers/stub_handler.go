package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StubHandler provides placeholder endpoints for routes the frontend expects
// but are not yet fully implemented. These return valid empty JSON responses
// to satisfy the Zero Warning / Zero Error Policy.
type StubHandler struct{}

// NewStubHandler creates a new stub handler
func NewStubHandler() *StubHandler {
	return &StubHandler{}
}

// GetRecentMedia returns an empty list of recent media items.
// GET /api/v1/media/recent
func (h *StubHandler) GetRecentMedia(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []interface{}{},
		"total": 0,
	})
}

// GetPopularMedia returns an empty list of popular media items.
// GET /api/v1/media/popular
func (h *StubHandler) GetPopularMedia(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []interface{}{},
		"total": 0,
	})
}

// GetMediaByPath returns an empty list of media items filtered by path.
// GET /api/v1/media/by-path
func (h *StubHandler) GetMediaByPath(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []interface{}{},
		"total": 0,
	})
}

// RefreshMediaMetadata is a placeholder for triggering metadata refresh on a media item.
// POST /api/v1/media/:id/refresh
func (h *StubHandler) RefreshMediaMetadata(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "metadata refresh not yet implemented",
	})
}

// GetMediaQuality is a placeholder for retrieving quality analysis of a media item.
// GET /api/v1/media/:id/quality
func (h *StubHandler) GetMediaQuality(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"quality": gin.H{},
		"message": "quality analysis not yet implemented",
	})
}

// AnalyzeMedia is a placeholder for triggering media analysis.
// POST /api/v1/media/analyze
func (h *StubHandler) AnalyzeMedia(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "analysis not yet implemented",
	})
}

// ChangePassword is a placeholder for the password change endpoint.
// POST /api/v1/auth/change-password
func (h *StubHandler) ChangePassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "password change not yet implemented",
	})
}

// GetInitStatus returns the initialization status. This endpoint is public
// (no auth required) and always reports the system as initialized.
// GET /api/v1/auth/init-status
func (h *StubHandler) GetInitStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"initialized": true,
	})
}
