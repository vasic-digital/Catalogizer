# gosec HIGH findings — baseline 2026-04-22

Total: 24  |  Files scanned: 354

## G122 (7 findings)
**Example:** /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api/services/sync_service.go:853
- CWE: 367 https://cwe.mitre.org/data/definitions/367.html
- Details: Filesystem operation in filepath.Walk/WalkDir callback uses race-prone path; consider root-scoped APIs (e.g. os.Root) to prevent symlink TOCTOU traversal

## G704 (4 findings)
**Example:** /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api/internal/services/aggregation_service.go:191
- CWE: 918 https://cwe.mitre.org/data/definitions/918.html
- Details: SSRF via taint analysis

## G703 (4 findings)
**Example:** /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api/services/conversion_service.go:475
- CWE: 22 https://cwe.mitre.org/data/definitions/22.html
- Details: Path traversal via taint analysis

## G101 (3 findings)
**Example:** /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api/internal/tests/mock_servers.go:646-650
- CWE: 798 https://cwe.mitre.org/data/definitions/798.html
- Details: Potential hardcoded credentials

## G118 (3 findings)
**Example:** /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api/internal/services/media_player_service.go:328
- CWE: 400 https://cwe.mitre.org/data/definitions/400.html
- Details: Goroutine uses context.Background/TODO while request-scoped context is available

## G404 (2 findings)
**Example:** /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api/internal/tests/testutils/testdata.go:16
- CWE: 338 https://cwe.mitre.org/data/definitions/338.html
- Details: Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand)

## G115 (1 findings)
**Example:** /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api/challenges/ch201_220_performance_challenges.go:767
- CWE: 190 https://cwe.mitre.org/data/definitions/190.html
- Details: integer overflow conversion uint64 -> int64

