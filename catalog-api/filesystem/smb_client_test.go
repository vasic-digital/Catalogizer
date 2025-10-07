package filesystem

import (
	"testing"
)

func TestSmbClient_GetProtocol(t *testing.T) {
	config := &SmbConfig{
		Host:     "localhost",
		Port:     445,
		Share:    "test",
		Username: "user",
		Password: "pass",
		Domain:   "WORKGROUP",
	}

	client := NewSmbClient(config)

	if client.GetProtocol() != "smb" {
		t.Errorf("Expected protocol 'smb', got '%s'", client.GetProtocol())
	}
}

func TestSmbClient_IsConnected(t *testing.T) {
	config := &SmbConfig{
		Host:     "localhost",
		Port:     445,
		Share:    "test",
		Username: "user",
		Password: "pass",
		Domain:   "WORKGROUP",
	}

	client := NewSmbClient(config)

	// Should not be connected initially
	if client.IsConnected() {
		t.Error("Client should not be connected initially")
	}

	// Test connection to non-existent server (should fail gracefully)
	// Note: This test assumes no SMB server is running on localhost
	err := client.Connect(nil)
	if err == nil {
		t.Error("Connect should fail for non-existent SMB server")
		client.Disconnect(nil) // Clean up if somehow connected
	}

	if client.IsConnected() {
		t.Error("Client should not be connected after failed connection")
	}
}

func TestSmbClient_TestConnection(t *testing.T) {
	config := &SmbConfig{
		Host:     "localhost",
		Port:     445,
		Share:    "test",
		Username: "user",
		Password: "pass",
		Domain:   "WORKGROUP",
	}

	client := NewSmbClient(config)

	// Test connection when not connected
	err := client.TestConnection(nil)
	if err == nil {
		t.Error("TestConnection should fail when not connected")
	}
}