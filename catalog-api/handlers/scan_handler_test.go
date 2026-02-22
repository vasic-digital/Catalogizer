package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/database"
	"catalogizer/internal/services"
	"catalogizer/internal/tests"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// mockUniversalScanner is a mock implementation of UniversalScanner for testing
type mockUniversalScanner struct {
	queueScanCalled bool
	queueScanJob    services.ScanJob
	queueScanError  error

	activeStatuses map[string]*services.ScanStatus
	getAllCalled   bool
	getCalled      map[string]bool
}

func newMockUniversalScanner() *mockUniversalScanner {
	return &mockUniversalScanner{
		activeStatuses: make(map[string]*services.ScanStatus),
		getCalled:      make(map[string]bool),
	}
}

func (m *mockUniversalScanner) QueueScan(job services.ScanJob) error {
	m.queueScanCalled = true
	m.queueScanJob = job
	return m.queueScanError
}

func (m *mockUniversalScanner) GetAllActiveScanStatuses() map[string]*services.ScanStatus {
	m.getAllCalled = true
	return m.activeStatuses
}

func (m *mockUniversalScanner) GetActiveScanStatus(jobID string) (*services.ScanStatus, bool) {
	m.getCalled[jobID] = true
	status, exists := m.activeStatuses[jobID]
	return status, exists
}

// ScanHandlerTestSuite is the test suite for ScanHandler
type ScanHandlerTestSuite struct {
	suite.Suite
	handler *ScanHandler
	router  *gin.Engine
	db      *database.DB
	mock    *mockUniversalScanner
}

func (suite *ScanHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *ScanHandlerTestSuite) SetupTest() {
	// Create in-memory test database
	sqlDB := tests.SetupTestDB(suite.T())
	suite.db = &database.DB{DB: sqlDB}

	// Create mock scanner
	suite.mock = newMockUniversalScanner()

	// Initialize handler with mock scanner and test database
	suite.handler = NewScanHandler(suite.mock, suite.db)

	// Setup router
	suite.router = gin.New()
	suite.router.POST("/api/v1/storage/roots", suite.handler.CreateStorageRoot)
	suite.router.GET("/api/v1/storage/roots", suite.handler.GetStorageRoots)
	suite.router.POST("/api/v1/scans", suite.handler.QueueScan)
	suite.router.GET("/api/v1/scans", suite.handler.ListScans)
	suite.router.GET("/api/v1/scans/:job_id", suite.handler.GetScanStatus)
}

func (suite *ScanHandlerTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

// TestNewScanHandler tests the constructor
func (suite *ScanHandlerTestSuite) TestNewScanHandler() {
	handler := NewScanHandler(nil, nil)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.scanner)
	assert.Nil(suite.T(), handler.db)
}

// TestCreateStorageRoot_ValidRequest tests creating a new storage root
func (suite *ScanHandlerTestSuite) TestCreateStorageRoot_ValidRequest() {
	reqBody := map[string]interface{}{
		"name":      "Test SMB Share",
		"protocol":  "smb",
		"host":      "server.example.com",
		"port":      445,
		"path":      "/share",
		"username":  "user",
		"password":  "pass",
		"domain":    "WORKGROUP",
		"max_depth": 5,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/storage/roots", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test SMB Share", response["name"])
	assert.Equal(suite.T(), "smb", response["protocol"])
	assert.NotNil(suite.T(), response["id"])
}

// TestCreateStorageRoot_InvalidRequest tests missing required fields
func (suite *ScanHandlerTestSuite) TestCreateStorageRoot_InvalidRequest() {
	reqBody := map[string]interface{}{
		"protocol": "smb",
		// missing name
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/storage/roots", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestCreateStorageRoot_UpdateExisting tests updating an existing storage root
func (suite *ScanHandlerTestSuite) TestCreateStorageRoot_UpdateExisting() {
	// First create a storage root
	reqBody1 := map[string]interface{}{
		"name":      "Existing Share",
		"protocol":  "smb",
		"host":      "server.example.com",
		"max_depth": 10,
	}
	body1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/api/v1/storage/roots", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, req1)
	assert.Equal(suite.T(), http.StatusCreated, w1.Code)

	// Update with same name but different protocol
	reqBody2 := map[string]interface{}{
		"name":      "Existing Share",
		"protocol":  "ftp",
		"host":      "ftp.example.com",
		"max_depth": 15,
	}
	body2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/api/v1/storage/roots", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusCreated, w2.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ftp", response["protocol"])
}

// TestGetStorageRoots tests retrieving all storage roots
func (suite *ScanHandlerTestSuite) TestGetStorageRoots() {
	// Create a storage root first
	reqBody := map[string]interface{}{
		"name":      "Test Share",
		"protocol":  "smb",
		"max_depth": 10,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/storage/roots", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// Now get all storage roots
	req2 := httptest.NewRequest("GET", "/api/v1/storage/roots", nil)
	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	roots, ok := response["roots"].([]interface{})
	assert.True(suite.T(), ok)
	assert.GreaterOrEqual(suite.T(), len(roots), 1)
}

// TestQueueScan_ValidRequest tests queueing a scan job
func (suite *ScanHandlerTestSuite) TestQueueScan_ValidRequest() {
	// Create a storage root first
	reqBody := map[string]interface{}{
		"name":      "Scan Test Share",
		"protocol":  "smb",
		"max_depth": 10,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/storage/roots", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NoError(suite.T(), err)
	storageRootID := createResponse["id"].(float64)

	// Queue a scan for this storage root
	scanReqBody := map[string]interface{}{
		"storage_root_id": storageRootID,
		"scan_type":       "full",
		"max_depth":       5,
	}
	scanBody, _ := json.Marshal(scanReqBody)
	scanReq := httptest.NewRequest("POST", "/api/v1/scans", bytes.NewReader(scanBody))
	scanReq.Header.Set("Content-Type", "application/json")
	scanW := httptest.NewRecorder()
	suite.router.ServeHTTP(scanW, scanReq)

	assert.Equal(suite.T(), http.StatusAccepted, scanW.Code)
	var scanResponse map[string]interface{}
	err = json.Unmarshal(scanW.Body.Bytes(), &scanResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "queued", scanResponse["status"])
	assert.True(suite.T(), suite.mock.queueScanCalled)
}

// TestQueueScan_InvalidStorageRoot tests queueing a scan for non-existent storage root
func (suite *ScanHandlerTestSuite) TestQueueScan_InvalidStorageRoot() {
	scanReqBody := map[string]interface{}{
		"storage_root_id": 9999,
		"scan_type":       "full",
	}
	scanBody, _ := json.Marshal(scanReqBody)
	scanReq := httptest.NewRequest("POST", "/api/v1/scans", bytes.NewReader(scanBody))
	scanReq.Header.Set("Content-Type", "application/json")
	scanW := httptest.NewRecorder()
	suite.router.ServeHTTP(scanW, scanReq)

	assert.Equal(suite.T(), http.StatusNotFound, scanW.Code)
	assert.False(suite.T(), suite.mock.queueScanCalled)
}

// TestQueueScan_InvalidRequest tests missing required fields
func (suite *ScanHandlerTestSuite) TestQueueScan_InvalidRequest() {
	// Missing storage_root_id
	scanReqBody := map[string]interface{}{
		"scan_type": "full",
	}
	scanBody, _ := json.Marshal(scanReqBody)
	scanReq := httptest.NewRequest("POST", "/api/v1/scans", bytes.NewReader(scanBody))
	scanReq.Header.Set("Content-Type", "application/json")
	scanW := httptest.NewRecorder()
	suite.router.ServeHTTP(scanW, scanReq)

	assert.Equal(suite.T(), http.StatusBadRequest, scanW.Code)
}

// TestListScans tests retrieving active scan statuses
func (suite *ScanHandlerTestSuite) TestListScans() {
	// Setup mock to return some active scans
	suite.mock.activeStatuses["job1"] = &services.ScanStatus{
		StorageRootName: "Test Share",
		Protocol:        "smb",
		Status:          "scanning",
		FilesProcessed:  100,
		FilesFound:      200,
	}

	req := httptest.NewRequest("GET", "/api/v1/scans", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.True(suite.T(), suite.mock.getAllCalled)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	scans, ok := response["scans"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), 1, len(scans))
}

// TestGetScanStatus_Existing tests retrieving status of an existing scan
func (suite *ScanHandlerTestSuite) TestGetScanStatus_Existing() {
	// Setup mock to return a scan status
	suite.mock.activeStatuses["job123"] = &services.ScanStatus{
		StorageRootName: "Test Share",
		Protocol:        "smb",
		Status:          "scanning",
		FilesProcessed:  50,
		FilesFound:      100,
	}

	req := httptest.NewRequest("GET", "/api/v1/scans/job123", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.True(suite.T(), suite.mock.getCalled["job123"])

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "job123", response["job_id"])
	assert.Equal(suite.T(), "Test Share", response["storage_root"])
}

// TestGetScanStatus_NotFound tests retrieving status of a non-existent scan
func (suite *ScanHandlerTestSuite) TestGetScanStatus_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/scans/nonexistent", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	assert.True(suite.T(), suite.mock.getCalled["nonexistent"])
}

// Run the test suite
func TestScanHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ScanHandlerTestSuite))
}
