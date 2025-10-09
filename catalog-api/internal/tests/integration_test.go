package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"catalog-api/internal/handlers"
	"catalog-api/internal/models"
	"catalog-api/internal/services"
)

// Integration test suite that tests the entire media recognition pipeline
func TestMediaRecognitionIntegration(t *testing.T) {
	ctx := context.Background()

	// Start all mock servers
	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	// Create test directory structure
	testDir := createTestMediaStructure(t)
	defer os.RemoveAll(testDir)

	// Initialize services with mock server URLs
	recognitionService := services.NewMediaRecognitionService()
	duplicationService := services.NewDuplicateDetectionService()
	readerService := services.NewReaderService()

	// Configure providers with mock URLs
	configureProvidersWithMockServers(recognitionService, mockServers)

	t.Run("full media library scan and recognition", func(t *testing.T) {
		// Scan the test directory
		mediaFiles, err := scanMediaDirectory(testDir)
		require.NoError(t, err)
		assert.True(t, len(mediaFiles) > 10, "Should find multiple media files")

		recognizedMedia := make([]*models.MediaMetadata, 0)

		// Recognize all media files
		for _, file := range mediaFiles {
			mediaType := determineMediaType(file)
			if mediaType == "" {
				continue
			}

			metadata, err := recognitionService.RecognizeMedia(ctx, file, mediaType)
			if err != nil {
				t.Logf("Failed to recognize %s: %v", file, err)
				continue
			}

			recognizedMedia = append(recognizedMedia, metadata)
		}

		assert.True(t, len(recognizedMedia) >= 8, "Should recognize most media files")

		// Verify recognition quality
		for _, metadata := range recognizedMedia {
			assert.NotEmpty(t, metadata.Title, "All media should have titles")
			assert.True(t, metadata.Confidence > 0.5, "Recognition confidence should be reasonable")
		}
	})

	t.Run("duplicate detection across media types", func(t *testing.T) {
		// Create test media with known duplicates
		testMedia := []*models.MediaMetadata{
			{
				Title:      "The Matrix",
				Year:       "1999",
				MediaType:  models.MediaTypeVideo,
				FilePath:   filepath.Join(testDir, "movies", "The.Matrix.1999.1080p.mkv"),
				FileSize:   2048000000,
				Format:     "mkv",
				Confidence: 0.95,
			},
			{
				Title:      "Matrix",
				Year:       "1999",
				MediaType:  models.MediaTypeVideo,
				FilePath:   filepath.Join(testDir, "movies", "Matrix.1999.DVDRip.avi"),
				FileSize:   700000000,
				Format:     "avi",
				Confidence: 0.88,
			},
			{
				Title:        "Bohemian Rhapsody",
				Artist:       "Queen",
				Album:        "A Night at the Opera",
				Year:         "1975",
				MediaType:    models.MediaTypeAudio,
				FilePath:     filepath.Join(testDir, "music", "Queen - Bohemian Rhapsody.mp3"),
				FileSize:     14200000,
				Format:       "mp3",
				Confidence:   0.96,
			},
			{
				Title:        "Bohemian Rhapsody",
				Artist:       "Queen",
				Album:        "Greatest Hits",
				Year:         "1981",
				MediaType:    models.MediaTypeAudio,
				FilePath:     filepath.Join(testDir, "music", "Queen Greatest Hits - Bohemian Rhapsody.flac"),
				FileSize:     42600000,
				Format:       "flac",
				Confidence:   0.94,
			},
		}

		// Find duplicates
		duplicates := duplicationService.FindAllDuplicates(testMedia)
		assert.True(t, len(duplicates) >= 2, "Should find at least 2 duplicate pairs")

		// Verify duplicate detection results
		for _, duplicate := range duplicates {
			assert.True(t, duplicate.Similarity > 0.8, "Duplicates should have high similarity")
			assert.NotEqual(t, duplicate.Original.FilePath, duplicate.Duplicate.FilePath, "Duplicates should have different file paths")
		}
	})

	t.Run("reader service integration", func(t *testing.T) {
		userID := "integration-test-user"
		bookPath := filepath.Join(testDir, "books", "Harry Potter.pdf")

		// Open book
		session, err := readerService.OpenBook(ctx, userID, bookPath)
		require.NoError(t, err)

		// Simulate reading session
		positions := []*models.ReadingPosition{
			{Page: 1, Word: 0, Character: 0, Timestamp: time.Now().Add(-2 * time.Hour)},
			{Page: 10, Word: 500, Character: 7500, Timestamp: time.Now().Add(-1 * time.Hour)},
			{Page: 25, Word: 1250, Character: 18750, Timestamp: time.Now()},
		}

		for _, pos := range positions {
			err := readerService.UpdatePosition(ctx, session.ID, pos)
			require.NoError(t, err)
		}

		// Add bookmarks and highlights
		bookmark := &models.Bookmark{
			Page:      15,
			Position:  750,
			Title:     "Important Chapter",
			Note:      "Key concept",
			Timestamp: time.Now(),
		}
		_, err = readerService.AddBookmark(ctx, session.ID, bookmark)
		require.NoError(t, err)

		highlight := &models.Highlight{
			StartPage:     20,
			EndPage:       20,
			StartPosition: 1000,
			EndPosition:   1100,
			Text:          "Important highlighted text",
			Color:         "Yellow",
			Timestamp:     time.Now(),
		}
		_, err = readerService.AddHighlight(ctx, session.ID, highlight)
		require.NoError(t, err)

		// Verify reading analytics
		analytics, err := readerService.GetReadingAnalytics(ctx, userID, 7)
		require.NoError(t, err)
		assert.True(t, analytics.TotalReadingTime > 0)
	})
}

func TestAPIEndpointsIntegration(t *testing.T) {
	// Start mock servers
	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	// Initialize services
	recognitionService := services.NewMediaRecognitionService()
	duplicationService := services.NewDuplicateDetectionService()
	readerService := services.NewReaderService()

	configureProvidersWithMockServers(recognitionService, mockServers)

	// Initialize handlers
	mediaHandler := handlers.NewMediaHandler(recognitionService, duplicationService)
	readerHandler := handlers.NewReaderHandler(readerService)

	// Setup router
	router := mux.NewRouter()
	router.HandleFunc("/api/media/recognize", mediaHandler.RecognizeMedia).Methods("POST")
	router.HandleFunc("/api/media/duplicates", mediaHandler.FindDuplicates).Methods("POST")
	router.HandleFunc("/api/reader/sessions", readerHandler.CreateSession).Methods("POST")
	router.HandleFunc("/api/reader/sessions/{id}/position", readerHandler.UpdatePosition).Methods("PUT")
	router.HandleFunc("/api/reader/sessions/{id}/bookmarks", readerHandler.AddBookmark).Methods("POST")

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("media recognition API", func(t *testing.T) {
		requestBody := `{
			"file_path": "/movies/The.Matrix.1999.1080p.BluRay.x264.mkv",
			"media_type": "video"
		}`

		resp, err := http.Post(server.URL+"/api/media/recognize", "application/json", strings.NewReader(requestBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result models.MediaMetadata
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "The Matrix", result.Title)
		assert.Equal(t, "1999", result.Year)
		assert.True(t, result.Confidence > 0.9)
	})

	t.Run("duplicate detection API", func(t *testing.T) {
		requestBody := `{
			"media_items": [
				{
					"title": "The Matrix",
					"year": "1999",
					"media_type": "video",
					"file_path": "/movies/Matrix1.mkv"
				},
				{
					"title": "Matrix",
					"year": "1999",
					"media_type": "video",
					"file_path": "/movies/Matrix2.avi"
				}
			]
		}`

		resp, err := http.Post(server.URL+"/api/media/duplicates", "application/json", strings.NewReader(requestBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []models.DuplicatePair
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, len(result) >= 1, "Should find duplicate pairs")
		if len(result) > 0 {
			assert.True(t, result[0].Similarity > 0.8)
		}
	})

	t.Run("reader session API", func(t *testing.T) {
		// Create session
		sessionBody := `{
			"user_id": "api-test-user",
			"book_path": "/books/Test Book.pdf",
			"device_id": "test-device"
		}`

		resp, err := http.Post(server.URL+"/api/reader/sessions", "application/json", strings.NewReader(sessionBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var session models.ReadingSession
		err = json.NewDecoder(resp.Body).Decode(&session)
		require.NoError(t, err)

		assert.NotEmpty(t, session.ID)
		assert.Equal(t, "api-test-user", session.UserID)

		// Update position
		positionBody := `{
			"page": 10,
			"word": 500,
			"character": 7500,
			"timestamp": "` + time.Now().Format(time.RFC3339) + `"
		}`

		req, err := http.NewRequest("PUT", server.URL+"/api/reader/sessions/"+session.ID+"/position", strings.NewReader(positionBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestPerformanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	ctx := context.Background()

	// Start mock servers
	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	recognitionService := services.NewMediaRecognitionService()
	configureProvidersWithMockServers(recognitionService, mockServers)

	t.Run("concurrent media recognition", func(t *testing.T) {
		mediaPaths := []string{
			"/movies/Movie1.2023.mkv",
			"/movies/Movie2.2022.mkv",
			"/movies/Movie3.2021.mkv",
			"/music/Song1.mp3",
			"/music/Song2.mp3",
			"/music/Song3.mp3",
			"/books/Book1.pdf",
			"/books/Book2.epub",
			"/games/Game1.exe",
			"/software/App1.exe",
		}

		start := time.Now()

		// Process all media concurrently
		results := make(chan *models.MediaMetadata, len(mediaPaths))
		errors := make(chan error, len(mediaPaths))

		for _, path := range mediaPaths {
			go func(p string) {
				mediaType := determineMediaType(p)
				result, err := recognitionService.RecognizeMedia(ctx, p, mediaType)
				if err != nil {
					errors <- err
					return
				}
				results <- result
			}(path)
		}

		// Collect results
		successCount := 0
		for i := 0; i < len(mediaPaths); i++ {
			select {
			case result := <-results:
				assert.NotNil(t, result)
				successCount++
			case err := <-errors:
				t.Logf("Recognition error: %v", err)
			case <-time.After(30 * time.Second):
				t.Fatal("Performance test timed out")
			}
		}

		duration := time.Since(start)
		t.Logf("Processed %d/%d files in %v", successCount, len(mediaPaths), duration)

		assert.True(t, successCount >= len(mediaPaths)/2, "Should successfully process at least half the files")
		assert.True(t, duration < 15*time.Second, "Should complete within 15 seconds")
	})

	t.Run("bulk duplicate detection", func(t *testing.T) {
		duplicationService := services.NewDuplicateDetectionService()

		// Generate test media items
		mediaItems := make([]*models.MediaMetadata, 100)
		for i := 0; i < 100; i++ {
			mediaItems[i] = &models.MediaMetadata{
				Title:      fmt.Sprintf("Test Media %d", i%20), // Create some duplicates
				Year:       "2023",
				Genre:      "Test",
				MediaType:  models.MediaTypeVideo,
				Confidence: 0.9,
			}
		}

		start := time.Now()
		duplicates := duplicationService.FindAllDuplicates(mediaItems)
		duration := time.Since(start)

		t.Logf("Found %d duplicate pairs in %v", len(duplicates), duration)

		assert.True(t, len(duplicates) > 0, "Should find some duplicates")
		assert.True(t, duration < 10*time.Second, "Bulk detection should complete within 10 seconds")
	})
}

func TestErrorHandlingIntegration(t *testing.T) {
	ctx := context.Background()

	// Initialize services without mock servers to test error handling
	recognitionService := services.NewMediaRecognitionService()
	duplicationService := services.NewDuplicateDetectionService()
	readerService := services.NewReaderService()

	t.Run("graceful handling of API failures", func(t *testing.T) {
		// Test with unavailable external APIs
		result, err := recognitionService.RecognizeMedia(ctx, "/movies/Unknown.Movie.mkv", models.MediaTypeVideo)

		// Should not crash, should return some basic info
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Unknown Movie", result.Title)
		assert.True(t, result.Confidence < 0.5)
	})

	t.Run("handling invalid file paths", func(t *testing.T) {
		invalidPaths := []string{
			"",
			"/nonexistent/path/file.mkv",
			"invalid-path",
			"/path/with/invalid\x00char.mp3",
		}

		for _, path := range invalidPaths {
			result, err := recognitionService.RecognizeMedia(ctx, path, models.MediaTypeVideo)
			assert.Error(t, err, "Should return error for invalid path: %s", path)
			assert.Nil(t, result)
		}
	})

	t.Run("handling corrupted media files", func(t *testing.T) {
		// Create temporary corrupted files
		tmpDir := t.TempDir()
		corruptedFile := filepath.Join(tmpDir, "corrupted.mp3")

		// Write invalid content
		err := os.WriteFile(corruptedFile, []byte("invalid audio content"), 0644)
		require.NoError(t, err)

		// Should handle gracefully
		result, err := recognitionService.RecognizeMedia(ctx, corruptedFile, models.MediaTypeAudio)
		if err != nil {
			// Error is acceptable for corrupted files
			assert.Nil(t, result)
		} else {
			// If no error, should have low confidence
			assert.True(t, result.Confidence < 0.7)
		}
	})

	t.Run("reader service error handling", func(t *testing.T) {
		userID := "error-test-user"

		// Try to open non-existent book
		session, err := readerService.OpenBook(ctx, userID, "/nonexistent/book.pdf")
		assert.Error(t, err)
		assert.Nil(t, session)

		// Try to update position for non-existent session
		position := &models.ReadingPosition{Page: 1}
		err = readerService.UpdatePosition(ctx, "invalid-session-id", position)
		assert.Error(t, err)
	})
}

// Helper functions
func createTestMediaStructure(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create directory structure
	dirs := []string{
		"movies",
		"music",
		"books",
		"games",
		"software",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
		require.NoError(t, err)
	}

	// Create test files
	testFiles := map[string]string{
		"movies/The.Matrix.1999.1080p.BluRay.x264.mkv":     "fake movie content",
		"movies/Breaking.Bad.S01E01.Pilot.1080p.mkv":       "fake tv content",
		"movies/Matrix.1999.DVDRip.XviD.avi":                "fake movie duplicate",
		"music/Queen - Bohemian Rhapsody.mp3":              "fake audio content",
		"music/Queen Greatest Hits - Bohemian Rhapsody.flac": "fake audio duplicate",
		"music/Led Zeppelin - Stairway to Heaven.mp3":      "fake audio content",
		"books/Harry Potter and the Philosopher's Stone.pdf": "fake book content",
		"books/The Lord of the Rings.epub":                 "fake book content",
		"books/Programming in Go.pdf":                      "fake book content",
		"games/Cyberpunk 2077.exe":                         "fake game content",
		"software/Visual Studio Code.exe":                  "fake software content",
	}

	for fileName, content := range testFiles {
		filePath := filepath.Join(tmpDir, fileName)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	return tmpDir
}

func scanMediaDirectory(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func determineMediaType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".mp4", ".mkv", ".avi", ".mov", ".wmv":
		return models.MediaTypeVideo
	case ".mp3", ".flac", ".wav", ".aac", ".ogg":
		return models.MediaTypeAudio
	case ".pdf", ".epub", ".mobi", ".txt":
		return models.MediaTypeBook
	case ".exe", ".msi", ".dmg", ".deb", ".rpm":
		if strings.Contains(strings.ToLower(filePath), "game") {
			return models.MediaTypeGame
		}
		return models.MediaTypeSoftware
	default:
		return ""
	}
}

func configureProvidersWithMockServers(recognitionService *services.MediaRecognitionService, mockServers []*httptest.Server) {
	// This would normally configure the service providers with mock URLs
	// Implementation depends on how the service exposes configuration
	// For now, this is a placeholder for the configuration logic
}