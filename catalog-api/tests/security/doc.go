// Package security contains security validation tests for the Catalogizer API.
//
// These tests verify authentication enforcement, authorization boundaries, input validation,
// and protection against common attack vectors such as SQL injection, path traversal, and
// cross-site scripting. They complement the automated security scanning tools (govulncheck,
// Semgrep, Snyk) by testing application-level security behavior.
package security
