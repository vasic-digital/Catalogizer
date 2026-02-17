// Package challenges provides challenge registration for
// the Catalogizer system. Import this package to register
// all built-in challenges with the provided service.
package challenges

import (
	"catalogizer/services"
)

// RegisterAll registers all built-in Catalogizer challenges
// with the given challenge service.
func RegisterAll(svc *services.ChallengeService) error {
	// Challenges will be registered here as they are
	// implemented. Each challenge is a Go struct embedding
	// challenge.BaseChallenge with a custom Execute method.
	//
	// Example:
	//   svc.Register(NewAPIHealthChallenge())
	//   svc.Register(NewDatabaseChallenge())
	//
	return nil
}
