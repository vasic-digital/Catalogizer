#!/usr/bin/env bash
#
# push_all_submodules.sh — Commit the Helix Constitution inheritance pointer in
# every OWNED submodule and push each to ALL its configured upstream remotes.
#
# ─────────────────────────────────────────────────────────────────────────────
# WHAT THIS DOES (per owned submodule under submodules/*)
#   1. Stage ONLY the governance pointer files (CLAUDE.md, AGENTS.md) — never
#      `git add -A` (avoids sweeping nested third-party submodule churn, e.g.
#      inside submodules/helix_qa).
#   2. Commit them (if staged) with a descriptive message + co-author footer.
#   3. fetch --all, then integrate the latest remote tip with `merge --ff-only`
#      onto the current branch (§11.4.113 — NEVER force-push; if the branch has
#      DIVERGED, the submodule is SKIPPED for manual merge, never forced).
#   4. Push the branch to EVERY configured remote (§2.1 multi-upstream).
#   5. Capture a per-(submodule,remote) PASS/FAIL result to a log + summary.
#
# SAFETY: DRY_RUN=1 (default) prints the plan and changes nothing. DRY_RUN=0
# executes. No `--force`, no `--no-verify`. Per-submodule failures do not abort
# the run — they are reported at the end (so one unreachable mirror cannot block
# the rest).
#
# Usage:
#   Preview (default):   scripts/push_all_submodules.sh
#   Execute:             DRY_RUN=0 scripts/push_all_submodules.sh
#   Limit to a subset:   ONLY="submodules/auth submodules/cache" DRY_RUN=0 scripts/push_all_submodules.sh
# ─────────────────────────────────────────────────────────────────────────────

set -uo pipefail   # intentionally NOT -e: we continue past per-repo failures.

DRY_RUN="${DRY_RUN:-1}"
ONLY="${ONLY:-}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

LOG="${REPO_ROOT}/qa-results/phasef_push.log"
mkdir -p "$(dirname "${LOG}")"
: > "${LOG}"

c_reset=$'\033[0m'; c_blue=$'\033[34m'; c_green=$'\033[32m'; c_yellow=$'\033[33m'; c_red=$'\033[31m'; c_dim=$'\033[2m'
log()  { printf '%s[push]%s %s\n' "${c_blue}"   "${c_reset}" "$*"; }
ok()   { printf '%s[ ok ]%s %s\n' "${c_green}"  "${c_reset}" "$*"; }
warn() { printf '%s[warn]%s %s\n' "${c_yellow}" "${c_reset}" "$*"; }
err()  { printf '%s[fail]%s %s\n' "${c_red}"    "${c_reset}" "$*" >&2; }
act()  { if [ "${DRY_RUN}" = "1" ]; then printf '%s[plan]%s WOULD %s\n' "${c_yellow}" "${c_reset}" "$*"; else printf '%s[ do ]%s %s\n' "${c_green}" "${c_reset}" "$*"; fi; }

COMMIT_MSG="chore(constitution): inherit Helix Constitution; submodules/ layout

Add the project-agnostic Helix Constitution inheritance pointer to
CLAUDE.md + AGENTS.md (references find_constitution.sh; no hardcoded
path, so this module stays fully decoupled per §11.4.28).

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"

# Collect owned submodule paths (submodules/* incl. constitution; constitution
# is idempotent — already committed/pushed, so its commit step is a no-op).
mapfile -t SUBS < <(git config -f .gitmodules --get-regexp '^submodule\..*\.path$' \
  | awk '{print $2}' | grep '^submodules/' | sort)

if [ -n "${ONLY}" ]; then
  # filter to the requested subset
  filtered=()
  for p in "${SUBS[@]}"; do for q in ${ONLY}; do [ "$p" = "$q" ] && filtered+=("$p"); done; done
  SUBS=("${filtered[@]}")
fi

log "Repo root: ${REPO_ROOT}"
[ "${DRY_RUN}" = "1" ] && warn "DRY_RUN=1 (default) — NO commits/pushes. Set DRY_RUN=0 to execute." \
                       || warn "DRY_RUN=0 — commits + pushes WILL happen."
log "Owned submodules to process: ${#SUBS[@]}"

declare -a PUSH_OK=() PUSH_FAIL=() SKIPPED=() COMMITTED=()

for p in "${SUBS[@]}"; do
  echo
  log "──────── ${p} ────────"
  if [ ! -e "${p}/.git" ]; then warn "no git dir, skipping: ${p}"; SKIPPED+=("${p} (no .git)"); continue; fi

  br="$(git -C "${p}" symbolic-ref --short -q HEAD || true)"
  if [ -z "${br}" ]; then warn "DETACHED HEAD, skipping (needs a branch): ${p}"; SKIPPED+=("${p} (detached)"); continue; fi
  log "branch: ${br}"

  # 1+2. stage ALL decoupling/pointer/reorg changes, but NEVER the nested
  # third-party gitlink drift (pre-existing) under tools/opensource|external
  # (e.g. inside submodules/helix_qa). add -A then unstage the vendored trees.
  git -C "${p}" add -A
  git -C "${p}" restore --staged -- tools/opensource tools/external 2>/dev/null || true
  if ! git -C "${p}" diff --cached --quiet; then
    act "commit governance pointer in ${p}"
    if [ "${DRY_RUN}" = "0" ]; then
      if git -C "${p}" commit -q -m "${COMMIT_MSG}"; then ok "committed ${p}"; COMMITTED+=("${p}"); else err "commit failed ${p}"; fi
    else
      COMMITTED+=("${p}")
    fi
  else
    log "nothing to commit (pointer already committed or absent)"
  fi

  # 3. fetch + merge-onto-latest origin/<branch> (§11.4.113 — take latest, merge
  #    our changes on top, NEVER force; on real conflict abort + skip for manual).
  act "fetch --all --prune ${p}"
  if [ "${DRY_RUN}" = "0" ]; then
    git -C "${p}" fetch --all --prune --tags --quiet 2>>"${LOG}" || warn "fetch had issues ${p} (see log)"
    if git -C "${p}" rev-parse --verify -q "origin/${br}" >/dev/null; then
      behind="$(git -C "${p}" rev-list --count "HEAD..origin/${br}" 2>/dev/null || echo 0)"
      if [ "${behind}" != "0" ]; then
        act "merge-onto-latest origin/${br} (behind ${behind}) ${p}"
        if ! git -C "${p}" merge --no-edit "origin/${br}" >>"${LOG}" 2>&1; then
          git -C "${p}" merge --abort >>"${LOG}" 2>&1 || true
          err "MERGE CONFLICT with origin/${br}; SKIP for manual resolution (no force, §11.4.113): ${p}"
          SKIPPED+=("${p} (merge conflict)"); continue
        fi
      fi
    fi
  fi

  # 4. push to every configured remote, skipping no-op pushes (already up-to-date)
  #    — this both honours "push to all upstreams" AND verifies the unchanged
  #    submodules are genuinely present on every remote.
  for r in $(git -C "${p}" remote); do
    if [ "${DRY_RUN}" = "0" ]; then
      # if the remote already has our HEAD, it's verified-present — skip the no-op push
      ahead="$(git -C "${p}" rev-list --count "${r}/${br}..HEAD" 2>/dev/null || echo "unknown")"
      if [ "${ahead}" = "0" ]; then ok "up-to-date (verified present) ${p} → ${r}"; PUSH_OK+=("${p}→${r} (verified)"); continue; fi
      act "push ${p} → ${r} ${br} (ahead ${ahead})"
      if git -C "${p}" push "${r}" "${br}" >>"${LOG}" 2>&1; then ok "pushed ${p} → ${r}"; PUSH_OK+=("${p}→${r}");
      else err "push FAILED ${p} → ${r} (see ${LOG})"; PUSH_FAIL+=("${p}→${r}"); fi
    else
      act "push-or-verify ${p} → ${r} ${br}"
    fi
  done
done

echo
log "════════════════ SUMMARY ════════════════"
log "submodules processed: ${#SUBS[@]}"
log "committed:            ${#COMMITTED[@]}"
ok  "push OK:              ${#PUSH_OK[@]}"
[ "${#PUSH_FAIL[@]}" -gt 0 ] && err "push FAILED:          ${#PUSH_FAIL[@]}" || log "push FAILED:          0"
[ "${#SKIPPED[@]}"  -gt 0 ] && warn "skipped:              ${#SKIPPED[@]}"
if [ "${#PUSH_FAIL[@]}" -gt 0 ]; then echo; err "Failed pushes:"; printf '   %s\n' "${PUSH_FAIL[@]}" >&2; fi
if [ "${#SKIPPED[@]}"  -gt 0 ]; then echo; warn "Skipped:"; printf '   %s\n' "${SKIPPED[@]}"; fi
echo
log "Full git output: ${LOG}"
[ "${#PUSH_FAIL[@]}" -eq 0 ]
