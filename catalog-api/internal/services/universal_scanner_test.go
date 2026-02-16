package services

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewUniversalScanner(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()

	scanner := NewUniversalScanner(mockDB, mockLogger, nil, nil)

	assert.NotNil(t, scanner)
}

func TestNewLocalScanner(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewLocalScanner(mockLogger)

	assert.NotNil(t, scanner)
}

func TestNewSMBScanner(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewSMBScanner(mockLogger)

	assert.NotNil(t, scanner)
}

func TestNewFTPScanner(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewFTPScanner(mockLogger)

	assert.NotNil(t, scanner)
}

func TestNewNFSScanner(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewNFSScanner(mockLogger)

	assert.NotNil(t, scanner)
}

func TestNewWebDAVScanner(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewWebDAVScanner(mockLogger)

	assert.NotNil(t, scanner)
}

func TestLocalScanner_GetScanStrategy(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewLocalScanner(mockLogger)

	strategy := scanner.GetScanStrategy()
	assert.True(t, strategy.UseRecursiveListing)
	assert.True(t, strategy.ParallelDirectories)
	assert.True(t, strategy.ChecksumCalculation)
	assert.True(t, strategy.MetadataExtraction)
	assert.True(t, strategy.RealTimeChangeDetection)
	assert.Equal(t, 1000, strategy.BatchSize)
}

func TestSMBScanner_GetScanStrategy(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewSMBScanner(mockLogger)

	strategy := scanner.GetScanStrategy()
	assert.False(t, strategy.UseRecursiveListing)
	assert.False(t, strategy.ParallelDirectories)
	assert.False(t, strategy.ChecksumCalculation)
	assert.True(t, strategy.MetadataExtraction)
	assert.False(t, strategy.RealTimeChangeDetection)
	assert.Equal(t, 500, strategy.BatchSize)
}

func TestFTPScanner_GetScanStrategy(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewFTPScanner(mockLogger)

	strategy := scanner.GetScanStrategy()
	assert.False(t, strategy.UseRecursiveListing)
	assert.False(t, strategy.ParallelDirectories)
	assert.False(t, strategy.ChecksumCalculation)
	assert.False(t, strategy.MetadataExtraction)
	assert.False(t, strategy.RealTimeChangeDetection)
	assert.Equal(t, 100, strategy.BatchSize)
}

func TestNFSScanner_GetScanStrategy(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewNFSScanner(mockLogger)

	strategy := scanner.GetScanStrategy()
	assert.True(t, strategy.UseRecursiveListing)
	assert.True(t, strategy.ParallelDirectories)
	assert.True(t, strategy.ChecksumCalculation)
	assert.True(t, strategy.MetadataExtraction)
	assert.False(t, strategy.RealTimeChangeDetection)
	assert.Equal(t, 800, strategy.BatchSize)
}

func TestWebDAVScanner_GetScanStrategy(t *testing.T) {
	mockLogger := zap.NewNop()
	scanner := NewWebDAVScanner(mockLogger)

	strategy := scanner.GetScanStrategy()
	assert.False(t, strategy.UseRecursiveListing)
	assert.False(t, strategy.ParallelDirectories)
	assert.False(t, strategy.ChecksumCalculation)
	assert.True(t, strategy.MetadataExtraction)
	assert.False(t, strategy.RealTimeChangeDetection)
	assert.Equal(t, 200, strategy.BatchSize)
}

func TestScanners_SupportsIncrementalScan(t *testing.T) {
	mockLogger := zap.NewNop()

	tests := []struct {
		name     string
		scanner  ProtocolScanner
		expected bool
	}{
		{
			name:     "local supports incremental",
			scanner:  NewLocalScanner(mockLogger),
			expected: true,
		},
		{
			name:     "SMB supports incremental",
			scanner:  NewSMBScanner(mockLogger),
			expected: true,
		},
		{
			name:     "FTP does not support incremental",
			scanner:  NewFTPScanner(mockLogger),
			expected: false,
		},
		{
			name:     "NFS supports incremental",
			scanner:  NewNFSScanner(mockLogger),
			expected: true,
		},
		{
			name:     "WebDAV does not support incremental",
			scanner:  NewWebDAVScanner(mockLogger),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scanner.SupportsIncrementalScan()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScanners_GetOptimalBatchSize(t *testing.T) {
	mockLogger := zap.NewNop()

	tests := []struct {
		name     string
		scanner  ProtocolScanner
		expected int
	}{
		{
			name:     "local batch size",
			scanner:  NewLocalScanner(mockLogger),
			expected: 1000,
		},
		{
			name:     "SMB batch size",
			scanner:  NewSMBScanner(mockLogger),
			expected: 500,
		},
		{
			name:     "FTP batch size",
			scanner:  NewFTPScanner(mockLogger),
			expected: 100,
		},
		{
			name:     "NFS batch size",
			scanner:  NewNFSScanner(mockLogger),
			expected: 800,
		},
		{
			name:     "WebDAV batch size",
			scanner:  NewWebDAVScanner(mockLogger),
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scanner.GetOptimalBatchSize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUniversalScanner_RegisterProtocolScanner(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()
	scanner := NewUniversalScanner(mockDB, mockLogger, nil, nil)

	// Register a custom scanner
	customScanner := NewLocalScanner(mockLogger)
	scanner.RegisterProtocolScanner("custom", customScanner)

	// Verify it was registered (no direct accessor, but no panic means success)
	assert.NotNil(t, scanner)
}

func TestUniversalScanner_GetActiveScanStatus(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()
	scanner := NewUniversalScanner(mockDB, mockLogger, nil, nil)

	status, exists := scanner.GetActiveScanStatus("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, status)
}

func TestUniversalScanner_GetAllActiveScanStatuses(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()
	scanner := NewUniversalScanner(mockDB, mockLogger, nil, nil)

	statuses := scanner.GetAllActiveScanStatuses()
	assert.NotNil(t, statuses)
	assert.Equal(t, 0, len(statuses))
}

func TestScanStatus_UpdateStatus(t *testing.T) {
	status := &ScanStatus{
		JobID:  "test_job",
		Status: "running",
	}

	status.updateStatus("completed")
	snapshot := status.GetSnapshot()
	assert.Equal(t, "completed", snapshot.Status)
}

func TestScanStatus_UpdateCurrentPath(t *testing.T) {
	status := &ScanStatus{
		JobID: "test_job",
	}

	status.updateCurrentPath("/some/path")
	snapshot := status.GetSnapshot()
	assert.Equal(t, "/some/path", snapshot.CurrentPath)
}

func TestScanStatus_IncrementCounters(t *testing.T) {
	status := &ScanStatus{
		JobID: "test_job",
	}

	status.incrementCounters(10, 8, 5, 2, 1)
	snapshot := status.GetSnapshot()
	assert.Equal(t, int64(10), snapshot.FilesProcessed)
	assert.Equal(t, int64(8), snapshot.FilesFound)
	assert.Equal(t, int64(5), snapshot.FilesUpdated)
	assert.Equal(t, int64(2), snapshot.FilesDeleted)
	assert.Equal(t, int64(1), snapshot.ErrorCount)

	// Increment again
	status.incrementCounters(5, 3, 2, 1, 0)
	snapshot = status.GetSnapshot()
	assert.Equal(t, int64(15), snapshot.FilesProcessed)
	assert.Equal(t, int64(11), snapshot.FilesFound)
	assert.Equal(t, int64(7), snapshot.FilesUpdated)
	assert.Equal(t, int64(3), snapshot.FilesDeleted)
	assert.Equal(t, int64(1), snapshot.ErrorCount)
}

func TestScanStatus_GetSnapshot(t *testing.T) {
	status := &ScanStatus{
		JobID:           "test_job",
		StorageRootName: "media",
		Protocol:        "local",
		Status:          "running",
	}

	snapshot := status.GetSnapshot()
	assert.Equal(t, "test_job", snapshot.JobID)
	assert.Equal(t, "media", snapshot.StorageRootName)
	assert.Equal(t, "local", snapshot.Protocol)
	assert.Equal(t, "running", snapshot.Status)
}
