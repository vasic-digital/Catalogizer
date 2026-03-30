# LLMsVerifier Architecture

## Module Structure

```
LLMsVerifier/
├── pkg/
│   ├── strategy/      # Core Strategy pattern interface + default implementation
│   ├── recipe/        # Fluent builder for verification configurations
│   ├── helixqa/       # HelixQA-specific strategy (vision-weighted)
│   └── catalogizer/   # Catalogizer-specific strategy
├── go.mod
└── docs/
    └── ARCHITECTURE.md
```

## Design Pattern: Strategy

LLMsVerifier implements the Strategy pattern for LLM model selection:

```
                ┌──────────────────────────┐
                │  VerificationStrategy    │
                │  (interface)             │
                ├──────────────────────────┤
                │  Score(model) -> Score   │
                │  Validate(model) -> bool │
                │  Rank(models) -> ranked  │
                │  Select(ranked, reqs)    │
                └────────────┬─────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────┐  ┌─────▼──────┐ ┌────▼──────────┐
     │  Default     │  │  HelixQA   │ │  Catalogizer  │
     │  Strategy    │  │  Strategy  │ │  Strategy     │
     │  (balanced)  │  │  (vision)  │ │  (project)    │
     └─────────────┘  └────────────┘ └───────────────┘
```

## Scoring Pipeline

```
Probe providers → Collect ModelInfo → Score each model → Rank → Filter by Requirements → Select best
```

### Dimension Weights (DefaultStrategy)

| Dimension | Weight | Source |
|-----------|--------|--------|
| Quality | 35% | Benchmark scores (0-1) |
| Speed | 25% | Inverse of latency (ms) |
| Cost | 20% | Inverse of token pricing |
| Reliability | 20% | Uptime rating (0-1) |

HelixQA strategy overrides these weights to prioritize vision capability.

### Scoring Formula

```
dimension_score = raw_value * weight
overall = sum(dimension_scores) / sum(weights)
confidence = 0.95 if verified, 0.8 otherwise (decays 10% if stale > 24h)
```

## Dual Model Types

LLMsVerifier distinguishes two model types:

| Type | `SupportsVision` | Used For | Examples |
|------|-------------------|----------|----------|
| Vision | `true` | Screenshot analysis, UI comprehension | Astica, Gemini, GPT-4o, Ollama minicpm-v |
| Chat | `false` | Planning, reasoning, report generation | GPT-4, Claude, Mistral, Groq |

The HelixQA strategy selects vision models for Execute/Curiosity phases and chat models for Learn/Plan/Analyze phases.

## Local Model Support

Ollama-hosted models are included in the scoring pool:

```
HELIX_OLLAMA_URL=http://thinker.local:11434
                    │
                    ▼
        ┌───────────────────────┐
        │  Probe Ollama API     │
        │  GET /api/tags        │
        └───────────┬───────────┘
                    │ discovered models
                    ▼
        ┌───────────────────────┐
        │  Create ModelInfo     │
        │  cost = 0 (free)     │
        │  latency = measured   │
        │  quality = benchmark  │
        └───────────┬───────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │  Score alongside      │
        │  cloud providers      │
        └───────────────────────┘
```

Local models get a perfect cost score (1.0) since they are free. They compete with cloud providers on quality, speed, and reliability dimensions.

## Thread Safety

- `DefaultStrategy` uses `sync.RWMutex` for weight access and score cache
- Score cache entries expire after 5 minutes
- All public methods are safe for concurrent use

## Tier Classification

Models are classified into tiers based on overall score:

| Tier | Score Range | Description |
|------|-------------|-------------|
| Tier 1 | >= 0.8 | Top-tier models |
| Tier 2 | 0.6 - 0.8 | Mid-tier models |
| Tier 3 | < 0.6 | Budget-tier models |

## Recipe Builder

The `pkg/recipe` package provides a fluent API for composing strategies:

```go
recipe := recipe.NewBuilder().
    WithStrategy("helix-qa").
    WithMinScore(0.6).
    WithMaxModels(5).
    WithVisionRequired(true).
    Build()
```

Presets provide pre-configured recipes for common use cases.
