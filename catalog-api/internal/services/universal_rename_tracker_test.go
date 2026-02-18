package services

import (
	"catalogizer/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewUniversalRenameTracker(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()

	tracker := NewUniversalRenameTracker(mockDB, mockLogger)

	assert.NotNil(t, tracker)
}

func TestUniversalRenameTracker_RegisterProtocolHandler(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	tracker := NewUniversalRenameTracker(mockDB, mockLogger)

	// Default handlers should already be registered
	protocols := tracker.getSupportedProtocols()
	assert.NotEmpty(t, protocols)
	assert.GreaterOrEqual(t, len(protocols), 5)

	// Register a custom handler
	customHandler := NewLocalProtocolHandler(mockLogger)
	tracker.RegisterProtocolHandler("custom", customHandler)

	updatedProtocols := tracker.getSupportedProtocols()
	assert.Greater(t, len(updatedProtocols), len(protocols))
}

func TestUniversalRenameTracker_GetSupportedProtocols(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	tracker := NewUniversalRenameTracker(mockDB, mockLogger)

	protocols := tracker.getSupportedProtocols()
	assert.NotEmpty(t, protocols)

	// Check that default protocols are registered
	protocolSet := make(map[string]bool)
	for _, p := range protocols {
		protocolSet[p] = true
	}

	assert.True(t, protocolSet["local"])
	assert.True(t, protocolSet["smb"])
	assert.True(t, protocolSet["ftp"])
	assert.True(t, protocolSet["nfs"])
	assert.True(t, protocolSet["webdav"])
}

func TestUniversalRenameTracker_CreateFallbackKey(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	tracker := NewUniversalRenameTracker(mockDB, mockLogger)

	tests := []struct {
		name        string
		protocol    string
		storageRoot string
		fileHash    *string
		size        int64
		isDirectory bool
	}{
		{
			name:        "file with hash",
			protocol:    "local",
			storageRoot: "/media",
			fileHash:    strPtrHelper("abc123"),
			size:        1024,
			isDirectory: false,
		},
		{
			name:        "directory without hash",
			protocol:    "smb",
			storageRoot: "//server/share",
			fileHash:    nil,
			size:        0,
			isDirectory: true,
		},
		{
			name:        "file without hash",
			protocol:    "ftp",
			storageRoot: "ftp://server",
			fileHash:    nil,
			size:        2048,
			isDirectory: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tracker.createFallbackKey(tt.protocol, tt.storageRoot, tt.fileHash, tt.size, tt.isDirectory)
			assert.NotEmpty(t, key)
			assert.Contains(t, key, "fallback:")
			assert.Contains(t, key, tt.protocol)
		})
	}
}

func TestUniversalRenameTracker_CreateFallbackKey_Uniqueness(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	tracker := NewUniversalRenameTracker(mockDB, mockLogger)

	key1 := tracker.createFallbackKey("local", "/media", nil, 1024, false)
	key2 := tracker.createFallbackKey("smb", "/media", nil, 1024, false)
	key3 := tracker.createFallbackKey("local", "/media", nil, 2048, false)
	key4 := tracker.createFallbackKey("local", "/media", nil, 1024, true)

	assert.NotEqual(t, key1, key2)
	assert.NotEqual(t, key1, key3)
	assert.NotEqual(t, key1, key4)
}

func strPtrHelper(s string) *string {
	return &s
}
