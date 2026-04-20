package handlers

import (
	"context"
	"net/http"
	"strconv"

	"catalogizer/services"
	"catalogizer/utils"
	"digital.vasic.challenges/pkg/challenge"

	"github.com/gin-gonic/gin"
)

// challengeService defines the methods needed by ChallengeHandler.
type challengeService interface {
	ListChallenges() []services.ChallengeSummary
	RunChallenge(ctx context.Context, id string) (*challenge.Result, error)
	RunAll(ctx context.Context) ([]*challenge.Result, error)
	RunByCategory(ctx context.Context, category string) ([]*challenge.Result, error)
	GetResults() []*challenge.Result
}

// ChallengeHandler handles challenge API endpoints.
type ChallengeHandler struct {
	service challengeService
}

// NewChallengeHandler creates a new challenge handler.
func NewChallengeHandler(
	service challengeService,
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

	result, err := h.service.RunChallenge(context.Background(), id)
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
// Uses context.Background() because RunAll is long-running and
// must not be cancelled by HTTP write timeouts or client
// disconnection. The HTTP response is written after completion.
func (h *ChallengeHandler) RunAll(c *gin.Context) {
	results, err := h.service.RunAll(context.Background())
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
		context.Background(), category,
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

// GetResults returns stored challenge execution results.
//
// FIX-QA-2026-04-21-004 (partial mitigation of DEFER-QA-2026-04-21-001):
// after a 508-challenge RunAll the full result slice carries tens of MB
// of embedded assertion + output data; serialising all of it through
// gin's JSON encoder can sit well past the 60s RequestTimeout and
// present as a "hung" endpoint. Accept an optional `limit` query param
// (default: last 100 results; 0 = unlimited) so diagnostic callers get
// a fast answer without starving the server on huge post-RunAll
// payloads.
func (h *ChallengeHandler) GetResults(c *gin.Context) {
	all := h.service.GetResults()

	limit := 100
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			limit = n
		}
	}

	results := all
	if limit > 0 && len(all) > limit {
		results = all[len(all)-limit:]
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        results,
		"count":       len(results),
		"total_count": len(all),
	})
}
