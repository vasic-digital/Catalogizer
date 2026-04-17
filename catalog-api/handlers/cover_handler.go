package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"catalogizer/internal/services"
	"catalogizer/repository"

	"github.com/gin-gonic/gin"
)

// CoverHandler serves cover images for media entities, including
// type-specific placeholder SVGs for entities without cover art.
type CoverHandler struct {
	coverArtService *services.CoverArtService
	qualityRepo     *repository.ImageQualityRepository
}

// NewCoverHandler creates a new cover handler.
func NewCoverHandler(coverArtService *services.CoverArtService) *CoverHandler {
	return &CoverHandler{coverArtService: coverArtService}
}

// WithQualityRepository attaches the image quality repository so response
// headers can reflect the last assessed verdict for the served cover.
func (h *CoverHandler) WithQualityRepository(repo *repository.ImageQualityRepository) *CoverHandler {
	h.qualityRepo = repo
	return h
}

// setQualityHeader writes X-Cover-Quality and X-Cover-Source based on the
// most recent image_quality_assessments row for the given item. When no row
// exists, the header is set to "unknown".
func (h *CoverHandler) setQualityHeader(ctx context.Context, c *gin.Context, mediaItemID int64) {
	if h.qualityRepo == nil {
		c.Header("X-Cover-Quality", "unknown")
		return
	}
	rec, err := h.qualityRepo.Find(ctx, "media_item", mediaItemID, "primary")
	if errors.Is(err, repository.ErrNotFound) || rec == nil {
		c.Header("X-Cover-Quality", "unknown")
		return
	}
	if err != nil {
		c.Header("X-Cover-Quality", "unknown")
		return
	}
	c.Header("X-Cover-Quality", rec.Verdict)
	if rec.Source != "" {
		c.Header("X-Cover-Source", rec.Source)
	}
}

// ServePlaceholder serves a type-specific placeholder SVG image.
// GET /api/v1/cover/placeholder/:type
func (h *CoverHandler) ServePlaceholder(c *gin.Context) {
	mediaType := c.Param("type")
	if mediaType == "" {
		mediaType = "movie"
	}

	svg := services.GeneratePlaceholderSVG(mediaType)

	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", "public, max-age=86400") // Cache 24h
	c.Header("Content-Length", strconv.Itoa(len(svg)))
	c.Data(http.StatusOK, "image/svg+xml", svg)
}

// ServeCover serves the cached cover art for a specific media item.
// GET /api/v1/cover/:id
func (h *CoverHandler) ServeCover(c *gin.Context) {
	ctx := c.Request.Context()

	mediaItemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media item ID"})
		return
	}

	h.setQualityHeader(ctx, c, mediaItemID)

	coverArt, err := h.coverArtService.GetCoverArt(ctx, mediaItemID)
	if err != nil || coverArt == nil {
		// Fall back to generic placeholder
		c.Header("X-Cover-Quality", "placeholder_fallback")
		svg := services.GeneratePlaceholderSVG("movie")
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, "image/svg+xml", svg)
		return
	}

	if coverArt.LocalPath != nil && *coverArt.LocalPath != "" {
		contentType := "image/jpeg"
		if coverArt.Format == "png" {
			contentType = "image/png"
		} else if coverArt.Format == "webp" {
			contentType = "image/webp"
		}
		c.Header("Cache-Control", "public, max-age=86400")

		// If object store is configured and the path looks like an object key,
		// stream from S3/MinIO. Otherwise fall back to local filesystem.
		if h.coverArtService.HasObjectStore() && !strings.HasPrefix(*coverArt.LocalPath, "/") && !strings.HasPrefix(*coverArt.LocalPath, ".") && !strings.HasPrefix(*coverArt.LocalPath, "\\") {
			reader, err := h.coverArtService.GetObjectStore().GetObject(ctx, h.coverArtService.GetBucket(), *coverArt.LocalPath)
			if err == nil {
				defer reader.Close()
				c.Header("Content-Type", contentType)
				c.Status(http.StatusOK)
				_, _ = io.Copy(c.Writer, reader)
				return
			}
			// Fall through to local file on error
		}
		c.File(*coverArt.LocalPath)
		_ = contentType // File() auto-detects content type
		return
	}

	if coverArt.URL != nil && *coverArt.URL != "" {
		c.Redirect(http.StatusFound, *coverArt.URL)
		return
	}

	// No cover art data available, serve placeholder
	c.Header("X-Cover-Quality", "placeholder_fallback")
	svg := services.GeneratePlaceholderSVG("movie")
	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "image/svg+xml", svg)
}

// GetCoverURL returns the cover URL for a media item as JSON.
// GET /api/v1/cover/url/:id
func (h *CoverHandler) GetCoverURL(c *gin.Context) {
	ctx := c.Request.Context()

	mediaItemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media item ID"})
		return
	}

	mediaType := c.DefaultQuery("type", "movie")
	coverURL := h.coverArtService.GetCoverURL(ctx, mediaItemID, mediaType)

	c.JSON(http.StatusOK, gin.H{
		"media_item_id": mediaItemID,
		"cover_url":     coverURL,
	})
}
