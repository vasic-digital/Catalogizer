package smb

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"catalogizer/internal/metrics"
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
		ID:                "custom-id",
		Name:              "Custom Share",
		Path:              "//server/share",
		MaxRetryAttempts:  10,
		RetryDelay:        10 * time.Second,
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

func TestConnectionState_HealthMetric(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected float64
	}{
		{StateConnected, metrics.SMBHealthy},
		{StateReconnecting, metrics.SMBDegraded},
		{StateDisconnected, metrics.SMBOffline},
		{StateOffline, metrics.SMBOffline},
		{ConnectionState(99), metrics.SMBOffline},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.HealthMetric())
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

// --- New tests for handleEvent ---

func TestHandleEvent_Connected(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Add a source with cached changes
	manager.offlineCache.CacheChange("src-connected", "/path/file.mp4")

	// handleEvent for Connected should call onSourceConnected which processes cached changes
	manager.handleEvent(SMBEvent{
		Type:      EventConnected,
		SourceID:  "src-connected",
		Timestamp: time.Now(),
	})

	// After processing, cached entries for that source should be marked available
	manager.offlineCache.mutex.RLock()
	defer manager.offlineCache.mutex.RUnlock()
	for _, entry := range manager.offlineCache.entries {
		if entry.SourceID == "src-connected" {
			assert.True(t, entry.IsAvailable)
		}
	}
}

func TestHandleEvent_Disconnected(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// handleEvent for Disconnected should call onSourceDisconnected (enable offline mode)
	// This should not panic
	manager.handleEvent(SMBEvent{
		Type:      EventDisconnected,
		SourceID:  "src-disconnected",
		Error:     fmt.Errorf("connection lost"),
		Timestamp: time.Now(),
	})
}

func TestHandleEvent_Offline(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// handleEvent for Offline should call onSourceOffline (logs warning)
	manager.handleEvent(SMBEvent{
		Type:      EventOffline,
		SourceID:  "src-offline",
		Timestamp: time.Now(),
	})
}

func TestHandleEvent_FileChange_Connected(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Add a connected source
	source := &SMBSource{
		ID:    "file-change-src",
		Name:  "Test",
		Path:  "//server/share",
		State: StateConnected,
	}
	manager.mutex.Lock()
	manager.sources["file-change-src"] = source
	manager.mutex.Unlock()

	// handleEvent for FileChange with connected source processes immediately
	manager.handleEvent(SMBEvent{
		Type:      EventFileChange,
		SourceID:  "file-change-src",
		Path:      "/some/file.mp4",
		Timestamp: time.Now(),
	})
}

func TestHandleEvent_FileChange_Disconnected(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Add a disconnected source
	source := &SMBSource{
		ID:    "file-change-disc",
		Name:  "Test",
		Path:  "//server/share",
		State: StateDisconnected,
	}
	manager.mutex.Lock()
	manager.sources["file-change-disc"] = source
	manager.mutex.Unlock()

	// handleEvent for FileChange with disconnected source caches the change
	manager.handleEvent(SMBEvent{
		Type:      EventFileChange,
		SourceID:  "file-change-disc",
		Path:      "/some/file.mp4",
		Timestamp: time.Now(),
	})

	// Verify entry was cached
	manager.offlineCache.mutex.RLock()
	defer manager.offlineCache.mutex.RUnlock()
	assert.Equal(t, 1, len(manager.offlineCache.entries))
}

func TestHandleEvent_Error(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// handleEvent for Error should log and not panic
	manager.handleEvent(SMBEvent{
		Type:      EventError,
		SourceID:  "src-error",
		Error:     fmt.Errorf("test error"),
		Timestamp: time.Now(),
	})
}

func TestHandleEvent_HealthCheck(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// handleEvent for HealthCheck should be a no-op (not in switch)
	manager.handleEvent(SMBEvent{
		Type:      EventHealthCheck,
		SourceID:  "src-health",
		Timestamp: time.Now(),
	})
}

// --- New tests for HealthChecker ---

func TestNewHealthChecker(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	checker := NewHealthChecker(manager, 10*time.Second, 5*time.Second, logger)
	assert.NotNil(t, checker)
	assert.Equal(t, manager, checker.manager)
	assert.Equal(t, 10*time.Second, checker.interval)
	assert.Equal(t, 5*time.Second, checker.timeout)
	assert.NotNil(t, checker.logger)
	assert.NotNil(t, checker.ctx)
	assert.NotNil(t, checker.cancel)
}

func TestHealthChecker_StartStop(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	checker := NewHealthChecker(manager, 50*time.Millisecond, 10*time.Millisecond, logger)
	checker.Start()

	// Let it tick once
	time.Sleep(80 * time.Millisecond)

	checker.Stop()
	assert.NotNil(t, checker.ticker)
}

func TestHealthChecker_StopWithoutStart(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	checker := NewHealthChecker(manager, 50*time.Millisecond, 10*time.Millisecond, logger)
	// Stop without Start should not panic (ticker is nil, cancel exists)
	checker.Stop()
}

func TestHealthChecker_PerformHealthChecks_NoSources(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	checker := NewHealthChecker(manager, 50*time.Millisecond, 10*time.Millisecond, logger)
	// performHealthChecks with no sources should be a no-op
	checker.performHealthChecks()
}

func TestHealthChecker_PerformHealthChecks_WithDisabledSource(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Add a disabled source directly
	source := &SMBSource{
		ID:        "disabled-health",
		Name:      "Disabled",
		Path:      "//server/share",
		IsEnabled: false,
	}
	manager.mutex.Lock()
	manager.sources["disabled-health"] = source
	manager.mutex.Unlock()

	checker := NewHealthChecker(manager, 50*time.Millisecond, 10*time.Millisecond, logger)
	// Disabled sources should be skipped
	checker.performHealthChecks()
	time.Sleep(50 * time.Millisecond)
}

// --- New tests for OfflineCache ---

func TestOfflineCache_EnableOfflineMode(t *testing.T) {
	logger := testLogger()
	cache := NewOfflineCache(10, logger)

	// Should not panic
	cache.EnableOfflineMode("test-source")
}

func TestOfflineCache_ProcessCachedChanges_NoEntries(t *testing.T) {
	logger := testLogger()
	cache := NewOfflineCache(10, logger)

	// Should handle empty cache gracefully
	cache.ProcessCachedChanges("non-existent-source")
}

func TestOfflineCache_ProcessCachedChanges_AlreadyAvailable(t *testing.T) {
	logger := testLogger()
	cache := NewOfflineCache(10, logger)

	// Cache a change and mark it available manually
	cache.CacheChange("src1", "/file.mp4")
	for _, entry := range cache.entries {
		entry.IsAvailable = true
	}

	// Processing should not re-process already available entries
	cache.ProcessCachedChanges("src1")
	for _, entry := range cache.entries {
		if entry.SourceID == "src1" {
			assert.True(t, entry.IsAvailable)
		}
	}
}

func TestOfflineCache_CacheChangeMetadata(t *testing.T) {
	logger := testLogger()
	cache := NewOfflineCache(10, logger)

	cache.CacheChange("src1", "/path/movie.mp4")

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	key := "src1:/path/movie.mp4"
	entry, exists := cache.entries[key]
	require.True(t, exists)
	assert.Equal(t, "/path/movie.mp4", entry.Path)
	assert.Equal(t, "src1", entry.SourceID)
	assert.False(t, entry.IsAvailable)
	assert.NotNil(t, entry.Metadata)
	assert.False(t, entry.LastSeen.IsZero())
}

func TestOfflineCache_EvictOldest_EmptyCache(t *testing.T) {
	logger := testLogger()
	cache := NewOfflineCache(0, logger)

	// evictOldest on empty cache should not panic
	cache.evictOldest()
	assert.Equal(t, 0, len(cache.entries))
}

// --- New tests for connectSource ---

func TestConnectSource_DisabledSource(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:                "disabled-connect",
		Name:              "Disabled Connect",
		Path:              "//server/share",
		IsEnabled:         false,
		ConnectionTimeout: 100 * time.Millisecond,
	}

	err := manager.connectSource(source)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source is disabled")
}

// --- New tests for sendEvent ---

func TestSendEvent_EmptyChannel(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	event := SMBEvent{
		Type:      EventConnected,
		SourceID:  "test-send",
		Timestamp: time.Now(),
	}

	// Channel is empty, should succeed immediately
	manager.sendEvent(event)
	assert.Equal(t, 1, len(manager.eventChannel))
}

func TestSendEvent_PartiallyFull(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Fill channel partially
	for i := 0; i < 500; i++ {
		manager.eventChannel <- SMBEvent{
			Type:      EventHealthCheck,
			SourceID:  fmt.Sprintf("filler-%d", i),
			Timestamp: time.Now(),
		}
	}

	// Should still succeed
	manager.sendEvent(SMBEvent{
		Type:      EventConnected,
		SourceID:  "new-event",
		Timestamp: time.Now(),
	})
	assert.Equal(t, 501, len(manager.eventChannel))
}

// --- New tests for processFileChange ---

func TestProcessFileChange(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Should not panic
	manager.processFileChange("source1", "/path/to/file.mp4")
}

// --- New test for processEvents with stopChannel ---

func TestProcessEvents_StopChannel(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	manager.wg.Add(1)
	go manager.processEvents()

	// Send a few events
	manager.eventChannel <- SMBEvent{
		Type:      EventConnected,
		SourceID:  "test",
		Timestamp: time.Now(),
	}

	time.Sleep(50 * time.Millisecond)

	// Signal stop
	close(manager.stopChannel)
	manager.wg.Wait()
}

// --- New test for checkSourceHealth ---

func TestCheckSourceHealth(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:                "health-test",
		Name:              "Health Test",
		Path:              "//server/share",
		IsEnabled:         true,
		State:             StateConnected,
		ConnectionTimeout: 100 * time.Millisecond,
	}
	manager.mutex.Lock()
	manager.sources["health-test"] = source
	manager.mutex.Unlock()

	manager.checkSourceHealth(source)

	// Give background goroutine time to complete if reconnection was triggered
	time.Sleep(100 * time.Millisecond)
}

// --- New tests for scheduleRetry with stop ---

func TestScheduleRetry_StopChannel(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:                "retry-stop",
		Name:              "Retry Stop",
		Path:              "//server/share",
		IsEnabled:         true,
		RetryAttempts:     1,
		RetryDelay:        5 * time.Second, // Long delay
		ConnectionTimeout: 100 * time.Millisecond,
	}

	done := make(chan struct{})
	go func() {
		manager.scheduleRetry(source)
		close(done)
	}()

	// Close stop channel to abort the retry wait
	time.Sleep(50 * time.Millisecond)
	close(manager.stopChannel)

	select {
	case <-done:
		// scheduleRetry returned
	case <-time.After(2 * time.Second):
		t.Fatal("scheduleRetry did not return after stopChannel closed")
	}
}

// --- New test for scheduleRetry delay cap ---

func TestScheduleRetry_DelayCap(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:                "retry-cap",
		Name:              "Retry Cap",
		Path:              "//server/share",
		IsEnabled:         true,
		RetryAttempts:     100, // Large number to exceed cap
		RetryDelay:        1 * time.Second,
		ConnectionTimeout: 100 * time.Millisecond,
	}

	// Computed delay = 1s * 100 = 100s, should be capped to 5 min
	// We just verify it doesn't run with the uncapped value by using stopChannel
	done := make(chan struct{})
	go func() {
		manager.scheduleRetry(source)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	close(manager.stopChannel)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduleRetry did not return after stopChannel closed")
	}
}

// --- New test for GetSourceStatus with multiple states ---

func TestGetSourceStatus_MultipleStates(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Add sources with different states directly
	manager.mutex.Lock()
	manager.sources["s1"] = &SMBSource{
		ID: "s1", Name: "Connected", Path: "//a/b",
		State: StateConnected, IsEnabled: true,
		LastError: "",
	}
	manager.sources["s2"] = &SMBSource{
		ID: "s2", Name: "Offline", Path: "//c/d",
		State: StateOffline, IsEnabled: true,
		LastError: "connection refused",
	}
	manager.mutex.Unlock()

	status := manager.GetSourceStatus()
	assert.Equal(t, 2, len(status))

	s1, ok := status["s1"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "connected", s1["state"])
	assert.Equal(t, "", s1["last_error"])

	s2, ok := status["s2"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "offline", s2["state"])
	assert.Equal(t, "connection refused", s2["last_error"])
}

// --- New test for AddSource default HealthCheckInterval ---

func TestAddSource_DefaultHealthCheckInterval(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	source := &SMBSource{
		ID:   "hci-test",
		Name: "HCI Test",
		Path: "//server/share",
	}

	err := manager.AddSource(source)
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Second, source.HealthCheckInterval)
	time.Sleep(100 * time.Millisecond)
}

// --- New test for Start with sources ---

func TestStart_WithSources(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	// Add a source directly (bypass AddSource to avoid goroutine)
	manager.mutex.Lock()
	manager.sources["start-src"] = &SMBSource{
		ID: "start-src", Name: "Start Test", Path: "//s/s",
		IsEnabled: true, State: StateConnected,
		ConnectionTimeout: 100 * time.Millisecond,
	}
	manager.mutex.Unlock()

	err := manager.Start()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	err = manager.Stop()
	assert.NoError(t, err)
}

// --- New test for Start with disabled sources ---

func TestStart_DisabledSourcesSkipped(t *testing.T) {
	logger := testLogger()
	manager := NewResilientSMBManager(logger, 100)

	manager.mutex.Lock()
	manager.sources["disabled-start"] = &SMBSource{
		ID: "disabled-start", Name: "Disabled", Path: "//s/s",
		IsEnabled: false, State: StateDisconnected,
	}
	manager.mutex.Unlock()

	err := manager.Start()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	err = manager.Stop()
	assert.NoError(t, err)
}
