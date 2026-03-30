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

// =============================================================================
// COMPREHENSIVE METRICS TEST: All Prometheus metrics registered and emitting
// =============================================================================

// setupComprehensiveRouter creates a Gin router with metrics middleware, the
// /metrics scrape endpoint, and sample API routes that exercise the HTTP
// request metrics pipeline.
func setupComprehensiveRouter() *gin.Engine {
	router := gin.New()
	router.Use(metrics.GinMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.GET("/api/v1/catalog", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": []string{}})
	})

	router.POST("/api/v1/scan", func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"job_id": "comprehensive-test-1"})
	})

	router.GET("/api/v1/media", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": []string{}, "total": 0})
	})

	return router
}

// getComprehensiveMetrics makes a request to /metrics and returns the output.
func getComprehensiveMetrics(t *testing.T, router *gin.Engine) string {
	t.Helper()
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	return w.Body.String()
}

// populateAllMetrics generates activity across all metric families so they
// appear in the /metrics output.
func populateAllMetrics(t *testing.T, router *gin.Engine) {
	t.Helper()

	// Generate HTTP request metrics via middleware.
	for _, method := range []string{"GET", "POST"} {
		path := "/api/v1/catalog"
		if method == "POST" {
			path = "/api/v1/scan"
		}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(method, path, nil)
		router.ServeHTTP(w, req)
	}

	// Database query metrics.
	metrics.RecordDBQuery("SELECT", "comp_test_files", 8*time.Millisecond)
	metrics.RecordDBQuery("INSERT", "comp_test_media_items", 12*time.Millisecond)
	metrics.RecordDBQuery("UPDATE", "comp_test_storage_roots", 5*time.Millisecond)

	// Error metrics.
	metrics.RecordError("comp_api", "request_error")
	metrics.RecordError("comp_scanner", "scan_error")

	// Cache metrics.
	metrics.RecordCacheHit("comp_metadata")
	metrics.RecordCacheHit("comp_thumbnail")
	metrics.RecordCacheMiss("comp_metadata")
	metrics.RecordCacheMiss("comp_search")

	// Auth metrics.
	metrics.RecordAuthAttempt("password_comp", "success")
	metrics.RecordAuthAttempt("password_comp", "failure")
	metrics.RecordAuthAttempt("token_comp", "success")

	// Filesystem operation metrics.
	metrics.RecordFileSystemOperation("smb_comp", "read", "success", 25*time.Millisecond)
	metrics.RecordFileSystemOperation("nfs_comp", "write", "success", 15*time.Millisecond)
	metrics.RecordFileSystemOperation("ftp_comp", "list", "error", 100*time.Millisecond)

	// External API call metrics.
	metrics.RecordExternalAPICall("tmdb_comp", "success", 150*time.Millisecond)
	metrics.RecordExternalAPICall("omdb_comp", "success", 200*time.Millisecond)
	metrics.RecordExternalAPICall("musicbrainz_comp", "error", 500*time.Millisecond)

	// Media analysis metrics.
	metrics.RecordMediaAnalysis(75 * time.Millisecond)
	metrics.RecordMediaAnalysis(120 * time.Millisecond)
	metrics.MediaFilesScanned.Inc()
	metrics.MediaFilesScanned.Add(10)
	metrics.UpdateMediaByType("comp_movie", 50)
	metrics.UpdateMediaByType("comp_tv_show", 30)
	metrics.UpdateMediaByType("comp_music", 100)

	// WebSocket metrics.
	metrics.RecordWebSocketMessage("sent")
	metrics.RecordWebSocketMessage("received")
	metrics.UpdateWebSocketConnections(5)

	// Storage metrics.
	metrics.UpdateStorageRoots("smb_comp", "enabled", 3)
	metrics.StorageSpaceUsed.WithLabelValues("comp_root").Set(1073741824) // 1 GB

	// Session and connection metrics.
	metrics.UpdateActiveSessions(12)
	metrics.UpdateDBConnections(10, 5)

	// SMB health metrics.
	metrics.SetSMBHealth("comp-nas-1", metrics.SMBHealthy)
	metrics.SetSMBHealth("comp-nas-2", metrics.SMBDegraded)

	// Uptime metric.
	metrics.IncrementUptime()

	// Cache size metric.
	metrics.CacheSize.WithLabelValues("comp_cache").Set(2048)
}

// --- HTTP request metrics tests ---

// TestComprehensive_HTTPRequestMetrics verifies that HTTP request counter,
// duration histogram, and active connections gauge all emit data after API
// requests are made through the metrics middleware.
func TestComprehensive_HTTPRequestMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	// Make requests to generate metric data.
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"GET", "/api/v1/catalog"},
		{"POST", "/api/v1/scan"},
		{"GET", "/api/v1/media"},
	}

	for _, ep := range endpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(ep.method, ep.path, nil)
		router.ServeHTTP(w, req)
	}

	body := getComprehensiveMetrics(t, router)

	t.Run("RequestsTotal", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_http_requests_total",
			"should contain HTTP requests total counter")
		assert.Contains(t, body, `method="GET"`, "should track GET method")
		assert.Contains(t, body, `method="POST"`, "should track POST method")
	})

	t.Run("RequestDurationHistogram", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_http_request_duration_seconds_bucket",
			"should contain duration histogram buckets")
		assert.Contains(t, body, "catalogizer_http_request_duration_seconds_sum",
			"should contain duration histogram sum")
		assert.Contains(t, body, "catalogizer_http_request_duration_seconds_count",
			"should contain duration histogram count")
	})

	t.Run("ActiveConnections", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_http_active_connections",
			"should contain active connections gauge")
	})

	t.Run("StatusCodeTracking", func(t *testing.T) {
		assert.Contains(t, body, `status="200"`, "should track 200 responses")
		assert.Contains(t, body, `status="202"`, "should track 202 responses")
	})

	t.Run("PathTracking", func(t *testing.T) {
		assert.Contains(t, body, `path="/health"`, "should track health path")
		assert.Contains(t, body, `path="/api/v1/catalog"`, "should track catalog path")
	})
}

// --- Latency histogram tests ---

// TestComprehensive_LatencyHistograms verifies that all histogram metrics
// produce _bucket, _sum, and _count sub-metrics with correct TYPE declarations.
func TestComprehensive_LatencyHistograms(t *testing.T) {
	router := setupComprehensiveRouter()
	populateAllMetrics(t, router)
	body := getComprehensiveMetrics(t, router)

	histograms := []struct {
		name       string
		metricType string
	}{
		{"catalogizer_http_request_duration_seconds", "histogram"},
		{"catalogizer_db_query_duration_seconds", "histogram"},
		{"catalogizer_external_api_call_duration_seconds", "histogram"},
		{"catalogizer_filesystem_operation_duration_seconds", "histogram"},
		{"catalogizer_media_analysis_duration_seconds", "histogram"},
	}

	for _, h := range histograms {
		t.Run(h.name, func(t *testing.T) {
			assert.Contains(t, body, h.name+"_bucket",
				"%s should produce _bucket sub-metric", h.name)
			assert.Contains(t, body, h.name+"_sum",
				"%s should produce _sum sub-metric", h.name)
			assert.Contains(t, body, h.name+"_count",
				"%s should produce _count sub-metric", h.name)

			typeDecl := "# TYPE " + h.name + " " + h.metricType
			assert.Contains(t, body, typeDecl,
				"%s should be declared as %s", h.name, h.metricType)
		})
	}
}

// --- Cache metrics tests ---

// TestComprehensive_CacheMetrics verifies that cache hit and miss counters
// are emitted with correct labels after cache events are recorded.
func TestComprehensive_CacheMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.RecordCacheHit("comp_test_meta")
	metrics.RecordCacheHit("comp_test_meta")
	metrics.RecordCacheMiss("comp_test_meta")
	metrics.RecordCacheMiss("comp_test_thumb")
	metrics.CacheSize.WithLabelValues("comp_test_meta").Set(4096)

	body := getComprehensiveMetrics(t, router)

	t.Run("CacheHitsTotal", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_cache_hits_total",
			"should contain cache hits counter")
		assert.Contains(t, body, `cache_type="comp_test_meta"`,
			"should track cache name label")
	})

	t.Run("CacheMissesTotal", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_cache_misses_total",
			"should contain cache misses counter")
	})

	t.Run("CacheSizeBytes", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_cache_size_bytes",
			"should contain cache size gauge")
	})
}

// --- Runtime metrics tests ---

// TestComprehensive_RuntimeMetrics verifies that Go runtime metrics (goroutines,
// memory allocation, system memory, heap) are always present in the /metrics
// output, including both native Go metrics and custom catalogizer gauges.
func TestComprehensive_RuntimeMetrics(t *testing.T) {
	router := setupComprehensiveRouter()
	body := getComprehensiveMetrics(t, router)

	t.Run("NativeGoMetrics", func(t *testing.T) {
		goMetrics := []string{
			"go_goroutines",
			"go_memstats_alloc_bytes",
			"go_memstats_sys_bytes",
			"go_memstats_heap_inuse_bytes",
		}
		for _, m := range goMetrics {
			assert.Contains(t, body, m,
				"should contain native Go metric %s", m)
		}
	})

	t.Run("CustomRuntimeGauges", func(t *testing.T) {
		customMetrics := []string{
			"catalogizer_runtime_goroutines",
			"catalogizer_runtime_memory_alloc_bytes",
			"catalogizer_runtime_memory_sys_bytes",
			"catalogizer_runtime_memory_heap_inuse_bytes",
		}
		for _, m := range customMetrics {
			assert.Contains(t, body, m,
				"should contain custom runtime metric %s", m)
		}
	})
}

// --- Database metrics tests ---

// TestComprehensive_DatabaseMetrics verifies that database query counters,
// duration histograms, and connection pool gauges emit data.
func TestComprehensive_DatabaseMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.RecordDBQuery("SELECT", "comp_db_test", 3*time.Millisecond)
	metrics.RecordDBQuery("INSERT", "comp_db_test", 7*time.Millisecond)
	metrics.UpdateDBConnections(15, 8)

	body := getComprehensiveMetrics(t, router)

	t.Run("QueryCounter", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_db_queries_total",
			"should contain db queries counter")
		assert.Contains(t, body, `operation="SELECT"`,
			"should track SELECT operations")
		assert.Contains(t, body, `operation="INSERT"`,
			"should track INSERT operations")
	})

	t.Run("QueryDurationHistogram", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_db_query_duration_seconds",
			"should contain db query duration histogram")
	})

	t.Run("ConnectionPoolGauges", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_db_connections_active",
			"should contain active connections gauge")
		assert.Contains(t, body, "catalogizer_db_connections_idle",
			"should contain idle connections gauge")
	})
}

// --- Media processing metrics tests ---

// TestComprehensive_MediaProcessingMetrics verifies that media scanning,
// analysis, and type breakdown metrics emit data.
func TestComprehensive_MediaProcessingMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.MediaFilesScanned.Add(25)
	metrics.RecordMediaAnalysis(50 * time.Millisecond)
	metrics.RecordMediaAnalysis(200 * time.Millisecond)
	metrics.UpdateMediaByType("comp_media_movie", 100)
	metrics.UpdateMediaByType("comp_media_series", 50)

	body := getComprehensiveMetrics(t, router)

	t.Run("FilesScannedCounter", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_media_files_scanned_total",
			"should contain files scanned counter")
	})

	t.Run("FilesAnalyzedCounter", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_media_files_analyzed_total",
			"should contain files analyzed counter")
	})

	t.Run("AnalysisDurationHistogram", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_media_analysis_duration_seconds",
			"should contain analysis duration histogram")
	})

	t.Run("MediaByTypeGauge", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_media_by_type",
			"should contain media by type gauge")
		assert.Contains(t, body, `type="comp_media_movie"`,
			"should track movie type")
		assert.Contains(t, body, `type="comp_media_series"`,
			"should track series type")
	})
}

// --- Authentication metrics tests ---

// TestComprehensive_AuthMetrics verifies that authentication attempt counters
// track both successful and failed attempts with correct labels.
func TestComprehensive_AuthMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.RecordAuthAttempt("password_comp_auth", "success")
	metrics.RecordAuthAttempt("password_comp_auth", "failure")
	metrics.RecordAuthAttempt("token_comp_auth", "success")

	body := getComprehensiveMetrics(t, router)

	assert.Contains(t, body, "catalogizer_auth_attempts_total",
		"should contain auth attempts counter")
	assert.Contains(t, body, `method="password_comp_auth"`,
		"should track password auth method")
	assert.Contains(t, body, `method="token_comp_auth"`,
		"should track token auth method")
	assert.Contains(t, body, `status="success"`,
		"should track success status")
	assert.Contains(t, body, `status="failure"`,
		"should track failure status")
}

// --- Filesystem operation metrics tests ---

// TestComprehensive_FilesystemMetrics verifies that filesystem operation
// counters and duration histograms track all protocols and operations.
func TestComprehensive_FilesystemMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.RecordFileSystemOperation("smb_comp_fs", "read", "success", 20*time.Millisecond)
	metrics.RecordFileSystemOperation("nfs_comp_fs", "write", "success", 30*time.Millisecond)
	metrics.RecordFileSystemOperation("ftp_comp_fs", "list", "error", 150*time.Millisecond)
	metrics.RecordFileSystemOperation("webdav_comp_fs", "stat", "success", 10*time.Millisecond)

	body := getComprehensiveMetrics(t, router)

	t.Run("OperationsCounter", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_filesystem_operations_total",
			"should contain filesystem operations counter")
		for _, proto := range []string{"smb_comp_fs", "nfs_comp_fs", "ftp_comp_fs", "webdav_comp_fs"} {
			assert.Contains(t, body, `protocol="`+proto+`"`,
				"should track protocol %s", proto)
		}
	})

	t.Run("DurationHistogram", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_filesystem_operation_duration_seconds",
			"should contain filesystem duration histogram")
	})
}

// --- External API metrics tests ---

// TestComprehensive_ExternalAPIMetrics verifies that external API call
// counters and duration histograms track all providers.
func TestComprehensive_ExternalAPIMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.RecordExternalAPICall("tmdb_comp_ext", "success", 100*time.Millisecond)
	metrics.RecordExternalAPICall("omdb_comp_ext", "error", 300*time.Millisecond)
	metrics.RecordExternalAPICall("musicbrainz_comp_ext", "success", 250*time.Millisecond)

	body := getComprehensiveMetrics(t, router)

	t.Run("CallsCounter", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_external_api_calls_total",
			"should contain external API calls counter")
		for _, provider := range []string{"tmdb_comp_ext", "omdb_comp_ext", "musicbrainz_comp_ext"} {
			assert.Contains(t, body, `provider="`+provider+`"`,
				"should track provider %s", provider)
		}
	})

	t.Run("DurationHistogram", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_external_api_call_duration_seconds",
			"should contain external API call duration histogram")
	})
}

// --- WebSocket metrics tests ---

// TestComprehensive_WebSocketMetrics verifies that WebSocket message counters
// and connection gauges emit data.
func TestComprehensive_WebSocketMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.RecordWebSocketMessage("sent")
	metrics.RecordWebSocketMessage("received")
	metrics.UpdateWebSocketConnections(8)

	body := getComprehensiveMetrics(t, router)

	assert.Contains(t, body, "catalogizer_websocket_messages_total",
		"should contain WebSocket messages counter")
	assert.Contains(t, body, "catalogizer_websocket_connections",
		"should contain WebSocket connections gauge")
}

// --- Storage root metrics tests ---

// TestComprehensive_StorageMetrics verifies that storage root counts and
// space usage gauges emit data.
func TestComprehensive_StorageMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.UpdateStorageRoots("smb_comp_storage", "enabled", 5)
	metrics.StorageSpaceUsed.WithLabelValues("comp_storage_root").Set(5368709120) // 5 GB

	body := getComprehensiveMetrics(t, router)

	assert.Contains(t, body, "catalogizer_storage_roots_total",
		"should contain storage roots gauge")
	assert.Contains(t, body, "catalogizer_storage_space_used_bytes",
		"should contain storage space used gauge")
}

// --- SMB health metrics tests ---

// TestComprehensive_SMBHealthMetrics verifies that SMB health status gauges
// are set correctly for different health states.
func TestComprehensive_SMBHealthMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.SetSMBHealth("comp-healthy-nas", metrics.SMBHealthy)
	metrics.SetSMBHealth("comp-degraded-nas", metrics.SMBDegraded)

	body := getComprehensiveMetrics(t, router)

	assert.Contains(t, body, "catalogizer_smb_health_status",
		"should contain SMB health status gauge")
	assert.Contains(t, body, `source="comp-healthy-nas"`,
		"should track healthy NAS source")
	assert.Contains(t, body, `source="comp-degraded-nas"`,
		"should track degraded NAS source")
}

// --- Session and error metrics tests ---

// TestComprehensive_SessionAndErrorMetrics verifies that active session
// gauges and error counters emit data with correct labels.
func TestComprehensive_SessionAndErrorMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.UpdateActiveSessions(20)
	metrics.RecordError("comp_handler", "validation_error")
	metrics.RecordError("comp_scanner", "timeout_error")

	body := getComprehensiveMetrics(t, router)

	t.Run("ActiveSessions", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_active_sessions",
			"should contain active sessions gauge")
	})

	t.Run("ErrorsTotal", func(t *testing.T) {
		assert.Contains(t, body, "catalogizer_errors_total",
			"should contain errors counter")
		assert.Contains(t, body, `component="comp_handler"`,
			"should track handler component")
		assert.Contains(t, body, `type="validation_error"`,
			"should track validation error type")
	})
}

// --- Uptime metrics tests ---

// TestComprehensive_UptimeMetrics verifies that the uptime counter emits data.
func TestComprehensive_UptimeMetrics(t *testing.T) {
	router := setupComprehensiveRouter()

	metrics.IncrementUptime()

	body := getComprehensiveMetrics(t, router)

	assert.Contains(t, body, "catalogizer_uptime_seconds",
		"should contain uptime counter")
	assert.Contains(t, body, "# TYPE catalogizer_uptime_seconds counter",
		"uptime should be declared as counter type")
}

// --- Full metric registration completeness test ---

// TestComprehensive_AllMetricFamiliesPresent exercises every metric family and
// verifies that all expected metrics appear in a single /metrics scrape.
func TestComprehensive_AllMetricFamiliesPresent(t *testing.T) {
	router := setupComprehensiveRouter()
	populateAllMetrics(t, router)
	body := getComprehensiveMetrics(t, router)

	expectedMetrics := []string{
		// HTTP
		"catalogizer_http_requests_total",
		"catalogizer_http_request_duration_seconds",
		"catalogizer_http_active_connections",
		// Database
		"catalogizer_db_queries_total",
		"catalogizer_db_query_duration_seconds",
		"catalogizer_db_connections_active",
		"catalogizer_db_connections_idle",
		// Media
		"catalogizer_media_files_scanned_total",
		"catalogizer_media_files_analyzed_total",
		"catalogizer_media_analysis_duration_seconds",
		"catalogizer_media_by_type",
		// External API
		"catalogizer_external_api_calls_total",
		"catalogizer_external_api_call_duration_seconds",
		// Cache
		"catalogizer_cache_hits_total",
		"catalogizer_cache_misses_total",
		"catalogizer_cache_size_bytes",
		// WebSocket
		"catalogizer_websocket_connections",
		"catalogizer_websocket_connections_active",
		"catalogizer_websocket_messages_total",
		// Filesystem
		"catalogizer_filesystem_operations_total",
		"catalogizer_filesystem_operation_duration_seconds",
		// Storage
		"catalogizer_storage_roots_total",
		"catalogizer_storage_space_used_bytes",
		// Auth
		"catalogizer_auth_attempts_total",
		// Sessions
		"catalogizer_active_sessions",
		// Errors
		"catalogizer_errors_total",
		// Uptime
		"catalogizer_uptime_seconds",
		// Runtime
		"catalogizer_runtime_goroutines",
		"catalogizer_runtime_memory_alloc_bytes",
		"catalogizer_runtime_memory_sys_bytes",
		"catalogizer_runtime_memory_heap_inuse_bytes",
		// SMB Health
		"catalogizer_smb_health_status",
		// Go runtime (always present)
		"go_goroutines",
		"go_memstats_alloc_bytes",
	}

	for _, metric := range expectedMetrics {
		t.Run(metric, func(t *testing.T) {
			assert.Contains(t, body, metric,
				"metric %q should be present in /metrics output", metric)
		})
	}
}

// --- Metric naming convention enforcement ---

// TestComprehensive_AllCustomMetricsHaveCatalogizerPrefix verifies that every
// non-standard metric uses the "catalogizer_" namespace prefix.
func TestComprehensive_AllCustomMetricsHaveCatalogizerPrefix(t *testing.T) {
	router := setupComprehensiveRouter()
	populateAllMetrics(t, router)
	body := getComprehensiveMetrics(t, router)

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "# TYPE ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		metricName := parts[2]

		// Skip standard Go, process, and promhttp metrics.
		if strings.HasPrefix(metricName, "go_") ||
			strings.HasPrefix(metricName, "process_") ||
			strings.HasPrefix(metricName, "promhttp_") {
			continue
		}

		assert.True(t, strings.HasPrefix(metricName, "catalogizer_"),
			"custom metric %q should use 'catalogizer_' prefix", metricName)
	}
}

// --- Metric type correctness test ---

// TestComprehensive_MetricTypeDeclarations verifies that all key metrics have
// the correct TYPE declaration (counter, gauge, histogram).
func TestComprehensive_MetricTypeDeclarations(t *testing.T) {
	router := setupComprehensiveRouter()
	populateAllMetrics(t, router)
	body := getComprehensiveMetrics(t, router)

	typeChecks := []struct {
		metric     string
		metricType string
	}{
		// Counters
		{"catalogizer_http_requests_total", "counter"},
		{"catalogizer_db_queries_total", "counter"},
		{"catalogizer_media_files_scanned_total", "counter"},
		{"catalogizer_media_files_analyzed_total", "counter"},
		{"catalogizer_external_api_calls_total", "counter"},
		{"catalogizer_cache_hits_total", "counter"},
		{"catalogizer_cache_misses_total", "counter"},
		{"catalogizer_filesystem_operations_total", "counter"},
		{"catalogizer_auth_attempts_total", "counter"},
		{"catalogizer_errors_total", "counter"},
		{"catalogizer_websocket_messages_total", "counter"},
		{"catalogizer_uptime_seconds", "counter"},
		// Gauges
		{"catalogizer_http_active_connections", "gauge"},
		{"catalogizer_db_connections_active", "gauge"},
		{"catalogizer_db_connections_idle", "gauge"},
		{"catalogizer_active_sessions", "gauge"},
		{"catalogizer_runtime_goroutines", "gauge"},
		{"catalogizer_runtime_memory_alloc_bytes", "gauge"},
		{"catalogizer_smb_health_status", "gauge"},
		{"catalogizer_media_by_type", "gauge"},
		{"catalogizer_storage_roots_total", "gauge"},
		// Histograms
		{"catalogizer_http_request_duration_seconds", "histogram"},
		{"catalogizer_db_query_duration_seconds", "histogram"},
		{"catalogizer_media_analysis_duration_seconds", "histogram"},
		{"catalogizer_external_api_call_duration_seconds", "histogram"},
		{"catalogizer_filesystem_operation_duration_seconds", "histogram"},
	}

	for _, tc := range typeChecks {
		t.Run(tc.metric+"_is_"+tc.metricType, func(t *testing.T) {
			expected := "# TYPE " + tc.metric + " " + tc.metricType
			assert.Contains(t, body, expected,
				"metric %s should be declared as type %s", tc.metric, tc.metricType)
		})
	}
}
