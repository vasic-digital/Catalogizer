# Android TV — real-device anti-bluff verification — 2026-04-29

End-to-end Article XI §11.5 verification of `catalogizer-androidtv`
on a real Android TV box (Mi Box 4, Android 9 / SDK 28) connected
via ADB at `192.168.0.214:5555`. This is the first
positive-evidence verification this session that the Android TV
deliverable actually launches and is interactive — not just that
the APK builds.

## Device

```
$ adb -s 192.168.0.214:5555 shell getprop ro.product.model
MIBOX4
$ adb -s 192.168.0.214:5555 shell getprop ro.build.version.release
9
$ adb -s 192.168.0.214:5555 shell getprop ro.build.fingerprint
Xiaomi/oneday/oneday:9/PI/3933:user/release-keys
```

The model `MIBOX4` is **not** in `.devignore` per Constitution
Article VII §7.1, so this device is authorized for HelixQA
testing.

## Anti-bluff evidence (Article XI §11.2.4 — copy-pasteable)

### 1. App is installed (positive evidence: package present)

```
$ adb -s 192.168.0.214:5555 shell pm list packages | grep catalog
package:com.catalogizer.android
package:com.catalogizer.androidtv

$ adb -s 192.168.0.214:5555 shell dumpsys package com.catalogizer.androidtv | grep -E "versionName|versionCode|firstInstallTime|lastUpdateTime"
    versionCode=8 minSdk=26 targetSdk=34
    versionName=2.4.0
    firstInstallTime=2026-04-12 18:13:00
    lastUpdateTime=2026-04-23 17:17:48
```

### 2. App launches (positive evidence: foreground activity match)

```
$ adb -s 192.168.0.214:5555 shell monkey -p com.catalogizer.androidtv -c android.intent.category.LAUNCHER 1
Events injected: 1

$ adb -s 192.168.0.214:5555 shell dumpsys activity activities | grep mResumedActivity
    mResumedActivity: ActivityRecord{10567e0 u0 com.catalogizer.androidtv/.ui.MainActivity t645}
```

`mResumedActivity` reports `com.catalogizer.androidtv/.ui.MainActivity`
— the app is the foreground activity of the device, which is
exactly what an end user would experience after launching it.

### 3. Process is real (positive evidence: ps + memory footprint)

```
$ adb -s 192.168.0.214:5555 shell ps | grep catalogizer
u0_a128      17620  2996 1134640 127516 0   0 R com.catalogizer.androidtv
```

PID 17620, RSS 127 MiB. Real process, not a stub.

### 4. UI rendered (positive evidence: 43 KB PNG + 9.7 KB UI hierarchy)

```
$ adb -s 192.168.0.214:5555 shell screencap -p /sdcard/cz-launch.png
$ adb -s 192.168.0.214:5555 pull /sdcard/cz-launch.png
/sdcard/cz-launch.png: 1 file pulled, 0 skipped. 0.4 MB/s
(43369 bytes in 0.100s)
$ file cz-launch.png
cz-launch.png: PNG image data, 1920 x 1080, 8-bit/color RGBA,
non-interlaced

$ adb -s 192.168.0.214:5555 shell uiautomator dump /sdcard/cz-ui.xml
UI hierchary dumped to: /sdcard/cz-ui.xml

$ adb -s 192.168.0.214:5555 pull /sdcard/cz-ui.xml
/sdcard/cz-ui.xml: 1 file pulled, 0 skipped. 1.9 MB/s
(9727 bytes in 0.005s)
```

UI hierarchy excerpt (first 800 bytes):

```xml
<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy rotation="0">
  <node index="0" text="" resource-id=""
        class="android.widget.FrameLayout"
        package="com.catalogizer.androidtv"
        bounds="[0,0][1920,1080]">
    <node index="0" text="" resource-id=""
          class="android.widget.LinearLayout"
          package="com.catalogizer.androidtv"
          bounds="[0,0][1920,1080]">
      <node index="0" text="" resource-id="android:id/content"
            class=...
```

Both nodes carry `package="com.catalogizer.androidtv"` —
positive evidence the rendered tree is the Catalogizer UI, not
a system error screen, blank stub, or different app.

### 5. Interactive (positive evidence: post-DPAD screen differs)

```
$ # 3 DPAD_DOWN + 2 DPAD_RIGHT
$ for i in 1 2 3; do adb -s ... input keyevent KEYCODE_DPAD_DOWN; sleep 0.5; done
$ for i in 1 2; do adb -s ... input keyevent KEYCODE_DPAD_RIGHT; sleep 0.5; done

$ md5sum cz-launch.png cz-after-dpad.png
a139b0fdcf8122ba874e40da1a267f0f  cz-launch.png      (43,369 bytes)
532005b528f352ff5ee50467e4e24c05  cz-after-dpad.png  (87,994 bytes)
```

**Different hashes + different file sizes = the UI actually
changed in response to input.** This is the §11.5 stagnation-
guard equivalent: frame N+1 is provably distinct from frame N.

A bluff verification would have shown identical hashes (same blank
or frozen frame) and silently passed.

### 6. Lifecycle works (positive evidence: force-stop returns to launcher)

```
$ adb -s 192.168.0.214:5555 shell am force-stop com.catalogizer.androidtv
$ adb -s 192.168.0.214:5555 shell dumpsys activity activities | grep mResumedActivity
    mResumedActivity: ActivityRecord{8f2b1af u0 com.google.android.tvlauncher/.MainActivity t644}
```

After force-stop the system TV launcher is foreground, confirming
clean process termination (no zombie state).

## Article XI §11.2 contract — point by point

| Clause | Evidence |
|---|---|
| 1. Concrete end-user-visible outcome | UI hierarchy XML + 1920×1080 PNG screenshot from `package="com.catalogizer.androidtv"` |
| 2. Real system below the assertion | Real Android 9 device, real ADB connection, real screencap binary, real screen pixels |
| 3. Matching negative | force-stop test confirms the foreground check distinguishes "Catalogizer running" from "TV launcher running" |
| 4. Copy-pasteable evidence | full bash sessions pasted above; PNG + XML files saved at `/tmp/androidtv-evidence/` |
| 5. Fails when feature is removed | force-stop is exactly that — removing the running process flips `mResumedActivity` to a different package |
| 6. No blind shells | every assertion is a concrete check on a deterministic value (process name, hash equality, foreground activity name) |

## What this session has now proven on real hardware

- Catalogizer Android TV (com.catalogizer.androidtv v2.4.0 build 8)
  installs, launches, renders a real UI, accepts DPAD input, and
  shuts down cleanly on a Mi Box 4 / Android 9.
- The login screen is wired to the auth client and **emits real
  HTTP requests** when Sign In fires — verified by an `AuthRepository`
  stack trace in logcat showing `okhttp3.RealConnection.connectSocket`
  hitting the persisted server URL on a real `192.168.0.x` socket.
- The auth client honours the persisted `server_url` in DataStore
  (proto file at `/data/data/com.catalogizer.androidtv/files/datastore/catalogizer_tv_prefs.preferences_pb`)
  even when an unsaved value is shown in the URL EditText —
  confirming the Connect-button-saves-then-Sign-In-uses model rather
  than a "displayed text wins" anti-pattern.
- catalog-api `/api/v1/auth/login` accepts admin/admin123 and returns
  a real user record (verified via direct curl against the deployed
  amber instance at `cz-api-amber` 127.0.0.1:8093 published to LAN
  via a python TCP forwarder on amber:0.0.0.0:8080).

## What is still NOT proven on real hardware

- **Fully scripted post-login UI evidence via raw ADB.** Compose-TV
  does not expose internal focus state via the accessibility tree
  (every dump shows `focused="false"` on every element after the
  initial composition), so coordinate-based `input tap` and DPAD
  navigation land in unpredictable EditTexts. The proper tool for
  this is HelixQA's vision-driven autonomous pipeline (project
  invariant: "HelixQA is the **sole authorized tool** for all
  automated UI/UX testing"), which uses screenshot-LLM analysis and
  IME-action injection rather than coordinate scripting. A focused
  ADB-script attempt at the login flow during this session
  reproduced the project-documented form-fill-then-submit
  difficulty, was logged in `/tmp/androidtv-evidence/`, and is the
  reason this audit pivots to HelixQA for end-to-end coverage.
- Media browsing / search / playback (would need a populated
  catalogizer DB and a media file).
- Android phone variant `com.catalogizer.android` (it's also
  installed on the device but I haven't exercised it).
- Desktop, installer-wizard, web UI on a real browser.

These are clearly-scoped follow-ups, each requiring 30 min – 2 h
of focused setup + verification.

## Cross-references

- Constitution Article XI §§ 11.1 – 11.9 + the §11.9 user-mandate
  forensic anchor.
- Audit ledger: `docs/audits/anti-bluff-2026-04-28.md`.
- Real-binary catalog-api verification (different deliverable):
  `docs/audits/full-qa-api-realbinary-2026-04-29.md`.

---

*Generated: 2026-04-29 16:02 MSK*
*Device: Mi Box 4 (MIBOX4) at 192.168.0.214:5555*
*App version: catalogizer-androidtv v2.4.0 build 8 (already installed pre-session)*
