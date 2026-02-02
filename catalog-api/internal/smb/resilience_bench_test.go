package smb

import (
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newBenchManager(chanSize int) *ResilientSMBManager {
	logger, _ := zap.NewProduction()
	return &ResilientSMBManager{
		sources:      make(map[string]*SMBSource),
		logger:       logger,
		offlineCache: NewOfflineCache(1000, logger),
		eventChannel: make(chan SMBEvent, chanSize),
		stopChannel:  make(chan struct{}),
		startTime:    time.Now(),
	}
}

// --- sendEvent benchmarks ---

func BenchmarkSendEvent(b *testing.B) {
	benchmarks := []struct {
		name     string
		chanSize int
		prefill  int // how many events to pre-fill the channel with
	}{
		{"empty_channel", 1000, 0},
		{"half_full_channel", 1000, 500},
		{"nearly_full_channel", 1000, 999},
		{"full_channel", 1000, 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			manager := newBenchManager(bm.chanSize)

			// Pre-fill the channel
			for i := 0; i < bm.prefill; i++ {
				manager.eventChannel <- SMBEvent{
					Type:      EventHealthCheck,
					SourceID:  "filler",
					Timestamp: time.Now(),
				}
			}

			event := SMBEvent{
				Type:      EventConnected,
				SourceID:  "bench-source",
				Timestamp: time.Now(),
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				manager.sendEvent(event)
				// Drain one event to keep channel from blocking indefinitely
				select {
				case <-manager.eventChannel:
				default:
				}
			}
		})
	}
}

func BenchmarkSendEvent_FullChannel_Contention(b *testing.B) {
	// Benchmark sendEvent with a full channel (backpressure path) without draining
	// Each iteration creates a fresh full channel to isolate the backpressure path
	logger, _ := zap.NewProduction()

	event := SMBEvent{
		Type:      EventConnected,
		SourceID:  "bench-source",
		Timestamp: time.Now(),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := &ResilientSMBManager{
			sources:      make(map[string]*SMBSource),
			logger:       logger,
			offlineCache: NewOfflineCache(10, logger),
			eventChannel: make(chan SMBEvent, 10),
			stopChannel:  make(chan struct{}),
		}
		// Fill channel
		for j := 0; j < 10; j++ {
			manager.eventChannel <- SMBEvent{
				Type:      EventHealthCheck,
				SourceID:  "filler",
				Timestamp: time.Now(),
			}
		}
		manager.sendEvent(event)
	}
}

// --- OfflineCache benchmarks ---

func BenchmarkOfflineCache_CacheChange(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, maxSize := range sizes {
		b.Run(fmt.Sprintf("maxSize=%d", maxSize), func(b *testing.B) {
			logger, _ := zap.NewProduction()
			cache := NewOfflineCache(maxSize, logger)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				cache.CacheChange("source1", fmt.Sprintf("/path/file_%d.mp4", i))
			}
		})
	}
}

func BenchmarkOfflineCache_CacheChange_Eviction(b *testing.B) {
	// Cache is always at capacity, so every insert triggers eviction
	logger, _ := zap.NewProduction()
	cache := NewOfflineCache(100, logger)

	// Pre-fill to capacity
	for i := 0; i < 100; i++ {
		cache.CacheChange("source1", fmt.Sprintf("/prefill/file_%d.mp4", i))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cache.CacheChange("source1", fmt.Sprintf("/new/file_%d.mp4", i))
	}
}

// --- GetSourceStatus benchmark ---

func BenchmarkGetSourceStatus(b *testing.B) {
	counts := []int{1, 10, 50}
	for _, count := range counts {
		b.Run(fmt.Sprintf("sources=%d", count), func(b *testing.B) {
			manager := newBenchManager(100)

			for i := 0; i < count; i++ {
				id := fmt.Sprintf("source-%d", i)
				manager.sources[id] = &SMBSource{
					ID:        id,
					Name:      fmt.Sprintf("Source %d", i),
					Path:      fmt.Sprintf("//server/share%d", i),
					State:     StateConnected,
					IsEnabled: true,
				}
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = manager.GetSourceStatus()
			}
		})
	}
}
