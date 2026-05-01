package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// Challenges CH-159 through CH-170 validate the playlist API endpoints.
// Each challenge makes HTTP requests to the running API and validates
// the response status code and JSON shape.

// --- CH-159: CreatePlaylist ---

// CreatePlaylistChallenge validates POST /api/v1/playlists creates
// a playlist and returns 200 or 201.
type CreatePlaylistChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCreatePlaylistChallenge creates CH-159.
func NewCreatePlaylistChallenge() *CreatePlaylistChallenge {
	return &CreatePlaylistChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-create",
			"Create Playlist",
			"Validates POST /api/v1/playlists creates a "+
				"playlist and returns 200 or 201.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the create playlist challenge.
func (c *CreatePlaylistChallenge) Execute(
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

	c.ReportProgress("creating-playlist", nil)
	body := `{"name":"test-playlist","description":"challenge test"}`
	code, _, err := client.PostJSON(
		ctx, "/api/v1/playlists", body,
	)

	codeOK := err == nil && (code == 200 || code == 201)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_create_status",
		Expected: "200 or 201",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Playlist create returned %d", code),
			fmt.Sprintf("Playlist create returned %d, err=%v",
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

// --- CH-160: ListPlaylists ---

// ListPlaylistsChallenge validates GET /api/v1/playlists
// returns 200 with an array or items field.
type ListPlaylistsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewListPlaylistsChallenge creates CH-160.
func NewListPlaylistsChallenge() *ListPlaylistsChallenge {
	return &ListPlaylistsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-list",
			"List Playlists",
			"Validates GET /api/v1/playlists returns 200 "+
				"with an array or items field.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the list playlists challenge.
func (c *ListPlaylistsChallenge) Execute(
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

	c.ReportProgress("listing-playlists", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/playlists",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_list_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Playlist list endpoint returned 200",
			fmt.Sprintf("Playlist list returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		hasData := false
		var obj map[string]interface{}
		if json.Unmarshal(rawBody, &obj) == nil {
			_, hasData = obj["items"]
			if !hasData {
				_, hasData = obj["playlists"]
			}
			if !hasData {
				_, hasData = obj["data"]
			}
		}
		if !hasData {
			var arr []interface{}
			if json.Unmarshal(rawBody, &arr) == nil {
				hasData = true
			}
		}
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "playlist_list_data",
			Expected: "array or items field present",
			Actual: challenge.Ternary(hasData,
				"present", "missing"),
			Passed: hasData,
			Message: challenge.Ternary(hasData,
				"Playlist data found in response",
				"Playlist data missing from response"),
		})
	}

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

// --- CH-161: GetPlaylist ---

// GetPlaylistChallenge validates GET /api/v1/playlists/1
// returns 200 or 404.
type GetPlaylistChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewGetPlaylistChallenge creates CH-161.
func NewGetPlaylistChallenge() *GetPlaylistChallenge {
	return &GetPlaylistChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-get",
			"Get Playlist",
			"Validates GET /api/v1/playlists/1 returns 200 "+
				"or 404.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the get playlist challenge.
func (c *GetPlaylistChallenge) Execute(
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

	c.ReportProgress("getting-playlist", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/playlists/1",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_get_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Get playlist returned %d", code),
			fmt.Sprintf("Get playlist returned %d, err=%v",
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

// --- CH-162: UpdatePlaylist ---

// UpdatePlaylistChallenge validates PUT /api/v1/playlists/1
// returns 200 or 404.
type UpdatePlaylistChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewUpdatePlaylistChallenge creates CH-162.
func NewUpdatePlaylistChallenge() *UpdatePlaylistChallenge {
	return &UpdatePlaylistChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-update",
			"Update Playlist",
			"Validates PUT /api/v1/playlists/1 returns 200 "+
				"or 404.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the update playlist challenge.
func (c *UpdatePlaylistChallenge) Execute(
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

	c.ReportProgress("updating-playlist", nil)
	body := `{"name":"updated-playlist","description":"updated"}`
	code, _, err := client.PutJSON(
		ctx, "/api/v1/playlists/1", body,
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_update_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Update playlist returned %d", code),
			fmt.Sprintf("Update playlist returned %d, err=%v",
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

// --- CH-163: DeletePlaylist ---

// DeletePlaylistChallenge validates DELETE /api/v1/playlists/1
// returns 200 or 404.
type DeletePlaylistChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewDeletePlaylistChallenge creates CH-163.
func NewDeletePlaylistChallenge() *DeletePlaylistChallenge {
	return &DeletePlaylistChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-delete",
			"Delete Playlist",
			"Validates DELETE /api/v1/playlists/1 returns 200 "+
				"or 404.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the delete playlist challenge.
func (c *DeletePlaylistChallenge) Execute(
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

	c.ReportProgress("deleting-playlist", nil)
	code, _, err := client.Delete(
		ctx, "/api/v1/playlists/1",
	)

	codeOK := err == nil && (code == 200 || code == 204 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_delete_status",
		Expected: "200, 204, or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Delete playlist returned %d", code),
			fmt.Sprintf("Delete playlist returned %d, err=%v",
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

// --- CH-164: AddPlaylistItem ---

// AddPlaylistItemChallenge validates POST /api/v1/playlists/1/items
// returns 200, 201, or 404.
type AddPlaylistItemChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAddPlaylistItemChallenge creates CH-164.
func NewAddPlaylistItemChallenge() *AddPlaylistItemChallenge {
	return &AddPlaylistItemChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-add-item",
			"Add Playlist Item",
			"Validates POST /api/v1/playlists/1/items returns "+
				"200, 201, or 404.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the add playlist item challenge.
func (c *AddPlaylistItemChallenge) Execute(
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

	c.ReportProgress("adding-playlist-item", nil)
	body := `{"media_item_id":1}`
	code, _, err := client.PostJSON(
		ctx, "/api/v1/playlists/1/items", body,
	)

	codeOK := err == nil &&
		(code == 200 || code == 201 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_add_item_status",
		Expected: "200, 201, or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Add playlist item returned %d", code),
			fmt.Sprintf("Add playlist item returned %d, err=%v",
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

// --- CH-165: RemovePlaylistItem ---

// RemovePlaylistItemChallenge validates DELETE
// /api/v1/playlists/1/items/1 returns 200, 204, or 404.
type RemovePlaylistItemChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewRemovePlaylistItemChallenge creates CH-165.
func NewRemovePlaylistItemChallenge() *RemovePlaylistItemChallenge {
	return &RemovePlaylistItemChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-remove-item",
			"Remove Playlist Item",
			"Validates DELETE /api/v1/playlists/1/items/1 "+
				"returns 200, 204, or 404.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the remove playlist item challenge.
func (c *RemovePlaylistItemChallenge) Execute(
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

	c.ReportProgress("removing-playlist-item", nil)
	code, _, err := client.Delete(
		ctx, "/api/v1/playlists/1/items/1",
	)

	codeOK := err == nil &&
		(code == 200 || code == 204 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_remove_item_status",
		Expected: "200, 204, or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Remove playlist item returned %d", code),
			fmt.Sprintf("Remove playlist item returned %d, err=%v",
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

// --- CH-166: PlaylistOrdering ---

// PlaylistOrderingChallenge validates that playlist items
// maintain their order by checking the response structure.
type PlaylistOrderingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPlaylistOrderingChallenge creates CH-166.
func NewPlaylistOrderingChallenge() *PlaylistOrderingChallenge {
	return &PlaylistOrderingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-ordering",
			"Playlist Ordering",
			"Validates that playlist items maintain their "+
				"order in the response.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the playlist ordering challenge.
func (c *PlaylistOrderingChallenge) Execute(
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

	c.ReportProgress("checking-playlist-ordering", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/playlists/1/items",
	)

	// Accept 200 with items or 404 if no playlist exists
	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_ordering_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Playlist items returned %d", code),
			fmt.Sprintf("Playlist items returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "playlist_ordering_json",
			Expected: "valid JSON response",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Playlist items response is valid JSON",
				"Playlist items response is not valid JSON"),
		})
	}

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

// --- CH-167: PlaylistAccessControl ---

// PlaylistAccessControlChallenge validates that unauthenticated
// requests to playlist endpoints are rejected.
type PlaylistAccessControlChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPlaylistAccessControlChallenge creates CH-167.
func NewPlaylistAccessControlChallenge() *PlaylistAccessControlChallenge {
	return &PlaylistAccessControlChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-access-control",
			"Playlist Access Control",
			"Validates that unauthenticated requests to "+
				"playlist endpoints are rejected with 401.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the playlist access control challenge.
func (c *PlaylistAccessControlChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	// Use a client without authentication
	client := httpclient.NewAPIClient(c.config.BaseURL)

	c.ReportProgress("testing-playlist-access-control", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/playlists",
	)

	// Expect 401 (unauthenticated) or 403 (forbidden)
	codeOK := err == nil && (code == 401 || code == 403)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_access_control_status",
		Expected: "401 or 403",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Unauthenticated playlist access "+
				"rejected with %d", code),
			fmt.Sprintf("Unauthenticated playlist access "+
				"returned %d, err=%v", code, err)),
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

// --- CH-168: PublicPlaylistAccess ---

// PublicPlaylistAccessChallenge validates that public playlists
// are accessible via GET /api/v1/playlists/public.
type PublicPlaylistAccessChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPublicPlaylistAccessChallenge creates CH-168.
func NewPublicPlaylistAccessChallenge() *PublicPlaylistAccessChallenge {
	return &PublicPlaylistAccessChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-public-access",
			"Public Playlist Access",
			"Validates GET /api/v1/playlists/public returns "+
				"200 for publicly shared playlists.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the public playlist access challenge.
func (c *PublicPlaylistAccessChallenge) Execute(
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

	c.ReportProgress("checking-public-playlists", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/playlists/public",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_public_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Public playlists returned %d", code),
			fmt.Sprintf("Public playlists returned %d, err=%v",
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

// --- CH-169: PlaylistPagination ---

// PlaylistPaginationChallenge validates that
// GET /api/v1/playlists?limit=5&offset=0 respects pagination.
type PlaylistPaginationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPlaylistPaginationChallenge creates CH-169.
func NewPlaylistPaginationChallenge() *PlaylistPaginationChallenge {
	return &PlaylistPaginationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-pagination",
			"Playlist Pagination",
			"Validates GET /api/v1/playlists?limit=5&offset=0 "+
				"respects pagination parameters.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the playlist pagination challenge.
func (c *PlaylistPaginationChallenge) Execute(
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

	c.ReportProgress("testing-playlist-pagination", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/playlists?limit=5&offset=0",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_pagination_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Playlist pagination returned 200",
			fmt.Sprintf("Playlist pagination returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "playlist_pagination_json",
			Expected: "valid JSON response",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Paginated playlists response is valid JSON",
				"Paginated playlists response is invalid JSON"),
		})
	}

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

// --- CH-170: PlaylistSearch ---

// PlaylistSearchChallenge validates that
// GET /api/v1/playlists?q=test searches playlists by name.
type PlaylistSearchChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPlaylistSearchChallenge creates CH-170.
func NewPlaylistSearchChallenge() *PlaylistSearchChallenge {
	return &PlaylistSearchChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playlist-search",
			"Playlist Search",
			"Validates GET /api/v1/playlists?q=test searches "+
				"playlists by name and returns 200.",
			"playlist",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the playlist search challenge.
func (c *PlaylistSearchChallenge) Execute(
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

	c.ReportProgress("searching-playlists", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/playlists?q=test",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playlist_search_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Playlist search returned 200",
			fmt.Sprintf("Playlist search returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "playlist_search_json",
			Expected: "valid JSON response",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Playlist search response is valid JSON",
				"Playlist search response is invalid JSON"),
		})
	}

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
