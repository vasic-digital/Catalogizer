> **Historical Snapshot** — This report reflects the security scan results as of 2026-02-27. Many issues listed here have since been remediated. See the latest security scan report (2026-03-26) for current status.

---

# Security Scan Report - 2026-02-27

## Overview

Security and code quality scans were performed using multiple tools:
- `go vet` - Go static analysis
- `gosec` - Security scanner
- `staticcheck` - Code quality linter

## Summary

| Tool | Status | Issues Found |
|------|--------|--------------|
| go vet | ✅ PASS | 0 (only external library warnings) |
| gosec | ⚠️ WARNING | 524 issues |
| staticcheck | ⚠️ WARNING | 67 issues |

## gosec Results (524 issues)

### High Severity Issues

#### G115: Integer Overflow (CWE-190)
- `filesystem/ftp_client.go:183` - uint64 -> int64 conversion
- `challenges/auth_token_refresh.go:67` - int -> uint conversion

#### G404: Weak Random Number Generator (CWE-338)
- `internal/tests/testutils/testdata.go:16` - Uses math/rand instead of crypto/rand

#### G704: SSRF via Taint Analysis (CWE)
- `tests/manual/test_auth.go:29,47,58` - HTTP requests with user-controllable URLs

### Medium/Low Severity Issues

- G104: Unhandled errors (multiple locations in challenge files)
- G101: Hardcoded credentials (test configurations)
- G102: Bind to all network interfaces (development configurations)
- G106: SSH insecure ignore host key (WebDAV/SMB clients)

### Recommendations

1. **High Priority**: Fix integer overflow issues with proper bounds checking
2. **Medium Priority**: Use crypto/rand for security-sensitive random generation
3. **Low Priority**: Add error handling for close operations in challenge files

## staticcheck Results (67 issues)

### Categories

| Category | Count | Description |
|----------|-------|-------------|
| SA1029 | 38 | Using built-in string as context key |
| U1000 | 5 | Unused code |
| S1009 | 1 | Redundant nil check |
| SA1012 | 2 | Nil context passed |
| S1019 | 1 | Use make() instead |
| Other | 20 | Various style issues |

### Key Findings

1. **Context Keys (SA1029)**: 38 instances of using `string` as context key
   - Locations: `handlers/*_test.go`, `internal/handlers/*_test.go`
   - Fix: Define custom type for context keys

2. **Unused Code (U1000)**: 5 unused functions/fields
   - `internal/metrics/metrics_integration_test.go:getCounterVecValue`
   - `internal/services/aggregation_service_test.go:newMockAggregationServiceWithRegex`
   - `internal/services/book_recognition_provider.go:generateID`
   - `internal/services/recommendation_service.go:findSimilarBooks, findSimilarGames`
   - `internal/services/recommendation_service.go:igdbClientID, igdbClientSecret`

3. **Nil Context (SA1012)**: 2 instances in `duplicate_detection_service_test.go`

### Recommendations

1. **Context Keys**: Create a custom type for context keys:
   ```go
   type contextKey string
   const userIDKey contextKey = "user_id"
   ```

2. **Unused Code**: Remove or implement unused functions

3. **Nil Context**: Replace nil contexts with `context.TODO()`

## Snyk & SonarQube Status

- **Snyk**: Not configured (requires SNYK_TOKEN environment variable)
- **SonarQube**: Available as container image, requires server setup

### Running Snyk (when token available):
```bash
podman run --rm -e SNYK_TOKEN -v $(pwd):/app docker.io/snyk/snyk-cli:docker snyk test --file=/app/go.mod
```

### Running SonarQube (requires server):
```bash
podman run --rm -v $(pwd):/usr/src sonarsource/sonar-scanner-cli
```

## Next Steps

1. [ ] Fix high severity gosec issues (G115, G404)
2. [ ] Define custom context key type and update all tests
3. [ ] Remove unused code identified by staticcheck
4. [ ] Replace nil contexts with context.TODO()
5. [ ] Configure Snyk for dependency vulnerability scanning
5. [ ] Set up SonarQube server for continuous code quality monitoring
