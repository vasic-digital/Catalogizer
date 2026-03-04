package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// extractBucketBoundaries reads the bucket upper bounds from a histogram metric.
func extractBucketBoundaries(t *testing.T, observer prometheus.Observer) []float64 {
	t.Helper()
	hist, ok := observer.(prometheus.Metric)
	require.True(t, ok, "observer should implement prometheus.Metric")

	m := &dto.Metric{}
	require.NoError(t, hist.Write(m))
	require.NotNil(t, m.GetHistogram(), "metric should be a histogram")

	var boundaries []float64
	for _, bucket := range m.GetHistogram().GetBucket() {
		boundaries = append(boundaries, bucket.GetUpperBound())
	}
	return boundaries
}

// TestHTTPRequestDuration_BucketBoundaries verifies that HTTPRequestDuration
// uses the default Prometheus bucket boundaries. These default buckets cover
// from 5ms to 10s, which is appropriate for HTTP request latencies.
func TestHTTPRequestDuration_BucketBoundaries(t *testing.T) {
	// Observe once to ensure the metric is instantiated with labels.
	HTTPRequestDuration.WithLabelValues("GET", "/bucket-test-http", "200").Observe(0.001)

	boundaries := extractBucketBoundaries(t,
		HTTPRequestDuration.WithLabelValues("GET", "/bucket-test-http", "200"))

	// prometheus.DefBuckets = {.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	expected := prometheus.DefBuckets
	require.Len(t, boundaries, len(expected), "should have %d buckets", len(expected))

	for i, b := range expected {
		assert.InDelta(t, b, boundaries[i], 1e-9,
			"bucket %d should be %v", i, b)
	}
}

// TestDBQueryDuration_BucketBoundaries verifies that DBQueryDuration uses
// custom bucket boundaries tailored for database query latencies (1ms to 5s).
func TestDBQueryDuration_BucketBoundaries(t *testing.T) {
	DBQueryDuration.WithLabelValues("SELECT", "bucket_test_table").Observe(0.001)

	boundaries := extractBucketBoundaries(t,
		DBQueryDuration.WithLabelValues("SELECT", "bucket_test_table"))

	expected := []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}
	require.Len(t, boundaries, len(expected), "should have %d buckets", len(expected))

	for i, b := range expected {
		assert.InDelta(t, b, boundaries[i], 1e-9,
			"bucket %d should be %v", i, b)
	}
}

// TestMediaAnalysisDuration_BucketBoundaries verifies that MediaAnalysisDuration
// uses custom bucket boundaries tailored for media analysis (0.1s to 60s).
func TestMediaAnalysisDuration_BucketBoundaries(t *testing.T) {
	MediaAnalysisDuration.Observe(0.1)

	hist := MediaAnalysisDuration.(prometheus.Metric)
	m := &dto.Metric{}
	require.NoError(t, hist.Write(m))
	require.NotNil(t, m.GetHistogram())

	var boundaries []float64
	for _, bucket := range m.GetHistogram().GetBucket() {
		boundaries = append(boundaries, bucket.GetUpperBound())
	}

	expected := []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60}
	require.Len(t, boundaries, len(expected), "should have %d buckets", len(expected))

	for i, b := range expected {
		assert.InDelta(t, b, boundaries[i], 1e-9,
			"bucket %d should be %v", i, b)
	}
}

// TestExternalAPICallDuration_BucketBoundaries verifies that ExternalAPICallDuration
// uses custom bucket boundaries tailored for external API calls (0.1s to 10s).
func TestExternalAPICallDuration_BucketBoundaries(t *testing.T) {
	ExternalAPICallDuration.WithLabelValues("bucket_test_provider").Observe(0.1)

	boundaries := extractBucketBoundaries(t,
		ExternalAPICallDuration.WithLabelValues("bucket_test_provider"))

	expected := []float64{0.1, 0.5, 1, 2, 5, 10}
	require.Len(t, boundaries, len(expected), "should have %d buckets", len(expected))

	for i, b := range expected {
		assert.InDelta(t, b, boundaries[i], 1e-9,
			"bucket %d should be %v", i, b)
	}
}

// TestFileSystemOperationDuration_BucketBoundaries verifies that
// FileSystemOperationDuration uses custom bucket boundaries tailored for
// filesystem operations (10ms to 10s).
func TestFileSystemOperationDuration_BucketBoundaries(t *testing.T) {
	FileSystemOperationDuration.WithLabelValues("smb_bucket_test", "read").Observe(0.01)

	boundaries := extractBucketBoundaries(t,
		FileSystemOperationDuration.WithLabelValues("smb_bucket_test", "read"))

	expected := []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10}
	require.Len(t, boundaries, len(expected), "should have %d buckets", len(expected))

	for i, b := range expected {
		assert.InDelta(t, b, boundaries[i], 1e-9,
			"bucket %d should be %v", i, b)
	}
}

// TestDBQueryDuration_SubMillisecondBucket verifies that the database query
// histogram has a 1ms bucket, which is critical for detecting fast queries
// that might indicate N+1 patterns.
func TestDBQueryDuration_SubMillisecondBucket(t *testing.T) {
	DBQueryDuration.WithLabelValues("SELECT", "sub_ms_test").Observe(0.001)

	boundaries := extractBucketBoundaries(t,
		DBQueryDuration.WithLabelValues("SELECT", "sub_ms_test"))

	require.True(t, len(boundaries) > 0, "should have at least one bucket")
	assert.InDelta(t, 0.001, boundaries[0], 1e-9,
		"first bucket should be 1ms for sub-millisecond query detection")
}

// TestHTTPRequestDuration_CoversTypicalLatencyRange verifies that the HTTP
// request duration histogram covers the range typical for web API requests.
func TestHTTPRequestDuration_CoversTypicalLatencyRange(t *testing.T) {
	HTTPRequestDuration.WithLabelValues("GET", "/range-test", "200").Observe(0.001)

	boundaries := extractBucketBoundaries(t,
		HTTPRequestDuration.WithLabelValues("GET", "/range-test", "200"))

	require.True(t, len(boundaries) > 0, "should have at least one bucket")

	// The smallest bucket should be <= 10ms (fast requests).
	assert.LessOrEqual(t, boundaries[0], 0.01,
		"smallest bucket should cover fast requests (<= 10ms)")

	// The largest bucket should be >= 5s (slow requests).
	assert.GreaterOrEqual(t, boundaries[len(boundaries)-1], 5.0,
		"largest bucket should cover slow requests (>= 5s)")
}

// TestAllHistograms_BucketsAreMonotonicallyIncreasing verifies that bucket
// boundaries for all histogram metrics are strictly monotonically increasing.
func TestAllHistograms_BucketsAreMonotonicallyIncreasing(t *testing.T) {
	tests := []struct {
		name     string
		observer prometheus.Observer
	}{
		{
			"HTTPRequestDuration",
			HTTPRequestDuration.WithLabelValues("GET", "/monotonic-http", "200"),
		},
		{
			"DBQueryDuration",
			DBQueryDuration.WithLabelValues("SELECT", "monotonic_db"),
		},
		{
			"ExternalAPICallDuration",
			ExternalAPICallDuration.WithLabelValues("monotonic_provider"),
		},
		{
			"FileSystemOperationDuration",
			FileSystemOperationDuration.WithLabelValues("smb_monotonic", "read"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Observe once to instantiate.
			tt.observer.Observe(0.001)

			boundaries := extractBucketBoundaries(t, tt.observer)
			require.True(t, len(boundaries) >= 2,
				"histogram should have at least 2 buckets")

			for i := 1; i < len(boundaries); i++ {
				assert.Greater(t, boundaries[i], boundaries[i-1],
					"bucket %d (%v) should be greater than bucket %d (%v)",
					i, boundaries[i], i-1, boundaries[i-1])
			}
		})
	}
}
