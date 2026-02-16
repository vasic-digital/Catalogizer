package services

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewPlaybackPositionService(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()

	service := NewPlaybackPositionService(mockDB, mockLogger)

	assert.NotNil(t, service)
}

func TestNewPlaybackPositionService_WithNilDB(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewPlaybackPositionService(nil, mockLogger)

	assert.NotNil(t, service)
}
