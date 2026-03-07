package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// UserManagementChallenge validates the user management API:
// listing users, viewing profiles, updating settings, and
// role-based access control.
type UserManagementChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewUserManagementChallenge creates CH-054.
func NewUserManagementChallenge() *UserManagementChallenge {
	return &UserManagementChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"user-management",
			"User Management API",
			"Validates the user management endpoints: list users, "+
				"view user profile, update settings, and verify "+
				"role-based access restrictions.",
			"api",
			[]challenge.ID{"auth-required"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the user management challenge.
func (c *UserManagementChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Login as admin
	c.ReportProgress("authenticating", nil)
	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 5)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("admin login failed: %v", err),
		), nil
	}

	// Test 1: Get current user profile
	c.ReportProgress("testing-user-profile", nil)
	status, body, _ := client.Get(ctx, "/users/me")

	profileOK := status == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "user_profile",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", status),
		Passed:   profileOK,
		Message:  challenge.Ternary(profileOK, "User profile endpoint works", "User profile endpoint failed"),
	})

	if body != nil {
		if username, ok := body["username"]; ok {
			outputs["current_user"] = fmt.Sprintf("%v", username)
		}
	}

	// Test 2: List users (admin-only)
	c.ReportProgress("testing-list-users", nil)
	statusList, bodyList, _ := client.Get(ctx, "/users")

	listOK := statusList == 200 && bodyList != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "list_users",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusList),
		Passed:   listOK,
		Message:  challenge.Ternary(listOK, "List users endpoint works", "List users endpoint failed"),
	})

	// Test 3: Get init status
	c.ReportProgress("testing-init-status", nil)
	statusInit, bodyInit, _ := client.Get(ctx, "/auth/init-status")

	initOK := statusInit == 200 && bodyInit != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "init_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusInit),
		Passed:   initOK,
		Message:  challenge.Ternary(initOK, "Init status endpoint works", "Init status endpoint failed"),
	})

	if bodyInit != nil {
		if initialized, ok := bodyInit["initialized"]; ok {
			outputs["initialized"] = fmt.Sprintf("%v", initialized)
		}
	}

	// Test 4: Verify unauthorized access is blocked
	c.ReportProgress("testing-unauthorized", nil)
	unauthClient := httpclient.NewAPIClient(c.config.BaseURL)
	statusUnauth, _, _ := unauthClient.Get(ctx, "/users")

	unauthBlocked := statusUnauth == 401
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "unauthorized_blocked",
		Expected: "401",
		Actual:   fmt.Sprintf("%d", statusUnauth),
		Passed:   unauthBlocked,
		Message:  challenge.Ternary(unauthBlocked, "Unauthorized access properly blocked", "Unauthorized access not blocked"),
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
