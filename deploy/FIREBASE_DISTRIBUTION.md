# Firebase Distribution — status (2026-06-29)

**Project:** catalogizer-7a3f1 (CLI auth: milos85vasic@gmail.com). Per operator: generate/obtain
missing config via the Firebase CLI.

## ✅ Distributed (real, verified)
| App | Build | Result |
|-----|-------|--------|
| **Android TV** | dev (debug) | ✅ App Distribution release **2.4.0 (8)** — app `1:881377664729:android:751a0d0e2d873db47768c8`. console.firebase.google.com/project/catalogizer-7a3f1/appdistribution/app/android:com.catalogizer.androidtv/releases/1akea8hl857c8 |
| **Android phone** | dev (debug) | ✅ App Distribution release **2.4.0 (6)** — app `1:881377664729:android:04e3b720c192b30c7768c8` (REGISTERED via CLI this session; google-services.json generated via `firebase apps:sdkconfig`). console.firebase.google.com/project/catalogizer-7a3f1/appdistribution/app/android:com.catalogizer.android/releases/1r73e3ghu9lhg |
| **Android TV** | **prod (release, signed+ProGuard)** | ✅ App Distribution release **2.4.0(8)** — releases/41nac6lmgfp20 (signed with generated keystore). |
| **Android phone** | **prod (release, signed)** | ✅ App Distribution release **2.4.0(6)** — releases/3a9al34l3pj48 (shared keystore). |
| **Web** (catalog-web) | prod (dist) | ✅ Firebase **Hosting** deploy (95 files, release complete) → https://catalogizer-7a3f1.web.app — firebase.json hosting → catalog-web/dist. |

## ⏳ Remaining
| App | Build | Status / next |
|-----|-------|---------------|

| **Desktop** (catalogizer-desktop) | — | NOT a Firebase distribution target (App Distribution = Android/iOS; Hosting = web). Desktop needs another channel. |
| Web functional note | — | Hosting serves the static SPA; it talks to the catalog-api which is LAN-only (thinker.local) — public web users can't reach the LAN API (honest §11.4.6 boundary). |
| Testers | — | Releases created but `no testers/groups` — assign a tester group in console or `--groups`. |
