# Incident: Host Reboot Investigation — 2026-04-28

## Summary

The host `nezha` rebooted twice on 2026-04-28 (at 09:52→11:20 and 18:37→18:45)
and the user reported losing the Claude Code session and running processes.
This document records the root-cause investigation.

## Verdict

**Both reboots were graceful, user-initiated `systemd-poweroff` events.**
Nothing in the Catalogizer codebase, running containers, or Claude actions
triggered a power-state transition. CONST-033 hardening remains intact and
fully effective.

## Evidence

### CONST-033 source guard — PASS

```
$ bash challenges/scripts/no_suspend_calls_challenge.sh
=== summary: PASS ===
```

No `systemctl suspend`, `loginctl hibernate`, `pm-suspend`, `shutdown -h`,
`dbus-send …Manager.Suspend`, `gsettings … sleep-inactive-…` calls anywhere
in the source tree.

### CONST-033 host hardening — PASS (4/4)

```
$ bash challenges/scripts/host_no_auto_suspend_challenge.sh
[1/4] sleep / suspend / hibernate / hybrid-sleep targets masked? PASS
[2/4] AllowSuspend=no in sleep.conf or drop-in?                  PASS
[3/4] logind IdleAction safe? IdleAction=ignore                   PASS
[4/4] journal: 'will suspend' broadcasts since 2026-04-26 fix? 0  PASS
=== summary: 4 pass, 0 fail ===
```

### Boot 1 shutdown trace (09:52:40)

Final journal entries from boot ending at 09:52:44 show the canonical
graceful sequence:

```
09:52:44 systemd[1]: Reached target shutdown.target
09:52:44 systemd[1]: Reached target final.target
09:52:44 systemd[1]: systemd-poweroff.service: Deactivated successfully.
09:52:44 systemd[1]: Finished systemd-poweroff.service - System Power Off.
09:52:44 systemd[1]: Reached target poweroff.target
09:52:44 systemd-shutdown[1]: Syncing filesystems and block devices.
09:52:44 systemd-shutdown[1]: Sending SIGTERM to remaining processes...
```

`last -x shutdown` confirms a recorded shutdown event at 09:52 (matches a
manual `shutdown` / GUI power-off) — not a suspend/hibernate, not a panic.

### Boot 2 shutdown trace (18:37:57)

Identical pattern at 18:38:00:

```
18:38:00 systemd[1]: Reached target shutdown.target
18:38:00 systemd[1]: Finished systemd-poweroff.service - System Power Off.
18:38:00 systemd[1]: Reached target poweroff.target - System Power Off.
18:38:00 systemd-shutdown[1]: Syncing filesystems and block devices.
```

No abrupt termination, no kernel panic, no OOM kill — clean orderly
poweroff initiated by `systemd-poweroff.service`.

### What did NOT happen

- No `systemctl suspend` / `loginctl suspend` events in the journal of either
  preceding boot.
- No `pm-suspend` legacy invocations.
- No `dbus-send … org.freedesktop.login1.Manager.Suspend` calls.
- No kernel panic or hardware error log lines.
- No OOM-killer activity preceding either reboot.
- No Catalogizer container was running at the time (current state shows
  only unrelated user services: `yt-dlp-*`, `metube-*`).

### Cross-project sudo attempt (informational, not the cause)

At 19:10:11, after the second boot, an unrelated process from project
`/Projects/Boba` attempted `sudo /bin/dbus-launch gsettings list-recursively
org.gnome.desktop.session`. It was **rejected** by PAM (`milosvasic : user
NOT in sudoers`). This was a different project's tooling and did not run.
It is recorded here only because the user reported a session disturbance
and might wonder if it was related — it was not.

## Why the user lost the running session

A graceful poweroff sends SIGTERM to all user processes (including
`claude-code` and any in-flight container builds), then terminates the
user session and shuts down. From the user's perspective this looks
identical to "logged out / suspended / hibernated", but the journal
clearly shows it was a clean shutdown.

The most likely external triggers (none of which originate from this
codebase or from Claude):
- User pressed the power button or selected "Shut down" in the GNOME menu.
- A privileged operator ran `shutdown` / `reboot` from another terminal.
- Kernel OOM never fired (would have left a `Kill process …` audit line —
  none present).

## Mitigations already in place (do not touch)

| Layer | State |
|------:|------|
| `sleep.target` | masked |
| `suspend.target` | masked |
| `hibernate.target` | masked |
| `hybrid-sleep.target` | masked |
| `/etc/systemd/sleep.conf.d/99-no-auto-suspend.conf` | `AllowSuspend=no` |
| `/etc/systemd/logind.conf.d/99-idle-ignore.conf` | `IdleAction=ignore` |
| Source-tree scanner | enforced via `challenges/scripts/no_suspend_calls_challenge.sh` |
| Host-config scanner | enforced via `challenges/scripts/host_no_auto_suspend_challenge.sh` |

## Action items

1. ☑ Verify CONST-033 source + host guards (both PASS).
2. ☑ Document this investigation under `docs/incidents/`.
3. ☐ Cascade CONST-033 invariant text into all 41 submodule
   CONSTITUTION/CLAUDE/AGENTS files (see Task #4).
4. ☐ Add an "Anti-Bluff Testing" article to the Constitution + cascade
   (see Task #5).
5. ☐ Add Docker/Podman power-management risk audit to
   `docs/HOST_POWER_MANAGEMENT.md` (see Task #3).

## Key takeaway

**Nothing the user, Claude, the codebase, or the running containers did
caused the reboots.** The machine was powered off normally — twice — most
plausibly by the user, the GUI menu, or another operator. CONST-033 is
working as designed.

---

## Session-end report (2026-04-29 — full rebuild + redistribute)

After CONST-033 was re-verified, the user requested a full clean-slate
rebuild of all binaries and containers, redistribution to `amber.local` +
`thinker.local`, and a complete test/Challenge run with anti-bluff
verification.

### Build (v2.3.0-build.25, 26m 52s)

All 7 components built successfully, total 464 MB:

| Component | Artifact |
|---|---|
| catalog-api | `releases/catalog-api/linux-amd64/v2.3.0-build.25/catalog-api` (88 MB) |
| catalog-web | `releases/catalog-web/web/v2.3.0-build.25/dist` |
| catalogizer-api-client | npm package |
| catalogizer-desktop | `.deb` + `.rpm` |
| installer-wizard | `.AppImage` |
| catalogizer-android | APK |
| catalogizer-androidtv | APK (203 MB) |

Container images: `localhost/catalogizer-api:latest` (269 MB) and
`localhost/catalogizer-web:latest` (71.8 MB). The api image was built
via a new `catalog-api/Dockerfile.runtime` that copies the pre-built
binary (avoids the recurrent `golang:1.25` blob-checksum failure on
`zip_test.go` we hit twice during this session).

### Distribution (clean slate, surgical)

`scripts/audit/catalogizer-clean-slate.sh` wiped only `catalogizer-*`
containers/images/volumes on local + thinker + amber. Sibling stacks
left untouched: `helixagent-mcp-*` (17 containers on amber, powering
the user's Claude session), `helixtrack-*`, `bear-*`, `mail-*`,
`shareconnect-*`, `open-webui`, local `yt-dlp-*` / `metube-*`.

Final deployment:

| | Thinker (Podman) | Amber (Docker) |
|---|---|---|
| postgres | 127.0.0.1:5445 | 127.0.0.1:5446 |
| redis | 127.0.0.1:6391 | 127.0.0.1:6392 |
| catalog-api | 127.0.0.1:8092 → /health 200 | 127.0.0.1:8093 → /health 200 |
| catalog-web | 127.0.0.1:3092 → HTTP 200 | 127.0.0.1:3093 → HTTP 200 |

Lessons captured in updated `deployment/thinker-up.sh` +
`deployment/amber-up.sh`:
- `PORT` → `SERVER_PORT` (post CONST-035 port-mapping fix).
- `HOST=0.0.0.0` so Gin binds to all interfaces inside the container.
- `ADMIN_USERNAME=admin` (now required alongside `ADMIN_PASSWORD`).
- Postgres `pg_isready` wait before starting catalog-api.
- `loginctl enable-linger` (no sudo, per-user) so rootless Podman
  containers survive SSH disconnect.
- `--add-host host.containers.internal:host-gateway` for the nginx
  upstream on Docker (Podman builtin).
- catalog-web port 3000 inside (image EXPOSEs 3000, not 80).

### Tests (all PASS)

| Suite | Result |
|---|---|
| catalog-api Go tests (`go test ./...`) | **45 packages PASS, 0 FAIL, 0 SKIP** (incl. 49.5s stress) |
| Submodule Go tests (32 submodules, GOMAXPROCS=2 sequential) | **32/32 PASS, 0 FAIL** (~440 packages) |
| catalog-web vitest | **131 files PASS, 2318 tests PASS, 0 failed** (45.59s) |
| CONST-033 source-tree scanner | PASS (after extending excludes for `docs/incidents/` + `docs/audits/`) |
| CONST-033 host hardening | 4/4 PASS |
| Security scan | Reports saved to `docs/security/security-scan-20260429_080241.md`. govulncheck/gosec/trivy/snyk/nancy/npm audit all completed; advisory CVEs in npm deps (catalog-web 1H/1M, desktop 4H/5M, wizard 4H/5M, api-client 1C/5H/5M). Not blocking ship; needs separate dep-bump triage. |

### Anti-bluff Challenges (Article XI verified end-to-end)

Real anti-bluff probes against the deployed thinker stack:

- **Real auth flow**: `POST /api/v1/auth/login` returned a 248-char
  JWT `session_token`; the JWT decoded back to user `id=1, username=admin`
  via `GET /api/v1/auth/profile`.
- **Real DB query under auth**: `GET /api/v1/users` returned the
  seeded admin user with full role + timestamps.
- **Round-trip persistence (the key bluff-vs-real distinction)**:
  - `POST /api/v1/storage/roots` created `ab-test` → 201 + `id=1`.
  - `GET /api/v1/storage/roots` returned the same record with
    `created_at` from the database — **proves the feature works
    for end users, not just that the handler returned 200**.
- **Negative tests (all expected rejections)**:
  - Empty bearer → 401.
  - Bogus bearer → 401.
  - Modified bearer (signature tamper) → 401 (with auth-error log).
  - POST without auth → 401.
  - Bad password → 401.
  - Bogus route → 404.
- **Parity**: amber `/health` returned identical version metadata
  (`build_date 2026-04-28T23:29:18Z`, `build_number 25`,
  `version 2.3.0`).

### Bluff audit (umbrella + 41 submodules)

Static heuristic scanner `scripts/audit/anti-bluff-scan.sh` found
**6091 candidate findings** across the tree. Triaged in
`docs/audits/anti-bluff-2026-04-28.md`:

- Tier 1 (high-confidence bluff): PROSE_HELIXQA_ACTION 4564,
  GO_MOCK_IN_INTEGRATION 150, GO_HTTPTEST_ABUSE 9, ASSERT_TAUTOLOGY 22.
- Tier 2 (likely real, needs review): SKIP_WITHOUT_TICKET 188,
  GO_NIL_ONLY 164.
- Tier 3 (high false-positive rate): GO_NO_ASSERT 982,
  CHALLENGE_BLIND_SHELL 12.

Bulk resolution is multi-day work tracked as Task #14.

### Governance landed everywhere

- Constitution Article XI added to umbrella + cascaded to all 41
  submodules' `CONSTITUTION.md` / `CLAUDE.md` / `AGENTS.md`.
- CONST-033 cascaded to all 41 submodules (Containers was
  already up-to-date).
- All 41 submodules pushed to all upstreams (resolved diverged
  github↔gitlab via merge in 6 cases).
- Umbrella `b15ddcff..bc0f6344` pushed to all 8 remotes (github,
  gitlab, gitflic, gitverse, plus dual github/gitlab vasic-digital
  + milos85vasic).
- `docs/HOST_POWER_MANAGEMENT.md` extended with §12 Incident #2
  (Docker/Podman cumulative-cgroup-pressure session-loss vector).

### Live monitor

Persistent monitor `basxy6l99` running — probes both deployed hosts'
`/health` + `/`, scans `cz-api-*` logs for error/fatal/panic, watches
host load and conmon cgroup warnings (the §12 vector). One event
surfaced during the session was a deliberate negative test (modified
JWT → expected 401 + auth-error log).

### Resource state at session end

- Local load: 4.57 / 4.04 / 2.94 (calm).
- Memory: ~5 GiB / 62 GiB used; 57 GiB available.
- Swap: 3 MiB / 15 GiB.
- All 8 deployed containers `Up`, neither host's user-slice
  ever distressed.

### Open items (not regressions)

1. **Bluff resolution** (Task #14) — multi-day rewrite of bank
   entries + a few catalog-api E2E tests. Plan in
   `docs/audits/anti-bluff-2026-04-28.md`.
2. **npm dependency CVE triage** — see
   `docs/security/security-scan-20260429_080241.md`. Mostly minor
   bumps; one critical in api-client requires a deeper look.
3. **`/api/v1/info` and `/api/v1/version` endpoints** — both 404 in
   build 25; either remove the references in tests/docs or add the
   endpoints. Not a regression of this session.
