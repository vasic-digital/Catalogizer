# Open Points — Closure Brief

**Last refresh:** 2026-04-19 (OpenClawing4 Phase 1 Go-core further extended on HelixQA — eleven milestones shipped across six packages: M1 `pkg/capture/frames` [97.1%, `61d2696`], M2 `pkg/bridge/sidecarutil` [84.5%, `bcdc740`], M3 `pkg/bridge/scrcpy` v3 wire + devguard [91.4%, `25599bb`], M4 `pkg/navigator/linux/uinput` [42% pkg, event.go 100%, `341fe33`], M5 `pkg/bridge/scrcpy/{server,session}` [81.5% pkg, `ee83028`], M6 `pkg/bridges/registry` ToolKind + 13 native probes [100%, `a28657e`], M7 `pkg/capture/linux/sidecar.go` [72.8% starter, `0c53389`], M8 `pkg/capture/linux/router.go` [79.6% pkg, `801b04c`], M9 `pkg/capture/linux/portal.go` [83.9% pkg, `ad0c0ec`], M10 `pkg/capture/linux/{pipewire,kmsgrab}.go` BackendFactory helpers [84.3% pkg, `bdfc375`], M12 `pkg/capture/android/direct.go` scrcpy-direct delegation + new `scrcpy.NewSession` constructor [88.9% android / 81.3% scrcpy, `4bc738f`]; handover rollup [`cbcef46`] with 19-case `banks/phase1-gocore.yaml`. Remaining in Phase 1: production `portal_dbus.go` Caller wrapping `*dbus.Conn`, `x11grab.go` factory (needs companion `cmd/helixqa-x11grab` NAL-splitter sidecar), `xcbshm.go` fallback, `linux_capture.go` routing modification, `pkg/navigator/linux/libei.go`, `x11_executor.go` -tags migration, native sidecar binaries (`helixqa-capture-linux` C+GStreamer, `helixqa-kmsgrab` C+DRM, `helixqa-input` Rust+enigo, `helixqa-x11grab` Go+ffmpeg), scrcpy-server JAR pin, feature-level banks — see `OpenClawing4-Handover.md` §3.1 🚧 rows. Prior (2026-04-19): OpenClawing4 Phase 0 closed — retraction banners on OpenClawing2/3/Starting_Point, `scripts/hooks/no-sudo.sh` pre-commit hook + `.pre-commit-config.yaml`, new `banks/docs-audit.yaml` with 7 mechanical checks, 14 `fixes-validation.yaml` entries FIX-OC2-001..003 + FIX-OC3-001..011, HQA-DOCS-001 challenge [HelixQA `a2f3764` + handover `b2445ec`]. Prior (2026-04-19): Tauri auto-container dispatch: Rust toolchain relocated to /opt in `docker/Dockerfile.builder`, builder image rebuilt + `cargo --version && rustc --version` verified inside `--userns=keep-id` container [commits 35624a2f + 4c8dcefc]; pure-bash test suite under `scripts/tests/` wired into `scripts/run-all-tests.sh`; `docs/BUILD_CONTAINER_AUTO_DISPATCH.md` published. Phase-3 baseline re-green on current tip: catalog-api 44/44 packages, catalog-web 2318/2318, api-client 283/283, shell-layer 9/9. Phase-4 partial: 27/27 dep-clean category runs ✅, 10/11 dep-free leaf challenges ✅. Device pool regression: ADT-3 at 192.168.0.193 unreachable — see §2. Three new challenge-framework defects surfaced — see §6.)
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
      **STATUS 2026-04-19**: device pool **0 usable**:
      - ADT-3 at `192.168.0.193` was added to `.devconnect` on
        2026-04-18 but is **unreachable** today (100 % ping loss).
        Needs operator to power it on / fix network.
      - The two USB devices currently attached (transport_id 1+2,
        model=`ATMOSphere`) are in `.devignore` and forbidden.
      - Until at least one non-ATMOSphere device is online, Phase 6
        Android / Android-TV HelixQA is blocked. The remaining
        non-device lanes (API, web, desktop) still run.
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

- [x] **SSRF guard migration to canonical Security/pkg/ssrf** — **CLOSED
      2026-04-18**: catalog-api `internal/services/ssrf_guard.go` and
      HelixQA `pkg/nexus/ai/ssrf_guard.go` rewritten as thin adapters
      (type aliases + delegating wrappers). ~150 LOC removed from each.
      Public API preserved (SSRFGuardConfig, Resolver, GuardProviderURL /
      ValidateURL, ErrSSRFBlocked, AllowPrivateSSRFForTests). HelixQA
      vendor/modules.txt + vendor/digital.vasic.security/ updated; go.mod
      replace + require added to both consumers. All tests green (race-
      clean). Commits: catalog-api `ce5c1be1`, HelixQA `c170ad2`,
      submodule bump `e4abd73e`.
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
- [x] **OCU P5 — Recording & streaming** — **CLOSED 2026-04-18**:
      Recorder + 3 encoder stubs (x264/nvenc/vaapi) + bounded FrameRing +
      clipper (newline-delimited JSON frame metadata; real MKV/MP4 muxing
      deferred to P5.5) + WebRTC/WHIP publisher off by default (ErrNotWired;
      real ICE/DTLS/RTP deferred to P5.5). Priority-drain goroutine prevents
      frame loss when Stop() races a pre-closed source channel. NVENC stub
      documents P5.5 remote-dispatch path via ocuremote.Dispatcher (reuses
      P2 SSH trust, no new credential). WebRTC BindAddr defaults to
      127.0.0.1 (never 0.0.0.0 without explicit operator flag). 21-entry
      bank `ocu-record.json`. -race-clean 100-goroutine stress. govulncheck
      clean. go vet clean. gofmt clean. Integration smoke (Recorder +
      in-memory source + all 3 encoder kinds registered) green. Security
      audit `HelixQA/docs/security/ocu-p5-audit.md`. Plan
      `HelixQA/docs/superpowers/plans/2026-04-17-ocu-p5-record-plan.md`.
- [x] **OCU P6 — Unified automation surface** — **CLOSED 2026-04-18**:
      Engine composes P1–P5 (CaptureSource / VisionPipeline / Interactor /
      Observer / Recorder) behind single `Engine.Perform(ctx, Action)`.
      Action is pure data (Kind + At/To/Text/Key/Button/DX/DY/ClipAround/
      ClipWindow/Expected); Engine is a pure dispatcher — zero decision logic.
      Result carries Success, VerificationPassed, Elapsed, Evidence[],
      DispatchedTo. verifier/ sub-package: PixelVerifier (Vision.Diff
      threshold, errors never swallowed) + MultiVerifier (AND-chain,
      first-fail short-circuit). agent_bridge/ sub-package: Bridge.
      ExecuteAction is a provable one-liner adapter — LLM remains sole
      decider. 22-entry bank `ocu-automation.json`. -race-clean
      100-goroutine stress. govulncheck clean. go vet clean. gofmt clean.
      Integration smoke (Capture → Click → Analyze → RecordClip sequence +
      Bridge full sequence + all-Elapsed check) green. Security audit
      `HelixQA/docs/security/ocu-p6-audit.md`. Plan
      `HelixQA/docs/superpowers/plans/2026-04-17-ocu-p6-automation-plan.md`.
- [x] **OCU P7 — Magical tickets + tests + challenges** — **CLOSED
      2026-04-18**: `pkg/ticket` extended with 12 OCU evidence kind
      constants (`EvidenceKindClip`, `EvidenceKindDiffOverlay`,
      `EvidenceKindOCRDump`, `EvidenceKindElementTree`,
      `EvidenceKindHookTrace`, `EvidenceKindReplayScript`,
      `EvidenceKindLLMReasoning`, `EvidenceKindPerfMetrics`,
      `EvidenceKindAXTreeDiff`, `EvidenceKindHAR`,
      `EvidenceKindWebRTCStream`, `EvidenceKindRawDMA`) + `Evidence`
      struct + `FromAutomationResult` helper + `BuildReplayScript`
      generator + `.ocu-replay` DSL spec (`docs/ocu-replay-format.md`).
      4 cross-cutting challenge banks: `ocu-tickets.json` (36 entries),
      `ocu-adversarial.json` (20 entries), `ocu-cross-platform.json`
      (15 entries), `ocu-fixes-validation.json` (10 entries) = 81 total.
      10-category campaign script `scripts/ocu-full-campaign.sh` — all
      8 active categories PASS. Pre-existing vet/fmt/integration bugs
      fixed (Engine Elapsed named-return, pipeline stepCancel defer,
      ExtractionStats/DetectorStats mutex-copy, 149-file gofmt). Release
      notes `HelixQA/docs/releases/v4.0.0.md`. **OCU program fully
      closed — v4.0.0 tagged and pushed to all 4 HelixQA remotes.** Plan
      `HelixQA/docs/superpowers/plans/2026-04-17-ocu-p7-tickets-plan.md`.
- [x] **OCU P1.5 + P3.5 web + android wiring** — **CLOSED 2026-04-18**:
      Production chromedp backend for web capture + interact;
      production adb backend for android capture + interact.
      Graceful ErrNotWired fallback when browser/adb absent +
      env kill-switches (HELIXQA_CAPTURE_WEB_STUB=1,
      HELIXQA_INTERACT_WEB_STUB=1, HELIXQA_CAPTURE_ANDROID_STUB=1,
      HELIXQA_INTERACT_ANDROID_STUB=1, HELIXQA_ADB_SERIAL for
      multi-device). pngToBGRA8 helper, H.264 NAL splitter, ADB keycode
      map (10 keys). 40 nexus packages -race green, vet+gofmt+govulncheck
      clean. Linux uinput/xwd, FFmpeg NVENC/VAAPI, LD_PRELOAD/plthook/
      dbus/ax_tree still stubbed (P2.5/P3.5-linux/P4.5 scope).
- [x] **OCU P1.5 + P3.5 Linux wiring** — **CLOSED 2026-04-18**:
      Production xwd+convert pipeline for Linux capture
      (xwd → gnome-screenshot → grim fallback chain, BMP→BGRA8
      decoder, pngToBGRA8 helper). Production xdotool/ydotool backend
      for Linux interact (xdotool X11 preferred, ydotool Wayland
      fallback, KeyCode→X11-keysym map for 10 keys). Kill-switches:
      HELIXQA_CAPTURE_LINUX_STUB=1 / HELIXQA_INTERACT_LINUX_STUB=1.
      DISPLAY+WAYLAND_DISPLAY guard prevents capture without a display
      server. Raw /dev/uinput path deferred; xdotool covers 95% of QA
      interactions no-sudo. Operator setup: docs/ocu-udev-setup.md.
      44 nexus packages -race green, vet+gofmt clean.

- [x] **OCU P2.5/P4.5/P5.5 partial wiring** — **CLOSED 2026-04-18**:
      Pure-Go CPU vision (per-pixel |Δ| diff with contiguous
      flood-fill into ChangeRegions + Sobel X+Y edge detection →
      UIElements, Kind "contour", Source "cv"). D-Bus observer via
      godbus/dbus/v5 (ConnectSessionBus + AddMatchSignal per
      target.Labels["interface"], signalToEvent pure translation).
      CDP observer via chromedp ListenTarget (Network.responseReceived
      + Runtime.consoleAPICalled). x264 encoder via ffmpeg libx264
      subprocess (rawvideo bgra stdin → MP4 stdout, frag_keyframe).
      Kill-switches: HELIXQA_VISION_CPU_STUB=1 /
      HELIXQA_OBSERVE_DBUS_STUB=1 / HELIXQA_OBSERVE_CDP_STUB=1 /
      HELIXQA_RECORD_X264_STUB=1. Graceful ErrNotWired when binary
      absent or stub env set. All 30 nexus packages -race green,
      vet+gofmt clean. Remaining stubs: LD_PRELOAD hook, plthook,
      AT-SPI walker, NVENC, VAAPI, real OpenCV CUDA sidecar.

- [x] **OCU P4.5 AT-SPI + P5.5 VAAPI wiring** — **CLOSED 2026-04-18**:
      AT-SPI observer via godbus/dbus/v5 on the a11y bus (Object +
      Window signals). VAAPI encoder via ffmpeg h264_vaapi with
      hw_device + nv12 upload filter. Kill-switches
      HELIXQA_OBSERVE_AX_STUB / HELIXQA_RECORD_VAAPI_STUB.

- [x] **OCU P5.5 NVENC client + P4.5 LD_PRELOAD loader** — **CLOSED
      2026-04-18**: NVENC encoder now routes via ocuremote.Dispatcher
      to a thinker.local Worker (real gRPC sidecar is P5.6
      operator-setup). Per-frame call uses *structpb.Value placeholder;
      P5.6 replaces with generated NVENCRequest/NVENCResponse from
      proto/nvenc.proto. Close() releases the Worker. Kill-switch
      HELIXQA_RECORD_NVENC_STUB=1; Dispatcher error (no GPU host) →
      ErrNotWired. LD_PRELOAD observer launches target binary with
      LD_PRELOAD=<shim.so> + HELIXQA_LD_SHIM_FIFO=<fifo>; reads
      newline-delimited JSON from the FIFO and emits EventKindHook
      Events (Payload["fn"], Payload["arg"], nanosecond Timestamp).
      Shim path from target.Labels["shim_path"] or HELIXQA_LD_SHIM
      env. FIFO created via syscall.Mkfifo (!windows build tag). C shim
      template + operator README at docs/hooks/. Kill-switch
      HELIXQA_OBSERVE_LDPRELOAD_STUB=1. All nexus packages -race green,
      vet+gofmt+govulncheck clean. HelixQA commits cd923f7 + e07188e
      pushed to all three remotes.

      Remaining stubs: plthook (deep /proc/self/maps + unsafe work;
      standalone session), real thinker.local CUDA sidecar gRPC server
      (P5.6 operator infra), real OpenCV CUDA sidecar (P2.5).

- [x] **OCU CUDA sidecar source + Dockerfile** — **SOURCE READY
      2026-04-18**: `OCU-CUDA-Sidecar/` directory committed to main repo
      (commit `f35f6f53`). Standalone Go module
      `digital.vasic.ocu-cuda-sidecar` with:
        - `proto/ocu.proto` — canonical NVENC + OpenCV service definitions
        - `proto/ocu.go` — hand-authored Go types (no protoc required)
        - `internal/server/backend_stub.go` (!cuda) — unit-testable stub,
          no GPU deps; 20 tests green, race-clean
        - `internal/server/backend_cuda.go` (cuda) — gocv CUDA backend
          (Analyze/Diff/Match/OCR via OpenCV CUDA; NVENC via session registry)
        - `internal/server/server.go` — shared HTTP/JSON shim + routes
        - `cmd/ocu-cuda-sidecar/main.go` — entrypoint, --listen flag
        - `Dockerfile` — nvidia/cuda:12.2.0-devel multi-stage build
      Operator tasks remaining (still unchecked in §3):
        1. `podman build --network host -t ocu-cuda-sidecar .` on a CUDA host
        2. `podman save ocu-cuda-sidecar | ssh thinker.local "podman load"`
        3. `podman run --device nvidia.com/gpu=all -p 50060:50060 ocu-cuda-sidecar`
        4. Set `HELIX_OCU_CUDA_ADDR=thinker.local:50060` in HelixQA `.env`

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

## 8. Framework defects surfaced 2026-04-19 (Phase-4 retest)

These are **code defects** — not ops items — but they materially
block the "100 % category coverage" contract so they live here
until a fix ships. Evidence in
`docs/reports/qa-sessions/qa-session-2026-04-19/challenges/SUMMARY.md`.

- [ ] **`RunByCategory` orders challenges alphabetically, not
      topologically.** `digital.vasic.challenges/pkg/runner` should
      perform a Kahn sort on `dependencies` before executing. Today
      `asset-lazy-loading` (category `e2e`, depends on
      `browsing-api-health` in the same category) errors with
      "unmet dependency" because `a` sorts before `b`. Affects 11
      categories: `api`, `admin-ops`, `api-consistency`, `coverage`,
      `e2e`, `lint`, `middleware`, `module-integration`,
      `observability`, `security`, `test`.
- [ ] **`RunAll` is not practical on the current bank.** A 30-min
      partial run completed ~10 challenges. Needs:
      - per-challenge hard timeout distinct from the 5-min
        "stale-progress" detector,
      - optional `--parallel N` across challenges whose dep-set is
        satisfied,
      - streaming response writer (flush per-challenge result) so
        `RunAll` doesn't block read endpoints while it buffers the
        entire JSON body.
- [ ] **`database-connectivity` challenge hangs >5 min.** The
      underlying endpoint `GET /api/v1/stats/overall` responds in
      1.5 s against the 112 MB SQLite DB when called directly, but
      the challenge (which uses
      `digital.vasic.challenges/pkg/httpclient` with a 180 s
      timeout) never returns. Default challenge timeout is 5 min;
      the runner should kill at 5 min but the handler writes no
      response body to the HTTP client. Points to a goroutine /
      context-propagation bug between runner → challenge →
      handler. Evidence:
      `docs/reports/qa-sessions/qa-session-2026-04-19/logs/catalog-api-phase4.log`.

---

## 9. Contacts & handoff

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

## 10. OpenClawing4 — Phase 0 closed, Phases 1–6 remaining

### 10.1 Phase 0 — DONE

Closed by HelixQA commits `a2f3764` (artefacts) + `b2445ec` (handover),
pushed to all 4 HelixQA upstreams (GitHub × 2, GitLab × 2). Main-repo
pointer bumped by this closure commit.

What shipped:

- **Retraction banners** on `HelixQA/docs/openclawing/Starting_Point.md`,
  `OpenClawing2.md`, `OpenClawing3.md` — audit in
  `OpenClawing4-Audit.md` (850 lines) identified 9/24 dead URLs in
  Starting_Point, 3 fabricated internal paths in OpenClawing2, and a
  larger set of issues in OpenClawing3 (`src/...` fabricated tree,
  `sudo` violations, compile-blockers, wrong DXGI zero-copy claim,
  3–7× optimistic benchmarks, missing llama.cpp RPC primary, 16-week
  / 47-tech plan replaced by 24-week / 7-phase).
- **`HelixQA/scripts/hooks/no-sudo.sh`** (executable) — rejects bare
  `sudo` in committed content, allow-lists retraction docs,
  strike-through (`~~sudo~~`), and quoted (`"sudo"`) forms.
- **`HelixQA/.pre-commit-config.yaml`** — wires the hook into
  `pre-commit run --all-files` alongside standard pre-commit-hooks.
- **`HelixQA/banks/docs-audit.yaml`** — 7 mechanical checks
  (AUDIT-001..007) verifying banners, hook behaviour, real `pkg/...`
  citation in OpenClawing4, llama.cpp RPC primary invariant.
- **`HelixQA/banks/fixes-validation.yaml`** — 14 new regression
  entries: FIX-OC2-001..003 + FIX-OC3-001..011 (total now 44).
- **`HelixQA/challenges/config/helixqa-validation.yaml`** — HQA-DOCS-001
  registered.
- **`HelixQA/docs/openclawing/OpenClawing4-Handover.md`** — 478-line
  resume playbook for phases 1–6 (file-by-file landing points,
  acceptance per Article V category, blockers, fast-path resume).

Acceptance (Article V cat 6 security): YAML lint clean across 4 files
(44 + 7 + 30 test_cases); hook positively rejects bare-sudo fixture
(exit 1) and accepts strike-through / quoted (exit 0); OpenClawing4.md
1,347 lines with 42 `pkg/` landing rows and llama.cpp-RPC-primary
declared on lines 65 + 250.

### 10.1.1 Phase 1 Go-Core — DONE (eleven milestones shipped in one session)

Closed by HelixQA commits `61d2696` (M1) + `bcdc740` (M2) + `25599bb` (M3) + `8535f12` (rollup 1) + `341fe33` (M4 uinput) + `ee83028` (M5 server/session) + `a28657e` (M6 registry) + `b37fc66` (rollup 2) + `0c53389` (M7 sidecar) + `801b04c` (M8 router) + `ad0c0ec` (M9 portal) + `641535d` (rollup 3), all pushed to 4 HelixQA upstreams. Main-repo pointer bumped by this closure commit to `641535d`.

What shipped (nine milestones, pure-Go, no native sidecar binaries):

- **M1 `pkg/capture/frames/`** — normalised `Frame{PTS, Width, Height, Format, Source, DataFD, DataLen, Data, AXTree}` with `Format` enum (NV12/RGBA/BGRA/H264AnnexB), `New` + `NewFromFD`, `Validate`, idempotent nil-safe `Close`. 97.1% statement coverage.
- **M2 `pkg/bridge/sidecarutil/`** — stdio framing (length-prefixed JSON, 16 MiB ceiling, heartbeat, `DrainReader`), SCM_RIGHTS FD passing over `*net.UnixConn` (stdlib-only, CGO-free), `HealthProbe`/`MultiHealth` enforcing `<bin> --health → ok\n + exit 0`. 84.5% coverage.
- **M3 `pkg/bridge/scrcpy/`** — all 18 v3 client→server control messages byte-exact including `InjectTouchEvent` 31-byte body with `action_button` + `buttons` uint32s (fixes FIX-OC3-011), server→client `DeviceMessage`, video/audio packet decoders, `.devignore` enforcement.
- **M4 `pkg/navigator/linux/uinput/`** — pure-Go `/dev/uinput` driver: byte-exact `EncodeEvent` (24-byte input_event), high-level `WriteKeyTap` / `WriteClickAbs` / `WriteMoveRel` / `WriteScroll`, Linux-only `Open`/`Close` running UI_SET_EVBIT → UI_SET_*BIT → UI_DEV_SETUP → UI_DEV_CREATE via `unix.Syscall(SYS_IOCTL, ...)`. 42% pkg (event.go 100%).
- **M5 `pkg/bridge/scrcpy/{server,session}.go`** — `StartServer` full bring-up (devguard → push → reverse → listener → launcher → accept(1..3)); idempotent `Server.Stop` rollback; `Session.StartPumps` buffered channels; mutex'd `Session.Send` with 5s deadline. 81.5% pkg.
- **M6 `pkg/bridges/registry.go`** — `ToolKind` enum + `NativeTools`/`ExternalTools` partition helpers + 13 HelixQA-native sidecar probes. 100% pkg.
- **M7 `pkg/capture/linux/sidecar.go`** — envelope wire format (4B body length + 8B PTS + body), `Runner`/`Cmd` interfaces + `ExecRunner` production wrapper, `SidecarRunner` with single-shot Start + idempotent Stop + ctx-cancel clean termination; NoPTS sentinel (^uint64(0)) → `time.Since(startedAt)` fallback.
- **M8 `pkg/capture/linux/router.go`** — `Backend` enum (Auto/Portal/KMSGrab/X11Grab), `ParseBackend` alias support, `Source` interface, `BackendFactory` dispatch via `NewSource`, `ResolveBackend` precedence (override → HELIX_LINUX_CAPTURE → XDG_SESSION_TYPE → Portal default), `WrapSidecarAsSource` adapter.
- **M9 `pkg/capture/linux/portal.go`** — xdg-desktop-portal ScreenCast client via godbus: `Portal{Caller}` with `CreateSession`/`SelectSources`/`Start`/`OpenPipeWireRemote`; `Caller` interface hides the Request/Response signal handshake so tests inject a `fakeCaller` that records invocations and returns scripted responses; `ErrPortalStatus` + `IsUserCancelled`; unique `handle_token`/`session_handle_token` via `sync/atomic.Uint64`; raw `dbus.UnixFD` → `*os.File` ready for `exec.Cmd.ExtraFiles`.
- **M10 `pkg/capture/linux/{pipewire,kmsgrab}.go`** — concrete `BackendFactory` helpers. `NewPortalFactory` chains Portal + SidecarRunner deferred to Source.Start (so `NewSource` is cheap); `portalSource` Frames() returns a pre-closed non-nil channel before Start; full rollback (portal.Close, fd.Close) on failure. `NewKMSGrabFactory` is a thin SidecarRunner wrapper with operator-owned capability grant. Package now 84.3% combined.
- **M12 `pkg/capture/android/direct.go`** — `DirectSource` adapts `scrcpy.Server` + `scrcpy.Session.StartPumps().video` to emit `frames.Frame` values with `Source="scrcpy-direct"`; `IncludeConfig` gate (skip SPS/PPS by default); `IsDirectEnabled` helper for `HELIX_SCRCPY_DIRECT=1` env gating; `scrcpy.NewSession` constructor exposed so callers driving the server through custom transports (or tests with `net.Pipe`) participate in the same channel-based contract. 88.9% android pkg, 81.3% scrcpy pkg.
- **Test coverage rollup**: ~90 test functions across eleven milestones in six packages; all pass under `-race`.
- **`banks/phase1-gocore.yaml`** — 19 test cases across M1..M12 covering Article V categories 1 (unit), 2 (integration), 6 (security), 8 (benchmarking reference), 10 (HelixQA bank registration).
- **`challenges/config/helixqa-validation.yaml`** — HQA-PHASE1-GOCORE-001 rollup challenge (spans all nine milestones).
- **`scripts/hooks/no-sudo.sh`** — allow-list extended to `banks/phase[0-9]+-gocore\.(yaml|json)`.

Phase-1 items remaining (🚧 in `OpenClawing4-Handover.md` §3.1):
- `pkg/capture/linux/portal_dbus.go` — production `Caller` wrapping `*dbus.Conn` with real Response signal match + wait (the abstracted interface documents exactly what the wrapper must implement; needs a host with portal + Wayland for integration)
- `pkg/capture/linux/x11grab.go` — needs a companion `cmd/helixqa-x11grab` Go sidecar that runs ffmpeg x11grab and NAL-splits the raw H.264 output into envelopes (not a pure-refactor item)
- `pkg/capture/linux/xcbshm.go` — optional X11 fallback
- `pkg/capture/linux_capture.go` — modify the legacy file to route by `XDG_SESSION_TYPE` (today the new subpackage is additive; the legacy path still runs)
- `pkg/navigator/linux/libei.go` — Wayland input via RemoteDesktop portal + EI binary protocol writer
- `pkg/navigator/x11_executor.go` — move existing X11 path behind `-tags x11legacy`
- `cmd/helixqa-capture-linux/` (C+GStreamer) + `cmd/helixqa-kmsgrab/` (C+DRM+VAAPI) + `cmd/helixqa-input/` (Rust+enigo) + `cmd/helixqa-x11grab/` (Go+ffmpeg NAL wrapper) native sidecars
- scrcpy-server JAR pinned under `pkg/bridge/scrcpy/testdata/` (v3 release tag)
- `banks/capture-linux.yaml`, `banks/capture-android.yaml`, `banks/input-linux.yaml` — feature-level banks that require native sidecars for end-to-end runs

### 10.2 Phases 2–6 — NOT started, fully specified

Roadmap is in `HelixQA/docs/openclawing/OpenClawing4.md` §8 and
`OpenClawing4-Handover.md` §3. Any session resuming from here should:

1. `git pull` + recursive submodule update.
2. Read `OpenClawing4-Handover.md` §3 for the phase they are tackling
   (Phase 1 partial — Go-core done, 🚧 rows remain).
3. Create a feature branch `feat/openclawing4-phase-N` in HelixQA.
4. Implement per file list; commit per sub-phase; push.
5. Bump main-repo submodule pointer on every HelixQA commit that is
   ready for general consumption.

Phase-at-a-glance (all weeks estimated, see handover §3 for details):

| Phase | Scope | Weeks | Toolchain blockers | Status |
|---|---|---|---|---|
| 1 | Linux Wayland capture (PipeWire portal + kmsgrab), scrcpy-server v3 direct protocol (pure Go), libei + uinput input | 3–4 | pipewire + gstreamer + scrcpy-server JAR v3 | **Go-core DONE**; native sidecars + pkg/capture/linux + pkg/navigator/linux 🚧 |
| 2 | Unified AX tree, perception tiers (dHash → SSIM → DreamSim), BOCPD stagnation | 4 | OpenCV 4.x + gocv; Triton on GPU host for DreamSim | Not started |
| 3 | UI-TARS-1.5-7B + OmniParser v2 + LangGraph + SGLang | 4–6 | llama.cpp with mmproj; Python VLM sidecars | Not started |
| 4 | GPU compute sidecars: qa-vision-infer (TRT+NPP+OpenCV-CUDA), qa-video-decode (FFmpeg+NVDEC), qa-vulkan-compute PoC | 4 | CUDA 12.x + NVIDIA Container Toolkit | Not started |
| 5 | Observability: Frida sidecar, cilium/ebpf uprobes, LD_PRELOAD catalogue, rapid fuzzing, VLM-guided DFS | 3 | Linux 5.x BTF | Not started |
| 6 | macOS (SCKit) + Windows (WGC) + iOS (idb + WDA) + TUI (pty + ANSI grid) | 4–6 | macOS Xcode, Windows + WinRT SDK | Not started |

### 10.3 Operator-actionable items from OpenClawing4

These go in the sections above (§1 credentials, §2 hardware) when they
require operator input:

- [ ] UI-TARS-1.5-7B GGUF on `~/models/` + `llama-server` on port 18100
  with `--mmproj` (operator action, Phase 3).
- [ ] OmniParser v2 weights + Python 3.11 env in
  `cmd/helixqa-omniparser/` sidecar container (operator action, Phase 3).
- [ ] `pre-commit install` on every fresh HelixQA clone (one-time per
  clone, operator action).
- [ ] TensorRT engine rebuild on every NVIDIA driver major bump
  (operator action, Phase 4).
- [ ] ScreenCaptureKit entitlement on macOS host for unattended QA
  (operator action, Phase 6).
- [ ] Windows Graphics Capture entitlements on Windows host (operator
  action, Phase 6).

---

_Add new items only when they match the pattern "requires access /
hardware / credentials / human judgment." Routine code work does
not belong here — it belongs in `docs/nexus/remaining-work.md` as a
W- / B- / P- item._
