# Desktop + Performance Audit — Master Plan Phase 10 + Phase 13

> **Purpose.** Master Plan v2 Phase 10 "Desktop Hardening" (5 days)
> requires Tauri builds across the 4-OS matrix + zero bare Rust
> `unwrap()`. Phase 13 "Performance & Stress Testing" (5 days)
> requires a k6 load battery and performance budgets on every
> endpoint group. This audit (2026-04-22) inventories the current
> baseline.

## 1. Phase 10 — Desktop (Tauri)

### 1.1 Tauri projects

Two Tauri/Rust+React apps:

| App | Source tree | Purpose |
|---|---|---|
| `catalogizer-desktop` | `src-tauri/src/main.rs` + `src-tauri/src/vlc/` | Full desktop app with VLC media playback |
| `installer-wizard` | `src-tauri/src/main.rs` + `ftp.rs`, `local.rs`, `network.rs`, `nfs.rs` | First-run configurator with per-protocol UI |

### 1.2 Rust unwrap() hygiene (RULE-DESK-001)

```bash
scripts/detect-landmines.sh   # includes walk-the-AST Rust check
```

Output:
```
✓ RULE-DESK-001: catalogizer-desktop Rust unwrap() clean (non-test code)
✓ RULE-DESK-001: installer-wizard Rust unwrap() clean (non-test code)
```

The script walks each `.rs` file, skips `#[test]` functions and
`#[cfg(test)]` modules, and flags any bare `unwrap()` that isn't
followed by `// SAFE: <reason>`. **Zero production unwrap()**
violations.

### 1.3 Phase 10 Exit Criteria

| Criterion | Target | Status |
|---|---|:-:|
| Tauri build on macOS (Intel) | `npm run tauri build` | ⏳ needs macOS hardware |
| Tauri build on macOS (Apple Silicon) | `npm run tauri build` | ⏳ needs aarch64 mac |
| Tauri build on Windows 11 | `npm run tauri build` | ⏳ needs Windows host |
| Tauri build on Ubuntu | `npm run tauri build` (rootless container) | ✅ CI-capable |
| Zero Rust `unwrap()` in non-test code | 0 | ✅ |
| Auto-updater functional | E2E check→download→install | ⏳ staging verify |
| Native file dialogs | Smoke test | ⏳ per-OS |
| System tray | Smoke test | ⏳ per-OS |
| Keyboard shortcuts | Smoke test | ⏳ per-OS |
| Drag-and-drop | Smoke test | ⏳ per-OS |

**Automated gate (unwrap cleanliness) closed. Cross-OS build + smoke
tests remain operator tasks** — same pattern as Phase 7 cross-browser
and Phase 14 video recording.

## 2. Phase 13 — Performance + Stress Testing

### 2.1 k6 script inventory (`tests/k6/`)

15 scripts present — well over the master plan's 3-script baseline
(load + stress + soak):

| Script | Purpose |
|---|---|
| `load_test.js` | Baseline — ramp to 50 users, verify p95 < 500 ms |
| `stress_test.js` | Ramp to 300 users, find breaking point |
| `soak_test.js` | 20 users for 30 min, detect memory leaks |
| `breakpoint_test.js` | Incrementally increase until failure |
| `spike_test.js` | Sudden traffic surge (0 → N users in seconds) |
| `endurance_test.js` | Multi-hour sustained load |
| `auth_load_test.js` | `/auth/login` + `/auth/refresh` under concurrency |
| `ddos_ratelimit_test.js` | Verify rate-limit kicks in under DDoS simulation |
| `entity_browse_load_test.js` | `/entities/browse/:type` hot path |
| `concurrent_writers_test.js` | Multiple clients writing simultaneously |
| `database_stress_test.js` | DB-specific stress (connection pool, dialect rewriting) |
| `media_scan_stress_test.js` | `/scan` under many concurrent scan triggers |
| `mixed_workload_test.js` | Realistic mix of reads + writes + scans |
| `websocket_stress_test.js` | WS fan-out + reconnect loops |
| `monitoring_test.js` | Verifies `/metrics` stays responsive during load |

### 2.2 Master plan §12.1 performance budget targets

| Endpoint | p50 | p95 | p99 | Baseline status |
|---|---|---|---|---|
| `GET /api/v1/media` | 50 ms | 100 ms | 200 ms | ⏳ needs run |
| `POST /api/v1/auth/login` | 100 ms | 200 ms | 500 ms | ⏳ needs run |
| `GET /api/v1/media/:id` | 30 ms | 50 ms | 100 ms | ⏳ needs run |
| WebSocket events | 10 ms | 20 ms | 50 ms | ⏳ needs run |

### 2.3 How to run the battery

```bash
# Single script
podman run --rm --network host -v $(pwd)/tests/k6:/scripts \
  docker.io/grafana/k6:latest run /scripts/load_test.js

# Full battery
for script in tests/k6/*.js; do
  name=$(basename "$script" .js)
  echo "=== $name ==="
  podman run --rm --network host -v $(pwd)/tests/k6:/scripts \
    docker.io/grafana/k6:latest run "/scripts/$(basename $script)"
done
```

### 2.4 Phase 13 Exit Criteria

| Criterion | Status |
|---|:-:|
| All API endpoints meet p50/p95/p99 targets | ⏳ needs run against staging |
| 100 concurrent users × 10 min — zero 5xx | ⏳ `stress_test.js` |
| 10,000-media library scan < 1 h | ⏳ `media_scan_stress_test.js` |
| Search < 500 ms on 10k library | ⏳ `entity_browse_load_test.js` |
| Memory < 2 GB under sustained load | ⏳ `soak_test.js` + pprof heap |
| No memory leaks on 30-min soak | ⏳ `soak_test.js` |

## 3. Summary

**Phase 10 automated gate** (unwrap hygiene) is closed. Cross-OS
Tauri builds need operator hardware access — not automatable on a
single Linux dev box.

**Phase 13 infrastructure** is massively over-spec (15 k6 scripts vs
3-script baseline). Actual **run-against-staging** and budget-check
is the remaining operator task; scripts cover every dimension the
master plan's §12.1 budget table requires.

Both phases track the same pattern as Phase 7 + Phase 14:
**automated gates closed now, interactive/hardware gates queued for
the final integration cycle (Phase 15)**.
