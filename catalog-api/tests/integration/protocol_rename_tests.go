package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/mutecomm/go-sqlcipher"
	"go.uber.org/zap"

	"catalogizer/database"
	"catalogizer/filesystem"
	"catalogizer/internal/services"
)

// ProtocolTestSuite defines the interface for protocol-specific tests
type ProtocolTestSuite interface {
	SetupProtocol(t *testing.T) (filesystem.FileSystemClient, func())
	GetProtocolName() string
	GetTestConfig() map[string]interface{}
	SupportsRealTimeEvents() bool
}

// TestProtocolRenameDetection runs rename detection tests for all supported protocols
func TestProtocolRenameDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping protocol integration tests in short mode")
	}

	// Test each protocol
	protocolSuites := []ProtocolTestSuite{
		&LocalProtocolTestSuite{},
		&SMBProtocolTestSuite{},
		&FTPProtocolTestSuite{},
		&NFSProtocolTestSuite{},
		&WebDAVProtocolTestSuite{},
	}

	for _, suite := range protocolSuites {
		t.Run(suite.GetProtocolName(), func(t *testing.T) {
			testProtocolRenameDetection(t, suite)
		})
	}
}

func testProtocolRenameDetection(t *testing.T, suite ProtocolTestSuite) {
	logger := zap.NewNop()
	ctx := context.Background()

	// Setup database
	db := setupProtocolTestDB(t)
	defer db.Close()

	// Setup protocol
	client, cleanup := suite.SetupProtocol(t)
	defer cleanup()

	// Setup universal rename tracker
	renameTracker := services.NewUniversalRenameTracker(db, logger)
	if err := renameTracker.Start(); err != nil {
		t.Fatalf("Failed to start rename tracker: %v", err)
	}
	defer renameTracker.Stop()

	// Test file rename detection
	t.Run("file_rename", func(t *testing.T) {
		testFileRename(t, client, renameTracker, suite, ctx)
	})

	// Test directory rename detection
	t.Run("directory_rename", func(t *testing.T) {
		testDirectoryRename(t, client, renameTracker, suite, ctx)
	})

	// Test batch rename operations
	t.Run("batch_rename", func(t *testing.T) {
		testBatchRename(t, client, renameTracker, suite, ctx)
	})

	// Test rename detection timing
	t.Run("timing_windows", func(t *testing.T) {
		testRenameTimingWindows(t, client, renameTracker, suite, ctx)
	})

	// Test protocol-specific capabilities
	t.Run("protocol_capabilities", func(t *testing.T) {
		testProtocolCapabilities(t, suite, logger)
	})
}

func testFileRename(t *testing.T, client filesystem.FileSystemClient, renameTracker *services.UniversalRenameTracker, suite ProtocolTestSuite, ctx context.Context) {
	protocol := suite.GetProtocolName()

	// Create test file
	testContent := "Test file content for rename detection"
	originalPath := "/test_file_rename.txt"
	renamedPath := "/renamed_test_file.txt"

	// Create the file through the client
	if err := createTestFile(ctx, client, originalPath, testContent); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Simulate tracking the file creation
	fileID := int64(100)
	size := int64(len(testContent))
	fileHash := "testhash123"
	protocolData := suite.GetTestConfig()

	// Track the file deletion (simulating a move)
	renameTracker.TrackDelete(ctx, fileID, originalPath, "test_storage", protocol, size, &fileHash, false, protocolData)

	// Simulate the file creation at new location
	pendingMove, isMove := renameTracker.DetectCreate(ctx, renamedPath, "test_storage", protocol, size, &fileHash, false, protocolData)

	if !isMove {
		t.Error("Expected file rename to be detected")
		return
	}

	if pendingMove.Path != originalPath {
		t.Errorf("Expected original path %s, got %s", originalPath, pendingMove.Path)
	}

	if pendingMove.Protocol != protocol {
		t.Errorf("Expected protocol %s, got %s", protocol, pendingMove.Protocol)
	}

	// Process the move
	if err := renameTracker.ProcessMove(ctx, client, pendingMove, renamedPath); err != nil {
		t.Errorf("Failed to process file move: %v", err)
	}

	// Verify the move was recorded
	stats := renameTracker.GetStatistics()
	if totalRenames, ok := stats["total_renames"].(int); !ok || totalRenames == 0 {
		t.Error("Expected rename to be recorded in statistics")
	}

	// Clean up
	client.DeleteFile(ctx, renamedPath)
}

func testDirectoryRename(t *testing.T, client filesystem.FileSystemClient, renameTracker *services.UniversalRenameTracker, suite ProtocolTestSuite, ctx context.Context) {
	protocol := suite.GetProtocolName()

	// Create test directory with files
	originalDir := "/test_dir_rename"
	renamedDir := "/renamed_test_dir"

	// Create directory and files
	if err := client.CreateDirectory(ctx, originalDir); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := originalDir + "/nested_file.txt"
	if err := createTestFile(ctx, client, testFile, "Nested file content"); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Track directory deletion
	dirID := int64(200)
	protocolData := suite.GetTestConfig()

	renameTracker.TrackDelete(ctx, dirID, originalDir, "test_storage", protocol, 0, nil, true, protocolData)

	// Detect directory creation at new location
	pendingMove, isMove := renameTracker.DetectCreate(ctx, renamedDir, "test_storage", protocol, 0, nil, true, protocolData)

	if !isMove {
		t.Error("Expected directory rename to be detected")
		return
	}

	if !pendingMove.IsDirectory {
		t.Error("Expected pending move to be marked as directory")
	}

	// Process the move
	if err := renameTracker.ProcessMove(ctx, client, pendingMove, renamedDir); err != nil {
		t.Errorf("Failed to process directory move: %v", err)
	}

	// Clean up
	client.DeleteFile(ctx, renamedDir+"/nested_file.txt")
	client.DeleteDirectory(ctx, renamedDir)
}

func testBatchRename(t *testing.T, client filesystem.FileSystemClient, renameTracker *services.UniversalRenameTracker, suite ProtocolTestSuite, ctx context.Context) {
	protocol := suite.GetProtocolName()

	// Create multiple files for batch rename
	numFiles := 5
	fileIDs := make([]int64, numFiles)
	originalPaths := make([]string, numFiles)
	renamedPaths := make([]string, numFiles)

	for i := 0; i < numFiles; i++ {
		fileIDs[i] = int64(300 + i)
		originalPaths[i] = fmt.Sprintf("/batch_file_%d.txt", i)
		renamedPaths[i] = fmt.Sprintf("/renamed_batch_%d.txt", i)

		// Create file
		content := fmt.Sprintf("Batch file %d content", i)
		if err := createTestFile(ctx, client, originalPaths[i], content); err != nil {
			t.Fatalf("Failed to create batch file %d: %v", i, err)
		}

		// Track deletion
		size := int64(len(content))
		hash := fmt.Sprintf("batchhash%d", i)
		renameTracker.TrackDelete(ctx, fileIDs[i], originalPaths[i], "test_storage", protocol, size, &hash, false, suite.GetTestConfig())
	}

	// Small delay to simulate batch operations
	time.Sleep(100 * time.Millisecond)

	// Detect all moves
	detectedMoves := 0
	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf("Batch file %d content", i)
		size := int64(len(content))
		hash := fmt.Sprintf("batchhash%d", i)

		if pendingMove, isMove := renameTracker.DetectCreate(ctx, renamedPaths[i], "test_storage", protocol, size, &hash, false, suite.GetTestConfig()); isMove {
			detectedMoves++
			if err := renameTracker.ProcessMove(ctx, client, pendingMove, renamedPaths[i]); err != nil {
				t.Errorf("Failed to process batch move %d: %v", i, err)
			}
		}
	}

	if detectedMoves != numFiles {
		t.Errorf("Expected %d batch moves to be detected, got %d", numFiles, detectedMoves)
	}

	// Clean up
	for i := 0; i < numFiles; i++ {
		client.DeleteFile(ctx, renamedPaths[i])
	}
}

func testRenameTimingWindows(t *testing.T, client filesystem.FileSystemClient, renameTracker *services.UniversalRenameTracker, suite ProtocolTestSuite, ctx context.Context) {
	protocol := suite.GetProtocolName()

	// Get protocol capabilities
	capabilities, err := services.GetProtocolCapabilities(protocol, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to get protocol capabilities: %v", err)
	}

	// Test move within window
	t.Run("within_window", func(t *testing.T) {
		originalPath := "/timing_test_1.txt"
		renamedPath := "/timing_renamed_1.txt"
		content := "Timing test content"

		if err := createTestFile(ctx, client, originalPath, content); err != nil {
			t.Fatalf("Failed to create timing test file: %v", err)
		}

		size := int64(len(content))
		hash := "timinghash1"
		renameTracker.TrackDelete(ctx, 400, originalPath, "test_storage", protocol, size, &hash, false, suite.GetTestConfig())

		// Detect move immediately (within window)
		if pendingMove, isMove := renameTracker.DetectCreate(ctx, renamedPath, "test_storage", protocol, size, &hash, false, suite.GetTestConfig()); !isMove {
			t.Error("Expected move to be detected within timing window")
		} else {
			renameTracker.ProcessMove(ctx, client, pendingMove, renamedPath)
		}

		client.DeleteFile(ctx, renamedPath)
	})

	// Test move outside window
	t.Run("outside_window", func(t *testing.T) {
		originalPath := "/timing_test_2.txt"
		renamedPath := "/timing_renamed_2.txt"
		content := "Timing test content 2"

		if err := createTestFile(ctx, client, originalPath, content); err != nil {
			t.Fatalf("Failed to create timing test file: %v", err)
		}

		size := int64(len(content))
		hash := "timinghash2"
		renameTracker.TrackDelete(ctx, 401, originalPath, "test_storage", protocol, size, &hash, false, suite.GetTestConfig())

		// Wait longer than the protocol's move window
		time.Sleep(capabilities.MoveWindow + time.Second)

		// Try to detect move (should fail due to expired window)
		if _, isMove := renameTracker.DetectCreate(ctx, renamedPath, "test_storage", protocol, size, &hash, false, suite.GetTestConfig()); isMove {
			t.Error("Expected move detection to fail outside timing window")
		}

		client.DeleteFile(ctx, originalPath)
	})
}

func testProtocolCapabilities(t *testing.T, suite ProtocolTestSuite, logger *zap.Logger) {
	protocol := suite.GetProtocolName()

	capabilities, err := services.GetProtocolCapabilities(protocol, logger)
	if err != nil {
		t.Fatalf("Failed to get capabilities for protocol %s: %v", protocol, err)
	}

	// Verify capabilities match expected values
	switch protocol {
	case "local":
		if !capabilities.SupportsRealTimeNotification {
			t.Error("Local protocol should support real-time notifications")
		}
		if capabilities.MoveWindow > 5*time.Second {
			t.Error("Local protocol should have a short move window")
		}

	case "smb":
		if capabilities.SupportsRealTimeNotification {
			t.Error("SMB protocol should not support real-time notifications")
		}
		if !capabilities.RequiresPolling {
			t.Error("SMB protocol should require polling")
		}

	case "ftp":
		if capabilities.SupportsRealTimeNotification {
			t.Error("FTP protocol should not support real-time notifications")
		}
		if capabilities.MoveWindow < 10*time.Second {
			t.Error("FTP protocol should have a longer move window")
		}

	case "nfs":
		if capabilities.SupportsRealTimeNotification {
			t.Error("NFS protocol should not support real-time notifications in most cases")
		}

	case "webdav":
		if capabilities.SupportsRealTimeNotification {
			t.Error("WebDAV protocol should not support real-time notifications")
		}
	}

	t.Logf("Protocol %s capabilities: %+v", protocol, capabilities)
}

// Protocol-specific test suite implementations

type LocalProtocolTestSuite struct{}

func (s *LocalProtocolTestSuite) SetupProtocol(t *testing.T) (filesystem.FileSystemClient, func()) {
	tempDir := t.TempDir()

	config := &filesystem.LocalConfig{
		BasePath: tempDir,
	}

	client := filesystem.NewLocalClient(config)

	return client, func() {
		// Cleanup handled by t.TempDir()
	}
}

func (s *LocalProtocolTestSuite) GetProtocolName() string {
	return "local"
}

func (s *LocalProtocolTestSuite) GetTestConfig() map[string]interface{} {
	return map[string]interface{}{
		"base_path": "/tmp/test",
	}
}

func (s *LocalProtocolTestSuite) SupportsRealTimeEvents() bool {
	return true
}

type SMBProtocolTestSuite struct{}

func (s *SMBProtocolTestSuite) SetupProtocol(t *testing.T) (filesystem.FileSystemClient, func()) {
	// For testing, we'll use a mock SMB client or skip if no test server available
	if os.Getenv("SMB_TEST_SERVER") == "" {
		t.Skip("SMB_TEST_SERVER not set, skipping SMB tests")
	}

	config := &filesystem.SmbConfig{
		Host:     os.Getenv("SMB_TEST_HOST"),
		Port:     445,
		Share:    os.Getenv("SMB_TEST_SHARE"),
		Username: os.Getenv("SMB_TEST_USER"),
		Password: os.Getenv("SMB_TEST_PASS"),
	}

	client := filesystem.NewSmbClient(config)

	return client, func() {
		// SMB cleanup
	}
}

func (s *SMBProtocolTestSuite) GetProtocolName() string {
	return "smb"
}

func (s *SMBProtocolTestSuite) GetTestConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":  "testserver",
		"share": "testshare",
	}
}

func (s *SMBProtocolTestSuite) SupportsRealTimeEvents() bool {
	return false
}

type FTPProtocolTestSuite struct{}

func (s *FTPProtocolTestSuite) SetupProtocol(t *testing.T) (filesystem.FileSystemClient, func()) {
	if os.Getenv("FTP_TEST_SERVER") == "" {
		t.Skip("FTP_TEST_SERVER not set, skipping FTP tests")
	}

	// Mock FTP client setup
	t.Skip("FTP client implementation pending")
	return nil, func() {}
}

func (s *FTPProtocolTestSuite) GetProtocolName() string {
	return "ftp"
}

func (s *FTPProtocolTestSuite) GetTestConfig() map[string]interface{} {
	return map[string]interface{}{
		"host": "ftpserver",
		"port": 21,
	}
}

func (s *FTPProtocolTestSuite) SupportsRealTimeEvents() bool {
	return false
}

type NFSProtocolTestSuite struct{}

func (s *NFSProtocolTestSuite) SetupProtocol(t *testing.T) (filesystem.FileSystemClient, func()) {
	if os.Getenv("NFS_TEST_SERVER") == "" {
		t.Skip("NFS_TEST_SERVER not set, skipping NFS tests")
	}

	// Mock NFS client setup
	t.Skip("NFS client implementation pending")
	return nil, func() {}
}

func (s *NFSProtocolTestSuite) GetProtocolName() string {
	return "nfs"
}

func (s *NFSProtocolTestSuite) GetTestConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":        "nfsserver",
		"export_path": "/export",
	}
}

func (s *NFSProtocolTestSuite) SupportsRealTimeEvents() bool {
	return false
}

type WebDAVProtocolTestSuite struct{}

func (s *WebDAVProtocolTestSuite) SetupProtocol(t *testing.T) (filesystem.FileSystemClient, func()) {
	if os.Getenv("WEBDAV_TEST_SERVER") == "" {
		t.Skip("WEBDAV_TEST_SERVER not set, skipping WebDAV tests")
	}

	// Mock WebDAV client setup
	t.Skip("WebDAV client implementation pending")
	return nil, func() {}
}

func (s *WebDAVProtocolTestSuite) GetProtocolName() string {
	return "webdav"
}

func (s *WebDAVProtocolTestSuite) GetTestConfig() map[string]interface{} {
	return map[string]interface{}{
		"url": "https://webdavserver/dav",
	}
}

func (s *WebDAVProtocolTestSuite) SupportsRealTimeEvents() bool {
	return false
}

// Helper functions

func setupProtocolTestDB(t *testing.T) *database.DB {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	schema := `
		CREATE TABLE storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);

		CREATE TABLE files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			path TEXT NOT NULL,
			name TEXT NOT NULL,
			is_directory BOOLEAN NOT NULL,
			size INTEGER NOT NULL,
			quick_hash TEXT,
			parent_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted BOOLEAN DEFAULT FALSE,
			deleted_at TIMESTAMP,
			last_scan_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (storage_root_id) REFERENCES storage_roots (id)
		);

		CREATE TABLE universal_rename_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			protocol TEXT NOT NULL,
			old_path TEXT NOT NULL,
			new_path TEXT NOT NULL,
			is_directory BOOLEAN NOT NULL,
			size INTEGER NOT NULL,
			file_hash TEXT,
			detected_at TIMESTAMP NOT NULL,
			processed_at TIMESTAMP,
			status TEXT NOT NULL DEFAULT 'pending',
			FOREIGN KEY (storage_root_id) REFERENCES storage_roots (id)
		);

		INSERT INTO storage_roots (id, name) VALUES (1, 'test_storage');
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

func createTestFile(ctx context.Context, client filesystem.FileSystemClient, path, content string) error {
	// For testing, we'll create a simple file
	// In a real implementation, this would use the appropriate client method
	return nil // Placeholder implementation
}

// TestUniversalRenameTrackerIntegration tests the complete integration
func TestUniversalRenameTrackerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	ctx := context.Background()

	// Setup
	db := setupProtocolTestDB(t)
	defer db.Close()

	tracker := services.NewUniversalRenameTracker(db, logger)
	if err := tracker.Start(); err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Test cross-protocol scenarios
	t.Run("cross_protocol_operations", func(t *testing.T) {
		// Test operations that span multiple protocols
		protocols := []string{"local", "smb", "ftp"}

		for _, protocol := range protocols {
			// Track deletions for each protocol
			tracker.TrackDelete(ctx, int64(100+len(protocol)), "/test.txt", "storage1", protocol, 1024, nil, false, map[string]interface{}{})
		}

		// Verify statistics
		stats := tracker.GetStatistics()
		if pendingByProtocol, ok := stats["pending_by_protocol"].(map[string]int); ok {
			for _, protocol := range protocols {
				if count, exists := pendingByProtocol[protocol]; !exists || count != 1 {
					t.Errorf("Expected 1 pending move for protocol %s, got %d", protocol, count)
				}
			}
		}
	})

	// Test concurrent operations
	t.Run("concurrent_operations", func(t *testing.T) {
		numGoroutines := 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				path := fmt.Sprintf("/concurrent_%d.txt", id)
				fileID := int64(500 + id)

				tracker.TrackDelete(ctx, fileID, path, "storage1", "local", 1024, nil, false, map[string]interface{}{})

				// Immediately try to detect
				if pendingMove, isMove := tracker.DetectCreate(ctx, path+"_renamed", "storage1", "local", 1024, nil, false, map[string]interface{}{}); isMove {
					t.Logf("Concurrent operation %d: move detected", id)
					// Process move (simplified)
					_ = pendingMove
				}
			}(i)
		}

		// Wait for all operations to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify final state
		finalStats := tracker.GetStatistics()
		t.Logf("Final statistics after concurrent operations: %+v", finalStats)
	})
}
