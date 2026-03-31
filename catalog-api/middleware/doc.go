// Package middleware provides domain-level HTTP middleware for the Catalogizer API.
//
// It includes CORS configuration, structured request logging, rate limiting, and request ID
// generation. These middleware components are applied to the Gin router to handle cross-cutting
// concerns at the domain level, complementing the infrastructure middleware in the internal
// package.
package middleware
