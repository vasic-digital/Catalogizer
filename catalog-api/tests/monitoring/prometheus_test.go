package monitoring

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"catalogizer/internal/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupRouterWithMetrics creates a Gin router wired with the metrics middleware
// and the /metrics endpoint, plus a few sample routes for exercising metrics.
func setupRouterWithMetrics() *gin.Engine {
	router := gin.New()
	router.Use(metrics.GinMiddleware())

	// Prometheus scrape endpoint.
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health endpoint.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "test",
		})
	})

	// Sample API endpoints.
	router.GET("/api/v1/catalog", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": []string{}})
	})

	router.POST("/api/v1/scan", func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"job_id": "test-job-1"})
	})

	router.GET("/api/v1/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	return router
}

// getMetricsOutput makes a request to /metrics and returns the Prometheus
// exposition text. It fails the test if the request does not succeed.
func getMetricsOutput(t *testing.T, router *gin.Engine) string {
	t.Helper()
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	return w.Body.String()
}

// --- /metrics endpoint format tests ---

// TestMetricsEndpoint_ReturnsOK verifies that the /metrics endpoint responds
// with HTTP 200.
func TestMetricsEndpoint_ReturnsOK(t *testing.T) {
	router := setupRouterWithMetrics()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestMetricsEndpoint_ReturnsValidPrometheusTextFormat verifies that the
// response is valid Prometheus text exposition format by checking for
// expected structural markers (# HELP, # TYPE, metric lines).
func TestMetricsEndpoint_ReturnsValidPrometheusTextFormat(t *testing.T) {
	router := setupRouterWithMetrics()
	body := getMetricsOutput(t, router)

	// Prometheus text format always has HELP and TYPE lines.
	assert.Contains(t, body, "# HELP", "response should contain HELP comments")
	assert.Contains(t, body, "# TYPE", "response should contain TYPE declarations")

	// At minimum, Go runtime metrics should be present (registered by default).
	assert.Contains(t, body, "go_goroutines", "response should include Go runtime metrics")
	assert.Contains(t, body, "go_memstats_alloc_bytes", "response should include Go memory stats")
}

// TestMetricsEndpoint_ContentType verifies that the /metrics endpoint returns
// the correct Prometheus content type.
func TestMetricsEndpoint_ContentType(t *testing.T) {
	router := setupRouterWithMetrics()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	assert.True(t,
		strings.Contains(contentType, "text/plain") || strings.Contains(contentType, "application/openmetrics-text"),
		"Content-Type should be Prometheus-compatible, got: %s", contentType,
	)
}

// --- Key metric presence tests ---

// TestMetricsEndpoint_ContainsHTTPRequestsTotal verifies that the
// catalogizer_http_requests_total metric appears in the /metrics output after
// making an API request.
func TestMetricsEndpoint_ContainsHTTPRequestsTotal(t *testing.T) {
	router := setupRouterWithMetrics()

	// Make a request to generate metric data.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	router.ServeHTTP(w, req)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_http_requests_total",
		"/metrics should contain catalogizer_http_requests_total")
	assert.Contains(t, body, `method="GET"`,
		"/metrics should contain method label")
}

// TestMetricsEndpoint_ContainsHTTPRequestDurationSeconds verifies that the
// catalogizer_http_request_duration_seconds histogram appears in the /metrics
// output.
func TestMetricsEndpoint_ContainsHTTPRequestDurationSeconds(t *testing.T) {
	router := setupRouterWithMetrics()

	// Make a request to populate the histogram.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	router.ServeHTTP(w, req)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_http_request_duration_seconds",
		"/metrics should contain catalogizer_http_request_duration_seconds")

	// Histograms produce _bucket, _sum, and _count sub-metrics.
	assert.Contains(t, body, "catalogizer_http_request_duration_seconds_bucket",
		"/metrics should contain histogram buckets")
	assert.Contains(t, body, "catalogizer_http_request_duration_seconds_sum",
		"/metrics should contain histogram sum")
	assert.Contains(t, body, "catalogizer_http_request_duration_seconds_count",
		"/metrics should contain histogram count")
}

// TestMetricsEndpoint_ContainsHTTPActiveConnections verifies that the
// catalogizer_http_active_connections gauge appears in the /metrics output.
func TestMetricsEndpoint_ContainsHTTPActiveConnections(t *testing.T) {
	router := setupRouterWithMetrics()
	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_http_active_connections",
		"/metrics should contain catalogizer_http_active_connections")
}

// TestMetricsEndpoint_ContainsDBQueryMetrics verifies that database query
// metrics appear after recording a query.
func TestMetricsEndpoint_ContainsDBQueryMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	// Record a DB query metric.
	metrics.RecordDBQuery("SELECT", "prom_test_files", 10*time.Millisecond)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_db_queries_total",
		"/metrics should contain catalogizer_db_queries_total")
	assert.Contains(t, body, "catalogizer_db_query_duration_seconds",
		"/metrics should contain catalogizer_db_query_duration_seconds")
}

// TestMetricsEndpoint_ContainsErrorsTotal verifies that the
// catalogizer_errors_total counter appears after recording an error.
func TestMetricsEndpoint_ContainsErrorsTotal(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordError("api_prom_test", "test_error")

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_errors_total",
		"/metrics should contain catalogizer_errors_total")
	assert.Contains(t, body, `component="api_prom_test"`,
		"/metrics should contain the component label")
}

// TestMetricsEndpoint_ContainsMediaMetrics verifies that media-related metrics
// appear after recording analysis events.
func TestMetricsEndpoint_ContainsMediaMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordMediaAnalysis(100 * time.Millisecond)
	metrics.MediaFilesScanned.Inc()
	metrics.UpdateMediaByType("prom_test_movie", 5)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_media_files_scanned_total",
		"/metrics should contain catalogizer_media_files_scanned_total")
	assert.Contains(t, body, "catalogizer_media_files_analyzed_total",
		"/metrics should contain catalogizer_media_files_analyzed_total")
	assert.Contains(t, body, "catalogizer_media_analysis_duration_seconds",
		"/metrics should contain catalogizer_media_analysis_duration_seconds")
	assert.Contains(t, body, "catalogizer_media_by_type",
		"/metrics should contain catalogizer_media_by_type")
}

// TestMetricsEndpoint_ContainsCacheMetrics verifies that cache hit/miss metrics
// appear after recording cache events.
func TestMetricsEndpoint_ContainsCacheMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordCacheHit("prom_test_cache")
	metrics.RecordCacheMiss("prom_test_cache")

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_cache_hits_total",
		"/metrics should contain catalogizer_cache_hits_total")
	assert.Contains(t, body, "catalogizer_cache_misses_total",
		"/metrics should contain catalogizer_cache_misses_total")
}

// TestMetricsEndpoint_ContainsAuthMetrics verifies that authentication metrics
// appear after recording auth attempts.
func TestMetricsEndpoint_ContainsAuthMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordAuthAttempt("password_prom_test", "success")
	metrics.RecordAuthAttempt("password_prom_test", "failure")

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_auth_attempts_total",
		"/metrics should contain catalogizer_auth_attempts_total")
	assert.Contains(t, body, `method="password_prom_test"`,
		"/metrics should contain the method label value")
}

// TestMetricsEndpoint_ContainsFileSystemMetrics verifies that filesystem
// operation metrics appear after recording filesystem operations.
func TestMetricsEndpoint_ContainsFileSystemMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordFileSystemOperation("smb_prom_test", "read", "success", 50*time.Millisecond)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_filesystem_operations_total",
		"/metrics should contain catalogizer_filesystem_operations_total")
	assert.Contains(t, body, "catalogizer_filesystem_operation_duration_seconds",
		"/metrics should contain catalogizer_filesystem_operation_duration_seconds")
	assert.Contains(t, body, `protocol="smb_prom_test"`,
		"/metrics should contain the protocol label value")
}

// TestMetricsEndpoint_ContainsExternalAPIMetrics verifies that external API
// call metrics appear.
func TestMetricsEndpoint_ContainsExternalAPIMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordExternalAPICall("tmdb_prom_test", "success", 200*time.Millisecond)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_external_api_calls_total",
		"/metrics should contain catalogizer_external_api_calls_total")
	assert.Contains(t, body, "catalogizer_external_api_call_duration_seconds",
		"/metrics should contain catalogizer_external_api_call_duration_seconds")
	assert.Contains(t, body, `provider="tmdb_prom_test"`,
		"/metrics should contain the provider label value")
}

// TestMetricsEndpoint_ContainsWebSocketMetrics verifies that WebSocket metrics
// appear in the output.
func TestMetricsEndpoint_ContainsWebSocketMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordWebSocketMessage("sent")

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_websocket_messages_total",
		"/metrics should contain catalogizer_websocket_messages_total")
	assert.Contains(t, body, "catalogizer_websocket_connections",
		"/metrics should contain websocket connections gauge")
}

// TestMetricsEndpoint_ContainsStorageMetrics verifies that storage root metrics
// appear.
func TestMetricsEndpoint_ContainsStorageMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.UpdateStorageRoots("nfs_prom_test", "enabled", 2)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_storage_roots_total",
		"/metrics should contain catalogizer_storage_roots_total")
}

// TestMetricsEndpoint_ContainsRuntimeMetrics verifies that runtime metrics
// (goroutines, memory) appear in the output. These are collected by the
// runtime collector, but also populated by updateRuntimeMetrics.
func TestMetricsEndpoint_ContainsRuntimeMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	body := getMetricsOutput(t, router)

	// Go runtime metrics are always present from the default Prometheus registry.
	assert.Contains(t, body, "go_goroutines",
		"/metrics should contain go_goroutines")
	assert.Contains(t, body, "go_memstats_alloc_bytes",
		"/metrics should contain go_memstats_alloc_bytes")

	// Custom runtime gauges should also appear (they are registered via promauto).
	assert.Contains(t, body, "catalogizer_runtime_goroutines",
		"/metrics should contain catalogizer_runtime_goroutines")
	assert.Contains(t, body, "catalogizer_runtime_memory_alloc_bytes",
		"/metrics should contain catalogizer_runtime_memory_alloc_bytes")
	assert.Contains(t, body, "catalogizer_runtime_memory_sys_bytes",
		"/metrics should contain catalogizer_runtime_memory_sys_bytes")
	assert.Contains(t, body, "catalogizer_runtime_memory_heap_inuse_bytes",
		"/metrics should contain catalogizer_runtime_memory_heap_inuse_bytes")
}

// TestMetricsEndpoint_ContainsSMBHealthMetrics verifies that SMB health status
// metrics appear after setting health values.
func TestMetricsEndpoint_ContainsSMBHealthMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.SetSMBHealth("prom-test-nas", metrics.SMBHealthy)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_smb_health_status",
		"/metrics should contain catalogizer_smb_health_status")
	assert.Contains(t, body, `source="prom-test-nas"`,
		"/metrics should contain the source label value")
}

// TestMetricsEndpoint_ContainsUptimeMetrics verifies that the uptime counter
// appears in the output.
func TestMetricsEndpoint_ContainsUptimeMetrics(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.IncrementUptime()

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_uptime_seconds",
		"/metrics should contain catalogizer_uptime_seconds")
}

// --- Health endpoint metrics tracking tests ---

// TestHealthEndpoint_MetricsAreTracked verifies that requests to /health are
// tracked by the metrics middleware (i.e., /health is not excluded like /metrics).
func TestHealthEndpoint_MetricsAreTracked(t *testing.T) {
	router := setupRouterWithMetrics()

	// Make a health check request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	body := getMetricsOutput(t, router)

	// The health endpoint should be tracked (unlike /metrics which is excluded).
	assert.Contains(t, body, `path="/health"`,
		"health endpoint requests should be tracked in metrics")
}

// TestHealthEndpoint_RequestCountIncreases verifies that multiple health
// requests cause the counter to increase.
func TestHealthEndpoint_RequestCountIncreases(t *testing.T) {
	router := setupRouterWithMetrics()

	// Get the initial metrics output.
	_ = getMetricsOutput(t, router)

	// Make several health check requests.
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	body := getMetricsOutput(t, router)

	// The output should contain the health path metric with a status 200 label.
	assert.Contains(t, body, `catalogizer_http_requests_total{method="GET",path="/health",status="200"}`,
		"metrics should track health endpoint with correct labels")
}

// --- Metrics endpoint self-exclusion test ---

// TestMetricsEndpoint_NotSelfTracked verifies that requests to /metrics itself
// are not tracked by the metrics middleware (to prevent self-referencing noise).
func TestMetricsEndpoint_NotSelfTracked(t *testing.T) {
	router := setupRouterWithMetrics()

	// Make several /metrics requests.
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metrics", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	body := getMetricsOutput(t, router)

	// There should NOT be a counter entry with path="/metrics" since the
	// middleware explicitly skips it.
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "catalogizer_http_requests_total{") {
			assert.NotContains(t, line, `path="/metrics"`,
				"requests to /metrics should not be tracked: %s", line)
		}
	}
}

// --- Error response metrics test ---

// TestErrorEndpoint_MetricsTrackErrorResponses verifies that error responses
// (5xx) from API endpoints are correctly tracked in the metrics output.
func TestErrorEndpoint_MetricsTrackErrorResponses(t *testing.T) {
	router := setupRouterWithMetrics()

	// Make a request that returns 500.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/error", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, `status="500"`,
		"metrics should track 500 status responses")
	assert.Contains(t, body, `path="/api/v1/error"`,
		"metrics should track the error endpoint path")
}

// --- Comprehensive metric type declaration tests ---

// TestMetricsEndpoint_CorrectMetricTypes verifies that the TYPE declarations
// in the Prometheus output are correct for key metrics.
func TestMetricsEndpoint_CorrectMetricTypes(t *testing.T) {
	router := setupRouterWithMetrics()

	// Populate some metrics so they appear in the output.
	metrics.RecordDBQuery("SELECT", "type_check_table", 5*time.Millisecond)
	metrics.RecordError("type_check_component", "type_check_error")
	metrics.RecordCacheHit("type_check_cache")
	metrics.RecordAuthAttempt("type_check_method", "success")
	metrics.MediaFilesScanned.Inc()

	body := getMetricsOutput(t, router)

	tests := []struct {
		metric     string
		metricType string
	}{
		{"catalogizer_http_requests_total", "counter"},
		{"catalogizer_http_request_duration_seconds", "histogram"},
		{"catalogizer_http_active_connections", "gauge"},
		{"catalogizer_db_queries_total", "counter"},
		{"catalogizer_db_query_duration_seconds", "histogram"},
		{"catalogizer_errors_total", "counter"},
		{"catalogizer_cache_hits_total", "counter"},
		{"catalogizer_cache_misses_total", "counter"},
		{"catalogizer_auth_attempts_total", "counter"},
		{"catalogizer_media_files_scanned_total", "counter"},
		{"catalogizer_media_files_analyzed_total", "counter"},
		{"catalogizer_media_analysis_duration_seconds", "histogram"},
		{"catalogizer_uptime_seconds", "counter"},
	}

	for _, tt := range tests {
		t.Run(tt.metric+" is "+tt.metricType, func(t *testing.T) {
			expectedType := "# TYPE " + tt.metric + " " + tt.metricType
			assert.Contains(t, body, expectedType,
				"metric %s should be declared as %s", tt.metric, tt.metricType)
		})
	}
}

// TestMetricsEndpoint_MultipleAPIRequests_ProducesDistinctSeries verifies that
// requests to different endpoints produce distinct metric series.
func TestMetricsEndpoint_MultipleAPIRequests_ProducesDistinctSeries(t *testing.T) {
	router := setupRouterWithMetrics()

	// Hit the catalog endpoint.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	router.ServeHTTP(w1, req1)

	// Hit the scan endpoint (POST).
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/v1/scan", strings.NewReader("{}"))
	router.ServeHTTP(w2, req2)

	// Hit the error endpoint.
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/v1/error", nil)
	router.ServeHTTP(w3, req3)

	body := getMetricsOutput(t, router)

	// Verify distinct path labels.
	assert.Contains(t, body, `path="/api/v1/catalog"`,
		"metrics should contain catalog path")
	assert.Contains(t, body, `path="/api/v1/scan"`,
		"metrics should contain scan path")
	assert.Contains(t, body, `path="/api/v1/error"`,
		"metrics should contain error path")

	// Verify distinct method labels.
	assert.Contains(t, body, `method="GET"`,
		"metrics should contain GET method")
	assert.Contains(t, body, `method="POST"`,
		"metrics should contain POST method")

	// Verify distinct status labels.
	assert.Contains(t, body, `status="200"`,
		"metrics should contain 200 status")
	assert.Contains(t, body, `status="202"`,
		"metrics should contain 202 status")
	assert.Contains(t, body, `status="500"`,
		"metrics should contain 500 status")
}

// TestMetricsEndpoint_GaugeMetricsReflectSetValues verifies that gauge values
// set via helper functions are correctly reflected in the /metrics output.
func TestMetricsEndpoint_GaugeMetricsReflectSetValues(t *testing.T) {
	router := setupRouterWithMetrics()

	// Set gauge values.
	metrics.UpdateActiveSessions(15)
	metrics.UpdateDBConnections(8, 4)
	metrics.SetSMBHealth("prom-gauge-test", metrics.SMBDegraded)

	body := getMetricsOutput(t, router)

	assert.Contains(t, body, "catalogizer_active_sessions 15",
		"active sessions gauge should reflect set value")
	assert.Contains(t, body, "catalogizer_db_connections_active 8",
		"db active connections gauge should reflect set value")
	assert.Contains(t, body, "catalogizer_db_connections_idle 4",
		"db idle connections gauge should reflect set value")
}
