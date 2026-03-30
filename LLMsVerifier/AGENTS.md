# AGENTS.md - LLMsVerifier

## MANDATORY: No CI/CD Pipelines

**NO GitHub Actions, GitLab CI/CD, or any automated pipeline may exist in this repository!**

- No `.github/workflows/` directory
- No `.gitlab-ci.yml` file
- No Jenkinsfile, .travis.yml, .circleci, or any other CI configuration
- All builds and tests are run manually or via Makefile targets
- This rule is permanent and non-negotiable

## For AI Agents Working on This Codebase

### Module Purpose
LLMsVerifier provides the Strategy pattern for dynamic LLM model verification, scoring, ranking, and selection. It is used by HelixQA to choose the best vision and chat models for QA sessions.

### Key Packages
- `pkg/strategy` — Core `VerificationStrategy` interface, `DefaultStrategy` implementation, types (`ModelInfo`, `StrategyScore`, `Requirements`, `RankedModel`)
- `pkg/recipe` — Fluent builder for composing verification configurations with presets
- `pkg/helixqa` — HelixQA-specific strategy (vision-weighted scoring), known model registry
- `pkg/catalogizer` — Catalogizer-specific strategy configuration

### Dynamic Scoring (No Hardcoded Preferences)
All model selection is score-based with configurable dimension weights:
- Quality (35%), Speed (25%), Cost (20%), Reliability (20%) — defaults
- HelixQA strategy overrides weights to prioritize vision capability
- Models are probed, scored, ranked, and selected at runtime

### Dual Model Selection
- **Vision models** (`SupportsVision: true`) — selected for screenshot analysis phases
- **Chat models** — selected for reasoning, planning, and report generation phases
- Both types go through the same scoring pipeline

### Local Model Probing
- Ollama instances are discovered via `HELIX_OLLAMA_URL`
- Local models receive cost=1.0 (free) and compete on other dimensions
- Distributed hosts (`HELIX_VISION_HOSTS`) are probed individually

### Testing
```bash
go test ./... -race -count=1
```

### Key Interfaces
- `strategy.VerificationStrategy` — Score, Validate, Rank, Select (6 methods)
- `strategy.ModelInfo` — Model metadata with capabilities and benchmarks
- `strategy.Requirements` — Constraint specification for model selection
