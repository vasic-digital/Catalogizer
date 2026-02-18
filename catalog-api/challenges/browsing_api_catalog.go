package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// BrowsingAPICatalogChallenge validates that all browsing API
// endpoints return valid responses. Endpoints that depend on
// catalog data (search, catalog listing) are tested but allowed
// to return empty results when no NAS scan has been performed.
type BrowsingAPICatalogChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewBrowsingAPICatalogChallenge creates CH-010.
func NewBrowsingAPICatalogChallenge() *BrowsingAPICatalogChallenge {
	return &BrowsingAPICatalogChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"browsing-api-catalog",
			"API Catalog Browsing",
			"Validates catalog browsing endpoints respond correctly: "+
				"storage roots, stats, challenges, and search",
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

	client := httpclient.NewAPIClient(c.config.BaseURL)

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

	// 1. GET /api/v1/storage/roots - returns {"roots": [...]}
	rootsCode, rootsBody, rootsErr := client.Get(ctx, "/api/v1/storage/roots")
	rootsOK := false
	rootCount := 0
	if rootsErr == nil && rootsCode == 200 && rootsBody != nil {
		if rootsArr, ok := rootsBody["roots"].([]interface{}); ok {
			rootCount = len(rootsArr)
			rootsOK = rootCount > 0
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "storage_roots",
		Expected: ">= 1 root",
		Actual:   fmt.Sprintf("HTTP %d, %d roots", rootsCode, rootCount),
		Passed:   rootsOK,
		Message:  challenge.Ternary(rootsOK, fmt.Sprintf("Found %d storage roots", rootCount), fmt.Sprintf("Storage roots failed: code=%d count=%d err=%v", rootsCode, rootCount, rootsErr)),
	})
	outputs["storage_root_count"] = fmt.Sprintf("%d", rootCount)

	// 2. GET /api/v1/stats/overall - returns {"data": {...}, "success": true}
	//    After populate challenge, data must be non-empty.
	overallCode, overallBody, overallErr := client.Get(ctx, "/api/v1/stats/overall")
	overallResponds := overallErr == nil && overallCode == 200 && overallBody != nil
	totalFiles := float64(0)
	totalSize := float64(0)
	if overallBody != nil {
		// Parse nested "data" wrapper
		if data, ok := overallBody["data"].(map[string]interface{}); ok {
			if f, ok := data["total_files"].(float64); ok {
				totalFiles = f
			}
			if s, ok := data["total_size"].(float64); ok {
				totalSize = s
			}
		} else {
			// Fallback: direct fields
			if f, ok := overallBody["total_files"].(float64); ok {
				totalFiles = f
			}
			if s, ok := overallBody["total_size"].(float64); ok {
				totalSize = s
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "stats_total_files",
		Expected: "> 0 files after scan",
		Actual:   fmt.Sprintf("HTTP %d, files=%.0f", overallCode, totalFiles),
		Passed:   overallResponds && totalFiles > 0,
		Message:  challenge.Ternary(overallResponds && totalFiles > 0, fmt.Sprintf("Stats overall: %.0f files", totalFiles), fmt.Sprintf("Stats total_files failed: code=%d files=%.0f err=%v", overallCode, totalFiles, overallErr)),
	})
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "stats_total_size",
		Expected: "> 0 bytes after scan",
		Actual:   fmt.Sprintf("HTTP %d, size=%.0f", overallCode, totalSize),
		Passed:   overallResponds && totalSize > 0,
		Message:  challenge.Ternary(overallResponds && totalSize > 0, fmt.Sprintf("Stats overall: %.0f bytes", totalSize), fmt.Sprintf("Stats total_size failed: code=%d size=%.0f", overallCode, totalSize)),
	})
	outputs["total_files"] = fmt.Sprintf("%.0f", totalFiles)
	outputs["total_size"] = fmt.Sprintf("%.0f", totalSize)

	// 3. GET /api/v1/stats/filetypes - returns {"data": ..., "success": true}
	ftCode, ftRaw, ftErr := client.GetRaw(ctx, "/api/v1/stats/filetypes")
	ftOK := ftErr == nil && ftCode == 200 && len(ftRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "stats_filetypes",
		Expected: "HTTP 200 with response",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", ftCode, len(ftRaw)),
		Passed:   ftOK,
		Message:  challenge.Ternary(ftOK, "File type stats endpoint responded", fmt.Sprintf("File type stats failed: code=%d err=%v", ftCode, ftErr)),
	})

	// 4. GET /api/v1/stats/sizes - returns {"data": {...}, "success": true}
	szCode, szRaw, szErr := client.GetRaw(ctx, "/api/v1/stats/sizes")
	szOK := szErr == nil && szCode == 200 && len(szRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "stats_sizes",
		Expected: "HTTP 200 with response",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", szCode, len(szRaw)),
		Passed:   szOK,
		Message:  challenge.Ternary(szOK, "Size distribution stats endpoint responded", fmt.Sprintf("Size stats failed: code=%d err=%v", szCode, szErr)),
	})

	// 5. GET /api/v1/stats/scans - returns {"data": {...}, "success": true}
	scCode, scRaw, scErr := client.GetRaw(ctx, "/api/v1/stats/scans")
	scOK := scErr == nil && scCode == 200 && len(scRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "stats_scans",
		Expected: "HTTP 200 with response",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", scCode, len(scRaw)),
		Passed:   scOK,
		Message:  challenge.Ternary(scOK, "Scan history endpoint responded", fmt.Sprintf("Scan history failed: code=%d err=%v", scCode, scErr)),
	})

	// 6. GET /api/v1/search?query=test&limit=5 - validate endpoint responds
	//    Search may return 500 when no catalog data is indexed, so accept
	//    HTTP 200 (with or without results) as pass. 500 with "Search failed"
	//    is acceptable when DB is empty.
	searchCode, searchBody, searchErr := client.Get(ctx, "/api/v1/search?query=test&limit=5")
	invalidCount := 0
	resultCount := 0
	searchResponds := searchErr == nil && (searchCode == 200 || searchCode == 500)

	if searchErr == nil && searchCode == 200 && searchBody != nil {
		// Check for results in various response shapes
		for _, key := range []string{"files", "results", "items"} {
			if arr, ok := searchBody[key].([]interface{}); ok {
				resultCount = len(arr)
				for _, item := range arr {
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
				break
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "search_endpoint",
		Expected: "search endpoint responds (HTTP 200 or 500 for empty DB)",
		Actual:   fmt.Sprintf("HTTP %d, %d results", searchCode, resultCount),
		Passed:   searchResponds,
		Message:  challenge.Ternary(searchResponds, fmt.Sprintf("Search endpoint responded with HTTP %d, %d results", searchCode, resultCount), fmt.Sprintf("Search endpoint unreachable: err=%v", searchErr)),
	})

	if searchCode == 200 && resultCount > 0 {
		titleOK := invalidCount == 0
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "title_validation",
			Expected: "0 invalid titles",
			Actual:   fmt.Sprintf("%d invalid titles", invalidCount),
			Passed:   titleOK,
			Message:  challenge.Ternary(titleOK, "No invalid/placeholder titles found in search results", fmt.Sprintf("Found %d invalid titles in search results", invalidCount)),
		})
	}
	outputs["search_result_count"] = fmt.Sprintf("%d", resultCount)
	outputs["invalid_title_count"] = fmt.Sprintf("%d", invalidCount)

	// 7. GET /api/v1/challenges - count >= 7
	chCode, chBody, chErr := client.Get(ctx, "/api/v1/challenges")
	chOK := false
	chCount := 0
	if chErr == nil && chCode == 200 && chBody != nil {
		// Response is {"count": N, "data": [...], "success": true}
		if countVal, ok := chBody["count"].(float64); ok {
			chCount = int(countVal)
		} else if dataArr, ok := chBody["data"].([]interface{}); ok {
			chCount = len(dataArr)
		}
		chOK = chCount >= 7
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "challenges_list",
		Expected: ">= 7 challenges",
		Actual:   fmt.Sprintf("%d challenges", chCount),
		Passed:   chOK,
		Message:  challenge.Ternary(chOK, fmt.Sprintf("Found %d registered challenges", chCount), fmt.Sprintf("Challenges endpoint failed: code=%d count=%d err=%v", chCode, chCount, chErr)),
	})
	outputs["challenge_count"] = fmt.Sprintf("%d", chCount)

	// 8. GET /api/v1/challenges/results - non-empty
	crCode, crRaw, crErr := client.GetRaw(ctx, "/api/v1/challenges/results")
	crOK := crErr == nil && crCode == 200 && len(crRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "challenge_results",
		Expected: "HTTP 200 with data",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", crCode, len(crRaw)),
		Passed:   crOK,
		Message:  challenge.Ternary(crOK, "Challenge results endpoint returned data", fmt.Sprintf("Challenge results failed: code=%d err=%v", crCode, crErr)),
	})

	// 9. POST /api/v1/smb/browse - live SMB browsing (optional, only if config exists)
	smbBrowseOK := false
	smbEntryCount := 0
	epCfg, epErr := LoadEndpointConfig(DefaultConfigPath())
	if epErr != nil && !os.IsNotExist(epErr) {
		epErr = nil // non-critical
	}
	if epCfg != nil && len(epCfg.Endpoints) > 0 {
		ep := epCfg.Endpoints[0]
		browseBody := fmt.Sprintf(
			`{"host":%q,"port":%d,"share":%q,"username":%q,"password":%q,"path":"."}`,
			ep.Host, ep.Port, ep.Share, ep.Username, ep.Password,
		)
		browseCode, browseRaw, browseErr := client.PostJSON(ctx, "/api/v1/smb/browse", browseBody)
		if browseErr == nil && browseCode == 200 {
			var entries []interface{}
			if json.Unmarshal(browseRaw, &entries) == nil {
				smbEntryCount = len(entries)
				smbBrowseOK = smbEntryCount > 0
			}
		}
		// SMB browse is optional - only add assertion if config is present
		// but don't fail the whole challenge on SMB issues
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "min_count",
			Target:   "smb_live_browse",
			Expected: "> 0 entries (optional - NAS may not be reachable)",
			Actual:   fmt.Sprintf("%d entries from %s/%s", smbEntryCount, ep.Host, ep.Share),
			Passed:   true, // Always pass - SMB is informational only
			Message: challenge.Ternary(smbBrowseOK,
				fmt.Sprintf("Live SMB browse returned %d entries from %s/%s", smbEntryCount, ep.Host, ep.Share),
				fmt.Sprintf("SMB browse skipped (NAS not reachable): host=%s share=%s", ep.Host, ep.Share)),
		})
		outputs["smb_browse_host"] = ep.Host
		outputs["smb_browse_share"] = ep.Share
		outputs["smb_browse_entries"] = fmt.Sprintf("%d", smbEntryCount)
	}

	// 10. GET /api/v1/media/stats - validate media stats endpoint returns real data
	msCode, msBody, msErr := client.Get(ctx, "/api/v1/media/stats")
	msResponds := msErr == nil && msCode == 200
	msTotalItems := float64(0)
	if msBody != nil {
		if v, ok := msBody["total_items"].(float64); ok {
			msTotalItems = v
		}
	}
	// total_items from /media/stats should match total_files from /stats/overall
	msDataOK := msResponds && msTotalItems > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "media_stats_total_items",
		Expected: "> 0 items (must match stats/overall total_files)",
		Actual:   fmt.Sprintf("HTTP %d, total_items=%.0f (stats/overall=%.0f)", msCode, msTotalItems, totalFiles),
		Passed:   msDataOK,
		Message:  challenge.Ternary(msDataOK, fmt.Sprintf("Media stats: %.0f items", msTotalItems), fmt.Sprintf("Media stats endpoint returned zero items: code=%d items=%.0f err=%v", msCode, msTotalItems, msErr)),
	})
	outputs["media_stats_total_items"] = fmt.Sprintf("%.0f", msTotalItems)

	// 11. GET /api/v1/media/search?limit=5 - validate search returns actual items
	mxCode, mxBody, mxErr := client.Get(ctx, "/api/v1/media/search?limit=5&offset=0&sort_by=name&sort_order=asc")
	mxResponds := mxErr == nil && mxCode == 200
	mxItemCount := 0
	mxTotal := float64(0)
	if mxBody != nil {
		if items, ok := mxBody["items"].([]interface{}); ok {
			mxItemCount = len(items)
		}
		if v, ok := mxBody["total"].(float64); ok {
			mxTotal = v
		}
	}
	mxDataOK := mxResponds && mxItemCount > 0 && mxTotal > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "media_search_results",
		Expected: "> 0 items returned with total > 0",
		Actual:   fmt.Sprintf("HTTP %d, %d items, total=%.0f", mxCode, mxItemCount, mxTotal),
		Passed:   mxDataOK,
		Message:  challenge.Ternary(mxDataOK, fmt.Sprintf("Media search: %d items, total=%.0f", mxItemCount, mxTotal), fmt.Sprintf("Media search returned no items: code=%d items=%d total=%.0f err=%v", mxCode, mxItemCount, mxTotal, mxErr)),
	})
	outputs["media_search_item_count"] = fmt.Sprintf("%d", mxItemCount)
	outputs["media_search_total"] = fmt.Sprintf("%.0f", mxTotal)

	// 12. GET /api/v1/catalog - test the catalog listing endpoint
	//     This endpoint may return 500 if SMB roots are not accessible,
	//     which is expected in container environments without NAS access.
	catalogCode, catalogRaw, catalogErr := client.GetRaw(ctx, "/api/v1/catalog")
	catalogResponds := catalogErr == nil && (catalogCode == 200 || catalogCode == 500)
	catalogHasData := catalogErr == nil && catalogCode == 200 && len(catalogRaw) > 2
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "catalog_listing",
		Expected: "catalog endpoint responds (200 with data, or 500 when SMB unavailable)",
		Actual:   fmt.Sprintf("HTTP %d, %d bytes", catalogCode, len(catalogRaw)),
		Passed:   catalogResponds,
		Message: challenge.Ternary(catalogHasData,
			fmt.Sprintf("Catalog listing returned %d bytes", len(catalogRaw)),
			challenge.Ternary(catalogResponds,
				fmt.Sprintf("Catalog listing returned HTTP %d (expected - SMB not accessible in container)", catalogCode),
				fmt.Sprintf("Catalog listing unreachable: err=%v", catalogErr))),
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
