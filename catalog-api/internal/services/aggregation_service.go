package services

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"catalogizer/database"
	"catalogizer/internal/media/models"
	"catalogizer/repository"

	"go.uber.org/zap"
)

// gamePlatformRe detects game platform hints in directory names.
// Kept here after title_parser.go was refactored to use digital.vasic.entities.
var gamePlatformRe = regexp.MustCompile(`(?i)\b(?:PC|Windows|Linux|Mac|macOS|PS[2-5]|PlayStation[\s._-]*[2-5]?|Xbox(?:[\s._-]*(?:One|360|Series[\s._-]*[XS]))?|Switch|Nintendo|GOG|Steam)\b`)

// AggregationService bridges the scan pipeline to structured media entities.
// After a scan completes, it analyzes scanned directories, detects media types,
// creates/updates MediaItem entities, and links files to them.
type AggregationService struct {
	db              *database.DB
	logger          *zap.Logger
	itemRepo        *repository.MediaItemRepository
	fileRepo        *repository.MediaFileRepository
	dirAnalysisRepo *repository.DirectoryAnalysisRepository
	extMetaRepo     *repository.ExternalMetadataRepository
}

// NewAggregationService creates a new aggregation service.
func NewAggregationService(
	db *database.DB,
	logger *zap.Logger,
	itemRepo *repository.MediaItemRepository,
	fileRepo *repository.MediaFileRepository,
	dirAnalysisRepo *repository.DirectoryAnalysisRepository,
	extMetaRepo *repository.ExternalMetadataRepository,
) *AggregationService {
	return &AggregationService{
		db:              db,
		logger:          logger,
		itemRepo:        itemRepo,
		fileRepo:        fileRepo,
		dirAnalysisRepo: dirAnalysisRepo,
		extMetaRepo:     extMetaRepo,
	}
}

// AggregateAfterScan runs after a scan completes to create entities from files.
func (s *AggregationService) AggregateAfterScan(ctx context.Context, storageRootID int64) error {
	s.logger.Info("Starting post-scan aggregation", zap.Int64("storage_root_id", storageRootID))

	// Get top-level directories from the scanned storage root
	dirs, err := s.getTopLevelDirectories(ctx, storageRootID)
	if err != nil {
		return fmt.Errorf("get top-level directories: %w", err)
	}

	s.logger.Info("Found top-level directories to analyze",
		zap.Int("count", len(dirs)),
		zap.Int64("storage_root_id", storageRootID))

	created, updated := 0, 0
	for _, dir := range dirs {
		isNew, err := s.processDirectory(ctx, dir, storageRootID)
		if err != nil {
			s.logger.Warn("Failed to process directory",
				zap.String("path", dir.path),
				zap.Error(err))
			continue
		}
		if isNew {
			created++
		} else {
			updated++
		}
	}

	s.logger.Info("Post-scan aggregation completed",
		zap.Int64("storage_root_id", storageRootID),
		zap.Int("entities_created", created),
		zap.Int("entities_updated", updated))

	return nil
}

type directoryInfo struct {
	path       string
	name       string
	fileCount  int
	totalSize  int64
	fileIDs    []int64
	fileTypes  map[string]int
	extensions []string
}

// getTopLevelDirectories returns top-level directories and their file stats.
func (s *AggregationService) getTopLevelDirectories(ctx context.Context, storageRootID int64) ([]directoryInfo, error) {
	// Get directories (parent_id IS NULL means top-level)
	query := `SELECT id, path, name FROM files
		WHERE storage_root_id = ? AND is_directory = 1 AND deleted = 0 AND parent_id IS NULL
		ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query, storageRootID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dirs []directoryInfo
	for rows.Next() {
		var id int64
		var dir directoryInfo
		if err := rows.Scan(&id, &dir.path, &dir.name); err != nil {
			return nil, err
		}

		// Get child files for this directory
		childQuery := `SELECT id, extension, size FROM files
			WHERE storage_root_id = ? AND parent_id = ? AND is_directory = 0 AND deleted = 0`

		childRows, err := s.db.QueryContext(ctx, childQuery, storageRootID, id)
		if err != nil {
			continue
		}

		dir.fileTypes = make(map[string]int)
		for childRows.Next() {
			var fileID, size int64
			var ext *string
			if err := childRows.Scan(&fileID, &ext, &size); err != nil {
				continue
			}
			dir.fileIDs = append(dir.fileIDs, fileID)
			dir.totalSize += size
			dir.fileCount++
			if ext != nil {
				dir.fileTypes[*ext]++
				dir.extensions = append(dir.extensions, *ext)
			}
		}
		childRows.Close()

		if dir.fileCount > 0 {
			dirs = append(dirs, dir)
		}
	}

	return dirs, rows.Err()
}

// processDirectory analyzes a directory and creates/updates a media entity.
// Returns true if a new entity was created.
func (s *AggregationService) processDirectory(ctx context.Context, dir directoryInfo, storageRootID int64) (bool, error) {
	// Detect media type from directory name and file types
	mediaTypeName, parsed := s.detectMediaType(dir)
	if mediaTypeName == "" {
		return false, nil // Couldn't determine type
	}

	// Get media type ID
	_, typeID, err := s.itemRepo.GetMediaTypeByName(ctx, mediaTypeName)
	if err != nil {
		return false, fmt.Errorf("get media type %q: %w", mediaTypeName, err)
	}

	// Check if entity already exists
	existing, err := s.itemRepo.GetByTitle(ctx, parsed.Title, typeID)
	if err != nil {
		return false, err
	}

	var itemID int64
	isNew := false

	if existing != nil {
		itemID = existing.ID
		// Update if needed
		if existing.Year == nil && parsed.Year != nil {
			existing.Year = parsed.Year
			_ = s.itemRepo.Update(ctx, existing)
		}
	} else {
		// Create new entity
		item := &models.MediaItem{
			MediaTypeID: typeID,
			Title:       parsed.Title,
			Year:        parsed.Year,
			Status:      "detected",
		}
		itemID, err = s.itemRepo.Create(ctx, item)
		if err != nil {
			return false, fmt.Errorf("create media item: %w", err)
		}
		isNew = true
	}

	// Link files to entity
	for i, fileID := range dir.fileIDs {
		isPrimary := i == 0 // First file is primary
		_, err := s.fileRepo.LinkFileToItem(ctx, itemID, fileID, nil, nil, isPrimary)
		if err != nil {
			s.logger.Warn("Failed to link file to entity",
				zap.Int64("file_id", fileID),
				zap.Int64("media_item_id", itemID),
				zap.Error(err))
		}
	}

	// Store directory analysis
	confidence := 0.5
	if parsed.Year != nil {
		confidence = 0.8
	}
	if len(parsed.QualityHints) > 0 {
		confidence = 0.9
	}

	da := &models.DirectoryAnalysis{
		DirectoryPath:   dir.path,
		MediaItemID:     &itemID,
		ConfidenceScore: confidence,
		DetectionMethod: "title_parser",
		FilesCount:      dir.fileCount,
		TotalSize:       dir.totalSize,
	}

	existingDA, _ := s.dirAnalysisRepo.GetByPath(ctx, dir.path)
	if existingDA != nil {
		da.ID = existingDA.ID
		_ = s.dirAnalysisRepo.Update(ctx, da)
	} else {
		_, _ = s.dirAnalysisRepo.Create(ctx, da)
	}

	// Build hierarchy for TV shows
	if mediaTypeName == "tv_show" && parsed.Season != nil {
		s.buildTVHierarchy(ctx, itemID, typeID, parsed)
	}

	return isNew, nil
}

// detectMediaType determines the media type from directory info and filename.
func (s *AggregationService) detectMediaType(dir directoryInfo) (string, ParsedTitle) {
	name := dir.name

	// Check file extensions to help classify
	hasVideo := false
	hasAudio := false
	hasISO := false
	hasEbook := false
	hasComic := false
	hasExecutable := false

	videoExts := map[string]bool{".mkv": true, ".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true, ".m4v": true, ".ts": true}
	audioExts := map[string]bool{".mp3": true, ".flac": true, ".wav": true, ".aac": true, ".ogg": true, ".m4a": true, ".wma": true, ".ape": true}
	isoExts := map[string]bool{".iso": true, ".img": true, ".bin": true, ".nrg": true}
	ebookExts := map[string]bool{".epub": true, ".mobi": true, ".azw3": true, ".pdf": true, ".djvu": true}
	comicExts := map[string]bool{".cbr": true, ".cbz": true, ".cb7": true}
	exeExts := map[string]bool{".exe": true, ".msi": true, ".dmg": true, ".deb": true, ".rpm": true, ".AppImage": true}

	for ext, count := range dir.fileTypes {
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if videoExts[ext] && count > 0 {
			hasVideo = true
		}
		if audioExts[ext] && count > 0 {
			hasAudio = true
		}
		if isoExts[ext] && count > 0 {
			hasISO = true
		}
		if ebookExts[ext] && count > 0 {
			hasEbook = true
		}
		if comicExts[ext] && count > 0 {
			hasComic = true
		}
		if exeExts[ext] && count > 0 {
			hasExecutable = true
		}
	}

	// TV show detection first (specific pattern)
	tvParsed := ParseTVShow(name)
	if tvParsed.Season != nil || tvParsed.Episode != nil {
		return "tv_show", tvParsed
	}
	nameLower := strings.ToLower(name)
	if strings.Contains(nameLower, "season") || strings.Contains(nameLower, "complete") ||
		strings.Contains(nameLower, "s01") || strings.Contains(nameLower, "s02") {
		return "tv_show", ParseTVShow(name)
	}

	// Comic detection
	if hasComic {
		return "comic", ParsedTitle{Title: CleanTitle(name)}
	}

	// Ebook detection
	if hasEbook && !hasVideo && !hasAudio {
		return "book", ParsedTitle{Title: CleanTitle(name)}
	}

	// Music detection
	if hasAudio && !hasVideo {
		return "music_album", ParseMusicAlbum(name)
	}

	// Software detection
	if hasExecutable || hasISO {
		return "software", ParseSoftwareTitle(name)
	}

	// Game detection (ISO with game-like names, or directory names with platform)
	if hasISO && gamePlatformRe.MatchString(name) {
		return "game", ParseGameTitle(name)
	}

	// Default: video content = movie
	if hasVideo {
		return "movie", ParseMovieTitle(name)
	}

	// Try to parse as movie from name alone
	parsed := ParseMovieTitle(name)
	if parsed.Year != nil {
		return "movie", parsed
	}

	return "", ParsedTitle{}
}

// buildTVHierarchy creates season and episode entities under a TV show.
func (s *AggregationService) buildTVHierarchy(ctx context.Context, showID, showTypeID int64, parsed ParsedTitle) {
	// Get season type
	_, seasonTypeID, err := s.itemRepo.GetMediaTypeByName(ctx, "tv_season")
	if err != nil {
		return
	}

	if parsed.Season != nil {
		seasonTitle := fmt.Sprintf("Season %d", *parsed.Season)
		existingSeason, _ := s.itemRepo.GetByTitle(ctx, seasonTitle, seasonTypeID)

		var seasonID int64
		if existingSeason != nil && existingSeason.ParentID != nil && *existingSeason.ParentID == showID {
			seasonID = existingSeason.ID
		} else {
			season := &models.MediaItem{
				MediaTypeID:  seasonTypeID,
				Title:        seasonTitle,
				Status:       "detected",
				ParentID:     &showID,
				SeasonNumber: parsed.Season,
			}
			seasonID, err = s.itemRepo.Create(ctx, season)
			if err != nil {
				return
			}
		}

		// Create episode if we have one
		if parsed.Episode != nil {
			_, epTypeID, err := s.itemRepo.GetMediaTypeByName(ctx, "tv_episode")
			if err != nil {
				return
			}

			epTitle := fmt.Sprintf("Episode %d", *parsed.Episode)
			ep := &models.MediaItem{
				MediaTypeID:   epTypeID,
				Title:         epTitle,
				Status:        "detected",
				ParentID:      &seasonID,
				SeasonNumber:  parsed.Season,
				EpisodeNumber: parsed.Episode,
			}
			_, _ = s.itemRepo.Create(ctx, ep)
		}
	}
}

// GetStorageRootName returns the storage root name for display.
func (s *AggregationService) getStorageRootName(ctx context.Context, storageRootID int64) string {
	var name string
	_ = s.db.QueryRowContext(ctx, "SELECT name FROM storage_roots WHERE id = ?", storageRootID).Scan(&name)
	return name
}

// DetectMediaTypeFromPath analyzes a file path and returns the likely media type.
func DetectMediaTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mkv", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".m4v":
		return "video"
	case ".mp3", ".flac", ".wav", ".aac", ".ogg", ".m4a", ".wma":
		return "audio"
	case ".iso", ".img", ".bin":
		return "disc_image"
	case ".epub", ".mobi", ".azw3":
		return "ebook"
	case ".cbr", ".cbz", ".cb7":
		return "comic"
	case ".exe", ".msi", ".dmg", ".deb", ".rpm":
		return "software"
	default:
		return "other"
	}
}
