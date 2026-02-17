package challenges

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// maxOutputBytes limits captured stdout/stderr from Playwright
// to avoid oversized challenge outputs.
const maxOutputBytes = 4096

// BrowsingWebAppChallenge invokes Playwright browser tests
// against the running web app to validate real browsing.
type BrowsingWebAppChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewBrowsingWebAppChallenge creates CH-010.
func NewBrowsingWebAppChallenge() *BrowsingWebAppChallenge {
	return &BrowsingWebAppChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"browsing-web-app",
			"Web App Browsing",
			"Runs Playwright browser tests against the live web app to validate "+
				"dashboard loading, media browsing, and navigation",
			"e2e",
			[]challenge.ID{"browsing-api-catalog"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the Playwright-based web app challenge.
func (c *BrowsingWebAppChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"web_app_url": c.config.WebAppURL,
		"web_app_dir": c.config.WebAppDir,
	}

	// Step 1: Check npx is available
	npxPath, npxErr := exec.LookPath("npx")
	npxOK := npxErr == nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "npx_available",
		Expected: "npx in PATH",
		Actual:   ternary(npxOK, npxPath, fmt.Sprintf("not found: %v", npxErr)),
		Passed:   npxOK,
		Message:  ternary(npxOK, fmt.Sprintf("npx found at %s", npxPath), fmt.Sprintf("npx not found in PATH: %v", npxErr)),
	})
	if !npxOK {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, "npx not available"), nil
	}

	// Step 2: Pre-flight check - web app URL is reachable
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, prefErr := httpClient.Get(c.config.WebAppURL)
	prefOK := prefErr == nil && resp != nil && resp.StatusCode < 500
	if resp != nil {
		resp.Body.Close()
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "web_app_reachable",
		Expected: "HTTP response from web app",
		Actual:   ternary(prefOK, fmt.Sprintf("HTTP %d", resp.StatusCode), fmt.Sprintf("err=%v", prefErr)),
		Passed:   prefOK,
		Message:  ternary(prefOK, fmt.Sprintf("Web app reachable at %s", c.config.WebAppURL), fmt.Sprintf("Web app not reachable: %v", prefErr)),
	})
	if !prefOK {
		errMsg := "web app not reachable"
		if prefErr != nil {
			errMsg = prefErr.Error()
		}
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, errMsg), nil
	}

	// Step 3: Run Playwright tests
	cmd := exec.CommandContext(ctx, "npx", "playwright", "test",
		"--project=chromium",
		"--reporter=json",
		"e2e/tests/browsing-challenge.spec.ts",
	)
	cmd.Dir = c.config.WebAppDir
	cmd.Env = append(cmd.Environ(),
		"PLAYWRIGHT_BASE_URL="+c.config.WebAppURL,
		"BROWSING_API_URL="+c.config.BaseURL,
		"ADMIN_USERNAME="+c.config.Username,
		"ADMIN_PASSWORD="+c.config.Password,
		"CI=true",
	)

	output, runErr := cmd.CombinedOutput()

	// Truncate output for storage
	outputStr := string(output)
	if len(outputStr) > maxOutputBytes {
		outputStr = outputStr[:maxOutputBytes] + "\n...(truncated)"
	}
	outputs["playwright_output"] = outputStr

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	outputs["exit_code"] = fmt.Sprintf("%d", exitCode)

	playwrightOK := runErr == nil && exitCode == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "playwright_tests",
		Expected: "exit code 0",
		Actual:   fmt.Sprintf("exit code %d", exitCode),
		Passed:   playwrightOK,
		Message:  ternary(playwrightOK, "All Playwright tests passed", fmt.Sprintf("Playwright tests failed with exit code %d", exitCode)),
	})

	metrics := map[string]challenge.MetricValue{
		"playwright_time": {
			Name:  "playwright_time",
			Value: float64(time.Since(start).Milliseconds()),
			Unit:  "ms",
		},
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	errMsg := ""
	if !playwrightOK && runErr != nil {
		errMsg = runErr.Error()
	}

	return c.CreateResult(status, start, assertions, metrics, outputs, errMsg), nil
}
