// Package services provides domain-level business logic for the Catalogizer API.
//
// It includes services for file synchronization, format conversion, log management, and error
// reporting. Each service is constructed with dependency injection via NewService constructors
// and orchestrates operations between handlers and repositories. This package represents the
// middle tier of the Handler -> Service -> Repository -> Database architecture.
package services
