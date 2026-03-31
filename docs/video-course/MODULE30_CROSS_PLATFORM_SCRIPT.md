# Module 30: Cross-Platform Consistency

## Video Script — Android, Android TV, Desktop, Web — Shared Patterns & Platform Testing

### Duration: ~25 minutes

---

### Scene 1: Introduction (2 min)

"Catalogizer runs on 6 platforms: Go API backend, React web app, Tauri desktop, Tauri installer wizard, Kotlin Android phone, and Kotlin Android TV. This module covers how we maintain consistency across all platforms while respecting platform-specific conventions."

---

### Scene 2: Shared API Contract (4 min)

"All clients communicate through the same REST API. The OpenAPI spec (`docs/api/openapi.yaml`) is the single source of truth."

**Shared patterns across all clients:**
- Authentication: JWT tokens via `POST /api/v1/auth/login`
- Pagination: `?page=1&limit=24` on all list endpoints
- Error format: `{ "error": "message", "code": "ERROR_CODE" }`
- Timestamps: RFC3339 format everywhere
- Content-Type: `application/json` for all API requests

**Challenges CH-221 to CH-230** validate API consistency:
- Consistent JSON structure, error format, pagination, status codes
- Consistent Content-Type, auth requirements, timestamp format

---

### Scene 3: Android Architecture (MVVM) (4 min)

**Directory:** `catalogizer-android/`

```
ViewModel (StateFlow) → Repository → Room + Retrofit
         ↓
Compose UI (observes StateFlow)
```

**Key patterns:**
- Hilt dependency injection
- Room for offline caching
- Retrofit for API calls
- Compose for declarative UI
- `jvmToolchain(17)` required

**Testing:** JUnit + Mockito for ViewModels, Compose test rules for UI

---

### Scene 4: Android TV Architecture (5 min)

**Directory:** `catalogizer-androidtv/`

"Android TV shares the Kotlin/Compose foundation but adds TV-specific patterns."

**TV-specific:**
- D-pad navigation (no touch)
- Focus management with `focusRequester`
- Leanback components for browse/detail layouts
- `dpad_center` BEFORE `type` for text input
- `KEYCODE_TAB` between form fields
- Banner/card grid layouts optimized for 10-foot UI

**Device:** Xiaomi Mi Box 4 (Android 9, 192.168.0.134:5555)
- ADB reverse proxy: `adb reverse tcp:8080 tcp:8080`

---

### Scene 5: Tauri Desktop & Installer Wizard (4 min)

**Directories:** `catalogizer-desktop/`, `installer-wizard/`

"Tauri apps combine a React frontend with a Rust backend, communicating via IPC."

```
React UI ←→ IPC Commands/Events ←→ Rust Backend
```

**Desktop app:** Full media browsing, search, collections, playback
**Installer wizard:** Step-by-step setup — system requirements, database, network, storage roots

**Build:**
```bash
npm run tauri:dev   # development with hot reload
npm run tauri:build # production build
```

---

### Scene 6: Shared TypeScript Libraries (3 min)

"9 TypeScript submodules are shared between web, desktop, and wizard."

| Module | Used By |
|--------|---------|
| @vasic-digital/auth-context | Web, Desktop |
| @vasic-digital/websocket-client | Web, Desktop |
| @vasic-digital/ui-components | Web, Desktop, Wizard |
| @vasic-digital/media-types | All TS clients |
| @vasic-digital/catalogizer-api-client | All TS clients |
| @vasic-digital/media-browser | Web, Desktop |
| @vasic-digital/media-player | Web, Desktop |
| @vasic-digital/collection-manager | Web, Desktop |
| @vasic-digital/dashboard-analytics | Web |

---

### Scene 7: HelixQA Cross-Platform Testing (3 min)

"HelixQA provides autonomous LLM-driven testing across all platforms."

```bash
# API testing
helixqa autonomous --platforms api

# Web testing (Playwright)
helixqa autonomous --platforms web

# Android TV testing (ADB + vision)
helixqa autonomous --platforms androidtv
```

**Pipeline:** Learn → Plan → Execute → Curiosity → Analyze

"The LLM sees screenshots, decides actions, validates results. No hardcoded test flows — fully autonomous."

**User flow challenges:** 49 API + 59 Web + 28 Desktop + 38 Mobile = 174 cross-platform tests

---

### Summary

- 6 platforms, 1 API contract (OpenAPI spec)
- Android: MVVM + Compose + Room + Retrofit + Hilt
- Android TV: D-pad navigation + focus management + Leanback
- Tauri: React + Rust IPC for desktop and installer
- 9 shared TypeScript libraries across web/desktop/wizard
- HelixQA autonomous testing validates all platforms
- 174 user flow challenges ensure cross-platform consistency
- Challenges CH-221-CH-250 enforce API contract compliance
