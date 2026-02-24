package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- Helper to read a CounterVec label combination ---

func getCounterVecValue(cv interface{ WithLabelValues(lvs ...string) interface{ Write(*dto.Metric) error } }, labels ...string) float64 {
	m := &dto.Metric{}
	cv.WithLabelValues(labels...).Write(m)
	return m.GetCounter().GetValue()
}

// TestHTTPRequestCounter_IncrementsOnAPICalls verifies that the
// catalogizer_http_requests_total counter increments after serving an API
// request through the Gin middleware.
func TestHTTPRequestCounter_IncrementsOnAPICalls(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())

	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/api/v1/items", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": 1})
	})
	router.DELETE("/api/v1/items/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus string
		routePattern   string
	}{
		{
			name:           "GET 200 increments counter",
			method:         "GET",
			path:           "/api/v1/health",
			expectedStatus: "200",
			routePattern:   "/api/v1/health",
		},
		{
			name:           "POST 201 increments counter",
			method:         "POST",
			path:           "/api/v1/items",
			expectedStatus: "201",
			routePattern:   "/api/v1/items",
		},
		{
			name:           "DELETE 204 increments counter",
			method:         "DELETE",
			path:           "/api/v1/items/42",
			expectedStatus: "204",
			routePattern:   "/api/v1/items/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := getCounterValue(HTTPRequestsTotal.WithLabelValues(tt.method, tt.routePattern, tt.expectedStatus))

			w := httptest.NewRecorder()
			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)
			router.ServeHTTP(w, req)

			after := getCounterValue(HTTPRequestsTotal.WithLabelValues(tt.method, tt.routePattern, tt.expectedStatus))
			assert.Equal(t, before+1, after, "counter should have incremented by 1")
		})
	}
}

// TestHTTPRequestCounter_MultipleRequestsAccumulate verifies that making
// multiple requests causes the counter to accumulate correctly.
func TestHTTPRequestCounter_MultipleRequestsAccumulate(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/accumulate-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	before := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/accumulate-test", "200"))

	requestCount := 5
	for i := 0; i < requestCount; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/accumulate-test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	after := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/accumulate-test", "200"))
	assert.Equal(t, before+float64(requestCount), after, "counter should accumulate across multiple requests")
}

// TestHTTPRequestDuration_RecordsLatencies verifies that the
// catalogizer_http_request_duration_seconds histogram records observations
// when requests are served.
func TestHTTPRequestDuration_RecordsLatencies(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/latency-test", func(c *gin.Context) {
		// Introduce a small delay to produce a non-zero duration.
		time.Sleep(1 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	beforeCount := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/latency-test", "200"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/latency-test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	afterCount := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/latency-test", "200"))
	assert.Equal(t, beforeCount+1, afterCount, "histogram observation count should increment by 1")

	// Verify the observed sum is > 0 (the request took some time).
	m := &dto.Metric{}
	HTTPRequestDuration.WithLabelValues("GET", "/api/v1/latency-test", "200").(interface{ Write(*dto.Metric) error }).Write(m)
	assert.Greater(t, m.GetHistogram().GetSampleSum(), float64(0), "histogram sum should be positive")
}

// TestHTTPRequestDuration_DifferentStatusCodesTrackedSeparately verifies that
// the histogram tracks different status codes as separate series.
func TestHTTPRequestDuration_DifferentStatusCodesTrackedSeparately(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/status-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	// Unmatched route returns 404.

	before200 := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/status-test", "200"))
	before404 := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/unknown-route", "404"))

	// Make a successful request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/status-test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Make a 404 request.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/unknown-route", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)

	after200 := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/status-test", "200"))
	after404 := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/unknown-route", "404"))

	assert.Equal(t, before200+1, after200, "200 histogram should increment")
	assert.Equal(t, before404+1, after404, "404 histogram should increment")
}

// TestErrorCounter_IncrementsOnErrorResponses verifies that the
// catalogizer_errors_total counter increments when RecordError is called
// during error response handling.
func TestErrorCounter_IncrementsOnErrorResponses(t *testing.T) {
	tests := []struct {
		name      string
		component string
		errType   string
	}{
		{"api validation error", "api", "validation"},
		{"database connection error", "database", "connection"},
		{"auth unauthorized error", "auth", "unauthorized"},
		{"filesystem permission error", "filesystem", "permission"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := getCounterValue(ErrorsTotal.WithLabelValues(tt.component, tt.errType))

			RecordError(tt.component, tt.errType)

			after := getCounterValue(ErrorsTotal.WithLabelValues(tt.component, tt.errType))
			assert.Equal(t, before+1, after, "error counter should increment by 1")
		})
	}
}

// TestErrorCounter_AccumulatesMultipleErrors verifies that multiple errors of
// the same type accumulate correctly.
func TestErrorCounter_AccumulatesMultipleErrors(t *testing.T) {
	before := getCounterValue(ErrorsTotal.WithLabelValues("api", "timeout"))

	for i := 0; i < 10; i++ {
		RecordError("api", "timeout")
	}

	after := getCounterValue(ErrorsTotal.WithLabelValues("api", "timeout"))
	assert.Equal(t, before+10, after, "error counter should accumulate 10 errors")
}

// TestErrorCounter_DifferentComponentsTrackedSeparately verifies that errors
// in different components are tracked independently.
func TestErrorCounter_DifferentComponentsTrackedSeparately(t *testing.T) {
	beforeAPI := getCounterValue(ErrorsTotal.WithLabelValues("api_sep", "error"))
	beforeDB := getCounterValue(ErrorsTotal.WithLabelValues("db_sep", "error"))

	RecordError("api_sep", "error")

	afterAPI := getCounterValue(ErrorsTotal.WithLabelValues("api_sep", "error"))
	afterDB := getCounterValue(ErrorsTotal.WithLabelValues("db_sep", "error"))

	assert.Equal(t, beforeAPI+1, afterAPI, "api error counter should increment")
	assert.Equal(t, beforeDB, afterDB, "db error counter should not change")
}

// TestCustomMetrics_AreRegistered verifies that all expected custom metrics
// are properly registered with the Prometheus default registry.
func TestCustomMetrics_AreRegistered(t *testing.T) {
	// HTTP metrics from metrics.go
	t.Run("HTTPRequestsTotal is registered", func(t *testing.T) {
		assert.NotNil(t, HTTPRequestsTotal)
		// Verify it accepts correct label cardinality (method, path, status).
		assert.NotPanics(t, func() {
			HTTPRequestsTotal.WithLabelValues("GET", "/test", "200")
		})
	})

	t.Run("HTTPRequestDuration is registered", func(t *testing.T) {
		assert.NotNil(t, HTTPRequestDuration)
		assert.NotPanics(t, func() {
			HTTPRequestDuration.WithLabelValues("GET", "/test", "200")
		})
	})

	t.Run("HTTPActiveConnections is registered", func(t *testing.T) {
		assert.NotNil(t, HTTPActiveConnections)
		assert.NotPanics(t, func() {
			HTTPActiveConnections.Inc()
			HTTPActiveConnections.Dec()
		})
	})

	// Database metrics
	t.Run("DBQueryTotal is registered", func(t *testing.T) {
		assert.NotNil(t, DBQueryTotal)
		assert.NotPanics(t, func() {
			DBQueryTotal.WithLabelValues("SELECT", "files")
		})
	})

	t.Run("DBQueryDuration is registered", func(t *testing.T) {
		assert.NotNil(t, DBQueryDuration)
		assert.NotPanics(t, func() {
			DBQueryDuration.WithLabelValues("SELECT", "files")
		})
	})

	t.Run("DBConnectionsActive is registered", func(t *testing.T) {
		assert.NotNil(t, DBConnectionsActive)
	})

	t.Run("DBConnectionsIdle is registered", func(t *testing.T) {
		assert.NotNil(t, DBConnectionsIdle)
	})

	// Media metrics
	t.Run("MediaFilesScanned is registered", func(t *testing.T) {
		assert.NotNil(t, MediaFilesScanned)
	})

	t.Run("MediaFilesAnalyzed is registered", func(t *testing.T) {
		assert.NotNil(t, MediaFilesAnalyzed)
	})

	t.Run("MediaAnalysisDuration is registered", func(t *testing.T) {
		assert.NotNil(t, MediaAnalysisDuration)
	})

	t.Run("MediaByType is registered", func(t *testing.T) {
		assert.NotNil(t, MediaByType)
		assert.NotPanics(t, func() {
			MediaByType.WithLabelValues("movie")
		})
	})

	// External API metrics
	t.Run("ExternalAPICallsTotal is registered", func(t *testing.T) {
		assert.NotNil(t, ExternalAPICallsTotal)
		assert.NotPanics(t, func() {
			ExternalAPICallsTotal.WithLabelValues("tmdb", "success")
		})
	})

	t.Run("ExternalAPICallDuration is registered", func(t *testing.T) {
		assert.NotNil(t, ExternalAPICallDuration)
		assert.NotPanics(t, func() {
			ExternalAPICallDuration.WithLabelValues("tmdb")
		})
	})

	// Cache metrics
	t.Run("CacheHits is registered", func(t *testing.T) {
		assert.NotNil(t, CacheHits)
		assert.NotPanics(t, func() {
			CacheHits.WithLabelValues("memory")
		})
	})

	t.Run("CacheMisses is registered", func(t *testing.T) {
		assert.NotNil(t, CacheMisses)
		assert.NotPanics(t, func() {
			CacheMisses.WithLabelValues("memory")
		})
	})

	t.Run("CacheSize is registered", func(t *testing.T) {
		assert.NotNil(t, CacheSize)
		assert.NotPanics(t, func() {
			CacheSize.WithLabelValues("memory")
		})
	})

	// WebSocket metrics
	t.Run("WebSocketConnections is registered (metrics.go)", func(t *testing.T) {
		assert.NotNil(t, WebSocketConnections)
	})

	t.Run("WebSocketConnectionsActive is registered (prometheus.go)", func(t *testing.T) {
		assert.NotNil(t, WebSocketConnectionsActive)
	})

	t.Run("WebSocketMessagesTotal is registered", func(t *testing.T) {
		assert.NotNil(t, WebSocketMessagesTotal)
		assert.NotPanics(t, func() {
			WebSocketMessagesTotal.WithLabelValues("sent")
			WebSocketMessagesTotal.WithLabelValues("received")
		})
	})

	// File system metrics
	t.Run("FileSystemOperationsTotal is registered", func(t *testing.T) {
		assert.NotNil(t, FileSystemOperationsTotal)
		assert.NotPanics(t, func() {
			FileSystemOperationsTotal.WithLabelValues("smb", "read", "success")
		})
	})

	t.Run("FileSystemOperationDuration is registered", func(t *testing.T) {
		assert.NotNil(t, FileSystemOperationDuration)
		assert.NotPanics(t, func() {
			FileSystemOperationDuration.WithLabelValues("smb", "read")
		})
	})

	// Storage metrics
	t.Run("StorageRootsTotal is registered", func(t *testing.T) {
		assert.NotNil(t, StorageRootsTotal)
		assert.NotPanics(t, func() {
			StorageRootsTotal.WithLabelValues("smb", "enabled")
		})
	})

	t.Run("StorageSpaceUsed is registered", func(t *testing.T) {
		assert.NotNil(t, StorageSpaceUsed)
		assert.NotPanics(t, func() {
			StorageSpaceUsed.WithLabelValues("root-1")
		})
	})

	// Auth metrics
	t.Run("AuthAttemptsTotal is registered", func(t *testing.T) {
		assert.NotNil(t, AuthAttemptsTotal)
		assert.NotPanics(t, func() {
			AuthAttemptsTotal.WithLabelValues("password", "success")
		})
	})

	t.Run("ActiveSessions is registered", func(t *testing.T) {
		assert.NotNil(t, ActiveSessions)
	})

	// Error metrics
	t.Run("ErrorsTotal is registered", func(t *testing.T) {
		assert.NotNil(t, ErrorsTotal)
		assert.NotPanics(t, func() {
			ErrorsTotal.WithLabelValues("api", "validation")
		})
	})

	// Runtime metrics
	t.Run("GoroutineCount is registered", func(t *testing.T) {
		assert.NotNil(t, GoroutineCount)
	})

	t.Run("MemoryAlloc is registered", func(t *testing.T) {
		assert.NotNil(t, MemoryAlloc)
	})

	t.Run("MemorySys is registered", func(t *testing.T) {
		assert.NotNil(t, MemorySys)
	})

	t.Run("MemoryHeapInuse is registered", func(t *testing.T) {
		assert.NotNil(t, MemoryHeapInuse)
	})

	// SMB health
	t.Run("SMBHealthStatus is registered", func(t *testing.T) {
		assert.NotNil(t, SMBHealthStatus)
		assert.NotPanics(t, func() {
			SMBHealthStatus.WithLabelValues("test-source")
		})
	})

	// System metrics
	t.Run("UptimeSeconds is registered", func(t *testing.T) {
		assert.NotNil(t, UptimeSeconds)
	})
}

// TestMiddlewareMetrics_EndToEnd verifies the full pipeline: a request through
// the Gin middleware increments both the counter and the histogram, and tracks
// active connections correctly.
func TestMiddlewareMetrics_EndToEnd(t *testing.T) {
	var connDuringHandler float64

	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/e2e-test", func(c *gin.Context) {
		connDuringHandler = getGaugeValue(HTTPActiveConnections)
		time.Sleep(1 * time.Millisecond) // small delay for duration
		c.JSON(http.StatusOK, gin.H{"result": "success"})
	})

	beforeCounter := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/e2e-test", "200"))
	beforeHistogram := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/e2e-test", "200"))
	beforeConn := getGaugeValue(HTTPActiveConnections)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/e2e-test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Counter incremented.
	afterCounter := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/e2e-test", "200"))
	assert.Equal(t, beforeCounter+1, afterCounter, "request counter should increment")

	// Histogram observation added.
	afterHistogram := getHistogramCount(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/e2e-test", "200"))
	assert.Equal(t, beforeHistogram+1, afterHistogram, "duration histogram should record an observation")

	// Active connections incremented during the handler.
	assert.Equal(t, beforeConn+1, connDuringHandler, "active connections should be elevated during request")

	// Active connections returned to previous value after the handler.
	afterConn := getGaugeValue(HTTPActiveConnections)
	assert.Equal(t, beforeConn, afterConn, "active connections should return to baseline after request")
}

// TestMiddlewareMetrics_ErrorResponsesRecorded verifies that error status codes
// (4xx, 5xx) are tracked by the middleware.
func TestMiddlewareMetrics_ErrorResponsesRecorded(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/bad-request", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	})
	router.GET("/api/v1/internal-error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})
	router.GET("/api/v1/forbidden", func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	})

	tests := []struct {
		name   string
		path   string
		status string
	}{
		{"400 bad request", "/api/v1/bad-request", "400"},
		{"500 internal error", "/api/v1/internal-error", "500"},
		{"403 forbidden", "/api/v1/forbidden", "403"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", tt.path, tt.status))

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			after := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", tt.path, tt.status))
			assert.Equal(t, before+1, after, "counter should track %s status", tt.status)
		})
	}
}

// TestRecordDBQuery_IntegrationWithHistogramAndCounter verifies that RecordDBQuery
// increments both the DBQueryTotal counter and the DBQueryDuration histogram.
func TestRecordDBQuery_IntegrationWithHistogramAndCounter(t *testing.T) {
	beforeCount := getCounterValue(DBQueryTotal.WithLabelValues("SELECT", "integration_test_table"))
	beforeHistCount := getHistogramCount(DBQueryDuration.WithLabelValues("SELECT", "integration_test_table"))

	RecordDBQuery("SELECT", "integration_test_table", 25*time.Millisecond)

	afterCount := getCounterValue(DBQueryTotal.WithLabelValues("SELECT", "integration_test_table"))
	afterHistCount := getHistogramCount(DBQueryDuration.WithLabelValues("SELECT", "integration_test_table"))

	assert.Equal(t, beforeCount+1, afterCount, "DBQueryTotal should increment")
	assert.Equal(t, beforeHistCount+1, afterHistCount, "DBQueryDuration should record observation")
}

// TestRecordExternalAPICall_IntegrationWithHistogramAndCounter verifies that
// RecordExternalAPICall increments both the counter and the histogram.
func TestRecordExternalAPICall_IntegrationWithHistogramAndCounter(t *testing.T) {
	beforeCalls := getCounterValue(ExternalAPICallsTotal.WithLabelValues("tmdb_int", "success"))
	beforeDur := getHistogramCount(ExternalAPICallDuration.WithLabelValues("tmdb_int"))

	RecordExternalAPICall("tmdb_int", "success", 150*time.Millisecond)

	afterCalls := getCounterValue(ExternalAPICallsTotal.WithLabelValues("tmdb_int", "success"))
	afterDur := getHistogramCount(ExternalAPICallDuration.WithLabelValues("tmdb_int"))

	assert.Equal(t, beforeCalls+1, afterCalls, "ExternalAPICallsTotal should increment")
	assert.Equal(t, beforeDur+1, afterDur, "ExternalAPICallDuration should record observation")
}

// TestRecordFileSystemOperation_IntegrationWithCounterAndHistogram verifies that
// RecordFileSystemOperation updates both the counter and the histogram.
func TestRecordFileSystemOperation_IntegrationWithCounterAndHistogram(t *testing.T) {
	beforeOps := getCounterValue(FileSystemOperationsTotal.WithLabelValues("smb_int", "read", "success"))
	beforeDur := getHistogramCount(FileSystemOperationDuration.WithLabelValues("smb_int", "read"))

	RecordFileSystemOperation("smb_int", "read", "success", 50*time.Millisecond)

	afterOps := getCounterValue(FileSystemOperationsTotal.WithLabelValues("smb_int", "read", "success"))
	afterDur := getHistogramCount(FileSystemOperationDuration.WithLabelValues("smb_int", "read"))

	assert.Equal(t, beforeOps+1, afterOps, "FileSystemOperationsTotal should increment")
	assert.Equal(t, beforeDur+1, afterDur, "FileSystemOperationDuration should record observation")
}

// TestMetricsEndpointExcluded verifies that the /metrics path itself is not
// counted by the metrics middleware (to avoid self-referencing noise).
func TestMetricsEndpointExcluded(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "metrics output")
	})

	beforeCount := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/metrics", "200"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	afterCount := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/metrics", "200"))
	assert.Equal(t, beforeCount, afterCount, "/metrics should be excluded from tracking")
}

// TestGaugeMetrics_SetAndRead verifies that gauge metrics can be set and read.
func TestGaugeMetrics_SetAndRead(t *testing.T) {
	t.Run("ActiveSessions gauge", func(t *testing.T) {
		UpdateActiveSessions(42)
		m := &dto.Metric{}
		ActiveSessions.Write(m)
		assert.Equal(t, float64(42), m.GetGauge().GetValue())
	})

	t.Run("WebSocketConnectionsActive gauge", func(t *testing.T) {
		UpdateWebSocketConnections(7)
		m := &dto.Metric{}
		WebSocketConnectionsActive.Write(m)
		assert.Equal(t, float64(7), m.GetGauge().GetValue())
	})

	t.Run("DBConnections gauges", func(t *testing.T) {
		UpdateDBConnections(15, 5)

		mActive := &dto.Metric{}
		DBConnectionsActive.Write(mActive)
		assert.Equal(t, float64(15), mActive.GetGauge().GetValue())

		mIdle := &dto.Metric{}
		DBConnectionsIdle.Write(mIdle)
		assert.Equal(t, float64(5), mIdle.GetGauge().GetValue())
	})

	t.Run("MediaByType gauge", func(t *testing.T) {
		UpdateMediaByType("movie_int", 250)
		m := &dto.Metric{}
		MediaByType.WithLabelValues("movie_int").Write(m)
		assert.Equal(t, float64(250), m.GetGauge().GetValue())
	})

	t.Run("StorageRootsTotal gauge", func(t *testing.T) {
		UpdateStorageRoots("nfs_int", "enabled", 3)
		m := &dto.Metric{}
		StorageRootsTotal.WithLabelValues("nfs_int", "enabled").Write(m)
		assert.Equal(t, float64(3), m.GetGauge().GetValue())
	})
}

// TestCacheMetrics_HitsAndMisses verifies that cache hit and miss counters
// increment correctly and independently.
func TestCacheMetrics_HitsAndMisses(t *testing.T) {
	beforeHits := getCounterValue(CacheHits.WithLabelValues("integration_test"))
	beforeMisses := getCounterValue(CacheMisses.WithLabelValues("integration_test"))

	RecordCacheHit("integration_test")
	RecordCacheHit("integration_test")
	RecordCacheMiss("integration_test")

	afterHits := getCounterValue(CacheHits.WithLabelValues("integration_test"))
	afterMisses := getCounterValue(CacheMisses.WithLabelValues("integration_test"))

	assert.Equal(t, beforeHits+2, afterHits, "cache hits should increment by 2")
	assert.Equal(t, beforeMisses+1, afterMisses, "cache misses should increment by 1")
}

// TestAuthMetrics_AttemptsTracked verifies that authentication attempt metrics
// track successes and failures independently.
func TestAuthMetrics_AttemptsTracked(t *testing.T) {
	beforeSuccess := getCounterValue(AuthAttemptsTotal.WithLabelValues("password_int", "success"))
	beforeFailure := getCounterValue(AuthAttemptsTotal.WithLabelValues("password_int", "failure"))

	RecordAuthAttempt("password_int", "success")
	RecordAuthAttempt("password_int", "success")
	RecordAuthAttempt("password_int", "failure")

	afterSuccess := getCounterValue(AuthAttemptsTotal.WithLabelValues("password_int", "success"))
	afterFailure := getCounterValue(AuthAttemptsTotal.WithLabelValues("password_int", "failure"))

	assert.Equal(t, beforeSuccess+2, afterSuccess, "success attempts should increment by 2")
	assert.Equal(t, beforeFailure+1, afterFailure, "failure attempts should increment by 1")
}

// TestMediaAnalysisMetrics_CounterAndHistogram verifies that RecordMediaAnalysis
// increments the counter and records in the histogram.
func TestMediaAnalysisMetrics_CounterAndHistogram(t *testing.T) {
	beforeCount := getCounterValue(MediaFilesAnalyzed)
	beforeHist := getHistogramCount(MediaAnalysisDuration)

	RecordMediaAnalysis(500 * time.Millisecond)

	afterCount := getCounterValue(MediaFilesAnalyzed)
	afterHist := getHistogramCount(MediaAnalysisDuration)

	assert.Equal(t, beforeCount+1, afterCount, "MediaFilesAnalyzed should increment")
	assert.Equal(t, beforeHist+1, afterHist, "MediaAnalysisDuration should record observation")
}

// TestWebSocketMessageMetrics verifies that WebSocket message counters track
// sent and received directions independently.
func TestWebSocketMessageMetrics(t *testing.T) {
	beforeSent := getCounterValue(WebSocketMessagesTotal.WithLabelValues("sent"))
	beforeReceived := getCounterValue(WebSocketMessagesTotal.WithLabelValues("received"))

	RecordWebSocketMessage("sent")
	RecordWebSocketMessage("sent")
	RecordWebSocketMessage("received")

	afterSent := getCounterValue(WebSocketMessagesTotal.WithLabelValues("sent"))
	afterReceived := getCounterValue(WebSocketMessagesTotal.WithLabelValues("received"))

	assert.Equal(t, beforeSent+2, afterSent, "sent counter should increment by 2")
	assert.Equal(t, beforeReceived+1, afterReceived, "received counter should increment by 1")
}

// TestMetricsMiddleware_ConcurrentRequests verifies that metrics are safely
// updated under concurrent request load.
func TestMetricsMiddleware_ConcurrentRequests(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/concurrent-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	before := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/concurrent-test", "200"))

	concurrency := 20
	done := make(chan struct{}, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/concurrent-test", nil)
			router.ServeHTTP(w, req)
		}()
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}

	after := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/concurrent-test", "200"))
	assert.Equal(t, before+float64(concurrency), after, "counter should be correct after concurrent requests")
}

// TestHTTPMethodsTrackedSeparately verifies that different HTTP methods on the
// same path are tracked as separate counter/histogram series.
func TestHTTPMethodsTrackedSeparately(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/methods-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "GET"})
	})
	router.POST("/api/v1/methods-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "POST"})
	})

	beforeGET := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/methods-test", "200"))
	beforePOST := getCounterValue(HTTPRequestsTotal.WithLabelValues("POST", "/api/v1/methods-test", "200"))

	// Make GET request.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/v1/methods-test", nil)
	router.ServeHTTP(w1, req1)

	afterGET := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/methods-test", "200"))
	afterPOST := getCounterValue(HTTPRequestsTotal.WithLabelValues("POST", "/api/v1/methods-test", "200"))

	assert.Equal(t, beforeGET+1, afterGET, "GET counter should increment")
	assert.Equal(t, beforePOST, afterPOST, "POST counter should not change for a GET request")

	// Make POST request.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/v1/methods-test", strings.NewReader("{}"))
	router.ServeHTTP(w2, req2)

	finalPOST := getCounterValue(HTTPRequestsTotal.WithLabelValues("POST", "/api/v1/methods-test", "200"))
	assert.Equal(t, beforePOST+1, finalPOST, "POST counter should now increment")
}
