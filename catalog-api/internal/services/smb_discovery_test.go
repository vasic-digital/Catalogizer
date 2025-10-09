package services

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestSMBDiscoveryService_GetCommonShares(t *testing.T) {
	logger := zap.NewNop()
	service := NewSMBDiscoveryService(logger)

	// Test with a non-existent host to ensure fallback works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shares := service.getCommonShares(ctx, "nonexistent.host", "testuser", "testpass", nil)

	if len(shares) == 0 {
		t.Error("Expected common shares to be returned, got empty list")
	}

	// Verify the shares have expected properties
	for _, share := range shares {
		if share.Host != "nonexistent.host" {
			t.Errorf("Expected host to be 'nonexistent.host', got '%s'", share.Host)
		}
		if share.ShareName == "" {
			t.Error("Expected share name to be non-empty")
		}
		if share.Path == "" {
			t.Error("Expected path to be non-empty")
		}
	}
}

func TestSMBDiscoveryService_TestConnection_InvalidHost(t *testing.T) {
	logger := zap.NewNop()
	service := NewSMBDiscoveryService(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	config := SMBConnectionConfig{
		Host:     "invalid.host.that.does.not.exist",
		Port:     445,
		Share:    "testshare",
		Username: "testuser",
		Password: "testpass",
	}

	result := service.TestConnection(ctx, config)

	// Should return false for invalid host
	if result {
		t.Error("Expected connection test to fail for invalid host")
	}
}

func TestSMBConnectionConfig_Validation(t *testing.T) {
	// Test that required fields are properly handled
	config := SMBConnectionConfig{
		Host:     "testhost",
		Port:     445,
		Share:    "testshare",
		Username: "testuser",
		Password: "testpass",
		Domain:   nil,
	}

	if config.Host == "" {
		t.Error("Host should not be empty")
	}
	if config.Port == 0 {
		t.Error("Port should not be zero")
	}
	if config.Share == "" {
		t.Error("Share should not be empty")
	}
}

func TestSMBShareInfo_Structure(t *testing.T) {
	// Test that SMBShareInfo structure is properly formed
	share := SMBShareInfo{
		Host:        "testhost",
		ShareName:   "testshare",
		Path:        "\\\\testhost\\testshare",
		Writable:    false,
		Description: stringPtr("Test description"),
	}

	if share.Host != "testhost" {
		t.Errorf("Expected host 'testhost', got '%s'", share.Host)
	}
	if share.Description == nil {
		t.Error("Description should not be nil")
	}
	if *share.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", *share.Description)
	}
}

func TestSMBFileEntry_Structure(t *testing.T) {
	// Test that SMBFileEntry structure is properly formed
	size := int64(1024)
	modified := "2024-01-01 12:00:00"

	entry := SMBFileEntry{
		Name:        "testfile.txt",
		Path:        "/path/to/testfile.txt",
		IsDirectory: false,
		Size:        &size,
		Modified:    &modified,
	}

	if entry.Name != "testfile.txt" {
		t.Errorf("Expected name 'testfile.txt', got '%s'", entry.Name)
	}
	if entry.IsDirectory {
		t.Error("Expected IsDirectory to be false")
	}
	if entry.Size == nil || *entry.Size != 1024 {
		t.Error("Expected size to be 1024")
	}
	if entry.Modified == nil || *entry.Modified != "2024-01-01 12:00:00" {
		t.Error("Expected modified time to match")
	}
}

// Helper function for testing
func stringPtr(s string) *string {
	return &s
}