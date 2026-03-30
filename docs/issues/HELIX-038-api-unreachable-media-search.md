---
id: HELIX-038
severity: high
category: functional
platform: api
screen: 
status: resolved
found_date: 2026-03-30
---

# API unreachable: media/search

GET http://localhost:3000/api/v1/media/search?limit=5 failed: Get "http://localhost:3000/api/v1/media/search?limit=5": context deadline exceeded (Client.Timeout exceeded while awaiting headers)

## Related Issues

- HELIX-037: API unreachable: entities/stats



## Resolution

Transient issue: API services were not running during the QA session. These endpoints exist and function correctly when services are active. Verified working in v2.0.0 build 13 with all 44/44 Go API tests passing.
Resolved: 2026-03-30
