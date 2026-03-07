package challenges

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"digital.vasic.challenges/pkg/challenge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- CH-045: API Docs Execute Tests ---

func TestAPIDocsChallenge_Execute_WithDocsDir(t *testing.T) {
	// Create a temp dir with docs structure the challenge looks for.
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "api"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "architecture"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "deployment"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "guides"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "security"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "api", "endpoints.md"),
		[]byte("# API Endpoints\n"),
		0644,
	))

	// ch045 probes relative paths; chdir to make it find our temp dir.
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	ch := NewAPIDocsChallenge()
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusPassed, result.Status)
	assert.GreaterOrEqual(t, len(result.Assertions), 2) // docs_directory + doc_files_exist + topics
}

func TestAPIDocsChallenge_Execute_NoDocs(t *testing.T) {
	dir := t.TempDir()

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	ch := NewAPIDocsChallenge()
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusFailed, result.Status)
}

func TestAPIDocsChallenge_Execute_EmptyDocsDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0755))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	ch := NewAPIDocsChallenge()
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	// Empty docs dir exists but has no files → fails on doc_files_exist
	assert.Equal(t, challenge.StatusFailed, result.Status)
}

// --- CH-046: Database Docs Execute Tests ---

func TestDatabaseDocsChallenge_Execute_WithMigrations(t *testing.T) {
	dir := t.TempDir()

	// Create migration dir with files
	migDir := filepath.Join(dir, "database", "migrations")
	require.NoError(t, os.MkdirAll(migDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(migDir, "001_init.go"),
		[]byte("package migrations\n"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(migDir, "002_media.go"),
		[]byte("package migrations\n"),
		0644,
	))

	// Create database package files
	dbDir := filepath.Join(dir, "database")
	require.NoError(t, os.WriteFile(
		filepath.Join(dbDir, "connection.go"),
		[]byte("package database\n"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dbDir, "dialect.go"),
		[]byte("package database\n"),
		0644,
	))

	// Create docs with table references
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "schema.md"),
		[]byte("# Schema\n\nmedia_items table stores entities.\nmedia_files links files.\nmedia_types are seeded.\nstorage_roots for mounts.\nusers table for auth.\n"),
		0644,
	))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	ch := NewDatabaseDocsChallenge()
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusPassed, result.Status)
	assert.GreaterOrEqual(t, len(result.Assertions), 3)
}

func TestDatabaseDocsChallenge_Execute_NoMigrations(t *testing.T) {
	dir := t.TempDir()

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	ch := NewDatabaseDocsChallenge()
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusFailed, result.Status)
}

// --- CH-047: Config Docs Execute Tests ---

func TestConfigDocsChallenge_Execute_WithDocs(t *testing.T) {
	dir := t.TempDir()

	// Create docs with config references
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "config.md"),
		[]byte("# Configuration\n\nSet PORT=8080\nGIN_MODE=release\nDB_TYPE=sqlite or postgres\nJWT_SECRET=your-secret\nADMIN_PASSWORD=admin\nTMDB_API_KEY=optional\n"),
		0644,
	))

	// Create .env.example
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".env.example"),
		[]byte("PORT=8080\nJWT_SECRET=changeme\n"),
		0644,
	))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	ch := NewConfigDocsChallenge()
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusPassed, result.Status)
	assert.GreaterOrEqual(t, len(result.Assertions), 3)
}

func TestConfigDocsChallenge_Execute_NoDocs(t *testing.T) {
	dir := t.TempDir()

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	ch := NewConfigDocsChallenge()
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusFailed, result.Status)
}

func TestConfigDocsChallenge_Execute_PartialDocs(t *testing.T) {
	dir := t.TempDir()

	// Only document 2 of 6 config options (below threshold of 4)
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "readme.md"),
		[]byte("# Setup\n\nSet PORT and GIN_MODE.\n"),
		0644,
	))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	ch := NewConfigDocsChallenge()
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusFailed, result.Status)
}

// --- Module Challenges: Additional Execute Tests ---

func TestModuleDocChallenge_Execute_PartialDocs(t *testing.T) {
	// Module with docs/ARCHITECTURE.md but missing CLAUDE.md and tests
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "docs", "ARCHITECTURE.md"),
		[]byte("# Arch\n"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module test\n"),
		0644,
	))

	ch := newModuleDocChallenge("test-partial", "Partial Module", dir, "TestModule")
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusFailed, result.Status)

	// Should have 4 assertions (arch, claude, tests, go.mod)
	assert.Len(t, result.Assertions, 4)

	// arch and go.mod should pass, claude and tests should fail
	passCount := 0
	for _, a := range result.Assertions {
		if a.Passed {
			passCount++
		}
	}
	assert.Equal(t, 2, passCount, "expected exactly 2 passing assertions (arch + go.mod)")
}

func TestModuleDocChallenge_Execute_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	ch := newModuleDocChallenge("test-empty", "Empty Module", dir, "TestModule")
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusFailed, result.Status)
	assert.Len(t, result.Assertions, 4)

	// All 4 should fail
	for _, a := range result.Assertions {
		assert.False(t, a.Passed, "assertion %s should fail for empty dir", a.Target)
	}
}

func TestModuleDocChallenge_Execute_Metrics(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "pkg", "foo"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "ARCHITECTURE.md"), []byte("# Arch"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg", "foo", "foo_test.go"), []byte("package foo"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg", "foo", "bar_test.go"), []byte("package foo"), 0644))

	ch := newModuleDocChallenge("test-metrics", "Metrics Module", dir, "TestModule")
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusPassed, result.Status)

	// Verify metrics include test file count
	require.Contains(t, result.Metrics, "test_files")
	assert.Equal(t, float64(2), result.Metrics["test_files"].Value)
}
