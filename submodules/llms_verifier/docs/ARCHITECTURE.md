# LLMsVerifier Architecture

## Module Structure

```
LLMsVerifier/
├── pkg/
│   ├── strategy/      # Core Strategy pattern interface + default implementation
│   ├── vision/        # Vision-specialised strategies (Vision, Navigation, Analysis)
│   ├── chat/          # Chat/reasoning-specialised strategies (Planning)
│   ├── bridge/        # Bridged CLI model discovery (Claude Code, Qwen, OpenCode)
│   ├── recipe/        # Fluent builder for verification configurations
│   ├── helixqa/       # HelixQA-specific strategy (vision-weighted)
│   ├── catalogizer/   # Catalogizer-specific strategy
│   ├── local/         # Local model discovery via SSH/Ollama probing
│   └── selector/      # Unified model selection across cloud and local providers
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
     ┌───────────┬───────────┼───────────┬───────────┬───────────┐
     │           │           │           │           │           │
┌────▼────┐ ┌───▼─────┐ ┌───▼─────┐ ┌───▼─────┐ ┌───▼────┐ ┌───▼──────┐
│ Default │ │ Vision  │ │ Navig-  │ │Analysis │ │Planning│ │ HelixQA  │
│Strategy │ │Strategy │ │ ation   │ │Strategy │ │Strategy│ │ /Catalog │
│(balanced│ │(vision) │ │Strategy │ │(descr.) │ │(reason)│ │ Strategy │
└─────────┘ └─────────┘ └─────────┘ └─────────┘ └────────┘ └──────────┘
 pkg/        pkg/        pkg/        pkg/        pkg/       pkg/helixqa
 strategy/   vision/     vision/     vision/     chat/      pkg/catalog
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

Presets provide pre-configured recipes for common use cases:

- `VisionPreset()` -- general vision model selection
- `NavigationPreset()` -- JSON-action navigation model selection
- `AnalysisPreset()` -- rich description analysis model selection
- `PlanningPreset()` -- reasoning/planning model selection

---

## NavigationStrategy (pkg/vision/)

Specialised strategy for UI navigation tasks. Prioritises models that produce
structured JSON action arrays from screenshots over models that return natural
language descriptions.

### Scoring Dimensions

| Dimension | Weight | Scoring Logic |
|-----------|--------|---------------|
| JSON compliance | 40% | 1.0 for json_output/structured_output capability; 0.2 otherwise |
| GUI understanding | 25% | Scaled by gui_analysis/gui_grounding/gui_navigation caps (0-4) |
| Speed | 20% | Normalised against 5s ideal (sub-1s = close to 1.0) |
| Cost | 15% | Free models = 1.0; cloud scaled by total token cost |

Non-vision models receive an overall score of 0. Vision models without
JSON output capability are heavily penalised (score 0.2 on the JSON
dimension) because navigation requires machine-parseable action arrays.

---

## AnalysisStrategy (pkg/vision/)

Specialised strategy for screenshot analysis tasks. Prioritises models that
produce rich, detailed image descriptions with OCR and object detection.

### Scoring Dimensions

| Dimension | Weight | Scoring Logic |
|-----------|--------|---------------|
| Description quality | 35% | QualityScore + bonus for vision/ocr/object_detection/face_detection caps |
| OCR accuracy | 20% | 0.95 for ocr capability; 0.4 * QualityScore otherwise |
| Object detection | 20% | 0.95 for object_detection capability; 0.35 * QualityScore otherwise |
| Comprehensiveness | 15% | Count of 6 analysis capabilities normalised to 0-1 |
| Cost | 10% | Free models = 1.0; cloud scaled by total token cost |

Models like Astica with specialised vision APIs (OCR, object detection, face
detection, content moderation) score highest. General-purpose LLMs with vision
(Gemini, GPT-4o) score lower.

---

## PlanningStrategy (pkg/chat/)

Specialised strategy for test plan generation. Prioritises models with strong
reasoning, large context windows, and structured output ability.

### Scoring Dimensions

| Dimension | Weight | Scoring Logic |
|-----------|--------|---------------|
| Reasoning quality | 35% | QualityScore + 0.10 bonus for "reasoning" capability |
| Context window | 25% | Normalised against 1M tokens (larger = better) |
| Structured output | 20% | 0.9 for json_output/structured_output; 0.7 for code; 0.3 baseline |
| Speed | 10% | Normalised against 10s (planning tolerates higher latency) |
| Cost | 10% | Free models = 1.0; cloud scaled by total token cost |

Claude and GPT-4 score highest due to strong reasoning and large context. Gemini
scores close second (1M context). Local Ollama models score lower (smaller
context, weaker reasoning).

---

## Bridged CLI Model Discovery (pkg/bridge/)

The bridge package discovers LLM models accessible via CLI coding assistants
installed on the developer's machine. Unlike cloud API models (API key required)
or local Ollama models (server required), bridged models are accessed by invoking
a CLI binary that manages its own authentication and routing.

### Discovery Flow

```
DiscoverBridgedModels()
  └─ For each known CLI spec (claude, qwen-coder, opencode):
       └─ exec.LookPath(binary_name)
            ├─ found → BridgedModel{Available: true, CLIPath: path}
            └─ not found → skip
  └─ Return []BridgedModel

BridgedModel.ToModelInfo()
  └─ strategy.ModelInfo{
       ID:           "bridged-claude-claude-sonnet-4",
       Provider:     "anthropic",
       Cost:         0 (CLI handles billing),
       Capabilities: ["text", "bridged", "coding", "vision"],
     }
```

### Known CLI Tools

| CLI | Binary Names | Provider | Model | Vision | Context |
|-----|-------------|----------|-------|--------|---------|
| Claude Code | `claude` | anthropic | claude-sonnet-4 | Yes | 200K |
| Qwen Coder | `qwen-coder`, `qwen` | qwen | qwen3-coder | No | 128K |
| OpenCode | `opencode`, `zen` | opencode | opencode | No | 128K |

Bridged models are scored alongside cloud and local providers by all strategies.
They get a perfect cost score (1.0) since they are free to the scoring system.

---

## Phase-Specific Strategy Mapping (HelixQA Integration)

When used by HelixQA's autonomous pipeline, each phase selects the optimal
strategy for its task:

```
┌──────────┐     ┌─────────────────────┐     ┌────────────────┐
│  Learn   │────►│  PlanningStrategy   │────►│ Chat model     │
│  Phase   │     │  (reasoning+context)│     │ (Gemini, Claude│
└──────────┘     └─────────────────────┘     └────────────────┘

┌──────────┐     ┌─────────────────────┐     ┌────────────────┐
│  Plan    │────►│  PlanningStrategy   │────►│ Chat model     │
│  Phase   │     │  (reasoning+struct) │     │ (Claude, GPT)  │
└──────────┘     └─────────────────────┘     └────────────────┘

┌──────────┐     ┌─────────────────────┐     ┌────────────────┐
│ Execute  │────►│ NavigationStrategy  │────►│ Vision model   │
│  Phase   │     │  (JSON+GUI+speed)   │     │ (Gemini Flash) │
└──────────┘     └─────────────────────┘     └────────────────┘

┌──────────┐     ┌─────────────────────┐     ┌────────────────┐
│Curiosity │────►│ NavigationStrategy  │────►│ Vision model   │
│  Phase   │     │  (JSON+GUI+speed)   │     │ (Gemini Flash) │
└──────────┘     └─────────────────────┘     └────────────────┘

┌──────────┐     ┌─────────────────────┐     ┌────────────────┐
│ Analyze  │────►│  AnalysisStrategy   │────►│ Vision model   │
│  Phase   │     │  (descr+OCR+detect) │     │ (Astica, GPT4o)│
└──────────┘     └─────────────────────┘     └────────────────┘
```
