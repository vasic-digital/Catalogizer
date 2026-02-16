package services

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewVideoPlayerService(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()

	service := NewVideoPlayerService(mockDB, mockLogger, nil, nil, nil, nil, nil)

	assert.NotNil(t, service)
}

func TestNewVideoPlayerService_WithNilDeps(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewVideoPlayerService(nil, mockLogger, nil, nil, nil, nil, nil)

	assert.NotNil(t, service)
}
