package handlers

import (
	"net/http"

	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

// SimpleRecommendationHandler for testing
type SimpleRecommendationHandler struct {
}

// NewSimpleRecommendationHandler creates a new simple recommendation handler
func NewSimpleRecommendationHandler() *SimpleRecommendationHandler {
	return &SimpleRecommendationHandler{}
}

// GetSimpleRecommendation returns a simple recommendation
func (h *SimpleRecommendationHandler) GetSimpleRecommendation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Simple recommendation works!",
	})
}

// GetTest returns a test endpoint
func (h *SimpleRecommendationHandler) GetTest(c *gin.Context) {
	utils.SendErrorResponse(c, http.StatusInternalServerError, "Test error", nil)
}