---
title: Helix Nexus + Cover-Quality — Research Findings & Phased Execution Plan
date: 2026-04-17
status: research complete, plan drafted
related:
  - docs/nexus/unfinished-and-issues.md
  - docs/plans/2026-04-17-helix-nexus-open-clawed-integration-plan.md
---

# Research Findings & Phased Execution Plan

This document captures the web research for every remaining 🔴 / 🟠 item
in the gap report, names concrete open-source libraries to adopt, and
lays out a fine-grained phased plan that the next sessions execute.
Every candidate library has a known license, maintenance status, and a
concrete integration point in the existing Nexus + catalog-api code.

## Legend

| Symbol | Meaning |
|---|---|
| 🆕 | New Go module dependency added via `go get` (no submodule) |
| 📦 | Added as a Git submodule so we can vendor or extract parts |
| 🔗 | Documentation-only reference (no code pulled) |

---

## 1. Research findings per gap

### 1.1 Live-browser integration harness

**Goal:** prove `pkg/nexus/browser` with the `nexus_chromedp` and
`nexus_rod` build tags works against a real Chromium instance and
recurring CI runs.

**Libraries / tools:**

- 🆕 **`testcontainers/testcontainers-go`** ([github.com/testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go)) —
  spin up a Chromium container from a test, expose the CDP port, and
  wire our Engine to it. Covers WebDriver reference flows without a
  persistent `chromium` install on the runner.
- 🆕 **`chromedp/chromedp`** — already a dep; runner uses this in
  `nexus_chromedp` tag.
- 🆕 **`go-rod/rod`** — already a dep.
- 📦 **`browserbase/stagehand`** — already vendored under
  `HelixQA/tools/opensource/stagehand`. We mine its "act / extract /
  observe" action primitives and self-healing patterns (MIT-licensed).
  Reference for Navigator + Healer prompt shapes.
- 🔗 **`rebrowser.net/blog/chromedp-tutorial-master-browser-automation-in-go-with-real-world-examples-and-best-practices`** —
  anti-detection + reliability tips.

**Integration point:** new `pkg/nexus/browser/integration_test.go`
guarded by `//go:build nexus_chromedp_integration`, using
testcontainers-go to start `zenika/alpine-chrome` or
`browserless/chrome` and asserting the full snapshot → action loop.

### 1.2 Real Appium hub integration

**Goal:** execute one full Android + iOS session against a running
Appium 2.0 hub (simulator and emulator lanes).

**Libraries / tools:**

- **Appium does not ship an official Go client.** The W3C WebDriver
  protocol is language-agnostic so our existing `pkg/nexus/mobile` HTTP
  client is the right direction.
- 🆕 **`sclevine/agouti`** — thin W3C WebDriver wrapper; useful only as
  a reference, since our HTTP client already does exactly what we need.
- 🔗 **`appium.io/docs/en/2.0/intro/`** — W3C endpoint contract.
- 🔗 **`appium/appium-uiautomator2-driver`** and **`appium-xcuitest-driver`** —
  the drivers the hub loads.
- 📦 **`appium/appium-inspector`** — desktop tool. Reference-only
  because we want CI-less headless operation.
- 📦 **`appium/appium`** — vendor-read as a submodule under
  `HelixQA/tools/opensource/appium/` to pull stable fixtures and test
  images; already present.

**Integration point:** `pkg/nexus/mobile/integration_test.go` behind a
`nexus_appium_integration` build tag that boots Appium in a container
(via testcontainers-go), drives UIAutomator2 against an emulator, and
records a session video.

### 1.3 Real Windows / macOS desktop harness

**Goal:** prove `pkg/nexus/desktop/windows.go` against WinAppDriver and
`pkg/nexus/desktop/macos.go` against XCUITest on a Mac runner.

**Libraries / tools:**

- 🔗 **`microsoft/WinAppDriver`** ([github.com/microsoft/WinAppDriver](https://github.com/microsoft/WinAppDriver)) —
  needs a Windows Pro/Enterprise host. Archived but still works.
- 📦 **`licanhua/YWinAppDriver`** ([github.com/licanhua/YWinAppDriver](https://github.com/licanhua/YWinAppDriver)) —
  open-source, actively maintained WinAppDriver-compatible reimplementation
  (MIT license). Candidate for a HelixQA submodule so we can vendor
  binaries and rebuild deterministically.
- 🔗 **`appium/WebDriverAgent`** — iOS + macOS XCUITest agent. Already
  referenced by the iOS runbook; the same binary drives macOS real-app
  automation.

**Integration point:** new `tools/nexus-windows-runner/` containing a
small Go program that talks to YWinAppDriver and exercises
`pkg/nexus/desktop/windows.go`. macOS gets a similar harness under
`tools/nexus-macos-runner/`.

### 1.4 axe-core + Playwright-compatible audit

**Goal:** make the `a11y.Auditor` test reach a live axe-core instance
and emit violations into the orchestrator evidence vault.

**Libraries / tools:**

- 📦 **`dequelabs/axe-core`** — vendor under
  `HelixQA/tools/opensource/axe-core/` so `a11y.InjectionScript()` can
  point at a local path instead of a CDN. License: Mozilla Public
  License 2.0 (compatible; redistribution requires preserving the
  license header).
- 🆕 **`@axe-core/playwright`** patterns — referenced as prompt shape
  for `pkg/nexus/ai/navigator.go` when the Navigator requests an
  a11y sweep.

**Integration point:** a Makefile target
`make vendor-axe-core` that checks the axe-core SHA and copies the
minified bundle into `docs/nexus/vendor/axe.min.js` (already referenced
by `InjectionScript`).

### 1.5 k6 embedded performance runner

**Goal:** let `pkg/nexus/perf` actually run `k6` without shelling out.

**Libraries / tools:**

- 📦 **`grafana/k6`** ([github.com/grafana/k6](https://github.com/grafana/k6)) —
  import the JS runtime package `go.k6.io/k6/cmd` to run scripts
  in-process. License AGPL-3.0; acceptable inside HelixQA because
  the consumer stays internal and we do not distribute a modified
  binary.
- 📦 **`grafana/xk6-browser`** ([github.com/grafana/xk6-browser](https://github.com/grafana/xk6-browser)) —
  same license, provides the browser module. Vendor under
  `HelixQA/tools/opensource/xk6-browser/` if we want a bundled
  `k6 browser` experience.
- 🆕 **`go.k6.io/k6`** — direct import if AGPL vendor constraints are
  acceptable; otherwise keep shelling out.

**Integration point:** `pkg/nexus/perf/run.go` exposes a
`Run(Scenario) (*Metrics, error)` that either invokes the embedded
runtime (AGPL path) or shells out to an installed `k6` binary.

### 1.6 Concrete LLMOrchestrator adapter

**Goal:** satisfy `ai.LLMClient` with a real bridge to our existing
`LLMOrchestrator` submodule instead of the generic OpenAI-compat
HTTPLLMClient.

**Libraries / tools:**

- 📦 **`HelixDevelopment/LLMOrchestrator`** — already a submodule.
  The bridge is a single Go file in `HelixQA/pkg/nexus/ai/` that
  imports the orchestrator's Go types and implements `Chat`.
- 🆕 **`openai/openai-go`** — optional direct SDK if we want richer
  features later.
- 🆕 **`sashabaranov/go-openai`** — well-maintained third-party Go
  SDK; useful for the generic fallback.
- 🔗 Anthropic provides an official Go SDK at
  `github.com/anthropics/anthropic-sdk-go` but we keep provider
  choice at the orchestrator layer.

**Integration point:** `HelixQA/pkg/nexus/ai/orchestrator_client.go`
importing `digital.vasic.llmorchestrator/pkg/...` and satisfying
`ai.LLMClient`.

### 1.7 Real OpenTelemetry exporter

**Goal:** swap `observability.NoopTracer` for an OTel exporter so spans
reach Jaeger / Tempo / OpenSearch.

**Libraries / tools:**

- 🆕 **`go.opentelemetry.io/otel`** — the core SDK.
- 🆕 **`go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`** —
  OTLP gRPC exporter (Jaeger accepts OTLP since v1.35).
- 🆕 **`go.opentelemetry.io/otel/exporters/prometheus`** — metrics
  exporter for the `helix_nexus_*` series.
- 🔗 **`opentelemetry.io/docs/languages/go/exporters/`** — official
  configuration guide.

**Integration point:** `HelixQA/pkg/nexus/observability/otel.go`
building an `OTelTracer` that implements the existing `Tracer`
interface.

### 1.8 SSO production-grade implementation

**Goal:** replace the `OIDCProvider` / `SAMLProvider` "verifier
function" shims with concrete implementations.

**Libraries / tools:**

- 🆕 **`coreos/go-oidc`** ([github.com/coreos/go-oidc](https://github.com/coreos/go-oidc)) —
  v3, actively maintained, integrates with `golang.org/x/oauth2`.
- 🆕 **`zitadel/oidc`** — OpenID-Foundation-certified alternative.
- 🆕 **`crewjam/saml`** ([github.com/crewjam/saml](https://github.com/crewjam/saml)) —
  mature SP + IdP implementation; `samlsp` package gives us the
  middleware out of the box.
- 🆕 **`russellhaering/gosaml2`** — lighter SP-only alternative.

**Integration point:** `HelixQA/pkg/nexus/orchestrator/sso_concrete.go`
holds the glue that turns a `crewjam/saml.ServiceProvider` into a
`SAMLProvider.Verifier` function.

### 1.9 Prometheus + metrics collector

**Goal:** serve `observability.Registry.Expose()` through an actual
`/metrics` HTTP handler compatible with Prometheus scraping.

**Libraries / tools:**

- 🆕 **`prometheus/client_golang`** ([github.com/prometheus/client_golang](https://github.com/prometheus/client_golang)) —
  canonical Go metrics + HTTP handler.
- 🆕 **`shirou/gopsutil/v4`** — host stats for process memory / CPU
  panels on the Grafana dashboard.

**Integration point:** `HelixQA/pkg/nexus/observability/prometheus.go`
adapts our `Registry` to `prometheus.Collector` so we get native
scraping for free.

### 1.10 Linux accessibility via AT-SPI DBus

**Goal:** replace the shell-script `atspi-helpers` with native Go
bindings for lower latency + better error surfacing.

**Libraries / tools:**

- 🆕 **`godbus/dbus/v5`** ([github.com/godbus/dbus](https://github.com/godbus/dbus)) —
  native D-Bus bindings.
- 🔗 **`wiki.linuxfoundation.org/accessibility/d-bus`** — AT-SPI
  D-Bus protocol spec.

**Integration point:** `HelixQA/pkg/nexus/desktop/linux_atspi.go`
using `godbus` to implement `atspi-find` / `atspi-action` /
`atspi-type` in-process. Shell helpers stay as the portable fallback.

### 1.11 Fanart.tv / IGDB / Cover Art Archive resolvers

**Goal:** add provider resolvers for richer high-res cover candidates.

**Libraries / tools:**

- 🆕 **`odwrtw/fanarttv`** ([github.com/odwrtw/fanarttv](https://github.com/odwrtw/fanarttv)) —
  Go wrapper for Fanart.tv's v3 API.
- 🆕 **`Henry-Sarabia/igdb`** ([github.com/Henry-Sarabia/igdb](https://github.com/Henry-Sarabia/igdb)) —
  IGDB v4 Go client (requires Twitch client id / secret).
- 🆕 **`gopkg.in/mineo/gocaa.v1`** — Cover Art Archive Go client for
  MusicBrainz cover retrieval.

**Integration point:** new files under
`catalog-api/internal/services/` implementing
`resolver.Resolver` for each provider, wrapped by `QualityGate` in
`main.go`.

### 1.12 HTML sanitization for user-provided cover / provider responses

**Goal:** defensive layer so a malicious cover-art URL response
cannot inject HTML/JS into any downstream admin UI.

**Libraries / tools:**

- 🆕 **`microcosm-cc/bluemonday`** ([github.com/microcosm-cc/bluemonday](https://github.com/microcosm-cc/bluemonday)) —
  OWASP-inspired Go sanitizer with `UGCPolicy` + `StrictPolicy`.

**Integration point:** `catalog-api/internal/services/quality_gate.go`
runs bluemonday on any text metadata before persisting.

### 1.13 Runtime wiring

**Goal:** make the nexus packages reachable from the `helixqa` binary
and the Catalogizer challenges.

**Integration points:**

- `HelixQA/cmd/helixqa/` grows a `--nexus` flag that selects the new
  adapters.
- `catalog-api/challenges/` registers challenges constructed around
  the `pkg/nexus/userflow.NexusBrowserAdapter` so one end-to-end run
  exercises the stack.
- `scripts/helixqa-orchestrator.sh` learns a `--nexus` mode.

---

## 2. Open-source codebases to add as Git submodules

All additions go under `HelixQA/tools/opensource/` unless noted.

| Submodule | Repository | License | Why |
|---|---|---|---|
| `licanhua/YWinAppDriver` | [github.com/licanhua/YWinAppDriver](https://github.com/licanhua/YWinAppDriver) | MIT | Open-source WinAppDriver replacement; keep the server source vendored so our Windows lane is reproducible. |
| `dequelabs/axe-core` | [github.com/dequelabs/axe-core](https://github.com/dequelabs/axe-core) | MPL 2.0 | Vendor axe-core so the a11y.InjectionScript points at a local asset, no CDN. |
| `grafana/xk6-browser` | [github.com/grafana/xk6-browser](https://github.com/grafana/xk6-browser) | AGPL-3.0 | Reference + optional embed for browser-level performance tests. |
| `grafana/k6` | [github.com/grafana/k6](https://github.com/grafana/k6) | AGPL-3.0 | Reference and optional embed of the k6 JS runtime. |
| `appium/WebDriverAgent` | [github.com/appium/WebDriverAgent](https://github.com/appium/WebDriverAgent) | Apache-2.0 | iOS + macOS driver source, needed for CI-less real-device builds. |
| `godbus/dbus` | [github.com/godbus/dbus](https://github.com/godbus/dbus) | BSD-2-Clause | Optional — usually pulled as a Go module but the repo makes a good reference for AT-SPI integration. |

Go modules (no submodule needed; `go get` adds them):

- `github.com/testcontainers/testcontainers-go`
- `github.com/coreos/go-oidc/v3`
- `github.com/crewjam/saml`
- `github.com/prometheus/client_golang`
- `github.com/shirou/gopsutil/v4`
- `go.opentelemetry.io/otel` + OTLP exporters + Prometheus exporter
- `github.com/minio/minio-go/v7` (concrete S3Client implementation for our `S3EvidenceStore`)
- `github.com/godbus/dbus/v5`
- `github.com/microcosm-cc/bluemonday`
- `github.com/odwrtw/fanarttv`
- `github.com/Henry-Sarabia/igdb`
- `gopkg.in/mineo/gocaa.v1`
- `github.com/sashabaranov/go-openai`

---

## 3. Phased execution plan

Every phase ends with a commit + push to every remote (HelixQA × 4,
main × 6) so no work stalls. Unit + integration tests must stay green
on every merge. Tasks are sized in developer-hours as a rough guide;
the actual execution may batch several into a single session.

### Phase R-0 — Dependency landing (~4 h)

| # | Task | Details |
|---|---|---|
| R0-01 | `go get` for every 🆕 module | Commits `go.mod` + `go.sum`. |
| R0-02 | Add 📦 submodules via `./scripts/setup-submodule.sh` | YWinAppDriver, dequelabs/axe-core, xk6-browser, k6, WebDriverAgent. |
| R0-03 | Makefile `vendor-axe-core` + `vendor-ywinappdriver` | Deterministic copy into `docs/nexus/vendor/` or a release artefact dir. |
| R0-04 | CI-less contract test: every submodule's SHA logged | Add a `scripts/check-nexus-submodules.sh` that `git diff --submodule=log` and fails on drift. |

### Phase R-1 — Observability pipeline (~6 h)

| # | Task |
|---|---|
| R1-01 | `HelixQA/pkg/nexus/observability/otel.go` — OTelTracer implementing `Tracer`; ingest via OTLP gRPC. |
| R1-02 | `HelixQA/pkg/nexus/observability/prometheus.go` — adapt our `Registry` into `prometheus.Collector`; expose `/metrics` HTTP handler. |
| R1-03 | `HelixQA/cmd/helixqa-metrics/` — tiny binary that runs the Prometheus handler on a configurable port. |
| R1-04 | Grafana dashboard operator guide under `docs/nexus/operator-manual/observability.md`. |
| R1-05 | 6+ tests per file (happy path, nil exporter fallback, /metrics text format). |

### Phase R-2 — SSO hardening (~5 h)

| # | Task |
|---|---|
| R2-01 | `HelixQA/pkg/nexus/orchestrator/oidc_concrete.go` — coreos/go-oidc-backed verifier. |
| R2-02 | `HelixQA/pkg/nexus/orchestrator/saml_concrete.go` — crewjam/saml SP wrapped into the `Verifier` signature. |
| R2-03 | Integration test with a fake IdP (an in-process SAML IdP from `samlidp`). |
| R2-04 | Operator runbook `docs/nexus/runbooks/sso-oidc-okta.md` + `sso-saml-azuread.md`. |

### Phase R-3 — LLMOrchestrator bridge (~4 h)

| # | Task |
|---|---|
| R3-01 | `HelixQA/pkg/nexus/ai/orchestrator_client.go` — wraps `LLMOrchestrator` Go types into `ai.LLMClient`. |
| R3-02 | Cost telemetry wired to Prometheus (`helix_nexus_ai_cost_cents`). |
| R3-03 | Integration test using the orchestrator's in-process fake. |
| R3-04 | Docs update `docs/nexus/ai.md` ("Concrete orchestrator bridge"). |

### Phase R-4 — Cover provider resolvers (~8 h)

| # | Task |
|---|---|
| R4-01 | `catalog-api/internal/services/fanart_resolver.go` — odwrtw/fanarttv-backed resolver. |
| R4-02 | `catalog-api/internal/services/igdb_resolver.go` — Henry-Sarabia/igdb-backed resolver with Twitch auth. |
| R4-03 | `catalog-api/internal/services/cover_art_archive_resolver.go` — mineo/gocaa-backed. |
| R4-04 | Each resolver gated by `QualityGate`; priorities 11/21/22 as per the design spec. |
| R4-05 | Per-resolver unit tests with httptest.Server mocks. |
| R4-06 | Registration in `main.go`'s resolver chain. |

### Phase R-5 — Linux AT-SPI native bindings (~5 h)

| # | Task |
|---|---|
| R5-01 | `HelixQA/pkg/nexus/desktop/linux_atspi.go` — godbus-backed `FindByName`, `FindByRole`, `Click`, `Type`. |
| R5-02 | Feature flag `NEXUS_LINUX_ATSPI_NATIVE=1` selects the native path; shell helpers remain the default. |
| R5-03 | Tests using a DBus test double (the `godbus/dbus/v5` test harness). |

### Phase R-6 — Runtime wiring (~7 h)

| # | Task |
|---|---|
| R6-01 | `HelixQA/cmd/helixqa/` grows a `--nexus` flag. |
| R6-02 | `catalog-api/challenges/nexus_web_flow.go` — a new challenge that spins up `NexusBrowserAdapter` and drives the Catalogizer web UI end-to-end. |
| R6-03 | `scripts/helixqa-orchestrator.sh` learns the new flag. |
| R6-04 | The cover-quality gate emits metrics through Nexus's `NexusMetrics`. |

### Phase R-7 — Integration harnesses (~8 h)

| # | Task |
|---|---|
| R7-01 | `HelixQA/pkg/nexus/browser/integration_test.go` under `//go:build nexus_chromedp_integration` using testcontainers-go + zenika/alpine-chrome. |
| R7-02 | `HelixQA/pkg/nexus/mobile/integration_test.go` under `//go:build nexus_appium_integration`; boots Appium + UIAutomator2 in a container and drives the Catalogizer Android TV APK. |
| R7-03 | `tools/nexus-windows-runner/` — manual runbook + helper binary for Windows Pro workstations. |
| R7-04 | `tools/nexus-macos-runner/` — manual runbook for macOS hosts. |

### Phase R-8 — Real QA campaign (~6 h wall-clock)

| # | Task |
|---|---|
| R8-01 | Build every component via `./scripts/release-build.sh --container --force`. |
| R8-02 | `./scripts/services-up.sh` spins up postgres, redis, catalog-api, Prometheus, Grafana, Jaeger/Tempo, MinIO. |
| R8-03 | `./scripts/helixqa-orchestrator.sh --nexus` runs every bank against the new stack. |
| R8-04 | Archive session under `docs/reports/qa-sessions/qa-session-YYYY-MM-DD/`. |
| R8-05 | Feed findings into a net-new `fixes-validation-nexus` bank. |

### Phase R-9 — Content + public website (~10 h)

| # | Task |
|---|---|
| R9-01 | Record video modules 01–08 (external content-guild task; scripts already shipped). |
| R9-02 | Build VitePress site and publish to helixqa.vasic.digital/nexus. |
| R9-03 | Cross-link from main Catalogizer docs + README. |

### Phase R-10 — Polish (~4 h)

| # | Task |
|---|---|
| R10-01 | Fix HTML entity decoding in `pkg/nexus/browser/snapshot.go`. |
| R10-02 | Case-insensitive host allowlist opt-in flag. |
| R10-03 | Improve Generator YAML validator (step-shape checks). |
| R10-04 | Add AUC-verified Predictor training data + regression test. |
| R10-05 | Admin endpoint `/api/v1/admin/image-quality/revalidate` to trigger QualityRevalidator manually. |
| R10-06 | Web / Android / desktop clients surface `X-Cover-Quality` in debug UX. |

---

## 4. Risk register (net-new items)

| Risk | Severity | Mitigation |
|---|---|---|
| AGPL k6 embed contaminates our binary distribution | 🟠 | Keep k6 as a shelled-out process; embed only inside HelixQA internal test binaries. |
| Fanart.tv / IGDB API quotas in CI | 🟠 | Cache responses in `catalog-api/cache/providers/` + skip tests that require live API when env keys are missing. |
| Twitch OAuth refresh for IGDB | 🟡 | Pre-refresh on scheduled cron; persist tokens to a read-only operator vault. |
| YWinAppDriver binary distribution | 🟡 | Rebuild from source on each release; cache binaries under `build/windows/`. |
| axe-core MPL attribution | 🟡 | Ensure `docs/nexus/vendor/axe.min.js` header preserves the Deque copyright. |
| crewjam/saml SP metadata rotation | 🟡 | Ship a helper CLI that generates + rotates SP keys every 90 days. |
| OTel OTLP endpoint availability | 🟡 | Default to NoopTracer so Nexus works without a collector. |

---

## 5. Success metrics (for the next milestone)

- `go test ./... -tags=nexus_chromedp_integration` passes on a CI-less
  workstation using testcontainers-go.
- A single `helixqa-orchestrator.sh --nexus` run produces green
  `FINAL-REPORT.md` under `docs/reports/qa-sessions/`.
- Grafana dashboard shows live `helix_nexus_*` panels during a run.
- `pkg/nexus/ai.HTTPLLMClient` is wrapped by `orchestrator_client.go`
  that calls `LLMOrchestrator` in-process.
- The cover-quality gate has three new provider resolvers (Fanart.tv,
  IGDB, Cover Art Archive).
- OIDC + SAML test accounts resolve to typed `User` rows with matching
  `AuditLog` entries persisted in SQL.

---

## 6. Documentation surfaces updated by this plan

| Surface | Additions |
|---|---|
| `docs/nexus/operator-manual/` | observability, SSO (OIDC + SAML), evidence vault (MinIO), QA orchestrator flags |
| `docs/nexus/runbooks/` | `sso-oidc-okta.md`, `sso-saml-azuread.md`, `prometheus-grafana-setup.md` |
| `docs/nexus/video-course/` | Supplementary module 09 "Running the real QA campaign" |
| `docs/nexus/api/` | Autogenerated Go docstrings for the new packages |
| `docs/nexus/sql/` | Migration scripts needed by the new provider resolvers (no schema change expected; cover_art table already covers it) |
| Website `/nexus/` | Landing updates listing newly integrated libraries |
| CHANGELOG | One entry per R-phase |

---

## 7. Sources

Core research sources consulted for this plan. Every link leads to the
upstream repository or canonical documentation.

- [testcontainers-go](https://github.com/testcontainers/testcontainers-go)
- [chromedp tutorial (Rebrowser)](https://rebrowser.net/blog/chromedp-tutorial-master-browser-automation-in-go-with-real-world-examples-and-best-practices)
- [Appium 2.0 Documentation](https://appium.io/docs/en/2.0/intro/)
- [WinAppDriver (Microsoft)](https://github.com/microsoft/WinAppDriver)
- [YWinAppDriver (licanhua)](https://github.com/licanhua/YWinAppDriver)
- [Fanart.tv Go client (odwrtw/fanarttv)](https://github.com/odwrtw/fanarttv)
- [IGDB Go client (Henry-Sarabia/igdb)](https://github.com/Henry-Sarabia/igdb)
- [Cover Art Archive Go client (gocaa)](https://pkg.go.dev/gopkg.in/mineo/gocaa.v1)
- [OpenTelemetry Go exporters](https://opentelemetry.io/docs/languages/go/exporters/)
- [coreos/go-oidc](https://github.com/coreos/go-oidc)
- [crewjam/saml](https://github.com/crewjam/saml)
- [xk6-browser](https://github.com/grafana/xk6-browser)
- [grafana/k6](https://github.com/grafana/k6)
- [godbus/dbus](https://github.com/godbus/dbus)
- [prometheus/client_golang](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus)
- [bluemonday](https://github.com/microcosm-cc/bluemonday)
- [Stagehand (browserbase)](https://github.com/browserbase/stagehand)
- [Stagehand v3 launch notes](https://www.browserbase.com/blog/stagehand-v3)
- [dequelabs/axe-core](https://github.com/dequelabs/axe-core)
- [MinIO Go SDK](https://github.com/minio/minio-go)
- [AT-SPI D-Bus protocol wiki](https://wiki.linuxfoundation.org/accessibility/d-bus)
