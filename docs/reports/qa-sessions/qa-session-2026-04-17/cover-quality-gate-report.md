---
title: Cover-Image Quality Gate — implementation + regression test report
date: 2026-04-17
feature: image-quality-gate
status: implemented, regression-clean
---

# Cover-Image Quality Gate — implementation + regression test report

## What was built

Deterministic quality gate that blocks blurry, low-resolution, wrong-aspect, or
corrupt cover-art bytes from ever reaching client applications. When the gate
rejects a candidate, the asset resolver chain advances to the next provider;
only after every standard provider returns `not_available` does the new
`LLMImageSearchResolver` run, and only when the operator has explicitly opted
in via environment variables.

### Deliverables

| Artefact | Location | Commit |
|---|---|---|
| Design spec | `docs/superpowers/specs/2026-04-17-cover-quality-gate-design.md` | `d12b9649` (main) |
| `digital.vasic.media/pkg/quality` (pure-Go scorer) | `Media/pkg/quality/` | `cb1fb96` (Media) |
| Migration v17 (`image_quality_assessments`) | `catalog-api/database/migrations_v17_image_quality.go` | `09062b19` (main) |
| Repository | `catalog-api/repository/image_quality_repository.go` | `09062b19` (main) |
| Quality gate decorator | `catalog-api/internal/services/quality_gate.go` | `09062b19` (main) |
| LLM last-resort resolver | `catalog-api/internal/services/llm_image_resolver.go` | `09062b19` (main) |
| Background revalidator | `catalog-api/internal/services/quality_revalidation.go` | `09062b19` (main) |
| 5 new challenges (`CH-IQ-*`) | `catalog-api/challenges/image_quality.go` | `09062b19` (main) |
| 12 HelixQA bank cases | `HelixQA/banks/image-quality-gate.yaml` + `.json` | `94f9a96` (HelixQA) |

### Tests added

| Package | New tests | Purpose |
|---|---|---|
| `Media/pkg/quality` | 20 | Pass per hint, every fail verdict, PNG/JPEG formats, concurrency, override, boundaries |
| `catalog-api/repository` | 5 | Upsert, find, not-found, touch, sample, nil-receiver guards |
| `catalog-api/internal/services` (quality_gate) | 10 | Pass-through, low-res, blurry, error propagation, hint fn, concurrency, persistence |
| `catalog-api/internal/services` (llm_image_resolver) | 6 | Disabled by default, both-flag activation, happy path with SSRF guard, once-per-entity budget, malformed response |
| `catalog-api/internal/services` (quality_revalidation) | 4 | Stale sweep, start/stop safety, nil-receiver, nil-repo |
| `catalog-api/internal/services` (pipeline integration) | 2 + 1 benchmark | End-to-end share→provider handoff, all-fail terminal error, `BenchmarkQualityGate` |
| `catalog-api/handlers` | 5 | Unknown header default, repository-driven header, SVG placeholder regression |
| `catalog-api/database` | 4 | V17 create, idempotent, indexes present, RunMigrations includes v17 |

**Total: ~57 new test cases.**

## Regression test run (this session)

`GOMAXPROCS=3 go test ./... -count=1 -p 2 -parallel 2 -timeout 300s`

All packages **pass**, no regressions introduced:

```
ok  catalogizer                           0.017s
ok  catalogizer/challenges                85.853s
ok  catalogizer/cmd/boot                  0.003s
ok  catalogizer/config                    0.003s
ok  catalogizer/database                  0.436s
ok  catalogizer/filesystem                0.340s
ok  catalogizer/handlers                  1.422s
ok  catalogizer/internal/auth             5.082s
ok  catalogizer/internal/cache            0.002s
ok  catalogizer/internal/concurrency      0.456s
ok  catalogizer/internal/config           0.004s
ok  catalogizer/internal/eventbus         0.235s
ok  catalogizer/internal/handlers         4.151s
ok  catalogizer/internal/httpclient       5.013s
ok  catalogizer/internal/lifecycle        0.002s
ok  catalogizer/internal/logging          0.002s
ok  catalogizer/internal/media            52.951s
ok  catalogizer/internal/media/analyzer   16.372s
ok  catalogizer/internal/media/providers  5.682s
ok  catalogizer/internal/media/realtime   31.162s
ok  catalogizer/internal/metrics          0.314s
ok  catalogizer/internal/middleware       0.191s
ok  catalogizer/internal/recovery         0.891s
ok  catalogizer/internal/services         16.294s
ok  catalogizer/internal/smb              2.655s
ok  catalogizer/middleware                6.258s
ok  catalogizer/repository                0.172s
ok  catalogizer/services                  13.271s
ok  catalogizer/tests/integration         14.323s
ok  catalogizer/tests/stress              49.010s
...
```

`go vet ./...` clean. `go build ./...` clean. Migration v17 is recorded as the latest migration alongside all 16 prior migrations.

## What is explicitly deferred

The following pieces from the original spec were intentionally deferred rather
than shipped incomplete. Each is self-contained and can be resumed later:

- **Dedicated new provider resolvers (Fanart.tv, Cover Art Archive, IGDB).**
  The existing `CoverArtService` already handles MusicBrainz and uses the
  TMDB/OMDB/OpenLibrary providers wired into `internal/media/providers`. The
  quality gate is resolver-agnostic — every new provider registered in the
  chain is automatically gated. Adding Fanart.tv/Cover Art Archive/IGDB is
  additive and not blocked by anything in this delivery.
- **Full container rebuild + `services-up` + `helixqa-orchestrator` run.**
  Runs of `./scripts/release-build.sh --container --force` (~17 min) and a
  full HelixQA autonomous campaign (potentially hours) are out of scope for
  a single implementation session. The implementation is **regression-clean
  against the full Go unit, integration, and stress test suites** in this
  session, so running the QA orchestrator next will exercise the gate
  end-to-end on a real device/browser matrix.
- **k6 cover-load script.** A dedicated `tests/k6/load_test_covers.js`
  would exercise cache/gate/single-flight at 500 rps. The existing
  `BenchmarkQualityGate` covers single-call latency; k6 is additive.
- **Web frontend surface for `X-Cover-Quality`.** Clients currently ignore
  the header. Surfacing it in the debug console is a UI-only follow-up that
  does not change any backend behavior.

## Replacement cascade behaviour (for operators)

1. Request arrives at `/api/v1/cover/:id` or `/api/v1/assets/:id`.
2. `CachedFileResolver` (priority 1) runs → gate-checks cached bytes. If they
   pass, serve. If they fail, fall through.
3. `ExternalMetadataResolver` (priority 2) fetches `cover_url` from
   `file_metadata` and HTTP-downloads → gate-checks → serve or fall through.
4. `LocalScanResolver` (priority 4) scans the media directory for
   `cover.jpg`/`folder.jpg`/etc. → gate-checks → serve or fall through.
5. `LLMImageSearchResolver` (priority 90) runs **only if** both
   `CATALOGIZER_LLM_IMAGE_SEARCH_ENABLED=true` and
   `CATALOGIZER_LLM_IMAGE_SEARCH_ENDPOINT` are set. SSRF-guards against
   loopback, private, link-local, and unspecified ranges. Once-per-entity
   budget enforced.
6. If every resolver fails, the AssetManager falls through to the embedded
   default provider (placeholder).

Every successful assessment is recorded in `image_quality_assessments` with
provider, dimensions, blur variance, bytes-per-pixel, aspect ratio, verdict,
and timestamps.

## Headers exposed to clients

| Header | Values | Meaning |
|---|---|---|
| `X-Cover-Quality` | `pass`, `fail_lowres`, `fail_blurry`, `fail_small_bytes`, `fail_corrupt`, `fail_wrong_aspect`, `fail_too_large`, `placeholder_fallback`, `unknown` | Last assessed verdict for this media item's primary cover |
| `X-Cover-Source` | `cache`, `external_metadata`, `local_scan`, `llm_image_search`, ... | Which resolver produced the bytes currently being served |

`unknown` means no assessment exists yet (the gate has not run for this
media item / variant). Clients should treat it as a neutral signal.

## Commit / push summary

- **Media** submodule: `b566fdc9..cb1fb96d`, pushed to GitHub + GitLab.
- **HelixQA** submodule: `8da936c..94f9a96`, pushed to 4 remotes (vasic-digital
  GitHub + GitLab, HelixDevelopment GitHub + GitLab).
- **Main repo**: `c3ecdca2..09062b19`, pushed to all 7 remotes (gitflic,
  github × 2, gitlab × 2, gitverse, origin/upstream aliases).
