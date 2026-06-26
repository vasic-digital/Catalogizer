package services

import (
	"context"
	"testing"
	"time"

	"catalogizer/utils"
	"go.uber.org/zap"
)

func TestSMBDiscoveryService_DiscoverShares_UnreachableHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network-dependent test in short mode")
	}
	logger := zap.NewNop()
	service := NewSMBDiscoveryService(logger)

	// Test with a non-existent host to verify anti-bluff: DiscoverShares returns
	// an error (not fabricated "common shares") when the host is unreachable.
	// §11.4.6 — guessing share names that don't exist on the target is forbidden.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := service.DiscoverShares(ctx, "nonexistent.host", "testuser", "testpass", nil)

	if err == nil {
		t.Fatal("§11.4.6: DiscoverShares must return an error for an unreachable host, not fabricated share names")
	}
	t.Logf("§11.4.6 PASS: DiscoverShares returned honest error: %v", err)
}

func TestSMBDiscoveryService_TestConnection_InvalidHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network-dependent test in short mode")
	}
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
		Description: utils.StringPtr("Test description"),
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
