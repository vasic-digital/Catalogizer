// Package integration contains integration tests that validate end-to-end workflows in the
// Catalogizer API.
//
// These tests exercise complete request paths from HTTP handler through service and repository
// layers to the database, verifying that all components work together correctly. Integration
// tests use in-memory SQLite databases for speed while testing realistic multi-layer
// interactions including authentication, scanning, and media entity creation.
package integration
