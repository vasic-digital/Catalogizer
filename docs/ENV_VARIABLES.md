# `.env` Variables — Canonical Reference

**Purpose:** single source of truth for every environment variable the Catalogizer + HelixQA + OCU + Containers stack reads. Prepare values here, hand back, and the next session extends the root `.env` file accordingly.

**File location:** project root `/run/media/milosvasic/DATA4TB/Projects/Catalogizer/.env` (gitignored; `chmod 600`).

**Legend:**
- 🔴 **REQUIRED** — stack fails or degrades significantly without it
- 🟡 **RECOMMENDED** — enables full feature set; defaults work but are limited
- 🟢 **OPTIONAL** — fine-tuning / opt-in features
- ⚫ **CI-ONLY** — leave UNSET in production (kill-switches, stubs)
- ✅ already in your `.env`
- ❌ missing from your `.env`

**Format for each entry:**
```
NAME=<example value>          # purpose / who reads it
```

---

## 1. Authentication & session secrets

```bash
JWT_SECRET=<32+ byte random>              # ✅ catalog-api — signs all JWT tokens
ADMIN_USERNAME=admin                       # ✅ catalog-api — seed admin on first boot
ADMIN_PASSWORD=<secure random>             # ✅ catalog-api — seed admin password
HELIXAGENT_API_KEY=<secret>                # ✅ HelixAgent — inter-service auth
```

## 2. Database

```bash
# PostgreSQL (production)
DATABASE_TYPE=postgres                     # ✅ catalog-api — postgres | sqlite
DATABASE_HOST=localhost                    # ✅
DATABASE_PORT=5433                         # ✅ (container port mapping; 5432 in-pod)
DATABASE_NAME=catalogizer                  # ✅
DATABASE_USER=catalogizer                  # ✅
DATABASE_PASSWORD=<secure>                 # ✅
DATABASE_SSL_MODE=disable                  # ✅ (or require for remote)

# Shorthand (legacy tools)
DB_HOST=localhost                          # ✅
DB_PORT=5433                               # ✅
DB_NAME=catalogizer                        # ✅
DB_USER=catalogizer                        # ✅
DB_PASSWORD=<secure>                       # ✅
POSTGRES_URL=postgres://...                # ✅ (URL form for some clients)
```

## 3. Caching / vector / eventing

```bash
REDIS_HOST=localhost                       # ✅
REDIS_PORT=6379                            # ✅
REDIS_PASSWORD=                            # ✅ (empty for local)
REDIS_URL=redis://localhost:6379           # ✅

CHROMA_URL=http://localhost:8000           # ✅ vector store
QDRANT_URL=http://localhost:6333           # ✅ vector store (alternative)

COGNEE_AUTH_EMAIL=<email>                  # ✅ Cognee memory
COGNEE_AUTH_PASSWORD=<pass>                # ✅
```

## 4. Service-level feature flags (already present)

```bash
SVC_POSTGRESQL_ENABLED=true                # ✅
SVC_POSTGRESQL_REMOTE=true|false           # ✅
SVC_POSTGRESQL_HOST=thinker.local          # ✅ (when REMOTE=true)
SVC_POSTGRESQL_PORT=5432                   # ✅
SVC_REDIS_REMOTE=true|false                # ✅
SVC_REDIS_HOST=thinker.local               # ✅
SVC_REDIS_PORT=6379                        # ✅
SVC_CHROMADB_ENABLED=true                  # ✅
SVC_CHROMADB_REMOTE=true|false             # ✅
SVC_CHROMADB_HOST=thinker.local            # ✅
SVC_CHROMADB_REQUIRED=true                 # ✅
SVC_COGNEE_ENABLED=true                    # ✅
SVC_COGNEE_REQUIRED=false                  # ✅
SVC_QDRANT_REMOTE=true                     # ✅
SVC_QDRANT_HOST=thinker.local              # ✅

BOOT_STRICT_MODE=false                     # ✅ Containers/pkg/boot
```

## 5. Server binding

```bash
SERVER_HOST=0.0.0.0                        # ✅
SERVER_PORT=8080                           # ✅ (catalog-api writes to .service-port)
```

## 6. LLM providers — vision & chat

### 6.1 🟡 RECOMMENDED (at least one required for autonomous QA)

```bash
# ✅ Already present in your .env:
GEMINI_API_KEY=<gemini>                    # Google Gemini (you have pay-as-you-go)
CLAUDE_CODE_USE_OAUTH_CREDENTIALS=true     # Anthropic via OAuth
KIMI_API_KEY=<kimi>                        # Moonshot Kimi
MISTRAL_API_KEY=<mistral>
GROQ_API_KEY=<groq>
DEEPSEEK_API_KEY=<deepseek>
CEREBRAS_API_KEY=<cerebras>
COHERE_API_KEY=<cohere>
HUGGINGFACE_API_KEY=<hf>
NVIDIA_API_KEY=<nvidia-nim>
OPENROUTER_API_KEY=<openrouter>            # routes to OpenAI/Anthropic/anything
REPLICATE_API_KEY=<replicate>
SAMBANOVA_API_KEY=<sambanova>
FIREWORKS_API_KEY=<fireworks>
HYPERBOLIC_API_KEY=<hyperbolic>
ZHIPU_API_KEY=<zhipu>
ZAI_API_KEY=<zai>
VERTEX_API_KEY=<vertex>
UPSTAGE_API_KEY=<upstage>
NOVITA_API_KEY=<novita>
PUBLICAI_API_KEY=<publicai>
SARVAM_API_KEY=<sarvam>
VULAVULA_API_KEY=<vulavula>
CLOUDFLARE_API_KEY=<cf>
CODESTRAL_API_KEY=<codestral>
SILICONFLOW_API_KEY=<silicon>
VENICE_API_KEY=<venice>
JUNIE_API_KEY=<junie>
KILO_API_KEY=<kilo>
LETTA_API_KEY=<letta>
MEMO_API_KEY=<memo>
NIA_API_KEY=<nia>
NLP_API_KEY=<nlpcloud>
MODAL_API_KEY=<modal>
MODAL_API_KEY_ID=<modal-id>
CHUTES_API_KEY=<chutes>
ASTICA_API_KEY=<astica>                    # Analyze-only (rich image description)
QWEN_CODE_USE_OAUTH_CREDENTIALS=true       # Qwen via OAuth
```

### 6.2 🟡 Direct vendor APIs (optional since OAuth/OpenRouter cover them)

```bash
OPENAI_API_KEY=<sk-...>                    # ❌ direct GPT-4o/omni (OpenRouter routes as fallback)
ANTHROPIC_API_KEY=<sk-ant-...>             # ❌ direct Claude (OAuth works without this)
```

### 6.3 🟢 Vision pipeline tuning (defaults are sensible)

```bash
HELIX_VISION_PROVIDER=                     # ❌ force provider (auto-ranked if unset)
HELIX_VISION_GEMINI_MODEL=gemini-2.5-flash # ❌ override
HELIX_VISION_OPENAI_MODEL=gpt-4o-mini      # ❌
HELIX_VISION_ANTHROPIC_MODEL=claude-sonnet-4-6  # ❌
HELIX_VISION_KIMI_MODEL=moonshot-v1-32k-vision  # ❌
HELIX_VISION_QWEN_MODEL=qwen-vl-max        # ❌
HELIX_VISION_ASTICA_MODEL=2.5_full         # ❌
HELIX_VISION_STEPGUI_MODEL=step-1v-32k     # ❌
HELIX_VISION_FALLBACK_CHAIN=gemini,openrouter,kimi  # ❌ provider order
HELIX_VISION_OPENCV_ENABLED=false          # ❌ local OpenCV fallback
HELIX_VISION_CHEAPER_ENABLED=true          # ❌ cost-saving routing
HELIX_VISION_LEARNING_ENABLED=false        # ❌ predictor learning loop
HELIX_VISION_MAX_IMAGE_SIZE=2097152        # ❌ upload cap (bytes)
HELIX_VISION_MAX_MEMORIES=1000             # ❌
HELIX_VISION_MULTI_USER=false              # ❌
HELIX_VISION_PERSIST_PATH=                 # ❌ optional cache dir
HELIX_VISION_TIMEOUT=60                    # ❌ seconds
HELIX_VISION_CHANGE_THRESHOLD=0.1          # ❌ diff sensitivity
HELIX_VISION_SSIM_THRESHOLD=0.95           # ❌ SSIM comparator threshold
HELIX_VISION_MODEL=                        # ❌ single-model override
```

### 6.4 🟢 llama.cpp RPC distributed inference (on-prem, multi-host)

```bash
HELIX_LLAMACPP=true                        # ❌ enable distributed inference
HELIX_LLAMACPP_MODEL=qwen2.5-vl-7b-instruct.gguf            # ❌ GGUF on master
HELIX_LLAMACPP_MMPROJ=mmproj-qwen2.5-vl-7b-instruct.gguf    # ❌ multimodal projector
HELIX_LLAMACPP_RPC_MODEL=                  # ❌ optional override
HELIX_LLAMACPP_RPC_WORKERS=thinker.local:50052,amber.local:50052  # ❌
HELIX_LLAMACPP_FREE_GPU=true               # ❌ stop Ollama to reclaim VRAM
HELIX_VISION_HOST=thinker.local            # ❌ master node
HELIX_VISION_HOSTS=thinker.local,amber.local  # ❌ all hosts
HELIX_VISION_USER=milosvasic               # ❌ SSH user
HELIX_OLLAMA_MODEL=minicpm-v:8b            # ❌ Ollama fallback model
HELIX_OLLAMA_URL=http://thinker.local:11434  # ❌ already common pattern
```

## 7. Metadata providers (for Catalogizer catalogue)

### 🟡 RECOMMENDED — real catalogue testing per Constitution §7.4

```bash
TMDB_API_KEY=<tmdb-v3>                     # ❌ movie/TV metadata
TMDB_ACCESS_TOKEN=<tmdb-v4>                # ❌ bearer token for TMDB v4
OMDB_API_KEY=<omdb>                        # ❌ fallback movie metadata
IGDB_CLIENT_ID=<twitch-igdb>               # ❌ game metadata
IGDB_CLIENT_SECRET=<twitch-igdb>           # ❌
# OR alternative:
IGDB_BEARER_TOKEN=<long-token>             # ❌ pre-exchanged bearer
FANART_TV_API_KEY=<fanart>                 # ❌ poster / art
```

### 🟢 OPTIONAL providers (auto-skip if absent)

```bash
# OpenLibrary, MusicBrainz, GitHub software search — no keys required
```

## 8. OCU (OpenClaw Ultimate) — GPU sidecar + distribution

### 🟡 Required to unlock the NVENC + CUDA OpenCV end-to-end path

```bash
# Containers distributed scheduling to thinker.local
CONTAINERS_REMOTE_ENABLED=true             # ❌
CONTAINERS_REMOTE_DEFAULT_SSH_USER=milosvasic  # ❌

CONTAINERS_REMOTE_HOST_1_NAME=thinker       # ❌
CONTAINERS_REMOTE_HOST_1_ADDRESS=thinker.local  # ❌
CONTAINERS_REMOTE_HOST_1_USER=milosvasic    # ❌
CONTAINERS_REMOTE_HOST_1_LABELS=gpu=true,gpu_vendor=nvidia,gpu_model=rtx3060,cuda=12.2,nvenc=true,vulkan=true  # ❌
CONTAINERS_REMOTE_HOST_1_GPU_AUTOPROBE=true # ❌

# OCU CUDA sidecar endpoint (runs the gRPC server — see OCU-CUDA-Sidecar/)
HELIX_OCU_CUDA_ADDR=thinker.local:50060     # ❌
```

## 9. HelixQA runtime

### 🟢 Infra health probes (HelixQA pre-session checks)

```bash
HELIX_INFRA_HOST=localhost                 # ❌ default ok for single-host
HELIX_INFRA_API_PORT=8080                  # ❌
HELIX_INFRA_API_SERVICE=catalog-api        # ❌
HELIX_INFRA_API_HEALTH_PATH=/api/v1/health # ❌
HELIX_WEB_URL=http://localhost:3000        # ❌ catalog-web dev URL
HELIX_QA_CLEAN_ALL_ON_NEW_RUN=true         # ✅ clean qa-results/ on new session
```

### 🟢 ADB & device control

```bash
ANDROID_HOME=/home/milosvasic/Android/Sdk  # ❌ auto-detected from PATH if unset
ANDROID_SDK_ROOT=/home/milosvasic/Android/Sdk # ❌ mirror of above
HELIX_ANDROID_DEVICE=192.168.0.193:5555    # ❌ force a device serial
HELIX_ANDROID_PACKAGE=com.catalogizer.androidtv  # ❌
HELIXQA_ADB_SERIAL=                        # ❌ alternative to HELIX_ANDROID_DEVICE
HELIX_ADB_EXCLUDE=ATMOSphere               # ❌ already enforced via .devignore
HELIX_DESKTOP_DISPLAY=:0                   # ❌ defaults to $DISPLAY
HELIX_FFMPEG_PATH=/home/milosvasic/bin/ffmpeg  # ❌ already defaulted inside scripts
HELIXQA_PLAYWRIGHT_NODE_PATH=              # ❌ rare override
HELIX_ENV_FILE=/run/media/.../Catalogizer/.env  # ❌ forced env path
```

### ⚫ CI / test kill-switches — LEAVE UNSET in production

```bash
HELIXQA_CAPTURE_WEB_STUB=1                 # forces chromedp path to ErrNotWired
HELIXQA_CAPTURE_ANDROID_STUB=1
HELIXQA_CAPTURE_LINUX_STUB=1
HELIXQA_INTERACT_WEB_STUB=1
HELIXQA_INTERACT_ANDROID_STUB=1
HELIXQA_INTERACT_LINUX_STUB=1
HELIXQA_VISION_CPU_STUB=1
HELIXQA_OBSERVE_DBUS_STUB=1
HELIXQA_OBSERVE_CDP_STUB=1
HELIXQA_OBSERVE_AX_STUB=1
HELIXQA_OBSERVE_LDPRELOAD_STUB=1
HELIXQA_RECORD_X264_STUB=1
HELIXQA_RECORD_VAAPI_STUB=1
HELIXQA_RECORD_NVENC_STUB=1
HELIXQA_VAAPI_DEVICE=/dev/dri/renderD128   # override VAAPI device node
```

### 🟢 LD_PRELOAD shim (opt-in observer)

```bash
HELIXQA_LD_SHIM=/tmp/helix-shim.so         # ❌ only when shim is compiled
HELIXQA_LD_SHIM_FIFO=/tmp/helix-shim.fifo  # ❌ named pipe for shim output
```

## 10. Git / release remotes

```bash
# ✅ Already in your .env
GITHUB_TOKEN=<pat>
GITLAB_TOKEN=<pat>
GITFLIC_TOKEN=<token>
GITVERSE_TOKEN=<token>
```

## 11. Observability — OPTIONAL

```bash
SENTRY_DSN=<https://...>                   # ❌ production error reporting
SEMGREP_APP_TOKEN=<semgrep-token>          # ❌ silences hook spam (user shell rc is also fine)
```

## 12. Container build overrides

```bash
CATALOGIZER_BUILDER_IMAGE=localhost/catalogizer-builder:latest  # ❌ default is fine
CATALOGIZER_BUILDER_AUTO=1                 # ❌ 0 disables auto-dispatch (forces local)
```

---

## Minimum to unlock Phase 4+ live cycle

If you only have time to add **one** block, make it this one (§8 — OCU distributed + §7 — TMDB):

```bash
# §7 — real catalogue testing
TMDB_API_KEY=<your-tmdb-v3-key>
TMDB_ACCESS_TOKEN=<your-tmdb-v4-token>
OMDB_API_KEY=<your-omdb-key>
FANART_TV_API_KEY=<your-fanart-key>
IGDB_CLIENT_ID=<your-twitch-igdb-id>
IGDB_CLIENT_SECRET=<your-twitch-igdb-secret>

# §8 — CUDA/NVENC routing to thinker.local
CONTAINERS_REMOTE_ENABLED=true
CONTAINERS_REMOTE_DEFAULT_SSH_USER=milosvasic
CONTAINERS_REMOTE_HOST_1_NAME=thinker
CONTAINERS_REMOTE_HOST_1_ADDRESS=thinker.local
CONTAINERS_REMOTE_HOST_1_USER=milosvasic
CONTAINERS_REMOTE_HOST_1_LABELS=gpu=true,gpu_vendor=nvidia,cuda=12.2,nvenc=true
CONTAINERS_REMOTE_HOST_1_GPU_AUTOPROBE=true
HELIX_OCU_CUDA_ADDR=thinker.local:50060
```

Everything else is tuning / fallback. Gemini (already in `.env`) is sufficient to run Phase 6 autonomous QA.

---

## Per-submodule references (authoritative `.env.example` files)

If you want the exact variable list for one specific component, read its own template:

- `catalog-api/.env.example` — backend
- `catalog-web/.env.example` — frontend
- `HelixQA/.env.example` — QA orchestrator
- `LLMsVerifier/.env.example` — provider scoring
- `LLMProvider/.env.example` — provider-specific runtime
- `LLMOrchestrator/.env.example` — agent runtime
- `VisionEngine/.env.example` — vision providers
- `DocProcessor/.env.example` — document processing
- `Containers/.env.example` — container distribution
- `Challenges/Containers/.env.example` — challenge-test stack
- `OCU-CUDA-Sidecar/.env.example` — GPU sidecar
- `.env.example` (project root) — consolidated template

This document supersedes the per-submodule templates for operator setup. When in doubt, use this file.

---

*Last updated: 2026-04-19. Regenerated by auditing actual `os.Getenv` / `os.LookupEnv` call sites across HelixQA + Containers + LLMsVerifier + LLMProvider + LLMOrchestrator + VisionEngine.*
