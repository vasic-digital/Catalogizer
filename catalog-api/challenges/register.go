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

	return nil
}
