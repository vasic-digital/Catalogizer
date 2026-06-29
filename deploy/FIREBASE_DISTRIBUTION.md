# Firebase App Distribution — status (2026-06-29)

**Operator mandate:** Firebase-distribute all client apps (Android phone, Android TV, desktop,
web), dev + prod. **Firebase project:** catalogizer-7a3f1 (CLI auth: milos85vasic@gmail.com).

## ✅ Done (real, verified)
| App | Build | Result |
|-----|-------|--------|
| **Android TV** (com.catalogizer.androidtv) | **dev (debug)** | ✅ **DISTRIBUTED** — release **2.4.0 (8)**, app id `1:881377664729:android:751a0d0e2d873db47768c8`. Console: https://console.firebase.google.com/project/catalogizer-7a3f1/appdistribution/app/android:com.catalogizer.androidtv/releases/1akea8hl857c8 |

## ⏳ Remaining + honest blockers (§11.4.6 — not bluffed)
| App | Blocker / next step |
|-----|---------------------|
| Android TV **prod (release)** | Needs the release signing config/keystore (release APK/AAB must be signed). |
| Android **phone** (catalogizer-android) | Debug APK exists BUT **no `google-services.json`** in catalogizer-android/app — the Firebase app for that package is not registered/configured. Register the phone app in project catalogizer-7a3f1 + add google-services.json, then distribute. |
| **Desktop** (catalogizer-desktop) | **NOT a Firebase App Distribution target** — App Distribution supports Android + iOS only (§11.4.112 platform boundary). Desktop distribution needs another channel. |
| **Web** (catalog-web) | **NOT an App Distribution target** — web deploys via **Firebase Hosting** (`firebase deploy --only hosting`), not App Distribution. |
| **Testers/groups** | The AndroidTV release was created but `no testers or groups specified, skipping` — assign a tester group in the console (or `--groups <id>`) to notify testers. |

## Notes
- AndroidTV dev APK distributed = the validated build (book/comic/image readers, player crash
  fixed, infra on thinker.local), main @16bfacf0.
- Per §11.4.173, builds should be containerized/distributed (task #32) — current APKs were
  host-built; the prod/release builds should move to the containerized path.
