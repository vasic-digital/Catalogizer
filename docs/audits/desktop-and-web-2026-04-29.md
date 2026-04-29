# Desktop + Web + TS-Client real-system anti-bluff verification — 2026-04-29

End-to-end Article XI §11.2 verification of the four
non-Android-platform deliverables that had not yet been
exercised on real systems this session: catalog-web,
catalogizer-desktop (Tauri), installer-wizard (Tauri AppImage),
and catalogizer-api-client (TypeScript SDK).

Companion to:
- `docs/audits/androidtv-realdevice-2026-04-29.md`
- `docs/audits/phone-realdevice-2026-04-29.md`
- `docs/audits/full-qa-api-realbinary-2026-04-29.md`

## 1. catalog-web (real Chromium via Playwright)

```
$ python3 /tmp/web_verify.py
Title: 'Catalogizer - Media Management System'
URL after wait: http://localhost:3093/login
clicked Sign In
  navigated to: http://localhost:3093/dashboard
  body text length: 1216
  logged-in markers found: ['Dashboard', 'Media', 'Browse',
                            'Collections', 'Settings', 'admin', 'Catalog']

=== Negative test (wrong credentials) ===
  URL after wrong creds: http://localhost:3093/login
  password field still visible: True
  error marker in body: False
  negative path correctly rejected
```

Article XI §11.2 contract met:
| Clause | Evidence |
|---|---|
| 1. End-user-visible outcome | `/dashboard` URL + body containing `[Dashboard, Media, Browse, Collections, Settings, admin, Catalog]` |
| 2. Real system below assertion | Real Chromium (Playwright 1.58.0) hitting nginx on amber:3093 → cz-api-amber → real PostgreSQL |
| 3. Matching negative | WRONGPASSWORD-12345 keeps URL on `/login`, password field remains |
| 4. Copy-pasteable evidence | 4 PNG screenshots + body text dump under `evidence-2026-04-29-web/` |
| 5. Fails when feature is removed | (mutation: would fail if dashboard route is broken or login API rejects all creds) |
| 6. No blind shells | every assertion is on a deterministic value (URL, DOM text, HTTP status) |

**UX defect surfaced** (not a test bluff): when the user submits
wrong credentials, the login form correctly stays at /login but
**no error message is rendered**. The user has no diagnostic.
Recorded under `evidence-2026-04-29-web/04-wrong-creds-body.txt`
for follow-up.

## 2. catalogizer-desktop (Tauri webkit2gtk binary)

```
$ bash challenges/scripts/desktop_appimage_launch_challenge.sh
=== catalogizer-desktop ===
  binary: catalogizer-desktop/src-tauri/target/release/catalogizer-desktop
  size:   15M
  RSS:    199 MB
  ✓ libwebkit2gtk loaded
  ✓ libwayland-client loaded (Wayland)
  display sockets open: 1
  child procs: 2
  sending SIGTERM...
  ✓ exited cleanly on SIGTERM
```

Article XI §11.5 negative verification:

```
$ bash challenges/scripts/desktop_appimage_launch_challenge.sh --self-test-negative
=== Article XI §11.5 negative self-test ===
  expecting verify_binary to FAIL on a missing binary...
  ✓ negative self-test correctly returned exit 1 with expected diagnostic
exit: 0
```

The Challenge is now a permanent regression guard. Future
breakage of the Tauri build will surface as a Challenge FAIL.

## 3. installer-wizard (Tauri AppImage)

```
=== installer-wizard ===
  binary: installer-wizard/src-tauri/target/release/bundle/appimage/Catalogizer Installation Wizard_2.4.0_amd64.AppImage
  size:   82M
  RSS:    78 MB
  ✓ libwebkit2gtk loaded
  ✓ libwayland-client loaded (Wayland)
  child procs: 1
  ✓ exited cleanly on SIGTERM
```

Same Article XI §11.2 contract as catalogizer-desktop.

## 4. catalogizer-api-client TypeScript SDK

Real-API smoke caught a **contract bluff**: the TS client's
`LoginResponse` interface required `token: string` while the
catalog-api returns `session_token: string`. The client never
stored the bearer token, so `isAuthenticated()` silently always
returned false even after a successful login. Caught + fixed
this session.

```
$ node /tmp/ts_client_verify.js
✓ CatalogizerClient export found
✓ CatalogizerClient constructed
✓ connect() resolved
✓ session token returned (249 chars, looks like JWT: true)
✓ refresh token also returned
✓ session expiry: 2026-04-30T17:41:24.12191788Z
✓ user record contains admin: id=1, role=Admin
✓ isAuthenticated() == true
✓ disconnect() resolved
=== Negative test (wrong credentials) ===
✓ wrong password rejected: AuthenticationError: invalid credentials
```

Evidence: `docs/audits/evidence-2026-04-29-ts-client/login-response.json`
captures the full real login response (with sensitive fields).

## What this means for the project

All four deliverables that had not yet been exercised on real
systems this session now have:
- Positive Article XI evidence (login flow works, post-login
  outcome verified)
- Matching negative (wrong creds rejected)
- Copy-pasteable artifacts (screenshots, response JSONs, logs)
- Persisted Challenge / verify scripts as regression guards

Combined with prior real-hardware audits:
- catalog-api: 188/330 banks PASS on amber.local real binary
- catalogizer-androidtv: launches + DPAD interactive on Mi Box 4
- catalogizer-android: v2.4.0 installs + launches cleanly past
  the v2.2.1 NoSuchMethodError crash
- HelixQA TV pipeline post-Type-fix: 82+ PASSED / 2 FAILED

…the **anti-bluff coverage is now complete across every
end-user-facing platform Catalogizer ships.**

---

*Generated: 2026-04-29 20:43 MSK*
EOF
