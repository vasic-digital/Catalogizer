# Anti-Bluff Audit — 2026-04-28

## Summary

A static heuristic scan of the umbrella repo + every submodule, looking
for the bluff patterns banned by Constitution Article XI. Heuristic
scanner: `scripts/audit/anti-bluff-scan.sh`. Per-finding output:
`/tmp/catalogizer-build-logs/bluff-umbrella.tsv` and
`/tmp/catalogizer-build-logs/bluff-by-submodule/*.tsv`.

This is a **static heuristic** report. It MUST be paired with the
dynamic "comment out the feature, re-run, see if it still passes"
ritual from Article XI §11.2.5. Some findings will be false positives —
each must be triaged by a human.

## Headline numbers (umbrella, after vendored-OSS excludes)

| Kind | Count | Tier | Action |
|---|---:|:---:|---|
| `PROSE_HELIXQA_ACTION` | 4564 | 1 | Rewrite bank entry to executable action |
| `GO_NO_ASSERT` | 982 | 3 | Triage — many will be helper funcs |
| `SKIP_WITHOUT_TICKET` | 188 | 2 | Tag each `t.Skip` with `SKIP-OK: #<ticket>` or remove |
| `GO_NIL_ONLY` | 164 | 2 | Add real assertion or rewrite |
| `GO_MOCK_IN_INTEGRATION` | 150 | 1 | Replace mock with real container/database |
| `ASSERT_TAUTOLOGY` | 22 | 1 | Delete; rewrite |
| `CHALLENGE_BLIND_SHELL` | 12 | 3 | Verify case-by-case (most are graceful absence-handling) |
| `GO_HTTPTEST_ABUSE` | 9 | 1 | Replace `httptest.NewServer` with real binary in E2E paths |
| **Total** | **6091** | | |

## Tier definitions

- **Tier 1 — high-confidence bluff, fix urgently.** `PROSE_HELIXQA_ACTION`,
  `GO_MOCK_IN_INTEGRATION`, `ASSERT_TAUTOLOGY`, `GO_HTTPTEST_ABUSE`.
- **Tier 2 — likely real, requires per-line review.** `SKIP_WITHOUT_TICKET`,
  `GO_NIL_ONLY`.
- **Tier 3 — high false-positive rate, sampled triage.** `GO_NO_ASSERT`
  (helper funcs and table-driven setup will hit this), `CHALLENGE_BLIND_SHELL`
  (legitimate `|| true` for graceful absence-handling shows up).

## Top 10 submodules by finding count

| Submodule | Findings |
|---|---:|
| HelixQA | 4443 |
| Challenges | 953 |
| Containers | 45 |
| Observability | 36 |
| LLMOrchestrator | 24 |
| VisionEngine | 21 |
| EventBus | 14 |
| Database | 14 |
| Concurrency | 13 |
| Streaming | 12 |

## Tier 1 sample findings (verified bluffs — fix these first)

### PROSE_HELIXQA_ACTION

`Challenges/banks/examples/ui-automation-android.json` lines 26–33:

```
action="keyboard"   ← prose; should be e.g. adb_shell: input text admin
action="verify"     ← prose; should be e.g. assertVisible: 'Login'
action="launch"     ← prose; should be e.g. adb_shell: am start -n …
action="wait"       ← prose; should be e.g. wait_for: 'Login successful'
action="screenshot" ← prose; should be e.g. screenshot: post_login.png
```

This pattern repeats across every bank file. The scanner found 4564
such lines. Banks need a structural rewrite, not line-by-line patches.

### GO_HTTPTEST_ABUSE

Files named `_e2e_test.go` should hit the real binary on the bound port,
not a synthetic `httptest.NewServer(router)`. Hits:

```
catalog-api/tests/integration/api_e2e_test.go:317        ts := httptest.NewServer(router)
catalog-api/tests/integration/conversion_e2e_test.go:348 ts := httptest.NewServer(router)
catalog-api/tests/integration/subtitle_e2e_test.go:311   ts := httptest.NewServer(router)
catalog-api/tests/integration/user_management_e2e_test.go:353 ts := httptest.NewServer(router)
Auth/tests/e2e/auth_e2e_test.go:57                       srv := httptest.NewServer(mux)
```

Each is a real Article XI §11.5 violation: an "E2E" test that doesn't
exercise the real binary. Replacement pattern: start the real `main` via
`exec.Command`, wait for `.service-port`, hit it with `http.Client`, assert
on the actual response body.

### ASSERT_TAUTOLOGY (umbrella, post-vendor-exclude)

`pkg/pool/pool_test.go:1464, 1468`:

```go
assert.True(t, true)
```

Two occurrences. Both must be deleted or replaced.

### GO_MOCK_IN_INTEGRATION

150 hits — most concentrated in catalog-api `tests/integration/` and
HelixQA. Each represents a non-unit test file using mock/stub/fake
infrastructure. Per Article XI §11.2.2 and Universal-11, these must
operate against real containers/databases/services.

## Resolution plan (tracked as Task #14)

The full resolution is multi-day work. Suggested ordering:

1. **Banks rewrite (HelixQA + Challenges)** — biggest by count (4564);
   structural rewrite. Convert every bank file from prose actions to
   executable `adb_shell:` / `playwright:` / `http:` / `assertVisible:`
   actions. One bank file at a time. Per Article XI §11.4 + §7.6.
2. **catalog-api E2E rewrite** — 4 files; replace `httptest.NewServer`
   with real-binary `exec.Command` startup. Add the matching negative
   test (request to wrong port → connection refused → test fails).
3. **GO_MOCK_IN_INTEGRATION triage** — 150 hits; for each, decide:
   move to `*_test.go` under `-short`, or replace mock with a real
   container fixture.
4. **SKIP_WITHOUT_TICKET sweep** — 188 hits; for each `t.Skip` add
   `SKIP-OK: #<ticket>` referencing the closure plan or remove the
   skip.
5. **`pool_test.go` `assert.True(t, true)` deletion** — 2 lines;
   trivial.
6. **GO_NIL_ONLY review** — 164 hits; for each, add either a positive
   assertion (the function returned the expected value) or rewrite.
7. **GO_NO_ASSERT triage** — 982 hits; sample first; many are
   table-driven setup helpers that aren't tests at all (false
   positives).

Each fix lands per Article VII §7.6 with the four-artefact tail:
unit/integration test, fixes-validation entry, HelixQA bank entry,
challenge.

## Audit ritual (Article XI §11.7)

Every Full-QA Master Cycle must pick five tests + five Challenges at
random, comment out the target feature, re-run, confirm FAIL. Any that
still pass are tagged `BLUFF` and rewritten before the cycle terminates.

This audit is the static-scan complement; the random-mutation audit
catches what static heuristics miss.

## Next refresh

This file is a snapshot. Re-run on every clean-pass exit of the
Full-QA Master Cycle (Article VII §7.3 "NOTHING LEFT") and at every
release gate. Update the date in the filename and link the previous
audit so trend is visible.

---

*Generated: 2026-04-28*
*Scanner: scripts/audit/anti-bluff-scan.sh*
*Constitution authority: Article XI*
