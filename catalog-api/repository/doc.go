// Package repository provides the data access layer for the Catalogizer API.
//
// It implements CRUD operations for files, media items, collections, storage roots, and
// statistics using the dual-dialect database abstraction. Repositories encapsulate all SQL
// queries and form the bottom tier of the Handler -> Service -> Repository -> Database
// architecture, ensuring that business logic in the services layer remains decoupled from
// database-specific concerns.
package repository
