package realtime

import (
	"catalog-api/internal/media/analyzer"
	"catalog-api/internal/media/database"
	"catalog-api/internal/media/models"
	"catalog-api/internal/services"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// EnhancedChangeWatcher monitors file system changes with intelligent rename detection
type EnhancedChangeWatcher struct {
	mediaDB        *database.MediaDatabase
	analyzer       *analyzer.MediaAnalyzer
	renameTracker  *services.RenameTracker
	logger         *zap.Logger
	watchers       map[string]*fsnotify.Watcher
	watcherMu      sync.RWMutex
	changeQueue    chan EnhancedChangeEvent
	workers        int
	stopCh         chan struct{}
	wg             sync.WaitGroup
	debounceMap    map[string]*time.Timer
	debounceMu     sync.Mutex
	debounceDelay  time.Duration
}

// EnhancedChangeEvent represents a file system change with additional metadata
type EnhancedChangeEvent struct {
	Path         string
	SmbRoot      string
	Operation    string // created, modified, deleted, moved
	Timestamp    time.Time
	Size         int64
	IsDir        bool
	FileHash     *string
	FileID       *int64
	PreviousPath *string // for move operations
}

// NewEnhancedChangeWatcher creates a new enhanced change watcher
func NewEnhancedChangeWatcher(mediaDB *database.MediaDatabase, analyzer *analyzer.MediaAnalyzer, renameTracker *services.RenameTracker, logger *zap.Logger) *EnhancedChangeWatcher {
	return &EnhancedChangeWatcher{
		mediaDB:       mediaDB,
		analyzer:      analyzer,
		renameTracker: renameTracker,
		logger:        logger,
		watchers:      make(map[string]*fsnotify.Watcher),
		changeQueue:   make(chan EnhancedChangeEvent, 10000),
		workers:       4, // Increased workers for better performance
		debounceMap:   make(map[string]*time.Timer),
		debounceDelay: 2 * time.Second,
		stopCh:        make(chan struct{}),
	}
}

// Start starts the enhanced change watcher
func (w *EnhancedChangeWatcher) Start() error {
	w.logger.Info("Starting enhanced change watcher", zap.Int("workers", w.workers))

	// Start worker goroutines
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.changeWorker(i)
	}

	return nil
}

// Stop stops the enhanced change watcher
func (w *EnhancedChangeWatcher) Stop() {
	w.logger.Info("Stopping enhanced change watcher")

	// Stop all file watchers
	w.watcherMu.Lock()
	for path, watcher := range w.watchers {
		watcher.Close()
		w.logger.Debug("Closed watcher", zap.String("path", path))
	}
	w.watchers = make(map[string]*fsnotify.Watcher)
	w.watcherMu.Unlock()

	// Stop workers
	close(w.stopCh)
	w.wg.Wait()

	w.logger.Info("Enhanced change watcher stopped")
}

// WatchPath adds a new path to watch
func (w *EnhancedChangeWatcher) WatchPath(smbRoot, localMountPath string) error {
	w.watcherMu.Lock()
	defer w.watcherMu.Unlock()

	// Check if already watching
	if _, exists := w.watchers[smbRoot]; exists {
		return nil
	}

	// Create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Add path to watcher recursively
	err = w.addPathRecursively(watcher, localMountPath)
	if err != nil {
		watcher.Close()
		return err
	}

	w.watchers[smbRoot] = watcher

	// Start monitoring goroutine
	w.wg.Add(1)
	go w.monitorPath(smbRoot, localMountPath, watcher)

	w.logger.Info("Started watching path",
		zap.String("smb_root", smbRoot),
		zap.String("local_path", localMountPath))

	return nil
}

// UnwatchPath removes a path from watching
func (w *EnhancedChangeWatcher) UnwatchPath(smbRoot string) {
	w.watcherMu.Lock()
	defer w.watcherMu.Unlock()

	if watcher, exists := w.watchers[smbRoot]; exists {
		watcher.Close()
		delete(w.watchers, smbRoot)
		w.logger.Info("Stopped watching path", zap.String("smb_root", smbRoot))
	}
}

// addPathRecursively adds a path and all its subdirectories to the watcher
func (w *EnhancedChangeWatcher) addPathRecursively(watcher *fsnotify.Watcher, rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip problematic paths
		}

		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
}

// monitorPath monitors a specific path for changes
func (w *EnhancedChangeWatcher) monitorPath(smbRoot, localPath string, watcher *fsnotify.Watcher) {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopCh:
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			w.handleFileSystemEvent(smbRoot, localPath, event)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			w.logger.Error("File watcher error",
				zap.String("smb_root", smbRoot),
				zap.Error(err))
		}
	}
}

// handleFileSystemEvent processes a file system event with enhanced logic
func (w *EnhancedChangeWatcher) handleFileSystemEvent(smbRoot, localPath string, event fsnotify.Event) {
	// Convert local path to relative path within the storage root
	relativePath, err := w.getRelativePath(localPath, event.Name)
	if err != nil {
		w.logger.Warn("Failed to get relative path", zap.Error(err))
		return
	}

	// Determine operation type
	var operation string
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		operation = "created"
	case event.Op&fsnotify.Write == fsnotify.Write:
		operation = "modified"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		operation = "deleted"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		operation = "moved"
	default:
		return // Ignore other events
	}

	// Get file info and metadata
	var size int64
	var isDir bool
	var fileHash *string
	var fileID *int64

	if operation != "deleted" {
		// Get file stats
		info, err := os.Stat(event.Name)
		if err == nil {
			size = info.Size()
			isDir = info.IsDir()

			// Calculate hash for files (not directories)
			if !isDir && size > 0 && size < 100*1024*1024 { // Only hash files < 100MB
				if hash := w.calculateFileHash(event.Name); hash != "" {
					fileHash = &hash
				}
			}
		}
	} else {
		// For deleted files, try to get info from database
		fileInfo := w.getFileInfoFromDB(relativePath, smbRoot)
		if fileInfo != nil {
			size = fileInfo.Size
			isDir = fileInfo.IsDirectory
			fileHash = fileInfo.QuickHash
			fileID = &fileInfo.ID
		}
	}

	// Handle directory creation by adding to watcher
	if operation == "created" && isDir {
		w.watcherMu.Lock()
		if watcher, exists := w.watchers[smbRoot]; exists {
			watcher.Add(event.Name)
		}
		w.watcherMu.Unlock()
	}

	changeEvent := EnhancedChangeEvent{
		Path:      relativePath,
		SmbRoot:   smbRoot,
		Operation: operation,
		Timestamp: time.Now(),
		Size:      size,
		IsDir:     isDir,
		FileHash:  fileHash,
		FileID:    fileID,
	}

	// Debounce changes to avoid processing rapid consecutive changes
	w.debounceChange(changeEvent)
}

// getRelativePath converts absolute path to relative path within storage root
func (w *EnhancedChangeWatcher) getRelativePath(basePath, fullPath string) (string, error) {
	relPath, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return "", err
	}

	// Normalize path separators for cross-platform compatibility
	relPath = filepath.ToSlash(relPath)

	// Ensure path starts with /
	if !strings.HasPrefix(relPath, "/") {
		relPath = "/" + relPath
	}

	return relPath, nil
}

// getFileInfoFromDB retrieves file information from database
func (w *EnhancedChangeWatcher) getFileInfoFromDB(path, smbRoot string) *models.File {
	query := `
		SELECT f.id, f.size, f.is_directory, f.quick_hash
		FROM files f
		JOIN storage_roots sr ON f.storage_root_id = sr.id
		WHERE f.path = ? AND sr.name = ?`

	var file models.File
	err := w.mediaDB.GetDB().QueryRow(query, path, smbRoot).Scan(
		&file.ID, &file.Size, &file.IsDirectory, &file.QuickHash)

	if err != nil {
		return nil
	}

	return &file
}

// calculateFileHash calculates MD5 hash of a file
func (w *EnhancedChangeWatcher) calculateFileHash(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// debounceChange debounces file changes to avoid excessive processing
func (w *EnhancedChangeWatcher) debounceChange(event EnhancedChangeEvent) {
	w.debounceMu.Lock()
	defer w.debounceMu.Unlock()

	key := event.SmbRoot + ":" + event.Path

	// Cancel existing timer
	if timer, exists := w.debounceMap[key]; exists {
		timer.Stop()
	}

	// Create new timer
	w.debounceMap[key] = time.AfterFunc(w.debounceDelay, func() {
		w.debounceMu.Lock()
		delete(w.debounceMap, key)
		w.debounceMu.Unlock()

		// Send to processing queue
		select {
		case w.changeQueue <- event:
		default:
			w.logger.Warn("Change queue full, dropping event",
				zap.String("path", event.Path),
				zap.String("operation", event.Operation))
		}
	})
}

// changeWorker processes change events
func (w *EnhancedChangeWatcher) changeWorker(workerID int) {
	defer w.wg.Done()

	w.logger.Info("Enhanced change worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-w.stopCh:
			return

		case event := <-w.changeQueue:
			w.processChange(event, workerID)
		}
	}
}

// processChange processes a single change event with rename detection
func (w *EnhancedChangeWatcher) processChange(event EnhancedChangeEvent, workerID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	w.logger.Debug("Processing enhanced change event",
		zap.Int("worker_id", workerID),
		zap.String("path", event.Path),
		zap.String("operation", event.Operation),
		zap.String("smb_root", event.SmbRoot))

	// Log the change
	if err := w.logChange(event); err != nil {
		w.logger.Error("Failed to log change", zap.Error(err))
	}

	switch event.Operation {
	case "created":
		w.handleCreate(ctx, event)
	case "modified":
		w.handleModify(ctx, event)
	case "deleted":
		w.handleDelete(ctx, event)
	case "moved":
		w.handleMove(ctx, event)
	}
}

// handleCreate handles file/directory creation with rename detection
func (w *EnhancedChangeWatcher) handleCreate(ctx context.Context, event EnhancedChangeEvent) {
	// Check if this creation is actually part of a move operation
	if pendingMove, isMove := w.renameTracker.DetectCreate(ctx, event.Path, event.SmbRoot, event.Size, event.FileHash, event.IsDir); isMove {
		// This is a move operation, not a new file
		err := w.renameTracker.ProcessMove(ctx, pendingMove, event.Path)
		if err != nil {
			w.logger.Error("Failed to process move operation",
				zap.String("old_path", pendingMove.Path),
				zap.String("new_path", event.Path),
				zap.Error(err))

			// Fall back to treating as delete + create
			w.handleDeleteFallback(ctx, pendingMove)
			w.handleCreateNew(ctx, event)
		} else {
			w.logger.Info("Successfully processed rename operation",
				zap.String("old_path", pendingMove.Path),
				zap.String("new_path", event.Path),
				zap.Bool("is_directory", event.IsDir))

			// No need for rescanning - just metadata update
			w.updateFileMetadata(ctx, event)
		}
		return
	}

	// This is a genuinely new file/directory
	w.handleCreateNew(ctx, event)
}

// handleCreateNew handles creation of a genuinely new file/directory
func (w *EnhancedChangeWatcher) handleCreateNew(ctx context.Context, event EnhancedChangeEvent) {
	if event.IsDir {
		// Directory creation - trigger directory analysis
		err := w.analyzer.AnalyzeDirectory(ctx, event.Path, event.SmbRoot, 7) // High priority
		if err != nil {
			w.logger.Error("Failed to queue directory analysis",
				zap.String("path", event.Path),
				zap.Error(err))
		}
	} else {
		// File creation - analyze parent directory
		parentDir := filepath.Dir(event.Path)
		err := w.analyzer.AnalyzeDirectory(ctx, parentDir, event.SmbRoot, 6) // Medium-high priority
		if err != nil {
			w.logger.Error("Failed to queue directory analysis for new file",
				zap.String("file", event.Path),
				zap.String("parent_dir", parentDir),
				zap.Error(err))
		}
	}
}

// handleModify handles file modification
func (w *EnhancedChangeWatcher) handleModify(ctx context.Context, event EnhancedChangeEvent) {
	if event.IsDir {
		// Directory modification - usually metadata changes
		w.updateDirectoryMetadata(ctx, event)
	} else {
		// File modification - update file metadata and check for changes
		w.updateFileMetadata(ctx, event)

		// If this is a media file, might need re-analysis
		if w.isMediaFile(event.Path) {
			parentDir := filepath.Dir(event.Path)
			w.analyzer.AnalyzeDirectory(ctx, parentDir, event.SmbRoot, 5) // Medium priority
		}
	}
}

// handleDelete handles file/directory deletion with move tracking
func (w *EnhancedChangeWatcher) handleDelete(ctx context.Context, event EnhancedChangeEvent) {
	// Track this deletion for potential move detection
	if event.FileID != nil {
		w.renameTracker.TrackDelete(ctx, *event.FileID, event.Path, event.SmbRoot, event.Size, event.FileHash, event.IsDir)
	}

	// Don't immediately delete from database - wait for move window to expire
	// The rename tracker's cleanup will handle actual deletions
}

// handleDeleteFallback handles deletion when move processing fails
func (w *EnhancedChangeWatcher) handleDeleteFallback(ctx context.Context, pendingMove *services.PendingMove) {
	// Mark files as deleted in database
	query := `
		UPDATE files
		SET deleted = true, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := w.mediaDB.GetDB().ExecContext(ctx, query, pendingMove.FileID)
	if err != nil {
		w.logger.Error("Failed to mark file as deleted",
			zap.String("path", pendingMove.Path),
			zap.Error(err))
	}
}

// handleMove handles explicit move operations (rare with fsnotify)
func (w *EnhancedChangeWatcher) handleMove(ctx context.Context, event EnhancedChangeEvent) {
	// This is handled by the create/delete pair in most file systems
	w.logger.Debug("Explicit move event received", zap.String("path", event.Path))
}

// updateFileMetadata updates file metadata without rescanning
func (w *EnhancedChangeWatcher) updateFileMetadata(ctx context.Context, event EnhancedChangeEvent) {
	query := `
		UPDATE files
		SET last_scan_at = CURRENT_TIMESTAMP, modified_at = ?, size = ?, quick_hash = ?
		WHERE path = ? AND storage_root_id = (SELECT id FROM storage_roots WHERE name = ?)`

	_, err := w.mediaDB.GetDB().ExecContext(ctx, query, event.Timestamp, event.Size, event.FileHash, event.Path, event.SmbRoot)
	if err != nil {
		w.logger.Error("Failed to update file metadata",
			zap.String("path", event.Path),
			zap.Error(err))
	}
}

// updateDirectoryMetadata updates directory metadata
func (w *EnhancedChangeWatcher) updateDirectoryMetadata(ctx context.Context, event EnhancedChangeEvent) {
	query := `
		UPDATE files
		SET last_scan_at = CURRENT_TIMESTAMP, modified_at = ?
		WHERE path = ? AND storage_root_id = (SELECT id FROM storage_roots WHERE name = ?) AND is_directory = true`

	_, err := w.mediaDB.GetDB().ExecContext(ctx, query, event.Timestamp, event.Path, event.SmbRoot)
	if err != nil {
		w.logger.Error("Failed to update directory metadata",
			zap.String("path", event.Path),
			zap.Error(err))
	}
}

// isMediaFile checks if a file is a media file based on extension
func (w *EnhancedChangeWatcher) isMediaFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	mediaExtensions := map[string]bool{
		".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true,
		".mp3": true, ".flac": true, ".wav": true, ".aac": true, ".ogg": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
		".pdf": true, ".epub": true, ".mobi": true,
	}
	return mediaExtensions[ext]
}

// logChange logs the change to the database
func (w *EnhancedChangeWatcher) logChange(event EnhancedChangeEvent) error {
	eventDataJSON, _ := json.Marshal(map[string]interface{}{
		"path":      event.Path,
		"operation": event.Operation,
		"timestamp": event.Timestamp,
		"size":      event.Size,
		"is_dir":    event.IsDir,
		"file_hash": event.FileHash,
		"file_id":   event.FileID,
	})

	query := `
		INSERT INTO change_log (entity_type, entity_id, change_type, new_data, detected_at, processed)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := w.mediaDB.GetDB().Exec(query,
		"file", event.Path, event.Operation, string(eventDataJSON), event.Timestamp, false)

	return err
}

// GetStatistics returns statistics about the enhanced watcher
func (w *EnhancedChangeWatcher) GetStatistics(since time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get change statistics
	query := `
		SELECT change_type, COUNT(*) as count
		FROM change_log
		WHERE detected_at >= ?
		GROUP BY change_type`

	rows, err := w.mediaDB.GetDB().Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changeTypes := make(map[string]int)
	for rows.Next() {
		var changeType string
		var count int
		if err := rows.Scan(&changeType, &count); err != nil {
			continue
		}
		changeTypes[changeType] = count
	}
	stats["changes_by_type"] = changeTypes

	// Get rename tracker statistics
	renameStats := w.renameTracker.GetStatistics()
	stats["rename_tracking"] = renameStats

	// Count currently watched paths
	w.watcherMu.RLock()
	stats["watched_paths"] = len(w.watchers)
	w.watcherMu.RUnlock()

	// Queue statistics
	stats["queue_length"] = len(w.changeQueue)
	stats["workers"] = w.workers

	return stats, nil
}