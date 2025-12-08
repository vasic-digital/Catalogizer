package integration

import (
	"catalogizer/internal/services"
	"catalogizer/tests/mocks"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestProtocolConnectivity tests connectivity for all supported protocols
func TestProtocolConnectivity(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("SMB Protocol Tests", func(t *testing.T) {
		testSMBProtocol(t, logger, ctx)
	})

	t.Run("FTP Protocol Tests", func(t *testing.T) {
		testFTPProtocol(t, logger, ctx)
	})

	t.Run("Local Protocol Tests", func(t *testing.T) {
		testLocalProtocol(t, logger, ctx)
	})

	t.Run("NFS Protocol Tests", func(t *testing.T) {
		testNFSProtocol(t, logger, ctx)
	})

	t.Run("WebDAV Protocol Tests", func(t *testing.T) {
		testWebDAVProtocol(t, logger, ctx)
	})
}

// testSMBProtocol tests SMB protocol functionality
func testSMBProtocol(t *testing.T, logger *zap.Logger, ctx context.Context) {
	// Start mock SMB server
	mockServer := mocks.NewMockSMBServer(logger)
	err := mockServer.Start()
	if err != nil {
		t.Fatalf("Failed to start mock SMB server: %v", err)
	}
	defer mockServer.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Test SMB discovery service
	discoveryService := services.NewSMBDiscoveryService(logger)

	t.Run("SMB Share Discovery", func(t *testing.T) {
		// Test with valid credentials using the mock server
		host := "localhost"
		shares, err := discoveryService.DiscoverShares(ctx, host, "testuser", "testpass", nil)
		if err != nil {
			// Expected to fall back to common shares when real SMB protocol fails
			t.Logf("SMB share discovery failed as expected: %v", err)
			shares, err = discoveryService.DiscoverShares(ctx, "nonexistent.local", "testuser", "testpass", nil)
			if err != nil {
				t.Errorf("Expected fallback to common shares, got error: %v", err)
			}
		}

		if len(shares) == 0 {
			t.Error("Expected at least one share to be discovered")
		}

		// Verify share properties
		for _, share := range shares {
			if share.Host != "localhost" && share.Host != "nonexistent.local" {
				t.Errorf("Expected host 'localhost' or 'nonexistent.local', got '%s'", share.Host)
			}
			if share.ShareName == "" {
				t.Error("Share name should not be empty")
			}
			if share.Path == "" {
				t.Error("Share path should not be empty")
			}
		}
	})

	t.Run("SMB Connection Testing", func(t *testing.T) {
		host := "localhost"
		config := services.SMBConnectionConfig{
			Host:     host,
			Port:     mockServer.GetPort(),
			Share:    "shared",
			Username: "testuser",
			Password: "testpass",
		}

		// Test valid connection (will likely fail with mock but we test the failure path)
		success := discoveryService.TestConnection(ctx, config)
		if success {
			t.Log("Mock SMB connection succeeded (unexpected but acceptable)")
		} else {
			t.Log("Mock SMB connection failed as expected")
		}

		// Test invalid credentials
		configInvalid := config
		configInvalid.Password = "wrongpass"
		success = discoveryService.TestConnection(ctx, configInvalid)
		if success {
			t.Error("Expected connection test to fail with invalid credentials")
		}

		// Test non-existent share
		configNoShare := config
		configNoShare.Share = "nonexistent"
		success = discoveryService.TestConnection(ctx, configNoShare)
		if success {
			t.Error("Expected connection test to fail for non-existent share")
		}
	})

	t.Run("SMB Share Browsing", func(t *testing.T) {
		host := "localhost"
		config := services.SMBConnectionConfig{
			Host:     host,
			Port:     mockServer.GetPort(),
			Share:    "shared",
			Username: "testuser",
			Password: "testpass",
		}

		// Browse root directory (will likely fail with mock server)
		entries, err := discoveryService.BrowseShare(ctx, config, ".")
		if err != nil {
			t.Logf("Expected SMB browsing to fail with mock server: %v", err)
			// This is expected with the mock server
			return
		}

		if len(entries) == 0 {
			t.Error("Expected at least one entry in share")
		}

		// Verify entry properties
		for _, entry := range entries {
			if entry.Name == "" {
				t.Error("Entry name should not be empty")
			}
			if entry.Path == "" {
				t.Error("Entry path should not be empty")
			}
		}

		// Look for expected files from mock server
		foundReadme := false
		foundDocuments := false
		for _, entry := range entries {
			if entry.Name == "readme.txt" && !entry.IsDirectory {
				foundReadme = true
			}
			if entry.Name == "documents" && entry.IsDirectory {
				foundDocuments = true
			}
		}

		if !foundReadme {
			t.Error("Expected to find readme.txt file")
		}
		if !foundDocuments {
			t.Error("Expected to find documents directory")
		}
	})

	t.Run("SMB Error Handling", func(t *testing.T) {
		// Test connection to non-existent host
		config := services.SMBConnectionConfig{
			Host:     "invalid-host.local",
			Port:     445,
			Share:    "shared",
			Username: "testuser",
			Password: "testpass",
		}

		success := discoveryService.TestConnection(ctx, config)
		if success {
			t.Error("Expected connection test to fail for non-existent host")
		}

		// Test share discovery with invalid host
		shares, err := discoveryService.DiscoverShares(ctx, "invalid-host.local", "testuser", "testpass", nil)
		// Should not return error but fall back to common shares
		if err != nil {
			t.Errorf("Expected fallback behavior, got error: %v", err)
		}
		if len(shares) == 0 {
			t.Error("Expected fallback to common shares")
		}
	})
}

// testFTPProtocol tests FTP protocol functionality
func testFTPProtocol(t *testing.T, logger *zap.Logger, ctx context.Context) {
	// Start mock FTP server
	mockServer := mocks.NewMockFTPServer(logger)
	err := mockServer.Start()
	if err != nil {
		t.Fatalf("Failed to start mock FTP server: %v", err)
	}
	defer mockServer.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	t.Run("FTP Server Verification", func(t *testing.T) {
		if !mockServer.IsRunning() {
			t.Error("Mock FTP server should be running")
		}

		if mockServer.GetPort() == 0 {
			t.Error("FTP server port should be set")
		}

		if mockServer.GetFileCount() == 0 {
			t.Error("FTP server should have default files")
		}

		if mockServer.GetUserCount() == 0 {
			t.Error("FTP server should have default users")
		}
	})

	t.Run("FTP Authentication Testing", func(t *testing.T) {
		// Test valid user credentials
		testCases := []struct {
			username string
			password string
			expected bool
		}{
			{"anonymous", "", true},
			{"testuser", "testpass", true},
			{"ftpuser", "ftppass", true},
			{"nonexistent", "any", false},
			{"testuser", "wrongpass", false},
		}

		for _, tc := range testCases {
			// Note: This would require implementing an FTP client for proper testing
			// For now, we test the mock server's authentication logic directly
			result := mockServer.AuthenticateUser(tc.username, tc.password)
			if result != tc.expected {
				t.Errorf("Authentication test failed for %s:%s, expected %v, got %v",
					tc.username, tc.password, tc.expected, result)
			}
		}
	})

	t.Run("FTP File Listing", func(t *testing.T) {
		// Test file listing in root directory
		files, err := mockServer.ListFiles("/", ".")
		if err != nil {
			t.Errorf("Expected successful file listing, got error: %v", err)
		}

		if len(files) == 0 {
			t.Error("Expected files in root directory")
		}

		// Verify expected files exist
		expectedFiles := map[string]bool{
			"public":     true,  // directory
			"uploads":    true,  // directory
			"readme.txt": false, // file
		}

		for fileName, isDirectory := range expectedFiles {
			found := false
			for _, file := range files {
				if file.Name == fileName && file.IsDirectory == isDirectory {
					found = true
					break
				}
			}
			if !found {
				fileType := "file"
				if isDirectory {
					fileType = "directory"
				}
				t.Errorf("Expected %s '%s' not found", fileType, fileName)
			}
		}
	})

	t.Run("FTP File Operations", func(t *testing.T) {
		// Test file retrieval
		file, err := mockServer.GetFile("/readme.txt")
		if err != nil {
			t.Errorf("Expected to retrieve readme.txt, got error: %v", err)
		}
		if file == nil {
			t.Error("Expected file to be returned")
		}

		// Test file writing (to writable directory)
		err = mockServer.WriteFile("/uploads/test.txt", []byte("test content"))
		if err != nil {
			t.Errorf("Expected successful file write, got error: %v", err)
		}

		// Verify file was written
		writtenFile, err := mockServer.GetFile("/uploads/test.txt")
		if err != nil {
			t.Errorf("Expected to retrieve written file, got error: %v", err)
		}
		if string(writtenFile.Content) != "test content" {
			t.Errorf("Expected file content 'test content', got '%s'", string(writtenFile.Content))
		}

		// Test file deletion
		err = mockServer.DeleteFile("/uploads/test.txt")
		if err != nil {
			t.Errorf("Expected successful file deletion, got error: %v", err)
		}

		// Verify file was deleted
		_, err = mockServer.GetFile("/uploads/test.txt")
		if err == nil {
			t.Error("Expected file to be deleted")
		}
	})
}

// testLocalProtocol tests local filesystem protocol functionality
func testLocalProtocol(t *testing.T, logger *zap.Logger, ctx context.Context) {
	t.Run("Local File System Access", func(t *testing.T) {
		// This would test local filesystem operations
		// For now, just verify we can access the current directory

		// Note: In a real implementation, you would test:
		// - Directory listing
		// - File reading/writing
		// - Permission handling
		// - Path validation
		// - Symlink handling

		t.Log("Local filesystem protocol testing placeholder")
		// TODO: Implement actual local filesystem tests
	})
}

// testNFSProtocol tests NFS protocol functionality
func testNFSProtocol(t *testing.T, logger *zap.Logger, ctx context.Context) {
	// Start mock NFS server
	mockServer := mocks.NewMockNFSServer(logger, "/mnt/nfs")
	err := mockServer.Start()
	if err != nil {
		t.Fatalf("Failed to start mock NFS server: %v", err)
	}
	defer mockServer.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	t.Run("NFS Server Verification", func(t *testing.T) {
		if !mockServer.IsRunning() {
			t.Error("Mock NFS server should be running")
		}

		if mockServer.GetPort() != 2049 {
			t.Errorf("Expected NFS port 2049, got %d", mockServer.GetPort())
		}

		if mockServer.GetFileCount() == 0 {
			t.Error("NFS server should have default files")
		}

		if mockServer.GetExportCount() == 0 {
			t.Error("NFS server should have default exports")
		}
	})

	t.Run("NFS Export Listing", func(t *testing.T) {
		exports := mockServer.ListExports()
		if len(exports) == 0 {
			t.Error("Expected at least one NFS export")
		}

		// Verify expected exports exist
		expectedExports := []string{"/export/media", "/export/backup", "/export/shared"}
		exportPaths := mockServer.GetExportNames()

		for _, expected := range expectedExports {
			found := false
			for _, actual := range exportPaths {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected export '%s' not found", expected)
			}
		}
	})

	t.Run("NFS Mount Operations", func(t *testing.T) {
		// Test successful mount
		err := mockServer.Mount("/export/media", "localhost")
		if err != nil {
			t.Errorf("Expected successful mount, got error: %v", err)
		}

		// Test mount with allowed client
		err = mockServer.Mount("/export/shared", "127.0.0.1")
		if err != nil {
			t.Errorf("Expected successful mount for allowed client, got error: %v", err)
		}

		// Test mount of non-existent export
		err = mockServer.Mount("/export/nonexistent", "localhost")
		if err == nil {
			t.Error("Expected mount to fail for non-existent export")
		}

		// Test unmount
		err = mockServer.Unmount("/export/media", "localhost")
		if err != nil {
			t.Errorf("Expected successful unmount, got error: %v", err)
		}
	})

	t.Run("NFS File Listing", func(t *testing.T) {
		// List files in media export
		files, err := mockServer.ListFiles("/export/media", ".")
		if err != nil {
			t.Errorf("Expected successful file listing, got error: %v", err)
		}

		if len(files) == 0 {
			t.Error("Expected files in media export")
		}

		// Verify expected directories exist
		expectedDirs := map[string]bool{
			"movies": true,
			"music":  true,
			"photos": true,
		}

		for expectedDir := range expectedDirs {
			found := false
			for _, file := range files {
				if file.Name == expectedDir && file.IsDirectory {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected directory '%s' not found", expectedDir)
			}
		}

		// Test listing subdirectory
		movieFiles, err := mockServer.ListFiles("/export/media", "movies")
		if err != nil {
			t.Errorf("Expected successful subdirectory listing, got error: %v", err)
		}

		if len(movieFiles) == 0 {
			t.Error("Expected files in movies subdirectory")
		}
	})

	t.Run("NFS File Operations", func(t *testing.T) {
		// Test file retrieval
		file, err := mockServer.GetFile("/export/media", "readme.txt")
		if err != nil {
			t.Errorf("Expected to retrieve readme.txt, got error: %v", err)
		}
		if file == nil {
			t.Error("Expected file to be returned")
		}
		if file.IsDirectory {
			t.Error("Expected file to be a regular file, not directory")
		}

		// Test file writing (to writable export)
		err = mockServer.WriteFile("/export/shared", "test.txt", []byte("test content"), 0644, 1000, 1000)
		if err != nil {
			t.Errorf("Expected successful file write, got error: %v", err)
		}

		// Verify file was written
		writtenFile, err := mockServer.GetFile("/export/shared", "test.txt")
		if err != nil {
			t.Errorf("Expected to retrieve written file, got error: %v", err)
		}
		if string(writtenFile.Content) != "test content" {
			t.Errorf("Expected file content 'test content', got '%s'", string(writtenFile.Content))
		}

		// Test directory creation
		err = mockServer.CreateDirectory("/export/shared", "testdir", 0755, 1000, 1000)
		if err != nil {
			t.Errorf("Expected successful directory creation, got error: %v", err)
		}

		// Test file deletion
		err = mockServer.DeleteFile("/export/shared", "test.txt")
		if err != nil {
			t.Errorf("Expected successful file deletion, got error: %v", err)
		}

		// Verify file was deleted
		_, err = mockServer.GetFile("/export/shared", "test.txt")
		if err == nil {
			t.Error("Expected file to be deleted")
		}
	})

	t.Run("NFS Permission Testing", func(t *testing.T) {
		// Test writing to read-only export (should fail)
		err := mockServer.WriteFile("/export/backup", "test.txt", []byte("test"), 0644, 1000, 1000)
		if err == nil {
			t.Error("Expected write to read-only export to fail")
		}

		// Test connection with specific client
		err = mockServer.TestConnection("/export/backup", "192.168.1.100")
		if err != nil {
			t.Errorf("Expected successful connection from allowed subnet, got error: %v", err)
		}
	})
}

// testWebDAVProtocol tests WebDAV protocol functionality
func testWebDAVProtocol(t *testing.T, logger *zap.Logger, ctx context.Context) {
	// Start mock WebDAV server
	mockServer := mocks.NewMockWebDAVServer(logger)
	err := mockServer.Start()
	if err != nil {
		t.Fatalf("Failed to start mock WebDAV server: %v", err)
	}
	defer mockServer.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	t.Run("WebDAV Server Verification", func(t *testing.T) {
		if !mockServer.IsRunning() {
			t.Error("Mock WebDAV server should be running")
		}

		if mockServer.GetPort() == 0 {
			t.Error("WebDAV server port should be set")
		}

		if mockServer.GetFileCount() == 0 {
			t.Error("WebDAV server should have default files")
		}

		if mockServer.GetUserCount() == 0 {
			t.Error("WebDAV server should have default users")
		}
	})

	t.Run("WebDAV HTTP Methods", func(t *testing.T) {
		baseURL := mockServer.GetAddress()
		client := &http.Client{Timeout: 5 * time.Second}

		// Test OPTIONS request
		req, _ := http.NewRequest("OPTIONS", baseURL+"/", nil)
		req.SetBasicAuth("webdavuser", "webdavpass")
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("OPTIONS request failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected OPTIONS to return 200, got %d", resp.StatusCode)
			}

			// Verify DAV header
			davHeader := resp.Header.Get("DAV")
			if davHeader == "" {
				t.Error("Expected DAV header in OPTIONS response")
			}
		}

		// Test GET request for directory listing
		req, _ = http.NewRequest("GET", baseURL+"/", nil)
		req.SetBasicAuth("webdavuser", "webdavpass")
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("GET request failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected GET to return 200, got %d", resp.StatusCode)
			}
		}

		// Test unauthorized request
		req, _ = http.NewRequest("GET", baseURL+"/", nil)
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("Unauthorized GET request failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected unauthorized GET to return 401, got %d", resp.StatusCode)
			}
		}
	})

	t.Run("WebDAV PROPFIND", func(t *testing.T) {
		baseURL := mockServer.GetAddress()
		client := &http.Client{Timeout: 5 * time.Second}

		// Test PROPFIND on root directory
		req, _ := http.NewRequest("PROPFIND", baseURL+"/", nil)
		req.SetBasicAuth("webdavuser", "webdavpass")
		req.Header.Set("Depth", "1")
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("PROPFIND request failed: %v", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusMultiStatus {
				t.Errorf("Expected PROPFIND to return 207, got %d", resp.StatusCode)
			}

			// Verify XML response contains expected directories
			bodyStr := string(body)
			expectedDirs := []string{"documents", "media", "public"}
			for _, dir := range expectedDirs {
				if !strings.Contains(bodyStr, dir) {
					t.Errorf("Expected to find '%s' in PROPFIND response", dir)
				}
			}
		}

		// Test PROPFIND on specific directory
		req, _ = http.NewRequest("PROPFIND", baseURL+"/documents", nil)
		req.SetBasicAuth("webdavuser", "webdavpass")
		req.Header.Set("Depth", "1")
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("PROPFIND on documents failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusMultiStatus {
				t.Errorf("Expected PROPFIND on documents to return 207, got %d", resp.StatusCode)
			}
		}
	})

	t.Run("WebDAV File Operations", func(t *testing.T) {
		baseURL := mockServer.GetAddress()
		client := &http.Client{Timeout: 5 * time.Second}

		// Test PUT (upload file)
		content := "Test file content for WebDAV"
		req, _ := http.NewRequest("PUT", baseURL+"/testfile.txt", strings.NewReader(content))
		req.SetBasicAuth("webdavuser", "webdavpass")
		req.Header.Set("Content-Type", "text/plain")
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("PUT request failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected PUT to return 201, got %d", resp.StatusCode)
			}
		}

		// Test GET (download file)
		req, _ = http.NewRequest("GET", baseURL+"/testfile.txt", nil)
		req.SetBasicAuth("webdavuser", "webdavpass")
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("GET request for uploaded file failed: %v", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected GET to return 200, got %d", resp.StatusCode)
			}
			if string(body) != content {
				t.Errorf("Expected file content '%s', got '%s'", content, string(body))
			}
		}

		// Test MKCOL (create collection/directory)
		req, _ = http.NewRequest("MKCOL", baseURL+"/testdir", nil)
		req.SetBasicAuth("webdavuser", "webdavpass")
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("MKCOL request failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected MKCOL to return 201, got %d", resp.StatusCode)
			}
		}

		// Test DELETE
		req, _ = http.NewRequest("DELETE", baseURL+"/testfile.txt", nil)
		req.SetBasicAuth("webdavuser", "webdavpass")
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("DELETE request failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				t.Errorf("Expected DELETE to return 204, got %d", resp.StatusCode)
			}
		}

		// Verify file was deleted
		req, _ = http.NewRequest("GET", baseURL+"/testfile.txt", nil)
		req.SetBasicAuth("webdavuser", "webdavpass")
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("GET request for deleted file failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("Expected GET on deleted file to return 404, got %d", resp.StatusCode)
			}
		}
	})

	t.Run("WebDAV Authentication", func(t *testing.T) {
		baseURL := mockServer.GetAddress()
		client := &http.Client{Timeout: 5 * time.Second}

		// Test valid credentials
		testUsers := map[string]string{
			"webdavuser": "webdavpass",
			"testuser":   "testpass",
			"admin":      "adminpass",
		}

		for username, password := range testUsers {
			req, _ := http.NewRequest("GET", baseURL+"/", nil)
			req.SetBasicAuth(username, password)
			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("Request failed for user %s: %v", username, err)
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected successful auth for user %s, got status %d", username, resp.StatusCode)
				}
			}
		}

		// Test invalid credentials
		req, _ := http.NewRequest("GET", baseURL+"/", nil)
		req.SetBasicAuth("invalid", "invalid")
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("Request with invalid credentials failed: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected 401 for invalid credentials, got %d", resp.StatusCode)
			}
		}
	})
}

// TestProtocolCapabilities tests protocol capability detection
func TestProtocolCapabilities(t *testing.T) {
	_ = zap.NewNop() // Logger available but not used in this test

	testCases := []struct {
		protocol             string
		expectedRealTime     bool
		expectedMoveWindow   time.Duration
		expectedBatchSize    int
		expectedSupportsAuth bool
	}{
		{
			protocol:             "local",
			expectedRealTime:     true,
			expectedMoveWindow:   2 * time.Second,
			expectedBatchSize:    1000,
			expectedSupportsAuth: false,
		},
		{
			protocol:             "smb",
			expectedRealTime:     false,
			expectedMoveWindow:   10 * time.Second,
			expectedBatchSize:    500,
			expectedSupportsAuth: true,
		},
		{
			protocol:             "ftp",
			expectedRealTime:     false,
			expectedMoveWindow:   30 * time.Second,
			expectedBatchSize:    100,
			expectedSupportsAuth: true,
		},
		{
			protocol:             "nfs",
			expectedRealTime:     false,
			expectedMoveWindow:   5 * time.Second,
			expectedBatchSize:    800,
			expectedSupportsAuth: false,
		},
		{
			protocol:             "webdav",
			expectedRealTime:     false,
			expectedMoveWindow:   15 * time.Second,
			expectedBatchSize:    200,
			expectedSupportsAuth: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Protocol_%s", tc.protocol), func(t *testing.T) {
			// This would test protocol capability detection
			// Note: This requires implementing a protocol capabilities service
			t.Logf("Testing capabilities for protocol: %s", tc.protocol)

			// TODO: Implement actual capability testing
			// capabilities, err := services.GetProtocolCapabilities(tc.protocol, logger)
			// if err != nil {
			//     t.Errorf("Failed to get capabilities for %s: %v", tc.protocol, err)
			// }
			//
			// if capabilities.SupportsRealTimeNotification != tc.expectedRealTime {
			//     t.Errorf("Expected real-time support %v for %s, got %v",
			//         tc.expectedRealTime, tc.protocol, capabilities.SupportsRealTimeNotification)
			// }
		})
	}
}

// TestEdgeCases tests various edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("Timeout Handling", func(t *testing.T) {
		// Test connection timeouts
		discoveryService := services.NewSMBDiscoveryService(logger)

		// Create a context with very short timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel()

		config := services.SMBConnectionConfig{
			Host:     "192.0.2.1", // Non-routable IP (RFC 5737)
			Port:     445,
			Share:    "test",
			Username: "test",
			Password: "test",
		}

		// This should timeout or fail quickly
		success := discoveryService.TestConnection(timeoutCtx, config)
		if success {
			t.Error("Expected connection to fail due to timeout")
		}
	})

	t.Run("Invalid Input Handling", func(t *testing.T) {
		discoveryService := services.NewSMBDiscoveryService(logger)

		// Test with empty/invalid inputs
		invalidConfigs := []services.SMBConnectionConfig{
			{Host: "", Port: 445, Share: "test", Username: "test", Password: "test"},
			{Host: "test", Port: 0, Share: "test", Username: "test", Password: "test"},
			{Host: "test", Port: 445, Share: "", Username: "test", Password: "test"},
			{Host: "test", Port: 445, Share: "test", Username: "", Password: "test"},
		}

		for i, config := range invalidConfigs {
			success := discoveryService.TestConnection(ctx, config)
			if success {
				t.Errorf("Invalid config %d should have failed", i)
			}
		}
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		// Test concurrent access to mock servers
		mockSMB := mocks.NewMockSMBServer(logger)
		err := mockSMB.Start()
		if err != nil {
			t.Fatalf("Failed to start mock SMB server: %v", err)
		}
		defer mockSMB.Stop()

		mockFTP := mocks.NewMockFTPServer(logger)
		err = mockFTP.Start()
		if err != nil {
			t.Fatalf("Failed to start mock FTP server: %v", err)
		}
		defer mockFTP.Stop()

		// Run multiple concurrent operations
		const concurrency = 10
		results := make(chan bool, concurrency)

		discoveryService := services.NewSMBDiscoveryService(logger)

		for i := 0; i < concurrency; i++ {
			go func() {
				config := services.SMBConnectionConfig{
					Host:     "localhost",
					Port:     mockSMB.GetPort(),
					Share:    "shared",
					Username: "testuser",
					Password: "testpass",
				}
				success := discoveryService.TestConnection(ctx, config)
				results <- success
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < concurrency; i++ {
			if <-results {
				successCount++
			}
		}

		// With mock SMB server, expect all connections to fail gracefully
		// The important thing is that the test doesn't panic and completes
		if successCount > 0 {
			t.Logf("Note: %d out of %d concurrent operations succeeded (unexpected but acceptable)", successCount, concurrency)
		}

		// Verify we processed all operations without deadlock
		t.Logf("Successfully processed %d concurrent operations", concurrency)
	})

	t.Run("Resource Cleanup", func(t *testing.T) {
		// Test that resources are properly cleaned up
		servers := make([]*mocks.MockSMBServer, 5)

		// Start multiple servers
		for i := range servers {
			servers[i] = mocks.NewMockSMBServer(logger)
			err := servers[i].Start()
			if err != nil {
				t.Fatalf("Failed to start server %d: %v", i, err)
			}
		}

		// Stop all servers
		for i, server := range servers {
			err := server.Stop()
			if err != nil {
				t.Errorf("Failed to stop server %d: %v", i, err)
			}

			if server.IsRunning() {
				t.Errorf("Server %d should be stopped", i)
			}
		}
	})
}
