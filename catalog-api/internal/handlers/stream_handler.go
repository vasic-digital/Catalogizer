package handlers

import (
	"catalogizer/database"
	"catalogizer/filesystem"
	"catalogizer/internal/models"
	"catalogizer/internal/services"
	root_models "catalogizer/models"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StreamHandler handles streaming media files from any storage backend (SMB, FTP, NFS, WebDAV, local)
// to HTTP clients with Range request support for video seeking.
type StreamHandler struct {
	catalogService services.CatalogServiceInterface
	db             *database.DB
	clientFactory  filesystem.ClientFactory
	logger         *zap.Logger
}

// NewStreamHandler creates a new stream handler.
func NewStreamHandler(catalogService services.CatalogServiceInterface, db *database.DB, clientFactory filesystem.ClientFactory, logger *zap.Logger) *StreamHandler {
	return &StreamHandler{
		catalogService: catalogService,
		db:             db,
		clientFactory:  clientFactory,
		logger:         logger,
	}
}

// StreamFile proxies file data from any storage backend to the HTTP client.
// It supports Range requests for video seeking via http.ServeContent when the
// underlying protocol supports seeking (SMB, local), and falls back to simple
// streaming for protocols that don't (FTP, WebDAV).
//
// GET /api/v1/stream/:id
func (h *StreamHandler) StreamFile(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse file ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// Look up file metadata via the catalog service
	fileInfo, err := h.catalogService.GetFileInfo(strconv.FormatInt(id, 10))
	if err != nil {
		h.logger.Error("Failed to get file info for streaming", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file information"})
		return
	}
	if fileInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	if fileInfo.IsDirectory {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot stream a directory"})
		return
	}

	// Look up the storage root from the database
	storageRoot, err := h.getStorageRootByName(ctx, fileInfo.SmbRoot)
	if err != nil {
		h.logger.Error("Storage root not found for streaming",
			zap.String("root_name", fileInfo.SmbRoot),
			zap.Int64("file_id", id),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Storage root not found"})
		return
	}

	// Build settings map from storage root
	settings := storageRootToSettings(storageRoot)

	// Create a protocol-appropriate filesystem client
	fsClient, err := h.clientFactory.CreateClient(&filesystem.StorageConfig{
		ID:       storageRoot.Name,
		Name:     storageRoot.Name,
		Protocol: storageRoot.Protocol,
		Settings: settings,
	})
	if err != nil {
		h.logger.Error("Failed to create filesystem client for streaming",
			zap.String("protocol", storageRoot.Protocol),
			zap.Int64("file_id", id),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to storage"})
		return
	}

	// Connect to the storage backend
	if err := fsClient.Connect(ctx); err != nil {
		h.logger.Error("Failed to connect to storage backend",
			zap.String("protocol", storageRoot.Protocol),
			zap.String("root_name", storageRoot.Name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to storage backend"})
		return
	}
	defer fsClient.Disconnect(ctx)

	// Open the file for reading
	reader, err := fsClient.ReadFile(ctx, fileInfo.Path)
	if err != nil {
		h.logger.Error("Failed to open file for streaming",
			zap.String("path", fileInfo.Path),
			zap.String("protocol", storageRoot.Protocol),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "File not accessible on storage"})
		return
	}
	defer reader.Close()

	// Determine content type
	contentType := detectContentType(fileInfo)

	// Set streaming headers
	disposition := "inline"
	if c.Query("download") == "true" {
		disposition = "attachment"
	}
	c.Header("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disposition, sanitizeContentDisposition(fileInfo.Name)))
	c.Header("Content-Type", contentType)

	// Try to use http.ServeContent for Range request support.
	// This works when the reader implements io.ReadSeeker (SMB and local protocols).
	if rs, ok := reader.(io.ReadSeeker); ok {
		// ServeContent handles Range requests, If-Modified-Since, Content-Length, etc.
		http.ServeContent(c.Writer, c.Request, fileInfo.Name, fileInfo.LastModified, rs)
		h.logger.Info("File streamed with seek support",
			zap.String("file", fileInfo.Name),
			zap.String("protocol", storageRoot.Protocol),
			zap.Int64("id", id))
		return
	}

	// Fallback: simple streaming without Range support (FTP, WebDAV).
	// Set Content-Length if we know the file size.
	if fileInfo.Size > 0 {
		c.Header("Content-Length", strconv.FormatInt(fileInfo.Size, 10))
	}
	c.Status(http.StatusOK)
	written, err := io.Copy(c.Writer, reader)
	if err != nil {
		h.logger.Error("Error during stream copy",
			zap.String("file", fileInfo.Name),
			zap.Int64("bytes_written", written),
			zap.Error(err))
		return
	}

	h.logger.Info("File streamed (no seek)",
		zap.String("file", fileInfo.Name),
		zap.String("protocol", storageRoot.Protocol),
		zap.Int64("bytes", written),
		zap.Int64("id", id))
}

// getStorageRootByName queries the database for a storage root by its name.
func (h *StreamHandler) getStorageRootByName(ctx context.Context, name string) (*root_models.StorageRoot, error) {
	query := `
		SELECT id, name, protocol, host, port, path, username, password, domain,
			   mount_point, options, url, enabled, max_depth,
			   enable_duplicate_detection, enable_metadata_extraction, include_patterns,
			   exclude_patterns, created_at, updated_at, last_scan_at
		FROM storage_roots
		WHERE name = ?`

	var root root_models.StorageRoot
	err := h.db.QueryRowContext(ctx, query, name).Scan(
		&root.ID, &root.Name, &root.Protocol, &root.Host, &root.Port, &root.Path,
		&root.Username, &root.Password, &root.Domain, &root.MountPoint, &root.Options,
		&root.URL, &root.Enabled, &root.MaxDepth, &root.EnableDuplicateDetection,
		&root.EnableMetadataExtraction, &root.IncludePatterns, &root.ExcludePatterns,
		&root.CreatedAt, &root.UpdatedAt, &root.LastScanAt,
	)
	if err != nil {
		return nil, fmt.Errorf("storage root %q not found: %w", name, err)
	}
	return &root, nil
}

// storageRootToSettings converts a StorageRoot model to the settings map expected
// by the filesystem ClientFactory. This mirrors the mapping in UniversalScanner.
func storageRootToSettings(root *root_models.StorageRoot) map[string]interface{} {
	settings := make(map[string]interface{})

	switch root.Protocol {
	case "local":
		if root.Path != nil {
			settings["base_path"] = *root.Path
		}

	case "smb":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Port != nil {
			settings["port"] = *root.Port
		}
		if root.Path != nil {
			settings["share"] = *root.Path
		}
		if root.Username != nil {
			settings["username"] = *root.Username
		}
		if root.Password != nil {
			settings["password"] = *root.Password
		}
		if root.Domain != nil {
			settings["domain"] = *root.Domain
		}

	case "ftp":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Port != nil {
			settings["port"] = *root.Port
		}
		if root.Username != nil {
			settings["username"] = *root.Username
		}
		if root.Password != nil {
			settings["password"] = *root.Password
		}

	case "nfs":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Path != nil {
			settings["export_path"] = *root.Path
		}
		if root.MountPoint != nil {
			settings["mount_point"] = *root.MountPoint
		}
		if root.Options != nil {
			settings["options"] = *root.Options
		}

	case "webdav":
		if root.URL != nil {
			settings["url"] = *root.URL
		}
		if root.Username != nil {
			settings["username"] = *root.Username
		}
		if root.Password != nil {
			settings["password"] = *root.Password
		}
	}

	return settings
}

// detectContentType determines the MIME type for streaming.
// Uses the file's stored MIME type if available, falls back to extension-based detection.
func detectContentType(fileInfo *models.FileInfo) string {
	if fileInfo.MimeType != nil && *fileInfo.MimeType != "" {
		return *fileInfo.MimeType
	}

	ext := strings.ToLower(filepath.Ext(fileInfo.Name))

	// Common media types that are important for streaming
	switch ext {
	// Video
	case ".mp4":
		return "video/mp4"
	case ".mkv":
		return "video/x-matroska"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".webm":
		return "video/webm"
	case ".flv":
		return "video/x-flv"
	case ".m4v":
		return "video/x-m4v"
	case ".ts":
		return "video/mp2t"
	case ".3gp":
		return "video/3gpp"

	// Audio
	case ".mp3":
		return "audio/mpeg"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	case ".wav":
		return "audio/wav"
	case ".wma":
		return "audio/x-ms-wma"
	case ".m4a":
		return "audio/mp4"
	case ".opus":
		return "audio/opus"
	case ".aiff", ".aif":
		return "audio/aiff"

	// Images
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".bmp":
		return "image/bmp"

	// Documents
	case ".pdf":
		return "application/pdf"
	case ".epub":
		return "application/epub+zip"
	case ".cbz":
		return "application/x-cbz"
	case ".cbr":
		return "application/x-cbr"

	default:
		return "application/octet-stream"
	}
}
