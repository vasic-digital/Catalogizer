#!/usr/bin/env python3
"""Anti-bluff fix: scan converted bank for negative-auth tests
where the converter auto-defaulted auth='admin' but the test's
prose says it should send NO token / forged / expired / logged-out.

Such tests with auth='admin' don't actually exercise what they
claim (the harness keeps sending a valid admin bearer regardless),
so they pass for the wrong reason or fail when the API correctly
rejects a request the test-harness was forcing through with a
token. Either outcome is a bluff."""

import json
import re
import sys

BANK_FILES = [
    "/run/media/milosvasic/DATA4TB/Projects/Catalogizer/submodules/helix_qa/banks/full-qa-api.json",
    "/run/media/milosvasic/DATA4TB/Projects/Catalogizer/submodules/helix_qa/banks/full-qa-api.yaml",
]

# Prose patterns that signal a negative-auth test
NEGATIVE_AUTH_PATTERNS = [
    r"\bno[\s_-]?(?:auth|authorization|bearer)[\s_-]?(?:header|token)?\b",
    r"\bwithout[\s_-]?(?:auth|authorization|token|bearer)\b",
    r"\bforged[\s_-]?token\b",
    r"\bexpired[\s_-]?token\b",
    r"\bmalformed[\s_-]?token\b",
    r"\binvalid[\s_-]?token\b",
    r"\blogged[\s_-]?out[\s_-]?token\b",
    r"\brevoked[\s_-]?token\b",
    r"\bunauth(?:enticated|orized)\b",
    r"\bno[\s_-]?credentials\b",
]

# Test-name-like fields and step-level fields to scan for prose
PROSE_FIELDS = ["name", "description", "expected", "_original_action", "_original_expected"]

def is_negative_auth(text: str) -> bool:
    if not text:
        return False
    t = text.lower()
    return any(re.search(p, t) for p in NEGATIVE_AUTH_PATTERNS)

def fix_bank(path: str) -> int:
    with open(path) as f:
        if path.endswith(".json"):
            data = json.load(f)
        else:
            import yaml
            data = yaml.safe_load(f)

    fixed = 0
    for tc in data.get("test_cases", []):
        # Build a context blob from test-level prose
        tc_prose = " ".join(str(tc.get(k, "")) for k in PROSE_FIELDS) + " " + tc.get("name", "")
        for step in tc.get("steps", []):
            step_prose = " ".join(str(step.get(k, "")) for k in PROSE_FIELDS)
            full_prose = tc_prose + " " + step_prose
            if is_negative_auth(full_prose):
                cur_auth = step.get("auth")
                if cur_auth in (None, "admin", "as:admin"):
                    step["auth"] = "none"
                    fixed += 1

    if fixed > 0:
        with open(path, "w") as f:
            if path.endswith(".json"):
                json.dump(data, f, indent=2, ensure_ascii=False)
                f.write("\n")
            else:
                import yaml
                yaml.dump(data, f, sort_keys=False, allow_unicode=True)
        print(f"  {path}: fixed {fixed} step(s)")
    else:
        print(f"  {path}: no negative-auth steps had wrong auth")
    return fixed

total = 0
for p in BANK_FILES:
    try:
        total += fix_bank(p)
    except FileNotFoundError:
        print(f"  {p}: missing — skipping")
print(f"\nTotal fixes: {total}")
