package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestTimeout_SetsContextDeadline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestTimeout(5 * time.Second))

	var hasDeadline bool
	r.GET("/test", func(c *gin.Context) {
		_, hasDeadline = c.Request.Context().Deadline()
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.True(t, hasDeadline, "context should have a deadline")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestTimeout_DeadlineMatchesDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	timeout := 3 * time.Second
	r := gin.New()
	r.Use(RequestTimeout(timeout))

	var deadline time.Time
	r.GET("/test", func(c *gin.Context) {
		deadline, _ = c.Request.Context().Deadline()
		c.String(http.StatusOK, "ok")
	})

	before := time.Now()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// Deadline should be roughly now + timeout
	expected := before.Add(timeout)
	assert.WithinDuration(t, expected, deadline, 500*time.Millisecond)
}

func TestRequestTimeout_ContextCancelledAfterExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestTimeout(50 * time.Millisecond))

	var ctxErr error
	r.GET("/test", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		ctxErr = c.Request.Context().Err()
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Error(t, ctxErr, "context should be cancelled after timeout")
}

func TestRequestTimeout_NormalRequestCompletes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestTimeout(5 * time.Second))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}
