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

## 2026-04-21T21:42 update — new evidence after KeyPress fix

After landing FIX-QA-2026-04-21-017 (KeyPress legacy-path fallback)
and force-stopping every non-Catalogizer foreground candidate, I
captured a fresh screenshot at 21:42:23. In the 2-second window
around that screenshot, the catalog-api server log shows the app
DID issue several image requests successfully:

    21:42:21.344  GET /api/v1/image-proxy?url=…tmdb…/yiBvCHwN…jpg  200  402 ms  okhttp/4.12.0
    21:42:21.344  GET /api/v1/image-proxy?url=…tmdb…/no5MpPR…jpg  200  406 ms  okhttp/4.12.0

Yet the rail cards still render as solid dark-teal with only the
play-arrow badge — **no artwork reaches the Compose surface** even
though OkHttp clearly received the JPEG bytes. This rules out
hypothesis #1 (empty items list) and hypothesis #3 (Retrofit converter
returning 0 items). The bug is purely client-side — between Coil
receiving the 200 response and the `SubcomposeAsyncImage` painting
the bitmap.

Updated ranked hypotheses:

1. **Coil's ImageLoader missing an OkHttp dispatcher** so the bytes
   are downloaded on a path that doesn't feed the compose painter.
2. **Image decoding fails silently** (ContentType mismatch? the API
   serves `image/jpeg`; the image-proxy may serve `application/
   octet-stream` or no Content-Type at all — worth checking).
3. **Card layout constraints zero the painter** so it draws into
   a 0×0 canvas.

Diagnostic commands for the APK-rebuild cycle:

    adb shell logcat -v time com.catalogizer.androidtv:V Coil:V ImageLoader:V

And in `catalog-api`: confirm the Content-Type header on
`/api/v1/image-proxy`:

    curl -sI -H "Authorization: Bearer $TOKEN" \
        "http://localhost:8080/api/v1/image-proxy?url=..." | grep -i content-type

## 2026-04-21T21:45 definitive root cause — dual-sided

### Server side

`GET /api/v1/image-proxy?url=https://image.tmdb.org/...` returns
`Content-Type: image/svg+xml` + a 1398-byte **placeholder SVG**, not
the TMDB JPEG:

    <svg xmlns="http://www.w3.org/2000/svg" width="300" height="450" viewBox="0 0 300 450">
      <defs>
        <linearGradient id="bg" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stop-color="#1a1a2e"/>
          <stop offset="60%" stop-color="#16213e"/>
          <stop offset="100%" stop-color="#0d0d0d"/>
          …

Traced to `catalog-api/main.go:1010-1022`:

    proxyClient := buildImageProxyClient(imageURL, cfg.Proxy)
    resp, err := proxyClient.Get(imageURL)
    if err != nil || resp.StatusCode != http.StatusOK {
        …
        svg := services.GeneratePlaceholderSVG(mediaType)
        c.Header("Content-Type", "image/svg+xml")
        c.Data(http.StatusOK, "image/svg+xml", svg)
        return
    }

So the proxy client is failing to fetch TMDB, and the fallback path
correctly serves a placeholder SVG. Evidence that TMDB is reachable
from the host:

    $ curl -s -o /dev/null -w "%{http_code} %{content_type}\n" \
        "https://image.tmdb.org/t/p/w500/mEWKXuCMv7mFMxXVSTI3v8UOQuq.jpg"
    200 image/jpeg

So the bug is in `buildImageProxyClient` — likely the dynamic-DNS
resolver (`resolveHostDynamic`) or SOCKS5/HTTP-proxy config, since
the handler routes through a potentially hijacked dial path.

### Client side

Even if the SVG is delivered legitimately, **Coil's default
`ImageLoader` cannot decode `image/svg+xml`** — the
`io.coil-kt:coil-svg` decoder is not in
`catalogizer-androidtv/app/build.gradle.kts`. So Coil gets 200 +
SVG bytes → decode fails silently → `SubcomposeAsyncImage` error
state → card paints the default teal background.

### Minimal fix — two small changes

1. **catalog-api** — buildImageProxyClient: log the Get error before
   falling back so operators see WHY TMDB is unreachable. Plus
   consider dropping the dynamic-DNS path when the default Go
   resolver would work (it does from curl); the hijacking workaround
   was likely for an older network scenario.
2. **catalogizer-androidtv** — add the Coil SVG decoder:

       implementation("io.coil-kt:coil-svg:2.5.0")

   And register `SvgDecoder.Factory()` in the `ImageLoader` builder.
   This alone makes the current placeholder SVG render (albeit as a
   placeholder), which is a strictly better UX than the blank teal
   card.

Both fixes are short. The catalog-api side is the real bug; the
client side is a defensive improvement. Covered by this ticket so
whichever arm lands first, the user stops seeing empty cards.
