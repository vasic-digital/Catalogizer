package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGinMiddleware_RecordsMetrics(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	beforeCount := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/test", "200"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	afterCount := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/test", "200"))
	assert.Equal(t, beforeCount+1, afterCount)
}

func TestGinMiddleware_TracksActiveConnections(t *testing.T) {
	connDuringRequest := float64(-1)

	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/api/v1/active-test", func(c *gin.Context) {
		connDuringRequest = getGaugeValue(HTTPActiveConnections)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	beforeConn := getGaugeValue(HTTPActiveConnections)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/active-test", nil)
	router.ServeHTTP(w, req)

	afterConn := getGaugeValue(HTTPActiveConnections)

	// During the request, active connections should have been incremented.
	assert.Equal(t, beforeConn+1, connDuringRequest)
	// After the request, active connections should return to the previous value.
	assert.Equal(t, beforeConn, afterConn)
}

func TestGinMiddleware_SkipsMetricsEndpoint(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"metrics": true})
	})

	beforeCount := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/metrics", "200"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	afterCount := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/metrics", "200"))
	// Should not have incremented since /metrics is skipped.
	assert.Equal(t, beforeCount, afterCount)
}

func TestGinMiddleware_Records404(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware())

	beforeCount := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/nonexistent", "404"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	afterCount := getCounterValue(HTTPRequestsTotal.WithLabelValues("GET", "/nonexistent", "404"))
	assert.Equal(t, beforeCount+1, afterCount)
}

func TestNormalizePath_UsesFullPath(t *testing.T) {
	router := gin.New()
	var capturedPath string

	router.GET("/api/v1/items/:id", func(c *gin.Context) {
		capturedPath = normalizePath(c)
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/items/123", nil)
	router.ServeHTTP(w, req)

	// Should use the route pattern, not the actual path with the parameter value.
	assert.Equal(t, "/api/v1/items/:id", capturedPath)
}
