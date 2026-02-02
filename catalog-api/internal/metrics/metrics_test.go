package metrics

import (
	"runtime"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getGaugeValue(g prometheus.Gauge) float64 {
	m := &dto.Metric{}
	g.Write(m)
	return m.GetGauge().GetValue()
}

func getCounterValue(c prometheus.Counter) float64 {
	m := &dto.Metric{}
	c.Write(m)
	return m.GetCounter().GetValue()
}

func getHistogramCount(h prometheus.Observer) uint64 {
	hist, ok := h.(prometheus.Metric)
	if !ok {
		return 0
	}
	m := &dto.Metric{}
	hist.Write(m)
	return m.GetHistogram().GetSampleCount()
}

func TestHTTPRequestsTotal(t *testing.T) {
	before := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/test", "200"))
	HTTPRequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
	after := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/test", "200"))
	assert.Equal(t, before+1, after)
}

func TestHTTPRequestDuration(t *testing.T) {
	beforeCount := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/test-duration", "200"))
	HTTPRequestDuration.WithLabelValues("GET", "/test-duration", "200").Observe(0.5)
	afterCount := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/test-duration", "200"))
	assert.Equal(t, beforeCount+1, afterCount)
}

func TestHTTPActiveConnections(t *testing.T) {
	initial := getGaugeValue(HTTPActiveConnections)
	HTTPActiveConnections.Inc()
	assert.Equal(t, initial+1, getGaugeValue(HTTPActiveConnections))
	HTTPActiveConnections.Dec()
	assert.Equal(t, initial, getGaugeValue(HTTPActiveConnections))
}

func TestWebSocketConnections(t *testing.T) {
	initial := getGaugeValue(WebSocketConnections)
	WebSocketConnections.Inc()
	assert.Equal(t, initial+1, getGaugeValue(WebSocketConnections))
	WebSocketConnections.Dec()
	assert.Equal(t, initial, getGaugeValue(WebSocketConnections))
}

func TestSMBHealthStatus(t *testing.T) {
	SetSMBHealth("nas-1", SMBHealthy)
	m := &dto.Metric{}
	SMBHealthStatus.WithLabelValues("nas-1").Write(m)
	assert.Equal(t, SMBHealthy, m.GetGauge().GetValue())

	SetSMBHealth("nas-1", SMBDegraded)
	m = &dto.Metric{}
	SMBHealthStatus.WithLabelValues("nas-1").Write(m)
	assert.Equal(t, SMBDegraded, m.GetGauge().GetValue())

	SetSMBHealth("nas-1", SMBOffline)
	m = &dto.Metric{}
	SMBHealthStatus.WithLabelValues("nas-1").Write(m)
	assert.Equal(t, SMBOffline, m.GetGauge().GetValue())
}

func TestObserveDBQuery(t *testing.T) {
	beforeCount := getHistogramCount(DBQueryDuration.WithLabelValues("SELECT", "files"))
	ObserveDBQuery("SELECT", "files", 50*time.Millisecond)
	afterCount := getHistogramCount(DBQueryDuration.WithLabelValues("SELECT", "files"))
	assert.Equal(t, beforeCount+1, afterCount)
}

func TestUpdateRuntimeMetrics(t *testing.T) {
	updateRuntimeMetrics()

	goroutines := getGaugeValue(GoroutineCount)
	assert.Greater(t, goroutines, float64(0))
	assert.InDelta(t, float64(runtime.NumGoroutine()), goroutines, 10)

	alloc := getGaugeValue(MemoryAlloc)
	assert.Greater(t, alloc, float64(0))

	sys := getGaugeValue(MemorySys)
	assert.Greater(t, sys, float64(0))

	heap := getGaugeValue(MemoryHeapInuse)
	assert.Greater(t, heap, float64(0))
}

func TestSMBHealthConstants(t *testing.T) {
	assert.Equal(t, 1.0, SMBHealthy)
	assert.Equal(t, 0.5, SMBDegraded)
	assert.Equal(t, 0.0, SMBOffline)
}

func TestStartAndStopRuntimeCollector(t *testing.T) {
	// Reset the once for testing purposes - we test the function directly instead
	// since sync.Once cannot be reset.
	done := make(chan struct{})
	stop := make(chan struct{})

	go func() {
		collectRuntimeMetrics(50*time.Millisecond, stop)
		close(done)
	}()

	// Let it run a few cycles.
	time.Sleep(200 * time.Millisecond)

	goroutines := getGaugeValue(GoroutineCount)
	require.Greater(t, goroutines, float64(0))

	close(stop)

	select {
	case <-done:
		// Collector stopped successfully.
	case <-time.After(2 * time.Second):
		t.Fatal("runtime collector did not stop within timeout")
	}
}
