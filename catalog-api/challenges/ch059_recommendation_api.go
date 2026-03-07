package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// RecommendationAPIChallenge validates the recommendation engine API:
// getting recommendations by media type, similar items, and
// personalized suggestions.
type RecommendationAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewRecommendationAPIChallenge creates CH-059.
func NewRecommendationAPIChallenge() *RecommendationAPIChallenge {
	return &RecommendationAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"recommendation-api",
			"Recommendation Engine API",
			"Validates the recommendation endpoints: get recommendations "+
				"for media type, find similar items, and personalized "+
				"suggestions based on user activity.",
			"api",
			[]challenge.ID{"entity-browsing"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the recommendation API challenge.
func (c *RecommendationAPIChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Login
	c.ReportProgress("authenticating", nil)
	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 5)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", err),
		), nil
	}

	// Test 1: Recommendations list
	c.ReportProgress("testing-recommendations", nil)
	status, _, _ := client.Get(ctx, "/recommendations")

	recsOK := status == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "recommendations_list",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", status),
		Passed:   recsOK,
		Message:  challenge.Ternary(recsOK, "Recommendations endpoint works", "Recommendations endpoint failed"),
	})

	// Test 2: Recommendations by type
	c.ReportProgress("testing-recs-by-type", nil)
	statusType, _, _ := client.Get(ctx, "/recommendations?type=movie&limit=5")

	typeOK := statusType == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "recommendations_by_type",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusType),
		Passed:   typeOK,
		Message:  challenge.Ternary(typeOK, "Recommendations by type works", "Recommendations by type failed"),
	})

	// Test 3: Similar items for non-existent entity
	c.ReportProgress("testing-similar-not-found", nil)
	statusSimilar, _, _ := client.Get(ctx, "/entities/99999999/similar")

	similarOK := statusSimilar == 404 || statusSimilar == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "similar_items",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", statusSimilar),
		Passed:   similarOK,
		Message:  challenge.Ternary(similarOK, "Similar items endpoint responds correctly", "Similar items endpoint error"),
	})

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(resultStatus, start, assertions, nil, outputs, ""), nil
}
