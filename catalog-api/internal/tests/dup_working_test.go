package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"catalogizer/internal/services"
)

func TestDuplicateDetectionService_BasicCreation(t *testing.T) {
	db := SetupTestDB(t)
	logger := zap.NewNop()

	// Create cache service first
	cacheService := services.NewCacheService(db, logger)

	// Test service creation
	service := services.NewDuplicateDetectionService(db, logger, cacheService)
	assert.NotNil(t, service, "DuplicateDetectionService should be created successfully")
}

func TestDuplicateDetectionService_DetectDuplicates(t *testing.T) {
	db := SetupTestDB(t)
	logger := zap.NewNop()

	// Create cache service first
	cacheService := services.NewCacheService(db, logger)

	service := services.NewDuplicateDetectionService(db, logger, cacheService)

	// Test duplicate detection request
	req := &services.DuplicateDetectionRequest{
		MediaTypes:       []services.MediaType{services.MediaTypeVideo},
		MinSimilarity:    0.8,
		DetectionMethods: []string{"hash", "metadata"},
		IncludeExisting:  false,
		BatchSize:        100,
		UserID:           1,
	}

	// This should not crash, though may not find duplicates without data
	_, err := service.DetectDuplicates(nil, req)
	assert.NoError(t, err, "DetectDuplicates should not error")
	// It's acceptable to get nil groups when no duplicates are found
	// The service may return nil or empty slice, both are valid
}
