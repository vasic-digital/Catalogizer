---
title: Cover-Image Quality Gate + Upstream Replacement Cascade
date: 2026-04-17
status: approved
owner: catalog-api + Media submodule
---

# Cover-Image Quality Gate + Upstream Replacement Cascade

## Problem

Client apps (web, Android TV, Android phone, desktop) currently receive whatever cover image the scan pipeline found on the media share. Some of those images are low resolution, blurry, wrong aspect ratio, or corrupt. Serving them degrades the UX and breaks brand guidelines. Users cannot easily tell which cover is "good enough" without LLM help, and we already have metadata providers (TMDB, Fanart.tv, OMDB, OpenLibrary, MusicBrainz/Cover Art Archive, IGDB) that host vetted high-resolution artwork for most titles we host.

## Goal

- No blurry, low-resolution, wrong-aspect, or corrupt cover ever leaves the API.
- If the share's copy fails the quality gate, the pipeline fetches a replacement from the metadata providers. Only if every standard provider fails do we fall back to an LLM-mediated image search.
- Accepted replacements are cached so we never pay the gate cost twice for the same entity.
- Every media type (11 types, per `CLAUDE.md`) is covered.
- 100% test coverage across all 10 constitution categories.

## Non-goals (YAGNI)

- No new image CDN, transcoder, or format converter.
- No new LLM providers; reuse `LLMProvider` + `LLMOrchestrator` submodules.
- No gating for UI icons or the placeholder asset itself.
- No override of user-set `user_metadata.override_cover_url`.

## Pipeline placement

```
AssetManager.Resolve(ctx, req)
 1. CachedFileResolver          -> hit? assess -> pass: serve | fail: purge + continue
 5. ShareSourceResolver [NEW]   -> pull bytes from scan source (SMB/FTP/local) -> assess
10. TMDBResolver                (movies / TV)
11. FanartTVResolver [NEW]      (high-res posters + backdrops)
12. OMDBResolver
20. OpenLibraryResolver         (books) — request -L.jpg for largest size
21. CoverArtArchiveResolver [NEW] (music)
22. IGDBResolver [NEW]          (games)
90. LLMImageSearchResolver [NEW] (last resort, gated by "no working API")
999. PlaceholderResolver        (final fallback)
```

Every resolver result walks through `QualityGate.Check()`. First pass wins and is written into `cache/cover_art/`. Failed candidates are discarded and the chain continues.

## Deterministic quality scoring

New Go package: `digital.vasic.media/pkg/quality` (extracted to `Media/` submodule for reuse by non-API consumers).

```go
type Hint string

const (
    HintMoviePoster Hint = "movie_poster"
    HintTvPoster    Hint = "tv_poster"
    HintMusicAlbum  Hint = "music_album"
    HintBookCover   Hint = "book_cover"
    HintGameCover   Hint = "game_cover"
    HintBackdrop    Hint = "backdrop"
    HintGeneric     Hint = "generic"
)

type Score struct {
    Width, Height int
    Megapixels    float64
    BytesPerPixel float64
    BlurVariance  float64
    AspectRatio   float64
    IsCorrupt     bool
    Verdict       Verdict
    FailReason    string
}

type Verdict int

const (
    Pass Verdict = iota
    FailLowRes
    FailBlurry
    FailSmallBytes
    FailCorrupt
    FailWrongAspect
    FailTooLarge // decompression-bomb guard
)

func Assess(data []byte, h Hint) (Score, error)
```

Per-hint thresholds (configurable via `config.json` → `image_quality`):

| Hint           | MinW | MinH | MinBlurVar | MinBPP | AspectTarget | AspectTol |
|----------------|------|------|------------|--------|--------------|-----------|
| movie_poster   | 600  | 900  | 80         | 0.40   | 2:3          | 0.05      |
| tv_poster      | 600  | 900  | 80         | 0.40   | 2:3          | 0.05      |
| music_album    | 500  | 500  | 70         | 0.35   | 1:1          | 0.03      |
| book_cover     | 400  | 600  | 60         | 0.30   | 2:3          | 0.10      |
| game_cover     | 600  | 800  | 80         | 0.40   | 3:4          | 0.10      |
| backdrop       | 1280 | 720  | 100        | 0.50   | 16:9         | 0.05      |
| generic        | 300  | 300  | 60         | 0.25   | any          | -         |

Blur detection: 3×3 Laplacian kernel on the luminance channel of the decoded image, return variance across all pixels. Pure Go, using stdlib `image/*` plus `golang.org/x/image/webp` and `golang.org/x/image/bmp`. No CGo, keeps container images small. Decompression-bomb defense: reject any image with `Width * Height > 64e6` pixels before full decode.

## Replacement cascade behavior

- Order is fixed and deterministic. All standard providers are tried before any LLM call.
- Per-provider budget: 5-second HTTP timeout, retry-with-backoff via existing `httpclient` helper, circuit breaker via `digital.vasic.recovery`. Quota exceeded / 4xx / 5xx / DNS failure / gate failure of returned candidate → treated as `not_available`, chain continues.
- **LLM last-resort trigger**: `LLMImageSearchResolver.Resolve()` fires only if every standard resolver returned `not_available` OR every returned candidate failed the gate. Implementation uses `LLMProvider` + `LLMOrchestrator` with Gemini web-search grounding to locate publicly listed high-res images. Downloaded bytes must themselves pass the gate before being cached. Budget: one LLM attempt per `(entity_id, variant)`; hit or miss is recorded so we never retry.
- **All-fail path**: return the existing `/api/v1/cover/placeholder/:type` with response header `X-Cover-Quality: placeholder_fallback` and a WARN log. HelixQA surfaces placeholder fallbacks.
- Single-flight deduplication (using existing `Concurrency` submodule primitives) prevents cache stampede when many clients request the same missing cover simultaneously.

## Caching and persistence

- **Byte cache (existing)**: `cache/cover_art/<sha256(entity_type|entity_id|variant)>.jpg`. `CachedFileResolver.StoreResult()` already handles this.
- **DB metadata (new)**: migration v10 adds

```sql
CREATE TABLE image_quality_assessments (
    id             INTEGER PRIMARY KEY,
    entity_type    TEXT NOT NULL,
    entity_id      INTEGER NOT NULL,
    variant        TEXT NOT NULL,
    source         TEXT NOT NULL,   -- share | tmdb | fanarttv | omdb | openlibrary | cover_art_archive | igdb | llm | placeholder
    provider_ref   TEXT,            -- provider-assigned id / URL fragment
    width          INTEGER NOT NULL,
    height         INTEGER NOT NULL,
    blur_var       REAL NOT NULL,
    bpp            REAL NOT NULL,
    aspect_ratio   REAL,
    verdict        TEXT NOT NULL,   -- pass | fail_lowres | fail_blurry | fail_small_bytes | fail_corrupt | fail_wrong_aspect
    assessed_at    TIMESTAMP NOT NULL,
    cache_path     TEXT,
    attempt_count  INTEGER NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX idx_iqa_entity_variant ON image_quality_assessments (entity_type, entity_id, variant);
CREATE INDEX idx_iqa_source ON image_quality_assessments (source);
CREATE INDEX idx_iqa_verdict ON image_quality_assessments (verdict);
```

Both SQLite and PostgreSQL variants via the existing dual-migration structure.

- **Revalidation**: background goroutine registered in `internal/lifecycle/registry.go` walks 5% of cached covers every 7 days, re-`Assess()`es current bytes against current thresholds, and re-resolves anything that now fails (thresholds may be tightened over time; LLM provider quality may have improved).

## Error handling summary

| Case                                    | Behavior                                                   |
|-----------------------------------------|------------------------------------------------------------|
| Share bytes corrupt                     | `fail_corrupt`, skip to providers                          |
| Provider timeout / 4xx / 5xx            | `not_available`, continue chain                            |
| Provider returns candidate but gate fails | candidate discarded, continue chain                       |
| All providers + LLM miss                | placeholder + `X-Cover-Quality: placeholder_fallback` + WARN|
| LLM disabled via config                 | skip LLM step, fall directly to placeholder                |
| Decompression bomb                      | `fail_too_large`, reject without full decode               |
| SSRF-looking URL in provider response    | blocked by `ssrf_filter`-equivalent Go check (private IP ranges, file://, localhost) before download |
| User override present                   | bypass entire gate; serve user's URL as-is                 |

## Testing plan (100% across all 10 constitution categories)

| Category         | Coverage                                                                                              |
|------------------|-------------------------------------------------------------------------------------------------------|
| Unit             | `Media/pkg/quality/*_test.go` — 60+ table-driven cases (pass per hint, each fail verdict, boundaries, PNG/JPEG/WebP/GIF/BMP, corrupt, bomb, concurrent `Assess()`); catalog-api resolver unit tests with HTTP + LLM mocked |
| Integration      | `cover_pipeline_integration_test.go` — blurry share asset → stub provider returns high-res → cached → repeat request served from cache; failure cascade ends at placeholder with correct header |
| E2E              | Playwright web: 500-item fixture, zero blurry covers after full sync; Android TV instrumented test with video frame analysis |
| Full automation  | HelixQA orchestrator run covers the flow unattended                                                   |
| Stress           | `tests/k6/load_test_covers.js` — 100 concurrent cover requests, p95 < 200 ms cold, < 50 ms hot        |
| Security         | SSRF defense, decompression bomb, polyglot content-type, govulncheck                                  |
| DDoS / rate-limit| cover endpoint verified under 500 rps; single-flight stampede test                                    |
| Benchmarking     | `BenchmarkQualityAssess` < 5 ms/1MP; `BenchmarkCoverPipeline` end-to-end                              |
| Challenges       | `CH-IQ-001`..`CH-IQ-014` registered in `catalog-api/challenges/register.go`                           |
| HelixQA          | 12 new bank entries (Android TV / web / phone / desktop); fixes-validation regressions for each bug fixed |

### Challenge catalogue

| ID        | Purpose                                                                                  |
|-----------|------------------------------------------------------------------------------------------|
| CH-IQ-001 | Blurry share cover is blocked by the gate                                                |
| CH-IQ-002 | Low-resolution share cover is blocked                                                    |
| CH-IQ-003 | Per-media-type thresholds enforced (asserts across all 11 hints)                         |
| CH-IQ-004 | TMDB replacement fetched and cached                                                      |
| CH-IQ-005 | Fanart.tv replacement fetched when TMDB fails                                             |
| CH-IQ-006 | OMDB fallback path                                                                        |
| CH-IQ-007 | OpenLibrary large-size fetch for books                                                    |
| CH-IQ-008 | Cover Art Archive fetch for music                                                         |
| CH-IQ-009 | IGDB fetch for games                                                                      |
| CH-IQ-010 | LLM fallback only runs after every API returns not_available                              |
| CH-IQ-011 | Placeholder served when all resolvers fail; `X-Cover-Quality: placeholder_fallback`       |
| CH-IQ-012 | Cache hit on second request skips re-scoring (verified via DB attempt_count)              |
| CH-IQ-013 | Concurrent requests for same missing cover deduplicated (single-flight)                   |
| CH-IQ-014 | Background revalidation re-resolves tightened-threshold cover                             |

## Rollout sequence

1. Implement `Media/pkg/quality` + tests. Commit + push Media submodule (2 upstreams).
2. Migration v10 + tests. Commit catalog-api.
3. Implement resolvers (share, gate, fanart, coverartarchive, igdb, llm). Unit tests.
4. Wire into main.go, cover handler, asset handler. Emit `X-Cover-Quality` response header.
5. Background revalidation job.
6. Challenges + HelixQA banks. Commit + push Challenges submodule (2 upstreams).
7. Integration, security, k6, benchmark tests.
8. `./scripts/release-build.sh --container --force --skip-tests`.
9. `./scripts/services-up.sh`.
10. `./scripts/run-all-tests.sh` and `./scripts/helixqa-orchestrator.sh`.
11. Report to `docs/reports/qa-sessions/qa-session-2026-04-17/`.
12. Commit main-repo submodule pointer bumps. Push to all 7 remotes.

Every logical phase ends with a commit and push so we never accumulate more than ~60 minutes of unsaved work.

## Open points for the implementer

- `FANART_TV_API_KEY` and `IGDB_CLIENT_ID`/`IGDB_CLIENT_SECRET` are new optional env vars. `.env.example` must be updated with placeholder values, never real keys.
- `ssrf_filter`-equivalent for Go: use the `Security` submodule's existing network helpers if available; otherwise add a small `internal/netsafety` helper that rejects private ranges, loopback, and link-local before any download.
- LLM resolver must gate the returned URL through the same `netsafety` check.
- Gate verdict and provider source are exposed via the `X-Cover-Quality` and `X-Cover-Source` response headers to help the web and Android clients surface provenance in debug mode.
