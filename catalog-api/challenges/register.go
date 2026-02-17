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

	return nil
}
