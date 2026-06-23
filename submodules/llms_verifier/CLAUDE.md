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
- `pkg/vision/` - Vision-specialised strategies for model selection
  - `strategy.go` - `VisionStrategy` scoring: image quality, GUI detection, OCR, speed, cost
  - `navigation.go` - `NavigationStrategy` scoring: JSON compliance, GUI understanding, speed, cost
  - `analysis.go` - `AnalysisStrategy` scoring: description quality, OCR, object detection, comprehensiveness, cost
  - `strategy_test.go`, `navigation_test.go`, `analysis_test.go` - Comprehensive tests
- `pkg/chat/` - Chat/reasoning-specialised strategies
  - `planning.go` - `PlanningStrategy` scoring: reasoning quality, context window, structured output, speed, cost
  - `planning_test.go` - Comprehensive tests
- `pkg/bridge/` - Bridged CLI model discovery and scoring
  - `discovery.go` - Discovers CLI coding assistants on PATH (Claude Code, Qwen Coder, OpenCode)
  - `provider.go` - Converts bridged models to `strategy.ModelInfo`, filtering and sorting utilities
  - `discovery_test.go` - Tests with injectable `LookPathFunc`
- `pkg/recipe/` - Recipe builder for composing verification configurations
  - `builder.go` - Fluent builder API for strategy configuration
  - `presets.go` - Pre-built strategy presets (`VisionPreset`, `NavigationPreset`, `AnalysisPreset`, `PlanningPreset`)
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
| `NavigationStrategy` | `pkg/vision/` | JSON-action producing vision models for UI navigation. Prioritises JSON compliance, GUI understanding, speed, cost |
| `AnalysisStrategy` | `pkg/vision/` | Rich image description models for screenshot analysis. Prioritises description quality, OCR accuracy, object detection |
| `PlanningStrategy` | `pkg/chat/` | Strong reasoning models for test plan generation. Prioritises reasoning quality, context window, structured output |
| `QAStrategy` | `pkg/helixqa/` | Autonomous QA sessions: weighs vision + speed + quality for interactive testing |
| `CatalogizerStrategy` | `pkg/catalogizer/` | Catalog-api tasks: metadata enrichment, content analysis, recommendations |

The `VisionStrategy` scores models across five vision-specific dimensions:
- **Image quality** (35%) -- image understanding ability, bonused for specialised vision capabilities
- **GUI detection** (25%) -- UI element detection (gui_analysis, gui_grounding, gui_navigation)
- **OCR accuracy** (15%) -- text recognition capability
- **Speed** (15%) -- response latency normalised to 6s window
- **Cost** (10%) -- free local models get 1.0; cloud models scaled by total token cost

Access strategies via presets or direct constructors:
- `recipe.VisionPreset()` or `vision.NewVisionStrategy()` -- general vision model selection
- `recipe.NavigationPreset()` or `vision.NewNavigationStrategy()` -- JSON-action navigation model selection
- `recipe.AnalysisPreset()` or `vision.NewAnalysisStrategy()` -- rich description analysis model selection
- `recipe.PlanningPreset()` or `chat.NewPlanningStrategy()` -- reasoning/planning model selection

The `NavigationStrategy` scores models across four navigation-specific dimensions:
- **JSON compliance** (40%) -- ability to produce structured JSON action arrays (not descriptions)
- **GUI understanding** (25%) -- GUI element detection, focus tracking, layout comprehension
- **Speed** (20%) -- response latency normalised to 5s window (interactive navigation needs fast responses)
- **Cost** (15%) -- free local models get 1.0; cloud models scaled by total token cost

The `AnalysisStrategy` scores models across five analysis-specific dimensions:
- **Description quality** (35%) -- rich image description ability with vision capability bonuses
- **OCR accuracy** (20%) -- text recognition capability (dedicated OCR models score highest)
- **Object detection** (20%) -- UI element and object identification
- **Comprehensiveness** (15%) -- breadth of vision analysis capabilities
- **Cost** (10%) -- free local models get 1.0; cloud models scaled by total token cost

The `PlanningStrategy` scores models across five planning-specific dimensions:
- **Reasoning quality** (35%) -- multi-step reasoning ability with bonus for "reasoning" capability
- **Context window** (25%) -- normalised against 1M tokens (larger = better for knowledge bases)
- **Structured output** (20%) -- JSON/structured output or code generation capability
- **Speed** (10%) -- normalised against 10s (planning tolerates higher latency)
- **Cost** (10%) -- free local models get 1.0; cloud models scaled by total token cost

## Bridged CLI Models

`pkg/bridge/` discovers LLM models available via CLI coding assistant tools installed on the developer's machine. These models are not called via HTTP APIs but through their CLI binaries, which handle authentication and billing internally.

### Discovered CLI Tools

| CLI Binary | Provider | Model | Vision | Context |
|------------|----------|-------|--------|---------|
| `claude` | Anthropic | claude-sonnet-4 | Yes | 200K |
| `qwen-coder` / `qwen` | Qwen | qwen3-coder | No | 128K |
| `opencode` / `zen` | OpenCode | opencode | No | 128K |

### How It Works

1. `DiscoverBridgedModels()` searches `PATH` for known CLI binaries
2. Found CLIs are converted to `BridgedModel` structs with capability metadata
3. `BridgedModel.ToModelInfo()` converts to `strategy.ModelInfo` for unified scoring
4. Bridged models get zero token cost (CLI handles billing) and a "bridged" capability tag
5. They compete alongside cloud and local models in the scoring/ranking pipeline

### Key Functions

- `bridge.DiscoverBridgedModels()` -- discover all available CLI tools
- `bridge.BridgedToModelInfos(models)` -- convert to `[]strategy.ModelInfo`
- `bridge.FilterVisionCapable(models)` -- filter to vision-capable CLIs only
- `bridge.GetProviderInfo(cliName)` -- get display metadata for a CLI tool

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


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** use `su` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Container-Based Solutions
When a build or runtime environment requires system-level dependencies, use containers instead of elevation:

- **Use the `Containers` submodule** (`https://github.com/vasic-digital/Containers`) for containerized build and runtime environments
- **Add the `Containers` submodule as a Git dependency** and configure it for local use within the project
- **Build and run inside containers** to avoid any need for privilege escalation
- **Rootless Podman/Docker** is the preferred container runtime

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo` or `su`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Use the `Containers` submodule for containerized builds
5. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**


