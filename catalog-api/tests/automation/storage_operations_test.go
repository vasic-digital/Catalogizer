package automation

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStorageOperationsFullFlow tests the complete storage operations workflow
func TestStorageOperationsFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping automation test in short mode")
	}

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api/v1"
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	t.Run("Get Storage Roots", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/storage/roots")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		assert.Contains(t, result, "roots")
		roots := result["roots"].([]interface{})
		assert.NotEmpty(t, roots)

		t.Logf("Found %d storage roots", len(roots))
	})

	t.Run("List Storage Path", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/storage/list/test?storage_id=local")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		assert.Contains(t, result, "path")
		assert.Contains(t, result, "storage_id")
		assert.Equal(t, "/test", result["path"])
		assert.Equal(t, "local", result["storage_id"])
	})

	t.Run("Copy to Storage", func(t *testing.T) {
		requestBody := map[string]string{
			"source_path": "/tmp/test.txt",
			"dest_path":   "/storage/test.txt",
			"storage_id":  "local",
		}

		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)

		resp, err := client.Post(
			baseURL+"/copy/storage",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		assert.Contains(t, result, "message")
		assert.Contains(t, result, "source")
		assert.Contains(t, result, "destination")
	})

	t.Run("Error Handling - Missing Parameters", func(t *testing.T) {
		requestBody := map[string]string{
			"source_path": "/tmp/test.txt",
			// Missing dest_path and storage_id
		}

		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)

		resp, err := client.Post(
			baseURL+"/copy/storage",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestFileSystemServiceIntegration tests filesystem service integration
func TestFileSystemServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "fs_service_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = ioutil.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	t.Run("Local Filesystem Operations", func(t *testing.T) {
		// Test that the file exists
		_, err := os.Stat(testFile)
		assert.NoError(t, err)

		// Test reading directory
		files, err := ioutil.ReadDir(tempDir)
		require.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, "test.txt", files[0].Name())
	})
}
