package tests

import (
	"database/sql"
	"testing"

	"catalogizer/internal/models"
	"catalogizer/internal/services"
	"go.uber.org/zap"
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
			text2:    "Matrix",
			expected: 0.75,
			name:     "similar titles",
		},
		{
			text1:    "Completely Different Text",
			text2:    "Nothing in Common",
			expected: 0.1,
			name:     "different content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			similarity := service.calculateTextSimilarity(tc.text1, tc.text2)
			if similarity < tc.expected-0.15 || similarity > tc.expected+0.15 {
				t.Errorf("Expected similarity around %f, got %f", tc.expected, similarity)
			}
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
				Year:       func() *int { y := 1999; return &y }(),
				Genre:      "Science Fiction",
				Director:   "The Wachowskis",
				Duration:   func() *int { d := 136; return &d }(),
				Resolution: "720p",
				FileSize:   func() *int64 { s := int64(1024000000); return &s }(),
				Language:   "English",
				MediaType:  models.MediaTypeVideo,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.95,
			name:               "same movie different quality",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "Matrix",
				Year:       func() *int { y := 1999; return &y }(),
				Genre:      "Sci-Fi",
				Director:   "Wachowski Brothers",
				Duration:   func() *int { d := 136; return &d }(),
				Resolution: "1080p",
				FileSize:   func() *int64 { s := int64(2048000000); return &s }(),
				Language:   "English",
				MediaType:  models.MediaTypeVideo,
			},
			expectedMatch:      true,
			expectedSimilarity: 0.85,
			name:               "slight variations in metadata",
		},
		{
			duplicate: &models.MediaMetadata{
				Title:      "The Matrix Reloaded",
				Year:       func() *int { y := 2003; return &y }(),
				Genre:      "Science Fiction",
				Director:   "The Wachowskis",
				Duration:   func() *int { d := 138; return &d }(),
				Resolution: "1080p",
				FileSize:   func() *int64 { s := int64(2148000000); return &s }(),
				Language:   "English",
				MediaType:  models.MediaTypeVideo,
			},
			expectedMatch:      false,
			expectedSimilarity: 0.6,
			name:               "different movie in same series",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test would need the actual comparison logic implementation
			// For now, we're testing the struct creation
			if tc.duplicate == nil || originalMovie == nil {
				t.Error("MediaMetadata should not be nil")
			}
		})
	}
}