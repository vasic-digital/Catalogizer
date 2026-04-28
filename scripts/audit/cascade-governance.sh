#!/usr/bin/env bash
# cascade-governance.sh
#
# Idempotently appends the CONST-033 (host-power-management hard-ban) and
# Article XI (anti-bluff testing) addenda to a submodule's CONSTITUTION.md,
# CLAUDE.md, and AGENTS.md if they are not already present.
#
# Usage:
#   cascade-governance.sh <submodule-path>
#
# Exit codes:
#   0  one or more files modified, OR all already present (idempotent OK)
#   1  required files missing
#   2  invocation error

set -uo pipefail

SUBMODULE="${1:-}"
if [ -z "$SUBMODULE" ] || [ ! -d "$SUBMODULE" ]; then
  echo "usage: $0 <submodule-path>" >&2
  exit 2
fi

cd "$SUBMODULE" || exit 2

# ---- CONST-033 addenda (one per file) ----
CONST033_FOR_CONSTITUTION='
<!-- BEGIN host-power-management addendum (CONST-033) -->

## CONST-033 — Host Power Management is Forbidden

**Status:** Mandatory. Non-negotiable. Inherited from the umbrella
project (`Catalogizer/CONSTITUTION.md` CONST-033).

**Rule:** No code in this submodule may invoke a host-level
power-state transition (suspend, hibernate, hybrid-sleep,
suspend-then-hibernate, poweroff, halt, reboot, kexec) on the host
machine. Forbidden invocations include — but are not limited to:

- `systemctl {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}`
- `loginctl {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}`
- `pm-{suspend,hibernate,suspend-hybrid}`
- `shutdown {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}`
- DBus / busctl calls to `org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}`
- DBus / busctl calls to `org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}`
- `gsettings set ... sleep-inactive-{ac,battery}-type` to any value other than `'"'"'nothing'"'"'` or `'"'"'blank'"'"'`

**Why:** The host runs mission-critical parallel CLI-agent and
container workloads. On 2026-04-26 18:23:43 the host was auto-suspended
mid-session, killing HelixAgent + 41 dependent services. On 2026-04-28
18:36:35 the user-slice was SIGKILLed under cumulative cgroup pressure
(see `docs/incidents/` in the umbrella). Both classes of event are
session-loss; both are now defended in depth.

**Defence in depth (umbrella project artifacts):**

1. `scripts/host-power-management/install-host-suspend-guard.sh`
2. `scripts/host-power-management/user_session_no_suspend_bootstrap.sh`
3. `scripts/host-power-management/check-no-suspend-calls.sh`
4. `challenges/scripts/host_no_auto_suspend_challenge.sh`
5. `challenges/scripts/no_suspend_calls_challenge.sh`

**Enforcement:** the umbrella project'"'"'s CI / `run_all_challenges.sh`
runs both challenges (host state + source tree). A violation in
either channel blocks merge. Adding files to the scanner'"'"'s
`EXCLUDE_PATHS` requires an explicit justification comment
identifying the non-host context.

**Cross-references:** umbrella `docs/HOST_POWER_MANAGEMENT.md`,
umbrella `CONSTITUTION.md` CONST-033 + Article X (no-sudo).

<!-- END host-power-management addendum (CONST-033) -->
'

ARTICLE_XI_ADDENDUM='
<!-- BEGIN anti-bluff-testing addendum (Article XI) -->

## Article XI — Anti-Bluff Testing (MANDATORY)

**Inherited from the umbrella project'"'"'s Constitution Article XI.
Tests and Challenges that pass without exercising real end-user
behaviour are forbidden in this submodule too.**

Every test, every Challenge, every HelixQA bank entry MUST:

1. **Assert on a concrete end-user-visible outcome** — rendered DOM,
   DB rows that a real query would return, files on disk, media that
   actually plays, search results that actually contain expected
   items. Not "no error" or "200 OK".
2. **Run against the real system below the assertion.** Mocks/stubs
   are permitted ONLY in unit tests (`*_test.go` under `go test
   -short` or language equivalent). Integration / E2E / Challenge /
   HelixQA tests use real containers, real databases, real
   renderers. Unreachable real-system → skip with `SKIP-OK:
   #<ticket>`, never silently pass.
3. **Include a matching negative.** Every positive assertion is
   paired with an assertion that fails when the feature is broken.
4. **Emit copy-pasteable evidence** — body, screenshot, frame, DB
   row, log excerpt. Boolean pass/fail is insufficient.
5. **Verify "fails when feature is removed."** Author runs locally
   with the feature commented out; the test MUST FAIL. If it still
   passes, it'"'"'s a bluff — delete and rewrite.
6. **No blind shells.** No `&& echo PASS`, `|| true`, `tee` exit
   laundering, `if [ -f file ]` without content assertion.

**Challenges in this submodule** must replay the user journey
end-to-end through the umbrella project'"'"'s deliverables — never via
raw `curl` or third-party scripts. Sub-1-second Challenges almost
always indicate a bluff.

**HelixQA banks** declare executable actions
(`adb_shell:`, `playwright:`, `http:`, `assertVisible:`,
`assertNotVisible:`), never prose. Stagnation guard from Article I
§1.3 applies — frame N+1 identical to frame N for >10 s = FAIL.

**PR requirement:** every PR adding/modifying a test or Challenge in
this submodule MUST include a fenced `## Anti-Bluff Verification`
block with: (a) command run, (b) pasted output, (c) proof the test
fails when the feature is broken (second run with feature
commented-out showing FAIL).

**Cross-reference:** umbrella `CONSTITUTION.md` Article XI
(§§ 11.1 — 11.8).

<!-- END anti-bluff-testing addendum (Article XI) -->
'

modified=0
for f in CONSTITUTION.md CLAUDE.md AGENTS.md; do
  [ -f "$f" ] || continue
  added=""
  # CONST-033
  if ! grep -q "CONST-033 — Host Power Management is Forbidden" "$f"; then
    if ! grep -q "BEGIN host-power-management addendum (CONST-033)" "$f"; then
      printf '%s' "$CONST033_FOR_CONSTITUTION" >> "$f"
      added="${added}+CONST-033 "
      modified=$((modified+1))
    fi
  fi
  # Article XI
  if ! grep -qE "Article XI — Anti-Bluff Testing|Anti-Bluff Testing — Mandatory|Anti-Bluff Testing \(MANDATORY\)|MANDATORY ANTI-BLUFF VALIDATION|MANDATORY ANTI-BLUFF COVENANT" "$f"; then
    if ! grep -q "BEGIN anti-bluff-testing addendum (Article XI)" "$f"; then
      printf '%s' "$ARTICLE_XI_ADDENDUM" >> "$f"
      added="${added}+ArtXI "
      modified=$((modified+1))
    fi
  fi
  if [ -n "$added" ]; then
    echo "  $f: $added"
  fi
done

if [ "$modified" -eq 0 ]; then
  echo "  (already up to date)"
fi
echo "modified $modified file(s) in $SUBMODULE"
exit 0
