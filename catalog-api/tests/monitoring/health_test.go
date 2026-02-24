package monitoring_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/internal/metrics"

	"github.com/gin-gonic/gin"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// openTestDB opens an in-memory SQLite database suitable for health check tests.
// MaxOpenConns is set to 10 so the connection pool check does not incorrectly
// report degraded status.
func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(10)
	return database.WrapDB(sqlDB, database.DialectSQLite)
}

// setupHealthRouter creates a Gin router that exposes /health (simple JSON),
// /health/detailed (HealthChecker-based), /health/live, /health/ready,
// /health/startup, and /metrics endpoints.
func setupHealthRouter(hc *metrics.HealthChecker) *gin.Engine {
	router := gin.New()
	router.Use(metrics.GinMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Simple health endpoint (mirrors main.go)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"time":         time.Now().UTC(),
			"version":      "test",
			"build_number": "0",
			"build_date":   "unknown",
		})
	})

	// Detailed health endpoint using HealthChecker
	router.GET("/health/detailed", func(c *gin.Context) {
		resp := hc.Check(c.Request.Context())
		status := http.StatusOK
		if resp.Status == metrics.HealthStatusUnhealthy {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, resp)
	})

	// Kubernetes-style probes
	router.GET("/health/live", func(c *gin.Context) {
		c.Status(hc.LivenessProbe())
	})

	router.GET("/health/ready", func(c *gin.Context) {
		c.Status(hc.ReadinessProbe(c.Request.Context()))
	})

	router.GET("/health/startup", func(c *gin.Context) {
		c.Status(hc.StartupProbe(c.Request.Context()))
	})

	return router
}

// --- Simple /health endpoint tests ---

// TestHealthEndpoint_ReturnsOK verifies that the /health endpoint returns
// HTTP 200 with a JSON body containing the expected fields.
func TestHealthEndpoint_ReturnsOK(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "test-version")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestHealthEndpoint_ReturnsExpectedFields verifies the simple /health
// endpoint returns all expected JSON fields.
func TestHealthEndpoint_ReturnsExpectedFields(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "test-version")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)

	assert.Equal(t, "healthy", body["status"])
	assert.Equal(t, "test", body["version"])
	assert.Equal(t, "0", body["build_number"])
	assert.Equal(t, "unknown", body["build_date"])
	assert.NotNil(t, body["time"], "response should include a timestamp")
}

// --- Detailed health check endpoint tests ---

// TestDetailedHealth_AllHealthyReturns200 verifies that when all components
// are healthy, the detailed health endpoint returns HTTP 200 with overall
// status "healthy".
func TestDetailedHealth_AllHealthyReturns200(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp metrics.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, metrics.HealthStatusHealthy, resp.Status)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.NotZero(t, resp.Timestamp)
	assert.NotEmpty(t, resp.Uptime)
	assert.Contains(t, resp.Components, "database")
	assert.Equal(t, metrics.HealthStatusHealthy, resp.Components["database"].Status)
}

// TestDetailedHealth_ContainsDatabaseComponent verifies the detailed health
// response always includes the database component.
func TestDetailedHealth_ContainsDatabaseComponent(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	var resp metrics.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	dbHealth, exists := resp.Components["database"]
	assert.True(t, exists, "database component should be present")
	assert.Equal(t, metrics.HealthStatusHealthy, dbHealth.Status)
	assert.NotEmpty(t, dbHealth.Latency, "database health should report latency")
}

// TestDetailedHealth_NilDBReturns503 verifies that when the database is nil,
// the detailed health endpoint returns HTTP 503 with overall status "unhealthy".
func TestDetailedHealth_NilDBReturns503(t *testing.T) {
	hc := metrics.NewHealthChecker(nil, "1.0.0")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp metrics.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, metrics.HealthStatusUnhealthy, resp.Status)
	assert.Equal(t, metrics.HealthStatusUnhealthy, resp.Components["database"].Status)
	assert.Equal(t, "Database not configured", resp.Components["database"].Message)
}

// TestDetailedHealth_ClosedDBReturns503 verifies that when the database
// connection has been closed, the detailed health endpoint returns HTTP 503.
func TestDetailedHealth_ClosedDBReturns503(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	sqlDB.Close() // close immediately

	hc := metrics.NewHealthChecker(database.WrapDB(sqlDB, database.DialectSQLite), "1.0.0")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp metrics.HealthCheckResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, metrics.HealthStatusUnhealthy, resp.Status)
}

// TestDetailedHealth_CustomComponentIncluded verifies that custom health checks
// registered on the HealthChecker are included in the detailed response.
func TestDetailedHealth_CustomComponentIncluded(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	hc.RegisterCheck("redis", func(ctx context.Context) metrics.ComponentHealth {
		return metrics.ComponentHealth{
			Status:  metrics.HealthStatusHealthy,
			Message: "connected",
		}
	})

	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	var resp metrics.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, metrics.HealthStatusHealthy, resp.Status)
	assert.Contains(t, resp.Components, "redis")
	assert.Equal(t, metrics.HealthStatusHealthy, resp.Components["redis"].Status)
	assert.Equal(t, "connected", resp.Components["redis"].Message)
}

// TestDetailedHealth_DegradedComponentMakesOverallDegraded verifies that a
// degraded custom component causes the overall status to become degraded.
func TestDetailedHealth_DegradedComponentMakesOverallDegraded(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	hc.RegisterCheck("cache", func(ctx context.Context) metrics.ComponentHealth {
		return metrics.ComponentHealth{
			Status:  metrics.HealthStatusDegraded,
			Message: "high latency",
		}
	})

	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "degraded should still return 200")

	var resp metrics.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, metrics.HealthStatusDegraded, resp.Status)
}

// TestDetailedHealth_UnhealthyOverridesDegraded verifies that if one component
// is degraded and another is unhealthy, the overall status is unhealthy.
func TestDetailedHealth_UnhealthyOverridesDegraded(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	hc.RegisterCheck("cache", func(ctx context.Context) metrics.ComponentHealth {
		return metrics.ComponentHealth{Status: metrics.HealthStatusDegraded}
	})
	hc.RegisterCheck("external_api", func(ctx context.Context) metrics.ComponentHealth {
		return metrics.ComponentHealth{Status: metrics.HealthStatusUnhealthy, Message: "down"}
	})

	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp metrics.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, metrics.HealthStatusUnhealthy, resp.Status)
}

// TestDetailedHealth_VersionReflectsConfig verifies that the version field
// in the health response matches what was configured.
func TestDetailedHealth_VersionReflectsConfig(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	versions := []string{"1.0.0", "2.5.3-beta", "dev", ""}
	for _, v := range versions {
		t.Run("version="+v, func(t *testing.T) {
			hc := metrics.NewHealthChecker(db, v)
			router := setupHealthRouter(hc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health/detailed", nil)
			router.ServeHTTP(w, req)

			var resp metrics.HealthCheckResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, v, resp.Version)
		})
	}
}

// TestDetailedHealth_UptimeIncreases verifies that the uptime value reported
// by the health endpoint is positive and reflects elapsed time.
func TestDetailedHealth_UptimeIncreases(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	router := setupHealthRouter(hc)

	// Small sleep to ensure non-zero uptime.
	time.Sleep(5 * time.Millisecond)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	var resp metrics.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Uptime, "uptime should not be empty")
	// The uptime string should be parseable as a Go duration.
	uptime, err := time.ParseDuration(resp.Uptime)
	require.NoError(t, err, "uptime should be a valid Go duration")
	assert.Greater(t, uptime, time.Duration(0), "uptime should be positive")
}

// --- Kubernetes probe tests (via HTTP) ---

// TestLivenessProbe_AlwaysReturns200 verifies the liveness probe always
// returns HTTP 200 regardless of component health.
func TestLivenessProbe_AlwaysReturns200(t *testing.T) {
	tests := []struct {
		name string
		db   *database.DB
	}{
		{"with healthy db", nil},
		{"with nil db", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var db *database.DB
			if tt.name == "with healthy db" {
				db = openTestDB(t)
				defer db.Close()
			}
			hc := metrics.NewHealthChecker(db, "1.0.0")
			router := setupHealthRouter(hc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health/live", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code,
				"liveness probe should always return 200")
		})
	}
}

// TestReadinessProbe_HealthyReturns200 verifies the readiness probe returns
// HTTP 200 when all components are healthy.
func TestReadinessProbe_HealthyReturns200(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestReadinessProbe_DegradedReturns200 verifies the readiness probe returns
// HTTP 200 when a component is degraded (service can still handle traffic).
func TestReadinessProbe_DegradedReturns200(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	hc.RegisterCheck("slow_service", func(ctx context.Context) metrics.ComponentHealth {
		return metrics.ComponentHealth{Status: metrics.HealthStatusDegraded}
	})
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code,
		"readiness probe should return 200 when degraded")
}

// TestReadinessProbe_UnhealthyReturns503 verifies the readiness probe returns
// HTTP 503 when a component is unhealthy.
func TestReadinessProbe_UnhealthyReturns503(t *testing.T) {
	hc := metrics.NewHealthChecker(nil, "1.0.0")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code,
		"readiness probe should return 503 when unhealthy")
}

// TestStartupProbe_HealthyReturns200 verifies the startup probe returns
// HTTP 200 when all components are healthy.
func TestStartupProbe_HealthyReturns200(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/startup", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestStartupProbe_UnhealthyReturns503 verifies the startup probe returns
// HTTP 503 when any component is unhealthy (service not ready to start).
func TestStartupProbe_UnhealthyReturns503(t *testing.T) {
	hc := metrics.NewHealthChecker(nil, "1.0.0")
	hc.RegisterCheck("init_service", func(ctx context.Context) metrics.ComponentHealth {
		return metrics.ComponentHealth{
			Status:  metrics.HealthStatusUnhealthy,
			Message: "not initialized",
		}
	})
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/startup", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code,
		"startup probe should return 503 when unhealthy")
}

// --- Health endpoint metrics integration tests ---

// TestHealthEndpoint_TrackedInPrometheusMetrics verifies that health check
// requests generate entries in the Prometheus metrics output.
func TestHealthEndpoint_TrackedInPrometheusMetrics(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	router := setupHealthRouter(hc)

	// Make several health check requests.
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Fetch /metrics to verify health requests appear.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, `path="/health"`,
		"Prometheus output should contain health endpoint path")
	assert.Contains(t, body, `method="GET"`,
		"Prometheus output should contain GET method label")
}

// TestDetailedHealth_TrackedInPrometheusMetrics verifies that the detailed
// health endpoint requests are also tracked in Prometheus metrics.
func TestDetailedHealth_TrackedInPrometheusMetrics(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Fetch /metrics to verify.
	mw := httptest.NewRecorder()
	mreq, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(mw, mreq)

	body := mw.Body.String()
	assert.Contains(t, body, `path="/health/detailed"`,
		"Prometheus output should contain detailed health endpoint path")
}

// TestHealthProbes_TrackedInPrometheusMetrics verifies that Kubernetes-style
// probe endpoints are tracked in Prometheus metrics.
func TestHealthProbes_TrackedInPrometheusMetrics(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	router := setupHealthRouter(hc)

	probeEndpoints := []string{"/health/live", "/health/ready", "/health/startup"}
	for _, endpoint := range probeEndpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", endpoint, nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Fetch /metrics to verify.
	mw := httptest.NewRecorder()
	mreq, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(mw, mreq)

	body := mw.Body.String()
	for _, endpoint := range probeEndpoints {
		assert.Contains(t, body, `path="`+endpoint+`"`,
			"Prometheus output should contain %s endpoint path", endpoint)
	}
}

// --- Multiple components health status combination tests ---

// TestDetailedHealth_MultipleHealthyComponents verifies correct behavior with
// multiple custom healthy components.
func TestDetailedHealth_MultipleHealthyComponents(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "1.0.0")
	components := []string{"redis", "smb", "scanner", "websocket"}
	for _, name := range components {
		hc.RegisterCheck(name, func(ctx context.Context) metrics.ComponentHealth {
			return metrics.ComponentHealth{Status: metrics.HealthStatusHealthy}
		})
	}

	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp metrics.HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, metrics.HealthStatusHealthy, resp.Status)
	// 4 custom + 1 database = 5 components
	assert.Len(t, resp.Components, 5)
	for _, name := range components {
		assert.Contains(t, resp.Components, name)
	}
}

// TestDetailedHealth_ResponseIsValidJSON verifies that the detailed health
// response is always valid JSON with the correct structure.
func TestDetailedHealth_ResponseIsValidJSON(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	hc := metrics.NewHealthChecker(db, "2.0.0")
	hc.RegisterCheck("test_component", func(ctx context.Context) metrics.ComponentHealth {
		return metrics.ComponentHealth{
			Status:  metrics.HealthStatusHealthy,
			Message: "operational",
			Latency: "1ms",
		}
	})

	router := setupHealthRouter(hc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	// Verify valid JSON.
	var raw json.RawMessage
	err := json.Unmarshal(w.Body.Bytes(), &raw)
	assert.NoError(t, err, "response should be valid JSON")

	// Verify full structure roundtrip.
	var resp metrics.HealthCheckResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "2.0.0", resp.Version)
	assert.NotEmpty(t, resp.Uptime)
	assert.NotZero(t, resp.Timestamp)
	assert.Equal(t, metrics.HealthStatusHealthy, resp.Components["test_component"].Status)
	assert.Equal(t, "operational", resp.Components["test_component"].Message)
	assert.Equal(t, "1ms", resp.Components["test_component"].Latency)
}
