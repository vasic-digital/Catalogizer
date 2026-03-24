# Service Discovery Feature — Design Document

**Date:** 2026-03-24
**Status:** Phases 1-6 Complete, Phase 7 In Progress
**Priority:** Critical

## Overview

All Catalogizer client apps must automatically discover backend API instances on the local network. No hardcoded API endpoint addresses are allowed. Users must be able to add/remove custom API endpoint entries.

## Architecture

### Backend (catalog-api)

1. **Enhanced /api/v1/discovery endpoint** — Returns full service info:
   ```json
   {
     "service": "catalogizer-api",
     "version": "1.0.0",
     "host": "192.168.0.213",
     "port": 8080,
     "protocol": "http",
     "capabilities": ["catalog", "media", "streaming", "sync"],
     "uptime": 3600
   }
   ```

2. **UDP Broadcast Announcer** — Broadcasts service presence on port 42069 (configurable) every 5 seconds:
   - Multicast group: `239.42.42.42:42069`
   - Message format: JSON with service info
   - Runs as goroutine, stopped on shutdown

3. **UDP Discovery Responder** — Listens for client discovery queries on the same port and responds with service info

### Discovery Module (digital.vasic.discovery)

New package: `pkg/broadcast/`
- `Announcer` — Sends periodic UDP multicast announcements
- `Listener` — Listens for announcements, returns discovered services
- `Config` — Multicast group, port, interval, timeout

### Client Apps

All clients implement the same discovery flow:

1. **On app launch / login screen load**:
   - Check stored server URL in local storage/preferences
   - If stored URL exists and is reachable → use it
   - If not → start network discovery

2. **Network Discovery**:
   - Listen on UDP multicast for announcements (3-5 seconds)
   - If servers found → show in a picker list
   - If only one → auto-select
   - If none → show manual entry field

3. **Server URL Management**:
   - Add custom server URL manually
   - Remove saved entries
   - Default fallback: `https://catalogizer.dev` (production)
   - Persist in platform-specific storage

### Platform-Specific Implementation

| Platform | Storage | Discovery Client | UI Component |
|----------|---------|-----------------|--------------|
| Android TV | DataStore | UDP multicast listener (NsdManager) | ServerPickerDialog on LoginScreen |
| Android Phone | DataStore | NsdManager | ServerPickerDialog on LoginScreen |
| Desktop (Tauri) | IPC config | Rust UDP listener via IPC | ServerUrlField on LoginPage |
| Installer Wizard | Tauri config | Rust UDP listener | ServerUrlField on WelcomeStep |
| Web | localStorage | HTTP fetch to known ports | ServerUrlField on LoginForm |

### API Changes

**Existing endpoint enhanced:**
```
GET /api/v1/discovery
Response: {
  "service": "catalogizer-api",
  "version": "1.0.0",
  "build": "677",
  "host": "<detected-ip>",
  "port": 8080,
  "protocol": "http",
  "capabilities": ["catalog", "media", "streaming", "sync", "websocket"],
  "uptime_seconds": 3600,
  "database": "sqlite",
  "storage_roots": 2
}
```

**New endpoint (public, no auth):**
```
GET /api/v1/discovery/announce
Response: Same as above (for HTTP-based discovery)
```

### UDP Protocol

```
Multicast Group: 239.42.42.42
Port: 42069
Interval: 5 seconds

Announcement Message (JSON):
{
  "type": "catalogizer-announce",
  "version": "1.0.0",
  "host": "192.168.0.213",
  "port": 8080,
  "name": "Catalogizer API",
  "id": "<unique-instance-id>"
}

Discovery Query:
{
  "type": "catalogizer-discover"
}
```

## Implementation Phases

### Phase 1: Backend (catalog-api + Discovery module)
- [ ] Add `pkg/broadcast/` to Discovery module (Announcer + Listener)
- [ ] Enhance `/api/v1/discovery` with full service info
- [ ] Add `/api/v1/discovery/announce` (public, no auth)
- [ ] Start UDP announcer on API startup
- [ ] Stop announcer on graceful shutdown
- [ ] Unit tests for broadcast package

### Phase 2: Android TV
- [ ] Add `ServerConfigRepository` (DataStore-based)
- [ ] Add `DiscoveryViewModel` with NsdManager or UDP multicast
- [ ] Modify `LoginScreen.kt` — add server URL field + discovery button
- [ ] Add `ServerPickerDialog` component
- [ ] Remove hardcoded `BuildConfig.API_BASE_URL`
- [ ] Make `CatalogizerApi` use dynamic base URL from repository
- [ ] Persist selected server, auto-reconnect on app restart

### Phase 3: Android Phone
- [ ] Same as Phase 2, adapted for phone UI

### Phase 4: Web + Desktop + Wizard
- [ ] Web: Add server URL field to LoginForm, HTTP-based discovery
- [ ] Desktop: Add IPC command for discovery, server URL in settings
- [ ] Wizard: Add discovery to welcome step

### Phase 5: Tests
- [ ] Unit tests for Discovery broadcast package
- [ ] Unit tests for each client's discovery implementation
- [ ] Challenges: discovery endpoint validation
- [ ] HelixQA bank: discovery UI tests for all platforms
- [ ] Integration test: API announces, client discovers

### Phase 6: Build + Deploy + QA Loop
- [ ] Rebuild all apps with discovery
- [ ] Increment all version codes
- [ ] Install Android TV via ADB
- [ ] HelixQA multi-pass curiosity testing
- [ ] Fix all discovered issues
- [ ] Repeat until zero defects

### Phase 7: Documentation
- [ ] Update CLAUDE.md with discovery architecture
- [ ] Update AGENTS.md with discovery commands
- [ ] Update API_REFERENCE.md
- [ ] Update user guides
- [ ] Update video course materials
- [ ] Update architecture diagrams

## Constraints

- No hardcoded API URLs in any app (except default fallback to catalogizer.dev)
- Discovery must work on LAN without internet
- Must handle multiple API instances on the same network
- Must persist user's server selection across app restarts
- UDP broadcast must be stoppable (not leak goroutines)
- All changes must have full test coverage
- Zero console errors, zero network failures policy maintained
