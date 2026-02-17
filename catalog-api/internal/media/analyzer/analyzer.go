package analyzer

import (
	"catalogizer/internal/media/detector"
	mediamodels "catalogizer/internal/media/models"
	"catalogizer/internal/media/providers"
	"catalogizer/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MediaAnalyzer handles real-time analysis of directory content
type MediaAnalyzer struct {
	db              *sql.DB
	detector        *detector.DetectionEngine
	providerManager *providers.ProviderManager
	logger          *zap.Logger
	analysisQueue   chan AnalysisRequest
	workers         int
	stopCh          chan struct{}
	wg              sync.WaitGroup
	mu              sync.RWMutex
	pendingAnalysis map[string]*AnalysisRequest
}

// AnalysisRequest represents a request to analyze a directory
type AnalysisRequest struct {
	DirectoryPath string
	SmbRoot       string
	Priority      int // Higher number = higher priority
	Timestamp     time.Time
	Callback      func(*AnalysisResult, error)
}

// AnalysisResult represents the result of directory analysis
type AnalysisResult struct {
	DirectoryAnalysis *mediamodels.DirectoryAnalysis
	MediaItem         *mediamodels.MediaItem
	ExternalMetadata  []mediamodels.ExternalMetadata
	QualityAnalysis   *QualityAnalysis
	UpdatedFiles      []mediamodels.MediaFile
}

// QualityAnalysis represents quality analysis of media files
type QualityAnalysis struct {
	BestQuality        *mediamodels.QualityInfo
	AvailableQualities []string
	TotalFiles         int
	TotalSize          int64
	DuplicateCount     int
	MissingQualities   []string
}

// NewMediaAnalyzer creates a new media analyzer
func NewMediaAnalyzer(db *sql.DB, detector *detector.DetectionEngine, providerManager *providers.ProviderManager, logger *zap.Logger) *MediaAnalyzer {
	return &MediaAnalyzer{
		db:              db,
		detector:        detector,
		providerManager: providerManager,
		logger:          logger,
		analysisQueue:   make(chan AnalysisRequest, 1000),
		workers:         4, // Number of concurrent workers
		stopCh:          make(chan struct{}),
		pendingAnalysis: make(map[string]*AnalysisRequest),
	}
}

// Start starts the analyzer workers
func (ma *MediaAnalyzer) Start() {
	ma.logger.Info("Starting media analyzer", zap.Int("workers", ma.workers))

	for i := 0; i < ma.workers; i++ {
		ma.wg.Add(1)
		go ma.worker(i)
	}
}

// Stop stops the analyzer workers
func (ma *MediaAnalyzer) Stop() {
	ma.logger.Info("Stopping media analyzer")
	close(ma.stopCh)
	ma.wg.Wait()
}

// AnalyzeDirectory queues a directory for analysis
func (ma *MediaAnalyzer) AnalyzeDirectory(ctx context.Context, directoryPath, smbRoot string, priority int) error {
	request := AnalysisRequest{
		DirectoryPath: directoryPath,
		SmbRoot:       smbRoot,
		Priority:      priority,
		Timestamp:     time.Now(),
	}

	// Check if already pending and add to pending map
	// Use a function scope to ensure lock is released with defer before select block
	shouldQueue := func() bool {
		ma.mu.Lock()
		defer ma.mu.Unlock()

		if existing, exists := ma.pendingAnalysis[directoryPath]; exists {
			// Update priority if higher
			if priority > existing.Priority {
				existing.Priority = priority
			}
			return false // Already pending, don't queue again
		}
		ma.pendingAnalysis[directoryPath] = &request
		return true // New request, should queue
	}()

	if !shouldQueue {
		return nil // Already pending
	}

	// Queue the analysis request
	select {
	case ma.analysisQueue <- request:
		return nil
	case <-ctx.Done():
		ma.mu.Lock()
		defer ma.mu.Unlock()
		delete(ma.pendingAnalysis, directoryPath)
		return ctx.Err()
	}
}

// AnalyzeDirectorySync performs synchronous directory analysis
func (ma *MediaAnalyzer) AnalyzeDirectorySync(ctx context.Context, directoryPath, smbRoot string) (*AnalysisResult, error) {
	// Get directory files
	files, err := ma.getDirectoryFiles(directoryPath, smbRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory files: %w", err)
	}

	// Convert to detector.FileInfo
	detectorFiles := make([]detector.FileInfo, len(files))
	for i, file := range files {
		extension := ""
		if file.Extension != nil {
			extension = *file.Extension
		}
		detectorFiles[i] = detector.FileInfo{
			Name:      file.Name,
			Path:      file.Path,
			Size:      file.Size,
			Extension: extension,
			IsDir:     file.IsDirectory,
		}
	}

	// Run detection
	detectionResult, err := ma.detector.AnalyzeDirectory(directoryPath, detectorFiles)
	if err != nil {
		return nil, fmt.Errorf("detection failed: %w", err)
	}

	if detectionResult == nil {
		return &AnalysisResult{}, nil // No detection
	}

	// Create or update directory analysis record
	dirAnalysis, err := ma.createDirectoryAnalysis(directoryPath, smbRoot, detectionResult)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory analysis: %w", err)
	}

	// Create or get media item
	mediaItem, err := ma.createOrUpdateMediaItem(ctx, detectionResult, dirAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to create media item: %w", err)
	}

	// Fetch external metadata
	externalMetadata, err := ma.fetchExternalMetadata(ctx, mediaItem)
	if err != nil {
		ma.logger.Error("Failed to fetch external metadata", zap.Error(err))
		// Don't fail the whole analysis for metadata errors
	}

	// Analyze quality
	qualityAnalysis, err := ma.analyzeQuality(files, mediaItem)
	if err != nil {
		ma.logger.Error("Failed to analyze quality", zap.Error(err))
	}

	// Update media files
	updatedFiles, err := ma.updateMediaFiles(mediaItem.ID, files, directoryPath, smbRoot)
	if err != nil {
		ma.logger.Error("Failed to update media files", zap.Error(err))
	}

	result := &AnalysisResult{
		DirectoryAnalysis: dirAnalysis,
		MediaItem:         mediaItem,
		ExternalMetadata:  externalMetadata,
		QualityAnalysis:   qualityAnalysis,
		UpdatedFiles:      updatedFiles,
	}

	return result, nil
}

// worker processes analysis requests
func (ma *MediaAnalyzer) worker(workerID int) {
	defer ma.wg.Done()

	ma.logger.Info("Media analyzer worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-ma.stopCh:
			return

		case request := <-ma.analysisQueue:
			ma.mu.Lock()
			delete(ma.pendingAnalysis, request.DirectoryPath)
			ma.mu.Unlock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			result, err := ma.AnalyzeDirectorySync(ctx, request.DirectoryPath, request.SmbRoot)
			cancel()

			if request.Callback != nil {
				request.Callback(result, err)
			}

			if err != nil {
				ma.logger.Error("Directory analysis failed",
					zap.String("directory", request.DirectoryPath),
					zap.String("smb_root", request.SmbRoot),
					zap.Error(err))
			} else {
				ma.logger.Info("Directory analysis completed",
					zap.String("directory", request.DirectoryPath),
					zap.String("media_type", result.MediaItem.MediaType.Name))
			}
		}
	}
}

// getDirectoryFiles retrieves files in a directory from the catalog database
func (ma *MediaAnalyzer) getDirectoryFiles(directoryPath, smbRoot string) ([]models.FileInfo, error) {
	query := `
		SELECT id, name, path, is_directory, size, last_modified, extension, mime_type
		FROM files
		WHERE path LIKE ? AND smb_root = ?
		ORDER BY is_directory DESC, name ASC
	`

	rows, err := ma.db.Query(query, directoryPath+"%", smbRoot)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.FileInfo
	for rows.Next() {
		var file models.FileInfo
		err := rows.Scan(
			&file.ID, &file.Name, &file.Path, &file.IsDirectory,
			&file.Size, &file.LastModified, &file.Extension, &file.MimeType,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// createDirectoryAnalysis creates or updates directory analysis record
func (ma *MediaAnalyzer) createDirectoryAnalysis(directoryPath, smbRoot string, detection *detector.DetectionResult) (*mediamodels.DirectoryAnalysis, error) {
	analysisDataJSON, _ := json.Marshal(detection.AnalysisData)

	query := `
		INSERT OR REPLACE INTO directory_analysis
		(directory_path, smb_root, media_item_id, confidence_score, detection_method, analysis_data, last_analyzed, files_count, total_size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Calculate files count and total size from analysis data
	filesCount := 0
	totalSize := int64(0)
	if detection.AnalysisData != nil {
		for _, count := range detection.AnalysisData.FileTypes {
			filesCount += count
		}
		for _, size := range detection.AnalysisData.SizeDistribution {
			totalSize += size
		}
	}

	_, err := ma.db.Exec(query,
		directoryPath, smbRoot, nil, // media_item_id will be set later
		detection.Confidence, detection.Method, string(analysisDataJSON),
		time.Now(), filesCount, totalSize,
	)
	if err != nil {
		return nil, err
	}

	// Return the created record
	return &mediamodels.DirectoryAnalysis{
		DirectoryPath:   directoryPath,
		SmbRoot:         smbRoot,
		ConfidenceScore: detection.Confidence,
		DetectionMethod: detection.Method,
		AnalysisData:    detection.AnalysisData,
		LastAnalyzed:    time.Now(),
		FilesCount:      filesCount,
		TotalSize:       totalSize,
	}, nil
}

// createOrUpdateMediaItem creates or updates media item
func (ma *MediaAnalyzer) createOrUpdateMediaItem(ctx context.Context, detection *detector.DetectionResult, dirAnalysis *mediamodels.DirectoryAnalysis) (*mediamodels.MediaItem, error) {
	// Check if media item already exists
	var existingID *int64
	err := ma.db.QueryRow(
		"SELECT media_item_id FROM directory_analysis WHERE directory_path = ?",
		dirAnalysis.DirectoryPath,
	).Scan(&existingID)

	if err == nil && existingID != nil {
		// Update existing media item
		return ma.updateExistingMediaItem(*existingID, detection)
	}

	// Create new media item
	genreJSON, _ := json.Marshal([]string{}) // Empty for now
	castCrewJSON, _ := json.Marshal(&mediamodels.CastCrew{})

	query := `
		INSERT INTO media_items
		(media_type_id, title, year, description, genre, cast_crew, status, first_detected, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?)
	`

	result, err := ma.db.Exec(query,
		detection.MediaTypeID, detection.SuggestedTitle, detection.SuggestedYear,
		nil, string(genreJSON), string(castCrewJSON),
		time.Now(), time.Now(),
	)
	if err != nil {
		return nil, err
	}

	mediaItemID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update directory analysis with media item ID
	_, err = ma.db.Exec(
		"UPDATE directory_analysis SET media_item_id = ? WHERE directory_path = ?",
		mediaItemID, dirAnalysis.DirectoryPath,
	)
	if err != nil {
		return nil, err
	}

	// Return the created media item
	mediaItem := &mediamodels.MediaItem{
		ID:            mediaItemID,
		MediaTypeID:   detection.MediaTypeID,
		MediaType:     detection.MediaType,
		Title:         detection.SuggestedTitle,
		Year:          detection.SuggestedYear,
		Status:        "active",
		FirstDetected: time.Now(),
		LastUpdated:   time.Now(),
	}

	return mediaItem, nil
}

// updateExistingMediaItem updates an existing media item
func (ma *MediaAnalyzer) updateExistingMediaItem(mediaItemID int64, detection *detector.DetectionResult) (*mediamodels.MediaItem, error) {
	// Get existing media item
	query := `
		SELECT id, media_type_id, title, year, description, genre, director, cast_crew, rating, runtime, language, country, status, first_detected, last_updated
		FROM media_items WHERE id = ?
	`

	var item mediamodels.MediaItem
	var genreJSON, castCrewJSON string

	err := ma.db.QueryRow(query, mediaItemID).Scan(
		&item.ID, &item.MediaTypeID, &item.Title, &item.Year, &item.Description,
		&genreJSON, &item.Director, &castCrewJSON, &item.Rating, &item.Runtime,
		&item.Language, &item.Country, &item.Status, &item.FirstDetected, &item.LastUpdated,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal([]byte(genreJSON), &item.Genre)
	json.Unmarshal([]byte(castCrewJSON), &item.CastCrew)

	// Update last_updated timestamp
	_, err = ma.db.Exec("UPDATE media_items SET last_updated = ? WHERE id = ?", time.Now(), mediaItemID)
	if err != nil {
		return nil, err
	}

	item.LastUpdated = time.Now()
	return &item, nil
}

// fetchExternalMetadata fetches metadata from external providers
func (ma *MediaAnalyzer) fetchExternalMetadata(ctx context.Context, mediaItem *mediamodels.MediaItem) ([]mediamodels.ExternalMetadata, error) {
	if mediaItem.MediaType == nil {
		return nil, fmt.Errorf("media type not available")
	}

	// Get the best match from providers
	bestResult, providerName, err := ma.providerManager.GetBestMatch(
		ctx, mediaItem.Title, mediaItem.MediaType.Name, mediaItem.Year,
	)
	if err != nil || bestResult == nil {
		return nil, err
	}

	// Get detailed metadata
	metadata, err := ma.providerManager.GetDetails(ctx, providerName, bestResult.ExternalID)
	if err != nil {
		return nil, err
	}

	metadata.MediaItemID = mediaItem.ID

	// Save to database
	query := `
		INSERT OR REPLACE INTO external_metadata
		(media_item_id, provider, external_id, data, rating, review_url, cover_url, trailer_url, last_fetched)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = ma.db.Exec(query,
		metadata.MediaItemID, metadata.Provider, metadata.ExternalID, metadata.Data,
		metadata.Rating, metadata.ReviewURL, metadata.CoverURL, metadata.TrailerURL, metadata.LastFetched,
	)
	if err != nil {
		return nil, err
	}

	return []mediamodels.ExternalMetadata{*metadata}, nil
}

// analyzeQuality analyzes the quality of media files
func (ma *MediaAnalyzer) analyzeQuality(files []models.FileInfo, mediaItem *mediamodels.MediaItem) (*QualityAnalysis, error) {
	if mediaItem.MediaType == nil {
		return nil, fmt.Errorf("media type not available")
	}

	analysis := &QualityAnalysis{
		AvailableQualities: make([]string, 0),
		TotalFiles:         len(files),
	}

	// Analyze video/audio files
	mediaFiles := ma.filterMediaFiles(files, mediaItem.MediaType.Name)
	analysis.TotalFiles = len(mediaFiles)

	for _, file := range mediaFiles {
		analysis.TotalSize += file.Size

		// Extract quality information from filename
		qualityInfo := ma.extractQualityFromFilename(file.Name, file.Extension)
		if qualityInfo != nil {
			qualityName := qualityInfo.GetDisplayName()
			if !contains(analysis.AvailableQualities, qualityName) {
				analysis.AvailableQualities = append(analysis.AvailableQualities, qualityName)
			}

			// Track best quality
			if analysis.BestQuality == nil || qualityInfo.IsBetterThan(analysis.BestQuality) {
				analysis.BestQuality = qualityInfo
			}
		}
	}

	return analysis, nil
}

// filterMediaFiles filters files relevant to the media type
func (ma *MediaAnalyzer) filterMediaFiles(files []models.FileInfo, mediaType string) []models.FileInfo {
	mediaExtensions := map[string][]string{
		"movie":     {".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".m4v"},
		"tv_show":   {".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".m4v"},
		"anime":     {".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".m4v"},
		"music":     {".mp3", ".flac", ".wav", ".m4a", ".aac", ".ogg", ".wma"},
		"audiobook": {".mp3", ".m4a", ".m4b", ".aac", ".ogg"},
		"podcast":   {".mp3", ".m4a", ".aac", ".ogg"},
		"comic":     {".cbr", ".cbz", ".cb7", ".cbt", ".pdf", ".epub"},
		"software":  {".exe", ".msi", ".dmg", ".pkg", ".iso", ".img", ".deb", ".rpm", ".apk", ".appimage"},
		"game":      {".exe", ".iso", ".img", ".bin", ".rom", ".nes", ".sfc", ".gba", ".nds"},
	}

	extensions, exists := mediaExtensions[mediaType]
	if !exists {
		return files // Return all files if media type not recognized
	}

	var filtered []models.FileInfo
	for _, file := range files {
		if file.IsDirectory {
			continue
		}

		if file.Extension != nil {
			for _, ext := range extensions {
				if strings.EqualFold(*file.Extension, ext) {
					filtered = append(filtered, file)
					break
				}
			}
		}
	}

	return filtered
}

// extractQualityFromFilename extracts quality information from filename
func (ma *MediaAnalyzer) extractQualityFromFilename(filename string, extension *string) *mediamodels.QualityInfo {
	lower := strings.ToLower(filename)
	quality := &mediamodels.QualityInfo{}

	// Resolution detection
	if strings.Contains(lower, "2160p") || strings.Contains(lower, "4k") || strings.Contains(lower, "uhd") {
		quality.Resolution = &mediamodels.Resolution{Width: 3840, Height: 2160}
		quality.QualityScore = 100
		profile := "4K/UHD"
		quality.QualityProfile = &profile
	} else if strings.Contains(lower, "1080p") || strings.Contains(lower, "fhd") {
		quality.Resolution = &mediamodels.Resolution{Width: 1920, Height: 1080}
		quality.QualityScore = 80
		profile := "1080p"
		quality.QualityProfile = &profile
	} else if strings.Contains(lower, "720p") || strings.Contains(lower, "hd") {
		quality.Resolution = &mediamodels.Resolution{Width: 1280, Height: 720}
		quality.QualityScore = 60
		profile := "720p"
		quality.QualityProfile = &profile
	} else if strings.Contains(lower, "480p") || strings.Contains(lower, "dvd") {
		quality.Resolution = &mediamodels.Resolution{Width: 720, Height: 480}
		quality.QualityScore = 40
		profile := "480p/DVD"
		quality.QualityProfile = &profile
	}

	// Source detection
	if strings.Contains(lower, "bluray") || strings.Contains(lower, "brrip") {
		source := "BluRay"
		quality.Source = &source
		quality.QualityScore += 10
	} else if strings.Contains(lower, "webdl") || strings.Contains(lower, "web-dl") {
		source := "WEB-DL"
		quality.Source = &source
		quality.QualityScore += 5
	} else if strings.Contains(lower, "webrip") {
		source := "WEB-RIP"
		quality.Source = &source
	}

	// Codec detection
	if strings.Contains(lower, "x265") || strings.Contains(lower, "h265") || strings.Contains(lower, "hevc") {
		codec := "H.265/HEVC"
		quality.VideoCodec = &codec
		quality.QualityScore += 5
	} else if strings.Contains(lower, "x264") || strings.Contains(lower, "h264") || strings.Contains(lower, "avc") {
		codec := "H.264/AVC"
		quality.VideoCodec = &codec
	}

	// Audio codec detection
	if strings.Contains(lower, "dts") {
		audioCodec := "DTS"
		quality.AudioCodec = &audioCodec
		quality.QualityScore += 5
	} else if strings.Contains(lower, "aac") {
		audioCodec := "AAC"
		quality.AudioCodec = &audioCodec
	} else if strings.Contains(lower, "ac3") {
		audioCodec := "AC3"
		quality.AudioCodec = &audioCodec
	}

	// HDR detection
	if strings.Contains(lower, "hdr") || strings.Contains(lower, "dolby.vision") {
		quality.HDR = true
		quality.QualityScore += 10
	}

	// For audio files
	if extension != nil {
		ext := strings.ToLower(*extension)
		if ext == ".flac" || ext == ".wav" {
			quality.QualityScore = 90
			profile := "Audio_Lossless"
			quality.QualityProfile = &profile
		} else if strings.Contains(lower, "320") || strings.Contains(lower, "320k") {
			quality.QualityScore = 70
			profile := "Audio_320k"
			quality.QualityProfile = &profile
		} else if ext == ".mp3" {
			quality.QualityScore = 50
			profile := "Audio_128k"
			quality.QualityProfile = &profile
		}
	}

	return quality
}

// updateMediaFiles creates or updates media file records
func (ma *MediaAnalyzer) updateMediaFiles(mediaItemID int64, files []models.FileInfo, directoryPath, smbRoot string) ([]mediamodels.MediaFile, error) {
	var updatedFiles []mediamodels.MediaFile

	for _, file := range files {
		if file.IsDirectory {
			continue
		}

		// Extract quality info
		qualityInfo := ma.extractQualityFromFilename(file.Name, file.Extension)
		qualityJSON, _ := json.Marshal(qualityInfo)

		// Generate SMB links
		directSmbLink := fmt.Sprintf("smb://%s/%s", smbRoot, file.Path)
		virtualSmbLink := fmt.Sprintf("virtual://%s/%d", smbRoot, file.ID)

		query := `
			INSERT OR REPLACE INTO media_files
			(media_item_id, file_path, smb_root, filename, file_size, file_extension, quality_info,
			 direct_smb_link, virtual_smb_link, last_verified, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err := ma.db.Exec(query,
			mediaItemID, file.Path, smbRoot, file.Name, file.Size, file.Extension,
			string(qualityJSON), directSmbLink, virtualSmbLink, time.Now(), time.Now(),
		)
		if err != nil {
			ma.logger.Error("Failed to update media file", zap.Error(err))
			continue
		}

		mediaFile := mediamodels.MediaFile{
			MediaItemID:    mediaItemID,
			FilePath:       file.Path,
			SmbRoot:        smbRoot,
			Filename:       file.Name,
			FileSize:       file.Size,
			FileExtension:  file.Extension,
			QualityInfo:    qualityInfo,
			DirectSmbLink:  directSmbLink,
			VirtualSmbLink: &virtualSmbLink,
			LastVerified:   time.Now(),
			CreatedAt:      time.Now(),
		}

		updatedFiles = append(updatedFiles, mediaFile)
	}

	return updatedFiles, nil
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
