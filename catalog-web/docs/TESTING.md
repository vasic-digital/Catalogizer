# catalog-web Testing

## Unit Tests (Vitest)

### Setup

Vitest is configured in `vite.config.ts` with jsdom environment and global test APIs. The setup file at `src/test-setup.ts` provides mocks for browser APIs not available in jsdom:

- `window.matchMedia` -- media query mock
- `IntersectionObserver` -- lazy-loading support
- `ResizeObserver` -- layout observer
- `WebSocket` -- full lifecycle mock (CONNECTING -> OPEN -> CLOSED)
- `localStorage` / `sessionStorage` -- Storage interface mock
- `fetch` -- global fetch stub
- `crypto.randomUUID` -- deterministic UUIDs
- `HTMLMediaElement` (play/pause/load) -- media playback
- `HTMLCanvasElement.getContext` -- canvas 2D context

### Conventions

- Test files use `.test.ts` or `.test.tsx` extensions.
- Tests are colocated with source in `__tests__/` directories within each module (e.g., `src/components/__tests__/`, `src/hooks/__tests__/`).
- Use `@testing-library/react` for rendering and `@testing-library/user-event` for interaction simulation.
- Vitest globals (`describe`, `it`, `expect`, `vi`) are available without imports.

### Running

```bash
npm run test              # single run
npm run test:watch        # watch mode
npm run test:coverage     # with @vitest/coverage-v8
```

### Mocking Patterns

- **API calls**: Mock axios or the specific `lib/` API module with `vi.mock()`.
- **React Query**: Wrap test components in a `QueryClientProvider` with a fresh `QueryClient`.
- **Auth context**: Wrap in `AuthProvider` or mock the `useAuth` hook.
- **Router**: Wrap in `MemoryRouter` for components that use `useNavigate` or `useParams`.

## E2E Tests (Playwright)

### Setup

Playwright is configured in `playwright.config.ts`. Tests live in the `e2e/` directory (excluded from Vitest via config). Playwright spec files use `.spec.ts` extension.

### Browser Matrix

| Project | Device |
|---------|--------|
| chromium | Desktop Chrome |
| firefox | Desktop Firefox |
| webkit | Desktop Safari |
| Mobile Chrome | Pixel 5 viewport |

All browser projects depend on a `setup` project that handles authentication state.

### Configuration

- **Base URL**: `http://localhost:3000` (overridable via `PLAYWRIGHT_BASE_URL`)
- **Web server**: Playwright auto-starts `npm run dev` and waits for the dev server
- **Test timeout**: 30 seconds per test, 5 seconds per assertion
- **Artifacts**: Screenshots on failure, video and trace on first retry
- **Reports**: HTML report in `playwright-report/`

### Test Suites

| File | Coverage |
|------|----------|
| `auth.spec.ts` | Login, registration, auth flows |
| `browse.spec.ts` | Media and entity browsing |
| `collections.spec.ts` | Collection CRUD |
| `favorites.spec.ts` | Favorites management |
| `search.spec.ts` | Search functionality |
| `accessibility.spec.ts` | Accessibility checks (axe-core) |
| `responsive.spec.ts` | Responsive layout validation |

### Running

```bash
npm run test:e2e           # headless, all browsers
npm run test:e2e:headed    # visible browser
npm run test:e2e:ui        # Playwright UI mode
npm run test:e2e:chromium  # Chromium only
npm run test:e2e:debug     # debug mode with inspector
```

### Global Setup

`e2e/global-setup.ts` handles pre-test setup (e.g., seeding test data). The `e2e/fixtures/` directory contains shared test fixtures and page objects.
