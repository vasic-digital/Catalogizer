// Package middleware provides infrastructure-level HTTP middleware for the Catalogizer API.
//
// It includes authentication verification middleware that validates JWT tokens and enforces
// role-based access, as well as metrics collection middleware that records request duration
// and status codes for Prometheus. These middleware complement the domain middleware in the
// top-level middleware package by handling infrastructure and security concerns.
package middleware
