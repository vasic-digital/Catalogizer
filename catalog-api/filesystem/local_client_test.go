package filesystem

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalClient_Connect(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "local_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		BasePath: tempDir,
	}

	client := NewLocalClient(config)

	// Test connection
	err = client.Connect(context.Background())
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Client should be connected after successful Connect")
	}

	// Test disconnection
	err = client.Disconnect(context.Background())
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("Client should not be connected after Disconnect")
	}
}

func TestLocalClient_TestConnection(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "local_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		BasePath: tempDir,
	}

	client := NewLocalClient(config)
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = client.TestConnection(context.Background())
	if err != nil {
		t.Errorf("TestConnection failed: %v", err)
	}
}

func TestLocalClient_WriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "local_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		BasePath: tempDir,
	}

	client := NewLocalClient(config)
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Test writing a file
	testContent := "Hello, World!"
	testPath := "test.txt"

	err = client.WriteFile(context.Background(), testPath, bytes.NewReader([]byte(testContent)))
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Verify the file was written
	fullPath := filepath.Join(tempDir, testPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Errorf("Failed to read written file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", testContent, string(content))
	}
}

func TestLocalClient_ReadFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "local_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		BasePath: tempDir,
	}

	client := NewLocalClient(config)
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Create a test file
	testContent := "Hello, World!"
	testPath := "test.txt"
	fullPath := filepath.Join(tempDir, testPath)

	err = os.WriteFile(fullPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test reading the file
	reader, err := client.ReadFile(context.Background(), testPath)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	defer reader.Close()

	content := make([]byte, len(testContent))
	_, err = reader.Read(content)
	if err != nil {
		t.Errorf("Failed to read from reader: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", testContent, string(content))
	}
}

func TestLocalClient_GetFileInfo(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "local_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		BasePath: tempDir,
	}

	client := NewLocalClient(config)
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Create a test file
	testContent := "Hello, World!"
	testPath := "test.txt"
	fullPath := filepath.Join(tempDir, testPath)

	err = os.WriteFile(fullPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test getting file info
	info, err := client.GetFileInfo(context.Background(), testPath)
	if err != nil {
		t.Errorf("GetFileInfo failed: %v", err)
	}

	if info.Name != "test.txt" {
		t.Errorf("File name mismatch. Expected: test.txt, Got: %s", info.Name)
	}

	if info.Size != int64(len(testContent)) {
		t.Errorf("File size mismatch. Expected: %d, Got: %d", len(testContent), info.Size)
	}

	if info.IsDir {
		t.Error("File should not be marked as directory")
	}

	if info.Path != testPath {
		t.Errorf("File path mismatch. Expected: %s, Got: %s", testPath, info.Path)
	}
}

func TestLocalClient_ListDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "local_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		BasePath: tempDir,
	}

	client := NewLocalClient(config)
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Create test files and directories
	testFile := filepath.Join(tempDir, "test.txt")
	testDir := filepath.Join(tempDir, "testdir")

	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test listing directory
	files, err := client.ListDirectory(context.Background(), "")
	if err != nil {
		t.Errorf("ListDirectory failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 items, got %d", len(files))
	}

	// Check that both file and directory are listed
	foundFile := false
	foundDir := false
	for _, file := range files {
		if file.Name == "test.txt" && !file.IsDir {
			foundFile = true
		}
		if file.Name == "testdir" && file.IsDir {
			foundDir = true
		}
	}

	if !foundFile {
		t.Error("Test file not found in directory listing")
	}
	if !foundDir {
		t.Error("Test directory not found in directory listing")
	}
}

func TestLocalClient_FileExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "local_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		BasePath: tempDir,
	}

	client := NewLocalClient(config)
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Create a test file
	testPath := "test.txt"
	fullPath := filepath.Join(tempDir, testPath)

	err = os.WriteFile(fullPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test file exists
	exists, err := client.FileExists(context.Background(), testPath)
	if err != nil {
		t.Errorf("FileExists failed: %v", err)
	}
	if !exists {
		t.Error("File should exist")
	}

	// Test file doesn't exist
	exists, err = client.FileExists(context.Background(), "nonexistent.txt")
	if err != nil {
		t.Errorf("FileExists failed: %v", err)
	}
	if exists {
		t.Error("File should not exist")
	}
}

func TestLocalClient_GetProtocol(t *testing.T) {
	config := &LocalConfig{
		BasePath: "/tmp",
	}

	client := NewLocalClient(config)

	if client.GetProtocol() != "local" {
		t.Errorf("Expected protocol 'local', got '%s'", client.GetProtocol())
	}
}
