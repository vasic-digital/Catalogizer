# PHASE 1: TEST COVERAGE - CRITICAL SERVICES
## Implementation Guide - Week 3-6

---

## OBJECTIVE

Achieve 95%+ test coverage on all critical services, handlers, and repositories. This phase focuses on the highest priority components with the lowest current coverage.

**Success Criteria:**
- sync_service.go: 12.6% → 95%
- webdav_client.go: 2.0% → 95%
- favorites_service.go: 14.1% → 95%
- auth_service.go: 27.2% → 95%
- conversion_service.go: 21.3% → 95%
- All handlers: ~30% → 95%
- Repository layer: 53% → 95%

---

## WEEK 3: SYNC SERVICE (12.6% → 95%)

### Day 1-2: Sync Service Architecture Analysis

#### Task 1.1: Analyze Current Implementation

**File:** `catalog-api/services/sync_service.go` (500+ lines)

**Key Components to Test:**
1. Cloud provider synchronization (Google Drive, Dropbox, OneDrive)
2. Conflict resolution strategies
3. Retry and error handling
4. Progress tracking
5. Batch operations

**Test Strategy:**
```go
// Test categories:
1. Unit tests (70%) - Individual function testing
2. Integration tests (20%) - With mock cloud APIs
3. Error scenarios (10%) - Failure handling
```

#### Task 1.2: Create Mock Cloud Providers

```go
// File: catalog-api/services/sync_service_mocks_test.go
package services

import (
	"context"
	"errors"
	"io"
	"time"
)

// MockCloudProvider implements the CloudProvider interface for testing
type MockCloudProvider struct {
	name           string
	files          map[string]*MockCloudFile
	shouldFail     bool
	failureMode    string
	latency        time.Duration
	callCount      map[string]int
}

type MockCloudFile struct {
	Path         string
	Content      []byte
	ModTime      time.Time
	Size         int64
	IsDir        bool
	ContentHash  string
}

func NewMockCloudProvider(name string) *MockCloudProvider {
	return &MockCloudProvider{
		name:      name,
		files:     make(map[string]*MockCloudFile),
		callCount: make(map[string]int),
	}
}

func (m *MockCloudProvider) SetFailure(mode string) {
	m.shouldFail = true
	m.failureMode = mode
}

func (m *MockCloudProvider) AddFile(path string, content []byte, modTime time.Time) {
	m.files[path] = &MockCloudFile{
		Path:    path,
		Content: content,
		ModTime: modTime,
		Size:    int64(len(content)),
	}
}

func (m *MockCloudProvider) ListFiles(ctx context.Context, prefix string) ([]CloudFile, error) {
	m.callCount["ListFiles"]++
	
	if m.shouldFail && m.failureMode == "list" {
		return nil, errors.New("mock list failure")
	}
	
	time.Sleep(m.latency)
	
	var files []CloudFile
	for path, file := range m.files {
		if prefix == "" || strings.HasPrefix(path, prefix) {
			files = append(files, CloudFile{
				Path:    path,
				Size:    file.Size,
				ModTime: file.ModTime,
				IsDir:   file.IsDir,
			})
		}
	}
	
	return files, nil
}

func (m *MockCloudProvider) DownloadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	m.callCount["DownloadFile"]++
	
	if m.shouldFail && m.failureMode == "download" {
		return nil, errors.New("mock download failure")
	}
	
	file, exists := m.files[path]
	if !exists {
		return nil, errors.New("file not found")
	}
	
	return io.NopCloser(bytes.NewReader(file.Content)), nil
}

func (m *MockCloudProvider) UploadFile(ctx context.Context, path string, content io.Reader) error {
	m.callCount["UploadFile"]++
	
	if m.shouldFail && m.failureMode == "upload" {
		return errors.New("mock upload failure")
	}
	
	data, _ := io.ReadAll(content)
	m.files[path] = &MockCloudFile{
		Path:    path,
		Content: data,
		Size:    int64(len(data)),
		ModTime: time.Now(),
	}
	
	return nil
}

func (m *MockCloudProvider) DeleteFile(ctx context.Context, path string) error {
	m.callCount["DeleteFile"]++
	
	if m.shouldFail && m.failureMode == "delete" {
		return errors.New("mock delete failure")
	}
	
	delete(m.files, path)
	return nil
}

func (m *MockCloudProvider) GetFileMetadata(ctx context.Context, path string) (*CloudFile, error) {
	m.callCount["GetFileMetadata"]++
	
	file, exists := m.files[path]
	if !exists {
		return nil, errors.New("file not found")
	}
	
	return &CloudFile{
		Path:    file.Path,
		Size:    file.Size,
		ModTime: file.ModTime,
		IsDir:   file.IsDir,
	}, nil
}

func (m *MockCloudProvider) GetCallCount(method string) int {
	return m.callCount[method]
}

func (m *MockCloudProvider) ResetCallCount() {
	m.callCount = make(map[string]int)
}
```

### Day 3-4: Unit Tests for Sync Operations

```go
// File: catalog-api/services/sync_service_test.go
package services

import (
	"bytes"
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSyncService_NewSyncService(t *testing.T) {
	tests := []struct {
		name        string
		config      *SyncConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &SyncConfig{
				Provider:     "google_drive",
				LocalPath:    "/tmp/sync",
				RemotePath:   "/remote",
				SyncInterval: time.Hour,
			},
			expectError: false,
		},
		{
			name: "missing provider",
			config: &SyncConfig{
				LocalPath:  "/tmp/sync",
				RemotePath: "/remote",
			},
			expectError: true,
		},
		{
			name: "missing local path",
			config: &SyncConfig{
				Provider:   "google_drive",
				RemotePath: "/remote",
			},
			expectError: true,
		},
		{
			name: "invalid sync interval",
			config: &SyncConfig{
				Provider:     "google_drive",
				LocalPath:    "/tmp/sync",
				RemotePath:   "/remote",
				SyncInterval: 0,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewSyncService(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestSyncService_StartStop(t *testing.T) {
	config := &SyncConfig{
		Provider:     "mock",
		LocalPath:    t.TempDir(),
		RemotePath:   "/remote",
		SyncInterval: time.Hour,
	}
	
	service, err := NewSyncService(config)
	assert.NoError(t, err)
	
	// Test start
	ctx := context.Background()
	err = service.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, service.IsRunning())
	
	// Test stop
	err = service.Stop()
	assert.NoError(t, err)
	assert.False(t, service.IsRunning())
}

func TestSyncService_SyncOnce(t *testing.T) {
	mockProvider := NewMockCloudProvider("mock")
	
	// Add files to remote
	mockProvider.AddFile("/remote/file1.txt", []byte("content1"), time.Now())
	mockProvider.AddFile("/remote/file2.txt", []byte("content2"), time.Now().Add(-time.Hour))
	
	config := &SyncConfig{
		Provider:   "mock",
		LocalPath:  t.TempDir(),
		RemotePath: "/remote",
	}
	
	service, err := NewSyncService(config)
	assert.NoError(t, err)
	
	// Inject mock provider
	service.provider = mockProvider
	
	// Perform sync
	ctx := context.Background()
	result, err := service.SyncOnce(ctx)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.FilesDownloaded)
	assert.Equal(t, 0, result.FilesUploaded)
	assert.Equal(t, 0, result.Conflicts)
}

func TestSyncService_ConflictResolution(t *testing.T) {
	tests := []struct {
		name           string
		localModTime   time.Time
		remoteModTime  time.Time
		strategy       ConflictStrategy
		expectWinner   string // "local" or "remote"
	}{
		{
			name:          "remote newer - take remote",
			localModTime:  time.Now().Add(-time.Hour),
			remoteModTime: time.Now(),
			strategy:      ConflictTakeNewer,
			expectWinner:  "remote",
		},
		{
			name:          "local newer - take local",
			localModTime:  time.Now(),
			remoteModTime: time.Now().Add(-time.Hour),
			strategy:      ConflictTakeNewer,
			expectWinner:  "local",
		},
		{
			name:          "always take local",
			localModTime:  time.Now().Add(-time.Hour),
			remoteModTime: time.Now(),
			strategy:      ConflictTakeLocal,
			expectWinner:  "local",
		},
		{
			name:          "always take remote",
			localModTime:  time.Now(),
			remoteModTime: time.Now().Add(-time.Hour),
			strategy:      ConflictTakeRemote,
			expectWinner:  "remote",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := NewMockCloudProvider("mock")
			mockProvider.AddFile("/remote/file.txt", []byte("remote content"), tt.remoteModTime)
			
			localPath := t.TempDir()
			localFile := filepath.Join(localPath, "file.txt")
			os.WriteFile(localFile, []byte("local content"), 0644)
			os.Chtimes(localFile, tt.localModTime, tt.localModTime)
			
			config := &SyncConfig{
				Provider:          "mock",
				LocalPath:         localPath,
				RemotePath:        "/remote",
				ConflictStrategy:  tt.strategy,
			}
			
			service, _ := NewSyncService(config)
			service.provider = mockProvider
			
			ctx := context.Background()
			result, err := service.SyncOnce(ctx)
			
			assert.NoError(t, err)
			
			// Verify which version won
			content, _ := os.ReadFile(localFile)
			if tt.expectWinner == "local" {
				assert.Equal(t, "local content", string(content))
			} else {
				assert.Equal(t, "remote content", string(content))
			}
		})
	}
}

func TestSyncService_RetryLogic(t *testing.T) {
	mockProvider := NewMockCloudProvider("mock")
	
	// First call fails, second succeeds
	callCount := 0
	mockProvider.ListFilesFunc = func(ctx context.Context, prefix string) ([]CloudFile, error) {
		callCount++
		if callCount == 1 {
			return nil, errors.New("temporary error")
		}
		return []CloudFile{}, nil
	}
	
	config := &SyncConfig{
		Provider:     "mock",
		LocalPath:    t.TempDir(),
		RemotePath:   "/remote",
		MaxRetries:   3,
		RetryDelay:   100 * time.Millisecond,
	}
	
	service, _ := NewSyncService(config)
	service.provider = mockProvider
	
	ctx := context.Background()
	result, err := service.SyncOnce(ctx)
	
	// Should succeed after retry
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, callCount) // Initial + 1 retry
}

func TestSyncService_ProgressTracking(t *testing.T) {
	mockProvider := NewMockCloudProvider("mock")
	
	// Add multiple files
	for i := 0; i < 10; i++ {
		mockProvider.AddFile(
			fmt.Sprintf("/remote/file%d.txt", i),
			[]byte(fmt.Sprintf("content%d", i)),
			time.Now(),
		)
	}
	
	config := &SyncConfig{
		Provider:   "mock",
		LocalPath:  t.TempDir(),
		RemotePath: "/remote",
	}
	
	service, _ := NewSyncService(config)
	service.provider = mockProvider
	
	// Track progress
	progressUpdates := make([]SyncProgress, 0)
	service.OnProgress = func(p SyncProgress) {
		progressUpdates = append(progressUpdates, p)
	}
	
	ctx := context.Background()
	service.SyncOnce(ctx)
	
	// Verify progress tracking
	assert.Greater(t, len(progressUpdates), 0)
	assert.Equal(t, 10, progressUpdates[len(progressUpdates)-1].TotalFiles)
	assert.Equal(t, 10, progressUpdates[len(progressUpdates)-1].ProcessedFiles)
}

func TestSyncService_Cancellation(t *testing.T) {
	mockProvider := NewMockCloudProvider("mock")
	mockProvider.latency = 100 * time.Millisecond
	
	// Add many files to ensure cancellation happens mid-sync
	for i := 0; i < 100; i++ {
		mockProvider.AddFile(
			fmt.Sprintf("/remote/file%d.txt", i),
			[]byte("content"),
			time.Now(),
		)
	}
	
	config := &SyncConfig{
		Provider:   "mock",
		LocalPath:  t.TempDir(),
		RemotePath: "/remote",
	}
	
	service, _ := NewSyncService(config)
	service.provider = mockProvider
	
	// Create cancellable context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	_, err := service.SyncOnce(ctx)
	
	// Should get cancellation error
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestSyncService_BatchOperations(t *testing.T) {
	mockProvider := NewMockCloudProvider("mock")
	
	// Add files to remote
	for i := 0; i < 50; i++ {
		mockProvider.AddFile(
			fmt.Sprintf("/remote/file%d.txt", i),
			[]byte(fmt.Sprintf("content%d", i)),
			time.Now(),
		)
	}
	
	config := &SyncConfig{
		Provider:    "mock",
		LocalPath:   t.TempDir(),
		RemotePath:  "/remote",
		BatchSize:   10,
	}
	
	service, _ := NewSyncService(config)
	service.provider = mockProvider
	
	ctx := context.Background()
	result, err := service.SyncOnce(ctx)
	
	assert.NoError(t, err)
	assert.Equal(t, 50, result.FilesDownloaded)
	
	// Verify batch processing (should make multiple calls)
	// In a real implementation, we'd track batch operations
}

func TestSyncService_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		failureMode string
		expectError bool
	}{
		{
			name:        "list files failure",
			failureMode: "list",
			expectError: true,
		},
		{
			name:        "download failure",
			failureMode: "download",
			expectError: true,
		},
		{
			name:        "upload failure",
			failureMode: "upload",
			expectError: false, // Should continue with other files
		},
		{
			name:        "delete failure",
			failureMode: "delete",
			expectError: false, // Should continue with other files
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := NewMockCloudProvider("mock")
			mockProvider.SetFailure(tt.failureMode)
			mockProvider.AddFile("/remote/file.txt", []byte("content"), time.Now())
			
			config := &SyncConfig{
				Provider:   "mock",
				LocalPath:  t.TempDir(),
				RemotePath: "/remote",
			}
			
			service, _ := NewSyncService(config)
			service.provider = mockProvider
			
			ctx := context.Background()
			_, err := service.SyncOnce(ctx)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncService_ConcurrentSyncs(t *testing.T) {
	mockProvider := NewMockCloudProvider("mock")
	mockProvider.AddFile("/remote/file.txt", []byte("content"), time.Now())
	
	config := &SyncConfig{
		Provider:   "mock",
		LocalPath:  t.TempDir(),
		RemotePath: "/remote",
	}
	
	service, _ := NewSyncService(config)
	service.provider = mockProvider
	
	// Try to start multiple concurrent syncs
	var wg sync.WaitGroup
	errors := make([]error, 3)
	
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx := context.Background()
			_, errors[idx] = service.SyncOnce(ctx)
		}(i)
	}
	
	wg.Wait()
	
	// Only one should succeed, others should get "sync in progress" error
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		} else {
			assert.Contains(t, err.Error(), "sync already in progress")
		}
	}
	
	assert.Equal(t, 1, successCount)
}
```

### Day 5-6: Integration Tests

```go
// File: catalog-api/tests/integration/sync_integration_test.go
package integration

import (
	"context"
	"testing"
	"time"
	
	"catalog-api/services"
	"catalog-api/internal/tests"
)

func TestSyncService_Integration_FullSync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Setup test database
	db := tests.SetupTestDB(t)
	defer tests.TeardownTestDB(t, db)
	
	// Create test fixtures
	fixtures := tests.NewTestFixtures(db)
	user, _ := fixtures.CreateTestUser("syncuser", "password")
	
	// Setup mock provider
	mockProvider := services.NewMockCloudProvider("test")
	mockProvider.AddFile("/test/file1.txt", []byte("content1"), time.Now())
	mockProvider.AddFile("/test/file2.txt", []byte("content2"), time.Now())
	
	config := &services.SyncConfig{
		Provider:     "test",
		UserID:       user.ID,
		LocalPath:    t.TempDir(),
		RemotePath:   "/test",
		SyncInterval: time.Hour,
	}
	
	service, err := services.NewSyncService(config)
	assert.NoError(t, err)
	
	// Inject mock provider
	service.SetProvider(mockProvider)
	
	// Perform full sync
	ctx := context.Background()
	result, err := service.SyncOnce(ctx)
	
	assert.NoError(t, err)
	assert.Equal(t, 2, result.FilesDownloaded)
	
	// Verify files in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM synced_files WHERE user_id = ?", user.ID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSyncService_Integration_DeltaSync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	db := tests.SetupTestDB(t)
	defer tests.TeardownTestDB(t, db)
	
	fixtures := tests.NewTestFixtures(db)
	user, _ := fixtures.CreateTestUser("syncuser2", "password")
	localPath := t.TempDir()
	
	// First sync - full sync
	mockProvider := services.NewMockCloudProvider("test")
	mockProvider.AddFile("/test/file1.txt", []byte("content1"), time.Now())
	
	config := &services.SyncConfig{
		Provider:   "test",
		UserID:     user.ID,
		LocalPath:  localPath,
		RemotePath: "/test",
	}
	
	service, _ := services.NewSyncService(config)
	service.SetProvider(mockProvider)
	
	ctx := context.Background()
	result1, _ := service.SyncOnce(ctx)
	assert.Equal(t, 1, result1.FilesDownloaded)
	
	// Add new file and modify existing
	mockProvider.AddFile("/test/file2.txt", []byte("new content"), time.Now())
	mockProvider.AddFile("/test/file1.txt", []byte("updated content"), time.Now())
	
	// Second sync - delta sync
	result2, _ := service.SyncOnce(ctx)
	
	assert.Equal(t, 1, result2.FilesDownloaded)  // New file
	assert.Equal(t, 1, result2.FilesUpdated)     // Modified file
}
```

---

## WEEK 4: WEBDAV CLIENT & FAVORITES SERVICE

### Day 1-3: WebDAV Client Tests (2.0% → 95%)

```go
// File: catalog-api/services/webdav_client_test.go
package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
)

func TestWebDAVClient_NewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      WebDAVConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: WebDAVConfig{
				URL:      "http://webdav.example.com",
				Username: "user",
				Password: "pass",
			},
			expectError: false,
		},
		{
			name: "missing URL",
			config: WebDAVConfig{
				Username: "user",
				Password: "pass",
			},
			expectError: true,
		},
		{
			name: "invalid URL",
			config: WebDAVConfig{
				URL:      "not-a-valid-url",
				Username: "user",
				Password: "pass",
			},
			expectError: true,
		},
		{
			name: "with timeout",
			config: WebDAVConfig{
				URL:      "http://webdav.example.com",
				Username: "user",
				Password: "pass",
				Timeout:  30 * time.Second,
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewWebDAVClient(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestWebDAVClient_ListFiles(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PROPFIND", r.Method)
		assert.Equal(t, "/remote/", r.URL.Path)
		
		// Return WebDAV response
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(207)
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<d:multistatus xmlns:d="DAV:">
  <d:response>
    <d:href>/remote/</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype>
          <d:collection/>
        </d:resourcetype>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/remote/file1.txt</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype/>
        <d:getcontentlength>1024</d:getcontentlength>
        <d:getlastmodified>Mon, 01 Jan 2024 00:00:00 GMT</d:getlastmodified>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`))
	}))
	defer server.Close()
	
	client, _ := NewWebDAVClient(WebDAVConfig{
		URL:      server.URL,
		Username: "user",
		Password: "pass",
	})
	
	ctx := context.Background()
	files, err := client.ListFiles(ctx, "/remote/")
	
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "/remote/file1.txt", files[0].Path)
	assert.Equal(t, int64(1024), files[0].Size)
}

func TestWebDAVClient_DownloadFile(t *testing.T) {
	expectedContent := []byte("test file content")
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/remote/file.txt", r.URL.Path)
		
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write(expectedContent)
	}))
	defer server.Close()
	
	client, _ := NewWebDAVClient(WebDAVConfig{
		URL:      server.URL,
		Username: "user",
		Password: "pass",
	})
	
	ctx := context.Background()
	reader, err := client.DownloadFile(ctx, "/remote/file.txt")
	assert.NoError(t, err)
	
	content, _ := io.ReadAll(reader)
	reader.Close()
	
	assert.Equal(t, expectedContent, content)
}

func TestWebDAVClient_UploadFile(t *testing.T) {
	uploaded := false
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/remote/file.txt", r.URL.Path)
		
		content, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test content", string(content))
		
		uploaded = true
		w.WriteHeader(201)
	}))
	defer server.Close()
	
	client, _ := NewWebDAVClient(WebDAVConfig{
		URL:      server.URL,
		Username: "user",
		Password: "pass",
	})
	
	ctx := context.Background()
	err := client.UploadFile(ctx, "/remote/file.txt", bytes.NewReader([]byte("test content")))
	
	assert.NoError(t, err)
	assert.True(t, uploaded)
}

func TestWebDAVClient_DeleteFile(t *testing.T) {
	deleted := false
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/remote/file.txt", r.URL.Path)
		
		deleted = true
		w.WriteHeader(204)
	}))
	defer server.Close()
	
	client, _ := NewWebDAVClient(WebDAVConfig{
		URL:      server.URL,
		Username: "user",
		Password: "pass",
	})
	
	ctx := context.Background()
	err := client.DeleteFile(ctx, "/remote/file.txt")
	
	assert.NoError(t, err)
	assert.True(t, deleted)
}

func TestWebDAVClient_CreateDirectory(t *testing.T) {
	created := false
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "MKCOL", r.Method)
		assert.Equal(t, "/remote/newdir", r.URL.Path)
		
		created = true
		w.WriteHeader(201)
	}))
	defer server.Close()
	
	client, _ := NewWebDAVClient(WebDAVConfig{
		URL:      server.URL,
		Username: "user",
		Password: "pass",
	})
	
	ctx := context.Background()
	err := client.CreateDirectory(ctx, "/remote/newdir")
	
	assert.NoError(t, err)
	assert.True(t, created)
}

func TestWebDAVClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedError  string
	}{
		{
			name:          "not found",
			statusCode:    404,
			expectedError: "not found",
		},
		{
			name:          "unauthorized",
			statusCode:    401,
			expectedError: "unauthorized",
		},
		{
			name:          "forbidden",
			statusCode:    403,
			expectedError: "forbidden",
		},
		{
			name:          "server error",
			statusCode:    500,
			expectedError: "server error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()
			
			client, _ := NewWebDAVClient(WebDAVConfig{
				URL:      server.URL,
				Username: "user",
				Password: "pass",
			})
			
			ctx := context.Background()
			_, err := client.ListFiles(ctx, "/")
			
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestWebDAVClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(200)
	}))
	defer server.Close()
	
	client, _ := NewWebDAVClient(WebDAVConfig{
		URL:      server.URL,
		Username: "user",
		Password: "pass",
		Timeout:  100 * time.Millisecond,
	})
	
	ctx := context.Background()
	_, err := client.ListFiles(ctx, "/")
	
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestWebDAVClient_RetryLogic(t *testing.T) {
	attemptCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(503) // Service unavailable
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`<?xml version="1.0"?>
<d:multistatus xmlns:d="DAV:">
  <d:response><d:href>/</d:href></d:response>
</d:multistatus>`))
	}))
	defer server.Close()
	
	client, _ := NewWebDAVClient(WebDAVConfig{
		URL:       server.URL,
		Username:  "user",
		Password:  "pass",
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	})
	
	ctx := context.Background()
	_, err := client.ListFiles(ctx, "/")
	
	assert.NoError(t, err)
	assert.Equal(t, 3, attemptCount)
}
```

### Day 4-6: Favorites Service Tests

```go
// File: catalog-api/services/favorites_service_test.go
package services

import (
	"context"
	"testing"
	"time"
	
	"catalog-api/internal/tests"
	"catalog-api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FavoritesServiceTestSuite struct {
	suite.Suite
	db      *tests.TestDB
	service *FavoritesService
	user    *models.User
}

func (s *FavoritesServiceTestSuite) SetupTest() {
	s.db = tests.SetupTestDB(s.T())
	s.service = NewFavoritesService(s.db.DB)
	
	fixtures := tests.NewTestFixtures(s.db.DB)
	var err error
	s.user, err = fixtures.CreateTestUser("testuser", "password")
	s.Require().NoError(err)
}

func (s *FavoritesServiceTestSuite) TearDownTest() {
	tests.TeardownTestDB(s.T(), s.db)
}

func TestFavoritesServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FavoritesServiceTestSuite))
}

func (s *FavoritesServiceTestSuite) TestAddFavorite() {
	// Create a media item
	fixtures := tests.NewTestFixtures(s.db.DB)
	item, err := fixtures.CreateTestMediaItem("Test Movie", "movie")
	s.Require().NoError(err)
	
	// Add to favorites
	ctx := context.Background()
	favorite, err := s.service.AddFavorite(ctx, s.user.ID, item.ID, "movie")
	
	s.NoError(err)
	s.NotNil(favorite)
	s.Equal(s.user.ID, favorite.UserID)
	s.Equal(item.ID, favorite.ItemID)
	s.Equal("movie", favorite.ItemType)
}

func (s *FavoritesServiceTestSuite) TestAddFavorite_Duplicate() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	item, _ := fixtures.CreateTestMediaItem("Test Movie", "movie")
	
	ctx := context.Background()
	
	// Add first time
	_, err := s.service.AddFavorite(ctx, s.user.ID, item.ID, "movie")
	s.NoError(err)
	
	// Add second time - should return error or existing favorite
	_, err = s.service.AddFavorite(ctx, s.user.ID, item.ID, "movie")
	s.Error(err)
	s.Contains(err.Error(), "already exists")
}

func (s *FavoritesServiceTestSuite) TestRemoveFavorite() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	item, _ := fixtures.CreateTestMediaItem("Test Movie", "movie")
	
	ctx := context.Background()
	
	// Add favorite
	favorite, _ := s.service.AddFavorite(ctx, s.user.ID, item.ID, "movie")
	
	// Remove favorite
	err := s.service.RemoveFavorite(ctx, favorite.ID)
	s.NoError(err)
	
	// Verify removed
	favorites, _ := s.service.GetFavorites(ctx, s.user.ID, FavoritesFilter{})
	s.Len(favorites, 0)
}

func (s *FavoritesServiceTestSuite) TestGetFavorites() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	
	// Create multiple media items
	item1, _ := fixtures.CreateTestMediaItem("Movie 1", "movie")
	item2, _ := fixtures.CreateTestMediaItem("Movie 2", "movie")
	item3, _ := fixtures.CreateTestMediaItem("TV Show", "tv_show")
	
	ctx := context.Background()
	
	// Add all to favorites
	s.service.AddFavorite(ctx, s.user.ID, item1.ID, "movie")
	s.service.AddFavorite(ctx, s.user.ID, item2.ID, "movie")
	s.service.AddFavorite(ctx, s.user.ID, item3.ID, "tv_show")
	
	// Get all favorites
	favorites, err := s.service.GetFavorites(ctx, s.user.ID, FavoritesFilter{})
	s.NoError(err)
	s.Len(favorites, 3)
}

func (s *FavoritesServiceTestSuite) TestGetFavorites_FilterByType() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	
	item1, _ := fixtures.CreateTestMediaItem("Movie 1", "movie")
	item2, _ := fixtures.CreateTestMediaItem("Movie 2", "movie")
	item3, _ := fixtures.CreateTestMediaItem("TV Show", "tv_show")
	
	ctx := context.Background()
	
	s.service.AddFavorite(ctx, s.user.ID, item1.ID, "movie")
	s.service.AddFavorite(ctx, s.user.ID, item2.ID, "movie")
	s.service.AddFavorite(ctx, s.user.ID, item3.ID, "tv_show")
	
	// Filter by type
	favorites, err := s.service.GetFavorites(ctx, s.user.ID, FavoritesFilter{
		ItemType: "movie",
	})
	
	s.NoError(err)
	s.Len(favorites, 2)
	
	for _, f := range favorites {
		s.Equal("movie", f.ItemType)
	}
}

func (s *FavoritesServiceTestSuite) TestGetFavorites_Pagination() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	
	// Create 25 favorites
	ctx := context.Background()
	for i := 0; i < 25; i++ {
		item, _ := fixtures.CreateTestMediaItem(fmt.Sprintf("Item %d", i), "movie")
		s.service.AddFavorite(ctx, s.user.ID, item.ID, "movie")
	}
	
	// Test page 1
	favorites1, total, err := s.service.GetFavoritesWithPagination(ctx, s.user.ID, FavoritesFilter{}, 1, 10)
	s.NoError(err)
	s.Len(favorites1, 10)
	s.Equal(25, total)
	
	// Test page 2
	favorites2, _, _ := s.service.GetFavoritesWithPagination(ctx, s.user.ID, FavoritesFilter{}, 2, 10)
	s.Len(favorites2, 10)
	
	// Test page 3
	favorites3, _, _ := s.service.GetFavoritesWithPagination(ctx, s.user.ID, FavoritesFilter{}, 3, 10)
	s.Len(favorites3, 5)
}

func (s *FavoritesServiceTestSuite) TestGetFavorites_Sorting() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	
	// Add favorites at different times
	ctx := context.Background()
	
	item1, _ := fixtures.CreateTestMediaItem("Item A", "movie")
	time.Sleep(10 * time.Millisecond)
	item2, _ := fixtures.CreateTestMediaItem("Item B", "movie")
	time.Sleep(10 * time.Millisecond)
	item3, _ := fixtures.CreateTestMediaItem("Item C", "movie")
	
	s.service.AddFavorite(ctx, s.user.ID, item1.ID, "movie")
	s.service.AddFavorite(ctx, s.user.ID, item2.ID, "movie")
	s.service.AddFavorite(ctx, s.user.ID, item3.ID, "movie")
	
	// Sort by created_at DESC
	favorites, err := s.service.GetFavorites(ctx, s.user.ID, FavoritesFilter{
		SortBy:    "created_at",
		SortOrder: "desc",
	})
	
	s.NoError(err)
	s.Len(favorites, 3)
	s.Equal(item3.ID, favorites[0].ItemID) // Most recent first
	s.Equal(item1.ID, favorites[2].ItemID) // Oldest last
}

func (s *FavoritesServiceTestSuite) TestIsFavorite() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	item, _ := fixtures.CreateTestMediaItem("Test Movie", "movie")
	
	ctx := context.Background()
	
	// Check before adding
	isFav, err := s.service.IsFavorite(ctx, s.user.ID, item.ID)
	s.NoError(err)
	s.False(isFav)
	
	// Add favorite
	s.service.AddFavorite(ctx, s.user.ID, item.ID, "movie")
	
	// Check after adding
	isFav, err = s.service.IsFavorite(ctx, s.user.ID, item.ID)
	s.NoError(err)
	s.True(isFav)
}

func (s *FavoritesServiceTestSuite) TestGetFavoriteStats() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	
	// Add favorites of different types
	ctx := context.Background()
	
	for i := 0; i < 5; i++ {
		item, _ := fixtures.CreateTestMediaItem(fmt.Sprintf("Movie %d", i), "movie")
		s.service.AddFavorite(ctx, s.user.ID, item.ID, "movie")
	}
	
	for i := 0; i < 3; i++ {
		item, _ := fixtures.CreateTestMediaItem(fmt.Sprintf("TV Show %d", i), "tv_show")
		s.service.AddFavorite(ctx, s.user.ID, item.ID, "tv_show")
	}
	
	stats, err := s.service.GetFavoriteStats(ctx, s.user.ID)
	s.NoError(err)
	s.Equal(8, stats.TotalFavorites)
	s.Equal(5, stats.ByType["movie"])
	s.Equal(3, stats.ByType["tv_show"])
}

func (s *FavoritesServiceTestSuite) TestUserIsolation() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	
	// Create second user
	user2, _ := fixtures.CreateTestUser("user2", "password")
	
	// Create items
	item1, _ := fixtures.CreateTestMediaItem("Item 1", "movie")
	item2, _ := fixtures.CreateTestMediaItem("Item 2", "movie")
	
	ctx := context.Background()
	
	// User 1 favorites item 1
	s.service.AddFavorite(ctx, s.user.ID, item1.ID, "movie")
	
	// User 2 favorites item 2
	s.service.AddFavorite(ctx, user2.ID, item2.ID, "movie")
	
	// User 1 should only see their favorite
	favorites1, _ := s.service.GetFavorites(ctx, s.user.ID, FavoritesFilter{})
	s.Len(favorites1, 1)
	s.Equal(item1.ID, favorites1[0].ItemID)
	
	// User 2 should only see their favorite
	favorites2, _ := s.service.GetFavorites(ctx, user2.ID, FavoritesFilter{})
	s.Len(favorites2, 1)
	s.Equal(item2.ID, favorites2[0].ItemID)
}

func (s *FavoritesServiceTestSuite) TestConcurrentAccess() {
	fixtures := tests.NewTestFixtures(s.db.DB)
	
	// Create multiple items
	items := make([]*models.MediaItem, 10)
	for i := 0; i < 10; i++ {
		items[i], _ = fixtures.CreateTestMediaItem(fmt.Sprintf("Item %d", i), "movie")
	}
	
	// Concurrently add all to favorites
	ctx := context.Background()
	var wg sync.WaitGroup
	errors := make([]error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errors[idx] = s.service.AddFavorite(ctx, s.user.ID, items[idx].ID, "movie")
		}(i)
	}
	
	wg.Wait()
	
	// Check all succeeded
	for i, err := range errors {
		s.NoError(err, "Failed at index %d", i)
	}
	
	// Verify all favorites exist
	favorites, _ := s.service.GetFavorites(ctx, s.user.ID, FavoritesFilter{})
	s.Len(favorites, 10)
}
```

---

## WEEKS 5-6: AUTH SERVICE, CONVERSION SERVICE & HANDLERS

### Week 5: Auth Service (27.2% → 95%)

**Priority Tests:**
1. JWT token generation/validation
2. Password hashing (bcrypt)
3. Session management
4. RBAC permission checks
5. Rate limiting integration
6. Multi-factor authentication
7. Token refresh
8. Logout/revocation

### Week 6: Handler Tests (~30% → 95%)

**All handlers need comprehensive tests:**
- auth_handler.go
- media_handler.go
- browse_handler.go
- copy_handler.go
- download_handler.go
- entity_handler.go
- recommendation_handler.go
- search_handler.go

Each handler test file should include:
1. Happy path tests
2. Error handling tests
3. Authentication/authorization tests
4. Request validation tests
5. Response format tests

---

## COVERAGE VALIDATION

Run these commands to validate coverage:

```bash
# Run all tests with coverage
cd catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -coverprofile=coverage.out

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out | grep total

# Check specific packages
go test -cover ./services/...
go test -cover ./handlers/...
go test -cover ./repository/...

# Validate coverage threshold
./scripts/validate-coverage.sh 95
```

**Target Coverage by End of Phase 1:**
- Services: 95%
- Handlers: 95%
- Repository: 95%
- Overall: 95%

---

**Phase 1 Complete: All critical services at 95%+ coverage**
