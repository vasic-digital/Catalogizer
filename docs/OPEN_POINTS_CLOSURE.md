# Open Points — Closure Brief

**Last refresh:** 2026-04-18 (OCU P4 closed)
**Owner:** Operator (you). Every item below is work Claude cannot do
autonomously — they need credentials, hardware, external accounts, or
human judgment / filming / editing time.

Everything else (all W-items, all B-items, P1–P10, OpenClawing2 phases
1–8 code/infra, Phase-4 resilience wiring, coord dispatch across all
three adapters, shared SSRF module, goxmldsig CVE, memory trim) is
**closed** in git — HelixQA, Security, and all six main-repo remotes.

Treat this document as the definitive "what WE must provide" list.
Once every checkbox is ticked, the program has zero open points.

---

## 1. Credentials & secrets (rotate + vault)

These are API / service credentials that must be issued, validated,
and stored in an operator vault. Claude refuses to touch production
secrets; you own this.

- [ ] **Fanart.tv API key** — register at <https://fanart.tv/get-an-api-key/>,
      set `FANART_TV_API_KEY` in `catalog-api/.env`. Tests the Fanart
      resolver CanResolve path (`FanartTVResolver_B1`). Without the
      key the resolver silently no-ops.
- [ ] **IGDB / Twitch client-id + bearer token** — Twitch dev console →
      IGDB app registration → `IGDB_CLIENT_ID` + `IGDB_CLIENT_SECRET`
      OR `IGDB_BEARER_TOKEN` in `catalog-api/.env`. IGDB v4 hard-rejects
      requests without both headers.
- [ ] **TMDB v3 + v4 keys** — <https://www.themoviedb.org/settings/api>
      → `TMDB_API_KEY` + `TMDB_ACCESS_TOKEN`. Already in most dev
      environments; double-check `.env.example` placeholders are
      replaced in every deployed instance.
- [ ] **OMDB API key** — <https://www.omdbapi.com/apikey.aspx> →
      `OMDB_API_KEY`. Free tier is 1000/day; bump to paid tier once
      the catalog size exceeds that threshold.
- [ ] **Astica.AI key** — <https://astica.ai/> account → `ASTICA_API_KEY`
      in HelixQA `.env`. Analyze-only (no JSON navigation).
- [ ] **Gemini / OpenAI / Anthropic / Kimi cloud LLM keys** — one or
      more is required for the autonomous QA pipeline unless
      llama.cpp RPC is fully self-contained. Set
      `GEMINI_API_KEY` / `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` /
      `MOONSHOT_API_KEY` in HelixQA `.env`.
- [ ] **SEMGREP_APP_TOKEN** — <https://semgrep.dev/orgs/-/settings/tokens>
      → export in shell rc. Silences the cosmetic
      `No SEMGREP_APP_TOKEN found` hook spam. Optional but recommended.
- [ ] **Sentry / error-reporting DSN** — if enabling, set
      `SENTRY_DSN`. Not wired in by default.

**Rotation policy:** all the above must rotate every 90 days. Store
only in the operator vault, never in git. `.env` is already listed in
every submodule's `.gitignore`; re-verify before each rotation.

---

## 2. Hardware / test harnesses

These require physical hardware Claude cannot acquire.

- [ ] **E2 — Windows Pro host** with WinAppDriver + YWinAppDriver
      installed. HelixQA's desktop-Windows branch is wired but idle
      without a real Windows runner reachable via SSH / VNC.
- [ ] **E2 — macOS host** with XCUITest bench. Same story: code is
      present; there is no Mac in the test lab.
- [ ] **Android TV device pool** — keep at least two working devices
      reachable for HelixQA (Mi Box 4 at 192.168.0.134:5555 is the
      current reference). `.devconnect` must list the IPs.
- [ ] **Physical Android phone** — running Android 13+ for
      `catalogizer-android` UI regression.
- [ ] **Synology (or equivalent) NAS** populated with the canonical
      media corpus — `synology.local` at 192.168.0.241 is the pin.
      First-catalog-populate challenges run ~25 min against it.

---

## 3. Infra & deployments

- [ ] **E4 — DNS for `helixqa.vasic.digital/nexus`** — point at the
      VitePress host, then `npm run docs:build && rsync dist/` to the
      web server. Docs already exist in
      `HelixQA/docs/website/nexus/`; just needs the deploy.
- [ ] **E1 — Section-9 live campaign** — `./scripts/openclaw-full-campaign.sh`
      end-to-end with:
      - PostgreSQL + Redis + catalog-api + MinIO + Prometheus +
        Grafana running in podman-compose
      - HelixQA infra pod alive (set `HELIX_INFRA_HOST`)
      - `.devconnect` devices online
      - All cloud LLM keys loaded (§1)
      - Two consecutive green runs required; Grafana dashboard
        (`monitoring/openclaw-dashboard.json`) populated.
- [ ] **E6 — Predictor training dataset** — export ≥6 months of
      `fixes-validation` bank run history into a CSV, train
      `Predictor.SaveWeights` against it, confirm AUC > 0.75 per the
      OpenClawing2 Phase-7 success criterion.
      Seed data: every `banks/fixes-validation-*.yaml` run result in
      `qa-results/session-*/`.
- [x] **GitLab remote for Security submodule** — ~~stale~~ — **DONE
      2026-04-17**: `gitlab` remote wired and mirrored up to
      `6ab39c5`.
- [ ] **Rotate signing keys** — `docker/signing/signing.properties`
      (Android) and Tauri code-signing certs (desktop). Per the
      90-day schedule.

---

## 4. Human / creative

- [ ] **E3 — Video courses 01–08 MP4s** — shot lists, VO scripts, and
      exercises are all shipped (see `docs/video-course/`). Filming
      and editing are outstanding. Deliverables: eight MP4s at
      1080p, 16:9, narrated; chaptered.
- [ ] **Graphics / brand review** — confirm Vasic Digital logo
      (rounded square with red border) meets brand-compliance every
      release. HelixQA's issues/HELIX-004 tracks any drift from
      this — resolve those in Figma before shipping.
- [ ] **Legal licence review** — confirm all OSS licences in
      `HelixQA/docs/opensource/LICENCE-INVENTORY.md` satisfy the
      distribution model. Four licences were added in OpenClawing2
      Phase 1 (browser-use / skyvern / ui-tars-desktop /
      anthropic-quickstarts); read before shipping a commercial build.

---

## 5. Optional hardening (deferred, not blockers)

These are *not* open points — they are "would be nice, can be scoped
later" items I'm listing so they aren't forgotten.

- [ ] Migrate catalog-api `internal/services/ssrf_guard.go` and
      HelixQA `pkg/nexus/ai/ssrf_guard.go` to import the canonical
      `Security/pkg/ssrf` package. Today both are hand-synced copies;
      the canonical module exists (commit `6ab39c5` in Security).
      Requires: add Security submodule to HelixQA's `go.mod` replace
      graph, same for catalog-api; vendor refresh; rewrite both
      local guards to `import "digital.vasic.security/pkg/ssrf"`.
- [ ] Wire `gorilla/csrf` on any future cookie-auth endpoints. JWT
      Bearer auth is CSRF-safe by default so this is latent until
      the auth model changes.
- [ ] Replace `database/sql` + manual rewriting (catalog-api
      `database/dialect.go`) with a library like `sqlc` or `ent`
      once the schema stabilises — removes a category of dialect
      rewriting bugs.
- [x] **OCU P0 — Foundation + Go↔Native bridging** — **CLOSED
      2026-04-17**: contracts (`pkg/nexus/native/contracts/`) + budget
      (`pkg/nexus/native/budget/`) + probe (`pkg/nexus/native/probe/`) +
      remote dispatcher (`pkg/nexus/native/remote/`) land in HelixQA;
      Containers GPU extension (`HostResources.GPU`, `GPURequirement`,
      `StrategyGPUAffinity`, `GPUHealthCheck`, `ProbeGPU`) land in
      Containers. Vertical-slice CLIs `cmd/ocu-probe` and
      `cmd/ocu-dispatch-test` prove thinker.local routing end-to-end.
      10-category test coverage (unit/integration/stress/security/
      benchmark/challenges) per Constitution §V. Spec + plan in
      `HelixQA/docs/superpowers/{specs,plans}/2026-04-17-*`. Next wave
      (P1 capture, P2 vision, P3 interact, P4 observe) can now proceed
      in parallel.
- [x] **OCU P1 — GPU capture engine** — **CLOSED 2026-04-18**: factory +
      web/CDP + linux/X11 + android/ADB (phone + TV) CaptureSource
      plumbing; stress tested under -race; per-source bench + audit
      filed; challenge bank `ocu-capture.json` shipped. Production
      subprocess wiring (chromedp / xwd / adb screenrecord) remains
      P1.5 scope via the injectable `newFrameProducer` pattern. Plan
      `HelixQA/docs/superpowers/plans/2026-04-17-ocu-p1-capture-plan.md`.
- [x] **OCU P2 — GPU vision pipeline** — **CLOSED 2026-04-18**:
      Pipeline + CPU backend + remote-dispatch plumbing via
      ocuremote.Dispatcher. Real OpenCV CUDA + TensorRT OCR deferred
      to P2.5 with the LocalBackend interface + stub remote path.
      13-entry challenge bank `ocu-vision.json`, 100-goroutine -race
      stress, bench baseline appended (Analyze 3.8 ns/0 allocs,
      Diff 48 ns/1 alloc). Integration smoke across all 4 methods
      green. Plan
      `HelixQA/docs/superpowers/plans/2026-04-17-ocu-p2-vision-plan.md`.
- [x] **OCU P3 — Interaction engine** — **CLOSED 2026-04-18**:
      Factory + 4 Interactor backends (linux/uinput-planned, web/CDP,
      android, androidtv) + verifier Wrap hook. Real evdev/CDP/ADB
      transport deferred to P3.5 via injectable newInjector pattern.
      19-entry bank `ocu-interact.json`. -race-clean 100-goroutine stress
      (per-goroutine private injectors). Bench: Wrap_Click ~86 ns/op
      0 allocs; NoOp_After ~0.34 ns/op 0 allocs (i7-1165G7). govulncheck
      clean. go vet clean. gofmt clean. Integration smoke (all 4 kinds
      open + Wrap composes) green. Plan
      `HelixQA/docs/superpowers/plans/2026-04-17-ocu-p3-interact-plan.md`.
- [x] **OCU P4 — Observation engine** — **CLOSED 2026-04-18**:
      Factory + 5 Observer backends (ld_preload, plthook, dbus, cdp,
      ax_tree) + shared BaseObserver + bounded RingBuffer. Injectable
      producerFunc sentinel; production returns ErrNotWired. Real shim
      install, PLT/GOT patching, D-Bus/CDP/AT-SPI2 subscription deferred
      to P4.5. 21-entry bank `ocu-observe.json`. -race-clean 100-goroutine
      stress (20 per kind × 5 kinds). govulncheck clean. go vet clean.
      gofmt clean. Integration smoke (all 5 kinds open) green. Plan
      `HelixQA/docs/superpowers/plans/2026-04-17-ocu-p4-observe-plan.md`.

---

## 6. What does "done" look like?

When every checkbox in sections 1–4 is ticked, you have:

- All provider resolvers fully populating the catalog.
- Two green consecutive section-9 live campaigns in CI/local.
- Android, Android TV, and Tauri clients all regression-tested on
  real hardware end-to-end.
- HelixQA running autonomously across all platforms with real
  vision models (cloud + distributed llama.cpp RPC).
- The public Nexus docs site live at `helixqa.vasic.digital/nexus`.
- The eight video courses published.

At that point the program has **zero open points**. Section 5 is
discretionary improvement — not a blocker.

---

## 7. Verification commands

Run these after each rotation / deploy to confirm the state is
actually green:

```bash
# Credentials wired
cd catalog-api && grep -E '^(TMDB|OMDB|IGDB|FANART)' .env | grep -v YOUR_
cd HelixQA && grep -E '^(GEMINI|OPENAI|ANTHROPIC|ASTICA)' .env | grep -v YOUR_

# Vulnerability scan — expect zero
cd HelixQA && GOTOOLCHAIN=local govulncheck -mode source ./...
cd catalog-api && GOTOOLCHAIN=local govulncheck -mode source ./...
cd catalog-web && npm audit --production

# SSRF guard tests green
cd catalog-api && GOMAXPROCS=3 go test ./internal/services/... -count=1 -race
cd HelixQA && GOTOOLCHAIN=local go test -mod=vendor ./pkg/nexus/ai/... -count=1 -race
cd Security && GOTOOLCHAIN=local go test ./pkg/ssrf/... -count=1 -race

# Devices reachable
./scripts/devconnect.sh

# Section-9 dry run
./scripts/openclaw-full-campaign.sh --skip-benchmarks

# Section-9 live (after §1 + §3 ticked)
./scripts/openclaw-full-campaign.sh
```

All of the above should exit 0 and report "OK". If any one reports
an error, the corresponding section-1–4 checkbox is not yet green.

---

## 8. Contacts & handoff

- **Upstream remotes:** six on the main repo (GitHub × 2, GitLab ×
  2, GitFlic, GitVerse port 2222); four on HelixQA (GitHub × 2,
  GitLab × 2); one on Security (GitHub — add GitLab when ready).
  Always push via `GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push
  origin main` to hit every configured remote in one shot.
- **Branch policy:** `main` only. No long-lived feature branches.
  No force-pushes to `main`.
- **Dependabot:** GitHub UI surfaces high-severity alerts
  immediately. Act on every one within 72 hours. The most recent
  (GO-2026-4753 in goxmldsig) was closed by commit `78ec16b` on
  HelixQA.

---

_Add new items only when they match the pattern "requires access /
hardware / credentials / human judgment." Routine code work does
not belong here — it belongs in `docs/nexus/remaining-work.md` as a
W- / B- / P- item._
