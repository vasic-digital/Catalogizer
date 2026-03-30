# CLAUDE.md - LLMsVerifier Module

## MANDATORY: No CI/CD Pipelines

**NO GitHub Actions, GitLab CI/CD, or any automated pipeline may exist in this repository!**

- No `.github/workflows/` directory
- No `.gitlab-ci.yml` file
- No Jenkinsfile, .travis.yml, .circleci, or any other CI configuration
- All builds and tests are run manually or via Makefile targets
- This rule is permanent and non-negotiable

## Overview

`digital.vasic.llmsverifier` is a Go module providing the Strategy pattern for LLM model verification, scoring, ranking, and selection. It dynamically evaluates available LLM providers and selects the best model for a given use case based on configurable criteria.

**Module**: `digital.vasic.llmsverifier` (Go 1.24+)

## Build and Test

```bash
go build ./...
go test ./... -race -count=1
go vet ./...
```

## Architecture

- `pkg/strategy/` - Strategy pattern interface and default implementation
  - `interface.go` - `VerificationStrategy` interface, `ModelInfo`, `StrategyScore`, `Requirements` types
  - `default.go` - `DefaultStrategy` with configurable dimension weights
  - `default_test.go` - Table-driven tests for scoring, ranking, selection
- `pkg/recipe/` - Recipe builder for composing verification configurations
  - `builder.go` - Fluent builder API for strategy configuration
  - `presets.go` - Pre-built strategy presets
  - `validator.go` - Recipe validation
- `pkg/helixqa/` - HelixQA-specific strategy and model definitions
  - `recipe.go` - HelixQA QA session strategy (vision-weighted)
  - `strategy.go` - QA-optimized strategy implementation
  - `models.go` - Known model registry with capabilities
- `pkg/catalogizer/` - Catalogizer-specific strategy configuration

## Dynamic Model Selection (No Hardcoded Preferences)

LLMsVerifier uses a scoring system with configurable dimension weights:
- **Quality** (default 35%) - Model benchmark scores
- **Speed** (default 25%) - Response latency
- **Cost** (default 20%) - Token pricing
- **Reliability** (default 20%) - Uptime and consistency

Models are probed, scored, ranked, and selected at runtime. There are no hardcoded model preferences. The `Requirements` struct allows callers to specify constraints (needs vision, max latency, min quality, etc.) and the strategy filters and ranks accordingly.

## Dual Model Types

LLMsVerifier distinguishes between two model types for HelixQA:
- **Vision models** (`SupportsVision: true`) - For screenshot analysis (Astica, Gemini, OpenAI GPT-4o, Ollama local models)
- **Chat models** - For reasoning, planning, and report generation (any text-capable model)

The HelixQA strategy weights vision capability higher when selecting models for the Execute and Curiosity phases.

## Local Model Support

LLMsVerifier can probe and score locally-running models via Ollama:
- Models running on `HELIX_OLLAMA_URL` are discovered and included in the scoring pool
- Local models get a cost score of 1.0 (free) and compete on quality/speed/reliability
- Distributed hosts (`HELIX_VISION_HOSTS`) are each probed for available models

## Key Interfaces

- `VerificationStrategy` - Core interface: Score, Validate, Rank, Select
- `ModelInfo` - Model metadata (provider, capabilities, benchmarks, costs)
- `StrategyScore` - Detailed scoring breakdown with confidence and reasoning
- `Requirements` - Caller-specified constraints for model selection
- `RankedModel` - Model with rank position, score, tier, and selection probability

## Functional Options

Strategies accept `StrategyOption` functions:
- `WithWeights(map[string]float64)` - Custom dimension weights
- `WithConstraints([]Constraint)` - Hard/soft verification constraints
- `WithFallbacks([]FallbackRule)` - Degradation rules for provider failures

## Code Style

- Standard Go conventions, `gofmt` formatting
- SPDX headers on every .go file
- Table-driven tests with `testify`
- Thread-safe scoring cache with `sync.RWMutex`
