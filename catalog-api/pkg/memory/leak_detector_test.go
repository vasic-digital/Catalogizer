package memory

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLeakDetector(t *testing.T) {
	detector := NewLeakDetector(time.Second, 2.0)

	assert.NotNil(t, detector)
	assert.Equal(t, time.Second, detector.interval)
	assert.Equal(t, 2.0, detector.thresholdRatio)
	assert.False(t, detector.running)
}

func TestLeakDetector_StartStop(t *testing.T) {
	detector := NewLeakDetector(100*time.Millisecond, 2.0)

	ctx := context.Background()
	err := detector.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, detector.running)

	time.Sleep(150 * time.Millisecond)

	detector.Stop()
	assert.False(t, detector.running)
}

func TestLeakDetector_DoubleStart(t *testing.T) {
	detector := NewLeakDetector(time.Second, 2.0)

	ctx := context.Background()
	err := detector.Start(ctx)
	assert.NoError(t, err)

	err = detector.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	detector.Stop()
}

func TestLeakDetector_GetReport(t *testing.T) {
	detector := NewLeakDetector(time.Second, 2.0)

	report := detector.GetReport()

	assert.NotZero(t, report.Timestamp)
	assert.Greater(t, report.HeapAlloc, uint64(0))
	assert.Greater(t, report.GoroutineCount, 0)
}

func TestLeakDetector_GetSamples(t *testing.T) {
	detector := NewLeakDetector(50*time.Millisecond, 2.0)

	ctx := context.Background()
	err := detector.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(120 * time.Millisecond)

	detector.Stop()

	samples := detector.GetSamples()
	assert.GreaterOrEqual(t, len(samples), 1)
}

func TestLeakDetector_PotentialLeakDetection(t *testing.T) {
	detector := NewLeakDetector(10*time.Millisecond, 1.001)

	report := detector.GetReport()
	assert.False(t, report.PotentialLeak)
}

func TestGetCurrentMemoryUsage(t *testing.T) {
	usage := GetCurrentMemoryUsage()

	assert.Contains(t, usage, "HeapAlloc")
	assert.Contains(t, usage, "HeapSys")
	assert.Contains(t, usage, "Goroutines")
}

func TestForceGC(t *testing.T) {
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	_ = make([]byte, 1024*1024)

	ForceGC()

	runtime.ReadMemStats(&m2)

	assert.GreaterOrEqual(t, m2.NumGC, m1.NumGC)
}

func TestNewMemoryMonitor(t *testing.T) {
	monitor := NewMemoryMonitor(time.Second, 2.0)

	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.detector)
	assert.NotNil(t, monitor.reportCh)
}

func TestMemoryMonitor_StartStop(t *testing.T) {
	monitor := NewMemoryMonitor(50*time.Millisecond, 2.0)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := monitor.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	monitor.Stop()
}

func TestMemoryMonitor_AlertCallback(t *testing.T) {
	alertReceived := false
	var receivedReport LeakReport

	monitor := NewMemoryMonitor(10*time.Millisecond, 0.001)
	monitor.SetAlertCallback(func(r LeakReport) {
		alertReceived = true
		receivedReport = r
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := monitor.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	monitor.Stop()

	assert.True(t, alertReceived)
	assert.NotZero(t, receivedReport.Timestamp)
}

func TestLeakReport_Fields(t *testing.T) {
	report := LeakReport{
		Timestamp:       time.Now(),
		HeapAlloc:       1024,
		HeapSys:         2048,
		HeapInUse:       1024,
		HeapObjects:     10,
		StackInUse:      512,
		GoroutineCount:  5,
		GCCount:         2,
		HeapGrowthRatio: 1.5,
		PotentialLeak:   true,
	}

	assert.Equal(t, uint64(1024), report.HeapAlloc)
	assert.Equal(t, uint64(2048), report.HeapSys)
	assert.Equal(t, uint64(1024), report.HeapInUse)
	assert.Equal(t, uint64(10), report.HeapObjects)
	assert.Equal(t, uint64(512), report.StackInUse)
	assert.Equal(t, 5, report.GoroutineCount)
	assert.Equal(t, uint32(2), report.GCCount)
	assert.Equal(t, 1.5, report.HeapGrowthRatio)
	assert.True(t, report.PotentialLeak)
}

func TestLeakDetector_ConcurrentAccess(t *testing.T) {
	detector := NewLeakDetector(10*time.Millisecond, 2.0)

	ctx := context.Background()
	err := detector.Start(ctx)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = detector.GetReport()
				_ = detector.GetSamples()
				time.Sleep(time.Millisecond)
			}
		}()
	}

	time.Sleep(100 * time.Millisecond)
	detector.Stop()
	wg.Wait()
}

func BenchmarkLeakDetector_GetReport(b *testing.B) {
	detector := NewLeakDetector(time.Second, 2.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.GetReport()
	}
}

func BenchmarkGetCurrentMemoryUsage(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetCurrentMemoryUsage()
	}
}
