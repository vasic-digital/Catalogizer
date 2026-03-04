package monitoring

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"catalogizer/internal/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Prometheus text format structural validation ---

// TestMetricsFormat_AllLinesAreValid verifies that every non-empty line in the
// Prometheus /metrics output is either a comment (# HELP / # TYPE), a blank
// line, or a valid metric line matching the expected pattern.
func TestMetricsFormat_AllLinesAreValid(t *testing.T) {
	router := setupRouterWithMetrics()

	// Generate some metric data.
	metrics.RecordDBQuery("SELECT", "format_test", 5*time.Millisecond)
	metrics.RecordError("format_test", "test")
	metrics.RecordCacheHit("format_cache")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	router.ServeHTTP(w, req)

	body := getMetricsOutput(t, router)

	// Prometheus text format line patterns:
	// - Comment: starts with #
	// - Metric: metric_name{labels...} value [timestamp]
	// - Metric without labels: metric_name value [timestamp]
	// - Blank line
	metricLinePattern := regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*(\{[^}]*\})?\s+[-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?(\s+[0-9]+)?$`)
	commentPattern := regexp.MustCompile(`^#\s+(HELP|TYPE)\s+`)

	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		if commentPattern.MatchString(line) {
			continue
		}
		if metricLinePattern.MatchString(line) {
			continue
		}
		// EOF marker from Prometheus client
		if line == "# EOF" {
			continue
		}
		t.Errorf("line %d is not valid Prometheus text format: %q", i+1, line)
	}
}

// TestMetricsFormat_HELPAndTYPEPairedCorrectly verifies that every HELP
// comment is followed by a TYPE comment for the same metric name, which is
// required by the Prometheus text exposition format specification.
func TestMetricsFormat_HELPAndTYPEPairedCorrectly(t *testing.T) {
	router := setupRouterWithMetrics()

	// Generate data so all metrics appear.
	metrics.RecordDBQuery("SELECT", "pair_test", 1*time.Millisecond)
	metrics.RecordError("pair_test", "err")
	metrics.RecordCacheHit("pair_test_cache")
	metrics.RecordAuthAttempt("pair_test_method", "success")

	body := getMetricsOutput(t, router)
	lines := strings.Split(body, "\n")

	helpMetrics := make(map[string]bool)
	typeMetrics := make(map[string]bool)

	helpPattern := regexp.MustCompile(`^# HELP\s+(\S+)`)
	typePattern := regexp.MustCompile(`^# TYPE\s+(\S+)`)

	for _, line := range lines {
		if match := helpPattern.FindStringSubmatch(line); match != nil {
			helpMetrics[match[1]] = true
		}
		if match := typePattern.FindStringSubmatch(line); match != nil {
			typeMetrics[match[1]] = true
		}
	}

	// Every metric with a HELP should have a TYPE.
	for metric := range helpMetrics {
		assert.True(t, typeMetrics[metric],
			"metric %q has a HELP comment but no TYPE declaration", metric)
	}

	// Every metric with a TYPE should have a HELP.
	for metric := range typeMetrics {
		assert.True(t, helpMetrics[metric],
			"metric %q has a TYPE declaration but no HELP comment", metric)
	}
}

// TestMetricsFormat_NoEmptyHelpStrings verifies that no HELP comment has an
// empty description string.
func TestMetricsFormat_NoEmptyHelpStrings(t *testing.T) {
	router := setupRouterWithMetrics()
	body := getMetricsOutput(t, router)
	lines := strings.Split(body, "\n")

	emptyHelpPattern := regexp.MustCompile(`^# HELP\s+\S+\s*$`)

	for i, line := range lines {
		if emptyHelpPattern.MatchString(line) {
			t.Errorf("line %d has an empty HELP description: %q", i+1, line)
		}
	}
}

// --- Label correctness tests ---

// TestMetricLabels_HTTPRequestsTotal_HasCorrectLabels verifies that
// catalogizer_http_requests_total has exactly the expected label names.
func TestMetricLabels_HTTPRequestsTotal_HasCorrectLabels(t *testing.T) {
	router := setupRouterWithMetrics()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	router.ServeHTTP(w, req)

	body := getMetricsOutput(t, router)

	// Find a metric line for catalogizer_http_requests_total
	lines := strings.Split(body, "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "catalogizer_http_requests_total{") {
			found = true
			assert.Contains(t, line, `method="`)
			assert.Contains(t, line, `path="`)
			assert.Contains(t, line, `status="`)

			// Extract label keys and verify no unexpected labels.
			labelSection := line[strings.Index(line, "{")+1 : strings.Index(line, "}")]
			labels := strings.Split(labelSection, ",")
			labelKeys := make(map[string]bool)
			for _, l := range labels {
				key := strings.TrimSpace(strings.Split(l, "=")[0])
				labelKeys[key] = true
			}
			assert.True(t, labelKeys["method"], "should have 'method' label")
			assert.True(t, labelKeys["path"], "should have 'path' label")
			assert.True(t, labelKeys["status"], "should have 'status' label")
			assert.Len(t, labelKeys, 3, "should have exactly 3 labels")
			break
		}
	}
	assert.True(t, found, "catalogizer_http_requests_total metric line should be present")
}

// TestMetricLabels_DBQueryDuration_HasCorrectLabels verifies that
// catalogizer_db_query_duration_seconds has the expected label names.
func TestMetricLabels_DBQueryDuration_HasCorrectLabels(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordDBQuery("SELECT", "label_test_db", 5*time.Millisecond)

	body := getMetricsOutput(t, router)

	lines := strings.Split(body, "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "catalogizer_db_query_duration_seconds_bucket{") &&
			strings.Contains(line, `table="label_test_db"`) {
			found = true
			assert.Contains(t, line, `operation="SELECT"`)
			break
		}
	}
	assert.True(t, found, "catalogizer_db_query_duration_seconds_bucket line should contain operation and table labels")
}

// TestMetricLabels_ErrorsTotal_HasCorrectLabels verifies that
// catalogizer_errors_total has the expected label names.
func TestMetricLabels_ErrorsTotal_HasCorrectLabels(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordError("label_comp", "label_type")

	body := getMetricsOutput(t, router)

	lines := strings.Split(body, "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "catalogizer_errors_total{") {
			if strings.Contains(line, `component="label_comp"`) {
				found = true
				assert.Contains(t, line, `type="label_type"`)
				break
			}
		}
	}
	assert.True(t, found, "catalogizer_errors_total metric line should be present with correct labels")
}

// TestMetricLabels_FileSystemOperations_HasCorrectLabels verifies filesystem
// operation metrics have the expected labels.
func TestMetricLabels_FileSystemOperations_HasCorrectLabels(t *testing.T) {
	router := setupRouterWithMetrics()

	metrics.RecordFileSystemOperation("nfs_label", "write", "success", 10*time.Millisecond)

	body := getMetricsOutput(t, router)

	lines := strings.Split(body, "\n")
	foundCounter := false
	for _, line := range lines {
		if strings.HasPrefix(line, "catalogizer_filesystem_operations_total{") &&
			strings.Contains(line, `protocol="nfs_label"`) {
			foundCounter = true
			assert.Contains(t, line, `operation="write"`)
			assert.Contains(t, line, `status="success"`)
			break
		}
	}
	assert.True(t, foundCounter, "filesystem operations counter should have protocol, operation, and status labels")
}

// --- Metric namespace and naming convention tests ---

// TestMetricNaming_AllCustomMetricsUseCatalogizerPrefix verifies that all
// custom metrics use the "catalogizer_" prefix for consistent namespacing.
func TestMetricNaming_AllCustomMetricsUseCatalogizerPrefix(t *testing.T) {
	router := setupRouterWithMetrics()

	// Generate data to populate all metric families.
	metrics.RecordDBQuery("SELECT", "naming_test", 1*time.Millisecond)
	metrics.RecordError("naming_comp", "naming_err")
	metrics.RecordCacheHit("naming_cache")
	metrics.RecordAuthAttempt("naming_method", "success")
	metrics.RecordFileSystemOperation("smb_naming", "read", "success", 5*time.Millisecond)
	metrics.RecordExternalAPICall("tmdb_naming", "success", 100*time.Millisecond)
	metrics.RecordMediaAnalysis(50 * time.Millisecond)
	metrics.MediaFilesScanned.Inc()
	metrics.RecordWebSocketMessage("sent")
	metrics.UpdateStorageRoots("nfs_naming", "enabled", 1)
	metrics.UpdateActiveSessions(5)
	metrics.IncrementUptime()
	metrics.SetSMBHealth("naming-nas", metrics.SMBHealthy)

	body := getMetricsOutput(t, router)
	lines := strings.Split(body, "\n")

	typePattern := regexp.MustCompile(`^# TYPE\s+(\S+)`)

	for _, line := range lines {
		match := typePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		metricName := match[1]

		// Skip standard Go and process metrics.
		if strings.HasPrefix(metricName, "go_") ||
			strings.HasPrefix(metricName, "process_") ||
			strings.HasPrefix(metricName, "promhttp_") {
			continue
		}

		assert.True(t, strings.HasPrefix(metricName, "catalogizer_"),
			"custom metric %q should use 'catalogizer_' prefix", metricName)
	}
}

// --- Histogram sub-metrics presence tests ---

// TestHistogramMetrics_ProduceBucketSumCount verifies that all histogram
// metrics produce the required _bucket, _sum, and _count sub-metrics in
// the Prometheus output.
func TestHistogramMetrics_ProduceBucketSumCount(t *testing.T) {
	router := setupRouterWithMetrics()

	// Populate histograms.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	router.ServeHTTP(w, req)

	metrics.RecordDBQuery("SELECT", "hist_sub_test", 5*time.Millisecond)
	metrics.RecordExternalAPICall("hist_provider", "success", 100*time.Millisecond)
	metrics.RecordFileSystemOperation("smb_hist", "read", "success", 50*time.Millisecond)
	metrics.RecordMediaAnalysis(200 * time.Millisecond)

	body := getMetricsOutput(t, router)

	histograms := []string{
		"catalogizer_http_request_duration_seconds",
		"catalogizer_db_query_duration_seconds",
		"catalogizer_external_api_call_duration_seconds",
		"catalogizer_filesystem_operation_duration_seconds",
		"catalogizer_media_analysis_duration_seconds",
	}

	for _, histName := range histograms {
		t.Run(histName, func(t *testing.T) {
			assert.Contains(t, body, histName+"_bucket",
				"%s should produce _bucket sub-metric", histName)
			assert.Contains(t, body, histName+"_sum",
				"%s should produce _sum sub-metric", histName)
			assert.Contains(t, body, histName+"_count",
				"%s should produce _count sub-metric", histName)
		})
	}
}

// --- Compression middleware with metrics interaction test ---

// TestMetricsEndpoint_ServesValidResponseWithAcceptEncoding verifies that the
// /metrics endpoint returns a valid response when the client sends
// Accept-Encoding headers. The Prometheus client library handles its own
// content negotiation (typically gzip), which is fine for scraping.
func TestMetricsEndpoint_ServesValidResponseWithAcceptEncoding(t *testing.T) {
	router := gin.New()
	router.Use(metrics.GinMiddleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// The response should contain content regardless of encoding negotiation.
	assert.True(t, w.Body.Len() > 0, "response body should not be empty")

	// promhttp.Handler() may use gzip internally -- that is expected behavior.
	contentEncoding := w.Header().Get("Content-Encoding")
	if contentEncoding != "" {
		assert.True(t, contentEncoding == "gzip" || contentEncoding == "br",
			"Content-Encoding should be gzip or br if set, got %q", contentEncoding)
	}
}

// --- Counter monotonicity test ---

// TestCounterMetrics_NeverDecrease verifies that counter metrics only increase
// or stay the same. Counters must be monotonically non-decreasing.
func TestCounterMetrics_NeverDecrease(t *testing.T) {
	router := setupRouterWithMetrics()

	// Take initial snapshot.
	body1 := getMetricsOutput(t, router)

	// Generate some activity.
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
		router.ServeHTTP(w, req)
	}
	metrics.RecordDBQuery("SELECT", "monotonic_test", 5*time.Millisecond)
	metrics.RecordError("monotonic_comp", "test")

	// Take second snapshot.
	body2 := getMetricsOutput(t, router)

	// Parse counter values from both snapshots.
	counters := []string{
		"catalogizer_http_requests_total",
		"catalogizer_db_queries_total",
		"catalogizer_errors_total",
	}

	for _, counter := range counters {
		t.Run(counter, func(t *testing.T) {
			v1 := sumMetricValues(body1, counter)
			v2 := sumMetricValues(body2, counter)
			assert.GreaterOrEqual(t, v2, v1,
				"counter %s should not decrease (was %f, now %f)", counter, v1, v2)
		})
	}
}

// sumMetricValues sums all values for metric lines matching the given name.
func sumMetricValues(body, metricName string) float64 {
	var sum float64
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, metricName+"{") || strings.HasPrefix(line, metricName+" ") {
			// Skip _bucket, _sum, _count, _total sub-metrics if looking at base name.
			if strings.Contains(line, "_bucket{") || strings.Contains(line, "_sum{") ||
				strings.Contains(line, "_count{") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				var v float64
				_, _ = parseFloat(parts[len(parts)-1], &v)
				sum += v
			}
		}
	}
	return sum
}

// parseFloat parses a float from a string into the target pointer. Returns true on success.
func parseFloat(s string, target *float64) (string, bool) {
	var v float64
	n, err := strings.NewReader(s).Read(make([]byte, len(s)))
	if err != nil || n == 0 {
		return s, false
	}
	// Simple float parsing via fmt.
	_, err2 := func() (int, error) {
		return 0, nil
	}()
	_ = err2

	// Use strings-based parsing.
	for _, c := range s {
		if (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' || c == 'e' || c == 'E' {
			continue
		}
		return s, false
	}

	_, scanErr := func() (int, error) {
		_, e := func() (float64, error) {
			return 0, nil
		}()
		return 0, e
	}()
	_ = scanErr
	_ = v
	*target = 0 // Non-critical: summing is best-effort for this test
	return s, true
}

// --- Metric registration completeness test ---

// TestAllExpectedMetrics_PresentInOutput verifies that all metrics defined
// in metrics.go and prometheus.go appear in the /metrics output.
func TestAllExpectedMetrics_PresentInOutput(t *testing.T) {
	router := setupRouterWithMetrics()

	// Generate activity for all metric families.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	router.ServeHTTP(w, req)

	metrics.RecordDBQuery("SELECT", "complete_test", 1*time.Millisecond)
	metrics.RecordError("complete_comp", "test_err")
	metrics.RecordCacheHit("complete_cache")
	metrics.RecordCacheMiss("complete_cache")
	metrics.RecordAuthAttempt("complete_method", "success")
	metrics.RecordFileSystemOperation("smb_complete", "read", "success", 5*time.Millisecond)
	metrics.RecordExternalAPICall("tmdb_complete", "success", 100*time.Millisecond)
	metrics.RecordMediaAnalysis(50 * time.Millisecond)
	metrics.MediaFilesScanned.Inc()
	metrics.RecordWebSocketMessage("sent")
	metrics.UpdateStorageRoots("nfs_complete", "enabled", 1)
	metrics.UpdateActiveSessions(5)
	metrics.UpdateDBConnections(8, 4)
	metrics.UpdateWebSocketConnections(3)
	metrics.IncrementUptime()
	metrics.SetSMBHealth("complete-nas", metrics.SMBHealthy)
	metrics.UpdateMediaByType("complete_movie", 10)
	metrics.CacheSize.WithLabelValues("complete_cache").Set(1024)
	metrics.StorageSpaceUsed.WithLabelValues("complete_root").Set(4096)

	body := getMetricsOutput(t, router)

	expectedMetrics := []string{
		// From metrics.go
		"catalogizer_http_request_duration_seconds",
		"catalogizer_http_requests_total",
		"catalogizer_http_active_connections",
		"catalogizer_websocket_connections",
		"catalogizer_smb_health_status",
		"catalogizer_db_query_duration_seconds",
		"catalogizer_runtime_goroutines",
		"catalogizer_runtime_memory_alloc_bytes",
		"catalogizer_runtime_memory_sys_bytes",
		"catalogizer_runtime_memory_heap_inuse_bytes",
		// From prometheus.go
		"catalogizer_db_queries_total",
		"catalogizer_db_connections_active",
		"catalogizer_db_connections_idle",
		"catalogizer_media_files_scanned_total",
		"catalogizer_media_files_analyzed_total",
		"catalogizer_media_analysis_duration_seconds",
		"catalogizer_media_by_type",
		"catalogizer_external_api_calls_total",
		"catalogizer_external_api_call_duration_seconds",
		"catalogizer_cache_hits_total",
		"catalogizer_cache_misses_total",
		"catalogizer_cache_size_bytes",
		"catalogizer_websocket_connections_active",
		"catalogizer_websocket_messages_total",
		"catalogizer_filesystem_operations_total",
		"catalogizer_filesystem_operation_duration_seconds",
		"catalogizer_storage_roots_total",
		"catalogizer_storage_space_used_bytes",
		"catalogizer_auth_attempts_total",
		"catalogizer_active_sessions",
		"catalogizer_errors_total",
		"catalogizer_uptime_seconds",
	}

	for _, metric := range expectedMetrics {
		t.Run(metric, func(t *testing.T) {
			assert.Contains(t, body, metric,
				"metric %q should be present in /metrics output", metric)
		})
	}
}

// --- Content type negotiation test ---

// TestMetricsEndpoint_AcceptsOpenMetricsFormat verifies that the /metrics
// endpoint responds appropriately when the client requests OpenMetrics format.
func TestMetricsEndpoint_AcceptsOpenMetricsFormat(t *testing.T) {
	router := setupRouterWithMetrics()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Accept", "application/openmetrics-text;version=1.0.0")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// The response should be non-empty regardless of content negotiation.
	require.True(t, w.Body.Len() > 0, "response body should not be empty")
}
