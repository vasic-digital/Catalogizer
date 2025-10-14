package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"catalogizer/internal/models"
	"go.uber.org/zap"
)

type MediaRecognitionService struct {
	db                    *sql.DB
	logger                *zap.Logger
	cacheService          *CacheService
	translationService    *TranslationService
	movieAPIBaseURL       string
	musicAPIBaseURL       string
	bookAPIBaseURL        string
	gameAPIBaseURL        string
	ocrAPIBaseURL         string
	fingerprintAPIBaseURL string
}

// Recognition request structure
type MediaRecognitionRequest struct {
	FilePath    string            `json:"file_path"`
	FileName    string            `json:"file_name"`
	FileSize    int64             `json:"file_size"`
	FileHash    string            `json:"file_hash"`
	MimeType    string            `json:"mime_type"`
	MediaType   MediaType         `json:"media_type,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	AudioSample []byte            `json:"audio_sample,omitempty"`
	VideoSample []byte            `json:"video_sample,omitempty"`
	ImageSample []byte            `json:"image_sample,omitempty"`
	TextSample  string            `json:"text_sample,omitempty"`
	UserHints   map[string]string `json:"user_hints,omitempty"`
	Languages   []string          `json:"languages,omitempty"`
}

// Recognition result structure
type MediaRecognitionResult struct {
	MediaID           string     `json:"media_id"`
	MediaType         MediaType  `json:"media_type"`
	Title             string     `json:"title"`
	OriginalTitle     string     `json:"original_title,omitempty"`
	AlternativeTitles []string   `json:"alternative_titles,omitempty"`
	Description       string     `json:"description"`
	Year              int        `json:"year,omitempty"`
	ReleaseDate       *time.Time `json:"release_date,omitempty"`
	Duration          int64      `json:"duration,omitempty"`
	Genres            []string   `json:"genres,omitempty"`
	Tags              []string   `json:"tags,omitempty"`

	// Movie/TV specific
	Director    string   `json:"director,omitempty"`
	Cast        []Person `json:"cast,omitempty"`
	IMDbID      string   `json:"imdb_id,omitempty"`
	TMDbID      string   `json:"tmdb_id,omitempty"`
	TVDBId      string   `json:"tvdb_id,omitempty"`
	Season      int      `json:"season,omitempty"`
	Episode     int      `json:"episode,omitempty"`
	SeriesTitle string   `json:"series_title,omitempty"`
	Rating      float64  `json:"rating,omitempty"`

	// Music specific
	Artist        string `json:"artist,omitempty"`
	AlbumArtist   string `json:"album_artist,omitempty"`
	Album         string `json:"album,omitempty"`
	TrackNumber   int    `json:"track_number,omitempty"`
	DiscNumber    int    `json:"disc_number,omitempty"`
	MusicBrainzID string `json:"musicbrainz_id,omitempty"`
	SpotifyID     string `json:"spotify_id,omitempty"`
	LastFMID      string `json:"lastfm_id,omitempty"`
	BPM           int    `json:"bpm,omitempty"`
	Key           string `json:"key,omitempty"`

	// Book/Publication specific
	Author    string   `json:"author,omitempty"`
	Authors   []Person `json:"authors,omitempty"`
	Publisher string   `json:"publisher,omitempty"`
	ISBN      string   `json:"isbn,omitempty"`
	ISBN10    string   `json:"isbn10,omitempty"`
	ISBN13    string   `json:"isbn13,omitempty"`
	ISSN      string   `json:"issn,omitempty"`
	DOI       string   `json:"doi,omitempty"`
	Language  string   `json:"language,omitempty"`
	PageCount int      `json:"page_count,omitempty"`
	WordCount int      `json:"word_count,omitempty"`
	Edition   string   `json:"edition,omitempty"`
	Series    string   `json:"series,omitempty"`
	Volume    int      `json:"volume,omitempty"`
	Issue     int      `json:"issue,omitempty"`

	// Game/Software specific
	Developer          string            `json:"developer,omitempty"`
	Publisher_Game     string            `json:"publisher_game,omitempty"`
	Platform           string            `json:"platform,omitempty"`
	Platforms          []string          `json:"platforms,omitempty"`
	Version            string            `json:"version,omitempty"`
	BuildNumber        string            `json:"build_number,omitempty"`
	License            string            `json:"license,omitempty"`
	SystemRequirements map[string]string `json:"system_requirements,omitempty"`
	IGDBId             string            `json:"igdb_id,omitempty"`
	SteamID            string            `json:"steam_id,omitempty"`

	// Cover art and media
	CoverArt    []models.CoverArtResult `json:"cover_art,omitempty"`
	Screenshots []string                `json:"screenshots,omitempty"`
	Trailer     string                  `json:"trailer,omitempty"`
	PreviewURL  string                  `json:"preview_url,omitempty"`

	// Recognition metadata
	Confidence        float64           `json:"confidence"`
	RecognitionMethod string            `json:"recognition_method"`
	APIProvider       string            `json:"api_provider"`
	RecognizedAt      time.Time         `json:"recognized_at"`
	ProcessingTime    int64             `json:"processing_time_ms"`
	Fingerprints      map[string]string `json:"fingerprints,omitempty"`

	// Additional metadata
	ExternalIDs  map[string]string      `json:"external_ids,omitempty"`
	Translations map[string]Translation `json:"translations,omitempty"`
	RelatedMedia []string               `json:"related_media,omitempty"`
	Duplicates   []DuplicateMatch       `json:"duplicates,omitempty"`
}

type Person struct {
	Name        string            `json:"name"`
	Role        string            `json:"role,omitempty"`
	Character   string            `json:"character,omitempty"`
	Biography   string            `json:"biography,omitempty"`
	BirthDate   *time.Time        `json:"birth_date,omitempty"`
	PhotoURL    string            `json:"photo_url,omitempty"`
	ExternalIDs map[string]string `json:"external_ids,omitempty"`
}

type Translation struct {
	Language    string   `json:"language"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Genres      []string `json:"genres,omitempty"`
}

type DuplicateMatch struct {
	MediaID    string  `json:"media_id"`
	FilePath   string  `json:"file_path"`
	Similarity float64 `json:"similarity"`
	MatchType  string  `json:"match_type"`
}

// Recognition providers interface
type RecognitionProvider interface {
	RecognizeMedia(ctx context.Context, req *MediaRecognitionRequest) (*MediaRecognitionResult, error)
	GetProviderName() string
	SupportsMediaType(mediaType MediaType) bool
	GetConfidenceThreshold() float64
}

// Audio fingerprinting structure
type AudioFingerprint struct {
	Algorithm  string               `json:"algorithm"`
	Hash       string               `json:"hash"`
	Duration   float64              `json:"duration"`
	SampleRate int                  `json:"sample_rate"`
	Channels   int                  `json:"channels"`
	Features   map[string]float64   `json:"features"`
	Segments   []FingerprintSegment `json:"segments"`
}

type FingerprintSegment struct {
	StartTime float64            `json:"start_time"`
	EndTime   float64            `json:"end_time"`
	Hash      string             `json:"hash"`
	Features  map[string]float64 `json:"features"`
}

// OCR result structure for text recognition
type OCRResult struct {
	Text       string            `json:"text"`
	Confidence float64           `json:"confidence"`
	Language   string            `json:"language"`
	Blocks     []TextBlock       `json:"blocks"`
	Layout     LayoutInfo        `json:"layout"`
	Metadata   map[string]string `json:"metadata"`
}

type TextBlock struct {
	Text        string    `json:"text"`
	Confidence  float64   `json:"confidence"`
	BoundingBox Rectangle `json:"bounding_box"`
	WordCount   int       `json:"word_count"`
	FontInfo    FontInfo  `json:"font_info"`
}

type Rectangle struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type FontInfo struct {
	Family string  `json:"family"`
	Size   float64 `json:"size"`
	Bold   bool    `json:"bold"`
	Italic bool    `json:"italic"`
	Color  string  `json:"color"`
}

type LayoutInfo struct {
	PageCount   int           `json:"page_count"`
	Orientation string        `json:"orientation"`
	TextColumns int           `json:"text_columns"`
	Images      []ImageRegion `json:"images"`
	Tables      []TableRegion `json:"tables"`
}

type ImageRegion struct {
	BoundingBox Rectangle `json:"bounding_box"`
	Caption     string    `json:"caption"`
	Type        string    `json:"type"`
}

type TableRegion struct {
	BoundingBox Rectangle `json:"bounding_box"`
	Rows        int       `json:"rows"`
	Columns     int       `json:"columns"`
	Headers     []string  `json:"headers"`
}

func NewMediaRecognitionService(
	db *sql.DB,
	logger *zap.Logger,
	cacheService *CacheService,
	translationService *TranslationService,
	movieAPIBaseURL string,
	musicAPIBaseURL string,
	bookAPIBaseURL string,
	gameAPIBaseURL string,
	ocrAPIBaseURL string,
	fingerprintAPIBaseURL string,
) *MediaRecognitionService {
	return &MediaRecognitionService{
		db:                    db,
		logger:                logger,
		cacheService:          cacheService,
		translationService:    translationService,
		movieAPIBaseURL:       movieAPIBaseURL,
		musicAPIBaseURL:       musicAPIBaseURL,
		bookAPIBaseURL:        bookAPIBaseURL,
		gameAPIBaseURL:        gameAPIBaseURL,
		ocrAPIBaseURL:         ocrAPIBaseURL,
		fingerprintAPIBaseURL: fingerprintAPIBaseURL,
	}
}

// Main recognition method that orchestrates all providers
func (s *MediaRecognitionService) RecognizeMedia(ctx context.Context, req *MediaRecognitionRequest) (*MediaRecognitionResult, error) {
	startTime := time.Now()
	s.logger.Info("Starting media recognition",
		zap.String("file_path", req.FilePath),
		zap.String("mime_type", req.MimeType),
		zap.String("media_type", string(req.MediaType)))

	// Check cache first
	cacheKey := fmt.Sprintf("media_recognition:%s", req.FileHash)
	var result MediaRecognitionResult
	if found, err := s.cacheService.Get(ctx, cacheKey, &result); err == nil && found {
		s.logger.Debug("Found cached recognition result", zap.String("media_id", result.MediaID))
		return &result, nil
	}

	// Determine media type if not provided
	if req.MediaType == "" {
		detectedType, confidence := s.detectMediaType(req)
		req.MediaType = detectedType
		s.logger.Debug("Detected media type",
			zap.String("type", string(detectedType)),
			zap.Float64("confidence", confidence))
	}

	// Get appropriate recognition providers
	providers := s.getProvidersForMediaType(req.MediaType)
	if len(providers) == 0 {
		return nil, fmt.Errorf("no recognition providers available for media type: %s", req.MediaType)
	}

	// Try recognition with multiple providers
	var bestResult *MediaRecognitionResult
	var bestConfidence float64

	for _, provider := range providers {
		result, err := provider.RecognizeMedia(ctx, req)
		if err != nil {
			s.logger.Warn("Recognition provider failed",
				zap.String("provider", provider.GetProviderName()),
				zap.Error(err))
			continue
		}

		if result.Confidence > bestConfidence && result.Confidence >= provider.GetConfidenceThreshold() {
			bestResult = result
			bestConfidence = result.Confidence
		}
	}

	if bestResult == nil {
		return nil, fmt.Errorf("no recognition provider returned confident results")
	}

	// Enhance with additional metadata
	s.enhanceRecognitionResult(ctx, bestResult, req)

	// Check for duplicates
	duplicates, err := s.findDuplicates(ctx, bestResult)
	if err != nil {
		s.logger.Warn("Failed to find duplicates", zap.Error(err))
	} else {
		bestResult.Duplicates = duplicates
	}

	// Translate metadata if requested
	if len(req.Languages) > 0 {
		translations, err := s.translateMetadata(ctx, bestResult, req.Languages)
		if err != nil {
			s.logger.Warn("Failed to translate metadata", zap.Error(err))
		} else {
			bestResult.Translations = translations
		}
	}

	// Set processing metadata
	bestResult.RecognizedAt = time.Now()
	bestResult.ProcessingTime = time.Since(startTime).Milliseconds()

	// Cache the result
	resultJSON, _ := json.Marshal(bestResult)
	s.cacheService.Set(ctx, cacheKey, string(resultJSON), 24*time.Hour)

	// Store in database
	if err := s.storeRecognitionResult(ctx, bestResult, req); err != nil {
		s.logger.Error("Failed to store recognition result", zap.Error(err))
	}

	s.logger.Info("Media recognition completed",
		zap.String("media_id", bestResult.MediaID),
		zap.String("title", bestResult.Title),
		zap.Float64("confidence", bestResult.Confidence),
		zap.Int64("processing_time_ms", bestResult.ProcessingTime))

	return bestResult, nil
}

// Detect media type from file characteristics
func (s *MediaRecognitionService) detectMediaType(req *MediaRecognitionRequest) (MediaType, float64) {
	// Video file detection
	videoMimes := []string{"video/mp4", "video/avi", "video/mkv", "video/mov", "video/wmv", "video/flv", "video/webm"}
	for _, mime := range videoMimes {
		if req.MimeType == mime {
			// Further distinguish between movie types
			if s.looksLikeTVEpisode(req.FileName) {
				return MediaTypeTVEpisode, 0.9
			}
			if s.looksLikeConcert(req.FileName) {
				return MediaTypeConcert, 0.8
			}
			if s.looksLikeDocumentary(req.FileName) {
				return MediaTypeDocumentary, 0.8
			}
			if s.looksLikeCourse(req.FileName) {
				return MediaTypeCourse, 0.8
			}
			return MediaTypeMovie, 0.7
		}
	}

	// Audio file detection
	audioMimes := []string{"audio/mp3", "audio/wav", "audio/flac", "audio/ogg", "audio/aac", "audio/m4a"}
	for _, mime := range audioMimes {
		if req.MimeType == mime {
			if s.looksLikeAudiobook(req.FileName) {
				return MediaTypeAudiobook, 0.9
			}
			if s.looksLikePodcast(req.FileName) {
				return MediaTypePodcast, 0.8
			}
			return MediaTypeMusic, 0.8
		}
	}

	// Text/Document file detection
	textMimes := []string{"application/pdf", "text/plain", "application/epub+zip", "application/x-mobipocket-ebook"}
	for _, mime := range textMimes {
		if req.MimeType == mime {
			if s.looksLikeComicBook(req.FileName) {
				return MediaTypeComicBook, 0.9
			}
			if s.looksLikeMagazine(req.FileName) {
				return MediaTypeMagazine, 0.8
			}
			if s.looksLikeManual(req.FileName) {
				return MediaTypeManual, 0.8
			}
			return MediaTypeBook, 0.7
		}
	}

	// Executable/Software detection
	execMimes := []string{"application/x-executable", "application/x-msdos-program", "application/x-msdownload"}
	for _, mime := range execMimes {
		if req.MimeType == mime {
			if s.looksLikeGame(req.FileName) {
				return MediaTypeGame, 0.8
			}
			return MediaTypeSoftware, 0.7
		}
	}

	// Default fallback based on file extension
	return s.detectFromFileName(req.FileName), 0.5
}

// Helper methods for media type detection
func (s *MediaRecognitionService) looksLikeTVEpisode(fileName string) bool {
	// Pattern matching for TV episodes (S01E01, 1x01, etc.)
	// Implementation would use regex patterns
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikeConcert(fileName string) bool {
	// Look for concert-related keywords
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikeDocumentary(fileName string) bool {
	// Look for documentary keywords
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikeCourse(fileName string) bool {
	// Look for course/training keywords
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikeAudiobook(fileName string) bool {
	// Look for audiobook patterns
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikePodcast(fileName string) bool {
	// Look for podcast patterns
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikeComicBook(fileName string) bool {
	// Look for comic book patterns
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikeMagazine(fileName string) bool {
	// Look for magazine patterns
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikeManual(fileName string) bool {
	// Look for manual/documentation patterns
	return false // Placeholder
}

func (s *MediaRecognitionService) looksLikeGame(fileName string) bool {
	// Look for game-related patterns
	return false // Placeholder
}

func (s *MediaRecognitionService) detectFromFileName(fileName string) MediaType {
	// Fallback detection based on file extension
	return MediaTypeMovie // Placeholder
}

// Get recognition providers for specific media type
func (s *MediaRecognitionService) getProvidersForMediaType(mediaType MediaType) []RecognitionProvider {
	var providers []RecognitionProvider

	switch mediaType {
	case MediaTypeMovie, MediaTypeTVSeries, MediaTypeTVEpisode, MediaTypeConcert, MediaTypeDocumentary:
		providers = append(providers, s.getMovieProviders()...)
	case MediaTypeMusic, MediaTypeAlbum, MediaTypeAudiobook, MediaTypePodcast:
		providers = append(providers, s.getMusicProviders()...)
	case MediaTypeBook, MediaTypeComicBook, MediaTypeMagazine, MediaTypeEbook:
		providers = append(providers, s.getBookProviders()...)
	case MediaTypeGame, MediaTypeSoftware:
		providers = append(providers, s.getGameProviders()...)
	}

	return providers
}

// Provider getter methods (to be implemented)
func (s *MediaRecognitionService) getMovieProviders() []RecognitionProvider {
	return []RecognitionProvider{} // Placeholder
}

func (s *MediaRecognitionService) getMusicProviders() []RecognitionProvider {
	return []RecognitionProvider{} // Placeholder
}

func (s *MediaRecognitionService) getBookProviders() []RecognitionProvider {
	return []RecognitionProvider{} // Placeholder
}

func (s *MediaRecognitionService) getGameProviders() []RecognitionProvider {
	return []RecognitionProvider{} // Placeholder
}

// Enhance recognition result with additional metadata
func (s *MediaRecognitionService) enhanceRecognitionResult(ctx context.Context, result *MediaRecognitionResult, req *MediaRecognitionRequest) {
	// Get additional cover art
	if coverArt, err := s.getAdditionalCoverArt(ctx, result); err == nil {
		result.CoverArt = append(result.CoverArt, coverArt...)
	}

	// Get additional metadata from alternative sources
	if metadata, err := s.getEnhancedMetadata(ctx, result); err == nil {
		// Merge additional metadata
		if result.ExternalIDs == nil {
			result.ExternalIDs = make(map[string]string)
		}
		for key, value := range metadata {
			result.ExternalIDs[key] = value
		}
	}
}

// Find duplicate content
func (s *MediaRecognitionService) findDuplicates(ctx context.Context, result *MediaRecognitionResult) ([]DuplicateMatch, error) {
	var duplicates []DuplicateMatch

	// Query database for potential duplicates based on:
	// 1. Exact title match
	// 2. External IDs (IMDb, ISBN, etc.)
	// 3. Fingerprint similarity
	// 4. File hash similarity

	query := `
		SELECT media_id, file_path, title, external_ids, fingerprints
		FROM media_recognition_results
		WHERE (title = ? OR ? IN (SELECT value FROM json_each(external_ids)))
		AND media_id != ?
	`

	rows, err := s.db.QueryContext(ctx, query, result.Title, result.IMDbID, result.MediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mediaID, filePath, title, externalIDsJSON, fingerprintsJSON string
		if err := rows.Scan(&mediaID, &filePath, &title, &externalIDsJSON, &fingerprintsJSON); err != nil {
			continue
		}

		similarity := s.calculateSimilarity(result, title, externalIDsJSON, fingerprintsJSON)
		if similarity > 0.8 { // High similarity threshold
			duplicates = append(duplicates, DuplicateMatch{
				MediaID:    mediaID,
				FilePath:   filePath,
				Similarity: similarity,
				MatchType:  s.determineMatchType(similarity),
			})
		}
	}

	return duplicates, nil
}

// Calculate similarity between media items
func (s *MediaRecognitionService) calculateSimilarity(result *MediaRecognitionResult, title, externalIDsJSON, fingerprintsJSON string) float64 {
	// Implement similarity calculation logic
	// Consider title similarity, external ID matches, fingerprint similarity
	return 0.0 // Placeholder
}

// Determine match type based on similarity score
func (s *MediaRecognitionService) determineMatchType(similarity float64) string {
	if similarity >= 0.95 {
		return "exact"
	} else if similarity >= 0.85 {
		return "high"
	} else if similarity >= 0.8 {
		return "medium"
	}
	return "low"
}

// Translate metadata to multiple languages
func (s *MediaRecognitionService) translateMetadata(ctx context.Context, result *MediaRecognitionResult, languages []string) (map[string]Translation, error) {
	translations := make(map[string]Translation)

	for _, lang := range languages {
		if lang == result.Language {
			continue // Skip if same as source language
		}

		translation := Translation{Language: lang}

		// Translate title
		if translatedTitle, err := s.translationService.TranslateText(ctx, TranslationRequest{
			Text:           result.Title,
			SourceLanguage: result.Language,
			TargetLanguage: lang,
		}); err == nil {
			translation.Title = translatedTitle.TranslatedText
		}

		// Translate description
		if result.Description != "" {
			if translatedDesc, err := s.translationService.TranslateText(ctx, TranslationRequest{
				Text:           result.Description,
				SourceLanguage: result.Language,
				TargetLanguage: lang,
			}); err == nil {
				translation.Description = translatedDesc.TranslatedText
			}
		}

		// Translate genres
		if len(result.Genres) > 0 {
			var translatedGenres []string
			for _, genre := range result.Genres {
				if translatedGenre, err := s.translationService.TranslateText(ctx, TranslationRequest{
					Text:           genre,
					SourceLanguage: result.Language,
					TargetLanguage: lang,
				}); err == nil {
					translatedGenres = append(translatedGenres, translatedGenre.TranslatedText)
				}
			}
			translation.Genres = translatedGenres
		}

		translations[lang] = translation
	}

	return translations, nil
}

// Store recognition result in database
func (s *MediaRecognitionService) storeRecognitionResult(ctx context.Context, result *MediaRecognitionResult, req *MediaRecognitionRequest) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	fingerprintsJSON, _ := json.Marshal(result.Fingerprints)
	externalIDsJSON, _ := json.Marshal(result.ExternalIDs)

	query := `
		INSERT OR REPLACE INTO media_recognition_results (
			media_id, file_path, file_hash, media_type, title, original_title,
			description, year, release_date, duration, genres, tags,
			recognition_data, fingerprints, external_ids, confidence,
			recognition_method, api_provider, recognized_at, processing_time_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	genresJSON, _ := json.Marshal(result.Genres)
	tagsJSON, _ := json.Marshal(result.Tags)

	_, err = s.db.ExecContext(ctx, query,
		result.MediaID, req.FilePath, req.FileHash, string(result.MediaType),
		result.Title, result.OriginalTitle, result.Description, result.Year,
		result.ReleaseDate, result.Duration, string(genresJSON), string(tagsJSON),
		string(resultJSON), string(fingerprintsJSON), string(externalIDsJSON),
		result.Confidence, result.RecognitionMethod, result.APIProvider,
		result.RecognizedAt, result.ProcessingTime,
	)

	return err
}

// Additional helper methods for enhancement
func (s *MediaRecognitionService) getAdditionalCoverArt(ctx context.Context, result *MediaRecognitionResult) ([]models.CoverArtResult, error) {
	// Implement additional cover art retrieval
	return []models.CoverArtResult{}, nil
}

func (s *MediaRecognitionService) getEnhancedMetadata(ctx context.Context, result *MediaRecognitionResult) (map[string]string, error) {
	// Implement enhanced metadata retrieval
	return make(map[string]string), nil
}

// Batch recognition for multiple files
func (s *MediaRecognitionService) RecognizeMediaBatch(ctx context.Context, requests []*MediaRecognitionRequest) ([]*MediaRecognitionResult, error) {
	results := make([]*MediaRecognitionResult, len(requests))

	for i, req := range requests {
		result, err := s.RecognizeMedia(ctx, req)
		if err != nil {
			s.logger.Error("Failed to recognize media in batch",
				zap.String("file_path", req.FilePath),
				zap.Error(err))
			continue
		}
		results[i] = result
	}

	return results, nil
}

// Get recognition statistics
func (s *MediaRecognitionService) GetRecognitionStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count by media type
	query := `
		SELECT media_type, COUNT(*) as count, AVG(confidence) as avg_confidence
		FROM media_recognition_results
		GROUP BY media_type
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	typeStats := make(map[string]map[string]interface{})
	for rows.Next() {
		var mediaType string
		var count int
		var avgConfidence float64
		if err := rows.Scan(&mediaType, &count, &avgConfidence); err != nil {
			continue
		}
		typeStats[mediaType] = map[string]interface{}{
			"count":          count,
			"avg_confidence": avgConfidence,
		}
	}

	stats["by_type"] = typeStats

	// Overall statistics
	totalQuery := `SELECT COUNT(*) FROM media_recognition_results`
	var totalCount int
	s.db.QueryRowContext(ctx, totalQuery).Scan(&totalCount)
	stats["total_recognized"] = totalCount

	return stats, nil
}
