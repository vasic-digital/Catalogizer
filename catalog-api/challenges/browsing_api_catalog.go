package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// BrowsingAPICatalogChallenge validates that all browsing API
// endpoints return valid, non-empty data after cataloging.
type BrowsingAPICatalogChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewBrowsingAPICatalogChallenge creates CH-009.
func NewBrowsingAPICatalogChallenge() *BrowsingAPICatalogChallenge {
	return &BrowsingAPICatalogChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"browsing-api-catalog",
			"API Catalog Browsing",
			"Validates all catalog browsing endpoints return non-empty data: "+
				"catalog listing, storage roots, stats, search, and challenges",
			"e2e",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the API catalog browsing challenge.
func (c *BrowsingAPICatalogChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := NewAPIClient(c.config.BaseURL)

	// Authenticate first
	_, err := client.Login(ctx, c.config.Username, c.config.Password)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "login",
			Passed:  false,
			Message: fmt.Sprintf("Login failed: %v", err),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	// 1. GET /api/v1/catalog - non-empty response
	catalogCode, catalogRaw, catalogErr := client.GetRaw(ctx, "/api/v1/catalog")
	catalogOK := catalogErr == nil && catalogCode == 200 && len(catalogRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "catalog_listing",
		Expected: "HTTP 200 with data",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", catalogCode, len(catalogRaw)),
		Passed:   catalogOK,
		Message:  ternary(catalogOK, fmt.Sprintf("Catalog listing returned %d bytes", len(catalogRaw)), fmt.Sprintf("Catalog listing failed: code=%d err=%v", catalogCode, catalogErr)),
	})

	// 2. GET /api/v1/storage/roots - at least 1 root
	rootsCode, rootsRaw, rootsErr := client.GetRaw(ctx, "/api/v1/storage/roots")
	rootsOK := false
	rootCount := 0
	if rootsErr == nil && rootsCode == 200 {
		var roots []interface{}
		if json.Unmarshal(rootsRaw, &roots) == nil {
			rootCount = len(roots)
			rootsOK = rootCount > 0
		} else {
			// May be an object with a data field
			var obj map[string]interface{}
			if json.Unmarshal(rootsRaw, &obj) == nil {
				rootsOK = len(obj) > 0
				rootCount = len(obj)
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "storage_roots",
		Expected: ">= 1 root",
		Actual:   fmt.Sprintf("%d roots", rootCount),
		Passed:   rootsOK,
		Message:  ternary(rootsOK, fmt.Sprintf("Found %d storage roots", rootCount), fmt.Sprintf("Storage roots empty or failed: code=%d err=%v", rootsCode, rootsErr)),
	})
	outputs["storage_root_count"] = fmt.Sprintf("%d", rootCount)

	// 3. GET /api/v1/stats/overall - total_files > 0, total_size > 0
	overallCode, overallBody, overallErr := client.Get(ctx, "/api/v1/stats/overall")
	overallOK := overallErr == nil && overallCode == 200 && overallBody != nil
	totalFiles := float64(0)
	totalSize := float64(0)
	if overallBody != nil {
		if f, ok := overallBody["total_files"].(float64); ok {
			totalFiles = f
		}
		if s, ok := overallBody["total_size"].(float64); ok {
			totalSize = s
		}
	}
	statsOK := overallOK && totalFiles > 0 && totalSize > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "stats_overall",
		Expected: "total_files > 0 and total_size > 0",
		Actual:   fmt.Sprintf("files=%.0f, size=%.0f", totalFiles, totalSize),
		Passed:   statsOK,
		Message:  ternary(statsOK, fmt.Sprintf("Overall stats: %.0f files, %.0f bytes", totalFiles, totalSize), fmt.Sprintf("Stats overall failed: code=%d files=%.0f size=%.0f err=%v", overallCode, totalFiles, totalSize, overallErr)),
	})
	outputs["total_files"] = fmt.Sprintf("%.0f", totalFiles)
	outputs["total_size"] = fmt.Sprintf("%.0f", totalSize)

	// 4. GET /api/v1/stats/filetypes - non-empty
	ftCode, ftRaw, ftErr := client.GetRaw(ctx, "/api/v1/stats/filetypes")
	ftOK := ftErr == nil && ftCode == 200 && len(ftRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "stats_filetypes",
		Expected: "HTTP 200 with data",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", ftCode, len(ftRaw)),
		Passed:   ftOK,
		Message:  ternary(ftOK, "File type stats returned data", fmt.Sprintf("File type stats failed: code=%d err=%v", ftCode, ftErr)),
	})

	// 5. GET /api/v1/stats/sizes - non-empty
	szCode, szRaw, szErr := client.GetRaw(ctx, "/api/v1/stats/sizes")
	szOK := szErr == nil && szCode == 200 && len(szRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "stats_sizes",
		Expected: "HTTP 200 with data",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", szCode, len(szRaw)),
		Passed:   szOK,
		Message:  ternary(szOK, "Size distribution stats returned data", fmt.Sprintf("Size stats failed: code=%d err=%v", szCode, szErr)),
	})

	// 6. GET /api/v1/stats/scans - non-empty
	scCode, scRaw, scErr := client.GetRaw(ctx, "/api/v1/stats/scans")
	scOK := scErr == nil && scCode == 200 && len(scRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "stats_scans",
		Expected: "HTTP 200 with data",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", scCode, len(scRaw)),
		Passed:   scOK,
		Message:  ternary(scOK, "Scan history returned data", fmt.Sprintf("Scan history failed: code=%d err=%v", scCode, scErr)),
	})

	// 7. GET /api/v1/search?q=&page=1&limit=50 - results > 0, no invalid titles
	searchCode, searchBody, searchErr := client.Get(ctx, "/api/v1/search?q=&page=1&limit=50")
	searchOK := searchErr == nil && searchCode == 200 && searchBody != nil
	invalidCount := 0
	resultCount := 0

	if searchBody != nil {
		// Check results array for invalid titles
		if results, ok := searchBody["results"].([]interface{}); ok {
			resultCount = len(results)
			for _, item := range results {
				if m, ok := item.(map[string]interface{}); ok {
					for _, field := range []string{"name", "title"} {
						if val, ok := m[field].(string); ok {
							if IsInvalidTitle(val) {
								invalidCount++
							}
						}
					}
				}
			}
		} else if items, ok := searchBody["items"].([]interface{}); ok {
			resultCount = len(items)
			for _, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					for _, field := range []string{"name", "title"} {
						if val, ok := m[field].(string); ok {
							if IsInvalidTitle(val) {
								invalidCount++
							}
						}
					}
				}
			}
		}
	}

	searchDataOK := searchOK && resultCount > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "search_results",
		Expected: "> 0 results",
		Actual:   fmt.Sprintf("%d results", resultCount),
		Passed:   searchDataOK,
		Message:  ternary(searchDataOK, fmt.Sprintf("Search returned %d results", resultCount), fmt.Sprintf("Search failed: code=%d results=%d err=%v", searchCode, resultCount, searchErr)),
	})

	titleOK := invalidCount == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "title_validation",
		Expected: "0 invalid titles",
		Actual:   fmt.Sprintf("%d invalid titles", invalidCount),
		Passed:   titleOK,
		Message:  ternary(titleOK, "No invalid/placeholder titles found in search results", fmt.Sprintf("Found %d invalid titles in search results", invalidCount)),
	})
	outputs["search_result_count"] = fmt.Sprintf("%d", resultCount)
	outputs["invalid_title_count"] = fmt.Sprintf("%d", invalidCount)

	// 8. GET /api/v1/challenges - count >= 7
	chCode, chRaw, chErr := client.GetRaw(ctx, "/api/v1/challenges")
	chOK := false
	chCount := 0
	if chErr == nil && chCode == 200 {
		var chList []interface{}
		if json.Unmarshal(chRaw, &chList) == nil {
			chCount = len(chList)
			chOK = chCount >= 7
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "challenges_list",
		Expected: ">= 7 challenges",
		Actual:   fmt.Sprintf("%d challenges", chCount),
		Passed:   chOK,
		Message:  ternary(chOK, fmt.Sprintf("Found %d registered challenges", chCount), fmt.Sprintf("Challenges endpoint failed: code=%d count=%d err=%v", chCode, chCount, chErr)),
	})
	outputs["challenge_count"] = fmt.Sprintf("%d", chCount)

	// 9. GET /api/v1/challenges/results - non-empty
	crCode, crRaw, crErr := client.GetRaw(ctx, "/api/v1/challenges/results")
	crOK := crErr == nil && crCode == 200 && len(crRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "challenge_results",
		Expected: "HTTP 200 with data",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", crCode, len(crRaw)),
		Passed:   crOK,
		Message:  ternary(crOK, "Challenge results endpoint returned data", fmt.Sprintf("Challenge results failed: code=%d err=%v", crCode, crErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"total_files": {Name: "total_files", Value: totalFiles, Unit: "count"},
		"total_size":  {Name: "total_size", Value: totalSize, Unit: "bytes"},
		"browse_time": {Name: "browse_time", Value: float64(time.Since(start).Milliseconds()), Unit: "ms"},
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
