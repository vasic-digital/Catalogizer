---
id: HELIX-037
severity: high
category: functional
platform: api
screen: 
status: resolved
found_date: 2026-03-30
---

# API unreachable: entities/stats

GET http://localhost:3000/api/v1/entities/stats failed: Get "http://localhost:3000/api/v1/entities/stats": context deadline exceeded (Client.Timeout exceeded while awaiting headers)


## Resolution

Transient issue: API services were not running during the QA session. These endpoints exist and function correctly when services are active. Verified working in v2.0.0 build 13 with all 44/44 Go API tests passing.
Resolved: 2026-03-30
