# React Frontend Architecture Guide (catalog-web)

This guide documents the architecture, component hierarchy, state management, and conventions used in the `catalog-web` React/TypeScript frontend.

## Technology Stack

- **React 18** with TypeScript
- **React Router v6** for routing
- **TanStack React Query** for server state management
- **Axios** for HTTP requests
- **Tailwind CSS** for styling
- **Vite** for build tooling
- **Jest + React Testing Library** for testing
- **jest-axe** for accessibility testing

## Component Hierarchy

```
App
├── ErrorBoundary
│   └── AuthProvider (context)
│       └── WebSocketProvider (context)
│           └── Router (BrowserRouter)
│               ├── ConnectionStatus (global WS status indicator)
│               └── Suspense (lazy loading boundary)
│                   └── Routes
│                       ├── /login           → LoginForm (public)
│                       ├── /register        → RegisterForm (public)
│                       └── Layout (authenticated shell with header/nav)
│                           ├── /dashboard   → ProtectedRoute → Dashboard
│                           ├── /media       → ProtectedRoute(read:media) → MediaBrowser
│                           ├── /analytics   → ProtectedRoute(view:analysis) → Analytics
│                           ├── /subtitles   → ProtectedRoute(manage:subtitles) → SubtitleManager
│                           ├── /collections → ProtectedRoute(read:collections) → Collections
│                           ├── /favorites   → ProtectedRoute → FavoritesPage
│                           ├── /playlists   → ProtectedRoute → PlaylistsPage
│                           ├── /conversion  → ProtectedRoute(convert:media) → ConversionTools
│                           ├── /admin       → ProtectedRoute(requireAdmin) → Admin
│                           └── /ai          → ProtectedRoute → AIDashboard
```

### Key architectural decisions

1. **Code splitting**: All page components are lazy-loaded with `React.lazy()` for smaller initial bundle
2. **Error boundary**: Wraps the entire app to catch rendering errors
3. **Provider nesting order**: ErrorBoundary > AuthProvider > WebSocketProvider > Router
4. **Layout component**: Renders the authenticated shell (Header, navigation) with `<Outlet>` for child routes

## Directory Structure

```
catalog-web/src/
├── App.tsx                  # Root component with provider wiring and route definitions
├── main.tsx                 # Entry point (renders App with QueryClientProvider)
├── contexts/
│   ├── AuthContext.tsx       # Authentication state and operations
│   └── WebSocketContext.tsx  # Real-time WebSocket connection management
├── lib/
│   ├── api.ts               # Axios instance + authApi (login, register, etc.)
│   ├── mediaApi.ts          # Media CRUD + search API calls
│   ├── subtitleApi.ts       # Subtitle search/download/sync API
│   ├── conversionApi.ts     # Media conversion job API
│   ├── collectionsApi.ts    # Collections CRUD API
│   ├── favoritesApi.ts      # Favorites API
│   ├── playlistsApi.ts      # Playlists API
│   ├── adminApi.ts          # Admin-only API calls
│   ├── websocket.ts         # WebSocketClient class + useWebSocket hook
│   ├── utils.ts             # Shared utility functions
│   └── webVitals.ts         # Performance monitoring
├── types/
│   ├── auth.ts              # User, LoginRequest, AuthStatus, etc.
│   ├── media.ts             # MediaItem, MediaSearchRequest, etc.
│   ├── subtitles.ts         # Subtitle types
│   ├── conversion.ts        # Conversion types
│   ├── collections.ts       # Collection types
│   ├── favorites.ts         # Favorites types
│   ├── playlists.ts         # Playlist types
│   ├── dashboard.ts         # Dashboard stat types
│   └── admin.ts             # Admin panel types
├── hooks/
│   └── useCollections.ts    # Custom hook for collection operations
├── components/
│   ├── ErrorBoundary.tsx    # Global error boundary
│   ├── auth/
│   │   ├── LoginForm.tsx    # Login page component
│   │   ├── RegisterForm.tsx # Registration page component
│   │   └── ProtectedRoute.tsx  # Auth-gated route wrapper
│   ├── layout/
│   │   ├── Layout.tsx       # Authenticated app shell
│   │   ├── Header.tsx       # Top navigation bar
│   │   └── PageHeader.tsx   # Page-level header
│   ├── media/
│   │   ├── MediaGrid.tsx    # Grid display of media items
│   │   ├── MediaCard.tsx    # Individual media card
│   │   ├── MediaDetailModal.tsx
│   │   ├── MediaFilters.tsx # Search/filter controls
│   │   └── MediaPlayer.tsx  # Media playback
│   ├── dashboard/
│   │   ├── DashboardStats.tsx
│   │   ├── ActivityFeed.tsx
│   │   └── MediaDistributionChart.tsx
│   ├── collections/         # Collection management components
│   ├── playlists/           # Playlist components
│   ├── subtitles/           # Subtitle management components
│   ├── favorites/           # Favorites components
│   ├── conversion/          # Format conversion components
│   ├── ai/                  # AI-powered features
│   ├── admin/               # Admin panel components
│   ├── upload/              # File upload components
│   ├── performance/         # Performance optimization components
│   └── ui/                  # Reusable UI primitives
│       ├── Button.tsx
│       ├── Card.tsx
│       ├── Input.tsx
│       ├── Badge.tsx
│       ├── Select.tsx
│       ├── Switch.tsx
│       ├── Tabs.tsx
│       ├── Textarea.tsx
│       ├── Progress.tsx
│       └── ConnectionStatus.tsx
├── pages/
│   ├── Dashboard.tsx
│   ├── MediaBrowser.tsx
│   ├── Analytics.tsx
│   ├── SubtitleManager.tsx
│   ├── Collections.tsx
│   ├── Favorites.tsx
│   ├── Playlists.tsx
│   ├── ConversionTools.tsx
│   ├── Admin.tsx
│   └── AIDashboard.tsx
└── test/
    └── setup.ts             # Jest test setup
```

## State Management

### Server State: React Query

All server data is managed via TanStack React Query. This handles caching, background refetching, and cache invalidation.

```tsx
// From contexts/AuthContext.tsx
const { data: authStatus, isLoading } = useQuery({
  queryKey: ['auth-status'],
  queryFn: authApi.getAuthStatus,
  retry: (failureCount, error: any) => {
    if (error?.response?.status === 401) return false
    return failureCount < 2
  },
  staleTime: 1000 * 60 * 5,  // 5 minutes
})
```

Mutations follow the same pattern:

```tsx
const loginMutation = useMutation({
  mutationFn: authApi.login,
  onSuccess: (data) => {
    localStorage.setItem('auth_token', data.token)
    setUser(data.user)
    queryClient.invalidateQueries({ queryKey: ['auth-status'] })
  },
  onError: (error: any) => {
    toast.error(error?.response?.data?.error || 'Login failed')
  },
})
```

### Client State: React Context

Two global contexts provide app-wide state:

1. **AuthContext** - user, permissions, login/logout/register actions
2. **WebSocketContext** - real-time connection management

There is no separate state library (like Zustand or Redux). Component-local state uses `useState` and `useReducer`.

## Auth Flow

### AuthProvider (`contexts/AuthContext.tsx`)

The `AuthProvider` wraps the entire app and manages:

- **Current user** state via `useState<User | null>`
- **Permissions** array via `useState<string[]>`
- **Authentication check** via `useQuery(['auth-status'])` on mount
- **Login** via `useMutation` that stores token in `localStorage`
- **Logout** via `useMutation` that clears token and query cache

```tsx
interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  permissions: string[]
  isAdmin: boolean
  login: (data: LoginRequest) => Promise<any>
  register: (data: RegisterRequest) => Promise<any>
  logout: () => Promise<void>
  hasPermission: (permission: string) => boolean
  canAccess: (resource: string, action: string) => boolean
}
```

### ProtectedRoute (`components/auth/ProtectedRoute.tsx`)

Wraps page components to enforce authentication and authorization:

```tsx
interface ProtectedRouteProps {
  children: React.ReactNode
  requireAdmin?: boolean
  requiredPermission?: string
  requiredRole?: string
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children, requireAdmin, requiredPermission, requiredRole,
}) => {
  const { isAuthenticated, isLoading, user, hasPermission } = useAuth()

  if (isLoading) return <Spinner />
  if (!isAuthenticated) return <Navigate to="/login" />
  if (requireAdmin && user?.role !== 'admin') return <Navigate to="/dashboard" />
  if (requiredPermission && !hasPermission(requiredPermission)) return <Navigate to="/dashboard" />

  return <>{children}</>
}
```

Usage in routes:

```tsx
<Route path="/media" element={
  <ProtectedRoute requiredPermission="read:media">
    <MediaBrowser />
  </ProtectedRoute>
} />
```

### Token Management

- Stored in `localStorage` under `auth_token`
- Attached to every request via Axios interceptor:
  ```ts
  api.interceptors.request.use((config) => {
    const token = localStorage.getItem('auth_token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  })
  ```
- 401 responses trigger automatic logout:
  ```ts
  api.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.response?.status === 401) {
        localStorage.removeItem('auth_token')
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }
  )
  ```

## WebSocket Integration

### WebSocketClient (`lib/websocket.ts`)

A class that manages a persistent WebSocket connection with:
- Automatic reconnection with exponential backoff (max 5 attempts)
- Message queuing when disconnected
- Channel-based subscribe/unsubscribe
- Connection state tracking

### useWebSocket hook

```ts
const { connect, disconnect, send, subscribe, unsubscribe, getConnectionState } = useWebSocket()
```

On connection, auto-subscribes to:
- `media_updates` - new/updated/deleted media
- `system_updates` - service health changes
- `analysis_updates` - media analysis completion

### WebSocketProvider (`contexts/WebSocketContext.tsx`)

Manages the WebSocket lifecycle tied to authentication state:

```tsx
useEffect(() => {
  if (isAuthenticated && user) {
    webSocket.connect()
    return () => webSocket.disconnect()
  } else {
    webSocket.disconnect()
  }
}, [isAuthenticated, user])
```

### Message handling

WebSocket messages trigger React Query cache invalidation:

```ts
case 'media_update':
  queryClient.invalidateQueries({ queryKey: ['media-search'] })
  queryClient.invalidateQueries({ queryKey: ['media-stats'] })
  break
case 'analysis_complete':
  queryClient.invalidateQueries({ queryKey: ['media-search'] })
  break
```

## API Layer

### Base configuration (`lib/api.ts`)

```ts
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})
```

### API modules

Each domain has its own API module in `lib/`:

```ts
// lib/mediaApi.ts
export const mediaApi = {
  searchMedia: (params: MediaSearchRequest): Promise<MediaSearchResponse> =>
    api.get('/media/search', { params }).then((res) => res.data),

  getMediaById: (id: number): Promise<MediaItem> =>
    api.get(`/media/${id}`).then((res) => res.data),

  deleteMedia: (id: number): Promise<void> =>
    api.delete(`/media/${id}`).then(() => {}),
}
```

## How to Add a New Page

### Step 1: Define types (`types/myfeature.ts`)

```ts
export interface MyFeatureItem {
  id: number
  name: string
  created_at: string
}
```

### Step 2: Create API module (`lib/myfeatureApi.ts`)

```ts
import api from './api'
import type { MyFeatureItem } from '@/types/myfeature'

export const myfeatureApi = {
  getAll: (): Promise<MyFeatureItem[]> =>
    api.get('/myfeature').then((res) => res.data),

  create: (data: Partial<MyFeatureItem>): Promise<MyFeatureItem> =>
    api.post('/myfeature', data).then((res) => res.data),
}
```

### Step 3: Create page component (`pages/MyFeature.tsx`)

```tsx
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { myfeatureApi } from '@/lib/myfeatureApi'

export const MyFeature: React.FC = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['myfeature'],
    queryFn: myfeatureApi.getAll,
  })

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Error loading data</div>

  return (
    <div>
      {data?.map(item => <div key={item.id}>{item.name}</div>)}
    </div>
  )
}
```

### Step 4: Add route in `App.tsx`

```tsx
const MyFeature = React.lazy(() =>
  import('@/pages/MyFeature').then(m => ({ default: m.MyFeature }))
)

// Inside Routes, within the Layout route:
<Route path="/myfeature" element={
  <ProtectedRoute requiredPermission="read:myfeature">
    <MyFeature />
  </ProtectedRoute>
} />
```

### Step 5: Add navigation link in `Header.tsx`

Add an entry to the navigation items array in `components/layout/Header.tsx`.

## How to Add a New API Call

1. Add the function to the appropriate API module in `lib/`:
   ```ts
   export const mediaApi = {
     // existing calls...
     getRelated: (id: number): Promise<MediaItem[]> =>
       api.get(`/media/${id}/related`).then((res) => res.data),
   }
   ```

2. Use it in a component with React Query:
   ```tsx
   const { data: related } = useQuery({
     queryKey: ['media-related', mediaId],
     queryFn: () => mediaApi.getRelated(mediaId),
     enabled: !!mediaId,
   })
   ```

3. For mutations (POST/PUT/DELETE):
   ```tsx
   const queryClient = useQueryClient()
   const deleteMutation = useMutation({
     mutationFn: (id: number) => mediaApi.deleteMedia(id),
     onSuccess: () => {
       queryClient.invalidateQueries({ queryKey: ['media-search'] })
       toast.success('Media deleted')
     },
   })
   ```

## Testing Conventions

- Test files live in `__tests__/` directories beside their components
- Component tests use React Testing Library (`render`, `screen`, `userEvent`)
- Mock contexts with `jest.mock('@/contexts/AuthContext', ...)`
- Mock `react-router-dom` for navigation assertions
- Accessibility tests use `jest-axe`:
  ```tsx
  const { container } = render(<Button>Click</Button>)
  const results = await axe(container)
  expect(results).toHaveNoViolations()
  ```
- Run tests: `cd catalog-web && npm run test`
- Run lint + type checks: `npm run lint && npm run type-check`

## Naming Conventions

- **Components**: PascalCase (`MediaCard.tsx`, `ProtectedRoute.tsx`)
- **Functions/hooks**: camelCase (`useAuth`, `handleMediaUpdate`)
- **API modules**: camelCase (`mediaApi`, `authApi`)
- **Types**: PascalCase (`MediaItem`, `LoginRequest`)
- **Files**: PascalCase for components, camelCase for utilities
- **CSS**: Tailwind utility classes (no separate CSS files)
