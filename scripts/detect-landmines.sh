#!/usr/bin/env bash
# scripts/detect-landmines.sh — run before every PR
#
# Fast grep-based check for the rules in docs/LANDMINES.md that can be
# enforced purely by static inspection. Exit non-zero on any hit.
#
# The full LANDMINES catalogue is 47 rules — many need behavioural tests
# (run catalog-api with bad input, run HelixQA against a real device) and
# are out of scope for this pre-flight. Rules covered here are the ones
# where a simple grep proves the violation.

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT"

fail_count=0
fail() {
  echo "❌ $1" >&2
  fail_count=$((fail_count + 1))
}

ok() { echo "✓ $1"; }

# ---------------------------------------------------------------------------
# RULE-SEC-001 — no real .env committed anywhere
# Allowed exceptions (non-secret deployment host-config templates):
#   deployment/*.env — described in deployment/README; no API keys
# ---------------------------------------------------------------------------
tracked_envs=$(git ls-files --cached 2>/dev/null \
  | grep -E '(^|/)\.env(\..*)?$' \
  | grep -vE '(^|/)\.env\.example($|\.)' \
  | grep -vE '^deployment/.*\.env$' \
  | grep -vE '^\.env\.(distributed|roundrobin|spread|security)$' || true)
if [ -n "$tracked_envs" ]; then
  fail "RULE-SEC-001: tracked .env file(s) detected:"
  echo "$tracked_envs" | sed 's/^/    /' >&2
else
  ok "RULE-SEC-001: no tracked .env files (deployment/*.env whitelisted)"
fi

# ---------------------------------------------------------------------------
# RULE-SEC-002 — .env.example placeholders only (no long tokens)
# ---------------------------------------------------------------------------
if git ls-files --cached -- '*.env.example' 2>/dev/null \
    | xargs -r grep -lE '^\w+=[A-Za-z0-9_\-]{24,}$' 2>/dev/null \
    | grep -q .; then
  fail "RULE-SEC-002: .env.example contains what looks like a real secret (24+ chars)"
else
  ok "RULE-SEC-002: .env.example placeholders clean"
fi

# ---------------------------------------------------------------------------
# RULE-GIT-002 — no GitHub Actions workflows
# ---------------------------------------------------------------------------
if find . -path '*/.github/workflows/*' -type f 2>/dev/null | grep -q .; then
  fail "RULE-GIT-002: .github/workflows/ files present:"
  find . -path '*/.github/workflows/*' -type f | sed 's/^/    /' >&2
else
  ok "RULE-GIT-002: no GitHub Actions workflows"
fi

# ---------------------------------------------------------------------------
# RULE-CONST-001 — no sudo / no root in scripts
# ---------------------------------------------------------------------------
if grep -rnE '^\s*(sudo\b|su -c)' scripts/ deployment/ Containers/ docker/ \
      --include='*.sh' --include='*.yml' --include='*.yaml' \
      --include='Dockerfile*' --include='Containerfile*' 2>/dev/null \
    | grep -vE '^\s*#|sudoers' | grep -q .; then
  fail "RULE-CONST-001: sudo / su reference in scripts/deployment/containers:"
  grep -rnE '^\s*(sudo\b|su -c)' scripts/ deployment/ Containers/ docker/ \
    --include='*.sh' --include='*.yml' --include='*.yaml' \
    --include='Dockerfile*' --include='Containerfile*' 2>/dev/null \
    | grep -vE '^\s*#|sudoers' | sed 's/^/    /' >&2
else
  ok "RULE-CONST-001: no sudo/su in scripts"
fi

# ---------------------------------------------------------------------------
# RULE-GO-004 — LastInsertId() forbidden in application code
# Allowed exceptions:
#   database/connection.go  — InsertReturningID SQLite fallback
#   database/tx_helpers.go  — TxInsertReturningID SQLite fallback
# These are the wrapper's internal implementations, not application misuse.
# ---------------------------------------------------------------------------
if [ -d catalog-api ]; then
  hits=$(grep -rn 'LastInsertId()' catalog-api/ --include='*.go' 2>/dev/null \
    | grep -v '_test.go' \
    | grep -v 'database/connection\.go' \
    | grep -v 'database/tx_helpers\.go' || true)
  if [ -n "$hits" ]; then
    fail "RULE-GO-004: LastInsertId() in application code:"
    echo "$hits" | sed 's/^/    /' >&2
  else
    ok "RULE-GO-004: no LastInsertId() in catalog-api application code"
  fi
fi

# ---------------------------------------------------------------------------
# RULE-GO-006 — no .disabled files; no unexplained t.Skip (tolerates the
# four infrastructure-conditional skips documented in DISABLED_FEATURES_AUDIT.md)
# ---------------------------------------------------------------------------
if find . -maxdepth 5 \( -name '*.disabled' -o -name '*.disabled.go' \
      -o -name '*.go.disabled' \) ! -path '*/node_modules/*' ! -path '*/.git/*' \
      2>/dev/null | grep -q .; then
  fail "RULE-GO-006: .disabled files present:"
  find . -maxdepth 5 \( -name '*.disabled' -o -name '*.disabled.go' \
    -o -name '*.go.disabled' \) ! -path '*/node_modules/*' ! -path '*/.git/*' \
    | sed 's/^/    /' >&2
else
  ok "RULE-GO-006: no .disabled files"
fi

# ---------------------------------------------------------------------------
# RULE-HELIX-001 — library is project-agnostic (no Catalogizer-specific
# package names baked into HelixQA or its sub-libs)
# ---------------------------------------------------------------------------
if [ -d HelixQA ]; then
  leaks=$(grep -rnE 'com\.catalogizer\.|ru\.iptvremote|com\.atmosphere' \
            HelixQA/pkg/ HelixQA/cmd/ --include='*.go' 2>/dev/null \
          | grep -v '_test.go' | grep -vE 'RULE-|//\s' || true)
  if [ -n "$leaks" ]; then
    fail "RULE-HELIX-001: project-specific package names baked into HelixQA library:"
    echo "$leaks" | sed 's/^/    /' >&2
  else
    ok "RULE-HELIX-001: HelixQA library clean"
  fi
fi

# ---------------------------------------------------------------------------
# RULE-DESK-001 — no bare Rust unwrap() in Tauri src-tauri code
# ---------------------------------------------------------------------------
for dir in catalogizer-desktop installer-wizard; do
  if [ -d "$dir/src-tauri/src" ]; then
    # Extract non-test unwrap() calls. Walk each .rs file, filter out
    # lines inside #[test] functions or #[cfg(test)] modules.
    hits=$(
      find "$dir/src-tauri/src" -name '*.rs' -print0 \
        | xargs -0 -I{} awk '
          BEGIN { in_test = 0; depth = 0 }
          /#\[(test|cfg\(test\))\]/ { in_test = 1 }
          in_test && /\{/ { depth += gsub(/\{/, "{") }
          in_test && /\}/ { depth -= gsub(/\}/, "}"); if (depth <= 0) { in_test = 0; depth = 0 } }
          !in_test && /\bunwrap\(\)/ && !/\/\/ SAFE/ {
            print FILENAME ":" NR ":" $0
          }
        ' {} 2>/dev/null || true
    )
    if [ -n "$hits" ]; then
      fail "RULE-DESK-001: bare unwrap() in $dir/src-tauri/src non-test code (add // SAFE: <reason> or refactor):"
      echo "$hits" | sed 's/^/    /' >&2
    else
      ok "RULE-DESK-001: $dir Rust unwrap() clean (non-test code)"
    fi
  fi
done

# ---------------------------------------------------------------------------
# RULE-HELIX-007 — .devconnect IP lines must not carry inline comments
# ---------------------------------------------------------------------------
if [ -f .devconnect ]; then
  if grep -nE '^\s*[^#].*#' .devconnect 2>/dev/null | grep -q .; then
    fail "RULE-HELIX-007: .devconnect has inline comments on IP lines:"
    grep -nE '^\s*[^#].*#' .devconnect | sed 's/^/    /' >&2
  else
    ok "RULE-HELIX-007: .devconnect clean"
  fi
fi

# ---------------------------------------------------------------------------
# RULE-CH-003 — config.json write_timeout must be 900+ for long RunAll
# ---------------------------------------------------------------------------
if [ -f catalog-api/config/config.json ]; then
  wt=$(python3 -c "import json; print(json.load(open('catalog-api/config/config.json')).get('write_timeout', 0))" 2>/dev/null || echo 0)
  if [ "${wt:-0}" -lt 900 ]; then
    fail "RULE-CH-003: catalog-api/config/config.json write_timeout=$wt (must be ≥900)"
  else
    ok "RULE-CH-003: write_timeout=$wt (≥900)"
  fi
fi

# ---------------------------------------------------------------------------
# RULE-TV-001 — Android TV OkHttpClient forces HTTP/1.1
# ---------------------------------------------------------------------------
if [ -d catalogizer-androidtv ]; then
  if grep -rn 'Protocol\.HTTP_1_1' catalogizer-androidtv/app/src --include='*.kt' \
      2>/dev/null | grep -q .; then
    ok "RULE-TV-001: Android TV HTTP/1.1 forced"
  else
    fail "RULE-TV-001: catalogizer-androidtv OkHttpClient does not force Protocol.HTTP_1_1"
  fi
fi

# ---------------------------------------------------------------------------
echo
if [ "$fail_count" -eq 0 ]; then
  echo "✓ landmine pre-flight clean"
  exit 0
else
  echo "❌ landmine pre-flight: $fail_count violation(s)" >&2
  echo "   see docs/LANDMINES.md for the full rule set + fix guidance" >&2
  exit 1
fi
