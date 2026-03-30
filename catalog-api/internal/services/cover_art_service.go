package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"catalogizer/database"

	"go.uber.org/zap"
	"golang.org/x/image/draw"
)

// CoverArtService handles cover art retrieval, processing, and caching
type CoverArtService struct {
	db         *database.DB
	logger     *zap.Logger
	httpClient *http.Client
	apiKeys    map[string]string
	cacheDir   string
}

// CoverArtProvider represents different cover art providers
type CoverArtProvider string

const (
	CoverArtProviderMusicBrainz CoverArtProvider = "musicbrainz"
	CoverArtProviderLastFM      CoverArtProvider = "lastfm"
	CoverArtProviderSpotify     CoverArtProvider = "spotify"
	CoverArtProviderDeezer      CoverArtProvider = "deezer"
	CoverArtProviderITunes      CoverArtProvider = "itunes"
	CoverArtProviderDiscogs     CoverArtProvider = "discogs"
	CoverArtProviderEmbedded    CoverArtProvider = "embedded"
	CoverArtProviderLocal       CoverArtProvider = "local"
)

// CoverArtQuality represents different quality levels
type CoverArtQuality string

const (
	QualityThumbnail CoverArtQuality = "thumbnail" // 150x150
	QualityMedium    CoverArtQuality = "medium"    // 300x300
	QualityHigh      CoverArtQuality = "high"      // 600x600
	QualityOriginal  CoverArtQuality = "original"  // Original size
)

// CoverArtSearchRequest represents a cover art search request
type CoverArtSearchRequest struct {
	Title         string             `json:"title"`
	Artist        string             `json:"artist"`
	Album         *string            `json:"album,omitempty"`
	Year          *int               `json:"year,omitempty"`
	MusicBrainzID *string            `json:"musicbrainz_id,omitempty"`
	SpotifyID     *string            `json:"spotify_id,omitempty"`
	Quality       CoverArtQuality    `json:"quality"`
	Providers     []CoverArtProvider `json:"providers,omitempty"`
	UseCache      bool               `json:"use_cache"`
}

// CoverArtSearchResult represents a cover art search result
type CoverArtSearchResult struct {
	ID           string           `json:"id"`
	Provider     CoverArtProvider `json:"provider"`
	Title        string           `json:"title"`
	Artist       string           `json:"artist"`
	Album        *string          `json:"album,omitempty"`
	URL          string           `json:"url"`
	ThumbnailURL *string          `json:"thumbnail_url,omitempty"`
	Width        int              `json:"width"`
	Height       int              `json:"height"`
	Format       string           `json:"format"`
	Quality      CoverArtQuality  `json:"quality"`
	Size         *int64           `json:"size,omitempty"`
	MatchScore   float64          `json:"match_score"`
	Copyright    *string          `json:"copyright,omitempty"`
	Source       string           `json:"source"`
}

// CoverArtDownloadRequest represents a cover art download request
type CoverArtDownloadRequest struct {
	MediaItemID   int64             `json:"media_item_id"`
	ResultID      string            `json:"result_id"`
	Quality       CoverArtQuality   `json:"quality"`
	GenerateSizes []CoverArtQuality `json:"generate_sizes,omitempty"`
	SetAsDefault  bool              `json:"set_as_default"`
}

// VideoThumbnailRequest represents a video thumbnail generation request
type VideoThumbnailRequest struct {
	MediaItemID   int64             `json:"media_item_id"`
	VideoPath     string            `json:"video_path"`
	Timestamps    []float64         `json:"timestamps,omitempty"` // Seconds
	Quality       CoverArtQuality   `json:"quality"`
	GenerateSizes []CoverArtQuality `json:"generate_sizes,omitempty"`
	Count         int               `json:"count"` // Number of thumbnails to generate
}

// LocalCoverArtScanRequest represents a request to scan for local cover art
type LocalCoverArtScanRequest struct {
	MediaItemID int64  `json:"media_item_id"`
	Directory   string `json:"directory"`
	Recursive   bool   `json:"recursive"`
}

// CoverArtProcessingOptions represents image processing options
type CoverArtProcessingOptions struct {
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	Quality         int     `json:"quality"`                    // JPEG quality 1-100
	Format          string  `json:"format"`                     // "jpeg", "png", "webp"
	Crop            bool    `json:"crop"`                       // Crop to exact dimensions
	PreserveAspect  bool    `json:"preserve_aspect"`            // Preserve aspect ratio
	BackgroundColor *string `json:"background_color,omitempty"` // Hex color for padding
}

// NewCoverArtService creates a new cover art service
func NewCoverArtService(db *database.DB, logger *zap.Logger) *CoverArtService {
	return &CoverArtService{
		db:         db,
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKeys:    make(map[string]string),
		cacheDir:   "./cache/cover_art",
	}
}

// SearchCoverArt searches for cover art across multiple providers
func (s *CoverArtService) SearchCoverArt(ctx context.Context, request *CoverArtSearchRequest) ([]CoverArtSearchResult, error) {
	s.logger.Info("Searching cover art",
		zap.String("title", request.Title),
		zap.String("artist", request.Artist),
		zap.String("album", getStringValue(request.Album)))

	// Check cache first if requested
	if request.UseCache {
		if cached := s.getCachedCoverArt(ctx, request); cached != nil {
			return []CoverArtSearchResult{*cached}, nil
		}
	}

	var allResults []CoverArtSearchResult

	// Default providers if none specified
	providers := request.Providers
	if len(providers) == 0 {
		providers = []CoverArtProvider{
			CoverArtProviderMusicBrainz,
			CoverArtProviderLastFM,
			CoverArtProviderITunes,
		}
	}

	// Search each provider
	for _, provider := range providers {
		results, err := s.searchProvider(ctx, provider, request)
		if err != nil {
			s.logger.Warn("Provider search failed",
				zap.String("provider", string(provider)),
				zap.Error(err))
			continue
		}
		allResults = append(allResults, results...)
	}

	// Sort by match score and quality
	s.sortCoverArtResults(allResults)

	s.logger.Info("Cover art search completed",
		zap.Int("total_results", len(allResults)))

	return allResults, nil
}

// DownloadCoverArt downloads and processes cover art
func (s *CoverArtService) DownloadCoverArt(ctx context.Context, request *CoverArtDownloadRequest) (*CoverArt, error) {
	s.logger.Info("Downloading cover art",
		zap.Int64("media_item_id", request.MediaItemID),
		zap.String("result_id", request.ResultID))

	// Get download info
	result, err := s.getCoverArtDownloadInfo(ctx, request.ResultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get download info: %w", err)
	}

	// Download image
	imageData, err := s.downloadImage(ctx, result.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	// Process and save image
	coverArt, err := s.processAndSaveCoverArt(ctx, request.MediaItemID, imageData, result, request)
	if err != nil {
		return nil, fmt.Errorf("failed to process cover art: %w", err)
	}

	// Generate additional sizes if requested
	if len(request.GenerateSizes) > 0 {
		go s.generateAdditionalSizes(ctx, coverArt, request.GenerateSizes)
	}

	// Set as default if requested
	if request.SetAsDefault {
		if err := s.setDefaultCoverArt(ctx, request.MediaItemID, coverArt.ID); err != nil {
			s.logger.Warn("Failed to set as default cover art", zap.Error(err))
		}
	}

	return coverArt, nil
}

// GenerateVideoThumbnails generates thumbnails for video files
func (s *CoverArtService) GenerateVideoThumbnails(ctx context.Context, request *VideoThumbnailRequest) ([]*CoverArt, error) {
	s.logger.Info("Generating video thumbnails",
		zap.Int64("media_item_id", request.MediaItemID),
		zap.String("video_path", request.VideoPath))

	// Get video duration
	duration, err := s.getVideoDuration(request.VideoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video duration: %w", err)
	}

	// Generate timestamps if not provided
	timestamps := request.Timestamps
	if len(timestamps) == 0 {
		timestamps = s.generateTimestamps(duration, request.Count)
	}

	var thumbnails []*CoverArt

	// Generate thumbnail for each timestamp
	for i, timestamp := range timestamps {
		thumbnail, err := s.generateVideoThumbnail(ctx, request, timestamp, i)
		if err != nil {
			s.logger.Warn("Failed to generate thumbnail",
				zap.Float64("timestamp", timestamp),
				zap.Error(err))
			continue
		}
		thumbnails = append(thumbnails, thumbnail)
	}

	return thumbnails, nil
}

// ScanLocalCoverArt scans directory for local cover art files
func (s *CoverArtService) ScanLocalCoverArt(ctx context.Context, request *LocalCoverArtScanRequest) ([]*CoverArt, error) {
	s.logger.Info("Scanning local cover art",
		zap.Int64("media_item_id", request.MediaItemID),
		zap.String("directory", request.Directory))

	var coverArts []*CoverArt

	// Common cover art filenames
	coverFilenames := []string{
		"cover.jpg", "cover.jpeg", "cover.png",
		"folder.jpg", "folder.jpeg", "folder.png",
		"album.jpg", "album.jpeg", "album.png",
		"front.jpg", "front.jpeg", "front.png",
		"albumart.jpg", "albumart.jpeg", "albumart.png",
	}

	// Scan directory
	files, err := os.ReadDir(request.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() && request.Recursive {
			// Recursively scan subdirectories
			subRequest := *request
			subRequest.Directory = filepath.Join(request.Directory, file.Name())
			subCoverArts, err := s.ScanLocalCoverArt(ctx, &subRequest)
			if err != nil {
				s.logger.Warn("Failed to scan subdirectory", zap.Error(err))
				continue
			}
			coverArts = append(coverArts, subCoverArts...)
			continue
		}

		// Check if file is a potential cover art
		filename := strings.ToLower(file.Name())
		for _, coverFilename := range coverFilenames {
			if filename == coverFilename || strings.HasPrefix(filename, strings.TrimSuffix(coverFilename, filepath.Ext(coverFilename))) {
				filePath := filepath.Join(request.Directory, file.Name())
				coverArt, err := s.processLocalCoverArt(ctx, request.MediaItemID, filePath)
				if err != nil {
					s.logger.Warn("Failed to process local cover art",
						zap.String("file", filePath),
						zap.Error(err))
					continue
				}
				coverArts = append(coverArts, coverArt)
				break
			}
		}
	}

	return coverArts, nil
}

// GetCoverArt returns cover art for a media item
func (s *CoverArtService) GetCoverArt(ctx context.Context, mediaItemID int64) (*CoverArt, error) {
	if s.db == nil {
		return nil, nil
	}

	query := `
		SELECT id, media_item_id, source, url, local_path, width, height,
		       format, size, quality, created_at, cached_at
		FROM cover_art WHERE media_item_id = ? AND is_default = 1
		ORDER BY created_at DESC LIMIT 1`

	var coverArt CoverArt
	var url, localPath sql.NullString
	var size sql.NullInt64
	var cachedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, mediaItemID).Scan(
		&coverArt.ID, &coverArt.MediaItemID, &coverArt.Source,
		&url, &localPath, &coverArt.Width, &coverArt.Height,
		&coverArt.Format, &size, &coverArt.Quality,
		&coverArt.CreatedAt, &cachedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No cover art found
		}
		return nil, fmt.Errorf("failed to get cover art: %w", err)
	}

	if url.Valid {
		coverArt.URL = &url.String
	}
	if localPath.Valid {
		coverArt.LocalPath = &localPath.String
	}
	if size.Valid {
		coverArt.Size = &size.Int64
	}
	if cachedAt.Valid {
		coverArt.CachedAt = &cachedAt.Time
	}

	return &coverArt, nil
}

// ProcessImage processes an image with specified options
func (s *CoverArtService) ProcessImage(inputPath string, outputPath string, options *CoverArtProcessingOptions) error {
	s.logger.Debug("Processing image",
		zap.String("input", inputPath),
		zap.String("output", outputPath))

	// Open input image
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Decode image
	img, format, err := image.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image
	resizedImg := s.resizeImage(img, options)

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Encode image
	switch strings.ToLower(options.Format) {
	case "jpeg", "jpg":
		return jpeg.Encode(outputFile, resizedImg, &jpeg.Options{Quality: options.Quality})
	case "png":
		return png.Encode(outputFile, resizedImg)
	default:
		// Default to original format
		switch format {
		case "jpeg":
			return jpeg.Encode(outputFile, resizedImg, &jpeg.Options{Quality: options.Quality})
		case "png":
			return png.Encode(outputFile, resizedImg)
		default:
			return jpeg.Encode(outputFile, resizedImg, &jpeg.Options{Quality: options.Quality})
		}
	}
}

// Provider-specific implementations
func (s *CoverArtService) searchProvider(ctx context.Context, provider CoverArtProvider, request *CoverArtSearchRequest) ([]CoverArtSearchResult, error) {
	switch provider {
	case CoverArtProviderMusicBrainz:
		return s.searchMusicBrainz(ctx, request)
	case CoverArtProviderLastFM:
		return s.searchLastFM(ctx, request)
	case CoverArtProviderITunes:
		return s.searchITunes(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (s *CoverArtService) searchMusicBrainz(ctx context.Context, request *CoverArtSearchRequest) ([]CoverArtSearchResult, error) {
	s.logger.Debug("Searching MusicBrainz",
		zap.String("artist", request.Artist),
		zap.String("album", getStringValue(request.Album)))

	// Mock implementation for demonstration
	result := CoverArtSearchResult{
		ID:         "mb_1",
		Provider:   CoverArtProviderMusicBrainz,
		Title:      request.Title,
		Artist:     request.Artist,
		Album:      request.Album,
		URL:        "https://coverartarchive.org/sample/front.jpg",
		Width:      500,
		Height:     500,
		Format:     "jpeg",
		Quality:    QualityHigh,
		MatchScore: 0.9,
		Source:     "coverartarchive.org",
	}

	return []CoverArtSearchResult{result}, nil
}

func (s *CoverArtService) searchLastFM(ctx context.Context, request *CoverArtSearchRequest) ([]CoverArtSearchResult, error) {
	s.logger.Debug("Searching Last.FM")

	result := CoverArtSearchResult{
		ID:         "lastfm_1",
		Provider:   CoverArtProviderLastFM,
		Title:      request.Title,
		Artist:     request.Artist,
		Album:      request.Album,
		URL:        "https://lastfm-img2.akamaized.net/sample.jpg",
		Width:      300,
		Height:     300,
		Format:     "jpeg",
		Quality:    QualityMedium,
		MatchScore: 0.85,
		Source:     "last.fm",
	}

	return []CoverArtSearchResult{result}, nil
}

func (s *CoverArtService) searchITunes(ctx context.Context, request *CoverArtSearchRequest) ([]CoverArtSearchResult, error) {
	s.logger.Debug("Searching iTunes")

	result := CoverArtSearchResult{
		ID:         "itunes_1",
		Provider:   CoverArtProviderITunes,
		Title:      request.Title,
		Artist:     request.Artist,
		Album:      request.Album,
		URL:        "https://is1-ssl.mzstatic.com/sample.jpg",
		Width:      600,
		Height:     600,
		Format:     "jpeg",
		Quality:    QualityHigh,
		MatchScore: 0.92,
		Source:     "itunes.apple.com",
	}

	return []CoverArtSearchResult{result}, nil
}

// Helper functions
func (s *CoverArtService) downloadImage(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (s *CoverArtService) resizeImage(img image.Image, options *CoverArtProcessingOptions) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	dstWidth := options.Width
	dstHeight := options.Height

	if options.PreserveAspect {
		// Calculate dimensions preserving aspect ratio
		aspectRatio := float64(srcWidth) / float64(srcHeight)
		if float64(dstWidth)/float64(dstHeight) > aspectRatio {
			dstWidth = int(float64(dstHeight) * aspectRatio)
		} else {
			dstHeight = int(float64(dstWidth) / aspectRatio)
		}
	}

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))

	// Resize image
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}

func (s *CoverArtService) generateTimestamps(duration float64, count int) []float64 {
	if count <= 0 {
		count = 3 // Default to 3 thumbnails
	}

	var timestamps []float64
	interval := duration / float64(count+1)

	for i := 1; i <= count; i++ {
		timestamps = append(timestamps, interval*float64(i))
	}

	return timestamps
}

func (s *CoverArtService) generateCoverArtID() string {
	return fmt.Sprintf("cover_%d", time.Now().UnixNano())
}

func (s *CoverArtService) generateCacheKey(request *CoverArtSearchRequest) string {
	data := fmt.Sprintf("%s_%s_%s_%s", request.Artist, request.Title,
		getStringValue(request.Album), request.Quality)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *CoverArtService) sortCoverArtResults(results []CoverArtSearchResult) {
	// Sort by match score descending, then by quality/size
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].MatchScore < results[j].MatchScore ||
				(results[i].MatchScore == results[j].MatchScore && results[i].Width < results[j].Width) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// getCachedCoverArt retrieves cached cover art from the database
func (s *CoverArtService) getCachedCoverArt(ctx context.Context, request *CoverArtSearchRequest) *CoverArtSearchResult {
	// Generate cache key based on request parameters
	cacheKey := s.generateCacheKey(request)

	query := `
		SELECT id, provider, title, artist, album, url, thumbnail_url, width, height,
		       format, quality, size, match_score, source, created_at
		FROM cover_art_cache
		WHERE cache_key = ? AND created_at > ?
		ORDER BY match_score DESC
		LIMIT 1
	`

	// Cache valid for 30 days
	cacheExpiry := time.Now().Add(-30 * 24 * time.Hour)

	var result CoverArtSearchResult
	var album sql.NullString
	var thumbnailURL sql.NullString
	var size sql.NullInt64
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, query, cacheKey, cacheExpiry).Scan(
		&result.ID, &result.Provider, &result.Title, &result.Artist,
		&album, &result.URL, &thumbnailURL, &result.Width, &result.Height,
		&result.Format, &result.Quality, &size, &result.MatchScore,
		&result.Source, &createdAt,
	)

	if err != nil {
		if err != sql.ErrNoRows {
			s.logger.Warn("Failed to get cached cover art", zap.Error(err))
		}
		return nil
	}

	if album.Valid {
		result.Album = &album.String
	}
	if thumbnailURL.Valid {
		result.ThumbnailURL = &thumbnailURL.String
	}
	if size.Valid {
		result.Size = &size.Int64
	}

	return &result
}

// getCoverArtDownloadInfo retrieves cover art download information
func (s *CoverArtService) getCoverArtDownloadInfo(ctx context.Context, resultID string) (*CoverArtSearchResult, error) {
	if s.db != nil {
		var url, provider, quality string
		err := s.db.QueryRowContext(ctx,
			"SELECT url, source, quality FROM cover_art WHERE id = ?", resultID,
		).Scan(&url, &provider, &quality)
		if err == nil {
			return &CoverArtSearchResult{
				ID:       resultID,
				Provider: CoverArtProvider(provider),
				URL:      url,
				Quality:  CoverArtQuality(quality),
				Source:   provider,
			}, nil
		}
	}

	return &CoverArtSearchResult{
		ID:       resultID,
		Provider: CoverArtProviderLocal,
		URL:      "",
		Quality:  QualityHigh,
		Source:   "cache",
	}, nil
}

// processAndSaveCoverArt processes and saves cover art
func (s *CoverArtService) processAndSaveCoverArt(ctx context.Context, mediaItemID int64, imageData []byte, result *CoverArtSearchResult, request *CoverArtDownloadRequest) (*CoverArt, error) {
	// Decode image
	img, format, err := image.Decode(strings.NewReader(string(imageData)))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Generate local path
	coverID := s.generateCoverArtID()
	filename := fmt.Sprintf("%s.%s", coverID, format)
	localPath := filepath.Join(s.cacheDir, filename)

	// Ensure cache directory exists
	if err := os.MkdirAll(s.cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Save image
	outFile, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(outFile, img)
	default:
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	// Get file size
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	size := fileInfo.Size()

	// Create CoverArt record
	now := time.Now()
	coverArt := &CoverArt{
		ID:          coverID,
		MediaItemID: mediaItemID,
		Source:      string(result.Provider),
		LocalPath:   &localPath,
		Width:       &width,
		Height:      &height,
		Format:      format,
		Size:        &size,
		Quality:     string(request.Quality),
		CreatedAt:   now,
		CachedAt:    &now,
	}

	return coverArt, nil
}

// generateAdditionalSizes generates additional sizes of cover art
func (s *CoverArtService) generateAdditionalSizes(ctx context.Context, coverArt *CoverArt, sizes []CoverArtQuality) {
	if coverArt.LocalPath == nil {
		s.logger.Warn("Cannot generate additional sizes: no local path")
		return
	}

	// Open source image
	srcFile, err := os.Open(*coverArt.LocalPath)
	if err != nil {
		s.logger.Error("Failed to open source image", zap.Error(err))
		return
	}
	defer srcFile.Close()

	srcImg, _, err := image.Decode(srcFile)
	if err != nil {
		s.logger.Error("Failed to decode source image", zap.Error(err))
		return
	}

	// Generate each size
	for _, quality := range sizes {
		targetSize := 0
		switch quality {
		case QualityThumbnail:
			targetSize = 150
		case QualityMedium:
			targetSize = 300
		case QualityHigh:
			targetSize = 600
		default:
			continue
		}

		// Resize image
		newImg := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
		draw.ApproxBiLinear.Scale(newImg, newImg.Bounds(), srcImg, srcImg.Bounds(), draw.Over, nil)

		// Save resized image
		resizedFilename := fmt.Sprintf("%s_%s.%s", coverArt.ID, quality, coverArt.Format)
		resizedPath := filepath.Join(s.cacheDir, resizedFilename)

		outFile, err := os.Create(resizedPath)
		if err != nil {
			s.logger.Error("Failed to create resized file",
				zap.String("quality", string(quality)),
				zap.Error(err))
			continue
		}

		if coverArt.Format == "png" {
			err = png.Encode(outFile, newImg)
		} else {
			err = jpeg.Encode(outFile, newImg, &jpeg.Options{Quality: 85})
		}
		outFile.Close()

		if err != nil {
			s.logger.Error("Failed to encode resized image",
				zap.String("quality", string(quality)),
				zap.Error(err))
		}
	}
}

// setDefaultCoverArt sets cover art as default for a media item
func (s *CoverArtService) setDefaultCoverArt(ctx context.Context, mediaItemID int64, coverArtID string) error {
	// In a real implementation, this would update the database to mark the cover art as default
	query := `UPDATE cover_art SET is_default = 0 WHERE media_item_id = ?`
	_, err := s.db.ExecContext(ctx, query, mediaItemID)
	if err != nil {
		return err
	}

	query = `UPDATE cover_art SET is_default = 1 WHERE id = ? AND media_item_id = ?`
	_, err = s.db.ExecContext(ctx, query, coverArtID, mediaItemID)
	return err
}

// getVideoDuration gets the duration of a video file using ffprobe
func (s *CoverArtService) getVideoDuration(videoPath string) (float64, error) {
	output, err := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	).Output()
	if err == nil {
		if duration, parseErr := strconv.ParseFloat(strings.TrimSpace(string(output)), 64); parseErr == nil && duration > 0 {
			return duration, nil
		}
	}

	// Fallback: estimate from file size (assume ~5 Mbps bitrate)
	if info, statErr := os.Stat(videoPath); statErr == nil {
		estimated := float64(info.Size()) / (5 * 1024 * 1024 / 8)
		if estimated > 0 {
			return estimated, nil
		}
	}

	return 120.0, nil
}

// generateVideoThumbnail generates a thumbnail for a video at a specific timestamp
func (s *CoverArtService) generateVideoThumbnail(ctx context.Context, request *VideoThumbnailRequest, timestamp float64, index int) (*CoverArt, error) {
	coverID := fmt.Sprintf("video_thumb_%d_%d", request.MediaItemID, index)
	now := time.Now()

	outputPath := filepath.Join(s.cacheDir, coverID+".jpg")

	// Try ffmpeg to extract a frame
	err := exec.CommandContext(ctx, "ffmpeg",
		"-ss", fmt.Sprintf("%.2f", timestamp),
		"-i", request.VideoPath,
		"-vframes", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	).Run()

	if err == nil {
		if info, statErr := os.Stat(outputPath); statErr == nil {
			size := info.Size()
			return &CoverArt{
				ID:          coverID,
				MediaItemID: request.MediaItemID,
				Source:      "video_thumbnail",
				LocalPath:   &outputPath,
				Format:      "jpeg",
				Size:        &size,
				Quality:     string(request.Quality),
				CreatedAt:   now,
				CachedAt:    &now,
			}, nil
		}
	}

	// Fallback: return metadata-only cover art record
	return &CoverArt{
		ID:          coverID,
		MediaItemID: request.MediaItemID,
		Source:      "video_thumbnail",
		Quality:     string(request.Quality),
		CreatedAt:   now,
	}, nil
}

// processLocalCoverArt processes local cover art file
func (s *CoverArtService) processLocalCoverArt(ctx context.Context, mediaItemID int64, filePath string) (*CoverArt, error) {
	// Open and decode image
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	size := fileInfo.Size()

	// Create CoverArt record
	coverID := s.generateCoverArtID()
	now := time.Now()

	coverArt := &CoverArt{
		ID:          coverID,
		MediaItemID: mediaItemID,
		Source:      "local",
		LocalPath:   &filePath,
		Width:       &width,
		Height:      &height,
		Format:      format,
		Size:        &size,
		Quality:     "original",
		CreatedAt:   now,
	}

	return coverArt, nil
}

// GetCoverURL returns a cover image URL for any media item.
// Tries in order:
//  1. Cached cover art from cover_art table
//  2. External metadata cover URL (TMDB, local_scan, etc.)
//  3. Type-specific placeholder SVG
func (s *CoverArtService) GetCoverURL(ctx context.Context, mediaItemID int64, mediaTypeName string) string {
	// 1. Check if we have cover art in the cover_art table
	if s.db != nil {
		var localPath sql.NullString
		var coverURL sql.NullString
		err := s.db.QueryRowContext(ctx,
			`SELECT local_path, url FROM cover_art
			 WHERE media_item_id = ? AND is_default = 1
			 ORDER BY created_at DESC LIMIT 1`, mediaItemID).Scan(&localPath, &coverURL)
		if err == nil {
			if coverURL.Valid && coverURL.String != "" {
				return coverURL.String
			}
			if localPath.Valid && localPath.String != "" {
				return fmt.Sprintf("/api/v1/cover/%d", mediaItemID)
			}
		}
	}

	// 2. Check external_metadata for a cover URL (TMDB, local scan, etc.)
	if s.db != nil {
		var coverURL sql.NullString
		err := s.db.QueryRowContext(ctx,
			`SELECT cover_url FROM external_metadata
			 WHERE media_item_id = ? AND cover_url IS NOT NULL AND cover_url != ''
			 ORDER BY last_fetched DESC LIMIT 1`, mediaItemID).Scan(&coverURL)
		if err == nil && coverURL.Valid && coverURL.String != "" {
			return coverURL.String
		}
	}

	// 3. Fall back to type-specific placeholder
	return fmt.Sprintf("/api/v1/cover/placeholder/%s", mediaTypeName)
}

// GetCoverURLsBatch returns cover image URLs for multiple media items at once.
// This is more efficient than calling GetCoverURL in a loop because it batches
// the database queries.
func (s *CoverArtService) GetCoverURLsBatch(ctx context.Context, items []CoverURLRequest) map[int64]string {
	result := make(map[int64]string, len(items))
	if len(items) == 0 {
		return result
	}

	// Set placeholder defaults for all items first
	itemTypeMap := make(map[int64]string, len(items))
	for _, item := range items {
		itemTypeMap[item.ID] = item.MediaTypeName
		result[item.ID] = fmt.Sprintf("/api/v1/cover/placeholder/%s", item.MediaTypeName)
	}

	if s.db == nil {
		return result
	}

	// Build IN clause for IDs
	ids := make([]interface{}, len(items))
	placeholders := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	// Batch query external_metadata cover URLs
	query := fmt.Sprintf(
		`SELECT media_item_id, cover_url FROM external_metadata
		 WHERE media_item_id IN (%s) AND cover_url IS NOT NULL AND cover_url != ''
		 ORDER BY last_fetched DESC`, inClause)

	rows, err := s.db.QueryContext(ctx, query, ids...)
	if err == nil {
		defer rows.Close()
		seen := make(map[int64]bool)
		for rows.Next() {
			var itemID int64
			var coverURL string
			if err := rows.Scan(&itemID, &coverURL); err == nil && !seen[itemID] {
				result[itemID] = coverURL
				seen[itemID] = true
			}
		}
	}

	// Batch query cover_art table (overrides external_metadata if present)
	query = fmt.Sprintf(
		`SELECT media_item_id, local_path, url FROM cover_art
		 WHERE media_item_id IN (%s) AND is_default = 1
		 ORDER BY created_at DESC`, inClause)

	rows2, err := s.db.QueryContext(ctx, query, ids...)
	if err == nil {
		defer rows2.Close()
		seen := make(map[int64]bool)
		for rows2.Next() {
			var itemID int64
			var localPath, urlVal sql.NullString
			if err := rows2.Scan(&itemID, &localPath, &urlVal); err == nil && !seen[itemID] {
				if urlVal.Valid && urlVal.String != "" {
					result[itemID] = urlVal.String
					seen[itemID] = true
				} else if localPath.Valid && localPath.String != "" {
					result[itemID] = fmt.Sprintf("/api/v1/cover/%d", itemID)
					seen[itemID] = true
				}
			}
		}
	}

	return result
}

// CoverURLRequest is used by GetCoverURLsBatch to identify items and their types.
type CoverURLRequest struct {
	ID            int64
	MediaTypeName string
}

// GeneratePlaceholderSVG generates a placeholder SVG image for a given media type.
// Returns SVG bytes with a colored background and an icon representing the type.
func GeneratePlaceholderSVG(mediaType string) []byte {
	colors := map[string]string{
		"movie":        "#E53E3E",
		"tv_show":      "#3182CE",
		"tv_season":    "#2B6CB0",
		"tv_episode":   "#4299E1",
		"music_artist": "#38A169",
		"music_album":  "#2F855A",
		"song":         "#48BB78",
		"game":         "#805AD5",
		"software":     "#DD6B20",
		"book":         "#2B6CB0",
		"comic":        "#ED64A6",
	}

	icons := map[string]string{
		"movie":        "&#127916;", // film clapper
		"tv_show":      "&#128250;", // TV
		"tv_season":    "&#128250;", // TV
		"tv_episode":   "&#128250;", // TV
		"music_artist": "&#127908;", // microphone
		"music_album":  "&#128191;", // CD
		"song":         "&#127925;", // music notes
		"game":         "&#127918;", // game controller
		"software":     "&#128187;", // computer
		"book":         "&#128214;", // book
		"comic":        "&#128214;", // book
	}

	labels := map[string]string{
		"movie":        "Movie",
		"tv_show":      "TV Show",
		"tv_season":    "Season",
		"tv_episode":   "Episode",
		"music_artist": "Artist",
		"music_album":  "Album",
		"song":         "Song",
		"game":         "Game",
		"software":     "Software",
		"book":         "Book",
		"comic":        "Comic",
	}

	color, ok := colors[mediaType]
	if !ok {
		color = "#718096" // gray fallback
	}
	icon, ok := icons[mediaType]
	if !ok {
		icon = "&#128196;" // generic document
	}
	label, ok := labels[mediaType]
	if !ok {
		label = strings.ReplaceAll(mediaType, "_", " ")
	}

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="300" height="300" viewBox="0 0 300 300">
  <defs>
    <linearGradient id="bg" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" style="stop-color:%s;stop-opacity:1"/>
      <stop offset="100%%" style="stop-color:%s;stop-opacity:0.8"/>
    </linearGradient>
  </defs>
  <rect width="300" height="300" rx="12" fill="url(#bg)"/>
  <text x="150" y="130" font-size="72" text-anchor="middle" dominant-baseline="middle" fill="white" opacity="0.9">%s</text>
  <text x="150" y="200" font-size="20" font-family="Arial, Helvetica, sans-serif" font-weight="bold" text-anchor="middle" dominant-baseline="middle" fill="white" opacity="0.85">%s</text>
</svg>`, color, color, icon, label)

	return []byte(svg)
}
