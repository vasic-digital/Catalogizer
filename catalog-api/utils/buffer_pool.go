package utils

import (
	"sync"
	"sync/atomic"
)

// BufferPool provides a pool of reusable byte slices to reduce GC pressure
type BufferPool struct {
	pools     map[int]*sync.Pool
	sizes     []int
	maxSize   int
	hits      atomic.Int64
	misses    atomic.Int64
	allocated atomic.Int64
	released  atomic.Int64
}

// NewBufferPool creates a new buffer pool with predefined sizes
func NewBufferPool() *BufferPool {
	sizes := []int{
		1024,    // 1 KB - small metadata
		4096,    // 4 KB - standard page size
		16384,   // 16 KB - small files
		65536,   // 64 KB - medium files
		262144,  // 256 KB - large files
		1048576, // 1 MB - very large files
	}

	pools := make(map[int]*sync.Pool, len(sizes))
	for _, size := range sizes {
		size := size // capture range variable
		pools[size] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		}
	}

	return &BufferPool{
		pools:   pools,
		sizes:   sizes,
		maxSize: sizes[len(sizes)-1],
	}
}

// Get retrieves a buffer of at least the requested size
func (bp *BufferPool) Get(size int) []byte {
	if size <= 0 {
		return nil
	}

	// Find appropriate pool size
	poolSize := bp.findPoolSize(size)
	if poolSize == 0 {
		// Size too large, allocate directly
		bp.misses.Add(1)
		bp.allocated.Add(int64(size))
		return make([]byte, size)
	}

	// Get from pool
	pool := bp.pools[poolSize]
	buf := pool.Get().([]byte)
	bp.hits.Add(1)

	// Return slice with exact requested size (may be smaller than pool size)
	if len(buf) > size {
		return buf[:size]
	}
	return buf
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf []byte) {
	if buf == nil || cap(buf) == 0 {
		return
	}

	// Find appropriate pool for this buffer
	cap := cap(buf)
	poolSize := bp.findPoolSize(cap)
	if poolSize == 0 || cap != poolSize {
		// Buffer wasn't from pool or wrong size, let GC collect it
		bp.released.Add(int64(cap))
		return
	}

	// Reset buffer before returning to pool
	buf = buf[:cap]
	for i := range buf {
		buf[i] = 0
	}

	pool := bp.pools[poolSize]
	pool.Put(buf)
}

// findPoolSize finds the smallest pool size that can accommodate the request
func (bp *BufferPool) findPoolSize(size int) int {
	for _, poolSize := range bp.sizes {
		if poolSize >= size {
			return poolSize
		}
	}
	return 0 // Size exceeds max pool size
}

// Stats returns pool statistics
func (bp *BufferPool) Stats() BufferPoolStats {
	return BufferPoolStats{
		Hits:      bp.hits.Load(),
		Misses:    bp.misses.Load(),
		Allocated: bp.allocated.Load(),
		Released:  bp.released.Load(),
		TotalGets: bp.hits.Load() + bp.misses.Load(),
	}
}

// Reset clears all pools and statistics
func (bp *BufferPool) Reset() {
	for _, pool := range bp.pools {
		for {
			if pool.Get() == nil {
				break
			}
		}
	}
	bp.hits.Store(0)
	bp.misses.Store(0)
	bp.allocated.Store(0)
	bp.released.Store(0)
}

// BufferPoolStats contains pool statistics
type BufferPoolStats struct {
	Hits      int64
	Misses    int64
	Allocated int64
	Released  int64
	TotalGets int64
}

// HitRate returns the cache hit rate as a percentage
func (s BufferPoolStats) HitRate() float64 {
	if s.TotalGets == 0 {
		return 0.0
	}
	return float64(s.Hits) / float64(s.TotalGets) * 100.0
}

// Global pool instance
var defaultPool = NewBufferPool()

// GetBuffer gets a buffer from the default pool
func GetBuffer(size int) []byte {
	return defaultPool.Get(size)
}

// PutBuffer returns a buffer to the default pool
func PutBuffer(buf []byte) {
	defaultPool.Put(buf)
}

// GetPoolStats returns statistics for the default pool
func GetPoolStats() BufferPoolStats {
	return defaultPool.Stats()
}
