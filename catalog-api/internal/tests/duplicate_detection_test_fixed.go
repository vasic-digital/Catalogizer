package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"catalogizer/internal/services"
	_ "github.com/mattn/go-sqlite3"
)

func TestDuplicateDetectionService_Basic(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	service := services.NewDuplicateDetectionService(db, logger, nil)

	// Test basic service creation
	require.NotNil(t, service)
	require.NotNil(t, db)
	require.NotNil(t, logger)

	// Test duplicate detection request with empty database
	ctx := context.Background()
	req := &services.DuplicateDetectionRequest{
		UserID:        1,
		MinSimilarity:  0.8,
		DetectionMethods: []string{"title", "metadata"},
		BatchSize:      100,
	}

	// This should not fail even with empty database
	groups, err := service.DetectDuplicates(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, groups)
	// Should be empty since database is empty
	require.Equal(t, 0, len(groups))
}