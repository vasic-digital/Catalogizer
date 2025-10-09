package services

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/image/draw"
)

// CoverArtService handles cover art retrieval, processing, and caching
type CoverArtService struct {
	db         *sql.DB
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
	Title       string             `json:"title"`
	Artist      string             `json:"artist"`
	Album       *string            `json:"album,omitempty"`
	Year        *int               `json:"year,omitempty"`
	MusicBrainzID *string          `json:"musicbrainz_id,omitempty"`
	SpotifyID   *string            `json:"spotify_id,omitempty"`
	Quality     CoverArtQuality    `json:"quality"`
	Providers   []CoverArtProvider `json:"providers,omitempty"`
	UseCache    bool               `json:"use_cache"`
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
	Quality         int     `json:"quality"`         // JPEG quality 1-100
	Format          string  `json:"format"`          // "jpeg", "png", "webp"
	Crop            bool    `json:"crop"`            // Crop to exact dimensions
	PreserveAspect  bool    `json:"preserve_aspect"` // Preserve aspect ratio
	BackgroundColor *string `json:"background_color,omitempty"` // Hex color for padding
}

// NewCoverArtService creates a new cover art service
func NewCoverArtService(db *sql.DB, logger *zap.Logger) *CoverArtService {
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
	hash := md5.Sum([]byte(data))
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

// Additional methods would be implemented for:
// - getCachedCoverArt
// - getCoverArtDownloadInfo
// - processAndSaveCoverArt
// - generateAdditionalSizes
// - setDefaultCoverArt
// - getVideoDuration
// - generateVideoThumbnail
// - processLocalCoverArt
// etc.