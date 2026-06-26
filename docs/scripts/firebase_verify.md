# `scripts/firebase_verify.sh` — Firebase Connectivity & Service Report (§11.4.18)

Tests Firebase CLI connectivity and reports on enabled services,
project access, App Distribution readiness, and build artifact status.

---

## Prerequisites

- `firebase` CLI installed (see `scripts/distribute.sh` prerequisites)
- `firebase login` completed
- At least one Firebase project accessible

---

## Usage

```bash
# Basic check:
scripts/firebase_verify.sh

# The script exits 0 on all-pass, 1 on any failure.
```

---

## Checks performed

| # | Check | What it verifies |
|---|---|---|
| 1 | Firebase CLI | `firebase --version` returns a version |
| 2 | Authentication | `firebase projects:list` returns data |
| 3 | Project access | Lists accessible Firebase projects |
| 4 | App Distribution config | FIREBASE_* env vars are set in `.env` |
| 5 | App Distribution API | CLI help for `appdistribution:distribute` is available |
| 6 | Build artifacts | Pre-existing `catalog-api` binary + APK found |

---

## Inputs

| Variable | Source | Purpose |
|---|---|---|
| `FIREBASE_TESTER_EMAILS` | `.env` | Checked for presence |
| `FIREBASE_API_APP_ID` | `.env` | Checked for presence |
| `FIREBASE_ANDROIDTV_APP_ID` | `.env` | Checked for presence |

---

## Outputs

| Destination | Content |
|---|---|
| stdout | Per-check PASS/FAIL/SKIP + summary table |
| Exit code | `0` = all passed, `1` = any failed |

---

## Side-effects

- Firebase CLI reads `~/.config/configstore/firebase-tools.json` (auth state)

---

## Dependencies

- `.env` at project root (optional — missing env vars are reported as FAIL)
- Firebase CLI authenticated session

---

## Related scripts

| Script | Purpose |
|---|---|
| `scripts/distribute.sh` | Full distribution workflow |
| `scripts/firebase_setup_env.sh` | Bootstrap missing env vars |

---

## Last verified

2026-06-26
