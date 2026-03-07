package challenges

import (
	"testing"

	"digital.vasic.challenges/pkg/challenge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- CH-021 to CH-025: Module Integration Challenges ---

func TestNewCollectionsAPIChallenge_Metadata(t *testing.T) {
	ch := NewCollectionsAPIChallenge()

	assert.Equal(t, challenge.ID("collections-api"), ch.ID())
	assert.Equal(t, "Collections API", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewEntityUserMetadataChallenge_Metadata(t *testing.T) {
	ch := NewEntityUserMetadataChallenge()

	assert.Equal(t, challenge.ID("entity-user-metadata"), ch.ID())
	assert.Equal(t, "Entity User Metadata", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

func TestNewEntitySearchChallenge_Metadata(t *testing.T) {
	ch := NewEntitySearchChallenge()

	assert.Equal(t, challenge.ID("entity-search"), ch.ID())
	assert.Equal(t, "Entity Search", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

func TestNewStorageRootsAPIChallenge_Metadata(t *testing.T) {
	ch := NewStorageRootsAPIChallenge()

	assert.Equal(t, challenge.ID("storage-roots-api"), ch.ID())
	assert.Equal(t, "Storage Roots API", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewAuthTokenRefreshChallenge_Metadata(t *testing.T) {
	ch := NewAuthTokenRefreshChallenge()

	assert.Equal(t, challenge.ID("auth-token-refresh"), ch.ID())
	assert.Equal(t, "Auth Token Refresh", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

// --- CH-026 to CH-035: Extended Validation Challenges ---

func TestNewStressTestChallenge_Metadata(t *testing.T) {
	ch := NewStressTestChallenge()

	assert.Equal(t, challenge.ID("stress-test"), ch.ID())
	assert.Equal(t, "API Stress Test", ch.Name())
	assert.Equal(t, "stress", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewRateLimitingChallenge_Metadata(t *testing.T) {
	ch := NewRateLimitingChallenge()

	assert.Equal(t, challenge.ID("rate-limiting"), ch.ID())
	assert.Equal(t, "Rate Limiting", ch.Name())
	assert.Equal(t, "security", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewFavoritesWorkflowChallenge_Metadata(t *testing.T) {
	ch := NewFavoritesWorkflowChallenge()

	assert.Equal(t, challenge.ID("favorites-workflow"), ch.ID())
	assert.Equal(t, "Favorites Workflow", ch.Name())
	assert.Equal(t, "workflow", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

func TestNewCollectionManagementChallenge_Metadata(t *testing.T) {
	ch := NewCollectionManagementChallenge()

	assert.Equal(t, challenge.ID("collection-management"), ch.ID())
	assert.Equal(t, "Collection Management", ch.Name())
	assert.Equal(t, "workflow", ch.Category())
	deps := ch.Dependencies()
	require.Len(t, deps, 2)
	depSet := map[challenge.ID]bool{}
	for _, d := range deps {
		depSet[d] = true
	}
	assert.True(t, depSet["collections-api"])
	assert.True(t, depSet["entity-aggregation"])
}

func TestNewMediaPlaybackChallenge_Metadata(t *testing.T) {
	ch := NewMediaPlaybackChallenge()

	assert.Equal(t, challenge.ID("media-playback"), ch.ID())
	assert.Equal(t, "Media Playback", ch.Name())
	assert.Equal(t, "playback", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

func TestNewSearchFilterChallenge_Metadata(t *testing.T) {
	ch := NewSearchFilterChallenge()

	assert.Equal(t, challenge.ID("search-filter"), ch.ID())
	assert.Equal(t, "Search & Filter", ch.Name())
	assert.Equal(t, "search", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-search"), ch.Dependencies()[0])
}

func TestNewCoverArtChallenge_Metadata(t *testing.T) {
	ch := NewCoverArtChallenge()

	assert.Equal(t, challenge.ID("cover-art"), ch.ID())
	assert.Equal(t, "Cover Art", ch.Name())
	assert.Equal(t, "media", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

func TestNewWebSocketEventsChallenge_Metadata(t *testing.T) {
	ch := NewWebSocketEventsChallenge()

	assert.Equal(t, challenge.ID("websocket-events"), ch.ID())
	assert.Equal(t, "WebSocket Events", ch.Name())
	assert.Equal(t, "realtime", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewSecurityChallenge_Metadata(t *testing.T) {
	ch := NewSecurityChallenge()

	assert.Equal(t, challenge.ID("security"), ch.ID())
	assert.Equal(t, "Security", ch.Name())
	assert.Equal(t, "security", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewConfigWizardChallenge_Metadata(t *testing.T) {
	ch := NewConfigWizardChallenge()

	assert.Equal(t, challenge.ID("config-wizard"), ch.ID())
	assert.Equal(t, "Configuration Wizard", ch.Name())
	assert.Equal(t, "configuration", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

// --- CH-021 to CH-035: ID Uniqueness ---

func TestExtendedChallenges_IDUniqueness(t *testing.T) {
	challenges := []challenge.Challenge{
		NewCollectionsAPIChallenge(),
		NewEntityUserMetadataChallenge(),
		NewEntitySearchChallenge(),
		NewStorageRootsAPIChallenge(),
		NewAuthTokenRefreshChallenge(),
		NewStressTestChallenge(),
		NewRateLimitingChallenge(),
		NewFavoritesWorkflowChallenge(),
		NewCollectionManagementChallenge(),
		NewMediaPlaybackChallenge(),
		NewSearchFilterChallenge(),
		NewCoverArtChallenge(),
		NewWebSocketEventsChallenge(),
		NewSecurityChallenge(),
		NewConfigWizardChallenge(),
	}

	seen := map[challenge.ID]bool{}
	for _, ch := range challenges {
		if seen[ch.ID()] {
			t.Errorf("duplicate challenge ID: %s", ch.ID())
		}
		seen[ch.ID()] = true
	}
	assert.Len(t, seen, 15)
}

// --- CH-021 to CH-035: Category Distribution ---

func TestExtendedChallenges_CategoryDistribution(t *testing.T) {
	challenges := []challenge.Challenge{
		NewCollectionsAPIChallenge(),
		NewEntityUserMetadataChallenge(),
		NewEntitySearchChallenge(),
		NewStorageRootsAPIChallenge(),
		NewAuthTokenRefreshChallenge(),
		NewStressTestChallenge(),
		NewRateLimitingChallenge(),
		NewFavoritesWorkflowChallenge(),
		NewCollectionManagementChallenge(),
		NewMediaPlaybackChallenge(),
		NewSearchFilterChallenge(),
		NewCoverArtChallenge(),
		NewWebSocketEventsChallenge(),
		NewSecurityChallenge(),
		NewConfigWizardChallenge(),
	}

	categories := map[string]int{}
	for _, ch := range challenges {
		categories[ch.Category()]++
	}

	// Every challenge must have a non-empty category
	for _, ch := range challenges {
		assert.NotEmpty(t, ch.Category(), "challenge %s has empty category", ch.ID())
	}

	// Verify we have multiple categories (not all lumped together)
	assert.GreaterOrEqual(t, len(categories), 5, "expected at least 5 distinct categories")
}
