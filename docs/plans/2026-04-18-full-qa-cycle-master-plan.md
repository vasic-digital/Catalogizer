# Full-QA Master Cycle — 2026-04-18

**Operator directive:** Rebuild everything clean-slate, run every test + Challenge + HelixQA bank, execute autonomous QA per app/platform (no ATMOSphere), review videos/screenshots, ticket every defect, root-cause fix with Fixes Validation Tests Suite entries, rebuild + retest loop until clean, release artefacts to `releases/<platform>/<app>/<version>/`, version-bump.

**Governance:** `CONSTITUTION.md` Article VII (§7.1–§7.11) is the enforceable contract.

**Mandatory reference:** `docs/OPEN_POINTS_CLOSURE.md` — tick items in the same commit that changes them.

---

## Session context

- **Session directory:** `docs/reports/qa-sessions/2026-04-18-T2158/`
- **Connected ADB devices at start:** 2 devices, both `ro.product.model=ATMOSphere` — **EXCLUDED** per Article VII §7.1 + `.devignore`.
- **`.env` keys available:** ~35 LLM providers (GEMINI, ANTHROPIC-via-OAuth, GROQ, KIMI, MISTRAL, NVIDIA, OPENROUTER, ASTICA, DEEPSEEK, CEREBRAS, COHERE, HUGGINGFACE, MOONSHOT-via-KIMI, REPLICATE, SAMBANOVA, FIREWORKS, HYPERBOLIC, ZHIPU, ZAI, VERTEX, UPSTAGE, NOVITA, PUBLICAI, SARVAM, VULAVULA, CLOUDFLARE, CODESTRAL, SILICONFLOW, VENICE, JUNIE, KILO, LETTA, MEMO, NIA, NLP, MODAL, CHUTES, HUGGINGFACE). Real autonomous QA is UNLOCKED.
- **Service credentials:** POSTGRES_URL, REDIS_URL, CHROMA_URL, QDRANT_URL — local + remote via `SVC_*_REMOTE` flags.
- **Git remotes:** 6 main, 4 HelixQA, 2 Security, 2 Containers.

## Fatal blocker inventory (as of 2026-04-18-T2158)

| Blocker | Scope | Unblocker |
|---|---|---|
| ATMOSphere-only ADB devices | Android + Android TV autonomous QA | Connect a non-ATMOSphere Android phone + Android TV via `adb connect`; add to `.devconnect` |
| No CUDA sidecar running on thinker.local | Vision CUDA path + NVENC encode | Build + deploy `OCU-CUDA-Sidecar/` image |
| No iOS / macOS / Windows runners | iOS autonomous QA + Windows/Mac native capture | E2 operator item |
| No LD_PRELOAD shim compiled per target | Hook-based observation against a specific target binary | Operator picks target + compiles `docs/hooks/ld-preload-shim.c` |

Everything else is in-session executable.

## Execution plan (this session)

### Phase 1 — Governance + plan + session directory [DONE]
- ✅ Constitution Article VII added
- ✅ CLAUDE.md + AGENTS.md reference the article
- ✅ `.devignore` already excludes ATMOSphere (line 14)
- ✅ This plan committed
- ✅ Session report directory created

### Phase 2 — Clean rebuild
Targets: catalog-api, catalog-web, catalogizer-desktop, installer-wizard, catalogizer-android, catalogizer-androidtv, catalogizer-api-client. Runner: `scripts/release-build.sh --container --force --skip-tests` OR `scripts/build-all-releases.sh`.

Log destination: `docs/reports/qa-sessions/2026-04-18-T2158/logs/release-build.log`.

### Phase 3 — Unit + integration tests (every submodule)
Runners:
- `cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1 -race -timeout 600s | tee ../docs/reports/qa-sessions/2026-04-18-T2158/logs/catalog-api-tests.log`
- `cd HelixQA && GOTOOLCHAIN=local go test -mod=vendor -race ./... -count=1 -timeout 600s | tee ../docs/reports/qa-sessions/2026-04-18-T2158/logs/helixqa-tests.log`
- `cd Containers && GOTOOLCHAIN=local go test -race ./... -count=1 -timeout 300s | tee ../docs/reports/qa-sessions/2026-04-18-T2158/logs/containers-tests.log`
- `cd Security && GOTOOLCHAIN=local go test ./... -race -count=1 | tee ../docs/reports/qa-sessions/2026-04-18-T2158/logs/security-tests.log`
- Each of the other Go submodules (Assets, Auth, Cache, Challenges, Concurrency, Config, Database, Discovery, Entities, EventBus, Filesystem, Lazy, Media, Memory, Middleware, Observability, RateLimiter, Recovery, Storage, Streaming, Watcher, DocProcessor, LLMOrchestrator, LLMProvider, VisionEngine, ReplayBuffer, ScreenDiff, TrainingCollector, VisualRegression) — same pattern.
- catalog-web: `npm run test && npm run test:e2e | tee …`
- catalogizer-desktop + installer-wizard: `npm run test`

### Phase 4 — Challenges bank run
Runner: the running catalog-api exposes `/api/v1/challenges`. Run ALL registered challenges via the API (per CLAUDE.md Challenge System: "never use shell scripts, curl, or third-party tools — use the catalog-api binary"). Log to `docs/reports/qa-sessions/2026-04-18-T2158/challenges/`.

### Phase 5 — HelixQA bank tests (non-Android)
- API bank: `./scripts/run-helixqa-api.sh`
- Web bank: `./scripts/run-helixqa-web.sh`
- Desktop bank: `./scripts/run-helixqa-desktop.sh`

Android + Android TV banks **SKIPPED** — see fatal blocker. Recorded as SKIP in the report.

### Phase 6 — HelixQA autonomous QA (non-Android)
- API: autonomous against `catalog-api` service
- Web: autonomous against `catalog-web` + chromedp (Chromium is installed)
- Desktop: autonomous against Tauri binary once built

Android + Android TV autonomous **SKIPPED** — ATMOSphere blocker.

### Phase 7 — Post-session review
Every video + screenshot produced by phases 5–6 reviewed (HelixQA's own visual-analysis pipeline via LLMsVerifier → Gemini/etc.). Tickets for every defect land in `docs/reports/qa-sessions/2026-04-18-T2158/tickets/` with full evidence per Constitution §7.5.

### Phase 8 — Fix loop
Each ticket → root-cause investigation → fix → 4-artefact regression tail (unit test + fixes-validation bank entry + HelixQA bank entry + challenge registration) → rebuild affected binary/container → re-run phases 3–6 for that scope only. Loop until clean pass (§7.3 "NOTHING LEFT") or until a FATAL BLOCKER surfaces.

### Phase 9 — Version bump + release artefacts
On clean pass:
- `scripts/release-build.sh --force` produces debug + release builds
- Copy to `releases/<platform>/<app>/<version>/`
- Bump version codes in `versions.json`
- Update `docs/releases/v<version>.md`
- Commit + push

### Phase 10 — Final session report
`docs/reports/qa-sessions/2026-04-18-T2158/FINAL-REPORT.md` with:
- Aggregated per-app / per-platform / per-category pass/fail/skip
- Deep analysis of gathered data
- Suggestions + conclusions for further improvements and fixes
- Missing-feature list for HelixQA + dependency submodules
- Next-session unblocker asks

## What I will complete in this Claude session

Honest scope: Phases 1–3 fully in-session; Phase 4 as far as the catalog-api binary + local services allow without operator intervention; Phases 5–9 partially (start, not finish). Phase 10 as a stub that a follow-up session can expand.

A genuine clean-pass loop takes multiple hours of wall-clock time per iteration — I cannot complete it inside a single conversation turn. The in-session outputs provide the baseline; subsequent sessions continue the loop per Article VII.

## Extensibility follow-ups (pre-committed to OPEN_POINTS_CLOSURE.md §5)

- Deep web research for cutting-edge OSS QA/automation frameworks to vendor + integrate
- Extend HelixQA's autonomous pipeline with comprehensive data-set combinatorics (positive + faulty + wrong + boundary + i18n + malicious)
- Extend Challenges submodule with a Full QA Tests Suite bank grouping every flow / screen / component / use case / edge case
- Add comprehensive live-monitoring dashboard (platform + app + test ID + progress + result) driven by a tests-executor binary that streams to operator console + archives to session directory
- Missing HelixQA features surfaced during review go into the closure brief and next-wave planning

## Stop condition for this session

Per §7.3: this session is already at a FATAL BLOCKER for Android + Android TV scope. For every other scope, the loop terminates on NOTHING LEFT or a second FATAL BLOCKER (e.g. test infra collapse).

---

*Plan committed 2026-04-18; execution log in the session report directory.*
