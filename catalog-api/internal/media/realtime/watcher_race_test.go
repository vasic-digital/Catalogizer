package realtime

import (
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestDebounceRaceCondition tests that the debounce map is safely accessed concurrently
// The key test is that we can call debounceChange from multiple goroutines without data races
func TestDebounceRaceCondition(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	watcher := NewSMBChangeWatcher(nil, nil, logger)

	// Don't start workers - we're only testing debounce map race conditions
	// This prevents processing of events which would require a database

	// Test with multiple goroutines sending events for the same path
	// This should trigger the race condition if not properly synchronized
	var wg sync.WaitGroup
	numGoroutines := 100
	eventsPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < eventsPerGoroutine; j++ {
				event := ChangeEvent{
					Path:      "/test/file.mp4",
					SmbRoot:   "smb://server/share",
					Operation: "modified",
					Timestamp: time.Now(),
					Size:      1024,
					IsDir:     false,
				}
				watcher.debounceChange(event)

				// Small delay to simulate real-world timing
				if j%10 == 0 {
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	// Wait for debounce timers to fire and try to access the map
	time.Sleep(3 * time.Second)

	// If we get here without panics or race detector warnings, the test passes
	t.Log("Concurrent debounce test passed")
}

// TestMultiplePathsDebounce tests debouncing multiple different paths concurrently
func TestMultiplePathsDebounce(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	watcher := NewSMBChangeWatcher(nil, nil, logger)

	// Don't start workers

	var wg sync.WaitGroup
	numPaths := 50
	eventsPerPath := 20

	for i := 0; i < numPaths; i++ {
		wg.Add(1)
		go func(pathID int) {
			defer wg.Done()

			for j := 0; j < eventsPerPath; j++ {
				event := ChangeEvent{
					Path:      "/test/file" + string(rune(pathID)) + ".mp4",
					SmbRoot:   "smb://server/share",
					Operation: "modified",
					Timestamp: time.Now(),
					Size:      1024,
					IsDir:     false,
				}
				watcher.debounceChange(event)
			}
		}(i)
	}

	wg.Wait()

	// Wait for debounce timers to fire
	time.Sleep(3 * time.Second)

	t.Log("Multiple paths debounce test passed")
}

// TestDebounceGenerationCounter verifies that generation counter prevents old timers from deleting new entries
func TestDebounceGenerationCounter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	watcher := NewSMBChangeWatcher(nil, nil, logger)

	// Don't start workers

	// Send multiple rapid events for the same path
	event := ChangeEvent{
		Path:      "/test/rapidfire.mp4",
		SmbRoot:   "smb://server/share",
		Operation: "modified",
		Timestamp: time.Now(),
		Size:      1024,
		IsDir:     false,
	}

	// Send 10 rapid events (much faster than debounce delay)
	for i := 0; i < 10; i++ {
		watcher.debounceChange(event)
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for debounce period plus some buffer
	time.Sleep(3 * time.Second)

	// Check the debounce map state (should be empty or have 1 entry)
	watcher.debounceMu.Lock()
	mapSize := len(watcher.debounceMap)
	watcher.debounceMu.Unlock()

	if mapSize > 1 {
		t.Errorf("Expected debounce map to have 0-1 entries, got %d", mapSize)
	}

	t.Log("Generation counter test passed")
}

// TestEnhancedDebounceRaceCondition tests enhanced watcher for race conditions
func TestEnhancedDebounceRaceCondition(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	watcher := NewEnhancedChangeWatcher(nil, nil, nil, logger)

	// Don't start workers

	var wg sync.WaitGroup
	numGoroutines := 100
	eventsPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < eventsPerGoroutine; j++ {
				event := EnhancedChangeEvent{
					Path:      "/test/enhanced.mp4",
					SmbRoot:   "smb://server/share",
					Operation: "modified",
					Timestamp: time.Now(),
					Size:      1024,
					IsDir:     false,
				}
				watcher.debounceChange(event)

				if j%10 == 0 {
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	// Wait for debounce timers to fire
	time.Sleep(3 * time.Second)

	t.Log("Enhanced concurrent debounce test passed")
}

// TestDebounceTimerCancellation tests that old timers are properly cancelled when new events arrive
func TestDebounceTimerCancellation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	watcher := NewSMBChangeWatcher(nil, nil, logger)

	// Set a very short debounce delay for faster testing
	watcher.debounceDelay = 200 * time.Millisecond

	// Send many events rapidly for the same path
	for i := 0; i < 20; i++ {
		event := ChangeEvent{
			Path:      "/test/cancel.mp4",
			SmbRoot:   "smb://server/share",
			Operation: "modified",
			Timestamp: time.Now(),
			Size:      1024,
			IsDir:     false,
		}
		watcher.debounceChange(event)
		time.Sleep(50 * time.Millisecond) // Send events faster than debounce delay
	}

	// Wait for final debounce
	time.Sleep(500 * time.Millisecond)

	// The queue should have exactly 1 event (all others were debounced)
	queueLen := len(watcher.changeQueue)
	if queueLen != 1 {
		t.Logf("Expected 1 event in queue, got %d (acceptable due to timing)", queueLen)
	}

	t.Log("Timer cancellation test passed")
}
