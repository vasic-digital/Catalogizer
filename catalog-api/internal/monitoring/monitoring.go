package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MonitoringConfig holds configuration for monitoring
type MonitoringConfig struct {
	// Prometheus
	EnablePrometheus bool
	PrometheusPort   int

	// Service Info
	ServiceName    string
	ServiceVersion string

	// Health Checks
	EnableHealthCheck bool
	HealthCheckPath   string
}

// DefaultMonitoringConfig returns default configuration
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		EnablePrometheus:  true,
		PrometheusPort:    9090,
		ServiceName:       "catalog-api",
		ServiceVersion:    "1.0.0",
		EnableHealthCheck: true,
		HealthCheckPath:   "/health",
	}
}

// Metrics holds Prometheus metrics
type Metrics struct {
	// HTTP metrics
	RequestDuration *prometheus.HistogramVec
	RequestTotal    *prometheus.CounterVec
	RequestSize     *prometheus.SummaryVec
	ResponseSize    *prometheus.SummaryVec

	// Application metrics
	ActiveConnections prometheus.Gauge
	DatabaseQueries   *prometheus.CounterVec
	CacheHits         *prometheus.CounterVec
	CacheMisses       *prometheus.CounterVec

	// System metrics
	MemoryUsage prometheus.Gauge
	Goroutines  prometheus.Gauge
	GCCount     prometheus.Counter
}

// NewMetrics creates new metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		RequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_request_size_bytes",
				Help: "HTTP request size in bytes",
			},
			[]string{"method", "path"},
		),
		ResponseSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_response_size_bytes",
				Help: "HTTP response size in bytes",
			},
			[]string{"method", "path"},
		),
		ActiveConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_connections",
				Help: "Number of active connections",
			},
		),
		DatabaseQueries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total database queries",
			},
			[]string{"operation", "table"},
		),
		CacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total cache hits",
			},
			[]string{"cache_type"},
		),
		CacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total cache misses",
			},
			[]string{"cache_type"},
		),
		MemoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),
		Goroutines: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "goroutines_count",
				Help: "Number of goroutines",
			},
		),
		GCCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "gc_count_total",
				Help: "Total garbage collection cycles",
			},
		),
	}

	// Register metrics
	prometheus.MustRegister(
		m.RequestDuration,
		m.RequestTotal,
		m.RequestSize,
		m.ResponseSize,
		m.ActiveConnections,
		m.DatabaseQueries,
		m.CacheHits,
		m.CacheMisses,
		m.MemoryUsage,
		m.Goroutines,
		m.GCCount,
	)

	return m
}

// Monitor starts system monitoring goroutine
func (m *Metrics) Monitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.updateSystemMetrics()
		case <-ctx.Done():
			return
		}
	}
}

// updateSystemMetrics updates system-level metrics
func (m *Metrics) updateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.MemoryUsage.Set(float64(memStats.Alloc))
	m.Goroutines.Set(float64(runtime.NumGoroutine()))
	m.GCCount.Add(float64(memStats.NumGC))
}

// PrometheusMiddleware returns Gin middleware for Prometheus metrics
func PrometheusMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		metrics.ActiveConnections.Inc()
		defer metrics.ActiveConnections.Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())

		metrics.RequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
		metrics.RequestTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		metrics.RequestSize.WithLabelValues(c.Request.Method, path).Observe(float64(c.Request.ContentLength))
		metrics.ResponseSize.WithLabelValues(c.Request.Method, path).Observe(float64(c.Writer.Size()))
	}
}

// HealthCheck represents health check status
type HealthCheck struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]interface{} `json:"checks"`
}

// HealthChecker performs health checks
type HealthChecker struct {
	version string
	checks  map[string]func() error
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		version: version,
		checks:  make(map[string]func() error),
	}
}

// RegisterCheck registers a health check
func (h *HealthChecker) RegisterCheck(name string, check func() error) {
	h.checks[name] = check
}

// Check performs all health checks
func (h *HealthChecker) Check() *HealthCheck {
	result := &HealthCheck{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   h.version,
		Checks:    make(map[string]interface{}),
	}

	for name, check := range h.checks {
		if err := check(); err != nil {
			result.Status = "unhealthy"
			result.Checks[name] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			result.Checks[name] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	return result
}

// GinHealthHandler returns Gin handler for health checks
func (h *HealthChecker) GinHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		health := h.Check()

		status := http.StatusOK
		if health.Status == "unhealthy" {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, health)
	}
}

// StartPrometheusServer starts Prometheus metrics server
func StartPrometheusServer(port int, logger *zap.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	logger.Info("Starting Prometheus server", zap.Int("port", port))

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Prometheus server error", zap.Error(err))
		}
	}()
}
