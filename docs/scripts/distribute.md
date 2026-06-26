# `scripts/distribute.sh` — Firebase App Distribution (§11.4.18)

Distributes **catalog-api** (Go binary) and **catalogizer-androidtv** (APK)
to Firebase App Distribution. Reads `FIREBASE_TESTER_EMAILS` from `.env`,
builds both artifacts with production settings, adds testers, and uploads.

---

## Prerequisites

| Dependency | Required for | Install |
|---|---|---|
| `firebase` CLI | Distribution | `npm install -g firebase-tools` |
| `go` | API build | [go.dev/dl](https://go.dev/dl/) |
| Java 17+ | Android TV APK | `brew install openjdk@17` |
| Git | Checksum / submodule tracking | `brew install git` |

**Firebase setup:**
1. Create/select a project at [console.firebase.google.com](https://console.firebase.google.com)
2. Register both apps (API + Android TV) in Project Settings
3. Run `firebase login`
4. Run `firebase use --add <project-id>`
5. Set `FIREBASE_TESTER_EMAILS`, `FIREBASE_API_APP_ID`, and
   `FIREBASE_ANDROIDTV_APP_ID` in `.env`
6. Enable **Firebase App Distribution** in the Console

---

## Usage

```bash
# Full flow — build + add testers + distribute:
scripts/distribute.sh

# Build only (skip App Distribution upload):
SKIP_DISTRIBUTE=true scripts/distribute.sh

# Distribute only (skip building — use pre-existing artifacts):
SKIP_BUILD=true scripts/distribute.sh

# Verbose output:
VERBOSE=true scripts/distribute.sh
```

---

## Inputs

| Variable | Source | Purpose |
|---|---|---|
| `FIREBASE_TESTER_EMAILS` | `.env` | Comma-separated tester emails |
| `FIREBASE_API_APP_ID` | `.env` | Firebase App ID for catalog-api |
| `FIREBASE_ANDROIDTV_APP_ID` | `.env` | Firebase App ID for catalogizer-androidtv |
| `SKIP_BUILD` | env | `true` → use pre-existing artifacts |
| `SKIP_DISTRIBUTE` | env | `true` → build only (no upload) |
| `VERBOSE` | env | `true` → print build command output |

---

## Outputs

| Path | Description |
|---|---|
| `build/catalog-api/catalog-api` | Compiled Go binary |
| `catalogizer-androidtv/**/release/*.apk` | Android TV release APK |

Tester emails are registered with Firebase App Distribution. Builds are
uploaded to the configured Firebase project.

---

## Side-effects

- Firebase CLI writes/reads `~/.config/configstore/firebase-tools.json`
- Gradle caches are created under `catalogizer-androidtv/.gradle/`
- Go module cache under `$(go env GOMODCACHE)`

---

## Dependencies

- `scripts/lib/project-config.sh` — shared build paths
- `scripts/firebase_setup_env.sh` — bootstrap Firebase env vars
- `.env` at project root

---

## Related scripts

| Script | Purpose |
|---|---|
| `scripts/firebase_setup_env.sh` | Create/migrate Firebase env vars |
| `scripts/firebase_verify.sh` | Verify Firebase connectivity |
| `scripts/lib/build-catalog-api.sh` | Underlying Go build helper |

---

## Last verified

2026-06-26
