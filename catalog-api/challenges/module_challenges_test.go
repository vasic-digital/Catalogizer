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

func TestModuleDocChallenge_Execute_ValidModule(t *testing.T) {
	// Create a fake module directory with required files.
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "pkg", "foo"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "ARCHITECTURE.md"), []byte("# Arch"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg", "foo", "foo_test.go"), []byte("package foo"), 0644))

	ch := newModuleDocChallenge("test-001", "Test Module", dir, "TestModule")
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusPassed, result.Status)
	assert.Len(t, result.Assertions, 4)
	for _, a := range result.Assertions {
		assert.True(t, a.Passed, "assertion %s should pass", a.Target)
	}
}

func TestModuleDocChallenge_Execute_MissingDocs(t *testing.T) {
	dir := t.TempDir()
	// Only create go.mod, no docs.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644))

	ch := newModuleDocChallenge("test-002", "Incomplete Module", dir, "TestModule")
	result, err := ch.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, challenge.StatusFailed, result.Status)
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(f, []byte("hello"), 0644))

	assert.True(t, fileExists(f))
	assert.False(t, fileExists(filepath.Join(dir, "nonexistent")))
	assert.False(t, fileExists(dir)) // directory, not file
}

func TestCountFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a_test.go"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b_test.go"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "c.go"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "d_test.go"), []byte(""), 0644))

	assert.Equal(t, 3, countFiles(dir, "*_test.go"))
	assert.Equal(t, 1, countFiles(dir, "c.go"))
}
