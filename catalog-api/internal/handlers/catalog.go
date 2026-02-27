package handlers

import (
	"catalogizer/internal/models"
	"catalogizer/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CatalogHandler struct {
	catalogService services.CatalogServiceInterface
	smbService     services.SMBServiceInterface
	logger         *zap.Logger
}

func NewCatalogHandler(catalogService services.CatalogServiceInterface, smbService services.SMBServiceInterface, logger *zap.Logger) *CatalogHandler {
	return &CatalogHandler{
		catalogService: catalogService,
		smbService:     smbService,
		logger:         logger,
	}
}

// @Summary List root directories
// @Description Get list of available SMB root directories
// @Tags catalog
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} map[string]string
// @Router /api/v1/catalog [get]
func (h *CatalogHandler) ListRoot(c *gin.Context) {
	roots, err := h.catalogService.GetSMBRoots()
	if err != nil {
		h.logger.Error("Failed to get SMB roots", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get SMB roots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roots": roots})
}

// @Summary List files in path
// @Description Get list of files and directories in the specified path
// @Tags catalog
// @Param path path string true "Path to browse"
// @Param sort_by query string false "Sort by field (name, size, modified)" default(name)
// @Param sort_order query string false "Sort order (asc, desc)" default(asc)
// @Param limit query int false "Limit number of results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Produce json
// @Success 200 {array} models.FileInfo
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/catalog/{path} [get]
func (h *CatalogHandler) ListPath(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Clean the path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	sortBy := c.DefaultQuery("sort_by", "name")
	sortOrder := c.DefaultQuery("sort_order", "asc")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	files, err := h.catalogService.ListPath(path, sortBy, sortOrder, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list path", zap.String("path", path), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list directory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files":  files,
		"count":  len(files),
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary Get file information
// @Description Get detailed information about a specific file or directory
// @Tags catalog
// @Param path path string true "Path to file/directory"
// @Produce json
// @Success 200 {object} models.FileInfo
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/catalog-info/{path} [get]
func (h *CatalogHandler) GetFileInfo(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Clean the path
	path = strings.TrimPrefix(path, "/")

	// Try to get file info by path or ID
	fileInfo, err := h.catalogService.GetFileInfo(path)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		h.logger.Error("Failed to get file info", zap.String("path", path), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file information"})
		return
	}

	if fileInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}

// @Summary Search files
// @Description Search for files and directories based on various criteria
// @Tags search
// @Param query query string false "Search query (filename)"
// @Param path query string false "Path filter"
// @Param extension query string false "File extension filter"
// @Param mime_type query string false "MIME type filter"
// @Param min_size query int false "Minimum file size"
// @Param max_size query int false "Maximum file size"
// @Param smb_roots query string false "Comma-separated list of SMB roots"
// @Param is_directory query bool false "Filter by directory status"
// @Param sort_by query string false "Sort by field" default(name)
// @Param sort_order query string false "Sort order" default(asc)
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/search [get]
func (h *CatalogHandler) Search(c *gin.Context) {
	var req models.SearchRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search parameters"})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "name"
	}
	if req.SortOrder == "" {
		req.SortOrder = "asc"
	}

	// Parse SMB roots if provided as comma-separated string
	smbRootsStr := c.Query("smb_roots")
	if smbRootsStr != "" {
		req.SmbRoots = strings.Split(smbRootsStr, ",")
	}

	files, total, err := h.catalogService.SearchFiles(&req)
	if err != nil {
		h.logger.Error("Failed to search files", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files":  files,
		"total":  total,
		"count":  len(files),
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

// @Summary Search duplicate files
// @Description Find groups of duplicate files
// @Tags search
// @Param smb_root query string false "SMB root to search in"
// @Param min_count query int false "Minimum number of duplicates" default(2)
// @Param limit query int false "Limit number of groups" default(50)
// @Produce json
// @Success 200 {array} models.DuplicateGroup
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/search/duplicates [get]
func (h *CatalogHandler) SearchDuplicates(c *gin.Context) {
	smbRoot := c.DefaultQuery("smb_root", "")
	minCount, _ := strconv.Atoi(c.DefaultQuery("min_count", "2"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if smbRoot == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SMB root is required"})
		return
	}

	groups, err := h.catalogService.GetDuplicateGroups(smbRoot, minCount, limit)
	if err != nil {
		h.logger.Error("Failed to get duplicate groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find duplicates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"count":  len(groups),
	})
}

// @Summary Get directories sorted by size
// @Description Get directories sorted by their total size
// @Tags stats
// @Param smb_root query string true "SMB root to analyze"
// @Param limit query int false "Limit number of results" default(50)
// @Produce json
// @Success 200 {array} models.DirectoryStats
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/stats/directories/by-size [get]
func (h *CatalogHandler) GetDirectoriesBySize(c *gin.Context) {
	smbRoot := c.Query("smb_root")
	if smbRoot == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SMB root is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	stats, err := h.catalogService.GetDirectoriesBySize(smbRoot, limit)
	if err != nil {
		h.logger.Error("Failed to get directories by size", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get directory statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"directories": stats,
		"count":       len(stats),
	})
}

// @Summary Get duplicate count statistics
// @Description Get statistics about duplicate files
// @Tags stats
// @Param smb_root query string false "SMB root to analyze"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/stats/duplicates/count [get]
func (h *CatalogHandler) GetDuplicatesCount(c *gin.Context) {
	smbRoot := c.DefaultQuery("smb_root", "")

	groups, err := h.catalogService.GetDuplicateGroups(smbRoot, 2, 1000)
	if err != nil {
		h.logger.Error("Failed to get duplicate statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get duplicate statistics"})
		return
	}

	var totalDuplicates int64
	var totalWastedSpace int64
	var groupCount int

	for _, group := range groups {
		if group.Count > 1 {
			groupCount++
			totalDuplicates += int64(group.Count - 1) // Don't count the original
			totalWastedSpace += int64(group.Count-1) * group.Size
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"duplicate_groups":   groupCount,
		"total_duplicates":   totalDuplicates,
		"total_wasted_space": totalWastedSpace,
		"smb_root":           smbRoot,
	})
}
