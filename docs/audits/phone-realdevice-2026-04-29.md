# catalogizer-android (phone APK) — real-device anti-bluff verification — 2026-04-29

End-to-end Article XI §11.5 verification of the **phone** APK
(`com.catalogizer.android`) running on the same Mi Box 4
(Android 9 / SDK 28) used for the TV-app verification — second
positive-evidence verification this session. Companion to
`docs/audits/androidtv-realdevice-2026-04-29.md`.

## Device

```
$ adb -s 192.168.0.214:5555 shell getprop ro.product.model
MIBOX4
$ adb -s 192.168.0.214:5555 shell getprop ro.build.version.release
9
```

`MIBOX4` is **not** in `.devignore` per Constitution Article VII
§7.1, so this device is authorized for testing.

## Anti-bluff evidence (Article XI §11.2.4 — copy-pasteable)

### 1. Bluff-buster: pre-existing v2.2.1 install **crashed on launch**

The phone APK was already installed at v2.2.1 (versionCode 5,
firstInstallTime 2026-04-07) — predating the current
versions.json bump. Launching it immediately threw a
`NoSuchMethodError`:

```
$ adb -s 192.168.0.214:5555 shell am start -n com.catalogizer.android/.ui.MainActivity
$ adb -s 192.168.0.214:5555 logcat -d | grep -A 5 AndroidRuntime
java.lang.NoSuchMethodError: No virtual method
  at(Ljava/lang/Object;I)Landroidx/compose/animation/core/KeyframesSpec$KeyframeEntity;
  in class Landroidx/compose/animation/core/KeyframesSpec$KeyframesSpecConfig;
  ...
  at com.catalogizer.android.ui.splash.SplashContentKt.SplashContent(SplashContent.kt:92)
  at com.catalogizer.android.ui.MainActivityKt.CatalogizerApp(MainActivity.kt:116)
  at com.catalogizer.android.ui.MainActivity$onCreate$3$1$1.invoke(MainActivity.kt:78)
  ...
ActivityManager: Force finishing activity com.catalogizer.android/.ui.MainActivity
ActivityManager: Process com.catalogizer.android (pid 23136) has died
```

This is **a real bug exposed by real-device testing** — an
older shipped APK from April 7 was incompatible with whatever
Compose runtime got bundled, and unit tests / CI did not catch it
because the failure is `Class.method()` resolution at runtime.
Recorded at `docs/audits/evidence-2026-04-29-phone/logcat.txt` +
`crash-screen.png`.

The current source tree is at v2.4.0 / versionCode 6 with
Compose BOM 2024.12.01 — built and installed cleanly (see step 2).

### 2. Fresh v2.4.0 install (positive evidence: package + version)

```
$ adb -s 192.168.0.214:5555 uninstall com.catalogizer.android
Success

$ adb -s 192.168.0.214:5555 install -d \
    releases/android/catalogizer-android/catalogizer-android-v2.4.0-build.25-debug.apk
Performing Streamed Install
Success

$ adb -s 192.168.0.214:5555 shell dumpsys package com.catalogizer.android \
    | grep -E "versionName|versionCode|firstInstallTime|lastUpdateTime"
    versionCode=6 minSdk=26 targetSdk=34
    versionName=2.4.0
    firstInstallTime=2026-04-29 15:42:49
    lastUpdateTime=2026-04-29 15:42:49
```

### 3. v2.4.0 launches cleanly (positive evidence: foreground + process)

```
$ adb -s 192.168.0.214:5555 shell am start -n com.catalogizer.android/.ui.MainActivity
Starting: Intent { cmp=com.catalogizer.android/.ui.MainActivity }

$ adb -s 192.168.0.214:5555 shell dumpsys activity activities | grep mResumedActivity
    mResumedActivity: ActivityRecord{63906e2 u0 com.catalogizer.android/.ui.MainActivity t660}

$ adb -s 192.168.0.214:5555 shell ps | grep "com.catalogizer.android$"
u0_a129  23370  2996 1133384 108420 0  0 R com.catalogizer.android
```

Real PID, real RSS, app is the foreground activity. **The crash
seen on v2.2.1 is gone** — proof that the version bump did
include a fix for this regression.

### 4. Splash → login screen advance (positive evidence: 2 distinct UI states)

```
$ md5sum cz-phone-launch.png cz-phone-after-splash.png
c0445566084a856aa35e15509ace96ec  cz-phone-launch.png
b75fa3889700070958c6a18e0fe8a100  cz-phone-after-splash.png
```

Different hashes prove the screen advanced. Splash UI tree shows:

```
ImageView         bounds=[864,267][1056,459]                          (logo)
TextView          text='Catalogizer'                                  bounds=[810,507][1111,573]
TextView          text='Advanced Multi-Protocol Media\n
                       Collection Management System'                 bounds=[748,589][1173,685]
ProgressBar                                                           bounds=[928,749][992,813]
TextView          text='Made with ❤️ by Vasic Digital'                bounds=[788,912][1133,960]
TextView          text='v2.4.0'                                       bounds=[928,968][994,1016]
```

Post-splash UI tree shows the **login form is wired**:

```
text='Catalogizer'
text='Sign in to your account'
text='Username'
text='Password'
text='Sign In'
text='Server Connection'
text='Server URL'
```

Same package (`com.catalogizer.android`), entirely different
node set from the splash → real composable transition, not a
frozen image.

### 5. No further crashes during 8-second observation

Logcat between splash and post-splash carries only the normal
`MitvPolicyDatabaseManager` package-usage entries —
no `AndroidRuntime: FATAL EXCEPTION`, no `Force finishing`, no
`InputDispatcher: Channel … unrecoverably broken`. The app is
**stable past the splash phase** that crashed v2.2.1.

## Article XI §11.2 contract — point by point

| Clause | Evidence |
|---|---|
| 1. Concrete end-user-visible outcome | login form rendered with 7 expected text labels from `com.catalogizer.android` package |
| 2. Real system below the assertion | Real Android 9 device, real Compose runtime, real ADB install + intent dispatch |
| 3. Matching negative | v2.2.1 *did* crash with `NoSuchMethodError` — captured in logcat — confirming the assertion can fail |
| 4. Copy-pasteable evidence | Full bash sessions pasted above; PNGs + UI XMLs at `docs/audits/evidence-2026-04-29-phone/` |
| 5. Fails when feature is removed | the v2.2.1 logcat *is* the failing-state evidence |
| 6. No blind shells | every assertion is on a deterministic value (`mResumedActivity`, hash, package=, text=) |

## What this session has now proven on real hardware (phone)

- catalogizer-android v2.4.0 build 25 installs, launches without
  crashing, advances past the splash, and renders a real login
  form on a Mi Box 4 / Android 9.
- The previous shipped build (v2.2.1) had a real Compose-runtime
  regression that crashed on launch — and the project's existing
  test suite did not catch it. The fix is implicit in v2.4.0; a
  regression Challenge for "phone APK launches without
  AndroidRuntime FATAL on Android 9" should be added per
  CONST-032 (Reproduction-Before-Fix).

## What is still NOT proven on real hardware (phone)

- Login flow against the deployed catalog-api (same Compose-TV
  scripted-input difficulty as the TV app — pivoting to HelixQA).
- Media browsing / search / playback (would need a populated
  catalogizer DB and a media file).
- Phone APK on a real phone form factor — Mi Box 4 is a TV set-top
  box; rendering on a 6-inch portrait phone is untested here.

## Cross-references

- Constitution Article XI §§ 11.1 – 11.9.
- TV-app verification: `docs/audits/androidtv-realdevice-2026-04-29.md`.
- Real-binary catalog-api verification:
  `docs/audits/full-qa-api-realbinary-2026-04-29.md`.
- Audit ledger: `docs/audits/anti-bluff-2026-04-28.md`.

---

*Generated: 2026-04-29 16:43 MSK*
*Device: Mi Box 4 (MIBOX4) at 192.168.0.214:5555*
*App version: catalogizer-android v2.4.0 build 25 (debug, freshly installed)*
*Previous crash version: catalogizer-android v2.2.1 build 5 (uninstalled)*
