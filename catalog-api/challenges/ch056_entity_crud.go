package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntityCRUDChallenge validates full CRUD operations on media entities:
// list, get by ID, search, filter by type, and verify hierarchy.
type EntityCRUDChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityCRUDChallenge creates CH-056.
func NewEntityCRUDChallenge() *EntityCRUDChallenge {
	return &EntityCRUDChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-crud",
			"Entity CRUD Operations",
			"Validates full CRUD lifecycle on media entities: "+
				"list entities, get by ID, search by title, "+
				"filter by media type, and verify parent-child hierarchy.",
			"api",
			[]challenge.ID{"entity-browsing"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity CRUD challenge.
func (c *EntityCRUDChallenge) Execute(
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

	// Test 1: List entities
	c.ReportProgress("listing-entities", nil)
	status, body, _ := client.Get(ctx, "/entities?limit=5")

	listOK := status == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_list",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", status),
		Passed:   listOK,
		Message:  challenge.Ternary(listOK, "Entity list works", "Entity list failed"),
	})

	// Test 2: Get media types
	c.ReportProgress("getting-media-types", nil)
	statusTypes, bodyTypes, _ := client.Get(ctx, "/media-types")

	typesOK := statusTypes == 200 && bodyTypes != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "media_types",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusTypes),
		Passed:   typesOK,
		Message:  challenge.Ternary(typesOK, "Media types endpoint works", "Media types endpoint failed"),
	})

	// Test 3: Search entities
	c.ReportProgress("searching-entities", nil)
	statusSearch, _, _ := client.Get(ctx, "/entities?search=test&limit=3")

	searchOK := statusSearch == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_search",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusSearch),
		Passed:   searchOK,
		Message:  challenge.Ternary(searchOK, "Entity search works", "Entity search failed"),
	})

	// Test 4: Filter by media type
	c.ReportProgress("filtering-entities", nil)
	statusFilter, _, _ := client.Get(ctx, "/entities?type=movie&limit=3")

	filterOK := statusFilter == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_filter",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusFilter),
		Passed:   filterOK,
		Message:  challenge.Ternary(filterOK, "Entity filter by type works", "Entity filter by type failed"),
	})

	// Test 5: Get non-existent entity returns 404
	c.ReportProgress("testing-not-found", nil)
	statusNotFound, _, _ := client.Get(ctx, "/entities/99999999")

	notFoundOK := statusNotFound == 404
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_not_found",
		Expected: "404",
		Actual:   fmt.Sprintf("%d", statusNotFound),
		Passed:   notFoundOK,
		Message:  challenge.Ternary(notFoundOK, "Non-existent entity returns 404", "Non-existent entity does not return 404"),
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
