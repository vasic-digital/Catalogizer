---
title: HelixQA Production-Readiness Validation Plan
date: 2026-04-17
status: active — required to pass before HelixQA is allowed in production
owners: [HelixQA core, SRE, Compliance]
---

# HelixQA Production-Readiness Validation Plan

HelixQA is production-ready only when **every row in this matrix
passes**. This document is the authoritative gate for declaring
HelixQA shippable. It references the existing 10-category coverage
bar from CLAUDE.md (Article V) and extends it with HelixQA-specific
functional domains.

Execute the plan top-down: each row ships its own tests + one or more
Challenges registered in `digital.vasic.challenges`, plus one bank
entry under `HelixQA/banks/` so the autonomous orchestrator validates
the row on every campaign.

## Legend

- **Test types** every row must cover: `U`nit, `I`ntegration, `E`2E,
  `F`ull-automation, `S`tress, `SEC`urity, `DDOS`, `B`enchmark,
  `C`hallenges, `H`elixQA-bank.
- Owner per row. Every row has a "Validation bank" identifier so the
  Fixes-Validation bank can flag regressions.

---

## 1. Browser engine (`pkg/nexus/browser`)

| # | Scenario | U | I | E | F | S | SEC | DDOS | B | C | H | Validation bank |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| B-001 | Engine open + close, no leaks under 500 iterations | ✅ | ✅ | ✅ | ✅ | ✅ | — | — | ✅ | CH-NX-BROWSER-001 | nexus-browser.NX-BROWSER-001 | fixes-validation-browser |
| B-002 | Snapshot produces stable e1..eN refs in document order | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | ✅ | CH-NX-BROWSER-002 | NX-BROWSER-002 | " |
| B-003 | URL allowlist blocks `file://`, `javascript:`, `data:`, `vbscript:` and private IPs | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | — | CH-NX-BROWSER-003 | NX-BROWSER-003 | " |
| B-004 | Pool caps concurrency + respects `ctx.Done()` | ✅ | ✅ | — | ✅ | ✅ | — | ✅ | ✅ | CH-NX-BROWSER-005 | NX-BROWSER-005 | " |
| B-005 | `ExtendedHandle` hover/drag/select/wait_for/tab_open/tab_close/pdf/console against real Chromium | — | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | CH-NX-BROWSER-EXT-001..008 | nexus-browser-extended | " |
| B-006 | `InstrumentedEngine` emits all `helix_nexus_*` counters + histograms | ✅ | ✅ | ✅ | ✅ | — | — | — | — | CH-NX-OBS-001 | nexus-observability | " |
| B-007 | Case-insensitive allowlist opt-in behaves identically across casings | ✅ | ✅ | — | — | — | ✅ | — | — | CH-NX-BROWSER-014 | nexus-browser | " |

## 2. Mobile engine (`pkg/nexus/mobile`)

| # | Scenario | U | I | E | F | S | SEC | DDOS | B | C | H | Validation bank |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| M-001 | Appium session lifecycle (Android + iOS simulator + Android TV) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | CH-NX-MOBILE-A-001 / I-001 / TV-001 | nexus-mobile-android / ios / androidtv | fixes-validation-mobile |
| M-002 | Capability validator rejects missing fields per platform | ✅ | — | — | — | — | ✅ | — | — | CH-NX-MOBILE-A-015 | NX-MOBILE-ANDROID-015 | " |
| M-003 | Gestures: tap, longPress, swipe, scroll, pinch, rotate, key all reach the driver | ✅ | ✅ | ✅ | ✅ | — | — | — | ✅ | CH-NX-MOBILE-A-005..011 | nexus-mobile-android | " |
| M-004 | Accessibility tree parser handles Android UIA + iOS XCUITest hierarchies | ✅ | ✅ | ✅ | ✅ | — | — | — | ✅ | CH-NX-MOBILE-A-010 | " | " |
| M-005 | Screen recording start+stop, base64 decode, non-empty MP4 | ✅ | ✅ | ✅ | ✅ | — | — | — | — | CH-NX-MOBILE-A-006 | " | " |
| M-006 | Deep-link navigation (catalogizer://media/:id) lands on correct activity | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-MOBILE-A-004 / I-007 | " | " |
| M-007 | iOS real-device lane via WebDriverAgent URL reuse | — | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-MOBILE-I-013 | nexus-mobile-ios | " |

## 3. Desktop engine (`pkg/nexus/desktop`)

| # | Scenario | U | I | E | F | S | SEC | DDOS | B | C | H | Validation bank |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| D-001 | Windows WinAppDriver session + click + type + shortcut + screenshot | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | CH-NX-DESKTOP-WIN-001..004 | nexus-desktop-windows | fixes-validation-desktop |
| D-002 | macOS AppleScript menu pick + modifier mapping + escape safety | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-DESKTOP-MAC-001..004 | nexus-desktop-macos | " |
| D-003 | Linux AT-SPI native backend replaces shell helpers without behaviour change | ✅ | ✅ | ✅ | ✅ | — | — | — | ✅ | CH-NX-DESKTOP-LIN-003 | nexus-desktop-linux | " |
| D-004 | Wayland `AsWayland()` refuses xdotool fallback + uses wtype for shortcuts | ✅ | ✅ | — | — | — | ✅ | — | — | CH-NX-DESKTOP-LIN-004..005 | " | " |
| D-005 | Installer flows (MSI, DMG, AppImage) complete end-to-end | — | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-DESKTOP-WIN-002, MAC-007, LIN-011 | " | " |

## 4. AI navigation + self-healing (`pkg/nexus/ai`)

| # | Scenario | U | I | E | F | S | SEC | DDOS | B | C | H | Validation bank |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| A-001 | Navigator parses JSON action response (incl. fenced) | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | ✅ | CH-NX-AI-001..002 | nexus-ai.NX-AI-001..002 | fixes-validation-ai |
| A-002 | CostTracker aborts on budget breach across every capability | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-AI-003 | NX-AI-003 | " |
| A-003 | Healer recovers from selector drift + refuses unsafe action | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-AI-004..005 | NX-AI-004..005 | " |
| A-004 | Generator validates YAML shape (step name/action/expected) + id prefix | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-AI-006..007 | NX-AI-006..007 | " |
| A-005 | Predictor ROC AUC ≥ 0.75 on historical holdout | ✅ | ✅ | — | ✅ | — | — | — | ✅ | CH-NX-AI-008..010 | NX-AI-008..010 | " |
| A-006 | HTTPLLMClient + OrchestratorClient roundtrip (incl. multi-modal image path) | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | ✅ | CH-NX-AI-011 | — | " |

## 5. Accessibility (`pkg/nexus/a11y`)

| # | Scenario | U | I | E | F | S | SEC | DDOS | B | C | H | Validation bank |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| X-001 | axe-core injection + parse + Assert(LevelA/AA/AAA) | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-A11Y-001..005 | nexus-a11y.NX-A11Y-001..005 | fixes-validation-a11y |
| X-002 | Section 508 filter returns correct subset | ✅ | ✅ | — | — | — | ✅ | — | — | CH-NX-A11Y-006 | NX-A11Y-006 | " |
| X-003 | Auditor drives real browser Engine + yields non-empty report on live UI | — | ✅ | ✅ | ✅ | — | — | — | ✅ | CH-NX-A11Y-LIVE-001 | nexus-a11y-live | " |
| X-004 | CI breaks when Catalogizer web hits a critical / serious WCAG violation | — | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-A11Y-REG-001 | " | " |

## 6. Performance (`pkg/nexus/perf`)

| # | Scenario | U | I | E | F | S | SEC | DDOS | B | C | H | Validation bank |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| P-001 | GenerateScript + ParseK6JSON + Metrics.Assert cover every threshold | ✅ | ✅ | ✅ | ✅ | — | — | — | ✅ | CH-NX-PERF-001..007 | nexus-perf.NX-PERF-* | fixes-validation-perf |
| P-002 | K6Runner.RunScenario succeeds against a live k6 binary + fails actionably when missing | ✅ | ✅ | ✅ | ✅ | ✅ | — | — | ✅ | CH-NX-PERF-RUN-001 | " | " |
| P-003 | Core Web Vitals regression detector fails build when LCP p95 > baseline + 10% | — | ✅ | ✅ | ✅ | ✅ | — | — | ✅ | CH-NX-PERF-REG-001 | " | " |

## 7. Cross-platform orchestration (`pkg/nexus/orchestrator`)

| # | Scenario | U | I | E | F | S | SEC | DDOS | B | C | H | Validation bank |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| O-001 | Flow executes every step in order; first error aborts; verify can fail the step | ✅ | ✅ | ✅ | ✅ | — | — | — | — | CH-NX-XFLOW-001..003 | nexus-xflow | fixes-validation-xflow |
| O-002 | ExecutionContext shares state across steps + Snapshot returns a copy | ✅ | ✅ | — | — | — | ✅ | — | — | CH-NX-XFLOW-004..005 | " | " |
| O-003 | File + S3 (MinIO) evidence stores Put/PutStream/List round-trip | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-XFLOW-006..007 | " | " |
| O-004 | RBAC enforces role ladder + every decision lands in the audit log | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | — | CH-NX-XFLOW-008..010 | " | " |
| O-005 | AuditPersister + FlowPersister write to helixqa_audit_log / helixqa_cross_flows | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — | — | CH-NX-XFLOW-PERSIST-001 | nexus-xflow-persist | " |
| O-006 | SSO OIDC + SAML verifiers round-trip via real (test) IdP | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | — | CH-NX-XFLOW-SSO-001 | nexus-sso | " |

## 8. Observability (`pkg/nexus/observability`)

| # | Scenario | U | I | E | F | S | SEC | DDOS | B | C | H | Validation bank |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Z-001 | InMemoryTracer captures Start/End + attributes + events + errors | ✅ | ✅ | — | — | — | — | — | — | CH-NX-OBS-001..003 | nexus-observability | fixes-validation-obs |
| Z-002 | OTelTracer delivers spans to OTLP collector | — | ✅ | ✅ | ✅ | — | ✅ | ✅ | ✅ | CH-NX-OBS-OTEL-001 | " | " |
| Z-003 | PrometheusBridge emits every `helix_nexus_*` metric the dashboard references | ✅ | ✅ | ✅ | ✅ | — | — | — | — | CH-NX-OBS-008 | " | " |
| Z-004 | `/metrics` endpoint scrape succeeds + passes `promtool check metrics` | — | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — | CH-NX-OBS-SCRAPE-001 | " | " |
| Z-005 | Grafana dashboard JSON imports cleanly and renders all panels | — | ✅ | ✅ | — | — | — | — | — | CH-NX-OBS-GRAFANA-001 | " | " |

## 9. Runtime campaign (the end-to-end bar)

The single test every deployment must pass before promotion:

`./scripts/helixqa-orchestrator.sh --nexus` against a fresh container
build (PostgreSQL + Redis + catalog-api + Prometheus + Grafana +
Jaeger + MinIO). The campaign **must**:

1. Execute every bank above without a single `FAIL` line in the
   `FINAL-REPORT.md`.
2. Produce a Grafana snapshot that shows non-zero values on every
   dashboard panel.
3. Surface zero Prometheus `ALERTMANAGER_CRITICAL` alerts during the
   run.
4. Close every ticket spawned by the run with an evidence-backed
   reproducer + regression test merged back into the relevant
   `fixes-validation-*` bank.
5. Complete within 2 hours wall-clock on the reference hardware
   budget (4 CPU, 8 GB RAM).

A campaign that reports "PASS" without satisfying all five points is
a critical infrastructure failure and voids the deployment.

---

## 10. Fixes-validation banks (regression firewall)

Every bug fix that lands after this document is written **must**
ship with a test in the matching `fixes-validation-*` bank:

- `banks/fixes-validation-browser.{yaml,json}` (B1 / B2 / B3 / B4 style)
- `banks/fixes-validation-mobile.{yaml,json}`
- `banks/fixes-validation-desktop.{yaml,json}`
- `banks/fixes-validation-ai.{yaml,json}`
- `banks/fixes-validation-a11y.{yaml,json}`
- `banks/fixes-validation-perf.{yaml,json}`
- `banks/fixes-validation-xflow.{yaml,json}`
- `banks/fixes-validation-obs.{yaml,json}`
- `banks/fixes-validation-cover.{yaml,json}` — cover-quality gate,
  provider resolvers, sanitizer.

A PR that introduces a bug fix without an entry in the matching bank
is rejected at review time.

## 11. Acceptance checklist for "production ready"

- [ ] Every row in sections 1–8 passes unit + integration + E2E +
      full-automation + challenges + HelixQA-bank.
- [ ] Stress, security, DDoS, and benchmark columns pass where the
      matrix marks them.
- [ ] Section 9 runtime campaign has passed twice in a row (flake
      protection).
- [ ] `/metrics` scrape green in production Prometheus.
- [ ] OTel spans visible in production Jaeger.
- [ ] Evidence vault contains artefacts from every test run.
- [ ] No `🔴` or `🟠` rows in `docs/nexus/remaining-work.md`.
- [ ] Compliance report signed off by the compliance owner.

Until every box is checked, HelixQA is "library-complete" but **not**
"production-ready". This document is updated in the same commit that
closes a box so the state stays honest.

## 12. Execution order for the next N sessions

1. **Close the runtime wiring items W1–W10** from remaining-work.md
   (W1 + W4 already done; W2, W3, W5, W7, W10 pending). Each closure
   ships with the bank entries above.
2. **Run section 9 campaign** once. Triage every failure, fix the
   root cause, add regression tests to `fixes-validation-*`.
3. **Re-run section 9.** Repeat until two consecutive passes.
4. **Tick the acceptance checklist.**
5. **Announce production readiness** — not before.

No shortcuts. No "it works on my machine". No "green in dev is good
enough". A QA platform that cannot verify itself cannot verify anything.
