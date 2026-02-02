package smb

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func TestNewResilientSMBManager(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.sources)
	assert.NotNil(t, manager.offlineCache)
	assert.NotNil(t, manager.healthChecker)
	assert.NotNil(t, manager.eventChannel)
	assert.NotNil(t, manager.stopChannel)
	assert.Equal(t, 0, len(manager.sources))
}

func TestAddSource(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		Name: "Test Share",
		Path: "//server/share",
	}

	err := manager.AddSource(source)
	assert.NoError(t, err)
	assert.NotEmpty(t, source.ID)
	assert.Equal(t, 5, source.MaxRetryAttempts)
	assert.Equal(t, 30*time.Second, source.RetryDelay)
	assert.Equal(t, 30*time.Second, source.ConnectionTimeout)
	assert.True(t, source.IsEnabled)
	assert.Equal(t, 1, len(manager.sources))

	// Wait for the connection goroutine to finish
	time.Sleep(200 * time.Millisecond)
}

func TestAddSourceWithCustomValues(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:               "custom-id",
		Name:             "Custom Share",
		Path:             "//server/share",
		MaxRetryAttempts: 10,
		RetryDelay:       10 * time.Second,
		ConnectionTimeout: 15 * time.Second,
	}

	err := manager.AddSource(source)
	assert.NoError(t, err)
	assert.Equal(t, "custom-id", source.ID)
	assert.Equal(t, 10, source.MaxRetryAttempts)
	assert.Equal(t, 10*time.Second, source.RetryDelay)
	assert.Equal(t, 15*time.Second, source.ConnectionTimeout)

	time.Sleep(200 * time.Millisecond)
}

func TestRemoveSource(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:   "test-source",
		Name: "Test Share",
		Path: "//server/share",
	}
	manager.AddSource(source)
	time.Sleep(200 * time.Millisecond)

	err := manager.RemoveSource("test-source")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(manager.sources))

	// Remove non-existent source
	err = manager.RemoveSource("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source not found")
}

func TestGetSourceStatus(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:   "status-test",
		Name: "Status Test Share",
		Path: "//server/share",
	}
	manager.AddSource(source)
	time.Sleep(200 * time.Millisecond)

	status := manager.GetSourceStatus()
	assert.Contains(t, status, "status-test")

	info, ok := status["status-test"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Status Test Share", info["name"])
	assert.Equal(t, "//server/share", info["path"])
	assert.Equal(t, true, info["is_enabled"])
}

func TestForceReconnect(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:   "reconnect-test",
		Name: "Reconnect Test",
		Path: "//server/share",
	}
	manager.AddSource(source)
	time.Sleep(200 * time.Millisecond)

	err := manager.ForceReconnect("reconnect-test")
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	// Non-existent source
	err = manager.ForceReconnect("non-existent")
	assert.Error(t, err)
}

func TestForceReconnectDisabledSource(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:   "disabled-test",
		Name: "Disabled Test",
		Path: "//server/share",
	}
	manager.AddSource(source)
	time.Sleep(200 * time.Millisecond)

	// Disable the source (use mutex to avoid data race with background goroutine)
	source.mutex.Lock()
	source.IsEnabled = false
	source.mutex.Unlock()

	err := manager.ForceReconnect("disabled-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source is disabled")
}

func TestConnectionState_String(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateConnected, "connected"},
		{StateDisconnected, "disconnected"},
		{StateReconnecting, "reconnecting"},
		{StateOffline, "offline"},
		{ConnectionState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestEventType_String(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventConnected, "connected"},
		{EventDisconnected, "disconnected"},
		{EventReconnecting, "reconnecting"},
		{EventOffline, "offline"},
		{EventFileChange, "file_change"},
		{EventError, "error"},
		{EventHealthCheck, "health_check"},
		{EventType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.eventType.String())
		})
	}
}

func TestSendEventBackpressure(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Fill the event channel
	for i := 0; i < 1000; i++ {
		manager.eventChannel <- SMBEvent{
			Type:      EventHealthCheck,
			SourceID:  "filler",
			Timestamp: time.Now(),
		}
	}

	// Channel is now full; sendEvent should handle backpressure (drop oldest)
	manager.sendEvent(SMBEvent{
		Type:      EventConnected,
		SourceID:  "important",
		Timestamp: time.Now(),
	})

	// Verify channel is still full (1000 items)
	assert.Equal(t, 1000, len(manager.eventChannel))
}

func TestOfflineCache(t *testing.T) {
	logger := testLogger()
	cache := NewOfflineCache(5, logger)

	// Cache some changes
	cache.CacheChange("source1", "/path/file1.mp4")
	cache.CacheChange("source1", "/path/file2.mp4")
	cache.CacheChange("source2", "/path/file3.mp4")

	assert.Equal(t, 3, len(cache.entries))

	// Process cached changes for source1
	cache.ProcessCachedChanges("source1")

	// Verify entries are marked available
	for _, entry := range cache.entries {
		if entry.SourceID == "source1" {
			assert.True(t, entry.IsAvailable)
		}
	}
}

func TestOfflineCacheEviction(t *testing.T) {
	logger := testLogger()
	cache := NewOfflineCache(3, logger)

	// Fill cache to max
	cache.CacheChange("s1", "/file1")
	time.Sleep(10 * time.Millisecond)
	cache.CacheChange("s1", "/file2")
	time.Sleep(10 * time.Millisecond)
	cache.CacheChange("s1", "/file3")

	assert.Equal(t, 3, len(cache.entries))

	// Adding one more should evict the oldest
	cache.CacheChange("s1", "/file4")
	assert.Equal(t, 3, len(cache.entries))
}

func TestStartAndStop(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	err := manager.Start()
	assert.NoError(t, err)

	// Give goroutines time to start
	time.Sleep(100 * time.Millisecond)

	err = manager.Stop()
	assert.NoError(t, err)
}

func TestIsSourceConnected(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Non-existent source
	assert.False(t, manager.isSourceConnected("non-existent"))

	source := &SMBSource{
		ID:    "test",
		Name:  "Test",
		Path:  "//server/share",
		State: StateConnected,
	}
	manager.mutex.Lock()
	manager.sources["test"] = source
	manager.mutex.Unlock()

	assert.True(t, manager.isSourceConnected("test"))

	source.mutex.Lock()
	source.State = StateDisconnected
	source.mutex.Unlock()

	assert.False(t, manager.isSourceConnected("test"))
}

func TestUpdateSource(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:   "update-test",
		Name: "Update Test",
		Path: "//server/share",
	}
	manager.AddSource(source)
	time.Sleep(200 * time.Millisecond)

	err := manager.UpdateSource("update-test", nil)
	assert.NoError(t, err)

	err = manager.UpdateSource("non-existent", nil)
	assert.Error(t, err)
}

func TestConcurrentSourceOperations(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	var wg sync.WaitGroup

	// Concurrently add sources with explicit unique IDs
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			source := &SMBSource{
				ID:   fmt.Sprintf("concurrent-%d", idx),
				Name: "Concurrent Test",
				Path: "//server/share",
			}
			manager.AddSource(source)
		}(i)
	}

	wg.Wait()
	time.Sleep(300 * time.Millisecond)

	status := manager.GetSourceStatus()
	assert.Equal(t, 10, len(status))
}

func TestGetStartTime(t *testing.T) {
	logger := testLogger()
	before := time.Now()
	manager := NewResilientSMBManager(logger, 100)
	after := time.Now()

	startTime := manager.GetStartTime()
	assert.True(t, !startTime.Before(before))
	assert.True(t, !startTime.After(after))
}
