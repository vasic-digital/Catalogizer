# Catalog Web -- React Frontend Course

**Component**: catalog-web
**Language**: TypeScript / React 18 / Vite
**Total Duration**: 5 hours (7 modules)
**Level**: Intermediate

---

## Course Overview

This course covers the complete architecture of catalog-web, the React frontend that provides the primary user interface for Catalogizer. You will learn the provider-based application shell, server and client state management, media browsing with the entity system, collection management, media playback, settings and admin panels, WebSocket real-time integration, and frontend performance optimization techniques.

---

### Module 1: Architecture

**Duration**: 45 minutes
**Prerequisites**: React 18 fundamentals, TypeScript, basic understanding of REST APIs

#### Learning Objectives
- Trace the application boot sequence through the provider hierarchy: AuthProvider, WebSocketProvider, Router
- Distinguish between server state (React Query) and client state (Zustand) and when to use each
- Navigate the codebase using path aliases (`@/components`, `@/hooks`, `@/lib`, `@/types`, `@/services`, `@/store`, `@/pages`, `@/assets`)
- Explain how the Vite dev server proxies API requests to catalog-api via `.service-port` auto-discovery

#### Topics Covered
1. **Application shell (`src/App.tsx`)**
   - Provider nesting order: `AuthProvider` -> `WebSocketProvider` -> `Router`
   - `AuthProvider` from `@vasic-digital/auth-context` managing JWT tokens, login/logout, and session state
   - `WebSocketProvider` from `@vasic-digital/websocket-client` establishing real-time connection after authentication
   - `ProtectedRoute` component gating authenticated-only pages
2. **State management strategy**
   - React Query (`@tanstack/react-query`) for all server state: media entities, collections, playlists, settings
   - Query invalidation patterns: mutation success triggers refetch of related queries
   - Zustand for client-only state: UI preferences, sidebar state, filter selections
   - React Hook Form + Zod for form state with schema-based validation
3. **Project structure and path aliases**
   - `src/pages/`: top-level route components (Dashboard, EntityBrowser, EntityDetail, Collections, MediaBrowser, Playlists, Favorites, Settings, Admin, Analytics, ConversionTools, SubtitleManager, AIDashboard)
   - `src/components/`: organized by domain (auth, entity, media, collections, playlists, favorites, admin, ai, dashboard, conversion, subtitles, upload, performance, layout, ui)
   - `src/hooks/`: custom hooks (useCollections, useDebounce, useFavorites, usePlaylists)
   - Path aliases defined in `vite.config.ts` for clean imports
4. **Vite configuration (`vite.config.ts`)**
   - API proxy reading `../catalog-api/.service-port` at dev-server startup, fallback to port 8080
   - Build chunks: `vendor` (react), `router`, `ui`, `charts`, `utils` for optimal caching
   - Dev server on port 3000 with proxy for `/api` routes to catalog-api
5. **Shared libraries (file-linked submodules)**
   - `@vasic-digital/ui-components`: reusable React UI library
   - `@vasic-digital/media-types`: shared media type definitions matching backend's 11 types
   - `@vasic-digital/catalogizer-api-client`: typed API client for all REST endpoints
   - `@vasic-digital/auth-context`: authentication context provider
   - `@vasic-digital/media-browser`, `@vasic-digital/media-player`, `@vasic-digital/collection-manager`, `@vasic-digital/dashboard-analytics`

#### Hands-On Exercise
Start the development server with `npm run dev` and open the browser DevTools. Inspect the React component tree to see the provider hierarchy. Open the Network tab and observe how API requests are proxied to catalog-api. Examine the React Query DevTools to see cached queries and their states. Modify a Zustand store value and observe the UI update.

#### Key Takeaways
- The provider hierarchy is strict: AuthProvider must wrap WebSocketProvider because real-time connections require authentication
- Server state belongs in React Query; client-only state belongs in Zustand -- mixing them causes stale data bugs
- The API proxy auto-discovers catalog-api's port via `.service-port`, eliminating hardcoded URLs
- Shared submodule libraries are linked via `file:../` in `package.json` for monorepo-style development

---

### Module 2: Media Browsing and Search

**Duration**: 45 minutes
**Prerequisites**: Module 1

#### Learning Objectives
- Build a media browsing interface using the entity system with hierarchical navigation
- Implement search with debounced input, type filters, and Cyrillic/special character support
- Create entity cards and detail views that adapt to the 11 media types
- Apply virtual scrolling for large result sets

#### Topics Covered
1. **Entity browser (`src/pages/EntityBrowser.tsx`)**
   - Grid layout with `EntityGrid` component rendering `EntityCard` for each media item
   - `TypeSelector` component filtering by media type (movie, tv_show, music_album, book, game, software, comic)
   - Hierarchical navigation: clicking a TV show opens its seasons, clicking a season opens its episodes
   - Breadcrumb trail tracking navigation depth through the hierarchy
2. **Entity detail view (`src/pages/EntityDetail.tsx`)**
   - `EntityDetailView` component adapting layout based on media type
   - Metadata display: title, year, genre, duration, external provider data (TMDB, OpenLibrary, MusicBrainz)
   - Related files list from `media_files` junction table
   - Actions: play, add to collection, add to favorites, add to playlist
3. **Search implementation**
   - `useDebounce` hook preventing excessive API calls during typing
   - Full-text search across entity titles with backend query optimization
   - Special character handling: Cyrillic titles, accented characters, CJK characters
   - `MediaFilters` component with multi-select for type, year range, and sort order
4. **Media grid (`src/components/media/`)**
   - `MediaGrid` component with responsive column count
   - `MediaCard` component showing cover art, title, year, type badge, and progress indicator
   - `ProgressBadge` component for partially-watched/listened items
   - `HistoryDrawer` for recently accessed items
5. **Advanced search (`src/components/collections/AdvancedSearch.tsx`)**
   - Compound filters: type AND year AND title pattern
   - Saved search queries for repeated use
   - Search results with relevance scoring

#### Hands-On Exercise
Navigate the entity browser through a TV show hierarchy: show -> season -> episode. Use the search bar to find a specific movie by title with a debounced query. Apply type and year filters simultaneously. Inspect the React Query cache to see how hierarchical queries are structured.

#### Key Takeaways
- The entity browser supports full hierarchical navigation matching the backend's parent-child media item structure
- Debounced search prevents API overload while maintaining responsive user experience
- Entity cards and detail views adapt their layout and available actions based on media type
- Search handles internationalized content (Cyrillic, CJK) through proper Unicode handling on both frontend and backend

---

### Module 3: Collection Management

**Duration**: 45 minutes
**Prerequisites**: Module 1, Module 2

#### Learning Objectives
- Implement CRUD operations for media collections with optimistic updates
- Build sharing workflows with permission levels
- Apply real-time sync for collaborative collection editing via WebSocket events
- Use the collection manager component library from `@vasic-digital/collection-manager`

#### Topics Covered
1. **Collection CRUD (`src/pages/Collections.tsx` and `src/components/collections/`)**
   - `CollectionsManager` component: list, create, edit, delete collections
   - `useCollections` hook wrapping React Query mutations with optimistic updates
   - Add/remove items from collections with drag-and-drop support
   - `CollectionPreview` component showing collection contents in a compact view
2. **Smart collections (`src/components/collections/SmartCollectionBuilder.tsx`)**
   - Rule-based automatic collections: "All movies from 2024", "All unplayed episodes"
   - Rule builder UI with condition type, operator, and value selectors
   - Dynamic membership: items automatically enter/leave based on rule evaluation
3. **Sharing and collaboration (`src/components/collections/CollectionSharing.tsx`)**
   - Share collections with other users via username or invite link
   - Permission levels: viewer (read-only), editor (add/remove items), owner (full control)
   - `CollectionSettings` component for managing shared access
4. **Real-time sync (`src/components/collections/CollectionRealTime.tsx`)**
   - WebSocket events for collection modifications by other users
   - Conflict resolution: last-write-wins with visual notification of concurrent edits
   - React Query cache invalidation triggered by WebSocket collection update events
5. **Bulk operations (`src/components/collections/BulkOperations.tsx`)**
   - Multi-select items across the entity browser for batch add to collection
   - Bulk remove, bulk move between collections
   - Progress indicator for large batch operations
6. **Export and templates**
   - `CollectionExport` component: export collection as JSON, CSV, or shareable link
   - `CollectionTemplates` component: pre-built collection templates (e.g., "Oscar Winners", "Top Albums")
   - `CollectionAnalytics` component: collection size, type distribution, growth over time

#### Hands-On Exercise
Create a new collection, add media items from the entity browser, and share it with another user account. Build a smart collection with rules that auto-populate based on media type and year. Open two browser windows and observe real-time sync as items are added from one window and appear in the other.

#### Key Takeaways
- Optimistic updates make collection mutations feel instant while the server confirms in the background
- Smart collections use rule-based membership that automatically updates as new media is scanned
- WebSocket events enable real-time collaborative editing with React Query cache invalidation
- The collection system supports the full lifecycle: creation, curation, sharing, export, and analytics

---

### Module 4: Media Playback

**Duration**: 40 minutes
**Prerequisites**: Module 1, Module 2

#### Learning Objectives
- Integrate the media player component for streaming playback of video and audio files
- Build playlist navigation with sequential and shuffle playback modes
- Handle streaming protocols and format negotiation with the backend
- Display playback progress and synchronize it with the backend for cross-device resume

#### Topics Covered
1. **Media player (`src/components/media/MediaPlayer.tsx`)**
   - Unified player component handling video and audio files
   - Integration with `@vasic-digital/media-player` submodule
   - Transport controls: play/pause, seek, volume, fullscreen
   - Adaptive quality selection based on network conditions
2. **Streaming integration**
   - Backend streaming endpoint delivering media content via chunked transfer
   - Range request support for seeking within large files
   - Format detection from file metadata and MIME type headers
3. **Playlists (`src/pages/Playlists.tsx`)**
   - `usePlaylists` hook managing playlist state and mutations
   - Sequential and shuffle playback modes
   - Auto-advance to next item with configurable behavior
   - Drag-and-drop reordering of playlist items
4. **Playback progress tracking**
   - `ProgressBadge` component showing completion percentage on media cards
   - Periodic progress reports sent to the backend during playback
   - Cross-device resume: start on web, continue on mobile or TV
   - `HistoryDrawer` showing recently played items with timestamps
5. **Media detail modal (`src/components/media/MediaDetailModal.tsx`)**
   - Inline player embedded in entity detail view
   - Episode navigation for TV shows: previous/next episode buttons
   - Album track list with current track highlighting for music

#### Hands-On Exercise
Play a video file from the entity browser and observe the streaming request in the Network tab. Create a playlist with multiple items and test sequential playback with auto-advance. Pause playback, navigate away, and return to verify resume position. Check the backend API to see the stored playback progress.

#### Key Takeaways
- The media player handles both video and audio through a unified component with adaptive transport controls
- Streaming uses chunked transfer with range request support for efficient seeking
- Playback progress is persisted to the backend, enabling cross-device resume
- Playlists support drag-and-drop reordering with sequential and shuffle modes

---

### Module 5: Settings and Configuration

**Duration**: 35 minutes
**Prerequisites**: Module 1

#### Learning Objectives
- Navigate the settings page structure: user preferences, server connection, and admin panel
- Implement storage root management for adding and removing media sources
- Build the admin panel with user management, backup controls, and system monitoring
- Apply configuration precedence: server-side defaults, user preferences, local storage overrides

#### Topics Covered
1. **Settings page (`src/pages/Settings.tsx`)**
   - User profile: display name, avatar, password change
   - Theme selection and UI density preferences
   - Notification preferences for scan completion, new media detection
   - Storage root management: add/edit/remove local and network media sources
2. **Admin panel (`src/pages/Admin.tsx`)**
   - User management: create, edit, deactivate user accounts
   - Role assignment: admin, user, viewer
   - System backup and restore controls
   - Server configuration: scan intervals, metadata provider API keys, cache settings
3. **Server connection management**
   - API base URL configuration for multi-server environments
   - Connection health indicator showing backend status
   - Auto-reconnection with exponential backoff on connection loss
4. **Analytics dashboard (`src/pages/Analytics.tsx`)**
   - Media library statistics: total items by type, storage usage, scan history
   - `@vasic-digital/dashboard-analytics` submodule providing chart components
   - User activity tracking: most played, recently added, favorites trends

#### Hands-On Exercise
Add a new storage root via the Settings page and observe the scan trigger. Create a new user account in the admin panel and assign the viewer role. Log in as the viewer and verify restricted access to admin endpoints. Explore the analytics dashboard to see library statistics.

#### Key Takeaways
- Settings are layered: server defaults, then user preferences, then local storage for UI-only options
- Admin operations (user management, backup, system config) are role-gated on both frontend and backend
- Storage root management integrates directly with the scan pipeline for immediate media discovery
- Analytics provide visibility into library composition, growth, and user engagement

---

### Module 6: WebSocket Integration

**Duration**: 35 minutes
**Prerequisites**: Module 1, understanding of WebSocket protocol

#### Learning Objectives
- Configure the WebSocket provider with authentication-aware connection management
- Handle real-time events for scan progress, media updates, and collection changes
- Implement automatic reconnection with exponential backoff
- Write tests for WebSocket-dependent components

#### Topics Covered
1. **WebSocket provider setup**
   - `@vasic-digital/websocket-client` submodule providing React hooks and context
   - Connection established after successful authentication (JWT token passed in handshake)
   - Provider placed below `AuthProvider` in the component tree to guarantee token availability
2. **Event handling**
   - Scan progress events: percentage, current file, estimated time remaining
   - Media detection events: new entity created, metadata enriched
   - Collection update events: item added/removed by another user
   - Error events: scan failure, connection issue, filesystem unavailable
3. **Reconnection strategy**
   - Automatic reconnection on connection drop with exponential backoff
   - Maximum retry attempts before showing user-visible connection error
   - Seamless state sync on reconnect: React Query refetch of stale data
4. **UI integration patterns**
   - Toast notifications for background events (scan complete, new media found)
   - Live progress bars during active scans
   - Real-time entity count updates in the sidebar
   - Connection status indicator in the application header
5. **Testing WebSocket components**
   - Mock WebSocket server for unit tests
   - Simulating connection drops and verifying reconnection behavior
   - Testing event-driven React Query cache invalidation

#### Hands-On Exercise
Open the browser console and observe WebSocket messages during a scan operation. Disconnect the network momentarily and watch the reconnection sequence. Write a test component that subscribes to a WebSocket event and verifies the UI updates when the event fires.

#### Key Takeaways
- WebSocket connection depends on authentication; the provider hierarchy enforces this ordering
- Reconnection is automatic with exponential backoff, and stale React Query data is refetched on reconnect
- Real-time events update the UI without polling, providing immediate feedback for scans and collaborative edits
- Toast notifications surface background events without interrupting the user's current workflow

---

### Module 7: Performance

**Duration**: 35 minutes
**Prerequisites**: Modules 1-6

#### Learning Objectives
- Apply React.lazy and Suspense for route-level code splitting
- Implement memoization strategies with `useMemo`, `useCallback`, and `React.memo`
- Use optimistic updates in React Query mutations for instant UI feedback
- Optimize bundle size with Vite's manual chunk configuration

#### Topics Covered
1. **Code splitting**
   - `React.lazy` for route-level splitting: each page loads only when navigated to
   - `Suspense` boundaries with `SplashScreen` fallback component
   - Dynamic imports for heavy components (charts, media player, subtitle editor)
2. **Memoization**
   - `React.memo` on `EntityCard`, `MediaCard`, and other frequently re-rendered list items
   - `useMemo` for computed values: filtered entity lists, search results, collection statistics
   - `useCallback` for event handlers passed to child components
   - `useDebounce` hook preventing excessive re-renders from rapid input changes
3. **Optimistic updates**
   - React Query `onMutate` callback for immediate UI update before server confirmation
   - Rollback on error: `onError` callback restoring previous cache state
   - Patterns for collection add/remove, playlist reorder, favorite toggle
4. **Bundle optimization (`vite.config.ts`)**
   - Manual chunks: `vendor` (react, react-dom), `router` (react-router), `ui` (component library), `charts` (recharts), `utils` (shared utilities)
   - Tree shaking for unused exports from submodule libraries
   - CSS purging via Tailwind's content configuration
5. **Performance monitoring (`src/components/performance/`)**
   - `PerformanceOptimizer` component for runtime monitoring
   - React Profiler integration for identifying slow renders
   - Network request timing correlation with UI responsiveness
6. **Error boundaries**
   - `ErrorBoundary` component catching render errors with fallback UI
   - `PageErrorBoundary` component for page-level error isolation
   - Graceful degradation: failed components show error state without crashing the application

#### Hands-On Exercise
Use the browser's Performance tab to profile a page navigation and identify code-split chunk loading. Add `React.memo` to a list item component and measure the render count reduction. Implement an optimistic update for a collection mutation and test the rollback behavior by simulating a server error. Compare bundle sizes before and after chunk optimization.

#### Key Takeaways
- Route-level code splitting ensures users download only the code they need for the current page
- Optimistic updates make mutations feel instant; rollback on server error preserves data integrity
- Memoization must be applied strategically: over-memoizing adds complexity without measurable benefit
- Vite's manual chunk configuration aligns with browser caching: vendor chunks change rarely and cache long
