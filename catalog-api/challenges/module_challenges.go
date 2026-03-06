package challenges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// moduleDocChallenge verifies that a decoupled Go module has proper
// documentation (ARCHITECTURE.md, CLAUDE.md, tests).
type moduleDocChallenge struct {
	challenge.BaseChallenge
	modulePath string
	moduleName string
}

func newModuleDocChallenge(id, name, modulePath, moduleName string) *moduleDocChallenge {
	return &moduleDocChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			challenge.ID(id),
			name,
			fmt.Sprintf("Verifies that the %s module has architecture docs, CLAUDE.md, and tests.", moduleName),
			"module-verification",
			nil,
		),
		modulePath: modulePath,
		moduleName: moduleName,
	}
}

func (c *moduleDocChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}

	// Check ARCHITECTURE.md exists
	archPath := filepath.Join(c.modulePath, "docs", "ARCHITECTURE.md")
	hasArch := fileExists(archPath)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "architecture_doc",
		Expected: "docs/ARCHITECTURE.md exists",
		Actual:   fmt.Sprintf("exists=%v", hasArch),
		Passed:   hasArch,
		Message:  challenge.Ternary(hasArch, "Architecture doc found", "Architecture doc missing"),
	})

	// Check CLAUDE.md exists
	claudePath := filepath.Join(c.modulePath, "CLAUDE.md")
	hasClaude := fileExists(claudePath)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "claude_md",
		Expected: "CLAUDE.md exists",
		Actual:   fmt.Sprintf("exists=%v", hasClaude),
		Passed:   hasClaude,
		Message:  challenge.Ternary(hasClaude, "CLAUDE.md found", "CLAUDE.md missing"),
	})

	// Check test files exist
	testCount := countFiles(c.modulePath, "*_test.go")
	hasTests := testCount > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "test_files",
		Expected: "at least 1 test file",
		Actual:   fmt.Sprintf("%d test files", testCount),
		Passed:   hasTests,
		Message:  challenge.Ternary(hasTests, fmt.Sprintf("%d test files found", testCount), "No test files found"),
	})

	// Check go.mod exists
	gomodPath := filepath.Join(c.modulePath, "go.mod")
	hasGomod := fileExists(gomodPath)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "go_mod",
		Expected: "go.mod exists",
		Actual:   fmt.Sprintf("exists=%v", hasGomod),
		Passed:   hasGomod,
		Message:  challenge.Ternary(hasGomod, "go.mod found", "go.mod missing"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	metrics := map[string]challenge.MetricValue{
		"test_files": {Name: "test_files", Value: float64(testCount), Unit: "count"},
	}

	return c.CreateResult(status, start, assertions, metrics, nil, ""), nil
}

// RegisterModuleChallenges registers challenges that verify each decoupled
// module has proper structure and documentation. These run without network
// access and validate the local filesystem layout.
func RegisterModuleChallenges(svc interface{ Register(challenge.Challenge) error }) {
	modules := []struct {
		id   string
		name string
		path string
	}{
		{"MOD-001", "Database Module Verification", "../Database"},
		{"MOD-002", "Concurrency Module Verification", "../Concurrency"},
		{"MOD-003", "Observability Module Verification", "../Observability"},
		{"MOD-004", "Security Module Verification", "../Security"},
		{"MOD-005", "Middleware Module Verification", "../Middleware"},
		{"MOD-006", "Media Module Verification", "../Media"},
		{"MOD-007", "Discovery Module Verification", "../Discovery"},
		{"MOD-008", "Streaming Module Verification", "../Streaming"},
		{"MOD-009", "Lazy Module Verification", "../Lazy"},
		{"MOD-010", "Memory Module Verification", "../Memory"},
		{"MOD-011", "Recovery Module Verification", "../Recovery"},
		{"MOD-012", "Storage Module Verification", "../Storage"},
		{"MOD-013", "Cache Module Verification", "../Cache"},
		{"MOD-014", "Watcher Module Verification", "../Watcher"},
		{"MOD-015", "RateLimiter Module Verification", "../RateLimiter"},
	}

	for _, m := range modules {
		_ = svc.Register(newModuleDocChallenge(m.id, m.name, m.path, m.name))
	}
}

// fileExists returns true if the path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// countFiles recursively counts files matching the given glob pattern.
func countFiles(root, pattern string) int {
	count := 0
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			count++
		}
		return nil
	})
	return count
}
