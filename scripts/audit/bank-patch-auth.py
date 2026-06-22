#!/usr/bin/env python3
"""
bank-patch-auth.py — patch already-converted HelixQA banks to
inject `auth: "admin"` for non-public HTTP endpoints that are
missing the field. Closes BLUFF-FQA-API-AUTH-INJECT-001.

Operates on the JSON form: walks every test_case's steps, and
for each step where action starts with "http: " AND no `auth`
key is set AND the path is not on the PUBLIC_ENDPOINT_PREFIXES
list, sets auth to "admin".

Idempotent: re-running on a patched bank is a no-op.

Usage:
  scripts/audit/bank-patch-auth.py submodules/helix_qa/banks/full-qa-api.json
"""

import json
import re
import sys

PUBLIC_ENDPOINT_PREFIXES = (
    "/health",
    "/api/v1/health",
    "/api/v1/auth/login",
    "/api/v1/auth/register",
    "/api/v1/auth/refresh",
    "/api/v1/auth/logout",
    "/api/v1/discovery",
    "/metrics",
)


def is_public(path: str) -> bool:
    p = path.split("?")[0].split("#")[0]
    return any(p == pref or p.startswith(pref + "/") for pref in PUBLIC_ENDPOINT_PREFIXES)


def patch_step(step: dict) -> bool:
    """Return True if step was modified."""
    action = step.get("action", "")
    if not isinstance(action, str) or not action.startswith("http:"):
        return False
    if step.get("auth"):
        return False  # already has auth
    # extract path: "http: METHOD /path"
    parts = action.split(maxsplit=2)
    if len(parts) < 3:
        return False
    path = parts[2]
    if is_public(path):
        return False
    step["auth"] = "admin"
    return True


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: bank-patch-auth.py <bank.json>", file=sys.stderr)
        return 2
    path = sys.argv[1]
    with open(path) as f:
        data = json.load(f)

    patched = 0
    total_http = 0
    for tc in data.get("test_cases", []):
        for step in tc.get("steps", []):
            if isinstance(step.get("action"), str) and step["action"].startswith("http:"):
                total_http += 1
                if patch_step(step):
                    patched += 1

    if patched > 0:
        with open(path, "w") as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
            f.write("\n")
    print(f"{path}: {patched}/{total_http} HTTP steps patched with auth: admin")
    return 0


if __name__ == "__main__":
    sys.exit(main())
