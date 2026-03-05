package stress

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// STRESS TEST: Memory Pressure - Allocation Storm
// =============================================================================

func TestMemoryPressure_AllocationStorm(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tests := []struct {
		name             string
		goroutines       int
		allocsPerRoutine int
		allocSizeBytes   int
		maxHeapGrowthMB  uint64
	}{
		{
			name:             "SmallAllocations_50Goroutines",
			goroutines:       50,
			allocsPerRoutine: 1000,
			allocSizeBytes:   1024, // 1 KB
			maxHeapGrowthMB:  200,
		},
		{
			name:             "MediumAllocations_20Goroutines",
			goroutines:       20,
			allocsPerRoutine: 500,
			allocSizeBytes:   64 * 1024, // 64 KB
			maxHeapGrowthMB:  300,
		},
		{
			name:             "LargeAllocations_10Goroutines",
			goroutines:       10,
			allocsPerRoutine: 100,
			allocSizeBytes:   512 * 1024, // 512 KB
			maxHeapGrowthMB:  400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			// Force GC and get baseline
			runtime.GC()
			time.Sleep(50 * time.Millisecond)

			var baselineStats runtime.MemStats
			runtime.ReadMemStats(&baselineStats)
			baselineHeap := baselineStats.HeapAlloc

			var wg sync.WaitGroup
			var totalAllocated int64

			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for j := 0; j < tt.allocsPerRoutine; j++ {
						// Allocate memory
						buf := make([]byte, tt.allocSizeBytes)
						// Touch the memory to ensure it is actually allocated
						buf[0] = byte(j)
						buf[len(buf)-1] = byte(j)
						atomic.AddInt64(&totalAllocated, int64(tt.allocSizeBytes))

						// Let GC reclaim periodically
						if j%100 == 0 {
							runtime.Gosched()
						}
					}
				}()
			}

			wg.Wait()

			// Force GC to reclaim released memory
			runtime.GC()
			time.Sleep(100 * time.Millisecond)
			runtime.GC()

			var afterStats runtime.MemStats
			runtime.ReadMemStats(&afterStats)

			heapGrowth := uint64(0)
			if afterStats.HeapAlloc > baselineHeap {
				heapGrowth = afterStats.HeapAlloc - baselineHeap
			}
			heapGrowthMB := heapGrowth / (1024 * 1024)

			t.Logf("Baseline heap: %d MB", baselineHeap/(1024*1024))
			t.Logf("After heap:    %d MB", afterStats.HeapAlloc/(1024*1024))
			t.Logf("Heap growth:   %d MB", heapGrowthMB)
			t.Logf("Total allocated (cumulative): %d MB", totalAllocated/(1024*1024))
			t.Logf("GC cycles: %d", afterStats.NumGC-baselineStats.NumGC)

			assert.LessOrEqual(t, heapGrowthMB, tt.maxHeapGrowthMB,
				"Heap should not grow unbounded after GC; growth was %d MB (max %d MB)",
				heapGrowthMB, tt.maxHeapGrowthMB)
		})
	}
}

// =============================================================================
// STRESS TEST: Memory Pressure - Goroutine Limit
// =============================================================================

func TestMemoryPressure_GoroutineLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tests := []struct {
		name             string
		spawnCount       int
		maxGoroutines    int
		workDuration     time.Duration
	}{
		{
			name:          "100BoundedGoroutines",
			spawnCount:    500,
			maxGoroutines: 100,
			workDuration:  10 * time.Millisecond,
		},
		{
			name:          "50BoundedGoroutines",
			spawnCount:    300,
			maxGoroutines: 50,
			workDuration:  20 * time.Millisecond,
		},
		{
			name:          "200BoundedGoroutines",
			spawnCount:    1000,
			maxGoroutines: 200,
			workDuration:  5 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			baseGoroutines := runtime.NumGoroutine()
			sem := make(chan struct{}, tt.maxGoroutines)
			var activeCount int64
			var peakCount int64

			var wg sync.WaitGroup
			for i := 0; i < tt.spawnCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					// Acquire semaphore slot
					sem <- struct{}{}
					defer func() { <-sem }()

					current := atomic.AddInt64(&activeCount, 1)
					defer atomic.AddInt64(&activeCount, -1)

					// Track peak
					for {
						old := atomic.LoadInt64(&peakCount)
						if current <= old || atomic.CompareAndSwapInt64(&peakCount, old, current) {
							break
						}
					}

					// Simulate work
					time.Sleep(tt.workDuration)
				}()
			}

			// Sample goroutine count during execution
			var maxObservedGoroutines int
			done := make(chan struct{})
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						count := runtime.NumGoroutine()
						if count > maxObservedGoroutines {
							maxObservedGoroutines = count
						}
						time.Sleep(5 * time.Millisecond)
					}
				}
			}()

			wg.Wait()
			close(done)

			// Allow goroutines to settle
			time.Sleep(100 * time.Millisecond)
			finalGoroutines := runtime.NumGoroutine()

			t.Logf("Base goroutines:  %d", baseGoroutines)
			t.Logf("Peak active:      %d", atomic.LoadInt64(&peakCount))
			t.Logf("Max observed:     %d", maxObservedGoroutines)
			t.Logf("Final goroutines: %d", finalGoroutines)

			assert.LessOrEqual(t, atomic.LoadInt64(&peakCount), int64(tt.maxGoroutines),
				"Peak active goroutines (%d) should not exceed limit (%d)",
				atomic.LoadInt64(&peakCount), tt.maxGoroutines)

			// Final goroutine count should be close to baseline (allow some variance for
			// test infrastructure goroutines)
			assert.InDelta(t, baseGoroutines, finalGoroutines, 20,
				"Goroutines should return to near baseline after completion")
		})
	}
}

// =============================================================================
// STRESS TEST: Memory Pressure - Cache Eviction
// =============================================================================

func TestMemoryPressure_CacheEviction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tests := []struct {
		name           string
		cacheEntries   int
		entrySize      int
		maxCacheLimit  int
		maxMemoryMB    uint64
	}{
		{
			name:          "SmallCache_1000Entries",
			cacheEntries:  5000,
			entrySize:     1024, // 1 KB per entry
			maxCacheLimit: 1000,
			maxMemoryMB:   100,
		},
		{
			name:          "MediumCache_500Entries",
			cacheEntries:  2000,
			entrySize:     16 * 1024, // 16 KB per entry
			maxCacheLimit: 500,
			maxMemoryMB:   100,
		},
		{
			name:          "LargeCache_200Entries",
			cacheEntries:  1000,
			entrySize:     64 * 1024, // 64 KB per entry
			maxCacheLimit: 200,
			maxMemoryMB:   150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			// Force GC and measure baseline
			runtime.GC()
			time.Sleep(50 * time.Millisecond)

			var baselineStats runtime.MemStats
			runtime.ReadMemStats(&baselineStats)

			// Simulate a bounded cache (map with eviction)
			type cacheEntry struct {
				key  string
				data []byte
			}

			var mu sync.RWMutex
			cache := make(map[string]*cacheEntry)
			evictionOrder := make([]string, 0, tt.maxCacheLimit)

			evict := func() {
				// Evict oldest entry
				if len(evictionOrder) > 0 {
					oldest := evictionOrder[0]
					evictionOrder = evictionOrder[1:]
					delete(cache, oldest)
				}
			}

			set := func(key string, data []byte) {
				mu.Lock()
				defer mu.Unlock()

				// Evict if at capacity
				for len(cache) >= tt.maxCacheLimit {
					evict()
				}

				entry := &cacheEntry{key: key, data: data}
				cache[key] = entry
				evictionOrder = append(evictionOrder, key)
			}

			// Fill cache with concurrent writes
			var wg sync.WaitGroup
			concurrentWriters := 10
			entriesPerWriter := tt.cacheEntries / concurrentWriters

			for w := 0; w < concurrentWriters; w++ {
				wg.Add(1)
				go func(writerID int) {
					defer wg.Done()
					for i := 0; i < entriesPerWriter; i++ {
						key := string(rune(writerID*entriesPerWriter + i))
						data := make([]byte, tt.entrySize)
						data[0] = byte(i)
						set(key, data)
					}
				}(w)
			}

			wg.Wait()

			// Measure memory with cache populated
			runtime.GC()
			time.Sleep(50 * time.Millisecond)

			var populatedStats runtime.MemStats
			runtime.ReadMemStats(&populatedStats)

			mu.RLock()
			cacheSize := len(cache)
			mu.RUnlock()

			// Verify cache size is bounded
			require.LessOrEqual(t, cacheSize, tt.maxCacheLimit,
				"Cache size (%d) should not exceed limit (%d)", cacheSize, tt.maxCacheLimit)

			populatedHeapMB := populatedStats.HeapAlloc / (1024 * 1024)
			baselineHeapMB := baselineStats.HeapAlloc / (1024 * 1024)

			t.Logf("Baseline heap:      %d MB", baselineHeapMB)
			t.Logf("Populated heap:     %d MB", populatedHeapMB)
			t.Logf("Cache entries:      %d (limit: %d)", cacheSize, tt.maxCacheLimit)
			t.Logf("Total entries written: %d", tt.cacheEntries)
			t.Logf("Entry size:         %d bytes", tt.entrySize)

			// Clear cache and verify memory is reclaimed
			mu.Lock()
			cache = nil
			evictionOrder = nil
			mu.Unlock()

			runtime.GC()
			time.Sleep(100 * time.Millisecond)
			runtime.GC()

			var afterClearStats runtime.MemStats
			runtime.ReadMemStats(&afterClearStats)

			afterClearHeapMB := afterClearStats.HeapAlloc / (1024 * 1024)
			t.Logf("After clear heap:   %d MB", afterClearHeapMB)

			// Memory should be bounded while cache is populated
			heapGrowth := uint64(0)
			if populatedStats.HeapAlloc > baselineStats.HeapAlloc {
				heapGrowth = (populatedStats.HeapAlloc - baselineStats.HeapAlloc) / (1024 * 1024)
			}
			assert.LessOrEqual(t, heapGrowth, tt.maxMemoryMB,
				"Memory should stay bounded with cache eviction; growth was %d MB (max %d MB)",
				heapGrowth, tt.maxMemoryMB)
		})
	}
}
