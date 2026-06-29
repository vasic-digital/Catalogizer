# Android TV Crashlytics Wiring — QA Evidence

**Revision:** 1
**Last modified:** 2026-06-29T12:45:00Z
**Run ID:** crashlytics-wiring-20260629
**Device:** MIBOX4 @ 192.168.0.214:5555 (Mi Box 4, Android TV)
**APK:** com.catalogizer.androidtv v2.4.0 (versionCode 8), installed 2026-06-29 09:37
**Firebase project:** catalogizer-7a3f1 (project_number 881377664729)
**App ID:** 1:881377664729:android:751a0d0e2d873db47768c8
**Authority:** Constitution §11.4.152 (Crashlytics monitoring) + §11.4.108 (runtime-signature) + §11.4.6 (no-guessing) + §11.4.69 (sink-side evidence)

---

## Operator action items (read first)

| # | Item | Status |
|---|------|--------|
| 1 | Live console crash round-trip (enable → crash → relaunch → console event) | PENDING — autonomous-infeasible on blank Compose-TV hierarchy + opt-in privacy gate; see §4 |
| 2 | Settings "Crash Reporting" toggle does not reflect persisted SDK state (`remember{mutableStateOf(false)}`) | TRACKED follow-up — UI/state desync, needs RED test + rebuild |

---

## 1. Console data review (§11.4.152, all three surfaces)

Queried Firebase Crashlytics for app `1:881377664729:android:751a0d0e2d873db47768c8` over the
full available 90-day window (2026-04-01 → 2026-06-29) for **FATAL, NON_FATAL, and ANR**:

| Query | Result |
|-------|--------|
| `topIssues` (FATAL+NON_FATAL+ANR, 90d) | **0 results** |
| `topVersions` (90d) | **0 results** |
| `list_events` | **0 results** |

**Verdict:** The console is genuinely empty. This is a verified FACT across three independent
queries — NOT a single-query guess (§11.4.6).

## 2. Why the console is empty (root cause, FACT)

Two cooperating causes, both proven:

1. **Old crashes could never report.** The pre-2026-06-29 APK shipped a *dummy* `google-services.json`
   (placeholder `PASTE_YOUR_API_KEY`). With no valid Firebase project, the Crashlytics SDK had no
   backend to upload to — the ~41 startup FATALs were only visible in local logcat
   (FirebaseInitProvider failure), never uploaded. Root cause fixed: real `catalogizer-7a3f1`
   config now embedded (verified §3).

2. **Collection is opt-in by design.** `CatalogizerTVApplication.kt:55` calls
   `setCrashlyticsCollectionEnabled(false)` deliberately (class header lines 26-28: "opt-in by
   default — the user enables it via Settings > Crash Reporting"). On-device shared_prefs confirms
   `firebase_crashlytics_collection_enabled value="false"`. This is correct privacy-respecting
   default behavior, NOT a bug (§11.4.7 demotion: initial "critical bug" classification corrected
   to "intentional design" after reading full context). Forcing it true would be a §11.4.122 silent
   behavior change.

## 3. Runtime signature — Firebase/Crashlytics genuinely wired (§11.4.108) — **PASS**

Captured on-device via `adb logcat` after `am force-stop` + clean launch
(evidence: `firebase_full_init.txt`):

```
I/FirebaseApp:          Device unlocked: initializing all Firebase APIs for app [DEFAULT]
D/FirebaseSessions:     Initializing Firebase Sessions SDK.
I/FirebaseCrashlytics:  Initializing Firebase Crashlytics 18.6.0 for com.catalogizer.androidtv
I/FirebaseInitProvider: FirebaseApp initialization successful
```

The final line — `FirebaseApp initialization successful` — is **the exact line whose ABSENCE caused
the old 41-crash FATAL state**. Its presence is unambiguous regression-fix evidence.

**Sink-side proof (§11.4.69):** the SDK fetched its settings document from the Firebase backend —
`files/.com.google.firebase.crashlytics.files.v2:com.catalogizer.androidtv/com.crashlytics.settings.json`
(774 bytes) exists on-device. This file can only be created by a successful authenticated round-trip
to the real `catalogizer-7a3f1` project. Installation IDs were issued
(`crashlytics.installation.id`, `firebase.installation.id` in `crashlytics_prefs.xml`).

## 4. No-crash regression proof — **PASS**

After clean launch, app process is **ALIVE** (pidof returns a live pid), **zero FATAL EXCEPTION /
AndroidRuntime** lines in 12s+ of captured logcat (evidence: `launch_logcat.txt`). Contrast: the
pre-fix APK produced an immediate FirebaseInitProvider FATAL on every launch.

## 5. Honest gap (§11.4.52 / §11.4.3 — NOT a faked PASS)

The full live console round-trip (enable collection → trigger test crash → relaunch → poll console
for the event) is **not autonomously completable** on this device because:
- The Compose-TV accessibility hierarchy is blank (`uiautomator dump` empty — §11.4.117 case), so
  driving Settings → Crash Reporting toggle → Test Crash button via D-pad is non-introspectable.
- Collection is gated behind the opt-in toggle (correct privacy design), and the SDK overwrites
  direct shared_prefs edits.
- The app ships the correct instrumentation (`triggerTestCrash()` + `toggleCrashlytics()` in
  Settings, wired at `SettingsScreen.kt:409-425`), so the path EXISTS and is operator-drivable.

Classified `operator_attended` per §11.4.52 with this tracked migration item — **never reported as a
faked PASS**. A future autonomous path: add an androidTest instrumentation case that calls
`setCrashlyticsCollectionEnabled(true)` + `recordException()` then asserts upload.

## Evidence files

| File | Proves |
|------|--------|
| `firebase_full_init.txt` | Full Firebase + Crashlytics init chain, init successful |
| `firebase_tags.txt` | Crashlytics 18.6.0 initializing |
| `launch_logcat.txt` | No FATAL on launch (35 KB capture) |
| `crashlytics_prefs.xml` | Installation IDs issued, collection opt-in state |
| `app_running_screenshot.png` | App UI rendered, alive |

## Summary verdict

| Aspect | Verdict |
|--------|---------|
| Console reviewed (FATAL/NON_FATAL/ANR, 90d) | PASS — verified empty |
| Empty-console root cause established | PASS — FACT, dual cause |
| Firebase/Crashlytics SDK wired to real project | PASS — runtime signature + sink-side settings fetch |
| Startup-crash regression fixed | PASS — alive, zero FATAL |
| Live console round-trip | OPERATOR-ATTENDED (tracked, honest gap) |
| Settings toggle state-binding | TRACKED follow-up (UI desync) |
