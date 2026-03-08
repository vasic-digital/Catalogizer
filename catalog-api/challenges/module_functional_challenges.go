package challenges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// moduleFuncChallenge verifies that a module's Go source has specific
// functional characteristics (exported types, test coverage, etc.).
type moduleFuncChallenge struct {
	challenge.BaseChallenge
	modulePath    string
	requiredTypes []string
	requiredFiles []string
}

func newModuleFuncChallenge(
	id, name, desc, modulePath string,
	requiredTypes, requiredFiles []string,
) *moduleFuncChallenge {
	return &moduleFuncChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			challenge.ID(id),
			name,
			desc,
			"module-verification",
			nil,
		),
		modulePath:    modulePath,
		requiredTypes: requiredTypes,
		requiredFiles: requiredFiles,
	}
}

func (c *moduleFuncChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"module": c.modulePath}

	// Check required files exist
	c.ReportProgress("check-files", nil)
	for _, f := range c.requiredFiles {
		fpath := filepath.Join(c.modulePath, f)
		exists := fileExists(fpath)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   fmt.Sprintf("file_%s", filepath.Base(f)),
			Expected: fmt.Sprintf("%s exists", f),
			Actual:   fmt.Sprintf("exists=%v", exists),
			Passed:   exists,
			Message: challenge.Ternary(exists,
				fmt.Sprintf("File %s found", f),
				fmt.Sprintf("File %s missing", f)),
		})
	}

	// Check required types/functions appear in source
	c.ReportProgress("check-types", nil)
	for _, typeName := range c.requiredTypes {
		found := sourceContains(c.modulePath, typeName)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "contains",
			Target:   fmt.Sprintf("type_%s", typeName),
			Expected: fmt.Sprintf("source contains %s", typeName),
			Actual:   fmt.Sprintf("found=%v", found),
			Passed:   found,
			Message: challenge.Ternary(found,
				fmt.Sprintf("Type/func %s found in source", typeName),
				fmt.Sprintf("Type/func %s not found in source", typeName)),
		})
	}

	// Check test coverage (at least one test file with test functions)
	c.ReportProgress("check-tests", nil)
	testCount := countFiles(c.modulePath, "*_test.go")
	hasTests := testCount > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "test_coverage",
		Expected: "at least 1 test file",
		Actual:   fmt.Sprintf("%d test files", testCount),
		Passed:   hasTests,
		Message: challenge.Ternary(hasTests,
			fmt.Sprintf("Module has %d test files", testCount),
			"Module has no test files"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	metrics := map[string]challenge.MetricValue{
		"required_types_checked": {Name: "required_types_checked", Value: float64(len(c.requiredTypes)), Unit: "count"},
		"required_files_checked": {Name: "required_files_checked", Value: float64(len(c.requiredFiles)), Unit: "count"},
		"test_files":             {Name: "test_files", Value: float64(testCount), Unit: "count"},
	}

	return c.CreateResult(status, start, assertions, metrics, outputs, ""), nil
}

// sourceContains checks if any .go file in the module contains the given string.
func sourceContains(root, needle string) bool {
	found := false
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || found {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if strings.Contains(string(data), needle) {
			found = true
		}
		return nil
	})
	return found
}

// RegisterModuleFuncChallenges registers functional verification challenges
// for specific module capabilities (MOD-016 to MOD-021).
func RegisterModuleFuncChallenges(svc interface{ Register(challenge.Challenge) error }) {
	challenges := []struct {
		id            string
		name          string
		desc          string
		path          string
		requiredTypes []string
		requiredFiles []string
	}{
		{
			id:   "MOD-016",
			name: "Lazy Value[T] Concurrent Access",
			desc: "Verifies the Lazy module exposes Value[T] with Get, MustGet, Reset, and uses sync.Mutex for thread safety.",
			path: "../Lazy",
			requiredTypes: []string{
				"Value[T any]",
				"func (v *Value[T]) Get()",
				"func (v *Value[T]) MustGet()",
				"func (v *Value[T]) Reset()",
				"sync.Mutex",
			},
			requiredFiles: []string{
				"pkg/lazy/lazy.go",
				"CLAUDE.md",
			},
		},
		{
			id:   "MOD-017",
			name: "Lazy Service[T] Singleton Guarantee",
			desc: "Verifies the Lazy module exposes Service[T] with Get, Initialized, and uses sync.Once for singleton guarantee.",
			path: "../Lazy",
			requiredTypes: []string{
				"Service[T any]",
				"func (s *Service[T]) Get()",
				"func (s *Service[T]) Initialized()",
				"sync.Once",
			},
			requiredFiles: []string{
				"pkg/lazy/lazy.go",
				"go.mod",
			},
		},
		{
			id:   "MOD-018",
			name: "Recovery CircuitBreaker Transitions",
			desc: "Verifies the Recovery module has CircuitBreaker with state transitions (closed, open, half-open).",
			path: "../Recovery",
			requiredTypes: []string{
				"CircuitBreaker",
				"StateClosed",
				"StateOpen",
			},
			requiredFiles: []string{
				"CLAUDE.md",
				"go.mod",
			},
		},
		{
			id:   "MOD-019",
			name: "Recovery HealthChecker Verification",
			desc: "Verifies the Recovery module has health checking capabilities.",
			path: "../Recovery",
			requiredTypes: []string{
				"HealthCheck",
			},
			requiredFiles: []string{
				"CLAUDE.md",
			},
		},
		{
			id:   "MOD-020",
			name: "Memory LeakDetector Tracking",
			desc: "Verifies the Memory module has leak detection and resource tracking capabilities.",
			path: "../Memory",
			requiredTypes: []string{
				"LeakDetector",
				"Track",
			},
			requiredFiles: []string{
				"CLAUDE.md",
				"go.mod",
			},
		},
		{
			id:   "MOD-021",
			name: "Memory KnowledgeGraph BFS",
			desc: "Verifies the Memory module has knowledge graph with breadth-first search.",
			path: "../Memory",
			requiredTypes: []string{
				"KnowledgeGraph",
			},
			requiredFiles: []string{
				"CLAUDE.md",
			},
		},
	}

	for _, ch := range challenges {
		_ = svc.Register(newModuleFuncChallenge(
			ch.id, ch.name, ch.desc, ch.path,
			ch.requiredTypes, ch.requiredFiles,
		))
	}
}
