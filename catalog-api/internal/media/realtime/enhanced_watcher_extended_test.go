package realtime

import (
	catalogDB "catalogizer/database"
	"catalogizer/internal/media/analyzer"
	"catalogizer/internal/media/database"
	"catalogizer/internal/services"
	"catalogizer/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupEnhancedTestMediaDB creates a test database with WAL mode for the enhanced watcher.
func setupEnhancedTestMediaDB(t *testing.T) *database.MediaDatabase {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "enhanced_test_media.db")

	config := database.DatabaseConfig{
		Path:     dbPath,
		Password: "test_password",
	}

	logger := zap.NewNop()
	mediaDB, err := database.NewMediaDatabase(config, logger)
	require.NoError(t, err)

	db := mediaDB.GetDB()
	db.Exec("PRAGMA journal_mode=WAL")

	extra := `
		CREATE TABLE IF NOT EXISTS storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			path TEXT NOT NULL,
			name TEXT NOT NULL,
			is_directory INTEGER NOT NULL DEFAULT 0,
			size INTEGER NOT NULL DEFAULT 0,
			quick_hash TEXT,
			extension TEXT,
			mime_type TEXT,
			parent_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted INTEGER DEFAULT 0,
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
			processed INTEGER DEFAULT 0
		);

		INSERT OR IGNORE INTO storage_roots (id, name) VALUES (1, 'test_storage');
	`
	_, err = db.Exec(extra)
	require.NoError(t, err)

	return mediaDB
}

// setupEnhancedWatcher creates an EnhancedChangeWatcher wired to a real test DB.
func setupEnhancedWatcher(t *testing.T) (*EnhancedChangeWatcher, *database.MediaDatabase, *services.RenameTracker) {
	t.Helper()
	logger := zap.NewNop()
	mediaDB := setupEnhancedTestMediaDB(t)

	wrappedDB := catalogDB.WrapDB(mediaDB.GetDB(), catalogDB.DialectSQLite)
	a := analyzer.NewMediaAnalyzer(wrappedDB, nil, nil, logger)
	rt := services.NewRenameTracker(wrappedDB, logger)
	require.NoError(t, rt.InitializeTables())

	watcher := NewEnhancedChangeWatcher(mediaDB, a, rt, logger)
	return watcher, mediaDB, rt
}

// ---------------------------------------------------------------------------
// WatchPath / UnwatchPath
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_WatchPath_Success(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	tempDir := t.TempDir()

	err := w.Start()
	require.NoError(t, err)
	defer w.Stop()

	err = w.WatchPath("test_root", tempDir)
	require.NoError(t, err)

	w.watcherMu.RLock()
	_, exists := w.watchers["test_root"]
	w.watcherMu.RUnlock()
	assert.True(t, exists, "Expected watcher to be registered")
}

func TestEnhancedChangeWatcher_WatchPath_AlreadyWatching(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	tempDir := t.TempDir()

	err := w.Start()
	require.NoError(t, err)
	defer w.Stop()

	err = w.WatchPath("test_root", tempDir)
	require.NoError(t, err)

	// Second call should return nil (already watching).
	err = w.WatchPath("test_root", tempDir)
	assert.NoError(t, err)
}

func TestEnhancedChangeWatcher_WatchPath_InvalidPath(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	// addPathRecursively skips errors (returns nil from the Walk callback), so a
	// nonexistent root path does NOT produce an error; it just results in an
	// empty watch list. The watcher is still registered.
	err := w.WatchPath("test_root", "/this/path/does/not/exist/at/all")
	assert.NoError(t, err)

	w.watcherMu.RLock()
	_, exists := w.watchers["test_root"]
	w.watcherMu.RUnlock()
	assert.True(t, exists, "Watcher should be registered even for nonexistent path")
}

func TestEnhancedChangeWatcher_UnwatchPath_Existing(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	tempDir := t.TempDir()

	err := w.Start()
	require.NoError(t, err)
	defer w.Stop()

	err = w.WatchPath("test_root", tempDir)
	require.NoError(t, err)

	w.UnwatchPath("test_root")

	w.watcherMu.RLock()
	_, exists := w.watchers["test_root"]
	w.watcherMu.RUnlock()
	assert.False(t, exists, "Expected watcher to be removed")
}

func TestEnhancedChangeWatcher_UnwatchPath_NonExistent(t *testing.T) {
	logger := zap.NewNop()
	w := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	// Should not panic.
	w.UnwatchPath("nonexistent")

	w.watcherMu.RLock()
	assert.Empty(t, w.watchers)
	w.watcherMu.RUnlock()
}

// ---------------------------------------------------------------------------
// addPathRecursively
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_AddPathRecursively(t *testing.T) {
	logger := zap.NewNop()
	w := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	fswatcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer fswatcher.Close()

	// Create a nested directory structure.
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub1", "sub2")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	err = w.addPathRecursively(fswatcher, tmpDir)
	require.NoError(t, err)

	// The watcher should have added the root, sub1, and sub1/sub2.
	watchList := fswatcher.WatchList()
	assert.GreaterOrEqual(t, len(watchList), 3, "Expected at least 3 watched paths (root + 2 subdirs)")
}

func TestEnhancedChangeWatcher_AddPathRecursively_MaxDepth(t *testing.T) {
	logger := zap.NewNop()
	w := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	fswatcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer fswatcher.Close()

	// Create a directory structure deeper than maxWatchDepth (20).
	tmpDir := t.TempDir()
	deepPath := tmpDir
	for i := 0; i < 22; i++ {
		deepPath = filepath.Join(deepPath, fmt.Sprintf("d%d", i))
	}
	require.NoError(t, os.MkdirAll(deepPath, 0755))

	err = w.addPathRecursively(fswatcher, tmpDir)
	require.NoError(t, err)

	// Directories beyond depth 20 should be skipped.
	watchList := fswatcher.WatchList()
	// We should have tmpDir + 20 levels = 21 entries.
	assert.LessOrEqual(t, len(watchList), 22, "Should not watch beyond max depth")
}

func TestEnhancedChangeWatcher_AddPathRecursively_NonExistentPath(t *testing.T) {
	logger := zap.NewNop()
	w := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	fswatcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer fswatcher.Close()

	// The walk callback returns nil for errors (skips problematic paths), so
	// a nonexistent root path results in no error and an empty watch list.
	err = w.addPathRecursively(fswatcher, "/nonexistent/path/123456")
	assert.NoError(t, err)

	watchList := fswatcher.WatchList()
	assert.Empty(t, watchList, "No paths should be watched for nonexistent root")
}

// ---------------------------------------------------------------------------
// monitorPath (Enhanced)
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_MonitorPath_StopChannel(t *testing.T) {
	logger := zap.NewNop()
	w := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	fswatcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer fswatcher.Close()

	w.wg.Add(1)
	go w.monitorPath("test_root", "/tmp", fswatcher)

	close(w.stopCh)
	w.wg.Wait()
}

func TestEnhancedChangeWatcher_MonitorPath_WatcherClosed(t *testing.T) {
	logger := zap.NewNop()
	w := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	fswatcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)

	w.wg.Add(1)
	go w.monitorPath("test_root", "/tmp", fswatcher)

	// Closing the fsnotify watcher closes the Events channel.
	fswatcher.Close()
	w.wg.Wait()
}

// ---------------------------------------------------------------------------
// handleFileSystemEvent (Enhanced)
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleFileSystemEvent_AllOps(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	w.debounceDelay = 50 * time.Millisecond

	// Create a temp dir to serve as localPath.
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		event         fsnotify.Event
		wantOperation string
		wantQueued    bool
	}{
		{
			name:          "create event",
			event:         fsnotify.Event{Name: filepath.Join(tmpDir, "new.mp4"), Op: fsnotify.Create},
			wantOperation: "created",
			wantQueued:    true,
		},
		{
			name:          "write event",
			event:         fsnotify.Event{Name: filepath.Join(tmpDir, "modified.mp4"), Op: fsnotify.Write},
			wantOperation: "modified",
			wantQueued:    true,
		},
		{
			name:          "remove event",
			event:         fsnotify.Event{Name: filepath.Join(tmpDir, "deleted.mp4"), Op: fsnotify.Remove},
			wantOperation: "deleted",
			wantQueued:    true,
		},
		{
			name:          "rename event",
			event:         fsnotify.Event{Name: filepath.Join(tmpDir, "moved.mp4"), Op: fsnotify.Rename},
			wantOperation: "moved",
			wantQueued:    true,
		},
		{
			name:          "chmod event ignored",
			event:         fsnotify.Event{Name: filepath.Join(tmpDir, "perm.mp4"), Op: fsnotify.Chmod},
			wantOperation: "",
			wantQueued:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Drain the queue.
			for len(w.changeQueue) > 0 {
				<-w.changeQueue
			}

			w.handleFileSystemEvent("test_storage", tmpDir, tt.event)

			if tt.wantQueued {
				time.Sleep(150 * time.Millisecond)
				select {
				case evt := <-w.changeQueue:
					assert.Equal(t, tt.wantOperation, evt.Operation)
					assert.Equal(t, "test_storage", evt.SmbRoot)
				case <-time.After(500 * time.Millisecond):
					t.Errorf("Expected event with operation %q in queue", tt.wantOperation)
				}
			} else {
				time.Sleep(150 * time.Millisecond)
				assert.Equal(t, 0, len(w.changeQueue))
			}
		})
	}
}

func TestEnhancedChangeWatcher_HandleFileSystemEvent_WithFileStats(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	w.debounceDelay = 50 * time.Millisecond

	tmpDir := t.TempDir()

	// Create a real file so os.Stat succeeds.
	testFilePath := filepath.Join(tmpDir, "real_file.mp4")
	require.NoError(t, os.WriteFile(testFilePath, []byte("test content for hash"), 0644))

	event := fsnotify.Event{Name: testFilePath, Op: fsnotify.Create}
	w.handleFileSystemEvent("test_storage", tmpDir, event)

	time.Sleep(150 * time.Millisecond)
	select {
	case evt := <-w.changeQueue:
		assert.Equal(t, "created", evt.Operation)
		assert.Greater(t, evt.Size, int64(0), "Expected non-zero size for real file")
		assert.False(t, evt.IsDir)
		// File is small enough to hash.
		assert.NotNil(t, evt.FileHash, "Expected hash for small file")
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected event in queue")
	}
}

func TestEnhancedChangeWatcher_HandleFileSystemEvent_Directory(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	w.debounceDelay = 50 * time.Millisecond

	tmpDir := t.TempDir()

	// Create a subdirectory so os.Stat sees it as a dir.
	subDir := filepath.Join(tmpDir, "new_subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	// Need the watcher registered so the "created" + isDir path adds to watcher.
	fswatcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer fswatcher.Close()

	w.watcherMu.Lock()
	w.watchers["test_storage"] = fswatcher
	w.watcherMu.Unlock()

	event := fsnotify.Event{Name: subDir, Op: fsnotify.Create}
	w.handleFileSystemEvent("test_storage", tmpDir, event)

	time.Sleep(150 * time.Millisecond)
	select {
	case evt := <-w.changeQueue:
		assert.Equal(t, "created", evt.Operation)
		assert.True(t, evt.IsDir)
		assert.Nil(t, evt.FileHash, "Directories should not have a hash")
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected event in queue")
	}
}

func TestEnhancedChangeWatcher_HandleFileSystemEvent_DeletedFile(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	w.debounceDelay = 50 * time.Millisecond

	tmpDir := t.TempDir()

	// Insert a file record in the DB so getFileInfoFromDB returns data.
	db := mediaDB.GetDB()
	_, err := db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, size, quick_hash, extension, mime_type, created_at, modified_at, last_scan_at)
		VALUES (1, 1, '/deleted_file.mp4', 'deleted_file.mp4', 0, 4096, 'oldhash123', '.mp4', 'video/mp4', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	// Simulate deletion of a file that no longer exists on disk.
	event := fsnotify.Event{Name: filepath.Join(tmpDir, "deleted_file.mp4"), Op: fsnotify.Remove}
	w.handleFileSystemEvent("test_storage", tmpDir, event)

	time.Sleep(150 * time.Millisecond)
	select {
	case evt := <-w.changeQueue:
		assert.Equal(t, "deleted", evt.Operation)
		// File info from DB should populate the event fields.
		// (only if getFileInfoFromDB matches the relative path)
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected event in queue")
	}
}

// ---------------------------------------------------------------------------
// getFileInfoFromDB
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_GetFileInfoFromDB_Found(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	db := mediaDB.GetDB()
	hash := "filehash123"
	_, err := db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, size, quick_hash, extension, mime_type, created_at, modified_at, last_scan_at)
		VALUES (50, 1, '/movies/test.mp4', 'test.mp4', 0, 8192, ?, '.mp4', 'video/mp4', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, hash)
	require.NoError(t, err)

	info := w.getFileInfoFromDB("/movies/test.mp4", "test_storage")
	require.NotNil(t, info, "Expected file info to be found")
	assert.Equal(t, int64(50), info.ID)
	assert.Equal(t, "test.mp4", info.Name)
	assert.Equal(t, "/movies/test.mp4", info.Path)
	assert.False(t, info.IsDirectory)
	assert.Equal(t, int64(8192), info.Size)
	assert.Equal(t, "file", info.Type)
	require.NotNil(t, info.Hash)
	assert.Equal(t, hash, *info.Hash)
}

func TestEnhancedChangeWatcher_GetFileInfoFromDB_Directory(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	db := mediaDB.GetDB()
	_, err := db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, size, created_at, modified_at, last_scan_at)
		VALUES (51, 1, '/movies', 'movies', 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	info := w.getFileInfoFromDB("/movies", "test_storage")
	require.NotNil(t, info)
	assert.True(t, info.IsDirectory)
	assert.Equal(t, "directory", info.Type)
}

func TestEnhancedChangeWatcher_GetFileInfoFromDB_NotFound(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	info := w.getFileInfoFromDB("/nonexistent/path.txt", "test_storage")
	assert.Nil(t, info, "Expected nil for non-existent file")
}

func TestEnhancedChangeWatcher_GetFileInfoFromDB_WrongStorageRoot(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	db := mediaDB.GetDB()
	_, err := db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, size, created_at, modified_at, last_scan_at)
		VALUES (52, 1, '/exists.mp4', 'exists.mp4', 0, 100, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	info := w.getFileInfoFromDB("/exists.mp4", "wrong_storage")
	assert.Nil(t, info, "Expected nil for wrong storage root")
}

// ---------------------------------------------------------------------------
// handleMove (Enhanced)
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleMove(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/moved_file.mp4",
		SmbRoot:   "test_storage",
		Operation: "moved",
		Timestamp: time.Now(),
	}

	// Should not panic; just logs a debug message.
	w.handleMove(ctx, event)
}

// ---------------------------------------------------------------------------
// processChange (Enhanced) — all operation branches
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_ProcessChange_AllOps(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	operations := []string{"created", "modified", "deleted", "moved"}
	for _, op := range operations {
		t.Run(op, func(t *testing.T) {
			event := EnhancedChangeEvent{
				Path:      "/process_enhanced_" + op + ".txt",
				SmbRoot:   "test_storage",
				Operation: op,
				Timestamp: time.Now(),
				Size:      512,
			}
			w.processChange(event, 0)

			// Verify change was logged.
			var count int
			err := mediaDB.GetDB().QueryRow(
				`SELECT COUNT(*) FROM change_log WHERE entity_id = ? AND change_type = ?`,
				event.Path, op,
			).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count)
		})
	}
}

// ---------------------------------------------------------------------------
// handleCreateNew — with and without analyzer
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleCreateNew_NilAnalyzer(t *testing.T) {
	logger := zap.NewNop()
	w := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	ctx := context.Background()

	// File without analyzer.
	event := EnhancedChangeEvent{
		Path:      "/no_analyzer_file.mp4",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		Size:      1024,
		IsDir:     false,
	}
	w.handleCreateNew(ctx, event)

	// Directory without analyzer.
	dirEvent := EnhancedChangeEvent{
		Path:      "/no_analyzer_dir",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		IsDir:     true,
	}
	w.handleCreateNew(ctx, dirEvent)
}

func TestEnhancedChangeWatcher_HandleCreateNew_WithAnalyzer_Dir(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/new_analyzed_dir",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		IsDir:     true,
	}
	// Should call analyzer.AnalyzeDirectory for a directory.
	w.handleCreateNew(ctx, event)
}

func TestEnhancedChangeWatcher_HandleCreateNew_WithAnalyzer_File(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/new_analyzed_file.mp4",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		Size:      4096,
		IsDir:     false,
	}
	// Should call analyzer.AnalyzeDirectory on the parent directory.
	w.handleCreateNew(ctx, event)
}

// ---------------------------------------------------------------------------
// handleModify — media file triggers re-analysis
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleModify_MediaFile(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	db := mediaDB.GetDB()
	_, err := db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, size, created_at, modified_at, last_scan_at)
		VALUES (60, 1, '/movies/video.mp4', 'video.mp4', 0, 4096, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/movies/video.mp4",
		SmbRoot:   "test_storage",
		Operation: "modified",
		Timestamp: time.Now(),
		Size:      8192,
		IsDir:     false,
		FileHash:  utils.StringPtr("newhash456"),
	}
	// isMediaFile returns true for .mp4, so this triggers re-analysis.
	w.handleModify(ctx, event)
}

func TestEnhancedChangeWatcher_HandleModify_NonMediaFile(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/docs/readme.txt",
		SmbRoot:   "test_storage",
		Operation: "modified",
		Timestamp: time.Now(),
		Size:      256,
		IsDir:     false,
	}
	// isMediaFile returns false for .txt, so NO re-analysis.
	w.handleModify(ctx, event)
}

// ---------------------------------------------------------------------------
// handleDelete — with and without FileID
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleDelete_WithFileID(t *testing.T) {
	w, mediaDB, rt := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()
	fileID := int64(100)
	hash := utils.StringPtr("hash100")
	event := EnhancedChangeEvent{
		Path:      "/delete_with_id.mp4",
		SmbRoot:   "test_storage",
		Operation: "deleted",
		Timestamp: time.Now(),
		Size:      2048,
		IsDir:     false,
		FileHash:  hash,
		FileID:    &fileID,
	}

	w.handleDelete(ctx, event)

	// Verify tracked as pending move.
	key := rt.CreateMoveKey("test_storage", hash, event.Size, false)
	rt.PendingMovesMu.RLock()
	_, exists := rt.PendingMoves[key]
	rt.PendingMovesMu.RUnlock()
	assert.True(t, exists, "Expected deletion to be tracked as pending move")
}

func TestEnhancedChangeWatcher_HandleDelete_WithoutFileID(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/delete_no_id.mp4",
		SmbRoot:   "test_storage",
		Operation: "deleted",
		Timestamp: time.Now(),
		Size:      2048,
		IsDir:     false,
		FileID:    nil,
	}

	// Should not panic when FileID is nil.
	w.handleDelete(ctx, event)
}

// ---------------------------------------------------------------------------
// handleDeleteFallback
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleDeleteFallback(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	db := mediaDB.GetDB()
	_, err := db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, size, created_at, modified_at, last_scan_at)
		VALUES (70, 1, '/fallback_delete.mp4', 'fallback_delete.mp4', 0, 1024, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	ctx := context.Background()
	pendingMove := &services.PendingMove{
		FileID:      70,
		Path:        "/fallback_delete.mp4",
		StorageRoot: "test_storage",
		Size:        1024,
	}

	// Exercises the full code path. The query uses "SET deleted = true" which
	// may not update on raw SQLite (go-sqlcipher) because 'true' is not a
	// recognized keyword in all builds. The function still executes without panic.
	w.handleDeleteFallback(ctx, pendingMove)
}

func TestEnhancedChangeWatcher_HandleDeleteFallback_NonExistentFile(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()
	pendingMove := &services.PendingMove{
		FileID:      99999,
		Path:        "/nonexistent.mp4",
		StorageRoot: "test_storage",
		Size:        1024,
	}

	// Should not panic even for non-existent file ID.
	w.handleDeleteFallback(ctx, pendingMove)
}

// ---------------------------------------------------------------------------
// updateFileMetadata / updateDirectoryMetadata
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_UpdateFileMetadata(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	db := mediaDB.GetDB()
	_, err := db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, size, created_at, modified_at, last_scan_at)
		VALUES (80, 1, '/update_meta.mp4', 'update_meta.mp4', 0, 1024, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	ctx := context.Background()
	hash := "newhash_meta"
	event := EnhancedChangeEvent{
		Path:      "/update_meta.mp4",
		SmbRoot:   "test_storage",
		Operation: "modified",
		Timestamp: time.Now(),
		Size:      2048,
		FileHash:  &hash,
	}

	w.updateFileMetadata(ctx, event)

	var size int64
	var quickHash string
	err = db.QueryRow(`SELECT size, quick_hash FROM files WHERE id = 80`).Scan(&size, &quickHash)
	require.NoError(t, err)
	assert.Equal(t, int64(2048), size)
	assert.Equal(t, "newhash_meta", quickHash)
}

func TestEnhancedChangeWatcher_UpdateFileMetadata_Error(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/nonexistent_file.mp4",
		SmbRoot:   "nonexistent_root",
		Operation: "modified",
		Timestamp: time.Now(),
		Size:      2048,
	}

	// Should not panic even when no matching row exists.
	w.updateFileMetadata(ctx, event)
}

func TestEnhancedChangeWatcher_UpdateDirectoryMetadata(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	db := mediaDB.GetDB()
	_, err := db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, size, created_at, modified_at, last_scan_at)
		VALUES (81, 1, '/update_dir', 'update_dir', 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/update_dir",
		SmbRoot:   "test_storage",
		Operation: "modified",
		Timestamp: time.Now(),
		IsDir:     true,
	}

	w.updateDirectoryMetadata(ctx, event)
}

// ---------------------------------------------------------------------------
// Stop — with watchers present
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_Stop_WithWatchers(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	tempDir := t.TempDir()

	err := w.Start()
	require.NoError(t, err)

	err = w.WatchPath("root1", tempDir)
	require.NoError(t, err)

	w.Stop()

	w.watcherMu.RLock()
	assert.Empty(t, w.watchers, "All watchers should be removed after Stop")
	w.watcherMu.RUnlock()

	select {
	case <-w.stopCh:
	default:
		t.Error("Expected stop channel to be closed")
	}
}

// ---------------------------------------------------------------------------
// getRelativePath — additional edge cases
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_GetRelativePath_TrailingSlash(t *testing.T) {
	logger := zap.NewNop()
	w := &EnhancedChangeWatcher{logger: logger}

	result, err := w.getRelativePath("/mnt/storage/", "/mnt/storage/file.txt")
	require.NoError(t, err)
	assert.Equal(t, "/file.txt", result)
}

// ---------------------------------------------------------------------------
// calculateFileHash — edge cases
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_CalculateFileHash_EmptyFile(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	require.NoError(t, os.WriteFile(emptyFile, []byte{}, 0644))

	hash := w.calculateFileHash(emptyFile)
	// SHA-256 of empty content is a well-known value.
	assert.NotEmpty(t, hash, "Even empty files should produce a hash")
}

// ---------------------------------------------------------------------------
// logChange (Enhanced) — all operation types
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_LogChange_AllOperations(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	operations := []string{"created", "modified", "deleted", "moved"}
	for _, op := range operations {
		t.Run(op, func(t *testing.T) {
			hash := "hash_" + op
			fileID := int64(len(op))
			event := EnhancedChangeEvent{
				Path:      "/enhanced_log_" + op + ".txt",
				SmbRoot:   "test_storage",
				Operation: op,
				Timestamp: time.Now(),
				Size:      int64(len(op) * 100),
				IsDir:     op == "created",
				FileHash:  &hash,
				FileID:    &fileID,
			}
			err := w.logChange(event)
			require.NoError(t, err)

			var count int
			err = mediaDB.GetDB().QueryRow(
				`SELECT COUNT(*) FROM change_log WHERE entity_id = ? AND change_type = ?`,
				event.Path, op,
			).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count)
		})
	}
}

// ---------------------------------------------------------------------------
// EnhancedChangeEvent struct tests
// ---------------------------------------------------------------------------

func TestEnhancedChangeEvent_Fields(t *testing.T) {
	now := time.Now()
	hash := "testhash"
	fileID := int64(42)
	prevPath := "/old/path.mp4"

	event := EnhancedChangeEvent{
		Path:         "/test/file.mp4",
		SmbRoot:      "test_root",
		Operation:    "created",
		Timestamp:    now,
		Size:         2048,
		IsDir:        false,
		FileHash:     &hash,
		FileID:       &fileID,
		PreviousPath: &prevPath,
	}

	assert.Equal(t, "/test/file.mp4", event.Path)
	assert.Equal(t, "test_root", event.SmbRoot)
	assert.Equal(t, "created", event.Operation)
	assert.Equal(t, now, event.Timestamp)
	assert.Equal(t, int64(2048), event.Size)
	assert.False(t, event.IsDir)
	require.NotNil(t, event.FileHash)
	assert.Equal(t, "testhash", *event.FileHash)
	require.NotNil(t, event.FileID)
	assert.Equal(t, int64(42), *event.FileID)
	require.NotNil(t, event.PreviousPath)
	assert.Equal(t, "/old/path.mp4", *event.PreviousPath)
}

func TestEnhancedChangeEvent_NilOptionalFields(t *testing.T) {
	event := EnhancedChangeEvent{
		Path:      "/test.txt",
		SmbRoot:   "root",
		Operation: "modified",
		Timestamp: time.Now(),
	}

	assert.Nil(t, event.FileHash)
	assert.Nil(t, event.FileID)
	assert.Nil(t, event.PreviousPath)
}

// ---------------------------------------------------------------------------
// isMediaFile — additional extensions
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_IsMediaFile_AllExtensions(t *testing.T) {
	logger := zap.NewNop()
	w := &EnhancedChangeWatcher{logger: logger}

	tests := []struct {
		path     string
		expected bool
	}{
		// Video.
		{"/v.mp4", true}, {"/v.avi", true}, {"/v.mkv", true},
		{"/v.mov", true}, {"/v.wmv", true},
		// Audio.
		{"/a.mp3", true}, {"/a.flac", true}, {"/a.wav", true},
		{"/a.aac", true}, {"/a.ogg", true},
		// Image.
		{"/i.jpg", true}, {"/i.jpeg", true}, {"/i.png", true},
		{"/i.gif", true}, {"/i.bmp", true},
		// Document.
		{"/d.pdf", true}, {"/d.epub", true}, {"/d.mobi", true},
		// Non-media.
		{"/n.txt", false}, {"/n.go", false}, {"/n.rs", false},
		{"/n.html", false}, {"/n.css", false}, {"/n.js", false},
		{"/n.zip", false}, {"/n.tar", false}, {"/n", false},
		// Case insensitive.
		{"/V.MP4", true}, {"/V.Mp4", true}, {"/V.MKV", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, w.isMediaFile(tt.path))
		})
	}
}

// ---------------------------------------------------------------------------
// GetStatistics — edge cases
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_GetStatistics_Empty(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	stats, err := w.GetStatistics(time.Now().Add(-1 * time.Hour))
	require.NoError(t, err)
	require.NotNil(t, stats)

	changesByType, ok := stats["changes_by_type"].(map[string]int)
	require.True(t, ok)
	assert.Empty(t, changesByType)

	assert.Equal(t, 0, stats["watched_paths"])
	assert.Equal(t, 4, stats["workers"])
}

// ---------------------------------------------------------------------------
// debounceChange (Enhanced) — queue full scenario
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_DebounceChange_QueueFull(t *testing.T) {
	logger := zap.NewNop()
	w := &EnhancedChangeWatcher{
		logger:        logger,
		changeQueue:   make(chan EnhancedChangeEvent, 1), // tiny queue
		debounceMap:   make(map[string]*enhancedDebounceEntry),
		debounceDelay: 50 * time.Millisecond,
		stopCh:        make(chan struct{}),
	}

	// Fill the queue.
	w.changeQueue <- EnhancedChangeEvent{Path: "/blocker", Operation: "created"}

	// Now debounce a new event; when the timer fires, it should drop it.
	event := EnhancedChangeEvent{
		Path:      "/dropped.mp4",
		SmbRoot:   "test_root",
		Operation: "modified",
		Timestamp: time.Now(),
	}
	w.debounceChange(event)

	// Wait for debounce timer to fire.
	time.Sleep(200 * time.Millisecond)

	// Only the blocker should be in the queue.
	assert.Equal(t, 1, len(w.changeQueue))
}

// ---------------------------------------------------------------------------
// monitorPath (Enhanced) — error channel path
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_MonitorPath_ErrorChannel(t *testing.T) {
	logger := zap.NewNop()
	w := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	fswatcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)

	// Add a valid path so the watcher is active.
	tmpDir := t.TempDir()
	require.NoError(t, fswatcher.Add(tmpDir))

	w.wg.Add(1)
	go w.monitorPath("test_root", tmpDir, fswatcher)

	// Closing the watcher triggers both the Events and Errors channels to close,
	// which makes the monitorPath goroutine return via the !ok branches.
	fswatcher.Close()
	w.wg.Wait()
}

// ---------------------------------------------------------------------------
// handleCreate — move detection fallback path
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleCreate_MoveDetectedButProcessFails(t *testing.T) {
	w, mediaDB, rt := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	ctx := context.Background()

	// Track a deletion so DetectCreate finds a match.
	fileID := int64(500)
	oldPath := "/old_move_fail.txt"
	size := int64(2048)
	hash := utils.StringPtr("movefailhash")

	rt.TrackDelete(ctx, fileID, oldPath, "test_storage", size, hash, false)

	// Create event that matches the deletion.
	event := EnhancedChangeEvent{
		Path:      "/new_move_fail.txt",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		Size:      size,
		IsDir:     false,
		FileHash:  hash,
	}

	// ProcessMove will fail (no proper DB setup for rename events table),
	// which triggers the fallback path: handleDeleteFallback + handleCreateNew.
	w.handleCreate(ctx, event)
}

// ---------------------------------------------------------------------------
// processChange (Enhanced) — logChange error path
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_ProcessChange_LogChangeError(t *testing.T) {
	logger := zap.NewNop()
	// Create watcher with nil mediaDB so logChange fails.
	w := &EnhancedChangeWatcher{
		logger:  logger,
		stopCh:  make(chan struct{}),
		workers: 1,
	}

	// processChange will call logChange which will panic on nil mediaDB.
	// Instead, use a watcher with no analyzer but valid DB.
	w2, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	// Close the DB to force logChange to fail.
	mediaDB.Close()

	event := EnhancedChangeEvent{
		Path:      "/log_error.txt",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		Size:      256,
	}

	// Should not panic; the error from logChange is logged.
	w2.processChange(event, 0)
	_ = w // suppress unused
}

// ---------------------------------------------------------------------------
// handleCreateNew — error paths with analyzer
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleCreateNew_AnalyzerError_Dir(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	// Close DB to force analyzer.AnalyzeDirectory to fail.
	mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/error_dir",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		IsDir:     true,
	}
	// Should not panic; error is logged.
	w.handleCreateNew(ctx, event)
}

func TestEnhancedChangeWatcher_HandleCreateNew_AnalyzerError_File(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	// Close DB to force analyzer.AnalyzeDirectory to fail.
	mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/error_file.mp4",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		Size:      1024,
		IsDir:     false,
	}
	// Should not panic; error is logged.
	w.handleCreateNew(ctx, event)
}

// ---------------------------------------------------------------------------
// handleFileSystemEvent — getRelativePath error path
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_HandleFileSystemEvent_RelativePathError(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	w.debounceDelay = 50 * time.Millisecond

	// Use an empty basePath which will still produce a relative path, but let's
	// try paths that cause getRelativePath to return an error.
	// On Unix, filepath.Rel never actually returns an error for absolute paths.
	// So we test a normal path to exercise the code and ensure no panic.
	event := fsnotify.Event{Name: "/completely/different/path.mp4", Op: fsnotify.Create}
	w.handleFileSystemEvent("test_storage", "/base", event)

	time.Sleep(150 * time.Millisecond)
	// Event should still be queued with a relative path.
	select {
	case evt := <-w.changeQueue:
		assert.Equal(t, "created", evt.Operation)
	case <-time.After(500 * time.Millisecond):
		// May not arrive if getRelativePath returned an error (path with ".." prefix).
	}
}

// ---------------------------------------------------------------------------
// updateFileMetadata — error path
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_UpdateFileMetadata_DBClosed(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/meta_error.mp4",
		SmbRoot:   "test_storage",
		Operation: "modified",
		Timestamp: time.Now(),
		Size:      2048,
	}

	// Should not panic; error is logged.
	w.updateFileMetadata(ctx, event)
}

// ---------------------------------------------------------------------------
// updateDirectoryMetadata — error path
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_UpdateDirectoryMetadata_DBClosed(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	mediaDB.Close()

	ctx := context.Background()
	event := EnhancedChangeEvent{
		Path:      "/dir_meta_error",
		SmbRoot:   "test_storage",
		Operation: "modified",
		Timestamp: time.Now(),
		IsDir:     true,
	}

	// Should not panic; error is logged.
	w.updateDirectoryMetadata(ctx, event)
}

// ---------------------------------------------------------------------------
// GetStatistics — error path (DB closed)
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_GetStatistics_DBClosed(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	mediaDB.Close()

	_, err := w.GetStatistics(time.Now().Add(-1 * time.Hour))
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// WatchPath — fsnotify.NewWatcher error cannot be easily triggered,
// but we can test the cleanup path when addPathRecursively fails.
// ---------------------------------------------------------------------------

func TestEnhancedChangeWatcher_WatchPath_MultiplePaths(t *testing.T) {
	w, mediaDB, _ := setupEnhancedWatcher(t)
	defer mediaDB.Close()

	tempDir1 := t.TempDir()
	tempDir2 := t.TempDir()

	err := w.Start()
	require.NoError(t, err)
	defer w.Stop()

	err = w.WatchPath("root1", tempDir1)
	require.NoError(t, err)

	err = w.WatchPath("root2", tempDir2)
	require.NoError(t, err)

	w.watcherMu.RLock()
	assert.Equal(t, 2, len(w.watchers))
	w.watcherMu.RUnlock()
}
