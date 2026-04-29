#!/usr/bin/env bash
# cascade-user-mandate.sh
#
# Idempotently appends the Article XI §11.9 user-mandate forensic
# anchor to a submodule's CONSTITUTION.md, CLAUDE.md, and AGENTS.md
# if not already present. Per Article XI §11.9 the anchor MUST
# appear in every submodule's governance files.
#
# Usage:
#   cascade-user-mandate.sh <submodule-path>

set -uo pipefail
SUBMODULE="${1:-}"
[ -d "$SUBMODULE" ] || { echo "usage: $0 <submodule>" >&2; exit 2; }
cd "$SUBMODULE" || exit 2

ANCHOR_BLOCK='
<!-- BEGIN user-mandate forensic anchor (Article XI §11.9) -->

## ⚠️ User-Mandate Forensic Anchor (Article XI §11.9 — 2026-04-29)

Inherited from the umbrella project. Verbatim user mandate:

> "We had been in position that all tests do execute with success
> and all Challenges as well, but in reality the most of the
> features does not work and can'"'"'t be used! This MUST NOT be the
> case and execution of tests and Challenges MUST guarantee the
> quality, the completion and full usability by end users of the
> product!"

**The operative rule:** the bar for shipping is **not** "tests
pass" but **"users can use the feature."**

Every PASS in this codebase MUST carry positive evidence captured
during execution that the feature works for the end user. No
metadata-only PASS, no configuration-only PASS, no
"absence-of-error" PASS, no grep-based PASS — all are critical
defects regardless of how green the summary line looks.

Tests and Challenges (HelixQA) are bound equally. A Challenge that
scores PASS on a non-functional feature is the same class of
defect as a unit test that does.

**No false-success results are tolerable.** A green test suite
combined with a broken feature is a worse outcome than an honest
red one — it silently destroys trust in the entire suite.

Adding files to scanner allowlists to silence bluff findings
without resolving the underlying defect is itself a §11 violation.

**Full text:** umbrella `CONSTITUTION.md` Article XI §11.9.

<!-- END user-mandate forensic anchor (Article XI §11.9) -->
'

modified=0
for f in CONSTITUTION.md CLAUDE.md AGENTS.md; do
  [ -f "$f" ] || continue
  if grep -q "BEGIN user-mandate forensic anchor (Article XI §11.9)" "$f"; then
    continue
  fi
  printf '%s' "$ANCHOR_BLOCK" >> "$f"
  echo "  $f: +§11.9 anchor"
  modified=$((modified+1))
done

if [ "$modified" -eq 0 ]; then
  echo "  (already up to date)"
fi
echo "modified $modified file(s) in $SUBMODULE"
exit 0
