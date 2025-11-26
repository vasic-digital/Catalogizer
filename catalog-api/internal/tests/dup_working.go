package tests

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"catalogizer/internal/services"
	_ "github.com/mutecomm/go-sqlcipher"
)

func TestDuplicateDetectionService_BasicCreation(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	
	// Create cache service first
	cacheService := services.NewCacheService(db, logger)
	
	// Test service creation
	service := services.NewDuplicateDetectionService(db, logger, cacheService)
	assert.NotNil(t, service, "DuplicateDetectionService should be created successfully")
}

func TestDuplicateDetectionService_DetectDuplicates(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
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
		UserID:          1,
	}
	
	// This should not crash, though may not find duplicates without data
	groups, err := service.DetectDuplicates(nil, req)
	assert.NoError(t, err, "DetectDuplicates should not error")
	assert.NotNil(t, groups, "Should return groups array")
}