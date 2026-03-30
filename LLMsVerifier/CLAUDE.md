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
- `pkg/vision/` - Dedicated VisionStrategy for vision model selection (NOT default)
  - `strategy.go` - VisionStrategy scoring: image quality, GUI detection, OCR, speed, cost
  - `strategy_test.go` - Comprehensive tests (29 tests)
- `pkg/recipe/` - Recipe builder for composing verification configurations
  - `builder.go` - Fluent builder API for strategy configuration
  - `presets.go` - Pre-built strategy presets (includes `VisionPreset()`)
  - `validator.go` - Recipe validation
- `pkg/helixqa/` - HelixQA-specific strategy and model definitions
  - `recipe.go` - HelixQA QA session strategy (vision-weighted)
  - `strategy.go` - QA-optimized strategy implementation
  - `models.go` - Known vision model registry with capabilities (12 models across 11 providers)
- `pkg/catalogizer/` - Catalogizer-specific strategy configuration
- `pkg/local/` - Local model discovery via SSH/Ollama probing
- `pkg/selector/` - Unified model selection across cloud and local providers

## Dynamic Model Selection (No Hardcoded Preferences)

LLMsVerifier uses a scoring system with configurable dimension weights:
- **Quality** (default 35%) - Model benchmark scores
- **Speed** (default 25%) - Response latency
- **Cost** (default 20%) - Token pricing
- **Reliability** (default 20%) - Uptime and consistency

Models are probed, scored, ranked, and selected at runtime. There are no hardcoded model preferences. The `Requirements` struct allows callers to specify constraints (needs vision, max latency, min quality, etc.) and the strategy filters and ranks accordingly.

## Strategy Selection Guide

LLMsVerifier provides four strategies, each optimised for a different use case:

| Strategy | Package | When to Use |
|----------|---------|-------------|
| `DefaultStrategy` | `pkg/strategy/` | General-purpose model selection with balanced quality/speed/cost/reliability weights |
| `VisionStrategy` | `pkg/vision/` | Vision model selection: screenshot analysis, UI element detection, OCR. Used by HelixQA Execute/Curiosity phases |
| `QAStrategy` | `pkg/helixqa/` | Autonomous QA sessions: weighs vision + speed + quality for interactive testing |
| `CatalogizerStrategy` | `pkg/catalogizer/` | Catalog-api tasks: metadata enrichment, content analysis, recommendations |

The `VisionStrategy` scores models across five vision-specific dimensions:
- **Image quality** (35%) -- image understanding ability, bonused for specialised vision capabilities
- **GUI detection** (25%) -- UI element detection (gui_analysis, gui_grounding, gui_navigation)
- **OCR accuracy** (15%) -- text recognition capability
- **Speed** (15%) -- response latency normalised to 6s window
- **Cost** (10%) -- free local models get 1.0; cloud models scaled by total token cost

Access the VisionStrategy via `recipe.VisionPreset()` or `vision.NewVisionStrategy()`.

## Dual Model Types

LLMsVerifier distinguishes between two model types for HelixQA:
- **Vision models** (`SupportsVision: true`) - For screenshot analysis (Astica, Gemini, OpenAI GPT-4o, Ollama local models)
- **Chat models** - For reasoning, planning, and report generation (any text-capable model)

The HelixQA QAStrategy weights vision capability higher when selecting models for the Execute and Curiosity phases. The VisionStrategy is available for pure vision model selection without the QA-specific adjustments.

## Local Model Support

LLMsVerifier can probe and score locally-running models via Ollama:
- Models running on `HELIX_OLLAMA_URL` are discovered and included in the scoring pool
- Local models get a cost score of 1.0 (free) and compete on quality/speed/reliability
- Distributed hosts (`HELIX_VISION_HOSTS`) are each probed for available models via SSH
- `pkg/local/` probes hosts, parses `ollama list` output, estimates quality from parameter count
- `pkg/selector/` combines cloud and local candidates into a unified ranked list
- Vision-capable local models (LLaVA, MiniCPM-V, Qwen-VL, etc.) are auto-detected by name patterns

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
