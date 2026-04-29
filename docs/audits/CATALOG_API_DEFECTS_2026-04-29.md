# Catalog-API defects discovered by anti-bluff verification — 2026-04-29

These are real catalog-api defects surfaced during the
BLUFF-HELIXQA-BANKS-REWRITE-001 verification run against the
deployed v2.3.0-build.25 stack on `thinker.local:8092`.

**They are not bank-conversion artifacts.** Each was previously
hidden by a bank entry whose prose action couldn't actually
execute against the backend (so the entry trivially "passed" by
not running). Now that the converted bank fires real HTTP
requests, the divergence between intended and actual behavior is
visible.

Each entry below is a candidate ticket for the catalog-api
backlog. The bank entries that surfaced them are correct as
authored; the catalog-api implementation needs the fix.

---

## CATAPI-DEFECT-001: `/api/v1/entities` accepts requests without Authorization header

**Severity:** HIGH — authentication bypass on a list endpoint that
is supposed to be admin-only.

**Reproduction:**
```bash
curl -fsS -i http://127.0.0.1:8092/api/v1/entities
# Expected: HTTP/1.1 401 Unauthorized
# Actual:   HTTP/1.1 200 OK
#           {"items":[],"limit":24,"offset":0,"total":0}
```

**Surfaced by bank entries:**
- `FQA-API-007` "Use logged-out token"
- `FQA-API-008` "Call protected endpoint with no Authorization header"
- `FQA-API-009` "Send request with expired token"
- `FQA-API-020` "Use forged token"

All four entries expect 401. All four return 200 with an empty
result page. The list response itself is empty (which the test was
relying on to not leak data) but the **authorization gate isn't
firing at all**.

**Suspected root cause:** the `/api/v1/entities` route is registered
in `main.go` outside the JWT-auth middleware chain, OR the chain
was wired but doesn't have `middleware.RequireAuth()` applied to
this route group.

**Verification once fixed:** re-running
`HELIXQA_HTTP_BASE_URL=http://127.0.0.1:18092 go test
-run TestBankRealBinary_FullQAAPI` should turn FQA-API-007/008/009/020
from FAIL → PASS.

---

## CATAPI-DEFECT-002: `/api/v1/admin/system-info` accessible without admin role

**Severity:** HIGH — privilege-escalation surface (admin info
disclosure).

**Reproduction:**
```bash
# log in as a non-admin user (or no role check at all)
curl -fsS http://127.0.0.1:8092/api/v1/admin/system-info
# Expected: HTTP/1.1 403 Forbidden  (admin-only)
# Actual:   HTTP/1.1 200 OK
#           {"activeConnections":34, "cpuUsage":0, "memoryUsage":..., ...}
```

**Surfaced by:** `FQA-API-010` "Access admin system-info" expects 403, gets 200.

**Suspected root cause:** route is gated by `RequireAuth` (any
logged-in user passes) but not by `RequireRole("admin")`.

---

## CATAPI-DEFECT-003: No rate limit on `/api/v1/auth/login`

**Severity:** HIGH — credential-stuffing / brute-force susceptibility.

**Reproduction:**
```bash
for i in $(seq 1 10); do
  curl -fsS -X POST http://127.0.0.1:8092/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"wrong"}' \
    -w "%{http_code}\n" -o /dev/null
done
# Expected: First few return 401. After ~5 failures → 429.
# Actual:   All 10 return 401. No rate-limiting kicks in.
```

**Surfaced by:** `FQA-API-013` "Send 6 rapid failed login attempts"
expects 429 on the 6th attempt; gets 200 (when password is correct
between iterations) or 401 throughout.

**Suspected root cause:** `internal/middleware/ratelimit.go` (or
similar) is registered globally but doesn't include
`/api/v1/auth/login` in its protected route list, OR the limit
threshold is too high.

---

## CATAPI-DEFECT-004: `/api/v1/storage/roots` accepts unsupported protocols

**Severity:** MEDIUM — input validation gap.

**Reproduction:**
```bash
TOKEN=$(curl -fsS -X POST http://127.0.0.1:8092/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq -r .session_token)
curl -fsS -i -X POST http://127.0.0.1:8092/api/v1/storage/roots \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Bad Root","protocol":"gopher","path":"/x","enabled":true}'
# Expected: HTTP/1.1 400 Bad Request  (unsupported protocol)
# Actual:   HTTP/1.1 201 Created
#           {"id":4,"message":"storage root created","protocol":"gopher",...}
```

**Surfaced by:** `FQA-API-036` "Attempt to create with unsupported
protocol" expects 400, gets 201.

**Suspected root cause:** `repository/storage_root_repository.go`
or its handler doesn't validate `protocol` against the supported
set (`smb|ftp|nfs|webdav|local`). Any string is accepted.

---

## CATAPI-DEFECT-005: `/api/v1/storage/roots` allows duplicate names

**Severity:** MEDIUM — conflict handling missing.

**Reproduction:**
```bash
TOKEN=...  # as above
curl -fsS -X POST http://127.0.0.1:8092/api/v1/storage/roots \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Duplicate Test","protocol":"local","path":"/tmp/a","enabled":true}'
# Returns 201 with id=4

curl -fsS -i -X POST http://127.0.0.1:8092/api/v1/storage/roots \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Duplicate Test","protocol":"local","path":"/tmp/b","enabled":true}'
# Expected: HTTP/1.1 409 Conflict  (name already exists)
# Actual:   HTTP/1.1 201 Created  (silently creates a duplicate)
```

**Surfaced by:** `FQA-API-038` "Create second root with same name"
expects 409, gets 201.

**Suspected root cause:** the `name` column has no UNIQUE
constraint in the migration, OR the repository's Create() doesn't
check for a pre-existing row.

---

## CATAPI-DEFECT-006: Malformed-JSON login still authenticates

**Severity:** UNCLEAR — either a real test design bug in the bank
OR the catalog-api is unexpectedly forgiving of malformed JSON.

**Reproduction:**
```bash
curl -fsS -X POST http://127.0.0.1:8092/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin"'   # unterminated JSON
# Expected: HTTP/1.1 400 Bad Request
# Actual (per test output):
#   HTTP/1.1 200 OK
#   {"user":{"id":1,...},"session_token":"..."}
```

**Surfaced by:** `FQA-API-017` "Send malformed JSON" expects 400, gets 200.

**Note:** the test as authored asks the runner to send malformed
JSON. The bank-prose-to-http.py converter (correctly) generated
valid JSON in the body field. So this might be a converter
limitation (the malformedness intent was lost) rather than a
catalog-api bug. Needs investigation:
- If the converter is wrong: special-case "Send malformed JSON"
  prose so the body is sent as raw bytes (not JSON-encoded).
- If the catalog-api accepts malformed JSON as valid: that's a real
  defect.

---

## How these were discovered

The bank conversion + real-binary execution pipeline
(`HelixQA/pkg/autonomous/bank_realbinary_test.go`) loads the
converted `full-qa-api.json`, fires every step against the live
backend, and asserts on the response. Before this session, those
bank entries were prose strings the executor couldn't run, so
they were trivially "passing" — Article XI §11 calls this the
"bluff" failure mode.

After three rounds of patching (auth-injection → default-bodies →
ongoing), the bank now produces:
- 185 of 330 PASS (56.1%) — the auth flow + read endpoints work as expected.
- 145 FAIL — divided into the 6 catalog-api defects above (~10 cases),
  bank-design chaining issues (~30 cases), and the BLUFF-FQA-API-LOOP-CONSTRUCT-005
  multi-request patterns (~67 cases).

## Cross-references

- Audit: `docs/audits/anti-bluff-2026-04-28.md`
- Real-binary verification: `docs/audits/full-qa-api-realbinary-2026-04-29.md`
- Constitution Article XI §§ 11.1–11.8
- Test runner: `HelixQA/pkg/autonomous/bank_realbinary_test.go`
- Converter: `scripts/audit/bank-prose-to-http.py`
- Patchers: `scripts/audit/bank-patch-auth.py`,
  `scripts/audit/bank-patch-default-bodies.py`
