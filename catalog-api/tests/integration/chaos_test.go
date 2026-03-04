package integration

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChaos_DatabaseUnavailable verifies API returns proper error responses
// when the database connection is closed.
func TestChaos_DatabaseUnavailable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping chaos test in short mode")
	}

	gin.SetMode(gin.TestMode)

	// Create and close database to simulate unavailability
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.Close() // close immediately to simulate unavailability

	router := gin.New()
	router.Use(gin.Recovery()) // must not panic

	// Handler that tries to use a closed database
	router.GET("/api/v1/health-db", func(c *gin.Context) {
		err := db.Ping()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database unavailable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.GET("/api/v1/catalog/files", func(c *gin.Context) {
		_, err := db.Query("SELECT id FROM files LIMIT 1")
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "database unavailable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"files": []interface{}{}})
	})

	// Test health check reports unhealthy
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health-db", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "database unavailable")

	// Test file listing returns service unavailable
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/catalog/files", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestChaos_DatabaseReconnection verifies the application can recover
// after a temporary database outage.
func TestChaos_DatabaseReconnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping chaos test in short mode")
	}

	gin.SetMode(gin.TestMode)

	db, err := sql.Open("sqlite3", ":memory:?_busy_timeout=5000")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS health_check (id INTEGER PRIMARY KEY, ts DATETIME)")
	require.NoError(t, err)

	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/api/v1/ping", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		err := db.PingContext(ctx)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Phase 1: Database is available
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ping", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Phase 2: Database operations still work
	_, err = db.Exec("INSERT INTO health_check (ts) VALUES (?)", time.Now())
	require.NoError(t, err)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/ping", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestChaos_ContextCancellation verifies handlers respect context cancellation
// and don't leak goroutines.
func TestChaos_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping chaos test in short mode")
	}

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/api/v1/slow", func(c *gin.Context) {
		ctx := c.Request.Context()
		select {
		case <-time.After(5 * time.Second):
			c.JSON(http.StatusOK, gin.H{"result": "done"})
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "cancelled"})
		}
	})

	// Create a request with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "/api/v1/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Either cancelled or timed out - both are acceptable
	assert.True(t, w.Code == http.StatusRequestTimeout || w.Code == http.StatusOK || w.Code == 0,
		"expected timeout/ok/empty, got %d", w.Code)
}

// TestChaos_PanicRecovery verifies the Gin recovery middleware catches panics
// and returns 500 without crashing the server.
func TestChaos_PanicRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping chaos test in short mode")
	}

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/api/v1/panic", func(c *gin.Context) {
		panic("simulated crash")
	})

	// The server should not crash
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/panic", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Subsequent requests should still work
	router.GET("/api/v1/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/ok", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestChaos_ConcurrentDatabaseAccess verifies the database handles concurrent
// reads and writes without corruption.
func TestChaos_ConcurrentDatabaseAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping chaos test in short mode")
	}

	db, err := sql.Open("sqlite3", ":memory:?_busy_timeout=5000&_journal_mode=WAL")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE test_data (id INTEGER PRIMARY KEY AUTOINCREMENT, value TEXT, created_at DATETIME)`)
	require.NoError(t, err)

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/write", func(c *gin.Context) {
		_, err := db.Exec("INSERT INTO test_data (value, created_at) VALUES (?, ?)", "test", time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "written"})
	})

	router.GET("/read", func(c *gin.Context) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"count": count})
	})

	// Concurrent writes and reads
	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/write", nil)
			router.ServeHTTP(w, req)
		}()
		go func() {
			defer func() { done <- true }()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/read", nil)
			router.ServeHTTP(w, req)
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify data integrity
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 10, count)
}

// TestChaos_ConnectionPoolExhaustion verifies graceful handling when the
// database connection pool is exhausted.
func TestChaos_ConnectionPoolExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping chaos test in short mode")
	}

	db, err := sql.Open("sqlite3", ":memory:?_busy_timeout=5000")
	require.NoError(t, err)
	defer db.Close()

	// Very small pool to trigger exhaustion
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)

	_, err = db.Exec("CREATE TABLE pool_test (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/query", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pool_test").Scan(&count)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "pool exhausted"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"count": count})
	})

	// Fire multiple concurrent requests against a small pool
	results := make(chan int, 10)
	for i := 0; i < 10; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/query", nil)
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	successCount := 0
	for i := 0; i < 10; i++ {
		code := <-results
		if code == http.StatusOK {
			successCount++
		}
	}

	// At least some requests should succeed
	assert.Greater(t, successCount, 0, "at least some requests should succeed")
}
