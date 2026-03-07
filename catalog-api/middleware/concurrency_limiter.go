package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/semaphore"
)

// ConcurrencyLimiter limits the number of concurrent requests being processed.
func ConcurrencyLimiter(maxConcurrent int64) gin.HandlerFunc {
	sem := semaphore.NewWeighted(maxConcurrent)

	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := sem.Acquire(ctx, 1); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "server too busy, try again later",
			})
			return
		}
		defer sem.Release(1)

		c.Next()
	}
}
