package challenges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// ConfigDocsChallenge validates that all configuration options
// are documented by checking documentation files for references
// to key environment variables and configuration keys.
type ConfigDocsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConfigDocsChallenge creates CH-047.
func NewConfigDocsChallenge() *ConfigDocsChallenge {
	return &ConfigDocsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"config-docs",
			"Configuration Documentation Completeness",
			"Verifies all configuration options are documented: checks "+
				"that key environment variables (PORT, GIN_MODE, DB_TYPE, "+
				"JWT_SECRET, etc.) appear in documentation files.",
			"documentation",
			nil,
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the configuration documentation challenge.
func (c *ConfigDocsChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	// Key configuration options that should be documented
	keyConfigs := []string{
		"PORT",
		"GIN_MODE",
		"DB_TYPE",
		"JWT_SECRET",
		"ADMIN_PASSWORD",
		"TMDB_API_KEY",
	}

	// Search for documentation files
	c.ReportProgress("scanning-docs", nil)
	docsDirs := []string{
		"../docs",
		"docs",
		".",
		"..",
	}

	// Collect all doc content
	var allDocContent strings.Builder
	docFileCount := 0

	for _, docDir := range docsDirs {
		_ = filepath.Walk(docDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			name := info.Name()
			if ext == ".md" || ext == ".txt" || name == ".env.example" || name == "CLAUDE.md" {
				content, readErr := os.ReadFile(path)
				if readErr == nil {
					allDocContent.Write(content)
					allDocContent.WriteString("\n")
					docFileCount++
				}
			}
			return nil
		})
	}

	outputs["doc_files_scanned"] = fmt.Sprintf("%d", docFileCount)

	// Check for doc files
	hasDocFiles := docFileCount > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "doc_files",
		Expected: "documentation files exist",
		Actual:   fmt.Sprintf("%d files", docFileCount),
		Passed:   hasDocFiles,
		Message: challenge.Ternary(hasDocFiles,
			fmt.Sprintf("Found %d documentation files to scan", docFileCount),
			"No documentation files found"),
	})

	if !hasDocFiles {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "no docs found",
		), nil
	}

	// Check each key config is mentioned
	c.ReportProgress("checking-configs", nil)
	docContent := allDocContent.String()
	documented := 0
	undocumented := []string{}

	for _, cfg := range keyConfigs {
		found := strings.Contains(docContent, cfg)
		if found {
			documented++
		} else {
			undocumented = append(undocumented, cfg)
		}
	}

	configRatio := float64(documented) / float64(len(keyConfigs)) * 100
	configPassed := documented >= 4 // at least 4 of 6 key configs documented
	outputs["configs_documented"] = fmt.Sprintf("%d/%d", documented, len(keyConfigs))
	if len(undocumented) > 0 {
		outputs["undocumented"] = strings.Join(undocumented, ", ")
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_score",
		Target:   "config_documented",
		Expected: "at least 4/6 key config options documented",
		Actual:   fmt.Sprintf("%d/%d (%.0f%%)", documented, len(keyConfigs), configRatio),
		Passed:   configPassed,
		Message: challenge.Ternary(configPassed,
			fmt.Sprintf("%d/%d key configuration options documented", documented, len(keyConfigs)),
			fmt.Sprintf("Only %d/%d key configs documented, missing: %v", documented, len(keyConfigs), undocumented)),
	})

	// Check for config.json example or .env.example
	c.ReportProgress("checking-examples", nil)
	exampleFiles := []string{
		".env.example",
		"config.json.example",
		"config.json",
		".env",
	}

	hasExample := false
	for _, f := range exampleFiles {
		if _, err := os.Stat(f); err == nil {
			hasExample = true
			outputs["example_config"] = f
			break
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "config_example_file",
		Expected: "example config file exists",
		Actual:   challenge.Ternary(hasExample, outputs["example_config"], "not found"),
		Passed:   hasExample,
		Message: challenge.Ternary(hasExample,
			fmt.Sprintf("Example config file found: %s", outputs["example_config"]),
			"No example configuration file found (.env.example, config.json.example)"),
	})

	metrics := map[string]challenge.MetricValue{
		"configs_documented": {
			Name:  "configs_documented",
			Value: float64(documented),
			Unit:  "count",
		},
		"doc_files_scanned": {
			Name:  "doc_files_scanned",
			Value: float64(docFileCount),
			Unit:  "count",
		},
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, metrics, outputs, ""), nil
}
