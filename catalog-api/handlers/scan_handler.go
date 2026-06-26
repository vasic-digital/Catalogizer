package handlers

import (
	"catalogizer/database"
	"catalogizer/internal/services"
	"catalogizer/models"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// scannerInterface defines the scanner methods used by ScanHandler
type scannerInterface interface {
	QueueScan(job services.ScanJob) error
	GetAllActiveScanStatuses() map[string]*services.ScanStatus
	GetActiveScanStatus(jobID string) (*services.ScanStatus, bool)
}

// ScanHandler wraps UniversalScanner with REST API endpoints for
// managing storage roots and triggering scan operations.
type ScanHandler struct {
	scanner scannerInterface
	db      *database.DB
}

// NewScanHandler creates a new ScanHandler.
func NewScanHandler(scanner scannerInterface, db *database.DB) *ScanHandler {
	return &ScanHandler{scanner: scanner, db: db}
}

// createStorageRootRequest is the JSON body for POST /storage/roots.
type createStorageRootRequest struct {
	Name     string  `json:"name" binding:"required"`
	Protocol string  `json:"protocol" binding:"required"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Path     *string `json:"path"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Domain   *string `json:"domain"`
	MaxDepth int     `json:"max_depth"`
}

// supportedStorageProtocols enumerates the protocols a CreateStorageRoot
// request may declare. Closes CATAPI-DEFECT-004 (the handler used to
// accept arbitrary strings — gopher, foo, anything — and create
// undriveable storage roots).
var supportedStorageProtocols = map[string]struct{}{
	"local":  {},
	"smb":    {},
	"ftp":    {},
	"nfs":    {},
	"webdav": {},
}

// CreateStorageRoot handles POST /api/v1/storage/roots.
// Strictly creates a new storage root. If a row with the same name
// already exists, returns 409 Conflict (CATAPI-DEFECT-005). For
// idempotent upsert semantics the caller should use PUT
// /api/v1/storage/roots/{id} instead.
//
// Validates Protocol against supportedStorageProtocols
// (CATAPI-DEFECT-004). Unknown protocols return 400 with an
// explanatory message and the list of accepted values.
func (h *ScanHandler) CreateStorageRoot(c *gin.Context) {
	var req createStorageRootRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// CATAPI-DEFECT-004: enforce protocol allowlist.
	if _, ok := supportedStorageProtocols[req.Protocol]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "unsupported protocol",
			"protocol": req.Protocol,
			"accepted": []string{"local", "smb", "ftp", "nfs", "webdav"},
		})
		return
	}

	if req.MaxDepth <= 0 {
		req.MaxDepth = 10
	}

	// CATAPI-DEFECT-005: refuse duplicate names with 409. Previous
	// behavior silently upserted and returned 201, hiding the
	// conflict from clients that genuinely meant "create".
	var existingID int64
	err := h.db.QueryRowContext(c.Request.Context(),
		"SELECT id FROM storage_roots WHERE name = ?", req.Name,
	).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":       "storage root with this name already exists",
			"name":        req.Name,
			"existing_id": existingID,
			"hint":        "use PUT /api/v1/storage/roots/{id} to update an existing root",
		})
		return
	}

	newID, insertErr := h.db.InsertReturningID(c.Request.Context(),
		`INSERT INTO storage_roots (name, protocol, host, port, path, username, password, domain, enabled, max_depth)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Name, req.Protocol, req.Host, req.Port, req.Path,
		req.Username, req.Password, req.Domain, true, req.MaxDepth,
	)
	if insertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create storage root: %v", insertErr)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       newID,
		"name":     req.Name,
		"protocol": req.Protocol,
		"message":  "storage root created",
	})
}

// GetStorageRoots handles GET /api/v1/storage/roots.
// Returns all storage roots from the database.
func (h *ScanHandler) GetStorageRoots(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT id, name, protocol, host, port, path, username, domain, enabled, max_depth,
		        created_at, updated_at, last_scan_at
		 FROM storage_roots ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query storage roots: %v", err)})
		return
	}
	defer rows.Close()

	var roots []gin.H
	for rows.Next() {
		var (
			id                   int64
			name, protocol       string
			host, path, username *string
			domain               *string
			port                 *int
			enabled              bool
			maxDepth             int
			createdAt, updatedAt time.Time
			lastScanAt           *time.Time
		)
		if err := rows.Scan(&id, &name, &protocol, &host, &port, &path, &username, &domain, &enabled, &maxDepth, &createdAt, &updatedAt, &lastScanAt); err != nil {
			continue
		}
		roots = append(roots, gin.H{
			"id":           id,
			"name":         name,
			"protocol":     protocol,
			"host":         host,
			"port":         port,
			"path":         path,
			"username":     username,
			"domain":       domain,
			"enabled":      enabled,
			"max_depth":    maxDepth,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
			"last_scan_at": lastScanAt,
		})
	}

	if roots == nil {
		roots = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"roots": roots})
}

// GetStorageRootStatus handles GET /api/v1/storage-roots/:id/status.
// Returns connectivity status for a storage root.
func (h *ScanHandler) GetStorageRootStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid storage root ID"})
		return
	}

	// Check if storage root exists
	var exists bool
	err = h.db.QueryRowContext(c.Request.Context(),
		"SELECT 1 FROM storage_roots WHERE id = ?", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storage root not found"})
		return
	}

	// For now, return dummy connectivity status
	c.JSON(http.StatusOK, gin.H{
		"id":         id,
		"connected":  true,
		"status":     "online",
		"checked_at": time.Now().Format(time.RFC3339),
	})
}

// queueScanRequest is the JSON body for POST /scans.
type queueScanRequest struct {
	StorageRootID int64  `json:"storage_root_id" binding:"required"`
	Path          string `json:"path"`
	ScanType      string `json:"scan_type"`
	MaxDepth      int    `json:"max_depth"`
}

// QueueScan handles POST /api/v1/scans.
// Looks up the storage root, builds a ScanJob, and queues it.
func (h *ScanHandler) QueueScan(c *gin.Context) {
	var req queueScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ScanType == "" {
		req.ScanType = "full"
	}
	if req.MaxDepth <= 0 {
		req.MaxDepth = 10
	}

	// Load storage root from DB
	root, err := h.loadStorageRoot(c.Request.Context(), req.StorageRootID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("storage root not found: %v", err)})
		return
	}

	jobID := uuid.New().String()
	job := services.ScanJob{
		ID:          jobID,
		StorageRoot: root,
		Path:        req.Path,
		ScanType:    req.ScanType,
		MaxDepth:    req.MaxDepth,
		Context:     context.Background(),
	}

	if err := h.scanner.QueueScan(job); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("failed to queue scan: %v", err)})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":          jobID,
		"storage_root_id": req.StorageRootID,
		"scan_type":       req.ScanType,
		"status":          "queued",
		"message":         "scan job queued",
	})
}

// ListScans handles GET /api/v1/scans.
// Returns all active scan statuses.
func (h *ScanHandler) ListScans(c *gin.Context) {
	statuses := h.scanner.GetAllActiveScanStatuses()
	scans := make([]gin.H, 0, len(statuses))
	for id, s := range statuses {
		scans = append(scans, scanStatusToJSON(id, s))
	}
	c.JSON(http.StatusOK, gin.H{"scans": scans})
}

// GetScanStatus handles GET /api/v1/scans/:job_id.
// Returns the status of a specific scan job.
func (h *ScanHandler) GetScanStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	status, exists := h.scanner.GetActiveScanStatus(jobID)
	if !exists {
		// Check completed scans table if we add one later;
		// for now return a synthetic "completed" or "not found"
		c.JSON(http.StatusNotFound, gin.H{
			"job_id": jobID,
			"status": "not_found",
			"error":  "scan job not found (may have already completed)",
		})
		return
	}
	snapshot := status.GetSnapshot()
	c.JSON(http.StatusOK, scanStatusToJSON(jobID, &snapshot))
}

// loadStorageRoot reads a StorageRoot from the database by ID.
func (h *ScanHandler) loadStorageRoot(ctx context.Context, id int64) (*models.StorageRoot, error) {
	row := h.db.QueryRowContext(ctx,
		`SELECT id, name, protocol, host, port, path, username, password, domain, options, enabled, max_depth
		 FROM storage_roots WHERE id = ?`, id)

	var root models.StorageRoot
	if err := row.Scan(
		&root.ID, &root.Name, &root.Protocol,
		&root.Host, &root.Port, &root.Path,
		&root.Username, &root.Password, &root.Domain,
		&root.Options, &root.Enabled, &root.MaxDepth,
	); err != nil {
		return nil, err
	}
	return &root, nil
}

// scanStatusToJSON converts a ScanStatus to a JSON-friendly map.
func scanStatusToJSON(jobID string, s *services.ScanStatus) gin.H {
	elapsed := time.Since(s.StartTime).Milliseconds()
	return gin.H{
		"job_id":          jobID,
		"storage_root":    s.StorageRootName,
		"protocol":        s.Protocol,
		"status":          s.Status,
		"start_time":      s.StartTime,
		"elapsed_ms":      elapsed,
		"current_path":    s.CurrentPath,
		"files_processed": s.FilesProcessed,
		"files_found":     s.FilesFound,
		"files_updated":   s.FilesUpdated,
		"files_deleted":   s.FilesDeleted,
		"error_count":     s.ErrorCount,
	}
}
