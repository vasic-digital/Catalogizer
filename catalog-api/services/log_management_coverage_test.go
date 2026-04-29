package services

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCompressLogFile_Roundtrip exercises the compression path: creates a
// temp log, compresses it, verifies the .gz exists, decompresses it, and
// confirms the original contents are preserved.
func TestCompressLogFile_Roundtrip(t *testing.T) {
	svc := &LogManagementService{}

	dir := t.TempDir()
	srcPath := filepath.Join(dir, "test.log")
	content := []byte("line 1\nline 2\nline 3 with some more data to ensure gzip has something to compress\n")
	require.NoError(t, os.WriteFile(srcPath, content, 0600))

	require.NoError(t, svc.compressLogFile(srcPath))

	// Original file must be removed after compression.
	_, err := os.Stat(srcPath)
	require.True(t, os.IsNotExist(err), "original file should have been removed")

	// Compressed file must exist and decompress back to original.
	gzPath := srcPath + ".gz"
	f, err := os.Open(gzPath)
	require.NoError(t, err)
	defer f.Close()

	gr, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	require.Equal(t, content, decompressed)
}

// TestCompressLogFile_MissingSource verifies the error path when the
// source file doesn't exist.
func TestCompressLogFile_MissingSource(t *testing.T) {
	svc := &LogManagementService{}
	err := svc.compressLogFile(filepath.Join(t.TempDir(), "nonexistent.log"))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

// TestFileLogCollector_CollectLogs exercises the basic collector contract.
func TestFileLogCollector_CollectLogs(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "component.log")
	require.NoError(t, os.WriteFile(logPath, []byte("some log content"), 0600))

	collector := &FileLogCollector{
		logPath:       logPath,
		componentName: "component",
	}

	// Current implementation opens the file and returns nil/empty —
	// we assert only the contract: no error on a readable file.
	entries, err := collector.CollectLogs()
	require.NoError(t, err)
	_ = entries // implementation returns empty/nil; either is acceptable

	require.Equal(t, logPath, collector.GetLogPath())
	require.Equal(t, "component", collector.GetComponentName())
}

// TestFileLogCollector_CollectLogs_MissingFile verifies the error path
// when the log file doesn't exist.
func TestFileLogCollector_CollectLogs_MissingFile(t *testing.T) {
	collector := &FileLogCollector{
		logPath:       filepath.Join(t.TempDir(), "missing.log"),
		componentName: "component",
	}

	entries, err := collector.CollectLogs()
	require.Error(t, err)
	require.Nil(t, entries)
}

// TestLogManagementService_Close_IsSafeToCallTwice verifies the shutdown
// path doesn't panic on repeated calls when no goroutines are in flight.
func TestLogManagementService_Close_IsSafeToCallTwice(t *testing.T) {
	// bluff-scan: no-assert-ok (idempotency smoke — repeated/extra calls must not panic)
	svc := &LogManagementService{}
	svc.Close()
	svc.Close()
}

// TestLogManagementService_InitializeCollectors_RegistersAllDefaults
// exercises the collector registration path.
func TestLogManagementService_InitializeCollectors_RegistersAllDefaults(t *testing.T) {
	svc := NewLogManagementService(nil)
	require.NotNil(t, svc)
	require.NotNil(t, svc.logCollectors)

	// initializeCollectors is called from NewLogManagementService. It
	// should register the default file-based collectors for each of the
	// 6 standard components.
	want := []string{"api", "auth", "sync", "conversion", "stress_test", "error_reporting"}
	for _, name := range want {
		_, ok := svc.logCollectors[name]
		require.True(t, ok, "expected collector %q to be registered", name)
	}
}
