package handlers

import (
	"net/http"

	"catalogizer/services"
	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

// ChallengeHandler handles challenge API endpoints.
type ChallengeHandler struct {
	service *services.ChallengeService
}

// NewChallengeHandler creates a new challenge handler.
func NewChallengeHandler(
	service *services.ChallengeService,
) *ChallengeHandler {
	return &ChallengeHandler{service: service}
}

// ListChallenges returns all registered challenges.
func (h *ChallengeHandler) ListChallenges(c *gin.Context) {
	challenges := h.service.ListChallenges()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    challenges,
		"count":   len(challenges),
	})
}

// GetChallenge returns details of a specific challenge.
func (h *ChallengeHandler) GetChallenge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendErrorResponse(
			c, http.StatusBadRequest,
			"Challenge ID is required", nil,
		)
		return
	}

	challenges := h.service.ListChallenges()
	for _, ch := range challenges {
		if ch.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    ch,
			})
			return
		}
	}

	utils.SendErrorResponse(
		c, http.StatusNotFound,
		"Challenge not found", nil,
	)
}

// RunChallenge executes a single challenge by ID.
func (h *ChallengeHandler) RunChallenge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendErrorResponse(
			c, http.StatusBadRequest,
			"Challenge ID is required", nil,
		)
		return
	}

	result, err := h.service.RunChallenge(c.Request.Context(), id)
	if err != nil {
		utils.SendErrorResponse(
			c, http.StatusInternalServerError,
			"Failed to run challenge", err,
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// RunAll executes all registered challenges in dependency order.
func (h *ChallengeHandler) RunAll(c *gin.Context) {
	results, err := h.service.RunAll(c.Request.Context())
	if err != nil {
		utils.SendErrorResponse(
			c, http.StatusInternalServerError,
			"Failed to run challenges", err,
		)
		return
	}

	passed := 0
	failed := 0
	for _, r := range results {
		if r.Status == "passed" {
			passed++
		} else {
			failed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"summary": gin.H{
			"total":  len(results),
			"passed": passed,
			"failed": failed,
		},
	})
}

// RunByCategory executes all challenges in a category.
func (h *ChallengeHandler) RunByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		utils.SendErrorResponse(
			c, http.StatusBadRequest,
			"Category is required", nil,
		)
		return
	}

	results, err := h.service.RunByCategory(
		c.Request.Context(), category,
	)
	if err != nil {
		utils.SendErrorResponse(
			c, http.StatusInternalServerError,
			"Failed to run challenges", err,
		)
		return
	}

	passed := 0
	failed := 0
	for _, r := range results {
		if r.Status == "passed" {
			passed++
		} else {
			failed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"summary": gin.H{
			"total":  len(results),
			"passed": passed,
			"failed": failed,
		},
	})
}

// GetResults returns all stored challenge execution results.
func (h *ChallengeHandler) GetResults(c *gin.Context) {
	results := h.service.GetResults()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}
