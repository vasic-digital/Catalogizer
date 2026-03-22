package utils

import (
	"sync"
	"testing"
)

func TestNewBufferPool(t *testing.T) {
	pool := NewBufferPool()
	if pool == nil {
		t.Fatal("expected pool to be non-nil")
	}

	if len(pool.pools) == 0 {
		t.Error("expected pools to be initialized")
	}

	if len(pool.sizes) == 0 {
		t.Error("expected sizes to be initialized")
	}
}

func TestBufferPool_Get(t *testing.T) {
	pool := NewBufferPool()

	tests := []struct {
		name    string
		size    int
		wantLen int
		wantCap int
	}{
		{"1KB", 512, 512, 1024},
		{"4KB exact", 4096, 4096, 4096},
		{"16KB", 10000, 10000, 16384},
		{"64KB", 50000, 50000, 65536},
		{"256KB", 200000, 200000, 262144},
		{"1MB", 500000, 500000, 1048576},
		{"oversized", 2000000, 2000000, 2000000}, // Should allocate directly
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := pool.Get(tt.size)
			if len(buf) != tt.wantLen {
				t.Errorf("Get(%d) length = %d, want %d", tt.size, len(buf), tt.wantLen)
			}
			if cap(buf) < tt.wantCap {
				t.Errorf("Get(%d) capacity = %d, want at least %d", tt.size, cap(buf), tt.wantCap)
			}
			pool.Put(buf)
		})
	}
}

func TestBufferPool_Get_ZeroSize(t *testing.T) {
	pool := NewBufferPool()
	buf := pool.Get(0)
	if buf != nil {
		t.Error("Get(0) should return nil")
	}
}

func TestBufferPool_Get_NegativeSize(t *testing.T) {
	pool := NewBufferPool()
	buf := pool.Get(-1)
	if buf != nil {
		t.Error("Get(-1) should return nil")
	}
}

func TestBufferPool_Put(t *testing.T) {
	pool := NewBufferPool()

	// Get and put a buffer
	buf := pool.Get(1024)
	pool.Put(buf)

	// Get again - should reuse
	buf2 := pool.Get(1024)
	pool.Put(buf2)

	stats := pool.Stats()
	if stats.Hits < 1 {
		t.Error("expected at least one hit after Put and Get")
	}
}

func TestBufferPool_Put_Nil(t *testing.T) {
	pool := NewBufferPool()
	// Should not panic
	pool.Put(nil)
}

func TestBufferPool_Put_Empty(t *testing.T) {
	pool := NewBufferPool()
	// Should not panic
	pool.Put([]byte{})
}

func TestBufferPool_Stats(t *testing.T) {
	pool := NewBufferPool()
	pool.Reset()

	// Do some operations
	for i := 0; i < 10; i++ {
		buf := pool.Get(1024)
		pool.Put(buf)
	}

	// Get an oversized buffer (miss)
	_ = pool.Get(2000000)

	stats := pool.Stats()

	if stats.Hits != 10 {
		t.Errorf("expected 10 hits, got %d", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}

	if stats.TotalGets != 11 {
		t.Errorf("expected 11 total gets, got %d", stats.TotalGets)
	}
}

func TestBufferPoolStats_HitRate(t *testing.T) {
	tests := []struct {
		name     string
		hits     int64
		misses   int64
		expected float64
	}{
		{"100% hit rate", 100, 0, 100.0},
		{"50% hit rate", 50, 50, 50.0},
		{"0% hit rate", 0, 100, 0.0},
		{"empty", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := BufferPoolStats{
				Hits:      tt.hits,
				Misses:    tt.misses,
				TotalGets: tt.hits + tt.misses,
			}

			if stats.HitRate() != tt.expected {
				t.Errorf("HitRate() = %f, want %f", stats.HitRate(), tt.expected)
			}
		})
	}
}

func TestBufferPool_Reset(t *testing.T) {
	pool := NewBufferPool()

	// Do some operations
	for i := 0; i < 5; i++ {
		buf := pool.Get(1024)
		pool.Put(buf)
	}

	// Reset
	pool.Reset()

	// Check stats are cleared
	stats := pool.Stats()
	if stats.Hits != 0 {
		t.Errorf("expected hits to be 0 after reset, got %d", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("expected misses to be 0 after reset, got %d", stats.Misses)
	}
}

func TestBufferPool_ConcurrentAccess(t *testing.T) {
	pool := NewBufferPool()

	const numGoroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				// Vary sizes
				size := 1024 * ((id+j)%6 + 1)
				buf := pool.Get(size)

				// Simulate work
				for k := 0; k < len(buf) && k < 100; k++ {
					buf[k] = byte(k)
				}

				pool.Put(buf)
			}
		}(i)
	}

	wg.Wait()

	stats := pool.Stats()
	if stats.TotalGets != int64(numGoroutines*iterations) {
		t.Errorf("expected %d total gets, got %d", numGoroutines*iterations, stats.TotalGets)
	}
}

func TestGetBuffer_PutBuffer(t *testing.T) {
	// Reset default pool
	defaultPool.Reset()

	// Test global functions
	buf := GetBuffer(1024)
	if buf == nil {
		t.Fatal("GetBuffer should return non-nil buffer")
	}

	PutBuffer(buf)

	stats := GetPoolStats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
}

func BenchmarkBufferPool_Get_Put(b *testing.B) {
	pool := NewBufferPool()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get(4096)
			pool.Put(buf)
		}
	})
}

func BenchmarkBufferPool_Get_NoPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = make([]byte, 4096)
		}
	})
}

func TestBufferPool_MemoryReuse(t *testing.T) {
	pool := NewBufferPool()
	pool.Reset()

	// Get a buffer
	buf1 := pool.Get(1024)
	ptr1 := &buf1[0]
	pool.Put(buf1)

	// Get same size buffer again
	buf2 := pool.Get(1024)
	ptr2 := &buf2[0]
	pool.Put(buf2)

	// Should be the same underlying memory
	if ptr1 != ptr2 {
		t.Log("Buffers were not reused (this is ok, pool may have allocated new)")
	}

	stats := pool.Stats()
	if stats.Hits == 0 {
		t.Error("expected at least one buffer reuse (hit)")
	}
}

func TestBufferPool_BufferZeroing(t *testing.T) {
	pool := NewBufferPool()

	// Get a buffer and write data
	buf := pool.Get(1024)
	for i := range buf {
		buf[i] = 0xFF
	}
	pool.Put(buf)

	// Get buffer again
	buf2 := pool.Get(1024)

	// Check it's zeroed (first few bytes)
	for i := 0; i < 100 && i < len(buf2); i++ {
		if buf2[i] != 0 {
			t.Error("buffer was not zeroed after Put")
			break
		}
	}
	pool.Put(buf2)
}
