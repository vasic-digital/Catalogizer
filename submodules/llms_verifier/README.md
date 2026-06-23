# LLMsVerifier

Strategy-based LLM model verification, scoring, ranking, and selection for Go. Dynamically evaluates available LLM providers and selects the best model for a given use case based on configurable criteria.

**Module:** `digital.vasic.llmsverifier`

## Features

- **Strategy pattern** -- Pluggable verification strategies with configurable dimension weights
- **7 built-in strategies** -- Default, Vision, Navigation, Analysis, Planning, QA, Catalogizer
- **Dynamic scoring** -- Models probed, scored, ranked, and selected at runtime (no hardcoded preferences)
- **Bridged CLI discovery** -- Discovers Claude Code, Qwen Coder, OpenCode on PATH
- **Local model support** -- Probes Ollama instances via SSH across distributed hosts
- **Unified selection** -- Cloud, local, and bridged models compete in a single scoring pipeline
- **Recipe builder** -- Fluent API with presets for quick strategy configuration

## Quick Start

```go
import (
    "digital.vasic.llmsverifier/pkg/strategy"
    "digital.vasic.llmsverifier/pkg/vision"
)

// Create a vision strategy for screenshot analysis
vs := vision.NewVisionStrategy()

// Score and rank available models
models := []strategy.ModelInfo{
    {Name: "gemini-2.0-flash", Provider: "google", SupportsVision: true, ...},
    {Name: "gpt-4o", Provider: "openai", SupportsVision: true, ...},
}

ranked := vs.Rank(models)
best := vs.Select(models, &strategy.Requirements{NeedsVision: true})
```

## Strategies

| Strategy | Package | Use Case |
|----------|---------|----------|
| `DefaultStrategy` | `pkg/strategy/` | General-purpose balanced selection (quality/speed/cost/reliability) |
| `VisionStrategy` | `pkg/vision/` | Screenshot analysis, UI detection, OCR |
| `NavigationStrategy` | `pkg/vision/` | JSON-action producing models for UI navigation |
| `AnalysisStrategy` | `pkg/vision/` | Rich image descriptions for screenshot analysis |
| `PlanningStrategy` | `pkg/chat/` | Test plan generation with strong reasoning |
| `QAStrategy` | `pkg/helixqa/` | Autonomous QA sessions (vision + speed + quality) |
| `CatalogizerStrategy` | `pkg/catalogizer/` | Metadata enrichment and content analysis |

## Scoring Dimensions

Each strategy weights dimensions differently. For example, VisionStrategy:

- **Image quality** (35%) -- Image understanding ability
- **GUI detection** (25%) -- UI element detection capabilities
- **OCR accuracy** (15%) -- Text recognition capability
- **Speed** (15%) -- Response latency normalized to 6s window
- **Cost** (10%) -- Free local models score 1.0; cloud models scaled by token cost

## Bridged CLI Models

`pkg/bridge/` discovers LLM models available via CLI coding assistants:

| CLI Binary | Provider | Model | Vision | Context |
|------------|----------|-------|--------|---------|
| `claude` | Anthropic | claude-sonnet-4 | Yes | 200K |
| `qwen-coder` / `qwen` | Qwen | qwen3-coder | No | 128K |
| `opencode` / `zen` | OpenCode | opencode | No | 128K |

Bridged models get zero token cost (CLI handles billing) and compete alongside cloud and local models.

## Local Model Discovery

Models running on Ollama instances are discovered via SSH probing:

```go
import "digital.vasic.llmsverifier/pkg/local"

// Probe hosts for available models
hosts := []string{"thinker.local", "amber.local"}
localModels := local.DiscoverModels(hosts)
```

Vision-capable local models (LLaVA, MiniCPM-V, Qwen-VL) are auto-detected by name patterns.

## Recipe Builder

```go
import "digital.vasic.llmsverifier/pkg/recipe"

// Use presets
r := recipe.VisionPreset()     // Pre-configured for vision model selection
r := recipe.NavigationPreset() // Pre-configured for navigation
r := recipe.AnalysisPreset()   // Pre-configured for analysis
r := recipe.PlanningPreset()   // Pre-configured for planning
```

## Packages

| Package | Description |
|---------|-------------|
| `pkg/strategy` | Core VerificationStrategy interface, DefaultStrategy, ModelInfo, Requirements |
| `pkg/vision` | VisionStrategy, NavigationStrategy, AnalysisStrategy |
| `pkg/chat` | PlanningStrategy for reasoning/planning tasks |
| `pkg/bridge` | Bridged CLI model discovery (Claude Code, Qwen Coder, OpenCode) |
| `pkg/recipe` | Fluent builder API with strategy presets |
| `pkg/helixqa` | HelixQA-specific QA session strategy and model registry |
| `pkg/catalogizer` | Catalogizer-specific strategy configuration |
| `pkg/local` | Local model discovery via SSH/Ollama probing |
| `pkg/selector` | Unified model selection across cloud and local providers |

## Build

```bash
go build ./...
go test ./... -race -count=1
go vet ./...
```

## License

Apache-2.0
