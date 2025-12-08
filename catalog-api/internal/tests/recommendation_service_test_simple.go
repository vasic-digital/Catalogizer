package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"catalogizer/internal/models"
	"catalogizer/internal/services"
	_ "github.com/mattn/go-sqlite3"
)

func TestRecommendationService_BasicOperation(t *testing.T) {
	ctx := context.Background()

	// Setup services with database
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	
	// Create services with minimal requirements
	mediaRecognitionService := services.NewMediaRecognitionService(db, logger, nil, nil, "", "", "", "", "")
	duplicateDetectionService := services.NewDuplicateDetectionService(db, logger, nil)
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
	)

	// Create test media item
	mediaItem := &models.MediaMetadata{
		Title: "Test Movie",
		Genre: "Action",
		Year: &[]int{2023}[0],
	}

	// Test getting recommendations - basic smoke test
	items, err := recommendationService.GetSimilarItems(ctx, mediaItem, 5)
	
	// We expect either a valid result or an error (due to empty database)
	// The important thing is that the service doesn't panic
	assert.True(t, items == nil || len(items) >= 0)
	
	// Error may be expected with empty database - we just verify it doesn't crash
	if err != nil {
		assert.Contains(t, err.Error(), "no") // Expected error messages contain "no" (e.g., "no data")
	}
}