package monitoring

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// HealthChecker — no checks registered
// ---------------------------------------------------------------------------

func TestHealthChecker_NoChecks(t *testing.T) {
	hc := NewHealthChecker("1.0.0")
	result := hc.Check()

	assert.Equal(t, "healthy", result.Status)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Empty(t, result.Checks)
	assert.False(t, result.Timestamp.IsZero())
}

// ---------------------------------------------------------------------------
// HealthChecker — check override
// ---------------------------------------------------------------------------

func TestHealthChecker_OverrideCheck(t *testing.T) {
	hc := NewHealthChecker("1.0.0")

	hc.RegisterCheck("db", func() error { return errors.New("down") })
	result1 := hc.Check()
	assert.Equal(t, "unhealthy", result1.Status)

	// Override with healthy check
	hc.RegisterCheck("db", func() error { return nil })
	result2 := hc.Check()
	assert.Equal(t, "healthy", result2.Status)
}

// ---------------------------------------------------------------------------
// HealthChecker — all unhealthy
// ---------------------------------------------------------------------------

func TestHealthChecker_AllUnhealthy(t *testing.T) {
	hc := NewHealthChecker("2.0.0")
	hc.RegisterCheck("db", func() error { return errors.New("db down") })
	hc.RegisterCheck("cache", func() error { return errors.New("cache down") })
	hc.RegisterCheck("queue", func() error { return errors.New("queue down") })

	result := hc.Check()
	assert.Equal(t, "unhealthy", result.Status)
	assert.Len(t, result.Checks, 3)

	for name, detail := range result.Checks {
		m, ok := detail.(map[string]interface{})
		require.True(t, ok, "check %q should be a map", name)
		assert.Equal(t, "unhealthy", m["status"])
		assert.NotEmpty(t, m["error"])
	}
}

// ---------------------------------------------------------------------------
// HealthCheck — JSON serialization
// ---------------------------------------------------------------------------

func TestHealthCheck_JSONSerialization(t *testing.T) {
	hc := NewHealthChecker("3.0.0")
	hc.RegisterCheck("db", func() error { return nil })

	result := hc.Check()

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var restored HealthCheck
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "healthy", restored.Status)
	assert.Equal(t, "3.0.0", restored.Version)
	assert.NotZero(t, restored.Timestamp)
}

// ---------------------------------------------------------------------------
// GinHealthHandler — no checks registered
// ---------------------------------------------------------------------------

func TestGinHealthHandler_NoChecks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hc := NewHealthChecker("1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	handler := hc.GinHealthHandler()
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body HealthCheck
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "healthy", body.Status)
	assert.Empty(t, body.Checks)
}

// ---------------------------------------------------------------------------
// PrometheusMiddleware — multiple routes
// ---------------------------------------------------------------------------

func TestPrometheusMiddleware_MultipleRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := newTestMetrics(t)

	router := gin.New()
	router.Use(PrometheusMiddleware(m))
	router.GET("/api/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	router.POST("/api/items", func(c *gin.Context) {
		c.String(http.StatusCreated, "created")
	})
	router.GET("/api/not-found", func(c *gin.Context) {
		c.String(http.StatusNotFound, "not found")
	})

	// GET /api/health
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	assert.Equal(t, http.StatusOK, w1.Code)

	// POST /api/items
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/api/items", nil))
	assert.Equal(t, http.StatusCreated, w2.Code)

	// GET /api/not-found
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/api/not-found", nil))
	assert.Equal(t, http.StatusNotFound, w3.Code)
}

// ---------------------------------------------------------------------------
// Monitor — context cancellation stops monitoring
// ---------------------------------------------------------------------------

func TestMetrics_Monitor_ContextCancellation(t *testing.T) {
	m := newTestMetrics(t)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		m.Monitor(ctx)
		close(done)
	}()

	// Cancel context to stop monitoring
	cancel()

	select {
	case <-done:
		// Monitor returned — expected
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not stop after context cancellation")
	}
}

// ---------------------------------------------------------------------------
// Metrics — database and cache counters
// ---------------------------------------------------------------------------

func TestMetrics_DatabaseQueries(t *testing.T) {
	m := newTestMetrics(t)

	m.DatabaseQueries.WithLabelValues("SELECT", "files").Inc()
	m.DatabaseQueries.WithLabelValues("SELECT", "files").Inc()
	m.DatabaseQueries.WithLabelValues("INSERT", "media_items").Inc()

	selectCount := counterVecValue(m.DatabaseQueries, "SELECT", "files")
	assert.Equal(t, float64(2), selectCount)

	insertCount := counterVecValue(m.DatabaseQueries, "INSERT", "media_items")
	assert.Equal(t, float64(1), insertCount)
}

func TestMetrics_CacheHitsAndMisses(t *testing.T) {
	m := newTestMetrics(t)

	m.CacheHits.WithLabelValues("memory").Add(10)
	m.CacheMisses.WithLabelValues("memory").Add(3)

	hits := counterVecValue(m.CacheHits, "memory")
	assert.Equal(t, float64(10), hits)

	misses := counterVecValue(m.CacheMisses, "memory")
	assert.Equal(t, float64(3), misses)
}

// ---------------------------------------------------------------------------
// Metrics — active connections gauge
// ---------------------------------------------------------------------------

func TestMetrics_ActiveConnections(t *testing.T) {
	m := newTestMetrics(t)

	assert.Equal(t, float64(0), gaugeValue(m.ActiveConnections))

	m.ActiveConnections.Inc()
	m.ActiveConnections.Inc()
	m.ActiveConnections.Inc()
	assert.Equal(t, float64(3), gaugeValue(m.ActiveConnections))

	m.ActiveConnections.Dec()
	assert.Equal(t, float64(2), gaugeValue(m.ActiveConnections))
}

// ---------------------------------------------------------------------------
// MonitoringConfig defaults
// ---------------------------------------------------------------------------

func TestDefaultMonitoringConfig_AllFields(t *testing.T) {
	cfg := DefaultMonitoringConfig()

	assert.True(t, cfg.EnablePrometheus)
	assert.Equal(t, 9090, cfg.PrometheusPort)
	assert.Equal(t, "catalog-api", cfg.ServiceName)
	assert.Equal(t, "1.0.0", cfg.ServiceVersion)
	assert.True(t, cfg.EnableHealthCheck)
	assert.Equal(t, "/health", cfg.HealthCheckPath)
}

func TestMonitoringConfig_CustomValues(t *testing.T) {
	cfg := MonitoringConfig{
		EnablePrometheus:  false,
		PrometheusPort:    9191,
		ServiceName:       "custom-api",
		ServiceVersion:    "2.5.0",
		EnableHealthCheck: false,
		HealthCheckPath:   "/status",
	}

	assert.False(t, cfg.EnablePrometheus)
	assert.Equal(t, 9191, cfg.PrometheusPort)
	assert.Equal(t, "custom-api", cfg.ServiceName)
	assert.Equal(t, "2.5.0", cfg.ServiceVersion)
	assert.False(t, cfg.EnableHealthCheck)
	assert.Equal(t, "/status", cfg.HealthCheckPath)
}

// ---------------------------------------------------------------------------
// NewMetrics — constructor coverage
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// StartPrometheusServer — verify it starts without panic
// ---------------------------------------------------------------------------

func TestStartPrometheusServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	// Use port 0 to let the OS pick a free port — avoids conflicts.
	// StartPrometheusServer launches a goroutine, so we just verify
	// it doesn't panic on startup.
	assert.NotPanics(t, func() {
		StartPrometheusServer(0, logger)
	})
	// Give the goroutine a moment to start, then let it die with the test.
	time.Sleep(10 * time.Millisecond)
}
