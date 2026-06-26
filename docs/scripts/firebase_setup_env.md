# `scripts/firebase_setup_env.sh` — Firebase Env Bootstrap (§11.4.18)

Creates Firebase-related `.env` entries if they do not exist and
documents them in `.env.example`.

---

## Prerequisites

- Write access to the project root (to modify `.env` and `.env.example`)

---

## Usage

```bash
scripts/firebase_setup_env.sh
```

---

## Inputs

| Item | Source | Purpose |
|---|---|---|
| `.env` | project root | Existing env vars (read) |
| `.env.example` | project root | Documented env template (appended) |

---

## Outputs

| File | Change |
|---|---|
| `.env.example` | Firebase block appended if absent |
| stdout | Status report of missing keys + suggested commands |

The script **never writes secrets** — it only reports what is missing
and prints the placeholder block the operator should paste.

---

## Side-effects

None — `.env` is never modified; `.env.example` gets the documentation
block appended once.

---

## Dependencies

- `.env` at project root
- `.env.example` at project root

---

## Related scripts

| Script | Purpose |
|---|---|
| `scripts/distribute.sh` | Consumes the Firebase env vars |
| `scripts/firebase_verify.sh` | Verifies the env config |

---

## Last verified

2026-06-26
