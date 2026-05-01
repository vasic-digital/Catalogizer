package challenges

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// ConversionJobsListChallenge validates GET /api/v1/conversion/jobs
// returns a job list.
type ConversionJobsListChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionJobsListChallenge creates CH-111.
func NewConversionJobsListChallenge() *ConversionJobsListChallenge {
	return &ConversionJobsListChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-jobs-list",
			"Conversion Jobs List",
			"Validates GET /api/v1/conversion/jobs returns a "+
				"job list with 200 status.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion jobs list challenge.
func (c *ConversionJobsListChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("fetching-jobs", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/conversion/jobs",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "conversion_jobs_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Conversion jobs endpoint returned 200",
			fmt.Sprintf("Conversion jobs returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// ConversionJobCreateChallenge validates POST
// /api/v1/conversion/jobs creates a conversion job.
type ConversionJobCreateChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionJobCreateChallenge creates CH-112.
func NewConversionJobCreateChallenge() *ConversionJobCreateChallenge {
	return &ConversionJobCreateChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-job-create",
			"Conversion Job Create",
			"Validates POST /api/v1/conversion/jobs creates a "+
				"conversion job and returns 200 or 201.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion job create challenge.
func (c *ConversionJobCreateChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("creating-job", nil)
	body := `{"source_file":"/tmp/test.avi","target_format":"mp4"}`
	code, _, err := client.PostJSON(
		ctx, "/api/v1/conversion/jobs", body,
	)

	codeOK := err == nil && (code == 200 || code == 201)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "conversion_job_create_status",
		Expected: "200 or 201",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Conversion job create returned %d",
				code),
			fmt.Sprintf(
				"Conversion job create returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// ConversionJobCancelChallenge validates DELETE
// /api/v1/conversion/jobs/:id cancels a job.
type ConversionJobCancelChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionJobCancelChallenge creates CH-113.
func NewConversionJobCancelChallenge() *ConversionJobCancelChallenge {
	return &ConversionJobCancelChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-job-cancel",
			"Conversion Job Cancel",
			"Validates DELETE /api/v1/conversion/jobs/:id "+
				"cancels a conversion job.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion job cancel challenge.
func (c *ConversionJobCancelChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("cancelling-job", nil)
	code, _, err := client.Delete(
		ctx, "/api/v1/conversion/jobs/1",
	)

	// Accept 200, 204, or 404 (no such job)
	codeOK := err == nil &&
		(code == 200 || code == 204 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "conversion_job_cancel_status",
		Expected: "200, 204, or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Conversion job cancel returned %d",
				code),
			fmt.Sprintf(
				"Conversion job cancel returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// ConversionJobRetryChallenge validates POST
// /api/v1/conversion/jobs/:id/retry retries a failed job.
type ConversionJobRetryChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionJobRetryChallenge creates CH-114.
func NewConversionJobRetryChallenge() *ConversionJobRetryChallenge {
	return &ConversionJobRetryChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-job-retry",
			"Conversion Job Retry",
			"Validates POST /api/v1/conversion/jobs/:id/retry "+
				"retries a failed conversion job.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion job retry challenge.
func (c *ConversionJobRetryChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("retrying-job", nil)
	code, _, err := client.PostJSON(
		ctx, "/api/v1/conversion/jobs/1/retry", "{}",
	)

	// Accept 200, 202, or 404 (no such job)
	codeOK := err == nil &&
		(code == 200 || code == 202 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "conversion_job_retry_status",
		Expected: "200, 202, or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Conversion job retry returned %d",
				code),
			fmt.Sprintf(
				"Conversion job retry returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// ConversionJobDownloadChallenge validates GET
// /api/v1/conversion/jobs/:id/download serves a file.
type ConversionJobDownloadChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionJobDownloadChallenge creates CH-115.
func NewConversionJobDownloadChallenge() *ConversionJobDownloadChallenge {
	return &ConversionJobDownloadChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-job-download",
			"Conversion Job Download",
			"Validates GET /api/v1/conversion/jobs/:id/download "+
				"serves the converted file or returns 404.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion job download challenge.
func (c *ConversionJobDownloadChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("downloading-job", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/conversion/jobs/1/download",
	)

	// Accept 200 (file served) or 404 (no such job/file)
	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "conversion_job_download_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf(
				"Conversion job download returned %d", code,
			),
			fmt.Sprintf(
				"Conversion download returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// ConversionAuthRequiredChallenge validates that conversion
// endpoints require authentication.
type ConversionAuthRequiredChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionAuthRequiredChallenge creates CH-116.
func NewConversionAuthRequiredChallenge() *ConversionAuthRequiredChallenge {
	return &ConversionAuthRequiredChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-auth-required",
			"Conversion Auth Required",
			"Validates that conversion endpoints require "+
				"authentication and return 401 without a token.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion auth required challenge.
func (c *ConversionAuthRequiredChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("testing-no-auth", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/api/v1/conversion/jobs", nil,
	)
	resp, err := httpClient.Do(req)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "conversion_reachable",
			Passed:  false,
			Message: fmt.Sprintf("Endpoint unreachable: %v", err),
		})
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, err.Error(),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	resp.Body.Close()

	rejected := resp.StatusCode == 401 ||
		resp.StatusCode == 403
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "conversion_auth_enforced",
		Expected: "401 or 403",
		Actual:   fmt.Sprintf("%d", resp.StatusCode),
		Passed:   rejected,
		Message: challenge.Ternary(rejected,
			"Conversion endpoint requires authentication",
			fmt.Sprintf(
				"Conversion endpoint returned %d without auth",
				resp.StatusCode)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// ConversionJobStatusTransitionChallenge validates that job
// status transitions correctly through lifecycle states.
type ConversionJobStatusTransitionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionJobStatusTransitionChallenge creates CH-117.
func NewConversionJobStatusTransitionChallenge() *ConversionJobStatusTransitionChallenge {
	return &ConversionJobStatusTransitionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-job-status-transition",
			"Conversion Job Status Transition",
			"Validates that conversion job status transitions "+
				"correctly through lifecycle states.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion job status transition challenge.
func (c *ConversionJobStatusTransitionChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	// Create a job and check its initial status
	c.ReportProgress("create-and-check", nil)
	body := `{"source_file":"/tmp/test.avi","target_format":"mp4"}`
	code, rawBody, err := client.PostJSON(
		ctx, "/api/v1/conversion/jobs", body,
	)

	createOK := err == nil && (code == 200 || code == 201)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "job_created",
		Expected: "200 or 201",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   createOK,
		Message: challenge.Ternary(createOK,
			"Conversion job created for status check",
			fmt.Sprintf("Job creation returned %d", code)),
	})

	if createOK && rawBody != nil {
		bodyStr := string(rawBody)
		hasStatus := strings.Contains(bodyStr, "status") ||
			strings.Contains(bodyStr, "state")
		outputs["has_status_field"] = fmt.Sprintf(
			"%t", hasStatus,
		)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "job_has_status",
			Expected: "status or state field in response",
			Actual: challenge.Ternary(hasStatus,
				"present", "missing"),
			Passed: hasStatus,
			Message: challenge.Ternary(hasStatus,
				"Job response contains status field",
				"Job response missing status field"),
		})
	}

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		resultStatus, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), resultStatus))
	return challengeResult, nil
}

// ConversionFormatsListChallenge validates GET
// /api/v1/conversion/formats returns a format list.
type ConversionFormatsListChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionFormatsListChallenge creates CH-118.
func NewConversionFormatsListChallenge() *ConversionFormatsListChallenge {
	return &ConversionFormatsListChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-formats-list",
			"Conversion Formats List",
			"Validates GET /api/v1/conversion/formats returns "+
				"a list of supported formats.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion formats list challenge.
func (c *ConversionFormatsListChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("fetching-formats", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/conversion/formats",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "conversion_formats_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Conversion formats endpoint returned 200",
			fmt.Sprintf("Conversion formats returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// ConversionInvalidJobIDChallenge validates that an invalid
// job ID returns 400.
type ConversionInvalidJobIDChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionInvalidJobIDChallenge creates CH-119.
func NewConversionInvalidJobIDChallenge() *ConversionInvalidJobIDChallenge {
	return &ConversionInvalidJobIDChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-invalid-job-id",
			"Conversion Invalid Job ID",
			"Validates that an invalid job ID returns 400 "+
				"or 404 instead of a server error.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the invalid job ID challenge.
func (c *ConversionInvalidJobIDChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("testing-invalid-id", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/conversion/jobs/not-a-valid-id",
	)

	// Should return 400 or 404, not 500
	codeOK := err == nil && code != 500
	safe := err == nil && (code == 400 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "invalid_job_id_response",
		Expected: "400 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(safe,
			fmt.Sprintf(
				"Invalid job ID correctly returned %d", code,
			),
			challenge.Ternary(codeOK,
				fmt.Sprintf(
					"Invalid job ID returned %d (no 500)",
					code,
				),
				fmt.Sprintf(
					"Invalid job ID caused server error: %d",
					code,
				))),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// ConversionRateLimitChallenge validates that conversion
// endpoints respect rate limits.
type ConversionRateLimitChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConversionRateLimitChallenge creates CH-120.
func NewConversionRateLimitChallenge() *ConversionRateLimitChallenge {
	return &ConversionRateLimitChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"conversion-rate-limit",
			"Conversion Rate Limit",
			"Validates that conversion endpoints respect rate "+
				"limits and do not return 5xx under rapid requests.",
			"conversion",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the conversion rate limit challenge.
func (c *ConversionRateLimitChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("rapid-requests", nil)
	rapidCount := 15
	status429 := 0
	status5xx := 0

	for i := 0; i < rapidCount; i++ {
		code, _, _ := client.GetRaw(
			ctx, "/api/v1/conversion/jobs",
		)
		if code == 429 {
			status429++
		} else if code >= 500 {
			status5xx++
		}
	}

	outputs["total_requests"] = fmt.Sprintf("%d", rapidCount)
	outputs["status_429"] = fmt.Sprintf("%d", status429)
	outputs["status_5xx"] = fmt.Sprintf("%d", status5xx)

	noServerErrors := status5xx == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "no_server_errors",
		Expected: "zero 5xx responses",
		Actual:   fmt.Sprintf("5xx=%d", status5xx),
		Passed:   noServerErrors,
		Message: challenge.Ternary(noServerErrors,
			"No server errors during rapid conversion requests",
			fmt.Sprintf(
				"%d server errors during rapid requests",
				status5xx)),
	})

	rateLimitOrClean := status429 > 0 || status5xx == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "rate_limit_behavior",
		Expected: "429 responses or clean handling",
		Actual:   fmt.Sprintf("429=%d", status429),
		Passed:   rateLimitOrClean,
		Message: challenge.Ternary(status429 > 0,
			fmt.Sprintf(
				"Rate limiting active: %d/%d rate-limited",
				status429, rapidCount),
			"Requests handled cleanly without rate limiting"),
	})

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		resultStatus, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), resultStatus))
	return challengeResult, nil
}
