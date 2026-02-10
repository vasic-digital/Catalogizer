package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for the Catalogizer application
var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "catalogizer_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Database Metrics
	DBQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "catalogizer_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "catalogizer_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "catalogizer_db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// Media Processing Metrics
	MediaFilesScanned = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "catalogizer_media_files_scanned_total",
			Help: "Total number of media files scanned",
		},
	)

	MediaFilesAnalyzed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "catalogizer_media_files_analyzed_total",
			Help: "Total number of media files analyzed",
		},
	)

	MediaAnalysisDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "catalogizer_media_analysis_duration_seconds",
			Help:    "Media analysis duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
	)

	MediaByType = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "catalogizer_media_by_type",
			Help: "Number of media items by type",
		},
		[]string{"type"},
	)

	// External API Metrics
	ExternalAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_external_api_calls_total",
			Help: "Total number of external API calls",
		},
		[]string{"provider", "status"},
	)

	ExternalAPICallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "catalogizer_external_api_call_duration_seconds",
			Help:    "External API call duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"provider"},
	)

	// Cache Metrics
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	CacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "catalogizer_cache_size_bytes",
			Help: "Cache size in bytes",
		},
		[]string{"cache_type"},
	)

	// WebSocket Metrics
	WebSocketConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "catalogizer_websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)

	WebSocketMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_websocket_messages_total",
			Help: "Total number of WebSocket messages",
		},
		[]string{"direction"}, // "sent" or "received"
	)

	// File System Metrics
	FileSystemOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_filesystem_operations_total",
			Help: "Total number of filesystem operations",
		},
		[]string{"protocol", "operation", "status"},
	)

	FileSystemOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "catalogizer_filesystem_operation_duration_seconds",
			Help:    "Filesystem operation duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		},
		[]string{"protocol", "operation"},
	)

	// Storage Metrics
	StorageRootsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "catalogizer_storage_roots_total",
			Help: "Number of storage roots by protocol",
		},
		[]string{"protocol", "status"}, // status: "enabled" or "disabled"
	)

	StorageSpaceUsed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "catalogizer_storage_space_used_bytes",
			Help: "Storage space used in bytes",
		},
		[]string{"root_id"},
	)

	// Authentication Metrics
	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "status"}, // method: "password", "token", etc. status: "success", "failure"
	)

	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "catalogizer_active_sessions",
			Help: "Number of active user sessions",
		},
	)

	// Error Metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalogizer_errors_total",
			Help: "Total number of errors",
		},
		[]string{"component", "type"},
	)

	// System Metrics
	UptimeSeconds = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "catalogizer_uptime_seconds",
			Help: "Application uptime in seconds",
		},
	)
)

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, path, status string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordDBQuery records a database query metric
func RecordDBQuery(operation, table string, duration time.Duration) {
	DBQueryTotal.WithLabelValues(operation, table).Inc()
	DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordMediaAnalysis records a media analysis metric
func RecordMediaAnalysis(duration time.Duration) {
	MediaFilesAnalyzed.Inc()
	MediaAnalysisDuration.Observe(duration.Seconds())
}

// RecordExternalAPICall records an external API call metric
func RecordExternalAPICall(provider, status string, duration time.Duration) {
	ExternalAPICallsTotal.WithLabelValues(provider, status).Inc()
	ExternalAPICallDuration.WithLabelValues(provider).Observe(duration.Seconds())
}

// RecordCacheHit records a cache hit
func RecordCacheHit(cacheType string) {
	CacheHits.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(cacheType string) {
	CacheMisses.WithLabelValues(cacheType).Inc()
}

// RecordFileSystemOperation records a filesystem operation metric
func RecordFileSystemOperation(protocol, operation, status string, duration time.Duration) {
	FileSystemOperationsTotal.WithLabelValues(protocol, operation, status).Inc()
	FileSystemOperationDuration.WithLabelValues(protocol, operation).Observe(duration.Seconds())
}

// RecordAuthAttempt records an authentication attempt
func RecordAuthAttempt(method, status string) {
	AuthAttemptsTotal.WithLabelValues(method, status).Inc()
}

// RecordError records an error
func RecordError(component, errorType string) {
	ErrorsTotal.WithLabelValues(component, errorType).Inc()
}

// UpdateMediaByType updates the media count by type
func UpdateMediaByType(mediaType string, count float64) {
	MediaByType.WithLabelValues(mediaType).Set(count)
}

// UpdateStorageRoots updates the storage roots count
func UpdateStorageRoots(protocol, status string, count float64) {
	StorageRootsTotal.WithLabelValues(protocol, status).Set(count)
}

// UpdateActiveSessions updates the active sessions count
func UpdateActiveSessions(count float64) {
	ActiveSessions.Set(count)
}

// UpdateWebSocketConnections updates the active WebSocket connections count
func UpdateWebSocketConnections(count float64) {
	WebSocketConnectionsActive.Set(count)
}

// UpdateDBConnections updates the database connection metrics
func UpdateDBConnections(active, idle int) {
	DBConnectionsActive.Set(float64(active))
	DBConnectionsIdle.Set(float64(idle))
}

// IncrementUptime increments the uptime counter (call this every second)
func IncrementUptime() {
	UptimeSeconds.Inc()
}

// RecordWebSocketMessage records a WebSocket message
func RecordWebSocketMessage(direction string) {
	WebSocketMessagesTotal.WithLabelValues(direction).Inc()
}
