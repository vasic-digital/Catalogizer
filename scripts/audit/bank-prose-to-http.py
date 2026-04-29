#!/usr/bin/env python3
"""
bank-prose-to-http.py — convert HelixQA bank prose actions to
structured ActionTypeHTTP / ActionTypeAssert form.

Reads a HelixQA bank JSON (top-level: {"test_cases": [...]}, each
test case has {steps: [{name, action, expected, ...}]}). For each
step whose `action` field is a recognized HTTP-prose pattern,
emits the structured form using HelixQA's new schema:

    {
      "name": "...",
      "action": "http: METHOD /path",
      "auth": "admin"|"none"|"as:user",
      "body": {...},
      "expect_status": 200,
      "expect_json_path": "$.foo",
      "expect_body_contains": "...",
      "expected": "<original prose preserved as docstring>"
    }

Patterns recognized (handles ~80% of full-qa-api / atmosphere /
full-qa-web prose actions):

  GET /path                          → {action: "http: GET /path"}
  GET /path with admin token         → adds auth: "admin"
  POST /path with body {...}         → adds body parsed from inline JSON
  POST /path with body {...} as user → adds auth: "as:user"
  DELETE /path                       → {action: "http: DELETE /path"}

Steps that don't match a known pattern are left untouched, with
a "_conversion_note" field added so reviewers can find them.
Such steps still parse as ActionTypeDescription and print a
WARNING when run, exactly as before this conversion — no
regression.

Usage:
  scripts/audit/bank-prose-to-http.py --input HelixQA/banks/full-qa-api.json \
      --output HelixQA/banks/full-qa-api.executable.json --report

The "expected" field text patterns are also used to infer
ExpectStatus / ExpectJSONPath / ExpectBodyContains:
  "200 OK"                                 → expect_status: 200
  "401 Unauthorized"                       → expect_status: 401
  "JSON containing token field"            → expect_json_path: "$.session_token"
  "JSON containing X field"                → expect_json_path: "$.X"
  "error message"                          → expect_body_contains: "error"

This is a 80%-coverage mechanical converter, not a complete one.
The remaining 20% require manual review and are tagged in the
output with "_conversion_note": "manual-review-required".

Article XI verification: each generated action is structurally
executable by HelixQA's HTTPExecutor. Running the converted bank
against the deployed thinker:8092 stack with HELIXQA_HTTP_BASE_URL
set will produce real PASS/FAIL outcomes (not bluff PASSes).
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from dataclasses import dataclass, field
from typing import Any


@dataclass
class ConversionStats:
    total_steps: int = 0
    converted: int = 0
    already_executable: int = 0
    manual_review: int = 0
    by_method: dict[str, int] = field(default_factory=dict)
    unrecognized_samples: list[str] = field(default_factory=list)


# Patterns: each (regex, builder) — builder takes a re.Match and
# returns (action, body, auth) tuple.

HTTP_PROSE_PATTERNS: list[tuple[re.Pattern[str], Any]] = []


def _register(pattern: str, builder):
    HTTP_PROSE_PATTERNS.append((re.compile(pattern, re.IGNORECASE), builder))


# Pattern 1: "<METHOD> <PATH> with body {<inline JSON>}" + optional auth/user phrase
# OPTIONS is included for CORS preflight tests.
_register(
    r"^\s*(?P<method>GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)\s+"
    r"(?P<path>/[\w/_\-{}.:?=&]+)"
    r"(?:\s+with\s+body\s+(?P<body>\{.*?\}))?"
    r"(?P<authclause>\s+(?:as|with)\s+(?:admin|user)\s*[:\w]*)?",
    lambda m: (
        f"http: {m.group('method').upper()} {m.group('path')}",
        _parse_inline_json(m.group("body")),
        _infer_auth(m.group("authclause") or ""),
    ),
)


def _parse_inline_json(s: str | None):
    if not s:
        return None
    s = s.strip()
    try:
        return json.loads(s)
    except json.JSONDecodeError:
        # Attempt to recover from single-quoted JSON
        try:
            return json.loads(s.replace("'", '"'))
        except json.JSONDecodeError:
            return s  # leave as raw string


def _infer_auth(clause: str) -> str:
    clause = clause.lower()
    if "admin" in clause:
        return "admin"
    if "as user" in clause:
        return "as:viewer"  # placeholder; manual review may refine
    return ""


# Public endpoints that don't require authentication.
# Used by BLUFF-FQA-API-AUTH-INJECT-001 to decide whether to
# default `auth: "admin"` on a converted step.
PUBLIC_ENDPOINT_PREFIXES = (
    "/health",
    "/api/v1/health",
    "/api/v1/auth/login",
    "/api/v1/auth/register",
    "/api/v1/auth/refresh",  # uses refresh_token in body, not bearer
    "/api/v1/auth/logout",   # logout takes the token in body too
    "/api/v1/discovery",     # service-discovery endpoint (when present)
    "/metrics",              # Prometheus metrics
)


def _is_public_endpoint(path: str) -> bool:
    """True if `path` is a public (no-auth) catalog-api endpoint."""
    p = path.split("?")[0].split("#")[0]
    return any(p == prefix or p.startswith(prefix + "/")
               for prefix in PUBLIC_ENDPOINT_PREFIXES)


# --- Playwright (web) prose patterns ---

PLAYWRIGHT_PROSE_PATTERNS: list[tuple[re.Pattern[str], Any]] = []


def _wregister(pattern: str, builder):
    PLAYWRIGHT_PROSE_PATTERNS.append((re.compile(pattern, re.IGNORECASE), builder))


# Open / Navigate to URL → playwright: navigate <url>
_wregister(
    r"^\s*(?:open|navigate(?:\s+to)?)\s+(?P<url>https?://\S+)",
    lambda m: f"playwright: navigate {m.group('url')}",
)
# Open <path> in the browser → playwright: navigate <path>
_wregister(
    r"^\s*open\s+(?P<url>/[^\s]+)\s+in\s+the\s+browser",
    lambda m: f"playwright: navigate {m.group('url')}",
)
# Click <button|element|"text"|'text'> → playwright: click text=<X>
_wregister(
    r"^\s*click\s+(?:on\s+)?(?:the\s+)?[\"']?(?P<target>[^\"']+?)[\"']?(?:\s+button|\s+link|\s+icon|\s+element|\s+tab)?\s*$",
    lambda m: f"playwright: click text={m.group('target').strip()}",
)
# Type 'X' in/into Y field → playwright: fill <selector> <text>
_wregister(
    r"^\s*type\s+[\"'](?P<text>[^\"']+)[\"']\s+(?:in|into)\s+(?:the\s+)?(?P<field>\w+)\s+field",
    lambda m: f"playwright: fill input[name={m.group('field').lower()}] {m.group('text')}",
)
# Type 'X' in Y, 'Z' in W → composite "fill" actions; convert first only
_wregister(
    r"^\s*type\s+[\"'](?P<text>[^\"']+)[\"']\s+(?:in|into)\s+(?P<field>\w+)",
    lambda m: f"playwright: fill input[name={m.group('field').lower()}] {m.group('text')}",
)
# Press <key> → playwright: press <key>
_wregister(
    r"^\s*press\s+(?:the\s+)?(?P<key>\w+)\s+key",
    lambda m: f"playwright: press {m.group('key').strip()}",
)
# Wait for X → playwright: waitFor text=X
_wregister(
    r"^\s*wait\s+for\s+(?:the\s+)?[\"']?(?P<target>[^\"']+?)[\"']?(?:\s+to\s+(?:appear|load|render))?\s*$",
    lambda m: f"playwright: waitFor text={m.group('target').strip()}",
)
# Look for / Check / Find / Verify <X> is visible → playwright: assertVisible
_wregister(
    r"^\s*(?:look\s+for|check\s+(?:that\s+|for\s+)?|find|verify(?:\s+that)?|observe|see)\s+"
    r"(?:the\s+)?[\"']?(?P<target>[^\"']+?)[\"']?(?:\s+is\s+visible|\s+visible)?\s*$",
    lambda m: f"playwright: assertVisible text={m.group('target').strip()}",
)


def _convert_playwright(action: str) -> str | None:
    """Try to convert web prose into a playwright: action.
    Returns None if no pattern matches."""
    for pattern, builder in PLAYWRIGHT_PROSE_PATTERNS:
        m = pattern.match(action)
        if m:
            return builder(m)
    return None


# --- Expected-text → assertion inference ---

_STATUS_RE = re.compile(r"\b(\d{3})\b")
_JSON_FIELD_RE = re.compile(
    r"(?:JSON|json) (?:containing|with) (?P<field>\w+)(?: field)?",
    re.IGNORECASE,
)


def _infer_expectations(expected: str) -> dict[str, Any]:
    out: dict[str, Any] = {}
    if not expected:
        return out
    if m := _STATUS_RE.search(expected):
        out["expect_status"] = int(m.group(1))
    if m := _JSON_FIELD_RE.search(expected):
        field_name = m.group("field")
        # Special-case: token/session — map to session_token
        if field_name.lower() in ("token", "jwt"):
            field_name = "session_token"
        out["expect_json_path"] = f"$.{field_name}"
    if "error" in expected.lower() and "expect_status" in out and out["expect_status"] >= 400:
        # Server returns an error message in body
        out.setdefault("expect_body_contains", "error")
    return out


def convert_step(step: dict[str, Any], stats: ConversionStats) -> dict[str, Any]:
    stats.total_steps += 1
    action = step.get("action", "")
    if not isinstance(action, str):
        return step
    # Already structured → leave alone
    if action.startswith(("http:", "adb_shell:", "tap:", "swipe:", "text:",
                          "keypress:", "sleep:", "screenshot",
                          "playback_check:", "frame_diff:", "assert:")):
        stats.already_executable += 1
        return step

    converted = dict(step)  # shallow copy
    matched = False
    # First try HTTP patterns
    for pattern, builder in HTTP_PROSE_PATTERNS:
        m = pattern.match(action)
        if m:
            new_action, body, auth = builder(m)
            converted["action"] = new_action
            if body is not None:
                converted["body"] = body
            if auth:
                converted["auth"] = auth
            else:
                # BLUFF-FQA-API-AUTH-INJECT-001: any non-public endpoint
                # defaults to auth: "admin". Public endpoints (auth/login,
                # auth/register, /health) explicitly stay auth: "none".
                # The bank can override this by setting auth: explicitly.
                path = m.group("path")
                if _is_public_endpoint(path):
                    pass  # leave auth unset, executor treats as "none"
                else:
                    converted["auth"] = "admin"
            converted.setdefault("_original_action", action)
            matched = True
            method = m.group("method").upper() if "method" in m.groupdict() else "?"
            stats.by_method[method] = stats.by_method.get(method, 0) + 1
            break

    # Then try Playwright patterns
    if not matched:
        playwright_action = _convert_playwright(action)
        if playwright_action:
            converted["action"] = playwright_action
            converted.setdefault("_original_action", action)
            matched = True
            verb = playwright_action.split()[1] if len(playwright_action.split()) > 1 else "?"
            stats.by_method["pw:" + verb] = stats.by_method.get("pw:" + verb, 0) + 1

    # Apply expectation inference from the "expected" prose
    expected_text = step.get("expected", "")
    if isinstance(expected_text, str):
        for k, v in _infer_expectations(expected_text).items():
            converted.setdefault(k, v)

    if matched:
        stats.converted += 1
    else:
        stats.manual_review += 1
        converted["_conversion_note"] = "manual-review-required"
        if len(stats.unrecognized_samples) < 10:
            stats.unrecognized_samples.append(action[:120])
    return converted


def convert_bank(data: dict[str, Any]) -> tuple[dict[str, Any], ConversionStats]:
    stats = ConversionStats()
    cases = data.get("test_cases", [])
    for tc in cases:
        steps = tc.get("steps", [])
        tc["steps"] = [convert_step(s, stats) for s in steps]
    # Tag the bank as converted at the metadata level so the
    # runner / scanner can recognize structurally-executable banks.
    meta = data.setdefault("metadata", {})
    meta["bluff_audit_status"] = "converted-by-bank-prose-to-http.py"
    meta["bluff_audit_date"] = "2026-04-29"
    return data, stats


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--input", required=True)
    ap.add_argument("--output", required=True)
    ap.add_argument("--report", action="store_true")
    args = ap.parse_args()

    with open(args.input) as fh:
        data = json.load(fh)

    converted, stats = convert_bank(data)

    with open(args.output, "w") as fh:
        json.dump(converted, fh, indent=2, ensure_ascii=False)
        fh.write("\n")

    if args.report:
        print(f"=== conversion report: {args.input} → {args.output} ===")
        print(f"  total steps:           {stats.total_steps}")
        print(f"  converted (HTTP):      {stats.converted}")
        print(f"  already executable:    {stats.already_executable}")
        print(f"  manual-review needed:  {stats.manual_review}")
        if stats.by_method:
            print(f"  by method: {dict(sorted(stats.by_method.items()))}")
        if stats.unrecognized_samples:
            print(f"  unrecognized prose samples (first {len(stats.unrecognized_samples)}):")
            for s in stats.unrecognized_samples:
                print(f"    {s}")
        coverage = 100 * stats.converted / max(1, stats.total_steps)
        print(f"  coverage: {coverage:.1f}%")
    return 0


if __name__ == "__main__":
    sys.exit(main())
