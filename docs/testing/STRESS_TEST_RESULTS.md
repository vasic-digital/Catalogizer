# Stress Test Results & Baselines

## Test Framework

Stress tests in `catalog-api/tests/stress/` use Go's `testing` package with goroutine-based concurrency.

```bash
cd catalog-api && go test -run TestStress ./tests/stress/ -timeout 10m -count=1
```

## Middleware Chain Stress Tests

**File**: `tests/stress/middleware_chain_stress_test.go`

| Test | Concurrency | Requests | Target |
|------|-------------|----------|--------|
| HighConcurrency | 100 goroutines | 50,000 | Full middleware stack |
| BurstTraffic | 200 burst | 10,000 | Burst handling |
| SustainedLoad | 50 goroutines | 100,000 | Long-duration stability |
| MixedEndpoints | 100 goroutines | 50,000 | Multiple routes |
| ErrorPaths | 100 goroutines | 25,000 | Error handling under load |

### Expected Baselines

| Metric | Baseline | Threshold |
|--------|----------|-----------|
| Success rate | > 99.5% | Must not drop below 99% |
| p95 latency | < 5ms | Must not exceed 10ms |
| Goroutine leak | 0 | Must return to baseline +/- 5 |

## Benchmarks

**File**: `middleware/benchmark_test.go`

```bash
go test -bench=. ./middleware/ -benchmem -count=3
```

## Resource Monitoring

**File**: `tests/monitoring/resource_monitor_test.go`

| Test | Metric | Threshold |
|------|--------|-----------|
| CPUUsage | 1-min load avg | < 8x NumCPU |
| MemoryUsage | HeapAlloc | < 512 MB |
| GoroutineCount | Baseline | < 100 |
| FileDescriptors | Total | < 4096 |

## Fuzz Tests

**File**: `middleware/fuzz_test.go`

```bash
go test -fuzz=FuzzSanitizeInput ./middleware/ -fuzztime=30s
```

5 fuzz targets: SanitizeInput, AuthHeader, URLPath, ContentType, QueryParams.
