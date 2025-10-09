package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"catalog-api/internal/services"
)

func TestSubtitleServiceIntegration(t *testing.T) {
	// Setup with mock server
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	translationService := services.NewTranslationService(db, logger, mockServer.URL())
	cacheService := services.NewCacheService(db, logger)
	subtitleService := services.NewSubtitleService(db, logger, mockServer.URL(), cacheService, translationService)

	t.Run("SearchSubtitles", func(t *testing.T) {
		req := &services.SubtitleSearchRequest{
			IMDbID:    "tt1234567",
			Languages: []string{"en", "es"},
			Year:      2023,
		}

		results, err := subtitleService.SearchSubtitles(context.Background(), req)
		require.NoError(t, err)
		require.NotEmpty(t, results)

		// Verify mock server was called
		requests := mockServer.GetRequestLog()
		assert.True(t, len(requests) > 0)

		// Verify results structure
		result := results[0]
		assert.NotEmpty(t, result.ID)
		assert.NotEmpty(t, result.Language)
		assert.NotEmpty(t, result.DownloadURL)
		assert.True(t, result.Rating > 0)
	})

	t.Run("DownloadSubtitle", func(t *testing.T) {
		// First search for subtitles
		searchReq := &services.SubtitleSearchRequest{
			IMDbID:    "tt1234567",
			Languages: []string{"en"},
		}

		results, err := subtitleService.SearchSubtitles(context.Background(), searchReq)
		require.NoError(t, err)
		require.NotEmpty(t, results)

		// Download the first result
		downloadReq := &services.SubtitleDownloadRequest{
			SubtitleID: results[0].ID,
			VideoID:    123,
			Language:   "en",
		}

		subtitle, err := subtitleService.DownloadSubtitle(context.Background(), downloadReq)
		require.NoError(t, err)
		require.NotNil(t, subtitle)

		assert.Equal(t, downloadReq.VideoID, subtitle.MediaItemID)
		assert.Equal(t, downloadReq.Language, subtitle.Language)
		assert.NotEmpty(t, subtitle.SubtitleData)
		assert.Equal(t, "srt", subtitle.Format)
	})

	t.Run("TranslateSubtitle", func(t *testing.T) {
		// Create test subtitle
		originalSubtitle := &services.SubtitleTrack{
			MediaItemID:  123,
			Language:     "en",
			SubtitleData: "1\n00:00:01,000 --> 00:00:03,000\nHello world\n\n2\n00:00:04,000 --> 00:00:06,000\nThis is a test",
			Format:       "srt",
		}

		req := &services.TranslationRequest{
			SubtitleTrack: originalSubtitle,
			TargetLanguage: "es",
		}

		translatedSubtitle, err := subtitleService.TranslateSubtitle(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, translatedSubtitle)

		assert.Equal(t, "es", translatedSubtitle.Language)
		assert.NotEmpty(t, translatedSubtitle.SubtitleData)
		assert.Contains(t, translatedSubtitle.SubtitleData, "Translated:")
		assert.Equal(t, originalSubtitle.Format, translatedSubtitle.Format)
	})

	t.Run("CachingBehavior", func(t *testing.T) {
		// Clear request log
		mockServer.ClearRequestLog()

		req := &services.SubtitleSearchRequest{
			IMDbID:    "tt7654321",
			Languages: []string{"en"},
		}

		// First request
		results1, err := subtitleService.SearchSubtitles(context.Background(), req)
		require.NoError(t, err)

		requestCount1 := len(mockServer.GetRequestLog())

		// Second identical request (should use cache)
		results2, err := subtitleService.SearchSubtitles(context.Background(), req)
		require.NoError(t, err)

		requestCount2 := len(mockServer.GetRequestLog())

		// Should have same results
		assert.Equal(t, len(results1), len(results2))
		if len(results1) > 0 && len(results2) > 0 {
			assert.Equal(t, results1[0].ID, results2[0].ID)
		}

		// Should not have made additional requests due to caching
		assert.Equal(t, requestCount1, requestCount2)
	})
}

func TestLyricsServiceIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	translationService := services.NewTranslationService(db, logger, mockServer.URL())
	cacheService := services.NewCacheService(db, logger)
	lyricsService := services.NewLyricsService(db, logger, mockServer.URL(), cacheService, translationService)

	t.Run("SearchLyrics", func(t *testing.T) {
		req := &services.LyricsSearchRequest{
			Artist: "Test Artist",
			Title:  "Test Song",
		}

		results, err := lyricsService.SearchLyrics(context.Background(), req)
		require.NoError(t, err)
		require.NotEmpty(t, results)

		// Verify results structure
		result := results[0]
		assert.NotEmpty(t, result.ID)
		assert.Contains(t, result.Artist, "Test Artist")
		assert.Contains(t, result.Title, "Test")
		assert.NotEmpty(t, result.URL)
		assert.True(t, result.Confidence > 0)
	})

	t.Run("GetConcertLyrics", func(t *testing.T) {
		req := &services.ConcertLyricsRequest{
			Artist:      "Test Artist",
			VenueCity:   "Test City",
			ConcertDate: time.Now().Format("2006-01-02"),
		}

		lyrics, err := lyricsService.GetConcertLyrics(context.Background(), req)
		require.NoError(t, err)
		require.NotEmpty(t, lyrics)

		// Verify setlist structure
		for _, lyric := range lyrics {
			assert.NotEmpty(t, lyric.Title)
			assert.NotEmpty(t, lyric.Artist)
			assert.NotEmpty(t, lyric.Content)
		}

		// Verify mock server was called
		requests := mockServer.GetRequestLog()
		assert.True(t, len(requests) > 0)

		// Check that setlist.fm API was called
		setlistFMCalled := false
		for _, req := range requests {
			if req.URL != "" && req.URL != "/" {
				setlistFMCalled = true
				break
			}
		}
		assert.True(t, setlistFMCalled)
	})

	t.Run("SynchronizeLyrics", func(t *testing.T) {
		// Test synchronized lyrics
		req := &services.LyricsSyncRequest{
			LyricsID:     "test-lyrics-123",
			AudioFile:    "/test/audio.mp3",
			TimingMethod: "auto",
		}

		lyrics, err := lyricsService.SynchronizeLyrics(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, lyrics)

		assert.NotEmpty(t, lyrics.Content)
		assert.True(t, lyrics.IsSynchronized)
		assert.NotEmpty(t, lyrics.SyncData)

		// Verify sync data contains timestamps
		assert.Contains(t, lyrics.SyncData, "timestamp")
	})
}

func TestCoverArtServiceIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	cacheService := services.NewCacheService(db, logger)
	coverArtService := services.NewCoverArtService(db, logger, mockServer.URL(), cacheService)

	t.Run("SearchCoverArt", func(t *testing.T) {
		req := &services.CoverArtSearchRequest{
			Artist: "Test Artist",
			Album:  "Test Album",
			Year:   2023,
		}

		results, err := coverArtService.SearchCoverArt(context.Background(), req)
		require.NoError(t, err)
		require.NotEmpty(t, results)

		// Verify results structure
		result := results[0]
		assert.NotEmpty(t, result.URL)
		assert.NotEmpty(t, result.Provider)
		assert.True(t, result.Width > 0)
		assert.True(t, result.Height > 0)
		assert.True(t, result.Quality > 0)
	})

	t.Run("MultipleProviders", func(t *testing.T) {
		mockServer.ClearRequestLog()

		req := &services.CoverArtSearchRequest{
			Artist: "Multi Provider Test",
			Album:  "Test Album",
		}

		results, err := coverArtService.SearchCoverArt(context.Background(), req)
		require.NoError(t, err)

		// Should get results from multiple providers
		providers := make(map[string]bool)
		for _, result := range results {
			providers[result.Provider] = true
		}

		// Verify multiple providers were called
		requests := mockServer.GetRequestLog()
		assert.True(t, len(requests) > 1, "Should have called multiple providers")

		// Verify results from different providers
		assert.True(t, len(providers) >= 1, "Should have results from at least one provider")
	})

	t.Run("GenerateVideoThumbnails", func(t *testing.T) {
		req := &services.VideoThumbnailRequest{
			FilePath: "/test/video.mp4",
			Position: 60000, // 1 minute
			Width:    320,
			Height:   180,
			Quality:  85,
		}

		thumbnails, err := coverArtService.GenerateVideoThumbnails(context.Background(), req)
		require.NoError(t, err)
		require.NotEmpty(t, thumbnails)

		thumbnail := thumbnails[0]
		assert.NotEmpty(t, thumbnail.URL)
		assert.Equal(t, req.Width, thumbnail.Width)
		assert.Equal(t, req.Height, thumbnail.Height)
		assert.Equal(t, "thumbnail", thumbnail.Provider)
	})
}

func TestTranslationServiceIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	translationService := services.NewTranslationService(db, logger, mockServer.URL())

	t.Run("TranslateText", func(t *testing.T) {
		req := services.TranslationRequest{
			Text:       "Hello world",
			SourceLang: "en",
			TargetLang: "es",
		}

		result, err := translationService.TranslateText(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, req.SourceLang, result.SourceLanguage)
		assert.Equal(t, req.TargetLang, result.TargetLanguage)
		assert.NotEmpty(t, result.TranslatedText)
		assert.Contains(t, result.TranslatedText, "Translated:")
		assert.True(t, result.Confidence > 0)
	})

	t.Run("DetectLanguage", func(t *testing.T) {
		req := &services.LanguageDetectionRequest{
			Text: "This is English text for language detection",
		}

		result, err := translationService.DetectLanguage(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "en", result.Language)
		assert.True(t, result.Confidence > 0.5)
	})

	t.Run("BatchTranslation", func(t *testing.T) {
		req := &services.BatchTranslationRequest{
			Texts: []string{
				"Hello",
				"Goodbye",
				"Thank you",
			},
			SourceLang: "en",
			TargetLang: "fr",
		}

		result, err := translationService.TranslateBatch(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Translations, 3)
		for i, translation := range result.Translations {
			assert.NotEmpty(t, translation.TranslatedText)
			assert.Equal(t, req.Texts[i], translation.OriginalText)
		}
	})

	t.Run("FallbackProviders", func(t *testing.T) {
		// Set a delay to test fallback behavior
		mockServer.SetResponseDelay(100 * time.Millisecond)

		req := services.TranslationRequest{
			Text:       "Fallback test",
			SourceLang: "en",
			TargetLang: "de",
		}

		result, err := translationService.TranslateText(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotEmpty(t, result.TranslatedText)
		assert.NotEmpty(t, result.Provider)

		// Reset delay
		mockServer.SetResponseDelay(0)
	})
}

func TestCacheServiceIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	cacheService := services.NewCacheService(db, logger)

	t.Run("BasicCaching", func(t *testing.T) {
		key := "test:basic:cache"
		value := map[string]interface{}{
			"test":      "value",
			"number":    42,
			"timestamp": time.Now().Unix(),
		}

		// Set cache entry
		err := cacheService.Set(context.Background(), key, value, 1*time.Hour)
		require.NoError(t, err)

		// Get cache entry
		var retrieved map[string]interface{}
		found, err := cacheService.Get(context.Background(), key, &retrieved)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, value["test"], retrieved["test"])
		assert.Equal(t, float64(42), retrieved["number"]) // JSON unmarshaling converts to float64
	})

	t.Run("CacheExpiration", func(t *testing.T) {
		key := "test:expiration:cache"
		value := "expires quickly"

		// Set with very short TTL
		err := cacheService.Set(context.Background(), key, value, 50*time.Millisecond)
		require.NoError(t, err)

		// Should be available immediately
		var retrieved string
		found, err := cacheService.Get(context.Background(), key, &retrieved)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, value, retrieved)

		// Wait for expiration
		time.Sleep(100 * time.Millisecond)

		// Should be expired
		found, err = cacheService.Get(context.Background(), key, &retrieved)
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("MediaMetadataCache", func(t *testing.T) {
		mediaID := int64(123)
		metadataType := "lyrics"
		provider := "genius"
		quality := 9.5

		metadata := map[string]interface{}{
			"title":  "Test Song",
			"artist": "Test Artist",
			"lyrics": "Test lyrics content...",
		}

		// Set metadata cache
		err := cacheService.SetMediaMetadata(context.Background(), mediaID, metadataType, provider, metadata, quality)
		require.NoError(t, err)

		// Get metadata cache
		var retrieved map[string]interface{}
		found, retrievedQuality, err := cacheService.GetMediaMetadata(context.Background(), mediaID, metadataType, provider, &retrieved)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, quality, retrievedQuality)
		assert.Equal(t, metadata["title"], retrieved["title"])
		assert.Equal(t, metadata["artist"], retrieved["artist"])
		assert.Equal(t, metadata["lyrics"], retrieved["lyrics"])
	})

	t.Run("APIResponseCache", func(t *testing.T) {
		provider := "test_api"
		endpoint := "/search"
		requestData := map[string]interface{}{
			"query": "test search",
			"limit": 10,
		}
		responseData := map[string]interface{}{
			"results": []string{"result1", "result2"},
			"total":   2,
		}
		statusCode := 200

		// Set API response cache
		err := cacheService.SetAPIResponse(context.Background(), provider, endpoint, requestData, responseData, statusCode, 1*time.Hour)
		require.NoError(t, err)

		// Get API response cache
		var retrieved map[string]interface{}
		found, retrievedStatus, err := cacheService.GetAPIResponse(context.Background(), provider, endpoint, requestData, &retrieved)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, statusCode, retrievedStatus)
		assert.Equal(t, responseData["total"], retrieved["total"])

		// Verify results array
		retrievedResults, ok := retrieved["results"].([]interface{})
		require.True(t, ok)
		assert.Len(t, retrievedResults, 2)
	})

	t.Run("CacheStats", func(t *testing.T) {
		// Add some test data
		cacheService.Set(context.Background(), "stats:test1", "value1", 1*time.Hour)
		cacheService.Set(context.Background(), "stats:test2", "value2", 1*time.Hour)
		cacheService.Set(context.Background(), "translation:test", "translated", 1*time.Hour)

		stats, err := cacheService.GetStats(context.Background())
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.True(t, stats.TotalEntries > 0)
		assert.True(t, stats.TotalSize > 0)
		assert.NotEmpty(t, stats.CachesByType)
	})

	t.Run("CacheCleanup", func(t *testing.T) {
		// Add expired entries
		cacheService.Set(context.Background(), "cleanup:test1", "value1", 1*time.Millisecond)
		cacheService.Set(context.Background(), "cleanup:test2", "value2", 1*time.Millisecond)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Run cleanup
		err := cacheService.CleanupExpired(context.Background())
		require.NoError(t, err)

		// Verify entries were cleaned up
		var value string
		found, err := cacheService.Get(context.Background(), "cleanup:test1", &value)
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func TestLocalizationServiceIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	translationService := services.NewTranslationService(db, logger, mockServer.URL())
	cacheService := services.NewCacheService(db, logger)
	localizationService := services.NewLocalizationService(db, logger, translationService, cacheService)

	t.Run("SetupUserLocalization", func(t *testing.T) {
		req := &services.WizardLocalizationStep{
			UserID:                1,
			PrimaryLanguage:       "en",
			SecondaryLanguages:    []string{"es", "fr"},
			SubtitleLanguages:     []string{"en", "es"},
			LyricsLanguages:       []string{"en"},
			MetadataLanguages:     []string{"en", "es", "fr"},
			AutoTranslate:         true,
			AutoDownloadSubtitles: true,
			AutoDownloadLyrics:    true,
			PreferredRegion:       "US",
			DateFormat:            "MM/DD/YYYY",
			TimeFormat:            "12h",
			NumberFormat:          "en-US",
			CurrencyCode:          "USD",
		}

		localization, err := localizationService.SetupUserLocalization(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, localization)

		assert.Equal(t, req.UserID, localization.UserID)
		assert.Equal(t, req.PrimaryLanguage, localization.PrimaryLanguage)
		assert.Equal(t, req.SecondaryLanguages, localization.SecondaryLanguages)
		assert.Equal(t, req.AutoTranslate, localization.AutoTranslate)
		assert.Equal(t, req.PreferredRegion, localization.PreferredRegion)
	})

	t.Run("GetPreferredLanguagesForContent", func(t *testing.T) {
		// Setup user localization first
		req := &services.WizardLocalizationStep{
			UserID:            2,
			PrimaryLanguage:   "es",
			SubtitleLanguages: []string{"es", "en"},
			LyricsLanguages:   []string{"es"},
		}

		_, err := localizationService.SetupUserLocalization(context.Background(), req)
		require.NoError(t, err)

		// Test subtitle preferences
		subtitleLangs, err := localizationService.GetPreferredLanguagesForContent(context.Background(), 2, "subtitles")
		require.NoError(t, err)
		assert.Equal(t, []string{"es", "en"}, subtitleLangs)

		// Test lyrics preferences
		lyricsLangs, err := localizationService.GetPreferredLanguagesForContent(context.Background(), 2, "lyrics")
		require.NoError(t, err)
		assert.Equal(t, []string{"es"}, lyricsLangs)
	})

	t.Run("SupportedLanguages", func(t *testing.T) {
		languages, err := localizationService.GetSupportedLanguages(context.Background())
		require.NoError(t, err)
		require.NotEmpty(t, languages)

		// Verify essential languages are supported
		assert.Contains(t, languages, "en")
		assert.Contains(t, languages, "es")
		assert.Contains(t, languages, "fr")

		// Verify language profile structure
		enProfile := languages["en"]
		assert.Equal(t, "English", enProfile.Name)
		assert.Equal(t, "English", enProfile.NativeName)
		assert.Equal(t, "ltr", enProfile.Direction)
		assert.Contains(t, enProfile.SupportedBy, "subtitles")
		assert.Contains(t, enProfile.SupportedBy, "lyrics")
	})

	t.Run("LanguageDetection", func(t *testing.T) {
		testCases := []struct {
			acceptLanguage string
			expectedLang   string
		}{
			{"en-US,en;q=0.9,es;q=0.8", "en"},
			{"es-ES,es;q=0.9", "es"},
			{"fr-FR,fr;q=0.9,en;q=0.8", "fr"},
			{"de-DE,de;q=0.9", "de"},
			{"invalid-lang", "en"}, // fallback
		}

		for _, tc := range testCases {
			detected := localizationService.DetectUserLanguage(context.Background(), "", tc.acceptLanguage)
			assert.Equal(t, tc.expectedLang, detected, "Failed for Accept-Language: %s", tc.acceptLanguage)
		}
	})

	t.Run("GetWizardDefaults", func(t *testing.T) {
		testCases := []struct {
			detectedLang   string
			expectedRegion string
			expectedCurrency string
		}{
			{"en", "US", "USD"},
			{"es", "ES", "EUR"},
			{"fr", "FR", "EUR"},
			{"de", "DE", "EUR"},
			{"ja", "JP", "JPY"},
		}

		for _, tc := range testCases {
			defaults := localizationService.GetWizardDefaults(context.Background(), tc.detectedLang)
			assert.Equal(t, tc.detectedLang, defaults.PrimaryLanguage)
			assert.Equal(t, tc.expectedRegion, defaults.PreferredRegion)
			assert.Equal(t, tc.expectedCurrency, defaults.CurrencyCode)
			assert.Contains(t, defaults.SubtitleLanguages, tc.detectedLang)
		}
	})

	t.Run("DateTimeFormatting", func(t *testing.T) {
		// Setup user with specific formatting preferences
		req := &services.WizardLocalizationStep{
			UserID:     3,
			PrimaryLanguage: "en",
			DateFormat: "DD/MM/YYYY",
			TimeFormat: "24h",
		}

		_, err := localizationService.SetupUserLocalization(context.Background(), req)
		require.NoError(t, err)

		// Test datetime formatting
		testTime := time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC)
		formatted, err := localizationService.FormatDateTimeForUser(context.Background(), 3, testTime)
		require.NoError(t, err)

		// Should use DD/MM/YYYY and 24h format
		assert.Contains(t, formatted, "25/12/2023")
		assert.Contains(t, formatted, "14:30")
	})
}

func TestJSONConfigurationIntegration(t *testing.T) {
	// Test JSON configuration import/export functionality
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	translationService := services.NewTranslationService(db, logger, "")
	cacheService := services.NewCacheService(db, logger)
	localizationService := services.NewLocalizationService(db, logger, translationService, cacheService)

	userID := int64(200)

	t.Run("ExportConfiguration", func(t *testing.T) {
		// Setup initial localization
		localizationReq := &services.WizardLocalizationStep{
			UserID:                userID,
			PrimaryLanguage:       "en",
			SecondaryLanguages:    []string{"es", "fr"},
			SubtitleLanguages:     []string{"en", "es", "fr"},
			AutoTranslate:         true,
			AutoDownloadSubtitles: true,
			PreferredRegion:       "US",
			CurrencyCode:          "USD",
			DateFormat:            "MM/DD/YYYY",
			TimeFormat:            "12h",
			Timezone:              "America/New_York",
		}

		_, err := localizationService.SetupUserLocalization(context.Background(), localizationReq)
		require.NoError(t, err)

		// Export configuration
		config, err := localizationService.ExportConfiguration(context.Background(), userID, "full", "Test export", []string{"test", "integration"})
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify export structure
		assert.Equal(t, "1.0", config.Version)
		assert.Equal(t, userID, config.ExportedBy)
		assert.Equal(t, "full", config.ConfigType)
		assert.Equal(t, "Test export", config.Description)
		assert.Contains(t, config.Tags, "test")
		assert.Contains(t, config.Tags, "integration")

		// Verify localization data
		assert.NotNil(t, config.Localization)
		assert.Equal(t, userID, config.Localization.UserID)
		assert.Equal(t, "en", config.Localization.PrimaryLanguage)
		assert.Contains(t, config.Localization.SecondaryLanguages, "es")
		assert.Contains(t, config.Localization.SecondaryLanguages, "fr")

		// Verify media settings are included
		assert.NotNil(t, config.MediaSettings)
		assert.NotEmpty(t, config.MediaSettings.PlaybackSettings)
		assert.NotEmpty(t, config.MediaSettings.VideoSettings)
		assert.NotEmpty(t, config.MediaSettings.AudioSettings)
	})

	t.Run("ValidateConfiguration", func(t *testing.T) {
		// Test valid configuration
		validConfig := map[string]interface{}{
			"version":      "1.0",
			"config_type":  "localization",
			"localization": map[string]interface{}{
				"user_id":           userID,
				"primary_language":  "de",
				"secondary_languages": []string{"en"},
				"preferred_region":  "DE",
				"currency_code":     "EUR",
			},
		}

		validJSON, _ := json.Marshal(validConfig)
		validation := localizationService.ValidateConfigurationJSON(context.Background(), string(validJSON))

		assert.True(t, validation.Valid)
		assert.Empty(t, validation.Errors)
		assert.NotEmpty(t, validation.Summary)

		// Test invalid configuration
		invalidConfig := map[string]interface{}{
			"version": "invalid",
			"localization": map[string]interface{}{
				"primary_language": "",
			},
		}

		invalidJSON, _ := json.Marshal(invalidConfig)
		invalidValidation := localizationService.ValidateConfigurationJSON(context.Background(), string(invalidJSON))

		assert.False(t, invalidValidation.Valid)
		assert.NotEmpty(t, invalidValidation.Errors)
	})

	t.Run("ImportConfiguration", func(t *testing.T) {
		// Create configuration to import
		importConfig := services.ConfigurationExport{
			Version:    "1.0",
			ConfigType: "localization",
			Localization: &services.UserLocalization{
				UserID:            userID + 1,
				PrimaryLanguage:   "de",
				SecondaryLanguages: []string{"en", "es"},
				SubtitleLanguages: []string{"de", "en"},
				AutoTranslate:     false,
				PreferredRegion:   "DE",
				CurrencyCode:      "EUR",
				DateFormat:        "DD/MM/YYYY",
				TimeFormat:        "24h",
				Timezone:          "Europe/Berlin",
			},
			Description: "German configuration",
			Tags:        []string{"german", "europe"},
		}

		configJSON, _ := json.Marshal(importConfig)

		// Import with backup
		options := map[string]bool{
			"overwrite_existing": true,
			"backup_current":     true,
			"validate_only":      false,
		}

		result, err := localizationService.ImportConfiguration(context.Background(), userID+1, string(configJSON), options)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Success)
		assert.NotEmpty(t, result.ImportedAt)
		assert.NotEmpty(t, result.BackupID)
		assert.Equal(t, "localization", result.ConfigType)

		// Verify the configuration was imported
		imported, err := localizationService.GetUserLocalization(context.Background(), userID+1)
		require.NoError(t, err)
		assert.Equal(t, "de", imported.PrimaryLanguage)
		assert.Contains(t, imported.SecondaryLanguages, "en")
		assert.Equal(t, "EUR", imported.CurrencyCode)
	})

	t.Run("EditConfiguration", func(t *testing.T) {
		// Export current configuration
		current, err := localizationService.ExportConfiguration(context.Background(), userID, "localization", "", nil)
		require.NoError(t, err)

		currentJSON, _ := json.Marshal(current)

		// Define edits
		edits := map[string]interface{}{
			"localization.primary_language": "fr",
			"localization.currency_code":    "EUR",
			"localization.timezone":         "Europe/Paris",
			"description":                   "Updated French configuration",
		}

		// Apply edits
		editedJSON, err := localizationService.EditConfiguration(context.Background(), userID, string(currentJSON), edits)
		require.NoError(t, err)

		// Parse edited configuration
		var editedConfig services.ConfigurationExport
		err = json.Unmarshal([]byte(editedJSON), &editedConfig)
		require.NoError(t, err)

		// Verify edits were applied
		assert.Equal(t, "fr", editedConfig.Localization.PrimaryLanguage)
		assert.Equal(t, "EUR", editedConfig.Localization.CurrencyCode)
		assert.Equal(t, "Europe/Paris", editedConfig.Localization.Timezone)
		assert.Equal(t, "Updated French configuration", editedConfig.Description)
	})

	t.Run("ConfigurationTemplates", func(t *testing.T) {
		templates := localizationService.GetConfigurationTemplates(context.Background())
		require.NotEmpty(t, templates)

		// Verify template structure
		template := templates[0]
		assert.NotEmpty(t, template.Name)
		assert.NotEmpty(t, template.Description)
		assert.NotEmpty(t, template.Template)

		// Verify template can be parsed as valid JSON
		var templateConfig map[string]interface{}
		err := json.Unmarshal([]byte(template.Template), &templateConfig)
		require.NoError(t, err)

		// Verify template has required fields
		assert.NotEmpty(t, templateConfig["version"])
		assert.NotEmpty(t, templateConfig["config_type"])
	})
}

func TestEndToEndWorkflow(t *testing.T) {
	// Test complete workflow from user setup to media playback
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Setup all services
	translationService := services.NewTranslationService(db, logger, mockServer.URL())
	cacheService := services.NewCacheService(db, logger)
	localizationService := services.NewLocalizationService(db, logger, translationService, cacheService)
	subtitleService := services.NewSubtitleService(db, logger, mockServer.URL(), cacheService, translationService)
	lyricsService := services.NewLyricsService(db, logger, mockServer.URL(), cacheService, translationService)
	coverArtService := services.NewCoverArtService(db, logger, mockServer.URL(), cacheService)
	positionService := services.NewPlaybackPositionService(db, logger)
	playlistService := services.NewPlaylistService(db, logger)
	mediaPlayerService := services.NewMediaPlayerService(db, logger)
	videoPlayerService := services.NewVideoPlayerService(db, logger, mediaPlayerService, positionService, subtitleService, coverArtService, translationService)

	t.Run("CompleteUserWorkflow", func(t *testing.T) {
		userID := int64(100)

		// 1. Setup user localization
		localizationReq := &services.WizardLocalizationStep{
			UserID:                userID,
			PrimaryLanguage:       "en",
			SecondaryLanguages:    []string{"es"},
			SubtitleLanguages:     []string{"en", "es"},
			AutoTranslate:         true,
			AutoDownloadSubtitles: true,
			PreferredRegion:       "US",
		}

		localization, err := localizationService.SetupUserLocalization(context.Background(), localizationReq)
		require.NoError(t, err)
		assert.Equal(t, userID, localization.UserID)

		// 2. Create test video
		videoID := insertTestVideo(t, db, "Integration Test Movie", services.VideoTypeMovie)

		// 3. Start video playback
		playReq := &services.PlayVideoRequest{
			UserID:  userID,
			VideoID: videoID,
			PlayMode: services.VideoPlayModeSingle,
			Quality: services.Quality1080p,
			DeviceInfo: services.DeviceInfo{
				DeviceID:   "integration-test-device",
				DeviceName: "Integration Test Device",
			},
		}

		session, err := videoPlayerService.PlayVideo(context.Background(), playReq)
		require.NoError(t, err)
		require.NotNil(t, session)

		// 4. Search and download subtitles
		subtitleReq := &services.SubtitleSearchRequest{
			IMDbID:    "tt1234567",
			Languages: []string{"en"},
		}

		subtitleResults, err := subtitleService.SearchSubtitles(context.Background(), subtitleReq)
		require.NoError(t, err)
		require.NotEmpty(t, subtitleResults)

		downloadReq := &services.SubtitleDownloadRequest{
			SubtitleID: subtitleResults[0].ID,
			VideoID:    videoID,
			Language:   "en",
		}

		subtitle, err := subtitleService.DownloadSubtitle(context.Background(), downloadReq)
		require.NoError(t, err)
		require.NotNil(t, subtitle)

		// 5. Translate subtitles to user's secondary language
		translateReq := &services.TranslationRequest{
			SubtitleTrack:  subtitle,
			TargetLanguage: "es",
		}

		translatedSubtitle, err := subtitleService.TranslateSubtitle(context.Background(), translateReq)
		require.NoError(t, err)
		require.NotNil(t, translatedSubtitle)
		assert.Equal(t, "es", translatedSubtitle.Language)

		// 6. Update playback position
		updateReq := &services.UpdateVideoPlaybackRequest{
			SessionID: session.ID,
			Position:  int64Ptr(300000), // 5 minutes
			State:     &[]services.PlaybackState{services.StatePaused}[0],
		}

		updatedSession, err := videoPlayerService.UpdateVideoPlayback(context.Background(), updateReq)
		require.NoError(t, err)
		assert.Equal(t, int64(300000), updatedSession.Position)

		// 7. Create bookmark
		bookmarkReq := &services.CreateVideoBookmarkRequest{
			SessionID:   session.ID,
			Title:       "Important Scene",
			Description: "Key plot point",
		}

		bookmark, err := videoPlayerService.CreateVideoBookmark(context.Background(), bookmarkReq)
		require.NoError(t, err)
		assert.Equal(t, "Important Scene", bookmark.Title)

		// 8. Verify continue watching
		continueWatching, err := videoPlayerService.GetContinueWatching(context.Background(), userID, 10)
		require.NoError(t, err)

		// Should include our video
		found := false
		for _, video := range continueWatching {
			if video.ID == videoID {
				found = true
				break
			}
		}
		assert.True(t, found, "Video should appear in continue watching")

		// 9. Verify all mock services were called
		requests := mockServer.GetRequestLog()
		assert.True(t, len(requests) > 0, "Mock services should have been called")

		// Verify different API endpoints were hit
		endpoints := make(map[string]bool)
		for _, req := range requests {
			if req.URL != "" && req.URL != "/" {
				endpoints[req.URL] = true
			}
		}
		assert.True(t, len(endpoints) > 0, "Multiple API endpoints should have been called")
	})
}