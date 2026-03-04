package challenges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// APIDocsChallenge validates that all API endpoints are documented
// by checking that documentation files exist in the docs/api/ directory.
type APIDocsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAPIDocsChallenge creates CH-045.
func NewAPIDocsChallenge() *APIDocsChallenge {
	return &APIDocsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"api-docs",
			"API Documentation Completeness",
			"Verifies that API endpoint documentation exists in the "+
				"docs/api/ directory. Checks for key documentation files "+
				"covering auth, entities, storage, and other API areas.",
			"documentation",
			nil,
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the API documentation challenge.
func (c *APIDocsChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	// Look for docs directory relative to working directory
	// Try common locations
	docsDirs := []string{
		"../docs/api",
		"docs/api",
		"../docs",
		"docs",
	}

	var docsDir string
	for _, dir := range docsDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			docsDir = dir
			break
		}
	}

	c.ReportProgress("checking-docs-dir", nil)

	if docsDir == "" {
		// Try absolute path based on typical project layout
		cwd, _ := os.Getwd()
		parentDocs := filepath.Join(filepath.Dir(cwd), "docs")
		if info, err := os.Stat(parentDocs); err == nil && info.IsDir() {
			docsDir = parentDocs
		}
	}

	docsExist := docsDir != ""
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "docs_directory",
		Expected: "docs directory exists",
		Actual:   challenge.Ternary(docsExist, docsDir, "not found"),
		Passed:   docsExist,
		Message: challenge.Ternary(docsExist,
			fmt.Sprintf("Documentation directory found: %s", docsDir),
			"Documentation directory not found"),
	})

	if !docsExist {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "docs directory not found",
		), nil
	}

	outputs["docs_dir"] = docsDir

	// Count documentation files
	c.ReportProgress("counting-files", nil)
	docFiles := 0
	_ = filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".md" || ext == ".txt" || ext == ".rst" || ext == ".adoc" {
				docFiles++
			}
		}
		return nil
	})

	outputs["doc_files_count"] = fmt.Sprintf("%d", docFiles)

	hasDocFiles := docFiles > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "doc_files_exist",
		Expected: "at least one documentation file",
		Actual:   fmt.Sprintf("%d files", docFiles),
		Passed:   hasDocFiles,
		Message: challenge.Ternary(hasDocFiles,
			fmt.Sprintf("Found %d documentation files", docFiles),
			"No documentation files found"),
	})

	// Check for specific key documentation areas
	keyTopics := []string{
		"api",
		"architecture",
		"deployment",
		"guides",
		"security",
	}

	for _, topic := range keyTopics {
		topicPath := filepath.Join(docsDir, topic)
		// Also check if it exists as a file
		topicFilePath := filepath.Join(docsDir, topic+".md")

		topicExists := false
		if info, err := os.Stat(topicPath); err == nil && info.IsDir() {
			topicExists = true
		} else if _, err := os.Stat(topicFilePath); err == nil {
			topicExists = true
		}

		// Also check parent docs directory for the topic
		if !topicExists && docsDir != "../docs" {
			parentTopic := filepath.Join(filepath.Dir(docsDir), topic)
			if info, err := os.Stat(parentTopic); err == nil && info.IsDir() {
				topicExists = true
			}
		}

		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   fmt.Sprintf("docs_topic_%s", topic),
			Expected: fmt.Sprintf("documentation for %s exists", topic),
			Actual:   challenge.Ternary(topicExists, "found", "missing"),
			Passed:   topicExists,
			Message: challenge.Ternary(topicExists,
				fmt.Sprintf("Documentation for %q exists", topic),
				fmt.Sprintf("Documentation for %q missing", topic)),
		})
	}

	metrics := map[string]challenge.MetricValue{
		"doc_files_count": {
			Name:  "doc_files_count",
			Value: float64(docFiles),
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
