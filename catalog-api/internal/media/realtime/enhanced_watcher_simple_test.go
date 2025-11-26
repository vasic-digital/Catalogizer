package realtime

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestEnhancedChangeWatcher_BasicFunctionality tests core functionality without database
func TestEnhancedChangeWatcher_BasicFunctionality(t *testing.T) {
	logger := zap.NewNop()

	// Create a basic watcher without database components
	watcher := &EnhancedChangeWatcher{
		logger: logger,
		stopCh: make(chan struct{}),
	}

	// Test getRelativePath method directly
	tests := []struct {
		name     string
		basePath string
		fullPath string
		expected string
		hasError bool
	}{
		{
			name:     "simple relative path",
			basePath: "/mnt/storage",
			fullPath: "/mnt/storage/test/file.txt",
			expected: "/test/file.txt",
			hasError: false,
		},
		{
			name:     "root file",
			basePath: "/mnt/storage",
			fullPath: "/mnt/storage/file.txt",
			expected: "/file.txt",
			hasError: false,
		},
		{
			name:     "same path",
			basePath: "/mnt/storage",
			fullPath: "/mnt/storage",
			expected: "/.",
			hasError: false,
		},
		{
			name:     "nested directory",
			basePath: "/mnt/storage",
			fullPath: "/mnt/storage/dir1/dir2/file.txt",
			expected: "/dir1/dir2/file.txt",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := watcher.getRelativePath(tt.basePath, tt.fullPath)

			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.hasError && result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestEnhancedChangeWatcher_MediaFileDetection tests media file detection without database
func TestEnhancedChangeWatcher_MediaFileDetection(t *testing.T) {
	logger := zap.NewNop()

	watcher := &EnhancedChangeWatcher{
		logger: logger,
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/video.mp4", true},
		{"/movie.avi", true},
		{"/song.mp3", true},
		{"/image.jpg", true},
		{"/document.pdf", true},
		{"/text.txt", false},
		{"/script.sh", false},
		{"/program.exe", false},
		{"/archive.zip", false},
		{"/no_extension", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := watcher.isMediaFile(tt.path)
			if result != tt.expected {
				t.Errorf("isMediaFile(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestEnhancedChangeWatcher_QueueOperations tests queue operations without database
func TestEnhancedChangeWatcher_QueueOperations(t *testing.T) {
	logger := zap.NewNop()

	// Create watcher with a reasonable queue size
	watcher := &EnhancedChangeWatcher{
		logger:      logger,
		changeQueue: make(chan EnhancedChangeEvent, 100),
		stopCh:      make(chan struct{}),
	}

	// Test change queue operations
	testEvent := EnhancedChangeEvent{
		Path:      "/test.txt",
		SmbRoot:   "test_storage",
		Operation: "created",
		Timestamp: time.Now(),
		Size:      1024,
		IsDir:     false,
	}

	// Test queue send
	select {
	case watcher.changeQueue <- testEvent:
		// Successfully sent
	case <-time.After(100 * time.Millisecond):
		t.Error("Failed to send event to queue")
	}

	// Test queue receive
	select {
	case receivedEvent := <-watcher.changeQueue:
		if receivedEvent.Path != testEvent.Path {
			t.Errorf("Expected path %s, got %s", testEvent.Path, receivedEvent.Path)
		}
		if receivedEvent.Operation != testEvent.Operation {
			t.Errorf("Expected operation %s, got %s", testEvent.Operation, receivedEvent.Operation)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Failed to receive event from queue")
	}
}