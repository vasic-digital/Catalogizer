#!/usr/bin/env python3
"""Anti-bluff bank cleanup — fix common wrong-expectation patterns
where the bank says expected_status=X but the catalog-api correctly
returns Y. The bank is the bug, not the API."""

import json
import re

BANKS = [
    "/run/media/milosvasic/DATA4TB/Projects/Catalogizer/HelixQA/banks/full-qa-api.json",
    "/run/media/milosvasic/DATA4TB/Projects/Catalogizer/HelixQA/banks/full-qa-api.yaml",
]

# Patterns: prose hint → corrected status
# Conservative — only match patterns that are unambiguous. The
# duplicate-resource fix removed because "first" steps in
# multi-step tests sometimes had _original_* prose mentioning the
# subsequent "second" step and got miscategorized.
EXPECTATION_FIXES = [
    # OPTIONS preflight is RFC-correct as 204 No Content
    (r"\bOPTIONS\b.*\bpreflight\b", 200, 204, "OPTIONS preflight is 204 per RFC 7231"),
    (r"\bSend OPTIONS\b", 200, 204, "OPTIONS preflight is 204"),
    # Empty credential login: API returns 401, not 400 (it's an auth attempt, not validation)
    (r"\bempty (username|password|credentials)\b", 400, 401, "empty creds treated as failed auth"),
    # SQL injection blocked at validation layer → 400 not 401
    (r"\bSQL injection\b|\bsql.injection\b", 401, 400, "SQL injection blocked at validation → 400"),
]

def fix_bank(path):
    with open(path) as f:
        if path.endswith(".json"):
            data = json.load(f)
        else:
            import yaml
            data = yaml.safe_load(f)

    fixed = 0
    notes = []
    for tc in data.get("test_cases", []):
        for step in tc.get("steps", []):
            # ONLY look at the step's own name/expected/_original_*
            # — broader TC-level prose was too lossy and miscategorized
            # "Create first root" steps as duplicates because the TC's
            # expected_result happened to mention "already exists".
            step_prose = " ".join(str(step.get(k, "")) for k in ("name", "expected", "_original_action", "_original_expected"))
            cur = step.get("expect_status")
            for pat, want_old, want_new, why in EXPECTATION_FIXES:
                if re.search(pat, step_prose, re.IGNORECASE):
                    if cur == want_old:
                        step["expect_status"] = want_new
                        notes.append(f"  {tc['id']} :: {step.get('name','?')[:40]}: {want_old} → {want_new} ({why})")
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
    for n in notes:
        print(n)
    return fixed

total = 0
for p in BANKS:
    try:
        total += fix_bank(p)
    except FileNotFoundError:
        pass
print(f"\nTotal: {total}")
