package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCopyHandler_CopyToStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	handler := &CopyHandler{
		logger: logger,
	}

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantError  bool
	}{
		{
			name: "Valid copy request",
			body: map[string]string{
				"source_path": "/tmp/test.txt",
				"dest_path":   "/storage/test.txt",
				"storage_id":  "local",
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "Missing source_path",
			body: map[string]string{
				"dest_path":  "/storage/test.txt",
				"storage_id": "local",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "Missing dest_path",
			body: map[string]string{
				"source_path": "/tmp/test.txt",
				"storage_id":  "local",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "Missing storage_id",
			body: map[string]string{
				"source_path": "/tmp/test.txt",
				"dest_path":   "/storage/test.txt",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "Empty request body",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.body)
			c.Request = httptest.NewRequest("POST", "/copy/storage", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CopyToStorage(c)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.wantError {
				assert.Contains(t, response, "error")
			} else {
				assert.Contains(t, response, "message")
				assert.Equal(t, "File copied to storage successfully", response["message"])
				assert.Equal(t, tt.body["source_path"], response["source"])
				assert.Equal(t, tt.body["dest_path"], response["destination"])
				assert.Equal(t, tt.body["storage_id"], response["storage_id"])
			}
		})
	}
}

func TestCopyHandler_CopyToStorage_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	handler := &CopyHandler{
		logger: logger,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Invalid JSON
	c.Request = httptest.NewRequest("POST", "/copy/storage", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CopyToStorage(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestCopyHandler_ListStoragePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	handler := &CopyHandler{
		logger: logger,
	}

	tests := []struct {
		name       string
		path       string
		storageID  string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "Valid list request",
			path:       "test",
			storageID:  "local",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "Valid list request with nested path",
			path:       "test/subdir",
			storageID:  "local",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "Missing storage_id",
			path:       "test",
			storageID:  "",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "Empty path with storage_id",
			path:       "",
			storageID:  "local",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			url := "/storage/list/" + tt.path
			if tt.storageID != "" {
				url += "?storage_id=" + tt.storageID
			}
			c.Request = httptest.NewRequest("GET", url, nil)
			c.Params = gin.Params{{Key: "path", Value: tt.path}}

			// Manually set query parameters
			if tt.storageID != "" {
				c.Request.URL.RawQuery = "storage_id=" + tt.storageID
			}

			handler.ListStoragePath(c)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.wantError {
				assert.Contains(t, response, "error")
			} else {
				assert.Contains(t, response, "path")
				assert.Contains(t, response, "storage_id")
				assert.Contains(t, response, "files")
				assert.Equal(t, tt.path, response["path"])
				assert.Equal(t, tt.storageID, response["storage_id"])
			}
		})
	}
}

func TestCopyHandler_GetStorageRoots(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	handler := &CopyHandler{
		logger: logger,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/storage/roots", nil)

	handler.GetStorageRoots(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "roots")

	roots, ok := response["roots"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, roots)

	// Verify structure of roots
	assert.GreaterOrEqual(t, len(roots), 2)

	// Check first root
	firstRoot, ok := roots[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, firstRoot, "id")
	assert.Contains(t, firstRoot, "name")
	assert.Contains(t, firstRoot, "path")
	assert.Equal(t, "local", firstRoot["id"])
	assert.Equal(t, "Local Storage", firstRoot["name"])
	assert.Equal(t, "/data/storage", firstRoot["path"])

	// Check second root
	secondRoot, ok := roots[1].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, secondRoot, "id")
	assert.Contains(t, secondRoot, "name")
	assert.Contains(t, secondRoot, "path")
	assert.Equal(t, "smb", secondRoot["id"])
	assert.Equal(t, "SMB Storage", secondRoot["name"])
	assert.Equal(t, "smb://server/share", secondRoot["path"])
}

func TestCopyHandler_GetStorageRoots_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	handler := &CopyHandler{
		logger: logger,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/storage/roots", nil)

	handler.GetStorageRoots(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify content type
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	// Verify the response is valid JSON
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestCopyHandler_NewCopyHandler(t *testing.T) {
	logger := zap.NewNop()
	tempDir := "/tmp/test"

	handler := NewCopyHandler(nil, nil, tempDir, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, tempDir, handler.tempDir)
}

func TestCopyHandler_StorageOperations_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	handler := &CopyHandler{
		logger: logger,
	}

	// Test 1: Get storage roots
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest("GET", "/storage/roots", nil)
	handler.GetStorageRoots(c1)
	assert.Equal(t, http.StatusOK, w1.Code)

	var rootsResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &rootsResponse)
	assert.NoError(t, err)
	roots := rootsResponse["roots"].([]interface{})
	firstRoot := roots[0].(map[string]interface{})
	storageID := firstRoot["id"].(string)

	// Test 2: List storage path with the storage ID from roots
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/storage/list/test?storage_id="+storageID, nil)
	c2.Params = gin.Params{{Key: "path", Value: "test"}}
	c2.Request.URL.RawQuery = "storage_id=" + storageID
	handler.ListStoragePath(c2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var listResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &listResponse)
	assert.NoError(t, err)
	assert.Equal(t, storageID, listResponse["storage_id"])

	// Test 3: Copy to storage
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	copyBody := map[string]string{
		"source_path": "/tmp/test.txt",
		"dest_path":   "/storage/test.txt",
		"storage_id":  storageID,
	}
	jsonBody, _ := json.Marshal(copyBody)
	c3.Request = httptest.NewRequest("POST", "/copy/storage", bytes.NewBuffer(jsonBody))
	c3.Request.Header.Set("Content-Type", "application/json")
	handler.CopyToStorage(c3)
	assert.Equal(t, http.StatusOK, w3.Code)

	var copyResponse map[string]interface{}
	err = json.Unmarshal(w3.Body.Bytes(), &copyResponse)
	assert.NoError(t, err)
	assert.Equal(t, storageID, copyResponse["storage_id"])
}
