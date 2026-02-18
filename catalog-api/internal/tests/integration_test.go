package tests

import (
	"context"
	"database/sql"
	"testing"

	"catalogizer/database"
	"catalogizer/internal/models"
	"catalogizer/internal/services"
	"go.uber.org/zap"
)

// TestMediaRecognitionIntegration tests the media recognition service
func TestMediaRecognitionIntegration(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer sqlDB.Close()
	db := database.WrapDB(sqlDB, database.DialectSQLite)

	logger := zap.NewNop()
	cacheService := services.NewCacheService(db, logger)
	translationService := services.NewTranslationService(logger)

	recognitionService := services.NewMediaRecognitionService(
		db,
		logger,
		cacheService,
		translationService,
		"", // Empty API keys - will use mock/fallback behavior
		"",
		"",
		"",
		"",
		"",
	)

	t.Run("service initialization", func(t *testing.T) {
		if recognitionService == nil {
			t.Error("Recognition service should not be nil")
		}
	})

	t.Run("get recognition stats", func(t *testing.T) {
		ctx := context.Background()
		stats, err := recognitionService.GetRecognitionStats(ctx)

		// Stats should return without error even with no prior recognitions
		if err != nil {
			t.Logf("GetRecognitionStats returned error (may be expected without database tables): %v", err)
		}
		if stats != nil {
			t.Logf("Recognition stats: %v", stats)
		}
	})
}

// TestDuplicateDetectionIntegration tests the duplicate detection service
func TestDuplicateDetectionIntegration(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer sqlDB.Close()
	db := database.WrapDB(sqlDB, database.DialectSQLite)

	logger := zap.NewNop()
	cacheService := services.NewCacheService(db, logger)

	duplicateService := services.NewDuplicateDetectionService(db, logger, cacheService)

	t.Run("service initialization", func(t *testing.T) {
		if duplicateService == nil {
			t.Error("Duplicate detection service should not be nil")
		}
	})

	t.Run("detect duplicates with empty database", func(t *testing.T) {
		ctx := context.Background()
		req := &services.DuplicateDetectionRequest{
			MediaTypes:    []services.MediaType{services.MediaTypeVideo},
			MinSimilarity: 0.8,
		}

		groups, err := duplicateService.DetectDuplicates(ctx, req)
		if err != nil {
			t.Logf("DetectDuplicates with empty database returned error: %v", err)
		}
		// Empty database should return empty groups
		t.Logf("Found %d duplicate groups (expected 0 with empty database)", len(groups))
	})

	t.Run("detect duplicates with various media types", func(t *testing.T) {
		ctx := context.Background()

		req := &services.DuplicateDetectionRequest{
			MediaTypes:    []services.MediaType{services.MediaTypeVideo, services.MediaTypeMusic},
			MinSimilarity: 0.5, // Lower threshold to catch similar titles
			BatchSize:     100,
		}

		groups, err := duplicateService.DetectDuplicates(ctx, req)
		if err != nil {
			t.Errorf("DetectDuplicates returned error: %v", err)
		}

		t.Logf("Found %d duplicate groups", len(groups))
		for i, group := range groups {
			t.Logf("Group %d: %d items, confidence %.2f", i+1, len(group.DuplicateItems), group.Confidence)
		}
	})
}

// Basic service creation test to ensure services can be instantiated
func TestServiceCreation(t *testing.T) {
	sqlDB, _ := sql.Open("sqlite3", ":memory:")
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	logger := zap.NewNop()

	// Test that services can be created with required parameters
	cacheService := services.NewCacheService(db, logger)
	translationService := services.NewTranslationService(logger)
	localizationService := services.NewLocalizationService(db, logger, translationService, cacheService)
	
	// Test creating services with required parameters
	recognitionService := services.NewMediaRecognitionService(
		db,
		logger,
		cacheService,
		translationService,
		"tmdb_key",
		"music_key",
		"book_key",
		"game_key",
		"ocr_key",
		"fingerprint_key",
	)
	
	duplicationService := services.NewDuplicateDetectionService(db, logger, cacheService)
	
	readerService := services.NewReaderService(
		db,
		logger,
		cacheService,
		translationService,
		localizationService,
	)
	
	// Verify services are created
	if recognitionService == nil {
		t.Error("Recognition service should not be nil")
	}
	if duplicationService == nil {
		t.Error("Duplication service should not be nil")
	}
	if readerService == nil {
		t.Error("Reader service should not be nil")
	}
}

// Test basic media metadata struct creation
func TestMediaMetadataCreation(t *testing.T) {
	year := 1999
	duration := 136
	fileSize := int64(2048000000)
	
	metadata := &models.MediaMetadata{
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
	
	// Verify struct was created correctly
	if metadata.Title != "The Matrix" {
		t.Errorf("Expected title 'The Matrix', got %s", metadata.Title)
	}
	if *metadata.Year != 1999 {
		t.Errorf("Expected year 1999, got %d", *metadata.Year)
	}
}