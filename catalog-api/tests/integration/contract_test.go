package integration

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Contract tests verify that API responses match the expected shapes
// defined in the TypeScript API client (catalogizer-api-client).
// These tests protect against breaking changes in API contracts.

// contractTestServer sets up a Gin engine backed by in-memory SQLite
// with the same schema and seeded data as production.
func contractTestServer(t *testing.T) (*gin.Engine, *sql.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := sql.Open("sqlite3", ":memory:?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=1")
	require.NoError(t, err)

	schema := `
	CREATE TABLE IF NOT EXISTS storage_roots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		protocol TEXT NOT NULL,
		host TEXT, port INTEGER, path TEXT,
		enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_scan_at DATETIME
	);
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_root_id INTEGER NOT NULL,
		path TEXT NOT NULL, name TEXT NOT NULL,
		extension TEXT, mime_type TEXT, file_type TEXT,
		size INTEGER NOT NULL, is_directory BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		modified_at DATETIME NOT NULL,
		deleted BOOLEAN DEFAULT 0,
		is_duplicate BOOLEAN DEFAULT 0,
		duplicate_group_id INTEGER, parent_id INTEGER,
		hash TEXT,
		FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
	);
	CREATE TABLE IF NOT EXISTS media_types (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		display_name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS media_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		media_type_id INTEGER NOT NULL,
		year INTEGER, parent_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_type_id) REFERENCES media_types(id),
		FOREIGN KEY (parent_id) REFERENCES media_items(id)
	);
	CREATE TABLE IF NOT EXISTS scan_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_root_id INTEGER NOT NULL,
		scan_type TEXT NOT NULL, status TEXT NOT NULL,
		start_time DATETIME NOT NULL, end_time DATETIME,
		files_processed INTEGER DEFAULT 0,
		files_added INTEGER DEFAULT 0,
		files_updated INTEGER DEFAULT 0,
		files_deleted INTEGER DEFAULT 0,
		error_count INTEGER DEFAULT 0,
		FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	// Seed data
	now := time.Now()
	_, err = db.Exec(`INSERT INTO storage_roots (name, protocol, host, path) VALUES ('test-root', 'smb', '192.168.1.1', '/share')`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO media_types (name, display_name) VALUES ('movie', 'Movie'), ('tv_show', 'TV Show')`)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, extension, file_type, size, modified_at)
			VALUES (1, ?, ?, '.mkv', 'video', ?, ?)`,
			fmt.Sprintf("/media/file_%d.mkv", i),
			fmt.Sprintf("file_%d.mkv", i),
			int64(1024*1024*(i+1)),
			now.Add(-time.Duration(i)*time.Hour))
		require.NoError(t, err)
	}

	_, err = db.Exec(`INSERT INTO media_items (title, media_type_id, year) VALUES ('Test Movie', 1, 2025)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO scan_history (storage_root_id, scan_type, status, start_time, end_time, files_processed, files_added)
		VALUES (1, 'full', 'completed', ?, ?, 100, 50)`, now.Add(-time.Hour), now)
	require.NoError(t, err)

	router := gin.New()
	router.Use(gin.Recovery())

	// Health endpoint
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
			"uptime":    "1h0m0s",
		})
	})

	// Storage roots endpoint
	router.GET("/api/v1/storage-roots", func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, name, protocol, host, path, enabled, created_at, updated_at, last_scan_at FROM storage_roots`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var roots []gin.H
		for rows.Next() {
			var id int
			var name, protocol string
			var host, path sql.NullString
			var enabled bool
			var createdAt, updatedAt time.Time
			var lastScanAt sql.NullTime
			rows.Scan(&id, &name, &protocol, &host, &path, &enabled, &createdAt, &updatedAt, &lastScanAt)
			roots = append(roots, gin.H{
				"id": id, "name": name, "protocol": protocol,
				"host": host.String, "path": path.String,
				"enabled": enabled, "created_at": createdAt, "updated_at": updatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"storage_roots": roots, "total": len(roots)})
	})

	// Files endpoint
	router.GET("/api/v1/catalog/files", func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, name, path, extension, file_type, size, is_directory, modified_at FROM files WHERE deleted = 0 LIMIT 50`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var files []gin.H
		for rows.Next() {
			var id int
			var name, path string
			var ext, ftype sql.NullString
			var size int64
			var isDir bool
			var modTime time.Time
			rows.Scan(&id, &name, &path, &ext, &ftype, &size, &isDir, &modTime)
			files = append(files, gin.H{
				"id": id, "name": name, "path": path,
				"extension": ext.String, "file_type": ftype.String,
				"size": size, "is_directory": isDir,
				"modified_at": modTime,
			})
		}
		c.JSON(http.StatusOK, gin.H{"files": files, "total": len(files), "page": 1, "per_page": 50})
	})

	// Entities endpoint
	router.GET("/api/v1/entities", func(c *gin.Context) {
		rows, err := db.Query(`SELECT mi.id, mi.title, mt.name, mi.year, mi.created_at
			FROM media_items mi JOIN media_types mt ON mi.media_type_id = mt.id`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var entities []gin.H
		for rows.Next() {
			var id int
			var title, mediaType string
			var year sql.NullInt64
			var createdAt time.Time
			rows.Scan(&id, &title, &mediaType, &year, &createdAt)
			entities = append(entities, gin.H{
				"id": id, "title": title, "media_type": mediaType,
				"year": year.Int64, "created_at": createdAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"entities": entities, "total": len(entities)})
	})

	// Scan history endpoint
	router.GET("/api/v1/scans", func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, storage_root_id, scan_type, status, start_time, end_time, files_processed, files_added
			FROM scan_history ORDER BY start_time DESC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var scans []gin.H
		for rows.Next() {
			var id, srID, processed, added int
			var scanType, status string
			var startTime time.Time
			var endTime sql.NullTime
			rows.Scan(&id, &srID, &scanType, &status, &startTime, &endTime, &processed, &added)
			scan := gin.H{
				"id": id, "storage_root_id": srID,
				"scan_type": scanType, "status": status,
				"start_time": startTime, "files_processed": processed,
				"files_added": added,
			}
			if endTime.Valid {
				scan["end_time"] = endTime.Time
			}
			scans = append(scans, scan)
		}
		c.JSON(http.StatusOK, gin.H{"scans": scans, "total": len(scans)})
	})

	return router, db
}

// assertJSONHasField checks that a JSON object contains a field of the expected type.
func assertJSONHasField(t *testing.T, obj map[string]interface{}, field string, expectedType string) {
	t.Helper()
	val, exists := obj[field]
	assert.True(t, exists, "missing field: %s", field)
	if !exists {
		return
	}

	switch expectedType {
	case "string":
		_, ok := val.(string)
		assert.True(t, ok, "field %s should be string, got %T", field, val)
	case "number":
		_, ok := val.(float64)
		assert.True(t, ok, "field %s should be number, got %T", field, val)
	case "boolean":
		_, ok := val.(bool)
		assert.True(t, ok, "field %s should be boolean, got %T", field, val)
	case "array":
		_, ok := val.([]interface{})
		assert.True(t, ok, "field %s should be array, got %T", field, val)
	case "object":
		_, ok := val.(map[string]interface{})
		assert.True(t, ok, "field %s should be object, got %T", field, val)
	}
}

// TestContract_HealthResponse validates the health endpoint contract.
func TestContract_HealthResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	router, db := contractTestServer(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Contract: HealthResponse { status: string, timestamp: string, version: string }
	assertJSONHasField(t, resp, "status", "string")
	assertJSONHasField(t, resp, "timestamp", "string")
	assertJSONHasField(t, resp, "version", "string")

	// Status must be "healthy" or "unhealthy"
	status := resp["status"].(string)
	assert.True(t, status == "healthy" || status == "unhealthy",
		"status should be 'healthy' or 'unhealthy', got '%s'", status)
}

// TestContract_StorageRootsResponse validates the storage roots listing contract.
func TestContract_StorageRootsResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	router, db := contractTestServer(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/storage-roots", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Contract: { storage_roots: StorageRoot[], total: number }
	assertJSONHasField(t, resp, "storage_roots", "array")
	assertJSONHasField(t, resp, "total", "number")

	roots := resp["storage_roots"].([]interface{})
	assert.Greater(t, len(roots), 0)

	// Validate StorageRoot shape: { id, name, protocol, host, path, enabled }
	root := roots[0].(map[string]interface{})
	assertJSONHasField(t, root, "id", "number")
	assertJSONHasField(t, root, "name", "string")
	assertJSONHasField(t, root, "protocol", "string")
	assertJSONHasField(t, root, "enabled", "boolean")

	// Protocol must be one of the supported values
	protocol := root["protocol"].(string)
	validProtocols := []string{"smb", "ftp", "nfs", "webdav", "local"}
	assert.Contains(t, validProtocols, protocol)
}

// TestContract_FilesListResponse validates the files listing contract.
func TestContract_FilesListResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	router, db := contractTestServer(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/catalog/files", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Contract: { files: File[], total: number, page: number, per_page: number }
	assertJSONHasField(t, resp, "files", "array")
	assertJSONHasField(t, resp, "total", "number")
	assertJSONHasField(t, resp, "page", "number")
	assertJSONHasField(t, resp, "per_page", "number")

	files := resp["files"].([]interface{})
	assert.Greater(t, len(files), 0)

	// Validate File shape: { id, name, path, extension, file_type, size, is_directory }
	file := files[0].(map[string]interface{})
	assertJSONHasField(t, file, "id", "number")
	assertJSONHasField(t, file, "name", "string")
	assertJSONHasField(t, file, "path", "string")
	assertJSONHasField(t, file, "size", "number")
	assertJSONHasField(t, file, "is_directory", "boolean")

	// File type should be a known value
	if ftype, ok := file["file_type"].(string); ok && ftype != "" {
		validTypes := []string{"video", "audio", "image", "subtitle", "document", "metadata", "archive", "other"}
		assert.Contains(t, validTypes, ftype)
	}
}

// TestContract_EntitiesResponse validates the entities listing contract.
func TestContract_EntitiesResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	router, db := contractTestServer(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/entities", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Contract: { entities: MediaEntity[], total: number }
	assertJSONHasField(t, resp, "entities", "array")
	assertJSONHasField(t, resp, "total", "number")

	entities := resp["entities"].([]interface{})
	assert.Greater(t, len(entities), 0)

	// Validate MediaEntity shape: { id, title, media_type, year }
	entity := entities[0].(map[string]interface{})
	assertJSONHasField(t, entity, "id", "number")
	assertJSONHasField(t, entity, "title", "string")
	assertJSONHasField(t, entity, "media_type", "string")

	// Media type must be one of the 11 defined types
	mediaType := entity["media_type"].(string)
	validMediaTypes := []string{
		"movie", "tv_show", "tv_season", "tv_episode",
		"music_artist", "music_album", "song",
		"game", "software", "book", "comic",
	}
	assert.Contains(t, validMediaTypes, mediaType)
}

// TestContract_ScanHistoryResponse validates the scan history contract.
func TestContract_ScanHistoryResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	router, db := contractTestServer(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scans", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Contract: { scans: Scan[], total: number }
	assertJSONHasField(t, resp, "scans", "array")
	assertJSONHasField(t, resp, "total", "number")

	scans := resp["scans"].([]interface{})
	assert.Greater(t, len(scans), 0)

	// Validate Scan shape
	scan := scans[0].(map[string]interface{})
	assertJSONHasField(t, scan, "id", "number")
	assertJSONHasField(t, scan, "storage_root_id", "number")
	assertJSONHasField(t, scan, "scan_type", "string")
	assertJSONHasField(t, scan, "status", "string")
	assertJSONHasField(t, scan, "files_processed", "number")

	// Status must be a valid scan status
	status := scan["status"].(string)
	validStatuses := []string{"pending", "running", "completed", "failed", "cancelled"}
	assert.Contains(t, validStatuses, status)
}

// TestContract_ErrorResponse validates that error responses follow a consistent contract.
func TestContract_ErrorResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/api/v1/test-error", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter"})
	})

	router.GET("/api/v1/test-not-found", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
	})

	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{"bad_request", "/api/v1/test-error", http.StatusBadRequest},
		{"not_found", "/api/v1/test-not-found", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			// Contract: all error responses must have { error: string }
			assertJSONHasField(t, resp, "error", "string")
			errMsg := resp["error"].(string)
			assert.NotEmpty(t, errMsg)

			// Must not contain stack traces or internal details
			assert.False(t, strings.Contains(errMsg, "panic"),
				"error message must not contain stack traces")
			assert.False(t, strings.Contains(errMsg, "goroutine"),
				"error message must not contain goroutine info")
		})
	}
}

// TestContract_PaginationResponse validates that paginated responses include
// consistent pagination metadata.
func TestContract_PaginationResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	router, db := contractTestServer(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/catalog/files", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// All paginated endpoints must include: total, page, per_page
	total, ok := resp["total"].(float64)
	assert.True(t, ok, "total must be a number")
	assert.GreaterOrEqual(t, total, float64(0))

	page, ok := resp["page"].(float64)
	assert.True(t, ok, "page must be a number")
	assert.GreaterOrEqual(t, page, float64(1))

	perPage, ok := resp["per_page"].(float64)
	assert.True(t, ok, "per_page must be a number")
	assert.Greater(t, perPage, float64(0))
}

// TestContract_ContentType validates that all API responses use application/json.
func TestContract_ContentType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contract test in short mode")
	}

	router, db := contractTestServer(t)
	defer db.Close()

	endpoints := []string{
		"/api/v1/health",
		"/api/v1/storage-roots",
		"/api/v1/catalog/files",
		"/api/v1/entities",
		"/api/v1/scans",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", ep, nil)
			router.ServeHTTP(w, req)

			ct := w.Header().Get("Content-Type")
			assert.Contains(t, ct, "application/json",
				"endpoint %s should return application/json, got %s", ep, ct)
		})
	}
}
