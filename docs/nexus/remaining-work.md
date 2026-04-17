---
title: Helix Nexus + Cover-Quality — Remaining Work & Known Issues
date: 2026-04-17
status: living document (update after every closure)
supersedes: docs/nexus/unfinished-and-issues.md (which is now summarised here)
---

# Remaining Work & Known Issues

This is the single authoritative list of what is still open across the
Helix Nexus + cover-quality program after the current session. Update
this file in the same commit that closes any item so the list stays
honest.

Severity legend:

- 🔴 **Blocker** — cannot ship the feature without closing.
- 🟠 **Significant** — library works but is not reachable.
- 🟡 **Polish** — nice to have; does not gate the release.
- ⚫ **External** — depends on hardware, a human, or a product decision.

---

## 🔴 Wiring gaps — code exists, nothing calls it

These are the highest-ROI items. Each one is a one-to-two line change
in `main.go` or a similar entry point. Until they land, the library
code they cover is effectively dead weight at runtime.

| # | Wiring needed | Lives in | Evidence it's missing |
|---|---|---|---|
| W1 | `router.Use(middleware.SecurityHeaders(cfg))` | `catalog-api/main.go` before route registration | No caller anywhere in catalog-api |
| W2 | `csrf := middleware.NewCSRF(secret); admin.Use(csrf.Handler())` | `catalog-api/main.go` admin group | Admin routes accept cross-origin `POST` today |
| W3 | `title = services.SanitizeMetadataString(raw)` inside Fanart / IGDB / CAA resolvers | `catalog-api/internal/services/*_resolver.go` | Provider strings persisted verbatim |
| W4 | `router.GET("/metrics", gin.WrapH(observability.Handler(registry)))` | `catalog-api/main.go` | Grafana dashboard reads zero data |
| W5 | `eng = browser.Instrument(base, metrics)` in every browser Engine constructor | `HelixQA/cmd/` or the helixqa runtime | NexusMetrics counters never increment |
| W6 | Navigator / Healer / Generator receive `OrchestratorClient` | `HelixQA/pkg/nexus/ai/navigator.go` callers | Defaults fall to `HTTPLLMClient` or nil |
| W7 | `persister.Save(ctx, entry)` attached to `AccessControl.Check` | `HelixQA/pkg/nexus/orchestrator/rbac.go` | Audit trail stays in RAM, `helixqa_audit_log` is empty |
| W8 | `k6.RunScenario` invoked from `perf` challenges | `HelixQA/banks/` or a new perf challenge | Generated scripts sit unrun |
| W9 | `Auditor.RunAndAssert` invoked from web banks | `HelixQA/banks/nexus-a11y.*` | No bank calls the Auditor |
| W10 | Nexus adapters registered with the `helixqa` CLI | `HelixQA/cmd/helixqa/` | `helixqa autonomous` still uses the pre-Nexus stack |

## 🟠 Suspected bugs (high-confidence, unverified against hardware)

| # | Bug | File | Fix sketch |
|---|---|---|---|
| B1 | Fanart.tv movie lookup passes `tmdb_id` where the library expects an IMDB id | `catalog-api/internal/services/fanart_resolver.go` | CanResolve should require `imdb_id` for movies; TV needs `tvdb_id`. Document this in the README. |
| B2 | IGDB client never sends `Client-ID` header | `catalog-api/internal/services/igdb_resolver.go` | `Henry-Sarabia/igdb` v1 takes a single `apiKey` argument; the Client-ID requirement needs a custom http.RoundTripper that injects the header. |
| B3 | chromedp `Hover` uses a non-existent `chromedp.MouseEvent("", ...)` call | `HelixQA/pkg/nexus/browser/chromedp_driver.go` | Use `cdp/input.DispatchMouseEvent` through `chromedp.ActionFunc` to send a real `mouseMoved` event. |
| B4 | rod `SavePDF` accumulates via `Read` on a `StreamReader` — may deadlock | `HelixQA/pkg/nexus/browser/rod_driver.go` | Use `io.ReadAll(reader)` instead of the manual chunked loop. |
| B5 | CSRF `__Host-` cookie prefix fails on plain HTTP dev | `catalog-api/middleware/csrf.go` | Either drop the prefix when `secure=false`, or document HTTPS-only use. |
| B6 | SAML `possibleRequestIDs` default `[]string{""}` may be rejected by strict checks | `HelixQA/pkg/nexus/orchestrator/saml_concrete.go` | Require callers to pass a non-empty slice; return an error when both args are empty. |
| B7 | OTel `semconv/v1.26.0` pin may drift from the installed SDK | `HelixQA/pkg/nexus/observability/otel_tracer.go` | Pin the semconv version in `go.mod` and add a compile-time check. |
| B8 | HTML entity decode only applied to `Element.Name`, not `Selector` | `HelixQA/pkg/nexus/browser/snapshot.go` | Apply `decodeHTMLEntities` to the built selector too. |
| B9 | Linux `Click(Element{})` on X11 clicks at the current cursor — not useful | `HelixQA/pkg/nexus/desktop/linux.go` | Refuse Click when the element handle is empty and the AT-SPI path is unavailable. |
| B10 | AUC sort is O(n²) insertion sort — slow on large holdouts | `HelixQA/pkg/nexus/ai/predictor_training.go` | Replace with `sort.Slice` when `len > 500`. |

## 🟡 Polish & hardening

- **P1** — no rate limiter on `/admin/image-quality/revalidate` or on provider API calls (Fanart.tv, IGDB).
- **P2** — no circuit breaker around provider calls; a slow upstream stalls the whole cover chain.
- **P3** — Predictor weights never persisted to disk; training evaporates on restart.
- **P4** — FileEvidenceStore grows unbounded; add a retention sweep.
- **P5** — Prometheus Go-runtime collector (`collectors.NewGoCollector()`) not mounted alongside the Nexus bridge.
- **P6** — CSRF test file parses `Set-Cookie` by hand; fragile against stdlib changes.
- **P7** — `useCoverQuality` hook has no Vitest coverage; `CoverQualityBadge` has no Storybook story.
- **P8** — no fuzz target on `decodeHTMLEntities`, `parseAttrs`, or the CSP string.
- **P9** — Generator collision on duplicate `id` fields (e.g. two `NX-GEN-demo`) silently overwrites prior banks.
- **P10** — MinIO test never spins up a real MinIO container; `PutObject` + `ListObjects` happy paths are mock-only.

## ⚫ External / human-gated

- **E1** — Full container QA campaign (`release-build --container` + `services-up` + `helixqa-orchestrator`) never executed.
- **E2** — Real WinAppDriver harness (Windows Pro or YWinAppDriver) + real macOS harness (XCUITest on a Mac runner).
- **E3** — Video course MP4s 01–08 (shot lists, VO scripts, exercises all shipped; filming + editing outstanding).
- **E4** — Public `helixqa.vasic.digital/nexus` DNS + VitePress deploy.
- **E5** — Android + Tauri desktop clients surfacing `X-Cover-Quality` debug UX (catalog-web landed; per-client PRs needed).
- **E6** — Predictor training against a real historical flake dataset (to verify the "AUC > 0.75" success criterion end-to-end).
- **E7** — Fanart.tv / IGDB / Twitch credentials acquired + rotated via the operator vault.

## Recommended next-session order

1. **Land W1–W10** in one cohesive commit (~90 minutes). This is the highest-leverage change in the whole backlog: every library we shipped starts producing value the moment it's wired up.
2. **Fix B1–B4** (two Go calls, one header injector, one `io.ReadAll`) against a real upstream; add an integration test each so the fix holds.
3. **Run E1** (container QA campaign) to surface the next tier of real-world issues.
4. **Close polish items P1–P5** once E1 has produced a signal-rich baseline.
5. **Address E2–E7** once hardware / content / ops access is available.

Every item above is a single file + tests. Nothing requires a spike or
redesign.

## Closed so far (for reference)

- All library-level R-phase deliverables: R-0 dep landing, R-1 observability, R-2 SSO (OIDC + SAML verifiers), R-3 LLMOrchestrator bridge, R-4 provider resolvers, R-5 AT-SPI native, R-6 Nexus web-flow challenge, R-10 polish (admin revalidate, generator validator, HTML entities, CSV training, AUC, k6 runner shell-out, case-insensitive allowlist, bluemonday sanitizer, security headers, CSRF middleware, catalog-web debug hook + badge).
- Cover-quality gate + 14 `CH-IQ-*` challenges + 3 provider resolvers.
- 10 Nexus Go packages with 260+ tests.
- 14 HelixQA banks (123 cases).
- Grafana dashboard JSON, 8 video module outlines, VitePress site source,
  all SQL schemas, operator runbooks.

Every item in the "Closed" list is regression-clean on
`go test ./...` (catalog-api + HelixQA + Media submodules).
