# run-helixqa-androidtv.sh

**Revision:** 1
**Last modified:** 2026-06-29T13:15:00Z

## Overview
Autonomous 4-phase HelixQA campaign for the Android TV app on all connected
ADB-reachable TV devices: Setup → Doc-Driven Verification → Curiosity Exploration
→ Report & Cleanup. Records screen video, captures screenshots, generates a report.

## Prerequisites
- `catalog-api` running and host-reachable (pass `--api-url http://localhost:<port>`).
- ADB-connected Android TV device(s) (leanback/television feature).
- `ffmpeg` for recording assembly.

## Usage
```bash
./scripts/run-helixqa-androidtv.sh --api-url http://localhost:28080 --timeout 30
```

## §11.4.117 pixel-oracle login (2026-06-29 fix)
The TV build is Jetpack-Compose-for-TV; its accessibility hierarchy is EMPTY
(`uiautomator dump` = 0 nodes). The prior login used hierarchy bounds (never found
the fields) and verified success by grepping the same empty dump for "Sign In" —
which on an empty dump ALWAYS reported "past login" (a §11.4.107(2) PASS-bluff).
`login_on_device` now:
- Enters server URL / username / password via fixed coordinates calibrated on the
  1920x1080 login screen, submits via the IME ENTER (DONE) action.
- VERIFIES login by §11.4.69 sink-side evidence: polls the device okhttp log for a
  real `/api/v1` catalog response. Cannot be faked by a blank UI dump.
The `adb reverse` now forwards device `:8080` to the host's actual API port (parsed
from `--api-url`), so the app's default `localhost:8080` reaches the API on `:28080`.

## Edge cases
- Empty Compose hierarchy → pixel-oracle path (above), no hierarchy dependency.
- API on a non-8080 host port → handled via API_PORT extraction + reverse forward.

## Related
- `scripts/run-helixqa.sh`, `scripts/run-helixqa-android.sh` (phone variant).

## Last verified
2026-06-29 — login VERIFIED (sink-side catalog fetch), Phase 2 reached on MIBOX4.
