package monitoring

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// RESOURCE MONITOR TEST: CPU Usage
// =============================================================================

func TestResourceMonitor_CPUUsage(t *testing.T) {
	tests := []struct {
		name         string
		maxLoadAvg1  float64
		maxLoadAvg5  float64
		maxLoadAvg15 float64
	}{
		{
			name:         "LoadAvg1Min_UnderThreshold",
			maxLoadAvg1:  float64(runtime.NumCPU()) * 5.0,
			maxLoadAvg5:  float64(runtime.NumCPU()) * 4.0,
			maxLoadAvg15: float64(runtime.NumCPU()) * 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			data, err := os.ReadFile("/proc/loadavg")
			if err != nil {
				t.Skipf("Skipping CPU usage test: /proc/loadavg not available: %v", err)
			}

			fields := strings.Fields(string(data))
			require.GreaterOrEqual(t, len(fields), 3,
				"/proc/loadavg should have at least 3 fields, got: %s", string(data))

			loadAvg1, err := strconv.ParseFloat(fields[0], 64)
			require.NoError(t, err, "Failed to parse 1-min load average")

			loadAvg5, err := strconv.ParseFloat(fields[1], 64)
			require.NoError(t, err, "Failed to parse 5-min load average")

			loadAvg15, err := strconv.ParseFloat(fields[2], 64)
			require.NoError(t, err, "Failed to parse 15-min load average")

			numCPU := runtime.NumCPU()
			t.Logf("CPUs: %d", numCPU)
			t.Logf("Load average (1m/5m/15m): %.2f / %.2f / %.2f", loadAvg1, loadAvg5, loadAvg15)
			t.Logf("Per-CPU load (1m/5m/15m): %.2f / %.2f / %.2f",
				loadAvg1/float64(numCPU), loadAvg5/float64(numCPU), loadAvg15/float64(numCPU))

			assert.LessOrEqual(t, loadAvg1, tt.maxLoadAvg1,
				"1-min load average (%.2f) should be under threshold (%.2f)",
				loadAvg1, tt.maxLoadAvg1)

			assert.LessOrEqual(t, loadAvg5, tt.maxLoadAvg5,
				"5-min load average (%.2f) should be under threshold (%.2f)",
				loadAvg5, tt.maxLoadAvg5)

			assert.LessOrEqual(t, loadAvg15, tt.maxLoadAvg15,
				"15-min load average (%.2f) should be under threshold (%.2f)",
				loadAvg15, tt.maxLoadAvg15)
		})
	}
}

// =============================================================================
// RESOURCE MONITOR TEST: Memory Usage
// =============================================================================

func TestResourceMonitor_MemoryUsage(t *testing.T) {
	tests := []struct {
		name               string
		maxHeapAllocMB     uint64
		maxSysMB           uint64
		maxHeapObjectCount uint64
	}{
		{
			name:               "HeapAlloc_WithinBounds",
			maxHeapAllocMB:     512,
			maxSysMB:           1024,
			maxHeapObjectCount: 5000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			// Run GC first to get a clean measurement
			runtime.GC()
			time.Sleep(50 * time.Millisecond)

			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			heapAllocMB := memStats.HeapAlloc / (1024 * 1024)
			sysMB := memStats.Sys / (1024 * 1024)

			t.Logf("=== Memory Stats ===")
			t.Logf("HeapAlloc:     %d MB", heapAllocMB)
			t.Logf("HeapSys:       %d MB", memStats.HeapSys/(1024*1024))
			t.Logf("HeapInuse:     %d MB", memStats.HeapInuse/(1024*1024))
			t.Logf("HeapIdle:      %d MB", memStats.HeapIdle/(1024*1024))
			t.Logf("HeapReleased:  %d MB", memStats.HeapReleased/(1024*1024))
			t.Logf("HeapObjects:   %d", memStats.HeapObjects)
			t.Logf("StackInuse:    %d KB", memStats.StackInuse/1024)
			t.Logf("Sys (total):   %d MB", sysMB)
			t.Logf("NumGC:         %d", memStats.NumGC)
			t.Logf("GCCPUFraction: %.6f", memStats.GCCPUFraction)

			assert.LessOrEqual(t, heapAllocMB, tt.maxHeapAllocMB,
				"HeapAlloc (%d MB) should be within bounds (%d MB)",
				heapAllocMB, tt.maxHeapAllocMB)

			assert.LessOrEqual(t, sysMB, tt.maxSysMB,
				"Sys memory (%d MB) should be within bounds (%d MB)",
				sysMB, tt.maxSysMB)

			assert.LessOrEqual(t, memStats.HeapObjects, tt.maxHeapObjectCount,
				"Heap object count (%d) should be within bounds (%d)",
				memStats.HeapObjects, tt.maxHeapObjectCount)
		})
	}
}

// =============================================================================
// RESOURCE MONITOR TEST: Goroutine Count
// =============================================================================

func TestResourceMonitor_GoroutineCount(t *testing.T) {
	tests := []struct {
		name            string
		spawnGoroutines int
		maxBaseline     int
		maxDuringWork   int
		maxAfterCleanup int
	}{
		{
			name:            "BaselineGoroutineCount",
			spawnGoroutines: 0,
			maxBaseline:     100,
			maxDuringWork:   100,
			maxAfterCleanup: 100,
		},
		{
			name:            "AfterSpawning50Goroutines",
			spawnGoroutines: 50,
			maxBaseline:     100,
			maxDuringWork:   200,
			maxAfterCleanup: 120,
		},
		{
			name:            "AfterSpawning200Goroutines",
			spawnGoroutines: 200,
			maxBaseline:     100,
			maxDuringWork:   400,
			maxAfterCleanup: 120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			baseline := runtime.NumGoroutine()
			t.Logf("Baseline goroutines: %d", baseline)

			assert.LessOrEqual(t, baseline, tt.maxBaseline,
				"Baseline goroutine count (%d) should be under threshold (%d)",
				baseline, tt.maxBaseline)

			if tt.spawnGoroutines == 0 {
				return
			}

			// Spawn goroutines with controlled lifetime
			done := make(chan struct{})
			var wg sync.WaitGroup

			for i := 0; i < tt.spawnGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-done
				}()
			}

			// Measure during active goroutines
			time.Sleep(50 * time.Millisecond)
			duringWork := runtime.NumGoroutine()
			t.Logf("During work goroutines: %d (spawned: %d)", duringWork, tt.spawnGoroutines)

			assert.LessOrEqual(t, duringWork, tt.maxDuringWork,
				"Goroutine count during work (%d) should be under threshold (%d)",
				duringWork, tt.maxDuringWork)

			// Signal goroutines to exit
			close(done)
			wg.Wait()

			// Allow goroutines to fully exit
			time.Sleep(100 * time.Millisecond)

			afterCleanup := runtime.NumGoroutine()
			t.Logf("After cleanup goroutines: %d", afterCleanup)

			assert.LessOrEqual(t, afterCleanup, tt.maxAfterCleanup,
				"Goroutine count after cleanup (%d) should return to near baseline (%d max)",
				afterCleanup, tt.maxAfterCleanup)
		})
	}
}

// =============================================================================
// RESOURCE MONITOR TEST: Open File Descriptors
// =============================================================================

func TestResourceMonitor_OpenFileDescriptors(t *testing.T) {
	tests := []struct {
		name   string
		maxFDs int
	}{
		{
			name:   "FileDescriptors_WithinBounds",
			maxFDs: 4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			fdDir := "/proc/self/fd"
			entries, err := os.ReadDir(fdDir)
			if err != nil {
				t.Skipf("Skipping file descriptor test: %s not available: %v", fdDir, err)
			}

			openFDs := len(entries)

			// Count FD types for additional insight
			var regularFiles, sockets, pipes, other int
			for _, entry := range entries {
				linkPath, err := os.Readlink(filepath.Join(fdDir, entry.Name()))
				if err != nil {
					other++
					continue
				}
				switch {
				case strings.HasPrefix(linkPath, "socket:"):
					sockets++
				case strings.HasPrefix(linkPath, "pipe:"):
					pipes++
				case strings.HasPrefix(linkPath, "/"):
					regularFiles++
				default:
					other++
				}
			}

			// Read the FD limit
			var softLimit, hardLimit string
			limitsData, err := os.ReadFile("/proc/self/limits")
			if err == nil {
				for _, line := range strings.Split(string(limitsData), "\n") {
					if strings.Contains(line, "Max open files") {
						fields := strings.Fields(line)
						if len(fields) >= 5 {
							softLimit = fields[3]
							hardLimit = fields[4]
						}
						break
					}
				}
			}

			t.Logf("=== File Descriptor Stats ===")
			t.Logf("Open FDs:      %d", openFDs)
			t.Logf("  Regular files: %d", regularFiles)
			t.Logf("  Sockets:       %d", sockets)
			t.Logf("  Pipes:         %d", pipes)
			t.Logf("  Other:         %d", other)
			t.Logf("Soft limit:    %s", softLimit)
			t.Logf("Hard limit:    %s", hardLimit)

			assert.LessOrEqual(t, openFDs, tt.maxFDs,
				"Open file descriptor count (%d) should be under threshold (%d)",
				openFDs, tt.maxFDs)

			// Verify FD count is reasonable for a test process
			assert.Greater(t, openFDs, 0,
				"Process should have at least some open file descriptors")

			// If soft limit is parseable, verify we are well under it
			if softLimit != "" && softLimit != "unlimited" {
				limit, err := strconv.Atoi(softLimit)
				if err == nil {
					usagePercent := float64(openFDs) / float64(limit) * 100
					t.Logf("FD usage:      %.1f%% of soft limit", usagePercent)
					assert.Less(t, usagePercent, 50.0,
						"FD usage (%.1f%%) should be under 50%% of soft limit", usagePercent)
				}
			}
		})
	}
}
