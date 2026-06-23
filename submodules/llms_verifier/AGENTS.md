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


