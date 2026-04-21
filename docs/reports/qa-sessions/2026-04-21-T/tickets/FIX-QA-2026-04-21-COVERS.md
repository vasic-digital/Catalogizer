# FIX-QA-2026-04-21-COVERS — Android TV home rail covers never fetch

**Severity:** HIGH (user-visible every launch)
**Discovered:** 2026-04-21 during X-cycle QA session on MIBOX4 (Android 9)
**Status:** OPEN — diagnosed, root cause needs APK rebuild with instrumentation

## Symptom

Home screen on `com.catalogizer.androidtv` (2.3.0 / code 7) loads stats
correctly (189 items / 174 Movies / 10 Comics / 2 Software / 2 TV Shows
/ 1 Books) but every rail card — "Recently Added Movies",
"Recently Added TV Shows", … — renders the teal placeholder background
with only a play-arrow icon. No cover artwork ever appears.

Evidence:

- `catalog-api` server log: **0 `GET /api/v1/cover/*`** and **0
  `GET /api/v1/image-proxy*`** in the 40-second window after a clean
  `am force-stop` + `am start` + 30 s hydration wait.
- Retrofit `/api/v1/entities/browse/movie?sort_by=created` returns
  HTTP 200 with `cover_url` populated (e.g., `/api/v1/cover/1`,
  `/api/v1/image-proxy?url=https://image.tmdb.org/...`).
- The app's okhttp interceptor logs confirm the JSON hits the
  client; Kotlinx should auto-map `cover_url` → `MediaItem.coverUrl`
  via `@SerialName("cover_url")`.

## Root-cause hypotheses (ranked)

1. **MediaItem deserialisation silently fails and the list ends up
   empty.** The rail's ~6 placeholder boxes come from a shimmer
   component, not from cards bound to real items. Coil never fires
   because MediaCard never renders. Evidence: "4× played" overlay on
   the leftmost card is suspicious — it may be a "continue watching"
   card that IS bound, while the others are shimmers.
2. **`thumbnailUrl` returns null** because `externalMetadata` is not
   null but its first element's `coverUrl`/`posterUrl` resolve to
   null, short-circuiting the chain before `coverUrl` fallback.
3. Retrofit/Kotlinx converter mis-registers the `MediaSearchResponse`
   wrapper and `items` ends up empty even with 200 OK.

## Next diagnostic step

Rebuild `catalogizer-androidtv` with a single `Log.d("MediaRail", "items=${items.size} firstCover=${items.firstOrNull()?.thumbnailUrl}")`
hook in `HomeViewModel.loadRecentByType` + `MediaCard.kt` (log the
`rawUrl`/`coverUrl` values). Install on MIBOX4, force stop + start,
grep logcat for `MediaRail`.

## Not yet in scope

- DeSerialiser-level tests for `MediaItem` against real API JSON
  payloads — recommended follow-up.
- A HelixQA bank entry `android-tv-home-covers-visible` that asserts
  Coil made at least N `GET /api/v1/cover/*` requests within N s of
  home-screen being focused.
