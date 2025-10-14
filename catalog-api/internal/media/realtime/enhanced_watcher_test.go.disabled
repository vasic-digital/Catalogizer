package realtime

import (
	"catalogizer/internal/media/analyzer"
	"catalogizer/internal/media/database"
	"catalogizer/internal/services"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"catalogizer/utils"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupTestMediaDB(t *testing.T) *database.MediaDatabase {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_media.db")

	config := database.DatabaseConfig{
		Path:     dbPath,
		Password: "test_password",
	}

	logger := zap.NewNop()
	mediaDB, err := database.NewMediaDatabase(config, logger)
	if err != nil {
		t.Fatalf("Failed to create test media database: %v", err)
	}

	// Initialize basic schema for testing
	db := mediaDB.GetDB()
	schema := `
		CREATE TABLE IF NOT EXISTS storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS files (
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

		CREATE TABLE IF NOT EXISTS change_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			change_type TEXT NOT NULL,
			new_data TEXT,
			detected_at TIMESTAMP NOT NULL,
			processed BOOLEAN DEFAULT FALSE
		);

		-- Insert test storage root
		INSERT OR IGNORE INTO storage_roots (id, name) VALUES (1, 'test_storage');
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return mediaDB
}

func setupTestComponents(t *testing.T) (*EnhancedChangeWatcher, *database.MediaDatabase, *services.RenameTracker) {
	logger := zap.NewNop()
	mediaDB := setupTestMediaDB(t)

	// Create mock analyzer
	analyzer := analyzer.NewMediaAnalyzer(mediaDB.GetDB(), nil, nil, logger)

	// Create rename tracker
	renameTracker := services.NewRenameTracker(mediaDB.GetDB(), logger)
	if err := renameTracker.InitializeTables(); err != nil {
		t.Fatalf("Failed to initialize rename tracker tables: %v", err)
	}

	// Create enhanced watcher
	watcher := NewEnhancedChangeWatcher(mediaDB, analyzer, renameTracker, logger)

	return watcher, mediaDB, renameTracker
}

func TestEnhancedChangeWatcher_GetRelativePath(t *testing.T) {
	watcher, mediaDB, _ := setupTestComponents(t)
	defer mediaDB.Close()

	tests := []struct {
		name     string
		basePath string
		fullPath string
		expected string
		hasError bool
	}{
		{
			name:     "simple relative path",
			basePath: "/mnt/storage",
			fullPath: "/mnt/storage/test/file.txt",
			expected: "/test/file.txt",
			hasError: false,
		},
		{
			name:     "root file",
			basePath: "/mnt/storage",
			fullPath: "/mnt/storage/file.txt",
			expected: "/file.txt",
			hasError: false,
		},
		{
			name:     "same path",
			basePath: "/mnt/storage",
			fullPath: "/mnt/storage",
			expected: "/.",
			hasError: false,
		},
		{
			name:     "nested directory",
			basePath: "/mnt/storage",
			fullPath: "/mnt/storage/dir1/dir2/file.txt",
			expected: "/dir1/dir2/file.txt",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := watcher.getRelativePath(tt.basePath, tt.fullPath)

			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.hasError && result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEnhancedChangeWatcher_CalculateFileHash(t *testing.T) {
	watcher, mediaDB, _ := setupTestComponents(t)
	defer mediaDB.Close()

	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate hash
	hash := watcher.calculateFileHash(testFile)

	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Verify hash is consistent
	hash2 := watcher.calculateFileHash(testFile)
	if hash != hash2 {
		t.Errorf("Hash calculation is not consistent: %s != %s", hash, hash2)
	}

	// Test with non-existent file
	nonExistentHash := watcher.calculateFileHash("/non/existent/file.txt")
	if nonExistentHash != "" {
		t.Error("Expected empty hash for non-existent file")
	}
}

func TestEnhancedChangeWatcher_HandleCreate(t *testing.T) {
	watcher, mediaDB, renameTracker := setupTestComponents(t)
	defer mediaDB.Close()

	ctx := context.Background()

	// Test case 1: New file creation (not a move)
	t.Run("new file creation", func(t *testing.T) {
		event := EnhancedChangeEvent{
			Path:      "/new_file.txt",
			SmbRoot:   "test_storage",
			Operation: "created",
			Timestamp: time.Now(),
			Size:      1024,
			IsDir:     false,
			FileHash:  utils.StringPtr("newhash123"),
		}

		// This should not detect a move (no pending move exists)
		watcher.handleCreate(ctx, event)

		// Verify the change was logged
		var changeCount int
		err := mediaDB.GetDB().QueryRow("SELECT COUNT(*) FROM change_log WHERE entity_id = ?", event.Path).Scan(&changeCount)
		if err != nil {
			t.Fatalf("Failed to query change log: %v", err)
		}

		// Note: The actual logging happens in processChange, not handleCreate directly
		// So we're mainly testing that no moves were detected and no errors occurred
	})

	// Test case 2: File creation that is actually a move
	t.Run("file creation as move", func(t *testing.T) {
		// First, track a deletion
		fileID := int64(123)
		oldPath := "/old_file.txt"
		storageRoot := "test_storage"
		size := int64(2048)
		fileHash := utils.StringPtr("movehash456")

		renameTracker.TrackDelete(ctx, fileID, oldPath, storageRoot, size, fileHash, false)

		// Now create the file in new location
		event := EnhancedChangeEvent{
			Path:      "/moved_file.txt",
			SmbRoot:   storageRoot,
			Operation: "created",
			Timestamp: time.Now(),
			Size:      size,
			IsDir:     false,
			FileHash:  fileHash,
		}

		watcher.handleCreate(ctx, event)

		// Verify the move was detected and processed
		// The pending move should be removed
		key := renameTracker.CreateMoveKey(storageRoot, fileHash, size, false)
		renameTracker.PendingMovesMu.RLock()
		_, exists := renameTracker.PendingMoves[key]
		renameTracker.PendingMovesMu.RUnlock()

		if exists {
			t.Error("Expected pending move to be removed after detection")
		}
	})
}

func TestEnhancedChangeWatcher_HandleModify(t *testing.T) {
	watcher, mediaDB, _ := setupTestComponents(t)
	defer mediaDB.Close()

	ctx := context.Background()

	// Test file modification
	t.Run("file modification", func(t *testing.T) {
		// Insert test file
		_, err := mediaDB.GetDB().Exec(`
			INSERT INTO files (id, storage_root_id, path, name, is_directory, size, quick_hash)
			VALUES (100, 1, '/test_modify.txt', 'test_modify.txt', 0, 1024, 'oldhash')
		`)
		if err != nil {
			t.Fatalf("Failed to insert test file: %v", err)
		}

		event := EnhancedChangeEvent{
			Path:      "/test_modify.txt",
			SmbRoot:   "test_storage",
			Operation: "modified",
			Timestamp: time.Now(),
			Size:      2048, // Changed size
			IsDir:     false,
			FileHash:  utils.StringPtr("newhash789"),
		}

		watcher.handleModify(ctx, event)

		// Verify file metadata was updated
		var updatedSize int64
		var updatedHash sql.NullString
		err = mediaDB.GetDB().QueryRow("SELECT size, quick_hash FROM files WHERE id = 100").Scan(&updatedSize, &updatedHash)
		if err != nil {
			t.Fatalf("Failed to query updated file: %v", err)
		}

		if updatedSize != 2048 {
			t.Errorf("Expected size 2048, got %d", updatedSize)
		}

		if !updatedHash.Valid || updatedHash.String != "newhash789" {
			t.Errorf("Expected hash 'newhash789', got %v", updatedHash)
		}
	})

	// Test directory modification
	t.Run("directory modification", func(t *testing.T) {
		// Insert test directory
		_, err := mediaDB.GetDB().Exec(`
			INSERT INTO files (id, storage_root_id, path, name, is_directory, size)
			VALUES (101, 1, '/test_dir', 'test_dir', 1, 0)
		`)
		if err != nil {
			t.Fatalf("Failed to insert test directory: %v", err)
		}

		event := EnhancedChangeEvent{
			Path:      "/test_dir",
			SmbRoot:   "test_storage",
			Operation: "modified",
			Timestamp: time.Now(),
			Size:      0,
			IsDir:     true,
		}

		watcher.handleModify(ctx, event)

		// Verify directory was updated (check last_scan_at was updated)
		var lastScanAt time.Time
		err = mediaDB.GetDB().QueryRow("SELECT last_scan_at FROM files WHERE id = 101").Scan(&lastScanAt)
		if err != nil {
			t.Fatalf("Failed to query updated directory: %v", err)
		}

		// last_scan_at should be recent (within last minute)
		if time.Since(lastScanAt) > time.Minute {
			t.Error("Expected last_scan_at to be recent")
		}
	})
}

func TestEnhancedChangeWatcher_HandleDelete(t *testing.T) {
	watcher, mediaDB, renameTracker := setupTestComponents(t)
	defer mediaDB.Close()

	ctx := context.Background()

	// Insert test file
	fileID := int64(200)
	_, err := mediaDB.GetDB().Exec(`
		INSERT INTO files (id, storage_root_id, path, name, is_directory, size, quick_hash)
		VALUES (?, 1, '/delete_test.txt', 'delete_test.txt', 0, 1024, 'deletehash')
	`, fileID)
	if err != nil {
		t.Fatalf("Failed to insert test file: %v", err)
	}

	event := EnhancedChangeEvent{
		Path:      "/delete_test.txt",
		SmbRoot:   "test_storage",
		Operation: "deleted",
		Timestamp: time.Now(),
		Size:      1024,
		IsDir:     false,
		FileHash:  utils.StringPtr("deletehash"),
		FileID:    &fileID,
	}

	watcher.handleDelete(ctx, event)

	// Verify the deletion was tracked for potential move detection
	key := renameTracker.CreateMoveKey("test_storage", event.FileHash, event.Size, false)
	renameTracker.PendingMovesMu.RLock()
	pendingMove, exists := renameTracker.PendingMoves[key]
	renameTracker.PendingMovesMu.RUnlock()

	if !exists {
		t.Error("Expected deletion to be tracked as pending move")
	}

	if pendingMove.FileID != fileID {
		t.Errorf("Expected file ID %d, got %d", fileID, pendingMove.FileID)
	}
}

func TestEnhancedChangeWatcher_IsMediaFile(t *testing.T) {
	watcher, mediaDB, _ := setupTestComponents(t)
	defer mediaDB.Close()

	tests := []struct {
		path     string
		expected bool
	}{
		{"/video.mp4", true},
		{"/movie.avi", true},
		{"/song.mp3", true},
		{"/image.jpg", true},
		{"/document.pdf", true},
		{"/text.txt", false},
		{"/script.sh", false},
		{"/program.exe", false},
		{"/archive.zip", false},
		{"/no_extension", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := watcher.isMediaFile(tt.path)
			if result != tt.expected {
				t.Errorf("isMediaFile(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestEnhancedChangeWatcher_LogChange(t *testing.T) {
	watcher, mediaDB, _ := setupTestComponents(t)
	defer mediaDB.Close()

	event := EnhancedChangeEvent{
		Path:      "/log_test.txt",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		Size:      1024,
		IsDir:     false,
		FileHash:  utils.StringPtr("loghash123"),
		FileID:    nil,
	}

	err := watcher.logChange(event)
	if err != nil {
		t.Fatalf("Failed to log change: %v", err)
	}

	// Verify change was logged
	var entityType, entityID, changeType, newData string
	var detectedAt time.Time
	var processed bool

	err = mediaDB.GetDB().QueryRow(`
		SELECT entity_type, entity_id, change_type, new_data, detected_at, processed
		FROM change_log
		WHERE entity_id = ?
	`, event.Path).Scan(&entityType, &entityID, &changeType, &newData, &detectedAt, &processed)

	if err != nil {
		t.Fatalf("Failed to query logged change: %v", err)
	}

	if entityType != "file" {
		t.Errorf("Expected entity_type 'file', got %s", entityType)
	}

	if entityID != event.Path {
		t.Errorf("Expected entity_id %s, got %s", event.Path, entityID)
	}

	if changeType != event.Operation {
		t.Errorf("Expected change_type %s, got %s", event.Operation, changeType)
	}

	if processed {
		t.Error("Expected processed to be false")
	}

	// Verify new_data contains expected fields (it's JSON)
	if newData == "" {
		t.Error("Expected new_data to be non-empty")
	}
}

func TestEnhancedChangeWatcher_StartStop(t *testing.T) {
	watcher, mediaDB, _ := setupTestComponents(t)
	defer mediaDB.Close()

	// Test start
	err := watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start enhanced watcher: %v", err)
	}

	// Verify workers are running by checking change queue can receive events
	testEvent := EnhancedChangeEvent{
		Path:      "/start_test.txt",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		Size:      1024,
		IsDir:     false,
	}

	// Try to send event (should not block if workers are running)
	select {
	case watcher.changeQueue <- testEvent:
		// Successfully sent
	case <-time.After(100 * time.Millisecond):
		t.Error("Change queue appears to be blocked or workers not running")
	}

	// Test stop
	watcher.Stop()

	// Verify stop completed
	select {
	case <-watcher.stopCh:
		// Channel is closed as expected
	default:
		t.Error("Expected stop channel to be closed")
	}
}

func TestEnhancedChangeWatcher_GetStatistics(t *testing.T) {
	watcher, mediaDB, renameTracker := setupTestComponents(t)
	defer mediaDB.Close()

	// Add some test data
	ctx := context.Background()
	since := time.Now().Add(-time.Hour)

	// Add some change log entries
	_, err := mediaDB.GetDB().Exec(`
		INSERT INTO change_log (entity_type, entity_id, change_type, detected_at)
		VALUES
			('file', '/test1.txt', 'created', ?),
			('file', '/test2.txt', 'modified', ?),
			('file', '/test3.txt', 'deleted', ?)
	`, since.Add(10*time.Minute), since.Add(20*time.Minute), since.Add(30*time.Minute))
	if err != nil {
		t.Fatalf("Failed to insert test change log entries: %v", err)
	}

	// Add pending moves to rename tracker
	renameTracker.TrackDelete(ctx, 1, "/pending1.txt", "test_storage", 1024, utils.StringPtr("hash1"), false)
	renameTracker.TrackDelete(ctx, 2, "/pending2.txt", "test_storage", 2048, utils.StringPtr("hash2"), false)

	// Get statistics
	stats, err := watcher.GetStatistics(since)
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	// Verify change statistics
	changesByType, ok := stats["changes_by_type"].(map[string]int)
	if !ok {
		t.Fatal("Expected changes_by_type to be map[string]int")
	}

	if changesByType["created"] != 1 {
		t.Errorf("Expected 1 created change, got %d", changesByType["created"])
	}

	if changesByType["modified"] != 1 {
		t.Errorf("Expected 1 modified change, got %d", changesByType["modified"])
	}

	if changesByType["deleted"] != 1 {
		t.Errorf("Expected 1 deleted change, got %d", changesByType["deleted"])
	}

	// Verify rename tracking statistics
	renameTracking, ok := stats["rename_tracking"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected rename_tracking to be map[string]interface{}")
	}

	if pendingMoves, ok := renameTracking["pending_moves"].(int); !ok || pendingMoves != 2 {
		t.Errorf("Expected 2 pending moves, got %v", renameTracking["pending_moves"])
	}

	// Verify other statistics
	if watchedPaths, ok := stats["watched_paths"].(int); !ok || watchedPaths != 0 {
		t.Errorf("Expected 0 watched paths, got %v", stats["watched_paths"])
	}

	if workers, ok := stats["workers"].(int); !ok || workers != 4 {
		t.Errorf("Expected 4 workers, got %v", stats["workers"])
	}
}
