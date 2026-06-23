#!/usr/bin/env bash
# scripts/hooks/pre-push-gate.sh — LLM-as-Judge pre-push wire-up
#
# Runs before every `git push` to main. Chains:
#   1. scripts/detect-landmines.sh      (hard gate — exits non-zero on violation)
#   1b. scripts/audit/anti-bluff-scan.sh (hard gate — §11.4 anti-bluff CI lane;
#       exits non-zero on ANY in-scope bluff finding so the push is BLOCKED.
#       This is the LOCAL gate mandated by §11.4.75 — remote CI is forbidden
#       by §11.4.156, so the anti-bluff lane runs here before every push.)
#   2. Generates /tmp/judge_prompt.md from templates/LLM_JUDGE_PREMERGE.md
#      + the current PR diff + docs/LANDMINES.md + docs/API_CONTRACTS.md
#   3. Prints instructions for the operator to submit the prompt to a
#      reasoning LLM (Claude Code, OpenCode headless, etc.) and review
#      the JSON verdict before proceeding with the push
#
# Master Plan v2 §7.3 Verification Loop + §8.3 LLM-as-Judge template.
#
# To enable as a git hook:
#   ln -sf ../../scripts/hooks/pre-push-gate.sh .git/hooks/pre-push
#   chmod +x scripts/hooks/pre-push-gate.sh
#
# Bypass (emergency only — files a post-push retrospective):
#   LLM_JUDGE_BYPASS=1 git push origin main

set -uo pipefail  # do NOT set -e — we want to capture landmine exit code

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT"

# -----------------------------------------------------------------------------
# Gate 1 — landmine pre-flight (hard)
# -----------------------------------------------------------------------------
echo "[pre-push-gate] 1/3 — scripts/detect-landmines.sh"
# NOTE: capture the command's own exit code DIRECTLY — `if ! cmd; then rc=$?`
# captures the negated-test result (always 0), not the command's exit code,
# which would launder a real failure into a 0 exit (push allowed despite the
# "refusing" message). §11.4.1 FAIL-bluff fix.
bash scripts/detect-landmines.sh; rc=$?
if [ "$rc" -ne 0 ]; then
  if [ "${LLM_JUDGE_BYPASS:-0}" = "1" ]; then
    echo "[pre-push-gate] LANDMINE VIOLATION but LLM_JUDGE_BYPASS=1 — continuing (emergency bypass)"
  else
    echo
    echo "[pre-push-gate] ❌ LANDMINE violation — refusing to push"
    echo "   Fix the flagged rule(s) above (see docs/LANDMINES.md) then retry."
    echo "   Emergency bypass: LLM_JUDGE_BYPASS=1 git push origin main"
    exit "$rc"
  fi
fi

# -----------------------------------------------------------------------------
# Gate 1b — anti-bluff scanner (hard, §11.4 / §11.4.75 LOCAL gate)
# -----------------------------------------------------------------------------
# The static anti-bluff scanner is the §11.4 covenant's mechanical CI lane.
# Per §11.4.156 server-side CI is disabled, so the lane MUST run locally before
# every push. The scanner exits non-zero on ANY in-scope bluff finding; a
# non-zero exit here BLOCKS the push (clean tree => exit 0 per its hardened
# contract). Same emergency-bypass switch as the landmine gate.
echo "[pre-push-gate] 1b/3 — scripts/audit/anti-bluff-scan.sh"
# Capture the scanner's own exit code directly (same §11.4.1 fix as Gate 1).
bash scripts/audit/anti-bluff-scan.sh; rc=$?
if [ "$rc" -ne 0 ]; then
  if [ "${LLM_JUDGE_BYPASS:-0}" = "1" ]; then
    echo "[pre-push-gate] ANTI-BLUFF finding(s) but LLM_JUDGE_BYPASS=1 — continuing (emergency bypass)"
  else
    echo
    echo "[pre-push-gate] ❌ ANTI-BLUFF finding(s) — refusing to push"
    echo "   Each TSV line above is file<TAB>line<TAB>kind<TAB>excerpt."
    echo "   Fix the flagged test/challenge/bank, or add the scanner-honored"
    echo "   inline marker ('# bluff-scan: ok (<why>)' / 'SKIP-OK: §11.4.3 <reason>')"
    echo "   for legitimate patterns, then retry."
    echo "   Emergency bypass: LLM_JUDGE_BYPASS=1 git push origin main"
    exit "$rc"
  fi
fi

# -----------------------------------------------------------------------------
# Gate 2 — generate judge prompt
# -----------------------------------------------------------------------------
echo "[pre-push-gate] 2/3 — generating LLM judge prompt"

PROMPT=/tmp/judge_prompt.md
DIFF=/tmp/pr.diff

# Collect the diff relative to main (skip if pushing main itself; then
# compare against origin/main).
if [ "$(git symbolic-ref --short HEAD 2>/dev/null)" = "main" ]; then
  git fetch origin main --quiet 2>/dev/null || true
  git diff origin/main...HEAD > "$DIFF"
else
  git diff main...HEAD > "$DIFF"
fi
diff_lines=$(wc -l < "$DIFF" 2>/dev/null || echo 0)

if [ "$diff_lines" -eq 0 ]; then
  echo "[pre-push-gate] empty diff — skipping judge prompt generation"
  exit 0
fi

# Assemble the prompt from the template + diff + landmines + contracts.
{
  cat templates/LLM_JUDGE_PREMERGE.md
  printf '\n<<<DIFF\n'
  cat "$DIFF"
  printf '\nDIFF\n'
  printf '\n<<<LANDMINES\n'
  cat docs/LANDMINES.md
  printf '\nLANDMINES\n'
  # API_CONTRACTS is large — include only when the diff touches
  # catalog-api handlers/services/repository
  if grep -qE '^(\+\+\+|---) .*catalog-api/(handlers|services|repository|internal)/' "$DIFF"; then
    printf '\n<<<CONTRACTS\n'
    cat docs/API_CONTRACTS.md
    printf '\nCONTRACTS\n'
  fi
} > "$PROMPT"

prompt_lines=$(wc -l < "$PROMPT")
echo "[pre-push-gate] judge prompt assembled: $PROMPT ($prompt_lines lines, $diff_lines lines of diff)"

# -----------------------------------------------------------------------------
# Gate 3 — operator instructions
# -----------------------------------------------------------------------------
echo "[pre-push-gate] 3/3 — LLM judge review"
echo
echo "BEFORE YOUR PUSH COMPLETES:"
echo "  Submit the prompt at $PROMPT to a reasoning LLM."
echo "  Examples:"
echo "    claude-code --prompt-file $PROMPT --format json > /tmp/verdict.json"
echo "    opencode --headless --prompt-file $PROMPT --format json > /tmp/verdict.json"
echo
echo "  Then check the verdict:"
echo '    veto=$(jq -r .veto /tmp/verdict.json)'
echo '    [ "$veto" = "false" ] && echo OK || echo "BLOCKER — see /tmp/verdict.json"'
echo
if [ "${LLM_JUDGE_INTERACTIVE:-0}" = "1" ] && command -v claude > /dev/null; then
  echo "LLM_JUDGE_INTERACTIVE=1 — running claude --headless now..."
  claude --headless --prompt-file "$PROMPT" --format json > /tmp/verdict.json 2>&1 || true
  if [ -s /tmp/verdict.json ]; then
    veto=$(jq -r '.veto // "unknown"' /tmp/verdict.json 2>/dev/null || echo unknown)
    if [ "$veto" = "true" ]; then
      echo "[pre-push-gate] ❌ LLM judge VETO — see /tmp/verdict.json"
      cat /tmp/verdict.json
      exit 1
    fi
    echo "[pre-push-gate] ✓ LLM judge cleared — veto=$veto"
  fi
fi

echo "[pre-push-gate] ✓ passed — push proceeds"
exit 0
