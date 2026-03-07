package challenges

import (
	"testing"

	"digital.vasic.challenges/pkg/challenge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- CH-036 to CH-040: Security Validation Challenges ---

func TestNewAuthRequiredChallenge_Metadata(t *testing.T) {
	ch := NewAuthRequiredChallenge()

	assert.Equal(t, challenge.ID("auth-required"), ch.ID())
	assert.Equal(t, "Auth Required on Protected Endpoints", ch.Name())
	assert.Equal(t, "security", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewJWTExpirationChallenge_Metadata(t *testing.T) {
	ch := NewJWTExpirationChallenge()

	assert.Equal(t, challenge.ID("jwt-expiration"), ch.ID())
	assert.Equal(t, "JWT Token Expiration", ch.Name())
	assert.Equal(t, "security", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewRateLimitAuthChallenge_Metadata(t *testing.T) {
	ch := NewRateLimitAuthChallenge()

	assert.Equal(t, challenge.ID("rate-limit-auth"), ch.ID())
	assert.Equal(t, "Rate Limiting on Auth Endpoints", ch.Name())
	assert.Equal(t, "security", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewCORSHeadersChallenge_Metadata(t *testing.T) {
	ch := NewCORSHeadersChallenge()

	assert.Equal(t, challenge.ID("cors-headers"), ch.ID())
	assert.Equal(t, "CORS Headers", ch.Name())
	assert.Equal(t, "security", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewNoSensitiveErrorsChallenge_Metadata(t *testing.T) {
	ch := NewNoSensitiveErrorsChallenge()

	assert.Equal(t, challenge.ID("no-sensitive-errors"), ch.ID())
	assert.Equal(t, "No Sensitive Data in Errors", ch.Name())
	assert.Equal(t, "security", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

// --- CH-041 to CH-044: Performance Regression Challenges ---

func TestNewHealthLatencyChallenge_Metadata(t *testing.T) {
	ch := NewHealthLatencyChallenge()

	assert.Equal(t, challenge.ID("health-latency"), ch.ID())
	assert.Equal(t, "Health Endpoint Latency", ch.Name())
	assert.Equal(t, "performance", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewFileListingLatencyChallenge_Metadata(t *testing.T) {
	ch := NewFileListingLatencyChallenge()

	assert.Equal(t, challenge.ID("file-listing-latency"), ch.ID())
	assert.Equal(t, "File Listing Latency Under Load", ch.Name())
	assert.Equal(t, "performance", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewEntitySearchLatencyChallenge_Metadata(t *testing.T) {
	ch := NewEntitySearchLatencyChallenge()

	assert.Equal(t, challenge.ID("entity-search-latency"), ch.ID())
	assert.Equal(t, "Entity Search Latency", ch.Name())
	assert.Equal(t, "performance", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewWebSocketLatencyChallenge_Metadata(t *testing.T) {
	ch := NewWebSocketLatencyChallenge()

	assert.Equal(t, challenge.ID("websocket-latency"), ch.ID())
	assert.Equal(t, "WebSocket Broadcast Latency", ch.Name())
	assert.Equal(t, "performance", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

// --- CH-045 to CH-047: Documentation Completeness Challenges ---

func TestNewAPIDocsChallenge_Metadata(t *testing.T) {
	ch := NewAPIDocsChallenge()

	assert.Equal(t, challenge.ID("api-docs"), ch.ID())
	assert.Equal(t, "API Documentation Completeness", ch.Name())
	assert.Equal(t, "documentation", ch.Category())
	assert.Empty(t, ch.Dependencies(), "doc challenges should have no dependencies")
}

func TestNewDatabaseDocsChallenge_Metadata(t *testing.T) {
	ch := NewDatabaseDocsChallenge()

	assert.Equal(t, challenge.ID("database-docs"), ch.ID())
	assert.Equal(t, "Database Documentation Completeness", ch.Name())
	assert.Equal(t, "documentation", ch.Category())
	assert.Empty(t, ch.Dependencies(), "doc challenges should have no dependencies")
}

func TestNewConfigDocsChallenge_Metadata(t *testing.T) {
	ch := NewConfigDocsChallenge()

	assert.Equal(t, challenge.ID("config-docs"), ch.ID())
	assert.Equal(t, "Configuration Documentation Completeness", ch.Name())
	assert.Equal(t, "documentation", ch.Category())
	assert.Empty(t, ch.Dependencies(), "doc challenges should have no dependencies")
}

// --- CH-048 to CH-050: System Resilience Challenges ---

func TestNewDBErrorRecoveryChallenge_Metadata(t *testing.T) {
	ch := NewDBErrorRecoveryChallenge()

	assert.Equal(t, challenge.ID("db-error-recovery"), ch.ID())
	assert.Equal(t, "Database Error Recovery", ch.Name())
	assert.Equal(t, "resilience", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewScannerRecoveryChallenge_Metadata(t *testing.T) {
	ch := NewScannerRecoveryChallenge()

	assert.Equal(t, challenge.ID("scanner-recovery"), ch.ID())
	assert.Equal(t, "Scanner Filesystem Recovery", ch.Name())
	assert.Equal(t, "resilience", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

func TestNewGracefulShutdownChallenge_Metadata(t *testing.T) {
	ch := NewGracefulShutdownChallenge()

	assert.Equal(t, challenge.ID("graceful-shutdown"), ch.ID())
	assert.Equal(t, "Graceful Shutdown Support", ch.Name())
	assert.Equal(t, "resilience", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("browsing-api-health"), ch.Dependencies()[0])
}

// --- Entity Challenges (CH-016 to CH-020) ---

func TestNewEntityAggregationChallenge_Metadata(t *testing.T) {
	ch := NewEntityAggregationChallenge()

	assert.Equal(t, challenge.ID("entity-aggregation"), ch.ID())
	assert.Equal(t, "Entity Aggregation", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("first-catalog-populate"), ch.Dependencies()[0])
}

func TestNewEntityBrowsingChallenge_Metadata(t *testing.T) {
	ch := NewEntityBrowsingChallenge()

	assert.Equal(t, challenge.ID("entity-browsing"), ch.ID())
	assert.Equal(t, "Entity Browsing", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

func TestNewEntityMetadataChallenge_Metadata(t *testing.T) {
	ch := NewEntityMetadataChallenge()

	assert.Equal(t, challenge.ID("entity-metadata"), ch.ID())
	assert.Equal(t, "Entity Metadata Enrichment", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

func TestNewEntityDuplicatesChallenge_Metadata(t *testing.T) {
	ch := NewEntityDuplicatesChallenge()

	assert.Equal(t, challenge.ID("entity-duplicates"), ch.ID())
	assert.Equal(t, "Entity Duplicate Detection", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

func TestNewEntityHierarchyChallenge_Metadata(t *testing.T) {
	ch := NewEntityHierarchyChallenge()

	assert.Equal(t, challenge.ID("entity-hierarchy"), ch.ID())
	assert.Equal(t, "Entity Hierarchical Navigation", ch.Name())
	assert.Equal(t, "e2e", ch.Category())
	require.Len(t, ch.Dependencies(), 1)
	assert.Equal(t, challenge.ID("entity-aggregation"), ch.Dependencies()[0])
}

// --- Cross-cutting: Security Challenges ID Uniqueness ---

func TestSecurityChallenges_IDUniqueness(t *testing.T) {
	challenges := []challenge.Challenge{
		NewAuthRequiredChallenge(),
		NewJWTExpirationChallenge(),
		NewRateLimitAuthChallenge(),
		NewCORSHeadersChallenge(),
		NewNoSensitiveErrorsChallenge(),
		NewRateLimitingChallenge(),
		NewSecurityChallenge(),
	}

	seen := map[challenge.ID]bool{}
	for _, ch := range challenges {
		assert.Equal(t, "security", ch.Category(), "challenge %s should be in security category", ch.ID())
		if seen[ch.ID()] {
			t.Errorf("duplicate security challenge ID: %s", ch.ID())
		}
		seen[ch.ID()] = true
	}
	assert.Len(t, seen, 7)
}

// --- Cross-cutting: Performance Challenges ID Uniqueness ---

func TestPerformanceChallenges_IDUniqueness(t *testing.T) {
	challenges := []challenge.Challenge{
		NewHealthLatencyChallenge(),
		NewFileListingLatencyChallenge(),
		NewEntitySearchLatencyChallenge(),
		NewWebSocketLatencyChallenge(),
	}

	seen := map[challenge.ID]bool{}
	for _, ch := range challenges {
		assert.Equal(t, "performance", ch.Category(), "challenge %s should be in performance category", ch.ID())
		if seen[ch.ID()] {
			t.Errorf("duplicate performance challenge ID: %s", ch.ID())
		}
		seen[ch.ID()] = true
	}
	assert.Len(t, seen, 4)
}

// --- Cross-cutting: Documentation Challenges No Dependencies ---

func TestDocumentationChallenges_NoDependencies(t *testing.T) {
	challenges := []challenge.Challenge{
		NewAPIDocsChallenge(),
		NewDatabaseDocsChallenge(),
		NewConfigDocsChallenge(),
	}

	for _, ch := range challenges {
		assert.Equal(t, "documentation", ch.Category(), "challenge %s should be in documentation category", ch.ID())
		assert.Empty(t, ch.Dependencies(), "documentation challenge %s should have no dependencies", ch.ID())
	}
}

// --- Cross-cutting: Resilience Challenges ---

func TestResilienceChallenges_IDUniqueness(t *testing.T) {
	challenges := []challenge.Challenge{
		NewDBErrorRecoveryChallenge(),
		NewScannerRecoveryChallenge(),
		NewGracefulShutdownChallenge(),
	}

	seen := map[challenge.ID]bool{}
	for _, ch := range challenges {
		assert.Equal(t, "resilience", ch.Category(), "challenge %s should be in resilience category", ch.ID())
		if seen[ch.ID()] {
			t.Errorf("duplicate resilience challenge ID: %s", ch.ID())
		}
		seen[ch.ID()] = true
	}
	assert.Len(t, seen, 3)
}

// --- Cross-cutting: Entity Dependency Chain ---

func TestEntityChallenges_DependencyChain(t *testing.T) {
	// All entity sub-challenges depend on entity-aggregation
	subChallenges := []challenge.Challenge{
		NewEntityBrowsingChallenge(),
		NewEntityMetadataChallenge(),
		NewEntityDuplicatesChallenge(),
		NewEntityHierarchyChallenge(),
		NewEntitySearchChallenge(),
		NewEntityUserMetadataChallenge(),
	}

	for _, ch := range subChallenges {
		deps := ch.Dependencies()
		found := false
		for _, d := range deps {
			if d == "entity-aggregation" {
				found = true
				break
			}
		}
		assert.True(t, found, "challenge %s should depend on entity-aggregation", ch.ID())
	}

	// entity-aggregation itself depends on first-catalog-populate
	agg := NewEntityAggregationChallenge()
	require.Len(t, agg.Dependencies(), 1)
	assert.Equal(t, challenge.ID("first-catalog-populate"), agg.Dependencies()[0])
}

// --- Cross-cutting: All 50 Challenge IDs Unique ---

func TestAllOriginalChallenges_GlobalIDUniqueness(t *testing.T) {
	ep := &Endpoint{Host: "localhost", Port: 445}
	dir := Directory{Path: "/media", ContentType: "movie"}

	all := []challenge.Challenge{
		// CH-001 to CH-007: First catalog
		NewSMBConnectivityChallenge(ep),
		NewDirectoryDiscoveryChallenge(ep),
		NewMusicScanChallenge(ep, dir),
		NewSeriesScanChallenge(ep, dir),
		NewMoviesScanChallenge(ep, dir),
		NewSoftwareScanChallenge(ep, dir),
		NewComicsScanChallenge(ep, dir),
		// CH-008: Populate
		NewFirstCatalogPopulateChallenge(),
		// CH-009 to CH-011: Browsing
		NewBrowsingAPIHealthChallenge(),
		NewBrowsingAPICatalogChallenge(),
		NewBrowsingWebAppChallenge(),
		// CH-012 to CH-013: Assets
		NewAssetServingChallenge(),
		NewAssetLazyLoadingChallenge(),
		// CH-014 to CH-015: Database
		NewDatabaseConnectivityChallenge(),
		NewDatabaseSchemaValidationChallenge(),
		// CH-016 to CH-020: Entity
		NewEntityAggregationChallenge(),
		NewEntityBrowsingChallenge(),
		NewEntityMetadataChallenge(),
		NewEntityDuplicatesChallenge(),
		NewEntityHierarchyChallenge(),
		// CH-021 to CH-025: Module Integration
		NewCollectionsAPIChallenge(),
		NewEntityUserMetadataChallenge(),
		NewEntitySearchChallenge(),
		NewStorageRootsAPIChallenge(),
		NewAuthTokenRefreshChallenge(),
		// CH-026 to CH-035: Extended Validation
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
		// CH-036 to CH-040: Security
		NewAuthRequiredChallenge(),
		NewJWTExpirationChallenge(),
		NewRateLimitAuthChallenge(),
		NewCORSHeadersChallenge(),
		NewNoSensitiveErrorsChallenge(),
		// CH-041 to CH-044: Performance
		NewHealthLatencyChallenge(),
		NewFileListingLatencyChallenge(),
		NewEntitySearchLatencyChallenge(),
		NewWebSocketLatencyChallenge(),
		// CH-045 to CH-047: Documentation
		NewAPIDocsChallenge(),
		NewDatabaseDocsChallenge(),
		NewConfigDocsChallenge(),
		// CH-048 to CH-050: Resilience
		NewDBErrorRecoveryChallenge(),
		NewScannerRecoveryChallenge(),
		NewGracefulShutdownChallenge(),
	}

	seen := map[challenge.ID]bool{}
	for _, ch := range all {
		if seen[ch.ID()] {
			t.Errorf("duplicate challenge ID: %s", ch.ID())
		}
		seen[ch.ID()] = true

		// Every challenge must have non-empty ID, Name, Category
		assert.NotEmpty(t, ch.ID(), "challenge has empty ID")
		assert.NotEmpty(t, ch.Name(), "challenge %s has empty name", ch.ID())
		assert.NotEmpty(t, ch.Category(), "challenge %s has empty category", ch.ID())
	}

	assert.Equal(t, 50, len(seen), "expected exactly 50 unique original challenge IDs")
}
