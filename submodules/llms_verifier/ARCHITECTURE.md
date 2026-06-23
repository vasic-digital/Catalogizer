# Architecture -- LLMsVerifier

## Purpose

Go module providing the Strategy pattern for LLM model verification, scoring, ranking, and selection. Dynamically evaluates available LLM providers and selects the best model for a given use case based on configurable criteria. Supports phase-specific strategies for vision, navigation, analysis, and planning tasks.

## Structure

```
pkg/
  strategy/      Core VerificationStrategy interface, DefaultStrategy with configurable dimension weights
  vision/        Vision-specialized strategies: VisionStrategy, NavigationStrategy, AnalysisStrategy
  chat/          Chat/reasoning-specialized: PlanningStrategy for test plan generation
  bridge/        Bridged CLI model discovery (Claude Code, Qwen Coder, OpenCode)
  recipe/        Fluent builder API for composing verification configurations with presets
  helixqa/       HelixQA-specific QA session strategy and known vision model registry
  catalogizer/   Catalogizer-specific strategy configuration
  local/         Local model discovery via SSH/Ollama probing
  selector/      Unified model selection across cloud and local providers
```

## Key Components

- **`strategy.VerificationStrategy`** -- Interface: Score, Validate, Rank, Select
- **`strategy.ModelInfo`** -- Model metadata: provider, capabilities, benchmarks, costs, context window
- **`strategy.DefaultStrategy`** -- Balanced scoring across Quality (35%), Speed (25%), Cost (20%), Reliability (20%)
- **`vision.VisionStrategy`** -- Image quality (35%), GUI detection (25%), OCR (15%), Speed (15%), Cost (10%)
- **`vision.NavigationStrategy`** -- JSON compliance (40%), GUI understanding (25%), Speed (20%), Cost (15%)
- **`vision.AnalysisStrategy`** -- Description quality (35%), OCR (20%), Object detection (20%), Comprehensiveness (15%), Cost (10%)
- **`chat.PlanningStrategy`** -- Reasoning (35%), Context window (25%), Structured output (20%), Speed (10%), Cost (10%)
- **`bridge.DiscoverBridgedModels()`** -- Discovers CLI tools on PATH, converts to ModelInfo for unified scoring
- **`local.*`** -- Probes Ollama hosts via SSH, estimates quality from parameter count

## Data Flow

```
strategy.Select(models, requirements)
    |
    filter by requirements (needs vision, max latency, min quality)
    |
    for each model: Score(model) -> weighted dimension scores -> StrategyScore
    |
    Rank(scored models) -> sort by total score -> assign tiers
    |
    Select() -> return highest-ranked model

Bridged models: bridge.DiscoverBridgedModels() -> check PATH for claude/qwen/opencode
    -> BridgedModel.ToModelInfo() -> compete alongside cloud/local in scoring pipeline
```

## Dependencies

- `github.com/stretchr/testify` -- Test assertions (only dependency)

## Testing Strategy

Table-driven tests with `testify` and race detection. Tests cover scoring dimension calculations, ranking order, requirement filtering, strategy weight customization, bridged model discovery with injectable LookPathFunc, and preset configuration validation.
