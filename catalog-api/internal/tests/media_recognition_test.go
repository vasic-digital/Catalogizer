package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"catalog-api/internal/models"
	"catalog-api/internal/services"
)

func TestMediaRecognitionService_Movies(t *testing.T) {
	ctx := context.Background()

	// Start all mock servers
	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	// Create recognition service with mock server URLs
	recognitionService := services.NewMediaRecognitionService()
	movieProvider := services.NewMovieRecognitionProvider()
	movieProvider.TMDbBaseURL = mockServers[0].URL
	movieProvider.OMDbBaseURL = mockServers[1].URL

	// Test movie recognition
	t.Run("recognize movie file", func(t *testing.T) {
		mediaPath := "/movies/The.Matrix.1999.1080p.BluRay.x264.mkv"
		mediaType := models.MediaTypeVideo

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "The Matrix", result.Title)
		assert.Equal(t, "1999", result.Year)
		assert.Equal(t, "Science Fiction", result.Genre)
		assert.Equal(t, 8.7, result.Rating)
		assert.NotEmpty(t, result.CoverArt)
		assert.True(t, result.Confidence > 0.9)
	})

	t.Run("recognize TV series", func(t *testing.T) {
		mediaPath := "/tv/Breaking.Bad.S01E01.Pilot.1080p.mkv"
		mediaType := models.MediaTypeVideo

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Breaking Bad", result.Title)
		assert.Equal(t, "Pilot", result.Episode)
		assert.Equal(t, 1, result.Season)
		assert.Equal(t, 1, result.EpisodeNumber)
		assert.Equal(t, "Drama", result.Genre)
		assert.True(t, result.Confidence > 0.85)
	})
}

func TestMediaRecognitionService_Music(t *testing.T) {
	ctx := context.Background()

	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	musicProvider := services.NewMusicRecognitionProvider()
	musicProvider.LastFmBaseURL = mockServers[2].URL
	musicProvider.MusicBrainzBaseURL = mockServers[3].URL
	musicProvider.AcoustIDBaseURL = mockServers[4].URL

	recognitionService := services.NewMediaRecognitionService()

	t.Run("recognize audio file with metadata", func(t *testing.T) {
		mediaPath := "/music/Queen - Bohemian Rhapsody.mp3"
		mediaType := models.MediaTypeAudio

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Bohemian Rhapsody", result.Title)
		assert.Equal(t, "Queen", result.Artist)
		assert.Equal(t, "A Night at the Opera", result.Album)
		assert.Equal(t, "1975", result.Year)
		assert.Equal(t, "Rock", result.Genre)
		assert.NotEmpty(t, result.CoverArt)
		assert.True(t, result.Confidence > 0.8)
	})

	t.Run("recognize audio by fingerprint", func(t *testing.T) {
		tempFile := createTempAudioFile(t)
		defer os.Remove(tempFile)

		result, err := recognitionService.RecognizeMedia(ctx, tempFile, models.MediaTypeAudio)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Stairway to Heaven", result.Title)
		assert.Equal(t, "Led Zeppelin", result.Artist)
		assert.True(t, result.Confidence > 0.7)
	})
}

func TestMediaRecognitionService_Games(t *testing.T) {
	ctx := context.Background()

	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	gameProvider := services.NewGameSoftwareRecognitionProvider()
	gameProvider.IGDBBaseURL = mockServers[5].URL
	gameProvider.SteamBaseURL = mockServers[6].URL
	gameProvider.GitHubBaseURL = mockServers[7].URL

	recognitionService := services.NewMediaRecognitionService()

	t.Run("recognize game executable", func(t *testing.T) {
		mediaPath := "/games/Cyberpunk 2077/bin/x64/Cyberpunk2077.exe"
		mediaType := models.MediaTypeGame

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Cyberpunk 2077", result.Title)
		assert.Equal(t, "CD Projekt RED", result.Developer)
		assert.Equal(t, "2020", result.Year)
		assert.Equal(t, "RPG", result.Genre)
		assert.Equal(t, "PC", result.Platform)
		assert.True(t, result.Confidence > 0.9)
	})

	t.Run("recognize software by name", func(t *testing.T) {
		mediaPath := "/software/Visual Studio Code/code.exe"
		mediaType := models.MediaTypeSoftware

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Visual Studio Code", result.Title)
		assert.Equal(t, "Microsoft", result.Developer)
		assert.Equal(t, "Code Editor", result.Category)
		assert.True(t, result.Confidence > 0.8)
	})
}

func TestMediaRecognitionService_Books(t *testing.T) {
	ctx := context.Background()

	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	bookProvider := services.NewBookRecognitionProvider()
	bookProvider.GoogleBooksBaseURL = mockServers[8].URL
	bookProvider.OpenLibraryBaseURL = mockServers[9].URL
	bookProvider.CrossrefBaseURL = mockServers[10].URL
	bookProvider.OCRSpaceBaseURL = mockServers[11].URL

	recognitionService := services.NewMediaRecognitionService()

	t.Run("recognize book by filename", func(t *testing.T) {
		mediaPath := "/books/Harry Potter and the Philosopher's Stone.pdf"
		mediaType := models.MediaTypeBook

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Harry Potter and the Philosopher's Stone", result.Title)
		assert.Equal(t, "J.K. Rowling", result.Author)
		assert.Equal(t, "9780747532699", result.ISBN)
		assert.Equal(t, "1997", result.Year)
		assert.Equal(t, "Fantasy", result.Genre)
		assert.True(t, result.Confidence > 0.9)
	})

	t.Run("recognize book with OCR", func(t *testing.T) {
		tempBook := createTempBookFile(t)
		defer os.Remove(tempBook)

		result, err := recognitionService.RecognizeMedia(ctx, tempBook, models.MediaTypeBook)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "The Lord of the Rings", result.Title)
		assert.Equal(t, "J.R.R. Tolkien", result.Author)
		assert.True(t, result.Confidence > 0.8)
	})
}

func TestMediaRecognitionService_DuplicateDetection(t *testing.T) {
	ctx := context.Background()

	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	recognitionService := services.NewMediaRecognitionService()
	duplicationService := services.NewDuplicateDetectionService()

	t.Run("detect movie duplicates", func(t *testing.T) {
		// Recognize original movie
		originalPath := "/movies/The.Matrix.1999.1080p.BluRay.x264.mkv"
		original, err := recognitionService.RecognizeMedia(ctx, originalPath, models.MediaTypeVideo)
		require.NoError(t, err)

		// Recognize potential duplicate
		duplicatePath := "/movies/Matrix.1999.DVDRip.XviD.avi"
		duplicate, err := recognitionService.RecognizeMedia(ctx, duplicatePath, models.MediaTypeVideo)
		require.NoError(t, err)

		// Check for duplicates
		isDuplicate, similarity := duplicationService.IsDuplicate(original, duplicate)
		assert.True(t, isDuplicate)
		assert.True(t, similarity > 0.8)
	})

	t.Run("detect music album duplicates", func(t *testing.T) {
		// Recognize original album
		originalPath := "/music/Queen - A Night at the Opera - 01 - Bohemian Rhapsody.mp3"
		original, err := recognitionService.RecognizeMedia(ctx, originalPath, models.MediaTypeAudio)
		require.NoError(t, err)

		// Recognize potential duplicate
		duplicatePath := "/music/Queen/A Night at the Opera/Bohemian Rhapsody.flac"
		duplicate, err := recognitionService.RecognizeMedia(ctx, duplicatePath, models.MediaTypeAudio)
		require.NoError(t, err)

		// Check for duplicates
		isDuplicate, similarity := duplicationService.IsDuplicate(original, duplicate)
		assert.True(t, isDuplicate)
		assert.True(t, similarity > 0.9)
	})
}

func TestMediaRecognitionService_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Create recognition service without mock servers to test fallbacks
	recognitionService := services.NewMediaRecognitionService()

	t.Run("handle unavailable API gracefully", func(t *testing.T) {
		mediaPath := "/movies/Unknown.Movie.2023.mkv"
		mediaType := models.MediaTypeVideo

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)

		// Should not error, but return basic info extracted from filename
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Unknown Movie", result.Title)
		assert.Equal(t, "2023", result.Year)
		assert.True(t, result.Confidence < 0.5) // Low confidence for filename-only recognition
	})

	t.Run("handle invalid file path", func(t *testing.T) {
		mediaPath := ""
		mediaType := models.MediaTypeVideo

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("handle unsupported media type", func(t *testing.T) {
		mediaPath := "/unknown/file.xyz"
		mediaType := "unsupported"

		result, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMediaRecognitionService_Performance(t *testing.T) {
	ctx := context.Background()

	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	recognitionService := services.NewMediaRecognitionService()

	t.Run("concurrent recognition", func(t *testing.T) {
		mediaPaths := []string{
			"/movies/Movie1.2023.mkv",
			"/movies/Movie2.2022.mkv",
			"/movies/Movie3.2021.mkv",
			"/music/Song1.mp3",
			"/music/Song2.mp3",
			"/books/Book1.pdf",
		}

		results := make(chan *models.MediaMetadata, len(mediaPaths))
		errors := make(chan error, len(mediaPaths))

		start := time.Now()

		for _, path := range mediaPaths {
			go func(p string) {
				var mediaType string
				if strings.Contains(p, "/movies/") {
					mediaType = models.MediaTypeVideo
				} else if strings.Contains(p, "/music/") {
					mediaType = models.MediaTypeAudio
				} else {
					mediaType = models.MediaTypeBook
				}

				result, err := recognitionService.RecognizeMedia(ctx, p, mediaType)
				if err != nil {
					errors <- err
					return
				}
				results <- result
			}(path)
		}

		// Collect results
		for i := 0; i < len(mediaPaths); i++ {
			select {
			case result := <-results:
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Title)
			case err := <-errors:
				t.Errorf("Recognition error: %v", err)
			case <-time.After(30 * time.Second):
				t.Fatal("Recognition timed out")
			}
		}

		duration := time.Since(start)
		t.Logf("Concurrent recognition took: %v", duration)
		assert.True(t, duration < 10*time.Second, "Recognition should complete within 10 seconds")
	})
}

func TestReaderService_Integration(t *testing.T) {
	ctx := context.Background()

	readerService := services.NewReaderService()

	t.Run("track reading position", func(t *testing.T) {
		userID := "test-user-123"
		bookPath := "/books/Test Book.pdf"

		// Open book
		session, err := readerService.OpenBook(ctx, userID, bookPath)
		require.NoError(t, err)
		require.NotNil(t, session)

		// Update reading position
		position := &models.ReadingPosition{
			Page:      10,
			Word:      150,
			Character: 2450,
			CFI:       "epubcfi(/6/4[chapter01]!/4/2/1:245)",
			Timestamp: time.Now(),
		}

		err = readerService.UpdatePosition(ctx, session.ID, position)
		require.NoError(t, err)

		// Retrieve reading position
		retrievedPosition, err := readerService.GetPosition(ctx, session.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedPosition)

		assert.Equal(t, position.Page, retrievedPosition.Page)
		assert.Equal(t, position.Word, retrievedPosition.Word)
		assert.Equal(t, position.Character, retrievedPosition.Character)
		assert.Equal(t, position.CFI, retrievedPosition.CFI)
	})

	t.Run("sync across devices", func(t *testing.T) {
		userID := "test-user-123"
		bookPath := "/books/Test Book.pdf"

		// Open book on device 1
		session1, err := readerService.OpenBookOnDevice(ctx, userID, bookPath, "device-1")
		require.NoError(t, err)

		// Open book on device 2
		session2, err := readerService.OpenBookOnDevice(ctx, userID, bookPath, "device-2")
		require.NoError(t, err)

		// Update position on device 1
		position1 := &models.ReadingPosition{
			Page:      20,
			Timestamp: time.Now(),
		}
		err = readerService.UpdatePosition(ctx, session1.ID, position1)
		require.NoError(t, err)

		// Sync to device 2
		err = readerService.SyncPosition(ctx, session2.ID)
		require.NoError(t, err)

		// Check position on device 2
		syncedPosition, err := readerService.GetPosition(ctx, session2.ID)
		require.NoError(t, err)
		assert.Equal(t, position1.Page, syncedPosition.Page)
	})
}

// Helper functions
func createTempAudioFile(t *testing.T) string {
	tmpDir := t.TempDir()
	audioFile := filepath.Join(tmpDir, "test_audio.mp3")

	// Create a dummy MP3 file with some basic metadata
	content := []byte("ID3\x03\x00\x00\x00\x00\x00\x00" + "fake audio data for testing")
	err := os.WriteFile(audioFile, content, 0644)
	require.NoError(t, err)

	return audioFile
}

func createTempBookFile(t *testing.T) string {
	tmpDir := t.TempDir()
	bookFile := filepath.Join(tmpDir, "test_book.pdf")

	// Create a dummy PDF file
	pdfHeader := "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj\nxref\n0 1\n0000000000 65535 f \ntrailer\n<<\n/Size 1\n/Root 1 0 R\n>>\nstartxref\n9\n%%EOF"
	err := os.WriteFile(bookFile, []byte(pdfHeader), 0644)
	require.NoError(t, err)

	return bookFile
}

func BenchmarkMediaRecognition(b *testing.B) {
	ctx := context.Background()

	mockServers := StartAllMockServers()
	defer func() {
		for _, server := range mockServers {
			server.Close()
		}
	}()

	recognitionService := services.NewMediaRecognitionService()

	b.Run("movie recognition", func(b *testing.B) {
		mediaPath := "/movies/The.Matrix.1999.1080p.BluRay.x264.mkv"
		mediaType := models.MediaTypeVideo

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
			if err != nil {
				b.Fatalf("Recognition failed: %v", err)
			}
		}
	})

	b.Run("music recognition", func(b *testing.B) {
		mediaPath := "/music/Queen - Bohemian Rhapsody.mp3"
		mediaType := models.MediaTypeAudio

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := recognitionService.RecognizeMedia(ctx, mediaPath, mediaType)
			if err != nil {
				b.Fatalf("Recognition failed: %v", err)
			}
		}
	})
}