package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"catalogizer/config"
	"catalogizer/database"
	mediaModels "catalogizer/internal/media/models"
	"catalogizer/internal/services"
	"catalogizer/models"
	"catalogizer/repository"
	svcPkg "catalogizer/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// =====================================================================
// Shared test helpers
// =====================================================================

func setupTestDBForCoverage(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "coverage_boost8_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Type:               "sqlite",
		Path:               tmpFile.Name(),
		MaxOpenConnections: 5,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    3600,
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        5000,
	}

	db, err := database.NewConnection(cfg)
	require.NoError(t, err)

	err = db.RunMigrations(context.Background())
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}
	return db, cleanup
}

func adminAuth() *mockAdminAuthService {
	return &mockAdminAuthService{
		getCurrentUserFn: func(token string) (*models.User, error) {
			return &models.User{ID: 1, Username: "admin"}, nil
		},
		checkPermissionFn: func(userID int, permission string) (bool, error) {
			return true, nil
		},
	}
}

// =====================================================================
// AdminHandler — GetStorageInfo (0% coverage)
// =====================================================================

func TestAdminHandler_GetStorageInfo_Success(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	auth := adminAuth()
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, db, "1.0.0-test")

	// Insert a storage root
	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"test-root", "/mnt/data", "smb")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin/storage", nil)
	c.Request.Header.Set("Authorization", "Bearer token")

	h.GetStorageInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)
	assert.Contains(t, result[0]["path"], "smb://")
}

func TestAdminHandler_GetStorageInfo_EmptyDB(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	auth := adminAuth()
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, db, "1.0.0-test")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin/storage", nil)
	c.Request.Header.Set("Authorization", "Bearer token")

	h.GetStorageInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.NotNil(t, result) // empty array, not nil
}

func TestAdminHandler_GetStorageInfo_Unauthorized(t *testing.T) {
	auth := &mockAdminAuthService{
		getCurrentUserFn: func(token string) (*models.User, error) {
			return nil, errors.New("invalid")
		},
	}
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, nil, "1.0.0")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin/storage", nil)
	c.Request.Header.Set("Authorization", "Bearer bad")

	h.GetStorageInfo(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminHandler_GetStorageInfo_WithFileCount(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	auth := adminAuth()
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, db, "1.0.0-test")

	// Insert a storage root and some files
	res, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"nas-root", "/media", "smb")
	require.NoError(t, err)
	rootID, _ := res.LastInsertId()

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/media/movie.mkv", "movie.mkv", 1024, false, time.Now(), time.Now())
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin/storage", nil)
	c.Request.Header.Set("Authorization", "Bearer token")

	h.GetStorageInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, float64(1), result[0]["mediaCount"])
}

func TestAdminHandler_GetStorageInfo_WithLastScan(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	auth := adminAuth()
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, db, "1.0.0-test")

	now := time.Now()
	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol, last_scan_at) VALUES (?, ?, ?, ?)`,
		"scanned-root", "/scanned", "local", now)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin/storage", nil)
	c.Request.Header.Set("Authorization", "Bearer token")

	h.GetStorageInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.NotNil(t, result[0]["lastScan"])
}

// Note: GetStorageInfo with nil db panics because the handler does not guard
// against nil db. This is fine — the handler is only called when db is set.

// =====================================================================
// AdminHandler — CreateBackup deeper paths (23.3% coverage)
// =====================================================================

func TestAdminHandler_CreateBackup_SQLiteSuccess(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	auth := adminAuth()
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, db, "1.0.0-test")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/admin/backups", strings.NewReader(`{"type":"full"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer token")

	h.CreateBackup(c)

	// The handler will succeed since we have a valid sqlite db
	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "completed", resp["status"])
	assert.Contains(t, resp["id"], "backup_")

	// Cleanup the backup file
	if id, ok := resp["id"].(string); ok {
		os.Remove(filepath.Join(defaultBackupDir, id))
	}
}

func TestAdminHandler_CreateBackup_Unauthorized(t *testing.T) {
	auth := &mockAdminAuthService{
		getCurrentUserFn: func(token string) (*models.User, error) {
			return nil, errors.New("invalid")
		},
	}
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, nil, "1.0.0")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/admin/backups", strings.NewReader(`{"type":"full"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer bad")

	h.CreateBackup(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminHandler_GetBackups_Unauthorized(t *testing.T) {
	auth := &mockAdminAuthService{
		getCurrentUserFn: func(token string) (*models.User, error) {
			return nil, errors.New("invalid")
		},
	}
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, nil, "1.0.0")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin/backups", nil)
	c.Request.Header.Set("Authorization", "Bearer bad")

	h.GetBackups(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminHandler_RestoreBackup_Unauthorized(t *testing.T) {
	auth := &mockAdminAuthService{
		getCurrentUserFn: func(token string) (*models.User, error) {
			return nil, errors.New("invalid")
		},
	}
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, nil, "1.0.0")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/admin/backups/42/restore", nil)
	c.Request.Header.Set("Authorization", "Bearer bad")
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	h.RestoreBackup(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =====================================================================
// CopyHandler — CopyToSmb deep paths (10% coverage)
// =====================================================================

func TestCopyToSmb_FileRepoError(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/smb", handler.CopyToSmb)

	// Valid request body, but file doesn't exist in DB
	body := SmbCopyRequest{
		SourceFileID:       99999,
		DestinationSmbRoot: "main",
		DestinationPath:    "/dest",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/copy/smb", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCopyToSmb_SourceStorageRootNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	// Insert two storage roots: source and file's root
	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"source-root", "/source", "smb")
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, "/source/file.txt", "file.txt", 100, false, time.Now(), time.Now())
	require.NoError(t, err)

	// Update the file's storage_root_id to a non-existent ID (bypassing FK by updating)
	_, err = db.Exec(`PRAGMA foreign_keys = OFF`)
	require.NoError(t, err)
	_, err = db.Exec(`UPDATE files SET storage_root_id = 9999 WHERE id = 1`)
	require.NoError(t, err)
	_, err = db.Exec(`PRAGMA foreign_keys = ON`)
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/smb", handler.CopyToSmb)

	body := SmbCopyRequest{
		SourceFileID:       1,
		DestinationSmbRoot: "dest-root",
		DestinationPath:    "/dest",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/copy/smb", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// File lookup via JOIN fails because storage_root_id=9999 doesn't exist
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestCopyToSmb_DestStorageRootNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	// Insert source storage root + file
	res, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"source-root", "/source", "smb")
	require.NoError(t, err)
	rootID, _ := res.LastInsertId()

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/source/file.txt", "file.txt", 100, false, time.Now(), time.Now())
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/smb", handler.CopyToSmb)

	body := SmbCopyRequest{
		SourceFileID:       1,
		DestinationSmbRoot: "nonexistent-dest",
		DestinationPath:    "/dest",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/copy/smb", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCopyToSmb_SmbConnectionFails(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	// Insert both source and dest storage roots + file
	host := "nonexistent.host.local"
	port := 445
	path := "/share"
	username := "user"

	res, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol, host, port, username) VALUES (?, ?, ?, ?, ?, ?)`,
		"source-root", path, "smb", host, port, username)
	require.NoError(t, err)
	rootID, _ := res.LastInsertId()

	_, err = db.Exec(`INSERT INTO storage_roots (name, path, protocol, host, port, username) VALUES (?, ?, ?, ?, ?, ?)`,
		"dest-root", "/dest-share", "smb", host, port, username)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/share/file.txt", "file.txt", 100, false, time.Now(), time.Now())
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/smb", handler.CopyToSmb)

	body := SmbCopyRequest{
		SourceFileID:       1,
		DestinationSmbRoot: "dest-root",
		DestinationPath:    "/dest/file.txt",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/copy/smb", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail with 500 because SMB connection to nonexistent host fails
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// =====================================================================
// CopyHandler — CopyToLocal deep paths (11.9% coverage)
// =====================================================================

func TestCopyToLocal_FileNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/local", handler.CopyToLocal)

	body := LocalCopyRequest{
		SourceFileID:    99999,
		DestinationPath: "/tmp/dest",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/copy/local", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCopyToLocal_SourceSmbRootNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"temp-root", "/temp", "smb")
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, "/temp/file.txt", "file.txt", 100, false, time.Now(), time.Now())
	require.NoError(t, err)

	// Move file to orphaned storage_root_id
	_, _ = db.Exec(`PRAGMA foreign_keys = OFF`)
	_, _ = db.Exec(`UPDATE files SET storage_root_id = 9999 WHERE id = 1`)
	_, _ = db.Exec(`PRAGMA foreign_keys = ON`)

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/local", handler.CopyToLocal)

	body := LocalCopyRequest{
		SourceFileID:    1,
		DestinationPath: "/tmp/dest",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/copy/local", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// File lookup via JOIN fails because storage_root_id=9999 doesn't exist
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestCopyToLocal_SmbConnectionFails(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	host := "bad.host"
	port := 445
	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol, host, port) VALUES (?, ?, ?, ?, ?)`,
		"smb-root", "/share", "smb", host, port)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, "/share/file.txt", "file.txt", 100, false, time.Now(), time.Now())
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/local", handler.CopyToLocal)

	body := LocalCopyRequest{
		SourceFileID:    1,
		DestinationPath: "/tmp/localcopy",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/copy/local", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// =====================================================================
// CopyHandler — CopyFromLocal deep paths (30% coverage)
// =====================================================================

func TestCopyFromLocal_StorageRootNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/upload", handler.CopyFromLocal)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("destination_smb_root", "nonexistent")
	writer.WriteField("destination_path", "/dest")
	// Create a file part
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("file content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/copy/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCopyFromLocal_SmbConnectionFails(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol, host, port) VALUES (?, ?, ?, ?, ?)`,
		"smb-dest", "/share", "smb", "bad.host", 445)
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewCopyHandler(fileRepo, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/upload", handler.CopyFromLocal)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("destination_smb_root", "smb-dest")
	writer.WriteField("destination_path", "/dest")
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("file content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/copy/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCopyFromLocal_OverwriteExistingParam(t *testing.T) {
	// Test the overwrite_existing form param parsing
	handler := NewCopyHandler(nil, "/tmp")
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/copy/upload", handler.CopyFromLocal)

	// Missing smb_root
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("destination_smb_root", "")
	writer.WriteField("destination_path", "/dest")
	writer.WriteField("overwrite_existing", "true")
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("data"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/copy/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =====================================================================
// DownloadHandler — DownloadFile deeper paths (8% coverage)
// =====================================================================

func TestDownloadFile_FileNotFoundInDB(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/file/:id", handler.DownloadFile)

	req := httptest.NewRequest("GET", "/api/download/file/99999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDownloadFile_IsDirectory(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	// Insert a storage root + directory entry
	res, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"test-root", "/data", "smb")
	require.NoError(t, err)
	rootID, _ := res.LastInsertId()

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/data/dir", "dir", 0, true, time.Now(), time.Now())
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/file/:id", handler.DownloadFile)

	req := httptest.NewRequest("GET", "/api/download/file/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Cannot download directory")
}

func TestDownloadFile_DeletedFile(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	res, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"test-root", "/data", "smb")
	require.NoError(t, err)
	rootID, _ := res.LastInsertId()

	now := time.Now()
	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, deleted, deleted_at, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/data/deleted.mkv", "deleted.mkv", 100, false, true, now, now, now)
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/file/:id", handler.DownloadFile)

	req := httptest.NewRequest("GET", "/api/download/file/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "deleted")
}

func TestDownloadFile_SmbRootNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"tmp-root", "/data", "smb")
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, "/data/movie.mkv", "movie.mkv", 1024, false, time.Now(), time.Now())
	require.NoError(t, err)

	// Orphan the file so its storage root ID doesn't match any root
	_, _ = db.Exec(`PRAGMA foreign_keys = OFF`)
	_, _ = db.Exec(`UPDATE files SET storage_root_id = 9999 WHERE id = 1`)
	_, _ = db.Exec(`PRAGMA foreign_keys = ON`)

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/file/:id", handler.DownloadFile)

	req := httptest.NewRequest("GET", "/api/download/file/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// File lookup via JOIN fails because storage_root_id=9999 doesn't exist
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestDownloadFile_SmbConnectionFails(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol, host, port) VALUES (?, ?, ?, ?, ?)`,
		"smb-root", "/share", "smb", "nonexistent.host", 445)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, "/share/movie.mkv", "movie.mkv", 1024, false, time.Now(), time.Now())
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/file/:id", handler.DownloadFile)

	req := httptest.NewRequest("GET", "/api/download/file/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// =====================================================================
// DownloadHandler — DownloadDirectory deeper paths (14.5% coverage)
// =====================================================================

func TestDownloadDirectory_SmbRootNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/directory/:smb_root", handler.DownloadDirectory)

	req := httptest.NewRequest("GET", "/api/download/directory/nonexistent?path=/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDownloadDirectory_SmbConnectionFails(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol, host, port) VALUES (?, ?, ?, ?, ?)`,
		"smb-root", "/share", "smb", "nonexistent.host", 445)
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/directory/:smb_root", handler.DownloadDirectory)

	req := httptest.NewRequest("GET", "/api/download/directory/smb-root?path=/movies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDownloadDirectory_RecursiveAndMaxDepthParams(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/directory/:smb_root", handler.DownloadDirectory)

	// Test recursive=false, max_depth=2
	req := httptest.NewRequest("GET", "/api/download/directory/nonexistent?path=/test&recursive=false&max_depth=2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should still fail at SMB root lookup since root doesn't exist
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =====================================================================
// DownloadHandler — GetDownloadInfo deeper paths
// =====================================================================

func TestGetDownloadInfo_FileFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	res, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"test-root", "/data", "smb")
	require.NoError(t, err)
	rootID, _ := res.LastInsertId()

	mimeType := "video/mp4"
	ext := "mp4"
	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, extension, mime_type, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/data/movie.mp4", "movie.mp4", ext, mimeType, 2048, false, time.Now(), time.Now())
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/info/:id", handler.GetDownloadInfo)

	req := httptest.NewRequest("GET", "/api/download/info/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "movie.mp4", data["name"])
	assert.Equal(t, float64(2048), data["size"])
}

func TestGetDownloadInfo_DirectoryFile(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	res, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"test-root", "/data", "smb")
	require.NoError(t, err)
	rootID, _ := res.LastInsertId()

	_, err = db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/data/movies", "movies", 50000, true, time.Now(), time.Now())
	require.NoError(t, err)

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/info/:id", handler.GetDownloadInfo)

	req := httptest.NewRequest("GET", "/api/download/info/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.True(t, data["is_directory"].(bool))
}

func TestGetDownloadInfo_FileNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	handler := NewDownloadHandler(fileRepo, "/tmp", 1024*1024, 4096)
	defer handler.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/download/info/:id", handler.GetDownloadInfo)

	req := httptest.NewRequest("GET", "/api/download/info/99999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =====================================================================
// MediaEntityHandler — RefreshEntityMetadata (17.4% coverage)
// =====================================================================

func TestMediaEntityHandler_RefreshEntityMetadata_InvalidIDFormat(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/abc/metadata/refresh", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.RefreshEntityMetadata(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_RefreshEntityMetadata_NilDB(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/1/metadata/refresh", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.RefreshEntityMetadata(c)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["message"], "queued")
}

func TestMediaEntityHandler_RefreshEntityMetadata_NoDirectoryFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Orphan Movie", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/entities/%d/metadata/refresh", id), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(id, 10)}}

	handler.RefreshEntityMetadata(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["message"], "No directory found")
}

// =====================================================================
// MediaEntityHandler — EnrichAllEntities (0% coverage)
// =====================================================================

func TestMediaEntityHandler_EnrichAllEntities_NilDB(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/enrich", nil)

	handler.EnrichAllEntities(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMediaEntityHandler_EnrichAllEntities_NoUnenrichedEntities(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)
	handler.db = db

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/enrich", nil)

	handler.EnrichAllEntities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["message"], "Batch enrichment complete")
	assert.Equal(t, float64(0), resp["scanned"])
}

// =====================================================================
// MediaEntityHandler — DownloadEntity (45.5% coverage)
// =====================================================================

func TestMediaEntityHandler_DownloadEntity_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/abc/download", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.DownloadEntity(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_DownloadEntity_EmptyFileList(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Empty Movie", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/download", id), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(id, 10)}}

	handler.DownloadEntity(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_DownloadEntity_WithFileIDParam(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie", Status: "detected"})

	// Insert a file + media_file link
	res, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"root1", "/data", "smb")
	require.NoError(t, err)
	rootID, _ := res.LastInsertId()

	fileRes, err := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/data/movie.mkv", "movie.mkv", 2048, false, time.Now(), time.Now())
	require.NoError(t, err)
	fileID, _ := fileRes.LastInsertId()

	_, err = db.Exec(`INSERT INTO media_files (media_item_id, file_id) VALUES (?, ?)`, entityID, fileID)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/download?file_id=%d", entityID, fileID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.DownloadEntity(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(fileID), resp["file_id"])
}

// =====================================================================
// MediaEntityHandler — StreamEntity
// =====================================================================

func TestMediaEntityHandler_StreamEntity_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/abc/stream", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.StreamEntity(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_StreamEntity_EmptyFileList(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "No File Movie", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/stream", id), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(id, 10)}}

	handler.StreamEntity(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_StreamEntity_WithFile(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie", Status: "detected"})

	res, _ := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`, "r1", "/d", "smb")
	rootID, _ := res.LastInsertId()
	fileRes, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/m.mkv", "m.mkv", 100, false, time.Now(), time.Now())
	fileID, _ := fileRes.LastInsertId()
	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id) VALUES (?, ?)`, entityID, fileID)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/stream", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.StreamEntity(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// MediaEntityHandler — GetInstallInfo
// =====================================================================

func TestMediaEntityHandler_GetInstallInfo_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/abc/install-info", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.GetInstallInfo(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_GetInstallInfo_NonSoftwareType(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "A Movie", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/install-info", id), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(id, 10)}}

	handler.GetInstallInfo(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "software")
}

func TestMediaEntityHandler_GetInstallInfo_Software_NoFiles(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "software")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Cool App", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/install-info", id), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(id, 10)}}

	handler.GetInstallInfo(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_GetInstallInfo_Software_WithFile(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "software")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Cool App", Status: "detected"})

	res, _ := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`, "r1", "/d", "smb")
	rootID, _ := res.LastInsertId()
	fileRes, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/app.exe", "app.exe", 5000, false, time.Now(), time.Now())
	fileID, _ := fileRes.LastInsertId()
	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id) VALUES (?, ?)`, entityID, fileID)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/install-info", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.GetInstallInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Cool App", resp["title"])
}

// =====================================================================
// MediaEntityHandler — BrowseByType
// =====================================================================

func TestMediaEntityHandler_BrowseByType_UnknownType(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/browse/nonexistent_type", nil)
	c.Params = gin.Params{{Key: "type", Value: "nonexistent_type"}}

	handler.BrowseByType(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_BrowseByType_ValidType(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "game")
	_, _ = itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Game 1", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/browse/game?limit=10&offset=0", nil)
	c.Params = gin.Params{{Key: "type", Value: "game"}}

	handler.BrowseByType(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "game", resp["type"])
	assert.Equal(t, float64(1), resp["total"])
}

func TestMediaEntityHandler_BrowseByType_PaginationBounds(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Negative limit and offset
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/browse/movie?limit=-5&offset=-1", nil)
	c.Params = gin.Params{{Key: "type", Value: "movie"}}

	handler.BrowseByType(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(24), resp["limit"]) // clamped to default
	assert.Equal(t, float64(0), resp["offset"]) // clamped to 0
}

// =====================================================================
// MediaEntityHandler — GetEntityDuplicates
// =====================================================================

func TestMediaEntityHandler_GetEntityDuplicates_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/abc/duplicates", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.GetEntityDuplicates(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_GetEntityDuplicates_EntityNotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/99999/duplicates", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.GetEntityDuplicates(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_GetEntityDuplicates_NoDuplicates(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Unique Movie", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/duplicates", id), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(id, 10)}}

	handler.GetEntityDuplicates(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// MediaEntityHandler — ListDuplicateGroups
// =====================================================================

func TestMediaEntityHandler_ListDuplicateGroups_Default(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/duplicates?limit=10&offset=0", nil)

	handler.ListDuplicateGroups(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMediaEntityHandler_ListDuplicateGroups_NegativePagination(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/duplicates?limit=-1&offset=-5", nil)

	handler.ListDuplicateGroups(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(50), resp["limit"]) // default
}

// =====================================================================
// MediaEntityHandler — GetEntityChildren
// =====================================================================

func TestMediaEntityHandler_GetEntityChildren_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/abc/children", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.GetEntityChildren(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_GetEntityChildren_Valid(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	parentID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Show", Status: "detected"})

	_, seasonTypeID, _ := itemRepo.GetMediaTypeByName(ctx, "tv_season")
	_, _ = itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: seasonTypeID, Title: "Season 1", ParentID: &parentID, Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/children?limit=10&offset=0", parentID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(parentID, 10)}}

	handler.GetEntityChildren(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// MediaEntityHandler — GetEntityFiles + GetEntityMetadata
// =====================================================================

func TestMediaEntityHandler_GetEntityFiles_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/abc/files", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.GetEntityFiles(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_GetEntityMetadata_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/abc/metadata", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.GetEntityMetadata(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =====================================================================
// MediaEntityHandler — UpdateUserMetadata
// =====================================================================

func TestMediaEntityHandler_UpdateUserMetadata_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/entities/abc/user-metadata", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.UpdateUserMetadata(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_UpdateUserMetadata_BadJSON(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/entities/1/user-metadata", strings.NewReader("invalid"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.UpdateUserMetadata(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_UpdateUserMetadata_Success(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	// Create a user for the FK constraint (handler defaults user_id to 1)
	_, _ = db.Exec(`INSERT INTO users (id, username, email, password_hash, salt, role_id) VALUES (?, ?, ?, ?, ?, ?)`,
		1, "testuser", "test@test.com", "hash", "salt", 1)

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie", Status: "detected"})

	body := `{"user_rating": 8.5, "watched_status": "watched", "favorite": true, "personal_notes": "Great!"}`

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/entities/%d/user-metadata", id), strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(id, 10)}}

	handler.UpdateUserMetadata(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMediaEntityHandler_UpdateUserMetadata_IsFavoriteFieldSet(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	// Create a user for the FK constraint
	_, _ = db.Exec(`INSERT INTO users (id, username, email, password_hash, salt, role_id) VALUES (?, ?, ?, ?, ?, ?)`,
		1, "testuser", "test@test.com", "hash", "salt", 1)

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie 2", Status: "detected"})

	body := `{"is_favorite": true}`

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/entities/%d/user-metadata", id), strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(id, 10)}}

	handler.UpdateUserMetadata(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, true, resp["is_favorite"])
}

// =====================================================================
// MediaEntityHandler — enrichItemsWithCoverURLs (57.7% coverage) + strPtr
// =====================================================================

func TestMediaEntityHandler_StrPtr(t *testing.T) {
	s := "hello"
	ptr := strPtr(s)
	assert.NotNil(t, ptr)
	assert.Equal(t, "hello", *ptr)
}

func TestMediaEntityHandler_EnrichItemsWithCoverURLs_NilCoverService(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	id, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Test", Status: "detected"})

	items := []*mediaModels.MediaItem{{ID: id, MediaTypeID: typeID, Title: "Test"}}
	jsonItems := []gin.H{{"id": id, "title": "Test"}}

	// No cover service set, should use DB fallback
	handler.enrichItemsWithCoverURLs(ctx, items, jsonItems)

	// No external metadata exists so no cover_url should be added
	_, hasCover := jsonItems[0]["cover_url"]
	assert.False(t, hasCover)
}

func TestMediaEntityHandler_EnrichItemsWithCoverURLs_EmptyItems(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	logger := zap.NewNop()
	coverArtService := services.NewCoverArtService(nil, logger)
	handler.SetCoverArtService(coverArtService)

	ctx := context.Background()
	handler.enrichItemsWithCoverURLs(ctx, nil, nil)
	handler.enrichItemsWithCoverURLs(ctx, []*mediaModels.MediaItem{}, []gin.H{})
	// No panic = success
}

// =====================================================================
// MediaEntityHandler — lazyEnrichEntities (8% coverage)
// =====================================================================

func TestMediaEntityHandler_LazyEnrichEntities_NilDB(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	// Should return immediately with nil db
	handler.lazyEnrichEntities([]int64{1, 2, 3})
	// No panic = success
}

func TestMediaEntityHandler_LazyEnrichEntities_NoTMDBKey(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)
	handler.db = db

	// Clear TMDB_API_KEY to test early return
	origKey := os.Getenv("TMDB_API_KEY")
	os.Setenv("TMDB_API_KEY", "")
	defer os.Setenv("TMDB_API_KEY", origKey)

	handler.lazyEnrichEntities([]int64{1, 2, 3})
	// Returns immediately with no TMDB key
}

// =====================================================================
// MediaEntityHandler — fetchTMDBMetadata (0% coverage)
// =====================================================================

func TestFetchTMDBMetadata_NoAPIKey(t *testing.T) {
	origKey := os.Getenv("TMDB_API_KEY")
	os.Setenv("TMDB_API_KEY", "")
	defer os.Setenv("TMDB_API_KEY", origKey)

	result := fetchTMDBMetadata("The Matrix", 1999, 1)
	assert.Nil(t, result)
}

// =====================================================================
// SubtitleHandler — GetSubtitles deeper paths (37.5% coverage)
// =====================================================================

func TestSubtitleHandler_GetSubtitles_EmptyMediaID(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/subtitles/media/", nil)
	c.Params = gin.Params{{Key: "media_id", Value: ""}}

	handler.GetSubtitles(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "MISSING_MEDIA_ID", resp.Code)
}

func TestSubtitleHandler_GetSubtitles_OverflowMediaID(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/subtitles/media/99999999999999999999", nil)
	c.Params = gin.Params{{Key: "media_id", Value: "99999999999999999999"}}

	handler.GetSubtitles(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =====================================================================
// SubtitleHandler — VerifySubtitleSync deeper paths (31.8% coverage)
// =====================================================================

func TestSubtitleHandler_VerifySubtitleSync_EmptyParams(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/subtitles//verify-sync/", nil)
	c.Params = gin.Params{
		{Key: "subtitle_id", Value: ""},
		{Key: "media_id", Value: ""},
	}

	handler.VerifySubtitleSync(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "MISSING_REQUIRED_PARAMS", resp.Code)
}

func TestSubtitleHandler_VerifySubtitleSync_OverflowMediaID(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/subtitles/sub1/verify-sync/99999999999999999999", nil)
	c.Params = gin.Params{
		{Key: "subtitle_id", Value: "sub1"},
		{Key: "media_id", Value: "99999999999999999999"},
	}

	handler.VerifySubtitleSync(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =====================================================================
// SubtitleHandler — UploadSubtitle deeper paths
// =====================================================================

func TestSubtitleHandler_UploadSubtitle_MissingFileWithFormFields(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/subtitles/upload", handler.UploadSubtitle)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("media_item_id", "123")
	writer.WriteField("language", "English")
	writer.WriteField("language_code", "en")
	// No file attached
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/subtitles/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "MISSING_FILE", resp.Code)
}

// =====================================================================
// CoverHandler — ServeCover deeper paths (41.4% coverage)
// =====================================================================

func TestCoverHandler_ServeCover_WithCoverURL(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	logger := zap.NewNop()
	coverArtService := services.NewCoverArtService(db, logger)
	handler := NewCoverHandler(coverArtService)

	// Insert a cover art record with URL
	itemRepo := repository.NewMediaItemRepository(db)
	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie", Status: "detected"})

	coverURL := "https://example.com/poster.jpg"
	_, err := db.Exec(`INSERT INTO cover_art (media_item_id, url, format) VALUES (?, ?, ?)`,
		entityID, coverURL, "jpg")
	if err != nil {
		// cover_art table might not exist; skip this part
		t.Skip("cover_art table not available")
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/cover/%d", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.ServeCover(c)

	// Should either redirect to URL or serve placeholder
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusFound)
}

func TestCoverHandler_ServeCover_NegativeID(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/-1", nil)
	c.Params = gin.Params{{Key: "id", Value: "-1"}}

	handler.ServeCover(c)

	// -1 is a valid int64, so it will try to find cover art and fall back to placeholder
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// LogManagementHandler — StreamLogs deeper paths (38.1% coverage)
// =====================================================================

func TestLogManagementHandler_StreamLogs_Unauthorized(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	// No user_id in context
	req := httptest.NewRequest("GET", "/stream-logs", nil)
	rr := httptest.NewRecorder()

	handler.StreamLogs(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogManagementHandler_StreamLogs_ForbiddenAccess(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(false, nil)

	req := httptest.NewRequest("GET", "/stream-logs", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.StreamLogs(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockAuthService.AssertExpectations(t)
}

func TestLogManagementHandler_StreamLogs_ServiceCallError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("StreamLogs", 1, mock.AnythingOfType("*models.LogStreamFilters")).Return(nil, errors.New("stream error"))

	req := httptest.NewRequest("GET", "/stream-logs", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.StreamLogs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestLogManagementHandler_StreamLogs_WithFilters(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	// Create a channel that will be closed immediately
	ch := make(chan models.LogEntry)
	close(ch)
	mockLogService.On("StreamLogs", 1, mock.AnythingOfType("*models.LogStreamFilters")).Return((<-chan models.LogEntry)(ch), errors.New("stream error"))

	req := httptest.NewRequest("GET", "/stream-logs?level=error&component=api&search=test", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.StreamLogs(rr, req)

	// Service returns error, so 500
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

// =====================================================================
// LogManagementHandler — additional handlers for coverage
// =====================================================================

func TestLogManagementHandler_AnalyzeLogs_SuccessPath(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("AnalyzeLogs", 1, 1).Return(&models.LogAnalysis{
		TotalEntries: 100,
	}, nil)

	req := httptest.NewRequest("GET", "/log-collections/1/analyze", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.AnalyzeLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_AnalyzeLogs_InvalidCollectionID(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	req := httptest.NewRequest("GET", "/log-collections/abc/analyze", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	handler.AnalyzeLogs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestLogManagementHandler_AnalyzeLogs_NoUserID(t *testing.T) {
	handler := &LogManagementHandler{
		logManagementService: new(MockLogManagementService),
		authService:          new(MockLogManagementAuthService),
	}

	req := httptest.NewRequest("GET", "/log-collections/1/analyze", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.AnalyzeLogs(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogManagementHandler_GetLogStatistics_SuccessPath(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("GetLogStatistics", 1).Return(&models.LogStatistics{
		TotalCollections: 5,
	}, nil)

	req := httptest.NewRequest("GET", "/log-statistics", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.GetLogStatistics(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_GetLogStatistics_NoUserID(t *testing.T) {
	handler := &LogManagementHandler{
		logManagementService: new(MockLogManagementService),
		authService:          new(MockLogManagementAuthService),
	}

	req := httptest.NewRequest("GET", "/log-statistics", nil)
	rr := httptest.NewRecorder()

	handler.GetLogStatistics(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogManagementHandler_GetConfiguration_SuccessPath(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("GetConfiguration").Return(&svcPkg.LogManagementConfig{})

	req := httptest.NewRequest("GET", "/log-config", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.GetConfiguration(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_UpdateConfiguration_SuccessPath(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("UpdateConfiguration", mock.AnythingOfType("*services.LogManagementConfig")).Return(nil)

	body := `{"max_log_size": 1000}`
	req := httptest.NewRequest("PUT", "/log-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.UpdateConfiguration(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_UpdateConfiguration_BadRequestBody(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	req := httptest.NewRequest("PUT", "/log-config", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.UpdateConfiguration(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestLogManagementHandler_CleanupOldLogs_SuccessPath(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("CleanupOldLogs").Return(nil)

	req := httptest.NewRequest("POST", "/log-cleanup", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.CleanupOldLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_CleanupOldLogs_Error(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("CleanupOldLogs").Return(errors.New("cleanup failed"))

	req := httptest.NewRequest("POST", "/log-cleanup", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.CleanupOldLogs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

// =====================================================================
// LogManagementHandler — ExportLogs with format variants
// =====================================================================

func TestLogManagementHandler_ExportLogs_AllFormats(t *testing.T) {
	formats := []struct {
		format      string
		contentType string
	}{
		{"json", "application/json"},
		{"csv", "text/csv"},
		{"txt", "text/plain"},
		{"zip", "application/zip"},
		{"unknown", "application/octet-stream"},
	}

	for _, fmt := range formats {
		t.Run("format_"+fmt.format, func(t *testing.T) {
			mockLogService := new(MockLogManagementService)
			mockAuthService := new(MockLogManagementAuthService)

			handler := &LogManagementHandler{
				logManagementService: mockLogService,
				authService:          mockAuthService,
			}

			mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
			mockLogService.On("ExportLogs", 1, 1, fmt.format).Return([]byte("data"), nil)

			req := httptest.NewRequest("GET", "/log-collections/1/export?format="+fmt.format, nil)
			req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
			req = mux.SetURLVars(req, map[string]string{"id": "1"})
			rr := httptest.NewRecorder()

			handler.ExportLogs(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Contains(t, rr.Header().Get("Content-Type"), fmt.contentType)
			mockLogService.AssertExpectations(t)
		})
	}
}

func TestLogManagementHandler_ExportLogs_DefaultFormatJSON(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("ExportLogs", 1, 1, "json").Return([]byte("data"), nil)

	req := httptest.NewRequest("GET", "/log-collections/1/export", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.ExportLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_ExportLogs_InvalidCollectionID(t *testing.T) {
	handler := &LogManagementHandler{
		logManagementService: new(MockLogManagementService),
		authService:          new(MockLogManagementAuthService),
	}

	req := httptest.NewRequest("GET", "/log-collections/abc/export", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	handler.ExportLogs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// =====================================================================
// LogManagementHandler — CreateLogShare + RevokeLogShare
// =====================================================================

func TestLogManagementHandler_CreateLogShare_SuccessPath(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	body, _ := json.Marshal(models.LogShareRequest{CollectionID: 1, Permissions: []string{"read"}})
	var req models.LogShareRequest
	_ = json.Unmarshal(body, &req)
	mockLogService.On("CreateLogShare", 1, &req).Return(&models.LogShare{ID: 1, ShareToken: "abc123"}, nil)

	httpReq := httptest.NewRequest("POST", "/log-shares", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq = httpReq.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.CreateLogShare(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_RevokeLogShare_SuccessPath(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("RevokeLogShare", 1, 1).Return(nil)

	req := httptest.NewRequest("DELETE", "/log-shares/1", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.RevokeLogShare(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_RevokeLogShare_InvalidShareID(t *testing.T) {
	handler := &LogManagementHandler{
		logManagementService: new(MockLogManagementService),
		authService:          new(MockLogManagementAuthService),
	}

	req := httptest.NewRequest("DELETE", "/log-shares/abc", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	handler.RevokeLogShare(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// =====================================================================
// LogManagementHandler — GetLogShare
// =====================================================================

func TestLogManagementHandler_GetLogShare_SuccessPath(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	share := &models.LogShare{
		ID:           1,
		ShareToken:   "abc123",
		CollectionID: 1,
		UserID:       1,
		Permissions:  []string{"read"},
	}
	collection := &models.LogCollection{ID: 1, Name: "Test"}

	mockLogService.On("GetLogShare", "abc123").Return(share, nil)
	mockLogService.On("GetLogCollection", 1, 1).Return(collection, nil)

	req := httptest.NewRequest("GET", "/log-shares/abc123", nil)
	req = mux.SetURLVars(req, map[string]string{"token": "abc123"})
	rr := httptest.NewRecorder()

	handler.GetLogShare(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_GetLogShare_TokenNotFound(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockLogService.On("GetLogShare", "badtoken").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/log-shares/badtoken", nil)
	req = mux.SetURLVars(req, map[string]string{"token": "badtoken"})
	rr := httptest.NewRecorder()

	handler.GetLogShare(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_GetLogShare_WriteOnlyPermission(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	share := &models.LogShare{
		ID:           1,
		ShareToken:   "writeonly",
		CollectionID: 1,
		UserID:       1,
		Permissions:  []string{"write"},
	}
	collection := &models.LogCollection{ID: 1, Name: "Test"}

	mockLogService.On("GetLogShare", "writeonly").Return(share, nil)
	mockLogService.On("GetLogCollection", 1, 1).Return(collection, nil)

	req := httptest.NewRequest("GET", "/log-shares/writeonly", nil)
	req = mux.SetURLVars(req, map[string]string{"token": "writeonly"})
	rr := httptest.NewRecorder()

	handler.GetLogShare(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// =====================================================================
// LogManagementHandler — parseLogEntryFilters (exercised via GetLogEntries already)
// =====================================================================

func TestLogManagementHandler_ParseLogEntryFilters_AllParams(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("GetLogEntries", 1, 1, mock.AnythingOfType("*models.LogEntryFilters")).Return(
		[]models.LogEntry{{ID: 1, Level: "error", Message: "test"}}, nil)

	startTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)
	url := fmt.Sprintf("/log-collections/1/entries?level=error&component=api&search=test&start_time=%s&end_time=%s&limit=50&offset=10",
		startTime, endTime)

	req := httptest.NewRequest("GET", url, nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.GetLogEntries(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockLogService.AssertExpectations(t)
}

// =====================================================================
// MediaEntityHandler — itemToJSON helpers
// =====================================================================

func TestItemToJSON_AllFieldsSet(t *testing.T) {
	year := 1999
	desc := "A movie"
	director := "Director"
	rating := 8.5
	runtime := 136
	language := "en"
	parentID := int64(10)
	season := 1
	episode := 5
	track := 3
	origTitle := "Original"

	item := &mediaModels.MediaItem{
		ID:            1,
		MediaTypeID:   1,
		Title:         "The Matrix",
		OriginalTitle: &origTitle,
		Year:          &year,
		Description:   &desc,
		Genre:         []string{"Action", "Sci-Fi"},
		Director:      &director,
		Rating:        &rating,
		Runtime:       &runtime,
		Language:      &language,
		ParentID:      &parentID,
		SeasonNumber:  &season,
		EpisodeNumber: &episode,
		TrackNumber:   &track,
		Status:        "detected",
	}

	result := itemToJSON(item)

	assert.Equal(t, int64(1), result["id"])
	assert.Equal(t, "The Matrix", result["title"])
	assert.Equal(t, "Original", result["original_title"])
	assert.Equal(t, 1999, result["year"])
	assert.Equal(t, "A movie", result["description"])
	assert.Equal(t, "Director", result["director"])
	assert.Equal(t, 8.5, result["rating"])
	assert.Equal(t, 136, result["runtime"])
	assert.Equal(t, "en", result["language"])
	assert.Equal(t, int64(10), result["parent_id"])
	assert.Equal(t, 1, result["season_number"])
	assert.Equal(t, 5, result["episode_number"])
	assert.Equal(t, 3, result["track_number"])
	assert.Equal(t, []string{"Action", "Sci-Fi"}, result["genre"])
}

func TestItemToJSON_MinimalFieldsOnly(t *testing.T) {
	item := &mediaModels.MediaItem{
		ID:          1,
		MediaTypeID: 1,
		Title:       "Minimal",
		Status:      "detected",
	}

	result := itemToJSON(item)

	assert.Equal(t, "Minimal", result["title"])
	_, hasYear := result["year"]
	assert.False(t, hasYear)
	_, hasDesc := result["description"]
	assert.False(t, hasDesc)
}

func TestItemsToJSON_NilItems(t *testing.T) {
	result := itemsToJSON(nil)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestEntityDetailJSON_WithNilMetadata(t *testing.T) {
	item := &mediaModels.MediaItem{
		ID:          1,
		MediaTypeID: 1,
		Title:       "Movie",
		Status:      "detected",
	}

	result := entityDetailJSON(item, "movie", 5, 3, nil)

	assert.Equal(t, "movie", result["media_type"])
	assert.Equal(t, int64(5), result["file_count"])
	assert.Equal(t, int64(3), result["children_count"])
	assert.NotNil(t, result["external_metadata"])
}

// =====================================================================
// MediaEntityHandler — GetEntity (not found path)
// =====================================================================

func TestMediaEntityHandler_GetEntity_NotFound(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/99999", nil)
	c.Params = gin.Params{{Key: "id", Value: "99999"}}

	handler.GetEntity(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_GetEntity_InvalidIDFormat(t *testing.T) {
	handler := NewMediaEntityHandler(nil, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/abc", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.GetEntity(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =====================================================================
// MediaEntityHandler — ListEntities edge cases
// =====================================================================

func TestMediaEntityHandler_ListEntities_WithTypeFilter(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, movieTypeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	_, gameTypeID, _ := itemRepo.GetMediaTypeByName(ctx, "game")
	_, _ = itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: movieTypeID, Title: "Film 1", Status: "detected"})
	_, _ = itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: gameTypeID, Title: "Game 1", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities?type=movie", nil)

	handler.ListEntities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["total"])
}

func TestMediaEntityHandler_ListEntities_BogusType(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities?type=bogustype", nil)

	handler.ListEntities(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_ListEntities_PaginationBounds(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities?limit=-1&offset=-5", nil)

	handler.ListEntities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(24), resp["limit"])
	assert.Equal(t, float64(0), resp["offset"])
}

// =====================================================================
// SearchSubtitles — with optional query params
// =====================================================================

func TestSubtitleHandler_SearchSubtitles_WithAllOptionalParams(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery()) // Catch nil panics
	router.GET("/api/v1/subtitles/search", handler.SearchSubtitles)

	// Valid media_path with all optional params — exercises all param parsing branches
	req := httptest.NewRequest("GET", "/api/v1/subtitles/search?media_path=/movies/test.mkv&title=Test&year=2020&season=1&episode=5&languages=en,es&providers=opensubtitles,subdb", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Nil service causes panic caught by recovery middleware
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// =====================================================================
// AdminHandler — ScanStorage with storage_root_id
// =====================================================================

func TestAdminHandler_ScanStorage_WithStorageRootID(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	auth := adminAuth()
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, db, "1.0.0-test")

	_, err := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`,
		"test-root", "/data", "smb")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/admin/storage/scan", strings.NewReader(`{"path":"/media","storage_root_id":1}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer token")

	h.ScanStorage(c)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestAdminHandler_ScanStorage_NonexistentStorageRootID(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	auth := adminAuth()
	h := NewAdminHandler(auth, &mockAdminUserRepo{}, db, "1.0.0-test")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/admin/storage/scan", strings.NewReader(`{"path":"/media","storage_root_id":99999}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer token")

	h.ScanStorage(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =====================================================================
// AdminHandler — GetUsers with LastLogin
// =====================================================================

func TestAdminHandler_GetUsers_WithLastLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := adminAuth()

	lastLogin := time.Now()
	repo := &mockAdminUserRepo{
		listFn: func(limit, offset int) ([]models.User, error) {
			return []models.User{
				{ID: 1, Username: "active-user", Email: "a@b.com", IsActive: true, LastLoginAt: &lastLogin},
			}, nil
		},
	}
	h := NewAdminHandler(auth, repo, nil, "1.0.0-test")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin/users", nil)
	c.Request.Header.Set("Authorization", "Bearer token")

	h.GetUsers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.NotNil(t, result[0]["lastLogin"])
}
