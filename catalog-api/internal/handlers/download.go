package handlers

import (
	"archive/tar"
	"archive/zip"
	"catalogizer/internal/models"
	"catalogizer/internal/services"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DownloadHandler struct {
	catalogService *services.CatalogService
	smbService     *services.SMBService
	tempDir        string
	maxArchiveSize int64
	chunkSize      int
	logger         *zap.Logger
}

func NewDownloadHandler(catalogService *services.CatalogService, smbService *services.SMBService, tempDir string, maxArchiveSize int64, chunkSize int, logger *zap.Logger) *DownloadHandler {
	return &DownloadHandler{
		catalogService: catalogService,
		smbService:     smbService,
		tempDir:        tempDir,
		maxArchiveSize: maxArchiveSize,
		chunkSize:      chunkSize,
		logger:         logger,
	}
}

// @Summary Download a single file
// @Description Download a file from the catalog
// @Tags download
// @Param id path int true "File ID"
// @Produce application/octet-stream
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/download/file/{id} [get]
func (h *DownloadHandler) DownloadFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	fileInfo, err := h.catalogService.GetFileInfo(strconv.FormatInt(id, 10))
	if err != nil {
		h.logger.Error("Failed to get file info", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file information"})
		return
	}

	if fileInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if fileInfo.IsDirectory {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot download directory as single file"})
		return
	}

	// Create temporary file for download
	tempFile, err := os.CreateTemp(h.tempDir, "download_*")
	if err != nil {
		h.logger.Error("Failed to create temp file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare download"})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Download from SMB to temp file
	err = h.smbService.DownloadFile(fileInfo.SmbRoot, fileInfo.Path, tempFile.Name())
	if err != nil {
		h.logger.Error("Failed to download from SMB", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		return
	}

	// Stream file to client
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size, 10))

	tempFile.Seek(0, 0)
	_, err = io.CopyBuffer(c.Writer, tempFile, make([]byte, h.chunkSize))
	if err != nil {
		h.logger.Error("Failed to stream file", zap.Error(err))
		return
	}

	h.logger.Info("File downloaded successfully", zap.String("file", fileInfo.Name), zap.Int64("id", id))
}

// @Summary Download directory as archive
// @Description Download a directory and its contents as a compressed archive
// @Tags download
// @Param path path string true "Directory path"
// @Param format query string false "Archive format (zip, tar, tar.gz)" default(zip)
// @Produce application/octet-stream
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/download/directory/{path} [get]
func (h *DownloadHandler) DownloadDirectory(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	format := c.DefaultQuery("format", "zip")
	if format != "zip" && format != "tar" && format != "tar.gz" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Supported: zip, tar, tar.gz"})
		return
	}

	// Clean the path
	path = strings.TrimPrefix(path, "/")

	// Get directory listing recursively
	files, err := h.getDirectoryContentsRecursive(path)
	if err != nil {
		h.logger.Error("Failed to get directory contents", zap.String("path", path), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
		return
	}

	if len(files) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found or empty"})
		return
	}

	// Check total size
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	if totalSize > h.maxArchiveSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Directory too large for download",
			"total_size": totalSize,
			"max_size":   h.maxArchiveSize,
		})
		return
	}

	// Create archive
	filename := filepath.Base(path) + "." + format
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	switch format {
	case "zip":
		c.Header("Content-Type", "application/zip")
		h.createZipArchive(c.Writer, files)
	case "tar":
		c.Header("Content-Type", "application/x-tar")
		h.createTarArchive(c.Writer, files, false)
	case "tar.gz":
		c.Header("Content-Type", "application/gzip")
		h.createTarArchive(c.Writer, files, true)
	}

	h.logger.Info("Directory downloaded successfully", zap.String("path", path), zap.String("format", format))
}

// @Summary Create archive from multiple files
// @Description Create and download an archive containing specified files
// @Tags download
// @Accept json
// @Param request body models.DownloadRequest true "Download request"
// @Produce application/octet-stream
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/download/archive [post]
func (h *DownloadHandler) DownloadArchive(c *gin.Context) {
	var req models.DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(req.Paths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No paths specified"})
		return
	}

	if req.Format == "" {
		req.Format = "zip"
	}

	if req.Format != "zip" && req.Format != "tar" && req.Format != "tar.gz" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Supported: zip, tar, tar.gz"})
		return
	}

	// Get file information for all paths
	var files []models.FileInfo
	var totalSize int64

	for _, path := range req.Paths {
		// This is a simplified implementation - you'd need to implement path-to-ID conversion
		// or modify the catalog service to search by path
		fileList, err := h.getFilesByPath(path, req.SmbRoot)
		if err != nil {
			h.logger.Error("Failed to get files for path", zap.String("path", path), zap.Error(err))
			continue
		}

		for _, file := range fileList {
			totalSize += file.Size
			files = append(files, file)
		}
	}

	if totalSize > h.maxArchiveSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Total size too large for download",
			"total_size": totalSize,
			"max_size":   h.maxArchiveSize,
		})
		return
	}

	// Create archive
	filename := fmt.Sprintf("archive_%d.%s", time.Now().Unix(), req.Format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	switch req.Format {
	case "zip":
		c.Header("Content-Type", "application/zip")
		h.createZipArchive(c.Writer, files)
	case "tar":
		c.Header("Content-Type", "application/x-tar")
		h.createTarArchive(c.Writer, files, false)
	case "tar.gz":
		c.Header("Content-Type", "application/gzip")
		h.createTarArchive(c.Writer, files, true)
	}

	h.logger.Info("Archive downloaded successfully", zap.Int("file_count", len(files)), zap.String("format", req.Format))
}

func (h *DownloadHandler) createZipArchive(w io.Writer, files []models.FileInfo) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, file := range files {
		if file.IsDirectory {
			continue // Skip directories for now
		}

		// Download file to temp location
		tempFile, err := os.CreateTemp(h.tempDir, "zip_*")
		if err != nil {
			h.logger.Error("Failed to create temp file for zip", zap.Error(err))
			continue
		}

		err = h.smbService.DownloadFile(file.SmbRoot, file.Path, tempFile.Name())
		if err != nil {
			h.logger.Error("Failed to download file for zip", zap.String("path", file.Path), zap.Error(err))
			tempFile.Close()
			os.Remove(tempFile.Name())
			continue
		}

		// Add to zip
		zipFile, err := zipWriter.Create(file.Path)
		if err != nil {
			h.logger.Error("Failed to create zip entry", zap.String("path", file.Path), zap.Error(err))
			tempFile.Close()
			os.Remove(tempFile.Name())
			continue
		}

		tempFile.Seek(0, 0)
		_, err = io.Copy(zipFile, tempFile)
		tempFile.Close()
		os.Remove(tempFile.Name())

		if err != nil {
			h.logger.Error("Failed to write zip entry", zap.String("path", file.Path), zap.Error(err))
			continue
		}
	}

	return nil
}

func (h *DownloadHandler) createTarArchive(w io.Writer, files []models.FileInfo, compress bool) error {
	var tarWriter *tar.Writer

	if compress {
		gzWriter := gzip.NewWriter(w)
		defer gzWriter.Close()
		tarWriter = tar.NewWriter(gzWriter)
	} else {
		tarWriter = tar.NewWriter(w)
	}
	defer tarWriter.Close()

	for _, file := range files {
		if file.IsDirectory {
			continue // Skip directories for now
		}

		// Download file to temp location
		tempFile, err := os.CreateTemp(h.tempDir, "tar_*")
		if err != nil {
			h.logger.Error("Failed to create temp file for tar", zap.Error(err))
			continue
		}

		err = h.smbService.DownloadFile(file.SmbRoot, file.Path, tempFile.Name())
		if err != nil {
			h.logger.Error("Failed to download file for tar", zap.String("path", file.Path), zap.Error(err))
			tempFile.Close()
			os.Remove(tempFile.Name())
			continue
		}

		// Create tar header
		header := &tar.Header{
			Name:    file.Path,
			Size:    file.Size,
			Mode:    0644,
			ModTime: file.LastModified,
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			h.logger.Error("Failed to write tar header", zap.String("path", file.Path), zap.Error(err))
			tempFile.Close()
			os.Remove(tempFile.Name())
			continue
		}

		tempFile.Seek(0, 0)
		_, err = io.Copy(tarWriter, tempFile)
		tempFile.Close()
		os.Remove(tempFile.Name())

		if err != nil {
			h.logger.Error("Failed to write tar entry", zap.String("path", file.Path), zap.Error(err))
			continue
		}
	}

	return nil
}

// Helper functions
func (h *DownloadHandler) getDirectoryContentsRecursive(path string) ([]models.FileInfo, error) {
	// This is a placeholder implementation
	// You would need to implement recursive directory traversal
	// using your catalog service
	return []models.FileInfo{}, nil
}

func (h *DownloadHandler) getFilesByPath(path, smbRoot string) ([]models.FileInfo, error) {
	// This is a placeholder implementation
	// You would need to implement path-based file lookup
	return []models.FileInfo{}, nil
}
