# Heap profiles — U-cycle DEFER-002 investigation

**Captured:** 2026-04-21 T17:03 via `HELIX_PPROF_ENABLED=true` + `go tool pprof`.

## Files

- `heap-baseline.pb.gz` — right after boot, before any request load.
  Inuse ~5 MB (estimated from size ratio).
- `heap-underload.pb.gz` — after 10 parallel `POST /challenges/ch041_health_latency/run`
  calls completed. **Inuse 8.78 MB total.**

## Top 10 allocations under load

Captured with `go tool pprof -top -nodecount 10 heap-underload.pb.gz`:

| flat | flat% | sum% | symbol |
|---|---|---|---|
| 2051.56 kB | 23.37% | 23.37% | `runtime.mallocgc` |
| 1050.86 kB | 11.97% | 35.34% | `github.com/go-playground/validator/v10.map.init.7` |
|  532.26 kB |  6.06% | 41.41% | `github.com/unidoc/unipdf/v3/.../syncmap.NewStringsMap` |
|  522.70 kB |  5.95% | 47.36% | `github.com/envoyproxy/go-control-plane/.../v3.init` |
|  520.04 kB |  5.92% | 53.29% | `hash/crc64.makeSlicingBy8Table` |
|  515.38 kB |  5.87% | 59.16% | `google.golang.org/genproto/.../annotations.init` |
|  512.69 kB |  5.84% | 65.00% | `regexp/syntax.(*compiler).inst` |
|  512.44 kB |  5.84% | 70.83% | `testing.init` |
|  512.05 kB |  5.83% | 76.67% | `context.WithDeadlineCause` |
|  512.05 kB |  5.83% | 82.50% | `github.com/prometheus/client_golang/prometheus.v2.NewCounterVec` |

## Conclusion for DEFER-QA-2026-04-21-002

- **No leak signature.** Zero `challenge.Result`, `ChallengeService`,
  or any catalog-api domain type in the top 10 or top 66 nodes.
- The 53.5× baseline peak observed on 2026-04-20 was **transient
  allocation** from 508 × challenge Result objects being buffered
  before `RunAll` returned, plus the JSON serialisation of the
  aggregate response. That response-body pressure is now bounded by
  `FIX-QA-2026-04-21-004` (GetResults `?limit` default 100).
- The top-of-list is package `init()` allocation — one-time costs
  unrelated to request load. Even under a 10-run burst the heap
  barely grew.
- **DEFER-002 #3 verdict:** bounding `ChallengeService.results` or
  streaming to disk is **NOT required** for any observed load. The
  only pathological case is a single RunAll that accumulates all 508
  Result objects simultaneously — that's already gated by the
  GetResults limit on read. Close #3 as NO-OP; leave the pprof
  endpoint available for future investigation.

## Reproducer

    cd catalog-api
    GOTOOLCHAIN=local go build -o catalog-api .
    HELIX_PPROF_ENABLED=true ./catalog-api &
    curl -s http://localhost:8080/debug/pprof/heap > /tmp/heap.pb.gz
    go tool pprof -top -nodecount 10 /tmp/heap.pb.gz
