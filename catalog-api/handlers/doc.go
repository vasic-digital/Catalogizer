// Package handlers provides domain-level HTTP request handlers for the Catalogizer API.
//
// These handlers implement the /api/v1 routes for catalog browsing, media entity management,
// search, file downloads, storage root operations, and collection management. Each handler
// follows the Gin framework conventions and delegates business logic to the services layer,
// forming the top tier of the Handler -> Service -> Repository -> Database architecture.
package handlers
