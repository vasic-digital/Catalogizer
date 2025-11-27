package tests

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"catalogizer/internal/models"
	"catalogizer/internal/services"
	_ "github.com/mattn/go-sqlite3"
)

func TestDuplicateDetectionService_TextSimilarity(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := services.NewDuplicateDetectionService(db, logger, nil)

	testCases := []struct {
		text1    string
		text2    string
		expected float64
		name     string
	}{
		{
			text1:    "The Matrix",
			text2:    "The Matrix",
			expected: 1.0,
			name:     "identical strings",
		},
		{
			text1:    "The Matrix",
			text2:    "Matrix",
			expected: 0.8,
			name:     "partial match",
		},
		{
			text1:    "The.Matrix.1999.1080p.BluRay.x264",
			text2:    "Matrix 1999 DVDRip XviD",
			expected: 0.7,
			name:     "different formats same movie",
		},
		{
			text1:    "Queen - Bohemian Rhapsody",
			text2:    "Bohemian Rhapsody by Queen",
			expected: 0.75,
			name:     "artist and title rearranged",
		},
		{
			text1:    "Harry Potter and the Philosopher's Stone",
			text2:    "Harry Potter and the Sorcerer's Stone",
			expected: 0.85,
			name:     "slight title variation",
		},
		{
			text1:    "Completely Different Title",
			text2:    "Another Unrelated Work",
			expected: 0.1,
			name:     "unrelated content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			similarity := service.CalculateTextSimilarity(tc.text1, tc.text2)
			assert.InDelta(t, tc.expected, similarity, 0.15, "Similarity should be within expected range")
		})
	}
}

func TestDuplicateDetectionService_MovieDuplicates(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := services.NewDuplicateDetectionService(db, logger, nil)

	year := 1999
	duration := 136
	fileSize := int64(2048000000) // 2GB
	originalMovie := &models.MediaMetadata{
		Title:      "The Matrix",
		Year:       &year,
		Genre:      "Science Fiction",
		Director:   "The Wachowskis",
		Duration:   &duration,
		Resolution: "1080p",
		FileSize:   &fileSize,
		Language:   "English",
		MediaType:  models.MediaTypeVideo,
	}

	testCases := []struct {
		duplicate          *models.MediaMetadata
		expectedMatch      bool
		expectedSimilarity float64
		name               string
	}{
		{
			duplicate: &models.MediaMetadata{
				Title:      "The Matrix",
				Year:       "1999",
				Genre:      "Science Fiction",
				Director:   "The Wachowskis",
				Duration:   136,
				Resolution: "720p",
				FileSize:   1024000000, // 1GB
				Format:     "avi",
				Language:   "English",
				MediaType:  models.MediaTypeVideo,
				Confidence: 0.92,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.95,
			name:               "same movie different quality",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "Matrix",
				Year:       "1999",
				Genre:      "Sci-Fi",
				Director:   "Wachowski Brothers",
				Duration:   136,
				Resolution: "1080p",
				FileSize:   2048000000,
				Format:     "mp4",
				Language:   "English",
				MediaType:  models.MediaTypeVideo,
				Confidence: 0.88,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.85,
			name:               "slight variations in metadata",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "The Matrix Reloaded",
				Year:       "2003",
				Genre:      "Science Fiction",
				Director:   "The Wachowskis",
				Duration:   138,
				Resolution: "1080p",
				FileSize:   2148000000,
				Format:     "mkv",
				Language:   "English",
				MediaType:  models.MediaTypeVideo,
				Confidence: 0.94,
			},
			expectedMatch:      false,
			expectedSimilarity: 0.6,
			name:               "sequel - not duplicate",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "Inception",
				Year:       "2010",
				Genre:      "Science Fiction",
				Director:   "Christopher Nolan",
				Duration:   148,
				Resolution: "1080p",
				FileSize:   2200000000,
				Format:     "mkv",
				Language:   "English",
				MediaType:  models.MediaTypeVideo,
				Confidence: 0.96,
			},
			expectedMatch:      false,
			expectedSimilarity: 0.2,
			name:               "different movie",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isDuplicate, similarity := service.IsDuplicate(originalMovie, tc.duplicate)
			assert.Equal(t, tc.expectedMatch, isDuplicate, "Duplicate detection result should match expected")
			assert.InDelta(t, tc.expectedSimilarity, similarity, 0.15, "Similarity should be within expected range")
		})
	}
}

func TestDuplicateDetectionService_MusicDuplicates(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(db, logger, nil)

	originalSong := &models.MediaMetadata{
		Title:       "Bohemian Rhapsody",
		Artist:      "Queen",
		Album:       "A Night at the Opera",
		Year:        "1975",
		Genre:       "Rock",
		Duration:    355, // 5:55
		TrackNumber: 11,
		Bitrate:     320,
		SampleRate:  44100,
		FileSize:    14200000, // ~14MB
		Format:      "mp3",
		Language:    "English",
		MediaType:   models.MediaTypeAudio,
		Confidence:  0.96,
	}

	testCases := []struct {
		duplicate          *models.MediaMetadata
		expectedMatch      bool
		expectedSimilarity float64
		name               string
	}{
		{
			duplicate: &models.MediaMetadata{
				Title:       "Bohemian Rhapsody",
				Artist:      "Queen",
				Album:       "A Night at the Opera",
				Year:        "1975",
				Genre:       "Rock",
				Duration:    355,
				TrackNumber: 11,
				Bitrate:     128,
				SampleRate:  44100,
				FileSize:    7100000, // ~7MB
				Format:      "mp3",
				Language:    "English",
				MediaType:   models.MediaTypeAudio,
				Confidence:  0.92,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.98,
			name:               "same song different bitrate",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:       "Bohemian Rhapsody",
				Artist:      "Queen",
				Album:       "Greatest Hits",
				Year:        "1981",
				Genre:       "Rock",
				Duration:    355,
				TrackNumber: 1,
				Bitrate:     320,
				SampleRate:  44100,
				FileSize:    14200000,
				Format:      "flac",
				Language:    "English",
				MediaType:   models.MediaTypeAudio,
				Confidence:  0.94,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.92,
			name:               "same song different album compilation",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:       "Bohemian Rhapsody",
				Artist:      "Various Artists",
				Album:       "Rock Ballads Collection",
				Year:        "2020",
				Genre:       "Rock",
				Duration:    355,
				TrackNumber: 5,
				Bitrate:     256,
				SampleRate:  44100,
				FileSize:    11400000,
				Format:      "m4a",
				Language:    "English",
				MediaType:   models.MediaTypeAudio,
				Confidence:  0.85,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.88,
			name:               "same song in compilation",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:       "We Will Rock You",
				Artist:      "Queen",
				Album:       "News of the World",
				Year:        "1977",
				Genre:       "Rock",
				Duration:    122,
				TrackNumber: 1,
				Bitrate:     320,
				SampleRate:  44100,
				FileSize:    4900000,
				Format:      "mp3",
				Language:    "English",
				MediaType:   models.MediaTypeAudio,
				Confidence:  0.95,
			},
			expectedMatch:      false,
			expectedSimilarity: 0.4,
			name:               "different song same artist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isDuplicate, similarity := service.IsDuplicate(originalSong, tc.duplicate)
			assert.Equal(t, tc.expectedMatch, isDuplicate, "Duplicate detection result should match expected")
			assert.InDelta(t, tc.expectedSimilarity, similarity, 0.15, "Similarity should be within expected range")
		})
	}
}

func TestDuplicateDetectionService_BookDuplicates(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(db, logger, nil)

	originalBook := &models.MediaMetadata{
		Title:      "Harry Potter and the Philosopher's Stone",
		Author:     "J.K. Rowling",
		ISBN:       "9780747532699",
		Year:       "1997",
		Genre:      "Fantasy",
		Publisher:  "Bloomsbury",
		Language:   "English",
		Pages:      223,
		FileSize:   5200000, // ~5MB
		Format:     "pdf",
		MediaType:  models.MediaTypeBook,
		Confidence: 0.98,
	}

	testCases := []struct {
		duplicate          *models.MediaMetadata
		expectedMatch      bool
		expectedSimilarity float64
		name               string
	}{
		{
			duplicate: &models.MediaMetadata{
				Title:      "Harry Potter and the Sorcerer's Stone",
				Author:     "J.K. Rowling",
				ISBN:       "9780439708180",
				Year:       "1998",
				Genre:      "Fantasy",
				Publisher:  "Scholastic",
				Language:   "English",
				Pages:      309,
				FileSize:   7800000, // ~7.8MB
				Format:     "epub",
				MediaType:  models.MediaTypeBook,
				Confidence: 0.97,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.92,
			name:               "US vs UK edition",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "Harry Potter and the Philosopher's Stone",
				Author:     "J.K. Rowling",
				ISBN:       "9780747532699",
				Year:       "1997",
				Genre:      "Fantasy",
				Publisher:  "Bloomsbury",
				Language:   "English",
				Pages:      223,
				FileSize:   12400000, // ~12MB higher quality
				Format:     "pdf",
				MediaType:  models.MediaTypeBook,
				Confidence: 0.98,
			},
			expectedMatch:      true,
			expectedSimilarity: 1.0,
			name:               "same book different file size",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "Harry Potter y la Piedra Filosofal",
				Author:     "J.K. Rowling",
				ISBN:       "9788478884452",
				Year:       "1999",
				Genre:      "Fantasía",
				Publisher:  "Emecé",
				Language:   "Spanish",
				Pages:      254,
				FileSize:   6100000,
				Format:     "epub",
				MediaType:  models.MediaTypeBook,
				Confidence: 0.94,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.85,
			name:               "Spanish translation",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "Harry Potter and the Chamber of Secrets",
				Author:     "J.K. Rowling",
				ISBN:       "9780747538493",
				Year:       "1998",
				Genre:      "Fantasy",
				Publisher:  "Bloomsbury",
				Language:   "English",
				Pages:      251,
				FileSize:   5800000,
				Format:     "pdf",
				MediaType:  models.MediaTypeBook,
				Confidence: 0.97,
			},
			expectedMatch:      false,
			expectedSimilarity: 0.6,
			name:               "sequel - not duplicate",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isDuplicate, similarity := service.IsDuplicate(originalBook, tc.duplicate)
			assert.Equal(t, tc.expectedMatch, isDuplicate, "Duplicate detection result should match expected")
			assert.InDelta(t, tc.expectedSimilarity, similarity, 0.15, "Similarity should be within expected range")
		})
	}
}

func TestDuplicateDetectionService_GameDuplicates(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(db, logger, nil)

	originalGame := &models.MediaMetadata{
		Title:      "Cyberpunk 2077",
		Developer:  "CD Projekt RED",
		Publisher:  "CD Projekt",
		Year:       "2020",
		Genre:      "RPG",
		Platform:   "PC",
		Version:    "1.63",
		Language:   "English",
		FileSize:   70000000000, // 70GB
		MediaType:  models.MediaTypeGame,
		Confidence: 0.96,
	}

	testCases := []struct {
		duplicate          *models.MediaMetadata
		expectedMatch      bool
		expectedSimilarity float64
		name               string
	}{
		{
			duplicate: &models.MediaMetadata{
				Title:      "Cyberpunk 2077",
				Developer:  "CD Projekt RED",
				Publisher:  "CD Projekt",
				Year:       "2020",
				Genre:      "RPG",
				Platform:   "PS4",
				Version:    "1.63",
				Language:   "English",
				FileSize:   68000000000, // 68GB
				MediaType:  models.MediaTypeGame,
				Confidence: 0.95,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.95,
			name:               "same game different platform",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "Cyberpunk 2077: Phantom Liberty",
				Developer:  "CD Projekt RED",
				Publisher:  "CD Projekt",
				Year:       "2023",
				Genre:      "RPG",
				Platform:   "PC",
				Version:    "2.0",
				Language:   "English",
				FileSize:   75000000000, // 75GB
				MediaType:  models.MediaTypeGame,
				Confidence: 0.97,
			},
			expectedMatch:      false,
			expectedSimilarity: 0.7,
			name:               "expansion pack - not duplicate",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "The Witcher 3: Wild Hunt",
				Developer:  "CD Projekt RED",
				Publisher:  "CD Projekt",
				Year:       "2015",
				Genre:      "RPG",
				Platform:   "PC",
				Version:    "1.32",
				Language:   "English",
				FileSize:   50000000000, // 50GB
				MediaType:  models.MediaTypeGame,
				Confidence: 0.98,
			},
			expectedMatch:      false,
			expectedSimilarity: 0.3,
			name:               "different game same developer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isDuplicate, similarity := service.IsDuplicate(originalGame, tc.duplicate)
			assert.Equal(t, tc.expectedMatch, isDuplicate, "Duplicate detection result should match expected")
			assert.InDelta(t, tc.expectedSimilarity, similarity, 0.15, "Similarity should be within expected range")
		})
	}
}

func TestDuplicateDetectionService_AlgorithmAccuracy(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(db, logger, nil)

	t.Run("levenshtein distance", func(t *testing.T) {
		testCases := []struct {
			str1     string
			str2     string
			expected int
		}{
			{"kitten", "sitting", 3},
			{"saturday", "sunday", 3},
			{"", "abc", 3},
			{"abc", "", 3},
			{"same", "same", 0},
		}

		for _, tc := range testCases {
			result := service.CalculateLevenshteinDistance(tc.str1, tc.str2)
			assert.Equal(t, tc.expected, result, "Levenshtein distance should match expected")
		}
	})

	t.Run("jaro winkler similarity", func(t *testing.T) {
		testCases := []struct {
			str1     string
			str2     string
			expected float64
		}{
			{"MARTHA", "MARHTA", 0.961},
			{"DIXON", "DICKSONX", 0.767},
			{"same", "same", 1.0},
			{"", "", 1.0},
			{"abc", "xyz", 0.0},
		}

		for _, tc := range testCases {
			result := service.CalculateJaroWinklerSimilarity(tc.str1, tc.str2)
			assert.InDelta(t, tc.expected, result, 0.01, "Jaro-Winkler similarity should match expected")
		}
	})

	t.Run("soundex matching", func(t *testing.T) {
		testCases := []struct {
			str1     string
			str2     string
			expected bool
		}{
			{"Smith", "Smyth", true},
			{"Johnson", "Jonson", true},
			{"Robert", "Rupert", true},
			{"Brown", "Green", false},
			{"", "", true},
		}

		for _, tc := range testCases {
			result := service.SoundexMatch(tc.str1, tc.str2)
			assert.Equal(t, tc.expected, result, "Soundex matching should match expected")
		}
	})
}

func TestDuplicateDetectionService_Performance(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(db, logger, nil)

	// Create a large number of media items for performance testing
	mediaItems := make([]*models.MediaMetadata, 1000)
	for i := 0; i < 1000; i++ {
		mediaItems[i] = &models.MediaMetadata{
			Title:      fmt.Sprintf("Movie %d", i),
			Year:       "2023",
			Genre:      "Action",
			MediaType:  models.MediaTypeVideo,
			Confidence: 0.9,
		}
	}

	// Add some actual duplicates
	mediaItems[500] = &models.MediaMetadata{
		Title:      "Movie 100",
		Year:       "2023",
		Genre:      "Action",
		MediaType:  models.MediaTypeVideo,
		Confidence: 0.85,
	}

	t.Run("bulk duplicate detection", func(t *testing.T) {
		start := time.Now()

		duplicates := service.FindAllDuplicates(mediaItems)

		duration := time.Since(start)
		t.Logf("Bulk duplicate detection took: %v", duration)

		// Should find at least one duplicate pair
		assert.True(t, len(duplicates) > 0, "Should find duplicate pairs")
		assert.True(t, duration < 5*time.Second, "Detection should complete within 5 seconds")
	})
}

func BenchmarkDuplicateDetection(b *testing.B) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := NewDuplicateDetectionService(db, logger, nil)

	movie1 := &models.MediaMetadata{
		Title:      "The Matrix",
		Year:       "1999",
		Genre:      "Science Fiction",
		MediaType:  models.MediaTypeVideo,
		Confidence: 0.95,
	}

	movie2 := &models.MediaMetadata{
		Title:      "Matrix",
		Year:       "1999",
		Genre:      "Sci-Fi",
		MediaType:  models.MediaTypeVideo,
		Confidence: 0.92,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.IsDuplicate(movie1, movie2)
	}
}
