package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create required tables
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

		CREATE TABLE rename_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
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

		-- Insert test storage root
		INSERT INTO storage_roots (id, name) VALUES (1, 'test_storage');

		-- Insert test files
		INSERT INTO files (id, storage_root_id, path, name, is_directory, size, quick_hash)
		VALUES
			(1, 1, '/test_file.txt', 'test_file.txt', 0, 1024, 'abcd1234'),
			(2, 1, '/test_dir', 'test_dir', 1, 0, NULL),
			(3, 1, '/test_dir/nested_file.txt', 'nested_file.txt', 0, 2048, 'efgh5678');
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

func TestRenameTracker_CreateMoveKey(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	tracker := NewRenameTracker(db, logger)

	tests := []struct {
		name        string
		storageRoot string
		fileHash    *string
		size        int64
		isDirectory bool
		expected    string
	}{
		{
			name:        "file with hash",
			storageRoot: "storage1",
			fileHash:    stringPtr("abcd1234"),
			size:        1024,
			isDirectory: false,
			expected:    "storage1:abcd1234:1024:false",
		},
		{
			name:        "file without hash",
			storageRoot: "storage1",
			fileHash:    nil,
			size:        2048,
			isDirectory: false,
			expected:    "storage1:nil:2048:false",
		},
		{
			name:        "directory",
			storageRoot: "storage2",
			fileHash:    nil,
			size:        0,
			isDirectory: true,
			expected:    "storage2:nil:0:true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tracker.createMoveKey(tt.storageRoot, tt.fileHash, tt.size, tt.isDirectory)
			if result != tt.expected {
				t.Errorf("createMoveKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRenameTracker_TrackDeleteAndDetectCreate(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	tracker := NewRenameTracker(db, logger)
	ctx := context.Background()

	// Track a file deletion
	fileID := int64(1)
	path := "/test_file.txt"
	storageRoot := "test_storage"
	size := int64(1024)
	fileHash := stringPtr("abcd1234")
	isDirectory := false

	tracker.TrackDelete(ctx, fileID, path, storageRoot, size, fileHash, isDirectory)

	// Verify the move is tracked
	key := tracker.createMoveKey(storageRoot, fileHash, size, isDirectory)
	tracker.pendingMovesMu.RLock()
	pendingMove, exists := tracker.pendingMoves[key]
	tracker.pendingMovesMu.RUnlock()

	if !exists {
		t.Fatal("Expected pending move to be tracked")
	}

	if pendingMove.Path != path {
		t.Errorf("Expected path %s, got %s", path, pendingMove.Path)
	}
	if pendingMove.FileID != fileID {
		t.Errorf("Expected file ID %d, got %d", fileID, pendingMove.FileID)
	}

	// Test detecting create (move)
	newPath := "/moved_file.txt"
	detectedMove, isMove := tracker.DetectCreate(ctx, newPath, storageRoot, size, fileHash, isDirectory)

	if !isMove {
		t.Fatal("Expected move to be detected")
	}

	if detectedMove.Path != path {
		t.Errorf("Expected original path %s, got %s", path, detectedMove.Path)
	}

	// Verify pending move is removed
	tracker.pendingMovesMu.RLock()
	_, stillExists := tracker.pendingMoves[key]
	tracker.pendingMovesMu.RUnlock()

	if stillExists {
		t.Error("Expected pending move to be removed after detection")
	}
}

func TestRenameTracker_DetectCreateNonExistent(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	tracker := NewRenameTracker(db, logger)
	ctx := context.Background()

	// Try to detect create for non-existent move
	newPath := "/new_file.txt"
	storageRoot := "test_storage"
	size := int64(1024)
	fileHash := stringPtr("nonexistent")
	isDirectory := false

	detectedMove, isMove := tracker.DetectCreate(ctx, newPath, storageRoot, size, fileHash, isDirectory)

	if isMove {
		t.Error("Expected no move to be detected for non-existent file")
	}

	if detectedMove != nil {
		t.Error("Expected nil detectedMove for non-existent file")
	}
}

func TestRenameTracker_DetectCreateExpired(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	// Set very short move window for testing
	tracker := NewRenameTracker(db, logger)
	tracker.moveWindow = 10 * time.Millisecond

	ctx := context.Background()

	// Track a file deletion
	fileID := int64(1)
	path := "/test_file.txt"
	storageRoot := "test_storage"
	size := int64(1024)
	fileHash := stringPtr("abcd1234")
	isDirectory := false

	tracker.TrackDelete(ctx, fileID, path, storageRoot, size, fileHash, isDirectory)

	// Wait for move window to expire
	time.Sleep(20 * time.Millisecond)

	// Try to detect create (should fail due to expired window)
	newPath := "/moved_file.txt"
	detectedMove, isMove := tracker.DetectCreate(ctx, newPath, storageRoot, size, fileHash, isDirectory)

	if isMove {
		t.Error("Expected no move to be detected for expired move")
	}

	if detectedMove != nil {
		t.Error("Expected nil detectedMove for expired move")
	}
}

func TestRenameTracker_ProcessMove(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	tracker := NewRenameTracker(db, logger)
	if err := tracker.initializeTables(); err != nil {
		t.Fatalf("Failed to initialize tables: %v", err)
	}

	ctx := context.Background()

	// Create a pending move
	pendingMove := &PendingMove{
		Path:        "/test_file.txt",
		StorageRoot: "test_storage",
		Size:        1024,
		FileHash:    stringPtr("abcd1234"),
		IsDirectory: false,
		DeletedAt:   time.Now(),
		FileID:      1,
	}

	newPath := "/moved_file.txt"

	// Process the move
	err := tracker.ProcessMove(ctx, pendingMove, newPath)
	if err != nil {
		t.Fatalf("Failed to process move: %v", err)
	}

	// Verify file path was updated
	var updatedPath string
	err = db.QueryRow("SELECT path FROM files WHERE id = ?", pendingMove.FileID).Scan(&updatedPath)
	if err != nil {
		t.Fatalf("Failed to query updated file path: %v", err)
	}

	if updatedPath != newPath {
		t.Errorf("Expected updated path %s, got %s", newPath, updatedPath)
	}

	// Verify rename event was recorded
	var eventCount int
	err = db.QueryRow("SELECT COUNT(*) FROM rename_events WHERE old_path = ? AND new_path = ?", pendingMove.Path, newPath).Scan(&eventCount)
	if err != nil {
		t.Fatalf("Failed to query rename events: %v", err)
	}

	if eventCount != 1 {
		t.Errorf("Expected 1 rename event, got %d", eventCount)
	}
}

func TestRenameTracker_ProcessDirectoryMove(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	tracker := NewRenameTracker(db, logger)
	if err := tracker.initializeTables(); err != nil {
		t.Fatalf("Failed to initialize tables: %v", err)
	}

	ctx := context.Background()

	// Create a pending directory move
	pendingMove := &PendingMove{
		Path:        "/test_dir",
		StorageRoot: "test_storage",
		Size:        0,
		FileHash:    nil,
		IsDirectory: true,
		DeletedAt:   time.Now(),
		FileID:      2,
	}

	newPath := "/moved_dir"

	// Process the directory move
	err := tracker.ProcessMove(ctx, pendingMove, newPath)
	if err != nil {
		t.Fatalf("Failed to process directory move: %v", err)
	}

	// Verify directory path was updated
	var updatedDirPath string
	err = db.QueryRow("SELECT path FROM files WHERE id = ?", pendingMove.FileID).Scan(&updatedDirPath)
	if err != nil {
		t.Fatalf("Failed to query updated directory path: %v", err)
	}

	if updatedDirPath != newPath {
		t.Errorf("Expected updated directory path %s, got %s", newPath, updatedDirPath)
	}

	// Verify nested file path was also updated
	var nestedFilePath string
	err = db.QueryRow("SELECT path FROM files WHERE id = 3").Scan(&nestedFilePath)
	if err != nil {
		t.Fatalf("Failed to query nested file path: %v", err)
	}

	expectedNestedPath := "/moved_dir/nested_file.txt"
	if nestedFilePath != expectedNestedPath {
		t.Errorf("Expected nested file path %s, got %s", expectedNestedPath, nestedFilePath)
	}
}

func TestRenameTracker_CleanupExpiredMoves(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	// Set very short move window for testing
	tracker := NewRenameTracker(db, logger)
	tracker.moveWindow = 10 * time.Millisecond

	ctx := context.Background()

	// Track some file deletions
	tracker.TrackDelete(ctx, 1, "/file1.txt", "test_storage", 1024, stringPtr("hash1"), false)
	tracker.TrackDelete(ctx, 2, "/file2.txt", "test_storage", 2048, stringPtr("hash2"), false)

	// Verify moves are tracked
	tracker.pendingMovesMu.RLock()
	initialCount := len(tracker.pendingMoves)
	tracker.pendingMovesMu.RUnlock()

	if initialCount != 2 {
		t.Errorf("Expected 2 pending moves, got %d", initialCount)
	}

	// Wait for moves to expire
	time.Sleep(20 * time.Millisecond)

	// Trigger cleanup
	tracker.cleanupExpiredMoves()

	// Verify moves were cleaned up
	tracker.pendingMovesMu.RLock()
	finalCount := len(tracker.pendingMoves)
	tracker.pendingMovesMu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 pending moves after cleanup, got %d", finalCount)
	}
}

func TestRenameTracker_GetStatistics(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	tracker := NewRenameTracker(db, logger)
	if err := tracker.initializeTables(); err != nil {
		t.Fatalf("Failed to initialize tables: %v", err)
	}

	ctx := context.Background()

	// Add some pending moves
	tracker.TrackDelete(ctx, 1, "/file1.txt", "test_storage", 1024, stringPtr("hash1"), false)
	tracker.TrackDelete(ctx, 2, "/file2.txt", "test_storage", 2048, stringPtr("hash2"), false)

	// Add some rename events to database
	_, err := db.Exec(`
		INSERT INTO rename_events (storage_root_id, old_path, new_path, is_directory, size, detected_at, status)
		VALUES
			(1, '/old1.txt', '/new1.txt', 0, 1024, datetime('now'), 'processed'),
			(1, '/old2.txt', '/new2.txt', 0, 2048, datetime('now'), 'processed'),
			(1, '/old3.txt', '/new3.txt', 0, 512, datetime('now'), 'failed')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test rename events: %v", err)
	}

	// Get statistics
	stats := tracker.GetStatistics()

	// Verify statistics
	if pendingMoves, ok := stats["pending_moves"].(int); !ok || pendingMoves != 2 {
		t.Errorf("Expected 2 pending moves in statistics, got %v", stats["pending_moves"])
	}

	if totalRenames, ok := stats["total_renames"].(int); !ok || totalRenames != 3 {
		t.Errorf("Expected 3 total renames in statistics, got %v", stats["total_renames"])
	}

	if successfulRenames, ok := stats["successful_renames"].(int); !ok || successfulRenames != 2 {
		t.Errorf("Expected 2 successful renames in statistics, got %v", stats["successful_renames"])
	}

	if successRate, ok := stats["success_rate"].(float64); !ok || successRate != 66.66666666666667 {
		t.Errorf("Expected success rate ~66.67%%, got %v", stats["success_rate"])
	}
}

func TestRenameTracker_GetRenameEvents(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	tracker := NewRenameTracker(db, logger)
	if err := tracker.initializeTables(); err != nil {
		t.Fatalf("Failed to initialize tables: %v", err)
	}

	ctx := context.Background()

	// Add some rename events to database
	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO rename_events (storage_root_id, old_path, new_path, is_directory, size, detected_at, status)
		VALUES
			(1, '/old1.txt', '/new1.txt', 0, 1024, ?, 'processed'),
			(1, '/old2.txt', '/new2.txt', 0, 2048, ?, 'pending')
	`, now, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert test rename events: %v", err)
	}

	// Get rename events
	events, err := tracker.GetRenameEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get rename events: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 rename events, got %d", len(events))
	}

	// Verify events are sorted by detected_at DESC
	if events[0].OldPath != "/old1.txt" {
		t.Errorf("Expected first event to be most recent, got %s", events[0].OldPath)
	}

	if events[1].OldPath != "/old2.txt" {
		t.Errorf("Expected second event to be older, got %s", events[1].OldPath)
	}
}

func TestRenameTracker_StartStop(t *testing.T) {
	logger := zap.NewNop()
	db := setupTestDB(t)
	defer db.Close()

	tracker := NewRenameTracker(db, logger)

	// Test start
	err := tracker.Start()
	if err != nil {
		t.Fatalf("Failed to start rename tracker: %v", err)
	}

	// Test that cleanup worker is running by adding a pending move and waiting
	ctx := context.Background()
	tracker.TrackDelete(ctx, 1, "/test.txt", "test_storage", 1024, stringPtr("hash"), false)

	// Verify move is tracked
	tracker.pendingMovesMu.RLock()
	count := len(tracker.pendingMoves)
	tracker.pendingMovesMu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 pending move, got %d", count)
	}

	// Test stop
	tracker.Stop()

	// Give some time for stop to complete
	time.Sleep(10 * time.Millisecond)

	// Verify channels are closed by trying to send (should not panic)
	select {
	case <-tracker.stopCh:
		// Channel is closed as expected
	default:
		t.Error("Expected stop channel to be closed")
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}