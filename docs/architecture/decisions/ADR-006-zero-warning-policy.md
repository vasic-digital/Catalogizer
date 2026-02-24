# ADR-006: Zero Warning / Zero Error Policy

## Status
Accepted (2026-02-23)

## Context

Media management applications are used by non-technical users who expect a polished, reliable experience. Browser console errors, failed network requests, and deprecation warnings are symptoms of integration defects that erode user trust and make debugging legitimate issues harder because signal is lost in noise.

During early development of Catalogizer, several categories of issues were observed:

1. **Missing API endpoints**: The frontend called endpoints that had not been implemented yet, producing 404 errors in the browser console and broken UI states.
2. **Schema mismatches**: API responses did not match the TypeScript types the frontend expected, causing runtime errors in React components.
3. **WebSocket connection failures**: The WebSocket client attempted connections before the server was ready, producing console errors.
4. **Framework deprecation warnings**: Outdated usage patterns of React, React Router, and Tailwind CSS produced console warnings that obscured real errors.
5. **Failed asset requests**: Cover art and thumbnail requests for non-existent assets returned 404s, producing console errors and broken image placeholders.

Each of these issues individually seems minor, but collectively they create a noisy, unreliable user experience and make it impossible to quickly identify genuine defects.

## Decision

All Catalogizer components must run with zero console warnings, zero console errors, and zero failed network requests in every environment (development, testing, production). This policy is enforced through code review, the challenge framework, and stub endpoints.

### Rules

1. **No browser console errors or warnings.** Every console error is treated as a defect to be fixed, not tolerated.

2. **Every failed network request is a defect.** If the frontend calls an API endpoint, that endpoint must exist, return a valid HTTP 2xx response, and match the expected response shape. A 404 from a missing endpoint is a backend defect. A failed fetch due to a network error is an infrastructure defect.

3. **No framework deprecation warnings.** All dependencies must be used according to their current API. Deprecated patterns must be migrated before they produce warnings.

4. **No WebSocket connection failures.** The WebSocket client must handle connection timing gracefully, with silent reconnection and no console errors during the connection lifecycle.

5. **If a feature is not yet implemented, provide a stub endpoint that returns a valid empty response.** For example, a not-yet-implemented search endpoint should return `{ "items": [], "total": 0 }` rather than a 404 or 500. The stub must match the expected response schema.

6. **Asset requests for non-existent assets return a default placeholder** (via the asset management system's `defaults.NewEmbeddedProvider()`), not a 404. The `X-Asset-Status: pending` header signals to the client that the real asset is being resolved in the background.

### Enforcement

The challenge framework (ADR-004) enforces this policy through several specific challenges:

- **CH-005 (browsing-web-app)**: Loads the web frontend and verifies zero console errors, zero failed network requests, and all expected API responses.
- **CH-004 (browsing-api-catalog)**: Verifies all catalog browsing API endpoints return valid responses.
- **CH-006, CH-007 (asset challenges)**: Verify that asset serving returns valid content (default or resolved) rather than 404s.
- **CH-008 (auth-token-refresh)**: Verifies the auth flow completes without errors.
- **CH-016 through CH-020 (entity challenges)**: Verify the entire entity API surface returns valid responses after aggregation.

Running all 20 challenges (117 assertions total) constitutes a comprehensive zero-warning validation. A challenge failure due to a console error or failed request is treated with the same severity as a functional failure.

### Implementation Patterns

**Backend stub endpoints** follow this pattern:
```go
// Stub for not-yet-implemented feature
api.GET("/some/future/endpoint", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "items": []interface{}{},
        "total": 0,
    })
})
```

**Frontend error boundaries** catch rendering errors without producing console errors:
```tsx
<ErrorBoundary fallback={<ErrorFallbackUI />}>
  <Component />
</ErrorBoundary>
```

**API client interceptors** handle non-2xx responses gracefully:
```typescript
// Returns empty result instead of throwing, preventing console errors
if (!response.ok) {
  return { items: [], total: 0 };
}
```

**Asset serving** always returns content:
```go
// If real asset not found, return type-appropriate default
rc, si, isDefault, err := h.manager.GetTyped(ctx, id, assetType)
// isDefault=true means placeholder was returned, not a 404
```

## Consequences

### Positive

- **Clean signal**: When a console error appears, it is a genuine defect that needs attention, not background noise to be ignored.
- **User trust**: A polished, error-free UI builds confidence in the software's reliability.
- **Faster debugging**: Without noise from known-benign errors, real issues are identified immediately.
- **API contract enforcement**: Stub endpoints document the expected API surface and prevent frontend-backend integration gaps.
- **Automated verification**: The challenge suite catches zero-warning violations before they reach users.

### Negative

- **Stub maintenance**: Stub endpoints must be maintained until the real implementation replaces them. Stale stubs that return incorrect shapes become their own category of defect.
- **Higher implementation cost**: Every API endpoint must be implemented (or stubbed) before the corresponding frontend feature can be developed, preventing frontend-first development workflows.
- **False sense of completeness**: Stub endpoints returning empty data may make features appear "working" when they are not yet implemented, potentially confusing testers.
- **Strictness overhead**: Developers must fix deprecation warnings and minor console messages that would normally be deferred, adding friction to development velocity.
