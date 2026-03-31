// Package auth provides JWT-based authentication with role-based access control for the
// Catalogizer API.
//
// It handles token generation, validation, and refresh using configurable secrets and
// expiration policies. The package supports role-based authorization with predefined roles
// (admin, user, viewer) and integrates with the middleware layer to protect API endpoints
// based on the authenticated user's permissions.
package auth
