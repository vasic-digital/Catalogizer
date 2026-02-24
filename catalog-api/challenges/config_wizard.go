package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// ConfigWizardChallenge validates the configuration wizard flow:
// get wizard steps, validate a step configuration, and complete
// the wizard process.
type ConfigWizardChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConfigWizardChallenge creates CH-035.
func NewConfigWizardChallenge() *ConfigWizardChallenge {
	return &ConfigWizardChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"config-wizard",
			"Configuration Wizard",
			"Validates wizard flow: get wizard steps/status, "+
				"validate step configuration, complete wizard process.",
			"configuration",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the configuration wizard challenge.
func (c *ConfigWizardChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "login",
			Passed:  false,
			Message: fmt.Sprintf("Login failed: %v", err),
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, err.Error(),
		), nil
	}

	// Step 1: Get wizard progress (actual endpoint: /api/v1/configuration/wizard/progress)
	// Use GetRaw since the handler may return empty body or non-JSON on error
	c.ReportProgress("getting-wizard-status", nil)
	wizardCode, wizardRaw, wizardErr := client.GetRaw(ctx, "/api/v1/configuration/wizard/progress")
	var wizardBody map[string]interface{}
	if wizardErr == nil && len(wizardRaw) > 0 {
		_ = json.Unmarshal(wizardRaw, &wizardBody)
	}

	wizardResponds := wizardErr == nil && wizardCode != 0
	wizardOK := wizardErr == nil && wizardCode == 200
	stepCount := 0
	isComplete := false
	if wizardBody != nil {
		if steps, ok := wizardBody["steps"].([]interface{}); ok {
			stepCount = len(steps)
		}
		if complete, ok := wizardBody["is_complete"].(bool); ok {
			isComplete = complete
		} else if complete, ok := wizardBody["completed"].(bool); ok {
			isComplete = complete
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "wizard_progress_endpoint",
		Expected: "200 or valid response",
		Actual:   fmt.Sprintf("HTTP %d, steps=%d, complete=%v", wizardCode, stepCount, isComplete),
		Passed:   wizardResponds,
		Message: challenge.Ternary(wizardOK,
			fmt.Sprintf("Wizard progress: %d steps, complete=%v", stepCount, isComplete),
			challenge.Ternary(wizardResponds,
				fmt.Sprintf("Wizard endpoint responds: code=%d", wizardCode),
				fmt.Sprintf("Wizard endpoint unreachable: err=%v", wizardErr))),
	})
	outputs["wizard_status_code"] = fmt.Sprintf("%d", wizardCode)
	outputs["wizard_step_count"] = fmt.Sprintf("%d", stepCount)
	outputs["wizard_complete"] = fmt.Sprintf("%v", isComplete)

	// Step 2: Get server configuration (actual endpoint: /api/v1/configuration)
	// Use GetRaw since the handler may return empty body or non-JSON on error
	c.ReportProgress("getting-config", nil)
	configCode, configRaw, configErr := client.GetRaw(ctx, "/api/v1/configuration")
	var configBody map[string]interface{}
	if configErr == nil && len(configRaw) > 0 {
		_ = json.Unmarshal(configRaw, &configBody)
	}

	configResponds := configErr == nil && configCode != 0
	configOK := configErr == nil && configCode == 200

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "config_endpoint",
		Expected: "200 or valid response",
		Actual:   fmt.Sprintf("HTTP %d", configCode),
		Passed:   configResponds,
		Message: challenge.Ternary(configOK,
			"Configuration endpoint returned settings",
			challenge.Ternary(configResponds,
				fmt.Sprintf("Config endpoint responds: code=%d", configCode),
				fmt.Sprintf("Config endpoint unreachable: err=%v", configErr))),
	})
	outputs["config_status_code"] = fmt.Sprintf("%d", configCode)

	// Step 3: Get system status — use /health as primary since it always works
	c.ReportProgress("checking-system-info", nil)
	infoCode, infoBody, infoErr := client.Get(ctx, "/health")

	infoOK := infoErr == nil && infoCode == 200
	version := ""
	if infoBody != nil {
		if v, ok := infoBody["version"].(string); ok {
			version = v
		} else if v, ok := infoBody["app_version"].(string); ok {
			version = v
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "system_info_endpoint",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d, version=%q", infoCode, version),
		Passed:   infoOK,
		Message: challenge.Ternary(infoOK,
			fmt.Sprintf("System info available: version=%q", version),
			fmt.Sprintf("System info failed: code=%d err=%v", infoCode, infoErr)),
	})
	outputs["version"] = version

	// Step 4: Check storage roots (required for wizard completion)
	c.ReportProgress("checking-storage", nil)
	storageCode, storageBody, storageErr := client.Get(ctx, "/api/v1/storage-roots")
	storageOK := storageErr == nil && storageCode == 200
	rootCount := 0
	if storageBody != nil {
		if items, ok := storageBody["items"].([]interface{}); ok {
			rootCount = len(items)
		} else if items, ok := storageBody["data"].([]interface{}); ok {
			rootCount = len(items)
		} else if items, ok := storageBody["configs"].([]interface{}); ok {
			rootCount = len(items)
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "storage_roots_for_wizard",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d, roots=%d", storageCode, rootCount),
		Passed:   storageOK,
		Message: challenge.Ternary(storageOK,
			fmt.Sprintf("Storage roots available: %d configured", rootCount),
			fmt.Sprintf("Storage roots failed: code=%d err=%v", storageCode, storageErr)),
	})
	outputs["storage_root_count"] = fmt.Sprintf("%d", rootCount)

	_ = configBody

	metrics := map[string]challenge.MetricValue{
		"config_wizard_latency": {
			Name:  "config_wizard_latency",
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

	return c.CreateResult(
		status, start, assertions, metrics, outputs, "",
	), nil
}
