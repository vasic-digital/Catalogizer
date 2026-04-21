# Bug Retrospective Template

> When a bug is found in manual testing that the automated tests missed,
> file a retrospective using this template. The retrospective feeds
> LANDMINES, fixes-validation, and HelixQA bank entries.
>
> Source: Master Plan v2 §8.4.

---

```markdown
## Bug Report: [YYYY-MM-DD] — [BRIEF DESCRIPTION]

### What the user did
1. [Step 1]
2. [Step 2]
3. [Step N]
4. [Result: What went wrong — exactly. Screenshot / video / log line.]

### What the automated tests did right
- [List every test that passed around this bug that SHOULD have caught
  it but did not. Include unit / integration / E2E / HelixQA / challenge
  names. This is the coverage-gap inventory.]

### Why the tests failed to catch this (Root Cause)

- **Missing Constraint:** [What invariant wasn't encoded. E.g.
  "tests did not verify foreground package per structured step"]
- **Detection Gap:** [What test TYPE was missing. E.g. "integration
  test mocked the vision provider instead of using a real screenshot"]
- **Coverage Gap:** [What specific scenario wasn't tested. E.g.
  "tv-channel-* tests never ran with a sibling app installed on the
  device"]

### New Landmine Rule

**RULE-[SCOPE]-[NNN]: [Rule Name]**
- **Context:** [why this matters — include the originating incident]
- **Detection:** [grep / test command that would have caught this]
- **Fix:** [how to fix when the rule is violated]

Append this to `docs/LANDMINES.md` in the SAME commit as the fix.

### Tests Added (all four are mandatory per Article VII)

- [ ] **Unit test**: [file path + test function name]
- [ ] **Integration test**: [file path + test function name]
- [ ] **E2E test**: [file path + test function name — if the bug was
      user-visible]
- [ ] **Regression entry**: new YAML block in
      `HelixQA/banks/fixes-validation.yaml` with ID
      `fix-qa-YYYY-MM-DD-NNN-slug`

### Evidence

- **Screenshot**: [path inside qa-results/session-<ts>/screenshots/]
- **Video**: [path + timestamp MM:SS inside qa-results/session-<ts>/
  videos/, per HelixQA CLAUDE.md "Evidence-Backed Issue Tickets"]
- **Log excerpt**: [tail of the relevant log file + line numbers]
- **Server state**: [relevant DB rows, cache entries, or API response
  that proves the backend was healthy, if applicable]

### Commit chain (once fix lands)

1. fix(<scope>): <subject>  — the actual fix
2. test(<scope>): <subject> — the four tests above
3. docs(landmines): add RULE-[SCOPE]-[NNN] for <root cause>
4. Bump the submodule pointer in the main repo + push to all 6 remotes

### Constitution check

- [ ] Article V — 100% coverage across 10 categories not reduced
- [ ] Article VI — `docs/OPEN_POINTS_CLOSURE.md` updated if this item
      was open there
- [ ] Article VII — Full-QA Master Cycle re-run post-fix
- [ ] Article VIII — device state unchanged by the fix (no settings
      put X)
- [ ] Article IX — no manual bash workaround; fix landed in Go code
```

## How to use

1. Copy the template into `docs/reports/qa-sessions/<date>/tickets/FIX-QA-YYYY-MM-DD-NNN-<slug>.md`.
2. Fill every section — an empty evidence block voids the ticket
   (HelixQA Constitution "Evidence-Backed Issue Tickets").
3. Land the fix + tests + LANDMINES update in a single commit (or
   commit series with an obvious join point).
4. Rerun HelixQA; verify the bank regression entry passes.
5. Update `docs/OPEN_POINTS_CLOSURE.md` if this closed an operator
   item.

## Do not

- File a ticket without a reproduction. "Sometimes it fails" is not a
  retrospective — it's a request for more data.
- Leave the "Tests Added" section with unchecked boxes. Land the tests
  in the same PR as the fix. A fix without tests is a future
  regression in waiting.
- Deduplicate retrospectives by deleting older ones. Each is a snapshot
  of what the code knew; the trail matters.
