package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"catalogizer/internal/services"
)

// TestRecommendationServiceConstruction verifies that the recommendation service
// can be properly instantiated with all required dependencies.
// Note: Full functional tests require mock implementations of FileRepositoryInterface.
func TestRecommendationServiceConstruction(t *testing.T) {
	// Setup database with schema
	db := SetupTestDB(t)
	defer db.Close()

	logger := zap.NewNop()

	// Create required service dependencies
	cacheService := services.NewCacheService(db, logger)
	translationService := services.NewTranslationService(logger)
	mediaRecognitionService := services.NewMediaRecognitionService(
		db, logger, cacheService, translationService,
		"", "", "", "", "", "", // Empty API URLs for testing
	)
	duplicateDetectionService := services.NewDuplicateDetectionService(db, logger, nil)

	// Verify services are created successfully
	assert.NotNil(t, cacheService, "CacheService should be created")
	assert.NotNil(t, translationService, "TranslationService should be created")
	assert.NotNil(t, mediaRecognitionService, "MediaRecognitionService should be created")
	assert.NotNil(t, duplicateDetectionService, "DuplicateDetectionService should be created")

	// Test recommendation service creation
	// Note: fileRepository is required for full functionality, but nil is accepted
	// for construction. Functional tests would need a mock repository.
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		nil, // FileRepository - would need mock for functional tests
		db,
	)

	assert.NotNil(t, recommendationService, "RecommendationService should be created")
}
