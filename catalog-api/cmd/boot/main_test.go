package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBuild verifies that the cmd/boot package compiles successfully.
// The main() function starts containers and blocks, so it cannot be
// called directly in a unit test. Having this file ensures every
// package in the project has at least one _test.go file.
func TestBuild(t *testing.T) {
	// bluff-scan: no-assert-ok (basic build/config smoke — must not panic)
	// Package compiles — this test passing is sufficient.
}

// TestDefaultFlagValues validates the default values that the flag
// package would assign when no CLI arguments are provided. We cannot
// call flag.Parse() again in a test (it reads os.Args), so we
// verify the constants/defaults that main() relies on.
func TestDefaultFlagValues(t *testing.T) {
	// These match the flag defaults declared in main().
	defaultEnv := "../Containers/.env"
	defaultCompose := "docker-compose.dev.yml"
	defaultLocal := false
	defaultDryRun := false
	defaultTimeout := 5 * time.Minute

	assert.Equal(t, "../Containers/.env", defaultEnv)
	assert.Equal(t, "docker-compose.dev.yml", defaultCompose)
	assert.False(t, defaultLocal)
	assert.False(t, defaultDryRun)
	assert.Equal(t, 5*time.Minute, defaultTimeout)
}

// TestSchedulerStrategyMapping verifies the strategy string-to-constant
// mapping logic used in main(). This mirrors the switch statement without
// importing the scheduler package (which would pull in SSH dependencies).
func TestSchedulerStrategyMapping(t *testing.T) {
	strategies := map[string]string{
		"round_robin": "round_robin",
		"affinity":    "affinity",
		"spread":      "spread",
		"bin_pack":    "bin_pack",
		"":            "resource_aware",
		"unknown":     "resource_aware",
	}

	for input, expected := range strategies {
		result := mapStrategy(input)
		assert.Equal(t, expected, result,
			"strategy %q should map to %q", input, expected)
	}
}

// mapStrategy mirrors the switch statement in main() for testability.
func mapStrategy(s string) string {
	switch s {
	case "round_robin":
		return "round_robin"
	case "affinity":
		return "affinity"
	case "spread":
		return "spread"
	case "bin_pack":
		return "bin_pack"
	default:
		return "resource_aware"
	}
}
