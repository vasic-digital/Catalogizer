# ADR-009: Non-Blocking UI Patterns

## Status
Accepted (2026-03-20)

## Context

Catalogizer's web frontend (`catalog-web`) communicates with a backend API that performs inherently slow operations: scanning large media libraries across network storage (SMB, NFS, WebDAV), querying external metadata providers (TMDB, OMDB, MusicBrainz), downloading subtitles, and converting media formats. Response times for these operations range from hundreds of milliseconds to several minutes.

The frontend initially followed a straightforward request-response pattern: the user clicks a button, the UI shows a loading spinner, and the UI updates when the response arrives. This approach created several user experience problems:

1. **Perceived slowness**: Even when the backend responds in 500ms, the combination of loading spinner appearance, data fetch, and re-render makes the UI feel sluggish. Users performing rapid navigation (browsing entities, switching between collections) experience constant loading flicker.

2. **Large list performance**: Media libraries can contain tens of thousands of items. Rendering the full entity list in the DOM causes visible jank during scrolling, high memory usage, and slow initial render times (2-3 seconds for 10,000+ items).

3. **Route-level code splitting absence**: The single-page application bundle contained all route components eagerly, resulting in a 1.2MB initial JavaScript payload. Users visiting the login page downloaded code for the entity browser, admin panel, and conversion queue -- pages they might never visit.

4. **Image loading storms**: The entity browser and media grid load dozens of poster/thumbnail images simultaneously. On slow connections, this blocks the main thread with layout recalculations as images load in random order, causing content to jump and shift unpredictably.

5. **Stale UI after mutations**: After operations like "add to collection" or "mark as favorite", the UI waited for the server to confirm before updating. On slow connections, users would click, see no change for 1-2 seconds, and click again -- causing duplicate requests or confusion.

## Decision

We adopt four complementary non-blocking UI patterns in `catalog-web` to eliminate perceived latency and maintain responsiveness regardless of backend performance:

### 1. Optimistic UI Updates

For mutations where the expected outcome is predictable (favorites, collection membership, user preferences), the UI updates immediately before the server confirms. If the server rejects the mutation, the UI rolls back to the previous state and shows an error notification.

Implementation uses React Query's `onMutate` / `onError` / `onSettled` lifecycle:

```
User Action (click "Add to Favorites")
    |
    | onMutate: snapshot current state, apply optimistic update to query cache
    v
UI reflects change immediately (heart icon filled, count incremented)
    |
    | Background: POST /api/v1/favorites
    v
Server Response:
    Success (onSettled): invalidate query to sync with server state
    Failure (onError): restore snapshot, show error toast
```

Optimistic updates are applied to:
- Favorite toggling (`useFavorites` hook)
- Collection item add/remove (`useCollections` hook)
- Playlist reordering (`usePlaylists` hook)
- User preference changes (`usePreferences` hook)
- Entity rating/review submission

Operations that are NOT optimistically updated (unpredictable outcomes):
- Media scanning (results unknown until scan completes)
- Subtitle search (depends on external providers)
- Format conversion (long-running, asynchronous)
- User creation/deletion (admin actions require confirmation)

### 2. React.lazy Route Splitting

All route-level components are loaded via `React.lazy()` with `Suspense` boundaries. The router configuration wraps each route in a lazy import:

```
/login          -> lazy(() => import('./pages/Login'))
/browse         -> lazy(() => import('./pages/EntityBrowser'))
/entity/:id     -> lazy(() => import('./pages/EntityDetail'))
/collections    -> lazy(() => import('./pages/Collections'))
/admin/*        -> lazy(() => import('./pages/AdminPanel'))
/conversion     -> lazy(() => import('./pages/ConversionQueue'))
/subtitles      -> lazy(() => import('./pages/SubtitleManager'))
```

Combined with Vite's build-time chunk splitting (vendor, router, ui, charts, utils), the initial JavaScript payload is reduced from 1.2MB to approximately 180KB for the login page. Other routes are loaded on demand when the user navigates to them.

The `Suspense` fallback renders a minimal skeleton layout matching the target page structure, avoiding layout shift when the chunk arrives.

### 3. Virtual Scrolling for Large Lists

The entity browser and media grid use virtual scrolling (windowed rendering) for lists exceeding 100 items. Only the visible items plus a small overscan buffer are rendered in the DOM. As the user scrolls, off-screen items are unmounted and new items are mounted.

Key implementation details:
- Variable-height row support for entity cards with different content lengths
- Overscan of 5 items above and below the viewport to prevent flicker during fast scrolling
- Scroll position preservation when navigating away and returning (stored in Zustand)
- Keyboard navigation support (arrow keys, Page Up/Down, Home/End)
- Accessible `role="listbox"` with `aria-setsize` and `aria-posinset` for screen readers

Performance impact: rendering 50,000 entity items uses approximately 30 DOM nodes (the visible window) instead of 50,000, reducing memory from ~200MB to ~5MB and eliminating scroll jank.

### 4. IntersectionObserver for Image Loading

Media poster images and thumbnails use the `IntersectionObserver` API to load only when they enter (or are about to enter) the viewport. Each image component registers with a shared observer instance:

```
Image Component Mount
    |
    | observer.observe(imageElement)
    v
IntersectionObserver (rootMargin: "200px")
    |
    | Image enters 200px proximity of viewport
    v
Set image src from data-src (triggers load)
    |
    | Image loads: remove placeholder, fade in
    | Image fails: show fallback poster
    v
observer.unobserve(imageElement)
```

Implementation details:
- **200px root margin**: Images start loading 200px before they scroll into view, so they are typically loaded by the time the user sees them.
- **Shared observer instance**: A single `IntersectionObserver` handles all image elements, avoiding the overhead of one observer per image.
- **Low-quality placeholder**: While loading, images display a blurred, low-resolution placeholder (generated server-side as a 20px-wide thumbnail encoded as base64 in the API response).
- **Fade-in transition**: Loaded images fade in over 200ms via CSS transition on opacity, eliminating the jarring pop-in effect.
- **Error fallback**: Failed image loads display a generic media-type-specific placeholder (movie reel, music note, game controller, etc.).

## Consequences

### Positive

- **Better perceived performance**: Optimistic updates make mutation responses feel instantaneous. Users see immediate feedback for favorites, collections, and preferences without waiting for server round-trips.
- **Smaller initial bundle**: Route splitting reduces the initial JavaScript payload by approximately 85% (1.2MB to 180KB), significantly improving first-contentful-paint time on slow connections and mobile devices.
- **Smooth scrolling at any scale**: Virtual scrolling maintains 60fps scrolling performance whether the library contains 100 or 100,000 items. DOM node count stays constant regardless of list size.
- **Reduced bandwidth**: Lazy image loading avoids downloading poster images that the user never scrolls to. For a library of 10,000 items where the user views 50, this saves approximately 49,950 unnecessary image requests.
- **Graceful degradation**: If the backend is slow or temporarily unavailable, the UI remains interactive. Optimistic updates show the expected state, and errors are communicated via non-blocking toast notifications rather than blocking spinners.

### Negative

- **More complex state reconciliation**: Optimistic updates require snapshot/rollback logic in every mutation hook. If the server state diverges from the optimistic state (e.g., another user deleted an entity), reconciliation can show briefly incorrect data before the query invalidation corrects it.
- **Route loading delay on first navigation**: The first visit to a lazily-loaded route incurs a chunk download delay (typically 50-200ms on broadband, longer on slow connections). The skeleton fallback mitigates the visual impact, but the delay is perceptible on very slow networks.
- **Virtual scrolling accessibility complexity**: Screen readers must navigate a virtualized list where most items do not exist in the DOM. The ARIA attributes (`aria-setsize`, `aria-posinset`) communicate the full list size, but some assistive technologies may not handle this pattern well.
- **Image loading order unpredictability**: With IntersectionObserver, images load in scroll-order rather than document-order. Users who scroll quickly may see placeholders flash briefly before the images load, even with the 200px root margin buffer.
- **Testing overhead**: Optimistic updates, lazy routes, virtual scrolling, and intersection observers all require specialized test setups. Vitest tests must mock `IntersectionObserver`, simulate scroll events for virtual lists, and handle `React.lazy` suspense boundaries.
