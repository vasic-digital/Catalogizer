package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type SMBServiceTestSuite struct {
	suite.Suite
	service *SMBService
	logger  *zap.Logger
}

func (suite *SMBServiceTestSuite) SetupTest() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	suite.logger = logger

	// Initialize service with mock config
	config := &Config{
		SMB: SMBConfig{
			Sources:   []string{"smb://test-server/media"},
			Username:  "testuser",
			Password:  "testpass",
			Domain:    "TEST",
		},
	}

	suite.service = NewSMBService(config, logger)
}

func (suite *SMBServiceTestSuite) TestNewSMBService() {
	service := NewSMBService(nil, suite.logger)
	assert.NotNil(suite.T(), service)
	assert.NotNil(suite.T(), service.logger)
}

func (suite *SMBServiceTestSuite) TestSMBConnection() {
	// Test connection to non-existent server (should fail gracefully)
	err := suite.service.Connect("smb://nonexistent-server/share")
	assert.Error(suite.T(), err)
}

func (suite *SMBServiceTestSuite) TestListSMBDirectory() {
	// Test listing directory from non-existent server
	files, err := suite.service.ListDirectory("smb://nonexistent-server/share")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), files)
}

func (suite *SMBServiceTestSuite) TestSMBConfig() {
	assert.Equal(suite.T(), []string{"smb://test-server/media"}, suite.service.config.SMB.Sources)
	assert.Equal(suite.T(), "testuser", suite.service.config.SMB.Username)
	assert.Equal(suite.T(), "testpass", suite.service.config.SMB.Password)
	assert.Equal(suite.T(), "TEST", suite.service.config.SMB.Domain)
}

func (suite *SMBServiceTestSuite) TestSMBPathParsing() {
	tests := []struct {
		input    string
		expected SMBPath
	}{
		{
			"smb://server/share",
			SMBPath{Server: "server", Share: "share", Path: ""},
		},
		{
			"smb://server/share/folder",
			SMBPath{Server: "server", Share: "share", Path: "folder"},
		},
		{
			"smb://server/share/folder/file.mp4",
			SMBPath{Server: "server", Share: "share", Path: "folder/file.mp4"},
		},
	}

	for _, test := range tests {
		result := suite.service.ParseSMBPath(test.input)
		assert.Equal(suite.T(), test.expected.Server, result.Server)
		assert.Equal(suite.T(), test.expected.Share, result.Share)
		assert.Equal(suite.T(), test.expected.Path, result.Path)
	}
}

func (suite *SMBServiceTestSuite) TestSMBPathValidation() {
	validPaths := []string{
		"smb://server/share",
		"smb://server.domain.com/share",
		"smb://192.168.1.100/share/folder",
		"smb://server/share/path/to/file.mp4",
	}

	for _, path := range validPaths {
		assert.True(suite.T(), suite.service.IsValidSMBPath(path), "Path should be valid: %s", path)
	}

	invalidPaths := []string{
		"",
		"ftp://server/share",
		"smb://",
		"smb://server",
		"not-an-smb-path",
		"http://server/share",
	}

	for _, path := range invalidPaths {
		assert.False(suite.T(), suite.service.IsValidSMBPath(path), "Path should be invalid: %s", path)
	}
}

func (suite *SMBServiceTestSuite) TestSMBConnectionPooling() {
	// Test that service maintains connection state
	initialState := suite.service.IsConnected("smb://test-server/media")
	assert.False(suite.T(), initialState)

	// Attempt connection (will fail but shouldn't panic)
	err := suite.service.Connect("smb://test-server/media")
	assert.Error(suite.T(), err)

	// State should still be false
	currentState := suite.service.IsConnected("smb://test-server/media")
	assert.False(suite.T(), currentState)
}

func (suite *SMBServiceTestSuite) TestSMBErrorHandling() {
	// Test various error conditions
	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{"Empty path", "", true},
		{"Invalid scheme", "ftp://server/share", true},
		{"Missing server", "smb:///share", true},
		{"Missing share", "smb://server", true},
		{"Valid path", "smb://server/share", false},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if tc.expectError {
				assert.False(t, suite.service.IsValidSMBPath(tc.path))
			} else {
				assert.True(t, suite.service.IsValidSMBPath(tc.path))
			}
		})
	}
}

func (suite *SMBServiceTestSuite) TestSMBFileOperations() {
	// Test file operations on non-existent server
	err := suite.service.CopyFile("smb://server1/file1.txt", "smb://server2/file1.txt")
	assert.Error(suite.T(), err)

	exists, err := suite.service.FileExists("smb://server/file.txt")
	assert.Error(suite.T(), err)
	assert.False(suite.T(), exists)

	size, err := suite.service.GetFileSize("smb://server/file.txt")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), int64(0), size)
}

func (suite *SMBServiceTestSuite) TestSMBDirectoryOperations() {
	// Test directory operations
	err := suite.service.CreateDirectory("smb://server/newdir")
	assert.Error(suite.T(), err)

	err = suite.service.DeleteDirectory("smb://server/olddir")
	assert.Error(suite.T(), err)

	exists, err := suite.service.DirectoryExists("smb://server/dir")
	assert.Error(suite.T(), err)
	assert.False(suite.T(), exists)
}

func (suite *SMBServiceTestSuite) TestSMBConnectionTimeout() {
	// Test with very short timeout (should fail quickly)
	shortTimeoutConfig := &Config{
		SMB: SMBConfig{
			Sources:        []string{"smb://192.0.2.1/share"}, // TEST-NET-1 (non-routable)
			Username:       "test",
			Password:       "test",
			ConnectionTimeout: 1, // 1 second
		},
	}

	service := NewSMBService(shortTimeoutConfig, suite.logger)

	// This should fail quickly due to timeout
	err := service.Connect("smb://192.0.2.1/share")
	assert.Error(suite.T(), err)
}

func (suite *SMBServiceTestSuite) TestSMBSourceManagement() {
	// Test adding/removing SMB sources
	initialCount := len(suite.service.config.SMB.Sources)
	assert.Equal(suite.T(), 1, initialCount)

	// Test source validation
	validSources := []string{
		"smb://server1/share1",
		"smb://server2/share2",
		"smb://192.168.1.100/media",
	}

	for _, source := range validSources {
		assert.True(suite.T(), suite.service.IsValidSMBPath(source))
	}
}

func TestSMBServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SMBServiceTestSuite))
}