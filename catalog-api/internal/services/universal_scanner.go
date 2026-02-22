package services

import (
	"catalogizer/database"
	"catalogizer/filesystem"
	"catalogizer/models"
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

// UniversalScanner handles file system scanning across all supported protocols
type UniversalScanner struct {
	db                 *database.DB
	logger             *zap.Logger
	renameTracker      *UniversalRenameTracker
	clientFactory      filesystem.ClientFactory
	aggregationService *AggregationService
	scanQueue          chan ScanJob
	workers            int
	maxConcurrentScans int
	scanSem            *semaphore.Weighted
	stopCh             chan struct{}
	wg                 sync.WaitGroup
	protocolScannersMu sync.RWMutex
	protocolScanners   map[string]ProtocolScanner
	activeScansMu      sync.RWMutex
	activeScans        map[string]*ScanStatus
}

// ScanJob represents a scan operation for any protocol
type ScanJob struct {
	ID              string
	StorageRoot     *models.StorageRoot
	Path            string
	Priority        int
	ScanType        string // full, incremental, verify
	MaxDepth        int
	IncludePatterns []string
	ExcludePatterns []string
	Context         context.Context
}

// ScanStatus tracks the status of an active scan
type ScanStatus struct {
	JobID           string
	StorageRootName string
	Protocol        string
	StartTime       time.Time
	CurrentPath     string
	FilesProcessed  int64
	FilesFound      int64
	FilesUpdated    int64
	FilesDeleted    int64
	ErrorCount      int64
	Status          string // running, completed, failed, cancelled
	mu              sync.RWMutex
}

// ProtocolScanner defines protocol-specific scanning behavior
type ProtocolScanner interface {
	// ScanPath performs a scan of the specified path
	ScanPath(ctx context.Context, client filesystem.FileSystemClient, job ScanJob, status *ScanStatus) error

	// GetScanStrategy returns the optimal scanning strategy for this protocol
	GetScanStrategy() ScanStrategy

	// SupportsIncrementalScan indicates if the protocol supports incremental scanning
	SupportsIncrementalScan() bool

	// GetOptimalBatchSize returns the optimal batch size for database operations
	GetOptimalBatchSize() int
}

// ScanStrategy defines how scanning should be performed
type ScanStrategy struct {
	UseRecursiveListing     bool
	BatchSize               int
	ParallelDirectories     bool
	ChecksumCalculation     bool
	MetadataExtraction      bool
	RealTimeChangeDetection bool
}

// NewUniversalScanner creates a new universal file system scanner
func NewUniversalScanner(db *database.DB, logger *zap.Logger, renameTracker *UniversalRenameTracker, clientFactory filesystem.ClientFactory) *UniversalScanner {
	scanner := &UniversalScanner{
		db:                 db,
		logger:             logger,
		renameTracker:      renameTracker,
		clientFactory:      clientFactory,
		scanQueue:          make(chan ScanJob, 1000),
		workers:            4,
		maxConcurrentScans: 4,
		scanSem:            semaphore.NewWeighted(4),
		stopCh:             make(chan struct{}),
		protocolScanners:   make(map[string]ProtocolScanner),
		activeScans:        make(map[string]*ScanStatus),
	}

	// Register protocol scanners
	scanner.RegisterProtocolScanner("local", NewLocalScanner(db, logger))
	scanner.RegisterProtocolScanner("smb", NewSMBScanner(db, logger))
	scanner.RegisterProtocolScanner("ftp", NewFTPScanner(logger))
	scanner.RegisterProtocolScanner("nfs", NewNFSScanner(logger))
	scanner.RegisterProtocolScanner("webdav", NewWebDAVScanner(logger))

	return scanner
}

// SetAggregationService sets the aggregation service for post-scan entity creation.
func (s *UniversalScanner) SetAggregationService(svc *AggregationService) {
	s.aggregationService = svc
}

// RegisterProtocolScanner registers a protocol-specific scanner
func (s *UniversalScanner) RegisterProtocolScanner(protocol string, scanner ProtocolScanner) {
	s.protocolScannersMu.Lock()
	defer s.protocolScannersMu.Unlock()
	s.protocolScanners[protocol] = scanner
}

// Start begins the universal scanning service
func (s *UniversalScanner) Start() error {
	s.logger.Info("Starting universal scanner service", zap.Int("workers", s.workers))

	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.scanWorker(i)
	}

	return nil
}

// Stop stops the universal scanning service
func (s *UniversalScanner) Stop() {
	s.logger.Info("Stopping universal scanner service")
	close(s.stopCh)
	s.wg.Wait()
	s.logger.Info("Universal scanner service stopped")
}

// QueueScan adds a scan job to the queue
func (s *UniversalScanner) QueueScan(job ScanJob) error {
	select {
	case s.scanQueue <- job:
		s.logger.Debug("Queued scan job",
			zap.String("job_id", job.ID),
			zap.String("storage_root", job.StorageRoot.Name),
			zap.String("protocol", job.StorageRoot.Protocol),
			zap.String("path", job.Path))
		return nil
	default:
		return fmt.Errorf("scan queue is full")
	}
}

// scanWorker processes scan jobs
func (s *UniversalScanner) scanWorker(workerID int) {
	defer s.wg.Done()

	s.logger.Info("Universal scan worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-s.stopCh:
			return
		case job := <-s.scanQueue:
			s.processScanJob(job, workerID)
		}
	}
}

// processScanJob processes a single scan job
func (s *UniversalScanner) processScanJob(job ScanJob, workerID int) {
	// Acquire semaphore to limit concurrent scans
	if err := s.scanSem.Acquire(job.Context, 1); err != nil {
		s.logger.Debug("Scan job cancelled before acquiring semaphore",
			zap.String("job_id", job.ID),
			zap.Error(err))
		return
	}
	defer s.scanSem.Release(1)

	s.logger.Debug("Processing scan job",
		zap.Int("worker_id", workerID),
		zap.String("job_id", job.ID),
		zap.String("storage_root", job.StorageRoot.Name),
		zap.String("protocol", job.StorageRoot.Protocol))

	// Create scan status
	status := &ScanStatus{
		JobID:           job.ID,
		StorageRootName: job.StorageRoot.Name,
		Protocol:        job.StorageRoot.Protocol,
		StartTime:       time.Now(),
		Status:          "running",
	}

	// Track active scan
	s.activeScansMu.Lock()
	s.activeScans[job.ID] = status
	s.activeScansMu.Unlock()

	// Retain completed scan status for 60 seconds so polling clients
	// can read the final result before it is garbage-collected.
	defer func() {
		go func() {
			time.Sleep(60 * time.Second)
			s.activeScansMu.Lock()
			delete(s.activeScans, job.ID)
			s.activeScansMu.Unlock()
		}()
	}()

	// Get protocol scanner
	s.protocolScannersMu.RLock()
	protocolScanner, exists := s.protocolScanners[job.StorageRoot.Protocol]
	s.protocolScannersMu.RUnlock()
	if !exists {
		s.logger.Error("No scanner for protocol",
			zap.String("protocol", job.StorageRoot.Protocol),
			zap.String("job_id", job.ID))
		status.updateStatus("failed")
		return
	}

	// Create filesystem client
	client, err := s.clientFactory.CreateClient(&filesystem.StorageConfig{
		ID:       job.StorageRoot.Name,
		Name:     job.StorageRoot.Name,
		Protocol: job.StorageRoot.Protocol,
		Settings: s.storageRootToSettings(job.StorageRoot),
	})
	if err != nil {
		s.logger.Error("Failed to create filesystem client",
			zap.String("protocol", job.StorageRoot.Protocol),
			zap.String("job_id", job.ID),
			zap.Error(err))
		status.updateStatus("failed")
		return
	}

	// Connect to filesystem
	if err := client.Connect(job.Context); err != nil {
		s.logger.Error("Failed to connect to filesystem",
			zap.String("protocol", job.StorageRoot.Protocol),
			zap.String("job_id", job.ID),
			zap.Error(err))
		status.updateStatus("failed")
		return
	}
	defer client.Disconnect(job.Context)

	// Perform the scan
	if err := protocolScanner.ScanPath(job.Context, client, job, status); err != nil {
		s.logger.Error("Scan failed",
			zap.String("job_id", job.ID),
			zap.Error(err))
		status.updateStatus("failed")
		return
	}

	status.updateStatus("completed")
	snapshot := status.GetSnapshot()
	s.logger.Info("Scan completed successfully",
		zap.String("job_id", job.ID),
		zap.String("storage_root", job.StorageRoot.Name),
		zap.Int64("files_processed", snapshot.FilesProcessed),
		zap.Duration("duration", time.Since(snapshot.StartTime)))

	// Run post-scan aggregation to create media entities from scanned files
	if s.aggregationService != nil {
		go func() {
			if err := s.aggregationService.AggregateAfterScan(job.Context, int64(job.StorageRoot.ID)); err != nil {
				s.logger.Error("Post-scan aggregation failed",
					zap.String("job_id", job.ID),
					zap.Error(err))
			}
		}()
	}
}

// GetActiveScanStatus returns the status of an active scan
func (s *UniversalScanner) GetActiveScanStatus(jobID string) (*ScanStatus, bool) {
	s.activeScansMu.RLock()
	defer s.activeScansMu.RUnlock()
	status, exists := s.activeScans[jobID]
	return status, exists
}

// GetAllActiveScanStatuses returns all active scan statuses
func (s *UniversalScanner) GetAllActiveScanStatuses() map[string]*ScanStatus {
	s.activeScansMu.RLock()
	defer s.activeScansMu.RUnlock()

	statuses := make(map[string]*ScanStatus)
	for id, status := range s.activeScans {
		// Use GetSnapshot for thread-safe copy of status fields
		snapshot := status.GetSnapshot()
		statuses[id] = &snapshot
	}
	return statuses
}

// storageRootToSettings converts StorageRoot to filesystem settings
func (s *UniversalScanner) storageRootToSettings(root *models.StorageRoot) map[string]interface{} {
	settings := make(map[string]interface{})

	switch root.Protocol {
	case "local":
		if root.Path != nil {
			settings["base_path"] = *root.Path
		}

	case "smb":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Port != nil {
			settings["port"] = *root.Port
		}
		if root.Path != nil {
			settings["share"] = *root.Path
		}
		if root.Username != nil {
			settings["username"] = *root.Username
		}
		if root.Password != nil {
			settings["password"] = *root.Password
		}
		if root.Domain != nil {
			settings["domain"] = *root.Domain
		}

	case "ftp":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Port != nil {
			settings["port"] = *root.Port
		}
		if root.Username != nil {
			settings["username"] = *root.Username
		}
		if root.Password != nil {
			settings["password"] = *root.Password
		}

	case "nfs":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Path != nil {
			settings["export_path"] = *root.Path
		}
		if root.MountPoint != nil {
			settings["mount_point"] = *root.MountPoint
		}
		if root.Options != nil {
			settings["options"] = *root.Options
		}

	case "webdav":
		if root.URL != nil {
			settings["url"] = *root.URL
		}
		if root.Username != nil {
			settings["username"] = *root.Username
		}
		if root.Password != nil {
			settings["password"] = *root.Password
		}
	}

	return settings
}

// updateStatus safely updates the scan status
func (s *ScanStatus) updateStatus(newStatus string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = newStatus
}

// updateCurrentPath safely updates the current path being scanned
func (s *ScanStatus) updateCurrentPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentPath = path
}

// incrementCounters safely increments the various counters
func (s *ScanStatus) incrementCounters(processed, found, updated, deleted, errors int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FilesProcessed += processed
	s.FilesFound += found
	s.FilesUpdated += updated
	s.FilesDeleted += deleted
	s.ErrorCount += errors
}

// GetSnapshot returns a thread-safe snapshot of the scan status
func (s *ScanStatus) GetSnapshot() ScanStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return ScanStatus{
		JobID:           s.JobID,
		StorageRootName: s.StorageRootName,
		Protocol:        s.Protocol,
		StartTime:       s.StartTime,
		CurrentPath:     s.CurrentPath,
		FilesProcessed:  s.FilesProcessed,
		FilesFound:      s.FilesFound,
		FilesUpdated:    s.FilesUpdated,
		FilesDeleted:    s.FilesDeleted,
		ErrorCount:      s.ErrorCount,
		Status:          s.Status,
	}
}

// LocalScanner implements protocol-specific scanning for local filesystem
type LocalScanner struct {
	db             *database.DB
	logger         *zap.Logger
	sem            *semaphore.Weighted
	maxConcurrency int
}

func NewLocalScanner(db *database.DB, logger *zap.Logger) *LocalScanner {
	maxConcurrency := 10 // optimal for local filesystem
	return &LocalScanner{
		db:             db,
		logger:         logger,
		sem:            semaphore.NewWeighted(int64(maxConcurrency)),
		maxConcurrency: maxConcurrency,
	}
}

func (s *LocalScanner) ScanPath(ctx context.Context, client filesystem.FileSystemClient, job ScanJob, status *ScanStatus) error {
	return s.scanDirectory(ctx, client, job.Path, job, status, 0)
}

func (s *LocalScanner) scanDirectory(ctx context.Context, client filesystem.FileSystemClient, path string, job ScanJob, status *ScanStatus, depth int) error {
	if depth > job.MaxDepth {
		return nil
	}

	status.updateCurrentPath(path)

	files, err := client.ListDirectory(ctx, path)
	if err != nil {
		status.incrementCounters(0, 0, 0, 0, 1)
		return fmt.Errorf("failed to list directory %s: %w", path, err)
	}

	for _, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fullPath := filepath.Join(path, file.Name)

		// Process file/directory
		if err := s.processFileInfo(ctx, client, fullPath, file, job, status); err != nil {
			s.logger.Error("Failed to process file",
				zap.String("path", fullPath),
				zap.Error(err))
			status.incrementCounters(0, 0, 0, 0, 1)
		}

		// Recurse into subdirectories
		if file.IsDir {
			if err := s.scanDirectory(ctx, client, fullPath, job, status, depth+1); err != nil {
				s.logger.Error("Failed to scan subdirectory",
					zap.String("path", fullPath),
					zap.Error(err))
			}
		}
	}

	return nil
}

func (s *LocalScanner) processFileInfo(ctx context.Context, client filesystem.FileSystemClient, path string, file *filesystem.FileInfo, job ScanJob, status *ScanStatus) error {
	return insertFileRecord(ctx, s.db, path, file, job, status, s.logger)
}

func (s *LocalScanner) GetScanStrategy() ScanStrategy {
	return ScanStrategy{
		UseRecursiveListing:     true,
		BatchSize:               1000,
		ParallelDirectories:     true,
		ChecksumCalculation:     true,
		MetadataExtraction:      true,
		RealTimeChangeDetection: true,
	}
}

func (s *LocalScanner) SupportsIncrementalScan() bool {
	return true
}

func (s *LocalScanner) GetOptimalBatchSize() int {
	return 1000
}

// SMBScanner implements protocol-specific scanning for SMB
type SMBScanner struct {
	db     *database.DB
	logger *zap.Logger
}

func NewSMBScanner(db *database.DB, logger *zap.Logger) *SMBScanner {
	return &SMBScanner{db: db, logger: logger}
}

func (s *SMBScanner) ScanPath(ctx context.Context, client filesystem.FileSystemClient, job ScanJob, status *ScanStatus) error {
	// SMB-specific scanning logic
	return s.scanDirectory(ctx, client, job.Path, job, status, 0)
}

func (s *SMBScanner) scanDirectory(ctx context.Context, client filesystem.FileSystemClient, path string, job ScanJob, status *ScanStatus, depth int) error {
	// Similar to LocalScanner but with SMB-specific optimizations
	if depth > job.MaxDepth {
		return nil
	}

	status.updateCurrentPath(path)

	files, err := client.ListDirectory(ctx, path)
	if err != nil {
		status.incrementCounters(0, 0, 0, 0, 1)
		return fmt.Errorf("failed to list SMB directory %s: %w", path, err)
	}

	// Process files in batches for better SMB performance
	batchSize := s.GetOptimalBatchSize()
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		for _, file := range batch {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			fullPath := filepath.Join(path, file.Name)

			if err := insertFileRecord(ctx, s.db, fullPath, file, job, status, s.logger); err != nil {
				s.logger.Error("Failed to insert file record",
					zap.String("path", fullPath),
					zap.Error(err))
				status.incrementCounters(0, 0, 0, 0, 1)
			}

			if file.IsDir {
				if err := s.scanDirectory(ctx, client, fullPath, job, status, depth+1); err != nil {
					s.logger.Error("Failed to scan SMB subdirectory",
						zap.String("path", fullPath),
						zap.Error(err))
				}
			}
		}
	}

	return nil
}

func (s *SMBScanner) GetScanStrategy() ScanStrategy {
	return ScanStrategy{
		UseRecursiveListing:     false, // SMB benefits from controlled recursion
		BatchSize:               500,   // Smaller batches for network efficiency
		ParallelDirectories:     false, // Avoid overwhelming SMB server
		ChecksumCalculation:     false, // Expensive over network
		MetadataExtraction:      true,
		RealTimeChangeDetection: false,
	}
}

func (s *SMBScanner) SupportsIncrementalScan() bool {
	return true
}

func (s *SMBScanner) GetOptimalBatchSize() int {
	return 500
}

// insertFileRecord inserts or updates a file record in the database.
// This is the core function that makes the scanner actually populate
// the catalog, shared by all protocol scanners.
func insertFileRecord(ctx context.Context, db *database.DB, path string, file *filesystem.FileInfo, job ScanJob, status *ScanStatus, logger *zap.Logger) error {
	if db == nil {
		status.incrementCounters(1, 1, 0, 0, 0)
		return nil
	}

	// Resolve storage root ID
	var storageRootID int64
	err := db.QueryRowContext(ctx,
		"SELECT id FROM storage_roots WHERE name = ? LIMIT 1",
		job.StorageRoot.Name,
	).Scan(&storageRootID)
	if err != nil {
		// Storage root not in DB yet — insert it
		if db.Dialect().IsPostgres() {
			// PostgreSQL: use INSERT ... ON CONFLICT DO NOTHING + RETURNING
			err2 := db.QueryRowContext(ctx,
				`INSERT INTO storage_roots (name, protocol, host, port, path, username, password, domain, enabled, max_depth)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, true, ?)
				 ON CONFLICT (name) DO NOTHING
				 RETURNING id`,
				job.StorageRoot.Name, job.StorageRoot.Protocol,
				job.StorageRoot.Host, job.StorageRoot.Port, job.StorageRoot.Path,
				job.StorageRoot.Username, job.StorageRoot.Password, job.StorageRoot.Domain,
				job.StorageRoot.MaxDepth,
			).Scan(&storageRootID)
			if err2 != nil {
				// ON CONFLICT DO NOTHING returns no rows — re-query
				_ = db.QueryRowContext(ctx,
					"SELECT id FROM storage_roots WHERE name = ? LIMIT 1",
					job.StorageRoot.Name,
				).Scan(&storageRootID)
			}
		} else {
			// SQLite path
			insertedID, insertErr := db.InsertReturningID(ctx,
				`INSERT OR IGNORE INTO storage_roots (name, protocol, host, port, path, username, password, domain, enabled, max_depth)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?)`,
				job.StorageRoot.Name, job.StorageRoot.Protocol,
				job.StorageRoot.Host, job.StorageRoot.Port, job.StorageRoot.Path,
				job.StorageRoot.Username, job.StorageRoot.Password, job.StorageRoot.Domain,
				job.StorageRoot.MaxDepth,
			)
			if insertErr != nil {
				return fmt.Errorf("insert storage root: %w", insertErr)
			}
			storageRootID = insertedID
			if storageRootID == 0 {
				_ = db.QueryRowContext(ctx,
					"SELECT id FROM storage_roots WHERE name = ? LIMIT 1",
					job.StorageRoot.Name,
				).Scan(&storageRootID)
			}
		}
	}

	name := file.Name
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	mimeType := mime.TypeByExtension("." + ext)
	fileType := classifyFileType(ext)

	// Resolve parent directory ID (if path has a parent)
	var parentID *int64
	parentPath := filepath.Dir(path)
	if parentPath != "." && parentPath != "/" && parentPath != "" {
		var pid int64
		err := db.QueryRowContext(ctx,
			"SELECT id FROM files WHERE path = ? AND storage_root_id = ? AND is_directory = 1 LIMIT 1",
			parentPath, storageRootID,
		).Scan(&pid)
		if err == nil {
			parentID = &pid
		}
	}

	modifiedAt := file.ModTime
	if modifiedAt.IsZero() {
		modifiedAt = time.Now()
	}

	isDir := file.IsDir

	// Upsert: insert or update on conflict (same storage_root + path)
	_, err = db.ExecContext(ctx,
		`INSERT INTO files (storage_root_id, path, name, extension, mime_type, file_type, size, is_directory, modified_at, last_scan_at, parent_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
		 ON CONFLICT(storage_root_id, path) DO UPDATE SET
		   size = excluded.size,
		   modified_at = excluded.modified_at,
		   last_scan_at = CURRENT_TIMESTAMP,
		   deleted = false,
		   deleted_at = NULL`,
		storageRootID, path, name, ext, mimeType, fileType, file.Size,
		isDir, modifiedAt, parentID,
	)
	if err != nil && db.Dialect().IsSQLite() {
		// SQLite fallback if UNIQUE constraint doesn't exist yet
		_, err = db.ExecContext(ctx,
			`INSERT OR REPLACE INTO files (storage_root_id, path, name, extension, mime_type, file_type, size, is_directory, modified_at, last_scan_at, parent_id)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)`,
			storageRootID, path, name, ext, mimeType, fileType, file.Size,
			isDir, modifiedAt, parentID,
		)
		if err != nil {
			return fmt.Errorf("insert file %s: %w", path, err)
		}
	} else if err != nil {
		return fmt.Errorf("insert file %s: %w", path, err)
	}

	status.incrementCounters(1, 1, 0, 0, 0)
	return nil
}

// classifyFileType returns a file_type category based on extension.
func classifyFileType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case "mp4", "mkv", "avi", "mov", "wmv", "flv", "m4v", "ts", "webm", "mpg", "mpeg":
		return "video"
	case "mp3", "flac", "wav", "m4a", "aac", "ogg", "wma", "ape", "opus":
		return "audio"
	case "jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "tiff", "ico":
		return "image"
	case "pdf", "epub", "mobi", "djvu", "cbr", "cbz", "cb7", "cbt":
		return "book"
	case "exe", "msi", "dmg", "pkg", "iso", "img", "deb", "rpm", "apk", "appimage":
		return "software"
	case "zip", "rar", "7z", "tar", "gz", "bz2", "xz", "zst":
		return "archive"
	case "txt", "md", "doc", "docx", "rtf", "odt", "csv", "xls", "xlsx":
		return "document"
	case "html", "htm", "css", "js", "go", "py", "java", "c", "cpp", "rs", "sh":
		return "code"
	default:
		return "other"
	}
}

// Similar implementations for FTP, NFS, and WebDAV scanners...

type FTPScanner struct {
	logger *zap.Logger
}

func NewFTPScanner(logger *zap.Logger) *FTPScanner {
	return &FTPScanner{logger: logger}
}

func (s *FTPScanner) ScanPath(ctx context.Context, client filesystem.FileSystemClient, job ScanJob, status *ScanStatus) error {
	// FTP-specific scanning logic
	return nil
}

func (s *FTPScanner) GetScanStrategy() ScanStrategy {
	return ScanStrategy{
		UseRecursiveListing:     false,
		BatchSize:               100,
		ParallelDirectories:     false,
		ChecksumCalculation:     false,
		MetadataExtraction:      false,
		RealTimeChangeDetection: false,
	}
}

func (s *FTPScanner) SupportsIncrementalScan() bool {
	return false
}

func (s *FTPScanner) GetOptimalBatchSize() int {
	return 100
}

type NFSScanner struct {
	logger *zap.Logger
}

func NewNFSScanner(logger *zap.Logger) *NFSScanner {
	return &NFSScanner{logger: logger}
}

func (s *NFSScanner) ScanPath(ctx context.Context, client filesystem.FileSystemClient, job ScanJob, status *ScanStatus) error {
	// NFS-specific scanning logic
	return nil
}

func (s *NFSScanner) GetScanStrategy() ScanStrategy {
	return ScanStrategy{
		UseRecursiveListing:     true,
		BatchSize:               800,
		ParallelDirectories:     true,
		ChecksumCalculation:     true,
		MetadataExtraction:      true,
		RealTimeChangeDetection: false,
	}
}

func (s *NFSScanner) SupportsIncrementalScan() bool {
	return true
}

func (s *NFSScanner) GetOptimalBatchSize() int {
	return 800
}

type WebDAVScanner struct {
	logger *zap.Logger
}

func NewWebDAVScanner(logger *zap.Logger) *WebDAVScanner {
	return &WebDAVScanner{logger: logger}
}

func (s *WebDAVScanner) ScanPath(ctx context.Context, client filesystem.FileSystemClient, job ScanJob, status *ScanStatus) error {
	// WebDAV-specific scanning logic
	return nil
}

func (s *WebDAVScanner) GetScanStrategy() ScanStrategy {
	return ScanStrategy{
		UseRecursiveListing:     false,
		BatchSize:               200,
		ParallelDirectories:     false,
		ChecksumCalculation:     false,
		MetadataExtraction:      true,
		RealTimeChangeDetection: false,
	}
}

func (s *WebDAVScanner) SupportsIncrementalScan() bool {
	return false
}

func (s *WebDAVScanner) GetOptimalBatchSize() int {
	return 200
}
