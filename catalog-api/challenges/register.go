// Package challenges provides challenge registration for
// the Catalogizer system. Import this package to register
// all built-in challenges with the provided service.
package challenges

import (
	"catalogizer/services"
	"fmt"
	"os"
)

// RegisterAll registers all built-in Catalogizer challenges
// with the given challenge service. If the endpoint config
// file is missing, registration is silently skipped (common
// in dev environments without NAS access).
func RegisterAll(svc *services.ChallengeService) error {
	cfg, err := LoadEndpointConfig(DefaultConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		// Log but don't fail - config might just be missing
		fmt.Printf("challenges: config not loaded: %v\n", err)
		return nil
	}

	for _, ep := range cfg.Endpoints {
		endpoint := ep // capture loop variable

		svc.Register(NewSMBConnectivityChallenge(&endpoint))
		svc.Register(NewDirectoryDiscoveryChallenge(&endpoint))

		for _, dir := range endpoint.Directories {
			d := dir // capture loop variable
			switch d.ContentType {
			case "music":
				svc.Register(NewMusicScanChallenge(&endpoint, d))
			case "tv_show":
				svc.Register(NewSeriesScanChallenge(&endpoint, d))
			case "movie":
				svc.Register(NewMoviesScanChallenge(&endpoint, d))
			case "software":
				svc.Register(NewSoftwareScanChallenge(&endpoint, d))
			case "comic":
				svc.Register(NewComicsScanChallenge(&endpoint, d))
			}
		}
	}

	// Populate challenge (triggers full scan pipeline via API)
	svc.Register(NewFirstCatalogPopulateChallenge())

	// Browsing challenges (validate the running API and web app)
	svc.Register(NewBrowsingAPIHealthChallenge())
	svc.Register(NewBrowsingAPICatalogChallenge())
	svc.Register(NewBrowsingWebAppChallenge())

	// Asset challenges (validate asset lazy loading pipeline)
	svc.Register(NewAssetServingChallenge())
	svc.Register(NewAssetLazyLoadingChallenge())

	// Database challenges (validate database connectivity and schema)
	svc.Register(NewDatabaseConnectivityChallenge())
	svc.Register(NewDatabaseSchemaValidationChallenge())

	// Entity challenges (validate entity system after scan)
	svc.Register(NewEntityAggregationChallenge())
	svc.Register(NewEntityBrowsingChallenge())
	svc.Register(NewEntityMetadataChallenge())
	svc.Register(NewEntityDuplicatesChallenge())
	svc.Register(NewEntityHierarchyChallenge())

	// Module integration challenges (CH-021 to CH-025)
	// Validate the API endpoints used by vasic-digital TypeScript modules
	svc.Register(NewCollectionsAPIChallenge())        // CH-021: @vasic-digital/collection-manager
	svc.Register(NewEntityUserMetadataChallenge())    // CH-022: @vasic-digital/media-browser
	svc.Register(NewEntitySearchChallenge())          // CH-023: @vasic-digital/media-browser
	svc.Register(NewStorageRootsAPIChallenge())       // CH-024: @vasic-digital/catalogizer-api-client
	svc.Register(NewAuthTokenRefreshChallenge())      // CH-025: @vasic-digital/catalogizer-api-client

	// Extended validation challenges (CH-026 to CH-035)
	svc.Register(NewStressTestChallenge())             // CH-026: API stress test
	svc.Register(NewRateLimitingChallenge())           // CH-027: Rate limiting
	svc.Register(NewFavoritesWorkflowChallenge())      // CH-028: Favorites workflow
	svc.Register(NewCollectionManagementChallenge())   // CH-029: Collection management
	svc.Register(NewMediaPlaybackChallenge())          // CH-030: Media playback
	svc.Register(NewSearchFilterChallenge())           // CH-031: Search & filter
	svc.Register(NewCoverArtChallenge())               // CH-032: Cover art
	svc.Register(NewWebSocketEventsChallenge())        // CH-033: WebSocket events
	svc.Register(NewSecurityChallenge())               // CH-034: Security
	svc.Register(NewConfigWizardChallenge())           // CH-035: Configuration wizard

	// Security validation challenges (CH-036 to CH-040)
	svc.Register(NewAuthRequiredChallenge())        // CH-036: Auth required on endpoints
	svc.Register(NewJWTExpirationChallenge())       // CH-037: JWT expiration enforced
	svc.Register(NewRateLimitAuthChallenge())       // CH-038: Rate limiting on auth
	svc.Register(NewCORSHeadersChallenge())         // CH-039: CORS headers
	svc.Register(NewNoSensitiveErrorsChallenge())   // CH-040: No sensitive data in errors

	// Performance regression challenges (CH-041 to CH-044)
	svc.Register(NewHealthLatencyChallenge())       // CH-041: Health < 10ms
	svc.Register(NewFileListingLatencyChallenge())  // CH-042: File listing < 200ms
	svc.Register(NewEntitySearchLatencyChallenge()) // CH-043: Entity search < 500ms
	svc.Register(NewWebSocketLatencyChallenge())    // CH-044: WebSocket < 50ms

	// Documentation completeness challenges (CH-045 to CH-047)
	svc.Register(NewAPIDocsChallenge())             // CH-045: API docs
	svc.Register(NewDatabaseDocsChallenge())        // CH-046: Database docs
	svc.Register(NewConfigDocsChallenge())          // CH-047: Config docs

	// System resilience challenges (CH-048 to CH-050)
	svc.Register(NewDBErrorRecoveryChallenge())     // CH-048: DB error recovery
	svc.Register(NewScannerRecoveryChallenge())     // CH-049: Scanner recovery
	svc.Register(NewGracefulShutdownChallenge())    // CH-050: Graceful shutdown

	// Extended API validation challenges (CH-051 to CH-060)
	svc.Register(NewInputValidationChallenge())     // CH-051: Input validation & sanitization
	svc.Register(NewPaginationChallenge())          // CH-052: Pagination consistency
	svc.Register(NewContentTypesChallenge())        // CH-053: Content-Type validation
	svc.Register(NewUserManagementChallenge())      // CH-054: User management API
	svc.Register(NewAnalyticsAPIChallenge())        // CH-055: Analytics & statistics API
	svc.Register(NewEntityCRUDChallenge())          // CH-056: Entity CRUD operations
	svc.Register(NewSyncAPIChallenge())             // CH-057: Synchronization API
	svc.Register(NewSubtitleAPIChallenge())         // CH-058: Subtitle management API
	svc.Register(NewRecommendationAPIChallenge())   // CH-059: Recommendation engine API
	svc.Register(NewLocalizationAPIChallenge())     // CH-060: Localization & i18n API

	// Search and browse challenges (CH-061 to CH-065)
	svc.Register(NewSearchAPIBasicQueryChallenge())           // CH-061: Search API basic query
	svc.Register(NewSearchAPIDuplicateDetectionChallenge())   // CH-062: Search API duplicate detection
	svc.Register(NewSearchAPIAdvancedFiltersChallenge())      // CH-063: Search API advanced filters
	svc.Register(NewBrowseAPIStorageRootsChallenge())         // CH-064: Browse API storage roots
	svc.Register(NewBrowseAPIDirectoryListingChallenge())     // CH-065: Browse API directory listing

	// Sync and security challenges (CH-066 to CH-070)
	svc.Register(NewSyncAPIEndpointCreationChallenge())       // CH-066: Sync API endpoint creation
	svc.Register(NewSyncAPICloudProvidersChallenge())         // CH-067: Sync API cloud providers
	svc.Register(NewSyncAPIUserEndpointsChallenge())          // CH-068: Sync API user endpoints
	svc.Register(NewSecurityHeadersAllChallenge())            // CH-069: Security headers present
	svc.Register(NewCORSRejectsUnauthorizedChallenge())       // CH-070: CORS rejects unauthorized origins

	// Security validation challenges (CH-071 to CH-075)
	svc.Register(NewInputValidationRejectsInjectionChallenge()) // CH-071: Input validation rejects injection
	svc.Register(NewRateLimitAuthEndpointsChallenge())        // CH-072: Rate limit auth endpoints
	svc.Register(NewJWTTokenLifecycleChallenge())             // CH-073: JWT token lifecycle
	svc.Register(NewFileUploadMagicBytesChallenge())          // CH-074: File upload magic bytes
	svc.Register(NewConversionRejectsPathTraversalChallenge()) // CH-075: Path traversal rejection

	// Performance challenges (CH-076 to CH-078)
	svc.Register(NewAPIResponseLatencyChallenge())            // CH-076: API response latency
	svc.Register(NewAPIConcurrentRequestsChallenge())         // CH-077: API concurrent requests
	svc.Register(NewGracefulDegradationChallenge())           // CH-078: Graceful degradation

	// Resilience challenges (CH-079 to CH-080)
	svc.Register(NewMemoryStableDuringLoadChallenge())        // CH-079: Memory stable during load
	svc.Register(NewDBPoolRecoveryChallenge())                // CH-080: DB pool recovery

	// WebSocket and runtime challenges (CH-081 to CH-083)
	svc.Register(NewWebSocketReconnectionChallenge())         // CH-081: WebSocket reconnection
	svc.Register(NewLazyInitOnFirstRequestChallenge())        // CH-082: Lazy init on first request
	svc.Register(NewSemaphorePreventsOverloadChallenge())     // CH-083: Semaphore prevents overload

	// Observability challenges (CH-084 to CH-088)
	svc.Register(NewPrometheusMetricsEndpointChallenge())     // CH-084: Prometheus metrics endpoint
	svc.Register(NewHTTPRequestMetricsIncrementChallenge())   // CH-085: HTTP request metrics increment
	svc.Register(NewRuntimeMetricsCurrentChallenge())         // CH-086: Runtime metrics current
	svc.Register(NewDBQueryDurationTrackedChallenge())        // CH-087: DB query duration tracked
	svc.Register(NewGrafanaDashboardRendersChallenge())       // CH-088: Grafana dashboard config exists

	// User flow challenges (UF-*): exhaustive multi-platform
	// user flow automation across all 6 Catalogizer applications
	RegisterUserFlowAPIChallenges(svc)     // 49 API challenges
	RegisterUserFlowWebChallenges(svc)     // 59 web browser challenges
	RegisterUserFlowDesktopChallenges(svc) // 28 desktop + wizard challenges
	RegisterUserFlowMobileChallenges(svc)  // 38 Android + TV challenges

	// Module verification challenges (MOD-001 to MOD-015)
	// Verify each decoupled Go module has proper structure and docs
	RegisterModuleChallenges(svc)

	// Module functional verification challenges (MOD-016 to MOD-021)
	// Verify specific module capabilities (types, functions, patterns)
	RegisterModuleFuncChallenges(svc)

	return nil
}
