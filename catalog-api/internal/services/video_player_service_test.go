package services

import (
	"catalogizer/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewVideoPlayerService(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()

	service := NewVideoPlayerService(mockDB, mockLogger, nil, nil, nil, nil, nil)

	assert.NotNil(t, service)
}

func TestNewVideoPlayerService_WithNilDeps(t *testing.T) {
	mockLogger := zap.NewNop()
	service := NewVideoPlayerService(nil, mockLogger, nil, nil, nil, nil, nil)

	assert.NotNil(t, service)
}
