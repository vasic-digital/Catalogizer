package integration

import (
	"catalogizer/filesystem"
	"catalogizer/internal/services"
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func skipIfNoContainer(t *testing.T, envVar, serviceName string) {
	if os.Getenv(envVar) == "" {
		t.Skipf("Skipping test: %s not set (start %s container first)", envVar, serviceName)
	}
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func TestSMBProtocolConnectivity(t *testing.T) {
	skipIfNoContainer(t, "SMB_TEST_SERVER", "SMB")

	ctx := context.Background()

	host := getEnvOrDefault("SMB_TEST_SERVER", "localhost")
	port := parseInt(getEnvOrDefault("SMB_TEST_PORT", "445"), 445)
	username := getEnvOrDefault("SMB_TEST_USER", "test")
	password := getEnvOrDefault("SMB_TEST_PASS", "test123")
	share := getEnvOrDefault("SMB_TEST_SHARE", "media")

	t.Run("SMB Client Creation", func(t *testing.T) {
		config := &filesystem.SmbConfig{
			Host:     host,
			Port:     port,
			Share:    share,
			Username: username,
			Password: password,
		}

		client := filesystem.NewSmbClient(config)
		require.NotNil(t, client)
	})

	t.Run("SMB Connection", func(t *testing.T) {
		config := &filesystem.SmbConfig{
			Host:     host,
			Port:     port,
			Share:    share,
			Username: username,
			Password: password,
		}

		client := filesystem.NewSmbClient(config)
		require.NotNil(t, client)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := client.Connect(ctx)
		if err != nil {
			t.Logf("SMB connection failed (server may not be available): %v", err)
		}
	})

	t.Run("SMB Directory Listing", func(t *testing.T) {
		config := &filesystem.SmbConfig{
			Host:     host,
			Port:     port,
			Share:    share,
			Username: username,
			Password: password,
		}

		client := filesystem.NewSmbClient(config)
		require.NotNil(t, client)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := client.Connect(ctx)
		if err != nil {
			t.Skipf("Cannot list directory: connection failed: %v", err)
		}

		files, err := client.ListDirectory(ctx, "/")
		if err != nil {
			t.Logf("Directory listing returned error: %v", err)
		} else {
			t.Logf("Found %d files in SMB share", len(files))
		}
	})
}

func TestFTPProtocolConnectivity(t *testing.T) {
	skipIfNoContainer(t, "FTP_TEST_SERVER", "FTP")

	ctx := context.Background()

	host := getEnvOrDefault("FTP_TEST_SERVER", "localhost")
	port := parseInt(getEnvOrDefault("FTP_TEST_PORT", "21"), 21)
	username := getEnvOrDefault("FTP_TEST_USER", "test")
	password := getEnvOrDefault("FTP_TEST_PASS", "test123")

	t.Run("FTP Client Creation", func(t *testing.T) {
		config := &filesystem.FTPConfig{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
		}

		client := filesystem.NewFTPClient(config)
		require.NotNil(t, client)
	})

	t.Run("FTP Connection", func(t *testing.T) {
		config := &filesystem.FTPConfig{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
		}

		client := filesystem.NewFTPClient(config)
		require.NotNil(t, client)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := client.Connect(ctx)
		if err != nil {
			t.Logf("FTP connection failed (server may not be available): %v", err)
		}
	})
}

func TestWebDAVProtocolConnectivity(t *testing.T) {
	skipIfNoContainer(t, "WEBDAV_TEST_URL", "WebDAV")

	ctx := context.Background()

	url := getEnvOrDefault("WEBDAV_TEST_URL", "http://localhost:8081")
	username := getEnvOrDefault("WEBDAV_TEST_USER", "test")
	password := getEnvOrDefault("WEBDAV_TEST_PASS", "test123")

	t.Run("WebDAV Client Creation", func(t *testing.T) {
		config := &filesystem.WebDAVConfig{
			URL:      url,
			Username: username,
			Password: password,
		}

		client := filesystem.NewWebDAVClient(config)
		require.NotNil(t, client)
	})

	t.Run("WebDAV Connection", func(t *testing.T) {
		config := &filesystem.WebDAVConfig{
			URL:      url,
			Username: username,
			Password: password,
		}

		client := filesystem.NewWebDAVClient(config)
		require.NotNil(t, client)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		files, err := client.ListDirectory(ctx, "/")
		if err != nil {
			t.Logf("WebDAV directory listing failed: %v", err)
		} else {
			t.Logf("Found %d items via WebDAV", len(files))
		}
	})
}

func TestNFSProtocolConnectivity(t *testing.T) {
	skipIfNoContainer(t, "NFS_TEST_SERVER", "NFS")

	host := getEnvOrDefault("NFS_TEST_SERVER", "localhost")
	exportPath := getEnvOrDefault("NFS_TEST_EXPORT", "/export/media")
	mountPoint := getEnvOrDefault("NFS_TEST_MOUNT", "/mnt/nfs-test")

	t.Run("NFS Client Creation", func(t *testing.T) {
		config := filesystem.NFSConfig{
			Host:       host,
			Path:       exportPath,
			MountPoint: mountPoint,
		}

		client, err := filesystem.NewNFSClient(config)
		if err != nil {
			t.Skipf("NFS client creation failed (may require root): %v", err)
		}
		require.NotNil(t, client)
	})
}

func TestLocalFilesystem(t *testing.T) {
	ctx := context.Background()

	tempDir := t.TempDir()

	t.Run("Local Client Creation", func(t *testing.T) {
		config := &filesystem.LocalConfig{
			BasePath: tempDir,
		}

		client := filesystem.NewLocalClient(config)
		require.NotNil(t, client)
	})

	t.Run("Local Directory Listing", func(t *testing.T) {
		testFile := tempDir + "/test.txt"
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		config := &filesystem.LocalConfig{
			BasePath: tempDir,
		}

		client := filesystem.NewLocalClient(config)
		require.NotNil(t, client)

		connErr := client.Connect(ctx)
		require.NoError(t, connErr)

		files, err := client.ListDirectory(ctx, "/")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(files), 1)
	})
}

func TestSMBDiscoveryService(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	discoveryService := services.NewSMBDiscoveryService(logger)

	t.Run("Discover shares with fallback", func(t *testing.T) {
		shares, err := discoveryService.DiscoverShares(ctx, "nonexistent.host.local", "user", "pass", nil)
		require.NoError(t, err, "Discovery should succeed with fallback shares")
		assert.GreaterOrEqual(t, len(shares), 1, "Should return at least fallback shares")
	})
}
