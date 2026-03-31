// Package tests provides test infrastructure and helpers for the Catalogizer API test suite.
//
// It includes the test_helper.go file that provides SQLite-based test database setup via
// database.WrapDB(), enabling fast in-memory database testing without requiring an external
// database server. This package establishes common test patterns and shared setup routines
// used across all unit and integration tests in the project.
package tests
