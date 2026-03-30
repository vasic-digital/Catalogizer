package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// Challenges CH-171 through CH-185 validate media analysis,
// entity management, and collection API endpoints.

// --- CH-171: MediaTypeDetection ---

// MediaTypeDetectionChallenge validates that the media type
// detection endpoint returns correct type for movie files.
type MediaTypeDetectionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMediaTypeDetectionChallenge creates CH-171.
func NewMediaTypeDetectionChallenge() *MediaTypeDetectionChallenge {
	return &MediaTypeDetectionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"media-type-detection",
			"Media Type Detection",
			"Validates GET /api/v1/entities?type=movie returns "+
				"200 with entities of type movie.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the media type detection challenge.
func (c *MediaTypeDetectionChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-media-type-detection", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities?type=movie&limit=5",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "media_type_detection_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Media type detection returned 200",
			fmt.Sprintf("Media type detection returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "media_type_detection_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Media type detection response is valid JSON",
				"Media type detection response is invalid JSON"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-172: TVShowHierarchy ---

// TVShowHierarchyChallenge validates that TV show entities
// maintain proper hierarchy (show > season > episode).
type TVShowHierarchyChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewTVShowHierarchyChallenge creates CH-172.
func NewTVShowHierarchyChallenge() *TVShowHierarchyChallenge {
	return &TVShowHierarchyChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"tv-show-hierarchy",
			"TV Show Hierarchy",
			"Validates GET /api/v1/entities?type=tv_show "+
				"returns entities with proper hierarchy.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the TV show hierarchy challenge.
func (c *TVShowHierarchyChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-tv-show-hierarchy", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities?type=tv_show&limit=5",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "tv_show_hierarchy_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"TV show hierarchy returned 200",
			fmt.Sprintf("TV show hierarchy returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "tv_show_hierarchy_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"TV show hierarchy response is valid JSON",
				"TV show hierarchy response is invalid JSON"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-173: MusicHierarchy ---

// MusicHierarchyChallenge validates that music entities
// maintain proper hierarchy (artist > album > song).
type MusicHierarchyChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMusicHierarchyChallenge creates CH-173.
func NewMusicHierarchyChallenge() *MusicHierarchyChallenge {
	return &MusicHierarchyChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"music-hierarchy",
			"Music Hierarchy",
			"Validates GET /api/v1/entities?type=music_artist "+
				"returns entities with proper hierarchy.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the music hierarchy challenge.
func (c *MusicHierarchyChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-music-hierarchy", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities?type=music_artist&limit=5",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "music_hierarchy_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Music hierarchy returned 200",
			fmt.Sprintf("Music hierarchy returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "music_hierarchy_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Music hierarchy response is valid JSON",
				"Music hierarchy response is invalid JSON"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-174: DuplicateDetection ---

// DuplicateDetectionChallenge validates that the duplicate
// detection endpoint identifies entities with same title+year.
type DuplicateDetectionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewDuplicateDetectionChallenge creates CH-174.
func NewDuplicateDetectionChallenge() *DuplicateDetectionChallenge {
	return &DuplicateDetectionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"duplicate-detection",
			"Duplicate Detection",
			"Validates GET /api/v1/entities/duplicates returns "+
				"200 with duplicate detection results.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the duplicate detection challenge.
func (c *DuplicateDetectionChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-duplicate-detection", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/duplicates",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "duplicate_detection_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Duplicate detection returned %d", code),
			fmt.Sprintf("Duplicate detection returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "duplicate_detection_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Duplicate detection response is valid JSON",
				"Duplicate detection response is invalid JSON"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-175: EntitySearchAPI ---

// EntitySearchAPIChallenge validates that
// GET /api/v1/entities/search?q=test returns matching entities.
type EntitySearchAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntitySearchAPIChallenge creates CH-175.
func NewEntitySearchAPIChallenge() *EntitySearchAPIChallenge {
	return &EntitySearchAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-search-api",
			"Entity Search API",
			"Validates GET /api/v1/entities/search?q=test "+
				"returns 200 with matching entities.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity search API challenge.
func (c *EntitySearchAPIChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("searching-entities", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/search?q=test",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_search_api_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entity search returned 200",
			fmt.Sprintf("Entity search returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "entity_search_api_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Entity search response is valid JSON",
				"Entity search response is invalid JSON"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-176: EntityPaginationAPI ---

// EntityPaginationAPIChallenge validates that
// GET /api/v1/entities?limit=5&offset=0 respects pagination.
type EntityPaginationAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityPaginationAPIChallenge creates CH-176.
func NewEntityPaginationAPIChallenge() *EntityPaginationAPIChallenge {
	return &EntityPaginationAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-pagination-api",
			"Entity Pagination API",
			"Validates GET /api/v1/entities?limit=5&offset=0 "+
				"respects pagination parameters.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity pagination API challenge.
func (c *EntityPaginationAPIChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("testing-entity-pagination", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities?limit=5&offset=0",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_pagination_api_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entity pagination returned 200",
			fmt.Sprintf("Entity pagination returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "entity_pagination_api_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Entity pagination response is valid JSON",
				"Entity pagination response is invalid JSON"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-177: MediaItemCreate ---

// MediaItemCreateChallenge validates POST /api/v1/entities
// creates a media item.
type MediaItemCreateChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMediaItemCreateChallenge creates CH-177.
func NewMediaItemCreateChallenge() *MediaItemCreateChallenge {
	return &MediaItemCreateChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"media-item-create",
			"Media Item Create",
			"Validates POST /api/v1/entities creates a media "+
				"item and returns 200 or 201.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the media item create challenge.
func (c *MediaItemCreateChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("creating-media-item", nil)
	body := `{"title":"Challenge Test Movie","type":"movie","year":2024}`
	code, _, err := client.PostJSON(
		ctx, "/api/v1/entities", body,
	)

	codeOK := err == nil &&
		(code == 200 || code == 201 || code == 400)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "media_item_create_status",
		Expected: "200, 201, or 400",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Media item create returned %d", code),
			fmt.Sprintf("Media item create returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-178: MediaItemUpdate ---

// MediaItemUpdateChallenge validates PUT /api/v1/entities/1
// updates a media item.
type MediaItemUpdateChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMediaItemUpdateChallenge creates CH-178.
func NewMediaItemUpdateChallenge() *MediaItemUpdateChallenge {
	return &MediaItemUpdateChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"media-item-update",
			"Media Item Update",
			"Validates PUT /api/v1/entities/1 updates a media "+
				"item and returns 200 or 404.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the media item update challenge.
func (c *MediaItemUpdateChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("updating-media-item", nil)
	body := `{"title":"Updated Challenge Movie","year":2025}`
	code, _, err := client.PutJSON(
		ctx, "/api/v1/entities/1", body,
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "media_item_update_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Media item update returned %d", code),
			fmt.Sprintf("Media item update returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-179: MediaFileLink ---

// MediaFileLinkChallenge validates POST /api/v1/entities/1/files
// links a file to a media item.
type MediaFileLinkChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMediaFileLinkChallenge creates CH-179.
func NewMediaFileLinkChallenge() *MediaFileLinkChallenge {
	return &MediaFileLinkChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"media-file-link",
			"Media File Link",
			"Validates POST /api/v1/entities/1/files links a "+
				"file to a media item.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the media file link challenge.
func (c *MediaFileLinkChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("linking-media-file", nil)
	body := `{"file_id":1}`
	code, _, err := client.PostJSON(
		ctx, "/api/v1/entities/1/files", body,
	)

	codeOK := err == nil &&
		(code == 200 || code == 201 || code == 400 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "media_file_link_status",
		Expected: "200, 201, 400, or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Media file link returned %d", code),
			fmt.Sprintf("Media file link returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-180: ExternalMetadata ---

// ExternalMetadataChallenge validates GET
// /api/v1/entities/1/metadata/external returns external metadata.
type ExternalMetadataChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewExternalMetadataChallenge creates CH-180.
func NewExternalMetadataChallenge() *ExternalMetadataChallenge {
	return &ExternalMetadataChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"external-metadata",
			"External Metadata",
			"Validates GET /api/v1/entities/1/metadata/external "+
				"returns 200 or 404.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the external metadata challenge.
func (c *ExternalMetadataChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("fetching-external-metadata", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities/1/metadata/external",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "external_metadata_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("External metadata returned %d", code),
			fmt.Sprintf("External metadata returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-181: UserMetadataAPI ---

// UserMetadataAPIChallenge validates GET
// /api/v1/entities/1/metadata/user returns user metadata.
type UserMetadataAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewUserMetadataAPIChallenge creates CH-181.
func NewUserMetadataAPIChallenge() *UserMetadataAPIChallenge {
	return &UserMetadataAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"user-metadata-api",
			"User Metadata API",
			"Validates GET /api/v1/entities/1/metadata/user "+
				"returns 200 or 404.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the user metadata API challenge.
func (c *UserMetadataAPIChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("fetching-user-metadata", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities/1/metadata/user",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "user_metadata_api_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("User metadata returned %d", code),
			fmt.Sprintf("User metadata returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-182: CollectionCreateAPI ---

// CollectionCreateAPIChallenge validates POST /api/v1/collections
// creates a collection.
type CollectionCreateAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCollectionCreateAPIChallenge creates CH-182.
func NewCollectionCreateAPIChallenge() *CollectionCreateAPIChallenge {
	return &CollectionCreateAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"collection-create-api",
			"Collection Create API",
			"Validates POST /api/v1/collections creates a "+
				"collection and returns 200 or 201.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the collection create API challenge.
func (c *CollectionCreateAPIChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("creating-collection", nil)
	body := `{"name":"Challenge Test Collection","description":"test"}`
	code, _, err := client.PostJSON(
		ctx, "/api/v1/collections", body,
	)

	codeOK := err == nil && (code == 200 || code == 201)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "collection_create_api_status",
		Expected: "200 or 201",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Collection create returned %d", code),
			fmt.Sprintf("Collection create returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-183: CollectionItemsAPI ---

// CollectionItemsAPIChallenge validates POST
// /api/v1/collections/1/items adds items to a collection.
type CollectionItemsAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCollectionItemsAPIChallenge creates CH-183.
func NewCollectionItemsAPIChallenge() *CollectionItemsAPIChallenge {
	return &CollectionItemsAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"collection-items-api",
			"Collection Items API",
			"Validates POST /api/v1/collections/1/items adds "+
				"items to a collection.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the collection items API challenge.
func (c *CollectionItemsAPIChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("adding-collection-items", nil)
	body := `{"media_item_id":1}`
	code, _, err := client.PostJSON(
		ctx, "/api/v1/collections/1/items", body,
	)

	codeOK := err == nil &&
		(code == 200 || code == 201 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "collection_items_api_status",
		Expected: "200, 201, or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Collection items returned %d", code),
			fmt.Sprintf("Collection items returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-184: CollectionSearchAPI ---

// CollectionSearchAPIChallenge validates GET
// /api/v1/collections?q=test searches collections.
type CollectionSearchAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCollectionSearchAPIChallenge creates CH-184.
func NewCollectionSearchAPIChallenge() *CollectionSearchAPIChallenge {
	return &CollectionSearchAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"collection-search-api",
			"Collection Search API",
			"Validates GET /api/v1/collections?q=test searches "+
				"collections and returns 200.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the collection search API challenge.
func (c *CollectionSearchAPIChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("searching-collections", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/collections?q=test",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "collection_search_api_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Collection search returned 200",
			fmt.Sprintf("Collection search returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "collection_search_api_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Collection search response is valid JSON",
				"Collection search response is invalid JSON"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-185: DirectoryAnalysis ---

// DirectoryAnalysisChallenge validates GET
// /api/v1/entities/directory-analysis returns directory
// analysis records.
type DirectoryAnalysisChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewDirectoryAnalysisChallenge creates CH-185.
func NewDirectoryAnalysisChallenge() *DirectoryAnalysisChallenge {
	return &DirectoryAnalysisChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"directory-analysis",
			"Directory Analysis",
			"Validates GET /api/v1/entities/directory-analysis "+
				"returns 200 with analysis records.",
			"media-analysis",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the directory analysis challenge.
func (c *DirectoryAnalysisChallenge) Execute(
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
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("fetching-directory-analysis", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/directory-analysis",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "directory_analysis_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Directory analysis returned %d", code),
			fmt.Sprintf("Directory analysis returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "directory_analysis_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Directory analysis response is valid JSON",
				"Directory analysis response is invalid JSON"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}
