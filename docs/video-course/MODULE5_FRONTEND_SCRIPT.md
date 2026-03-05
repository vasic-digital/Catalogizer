# Module 5: Frontend Development - Script

**Duration**: 60 minutes
**Module**: 5 - Frontend Development

---

## Scene 1: React + TypeScript Setup (0:00 - 15:00)

**[Visual: Terminal showing `catalog-web/` directory structure]**

**Narrator**: Welcome to Module 5. The Catalogizer frontend is a React 18 application written in TypeScript, bundled with Vite, styled with Tailwind CSS, and tested with Vitest and Playwright. Let us start with the build configuration.

**[Visual: Open `catalog-web/vite.config.ts`]**

**Narrator**: Vite is the build tool. The configuration defines path aliases that map `@/components`, `@/hooks`, `@/lib`, `@/types`, `@/services`, `@/store`, `@/pages`, and `@/assets` to their respective directories under `src/`. This eliminates deep relative imports throughout the codebase.

```typescript
// catalog-web/vite.config.ts
export default defineConfig({
  resolve: {
    alias: {
      '@/components': resolve(__dirname, 'src/components'),
      '@/hooks': resolve(__dirname, 'src/hooks'),
      '@/lib': resolve(__dirname, 'src/lib'),
      '@/types': resolve(__dirname, 'src/types'),
      '@/services': resolve(__dirname, 'src/services'),
      '@/store': resolve(__dirname, 'src/store'),
      '@/pages': resolve(__dirname, 'src/pages'),
      '@/assets': resolve(__dirname, 'src/assets'),
    },
  },
  // ...
});
```

**[Visual: Show proxy configuration]**

**Narrator**: The API proxy is dynamic. At dev server startup, Vite reads the `../catalog-api/.service-port` file to discover the backend's actual port. This is the same port file the Go backend writes on startup. If the file is missing, it falls back to port 8080.

**[Visual: Show build output chunk splitting]**

**Narrator**: The production build splits output into vendor chunks for optimal caching: `vendor` for React core, `router` for React Router, `ui` for component libraries, `charts` for visualization libraries, and `utils` for utility packages. Each chunk gets a content hash in its filename, so browsers only re-download what has actually changed.

**[Visual: Show linked submodule packages in `package.json`]**

**Narrator**: The frontend depends on several TypeScript submodules linked via `file:../` in `package.json`: `@vasic-digital/websocket-client` for WebSocket communication, `@vasic-digital/ui-components` for shared UI, `@vasic-digital/media-types` for type definitions, `@vasic-digital/catalogizer-api-client` for the API client, and `@vasic-digital/auth-context` for authentication state.

**[Visual: Show environment variables]**

**Narrator**: Environment variables follow Vite conventions -- prefixed with `VITE_`. The frontend reads API base URL, WebSocket URL, and feature flags from the environment.

---

## Scene 2: State Management (15:00 - 30:00)

**[Visual: Diagram showing React Query for server state and Zustand for client state]**

**Narrator**: Catalogizer uses a two-tier state management approach. Server state -- data from the API -- is managed by React Query. Client state -- UI preferences, sidebar visibility, filter selections -- is managed by Zustand.

**[Visual: Open `catalog-web/src/lib/api.ts`]**

**Narrator**: The API layer in `src/lib/api.ts` defines typed fetch functions. Each function specifies the endpoint, HTTP method, request body type, and response type. These functions are consumed by React Query hooks.

```typescript
// catalog-web/src/lib/api.ts
export async function fetchFiles(params: FileListParams): Promise<FileListResponse> {
  const searchParams = new URLSearchParams();
  // ... build query string from params
  const response = await fetch(`/api/v1/files?${searchParams}`);
  if (!response.ok) throw new Error('Failed to fetch files');
  return response.json();
}
```

**[Visual: Show React Query usage in a component]**

**Narrator**: Components use `useQuery` and `useMutation` hooks from React Query. Queries automatically cache, refetch on window focus, and deduplicate concurrent requests. Mutations invalidate related queries on success, keeping the UI in sync.

```typescript
// Example React Query usage
function FileList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['files', storageRootId, page],
    queryFn: () => fetchFiles({ storageRootId, page }),
    staleTime: 30_000, // 30 seconds
  });

  // Mutation that invalidates the file list on success
  const deleteMutation = useMutation({
    mutationFn: deleteFile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['files'] });
    },
  });
}
```

**[Visual: Show Zustand store]**

**Narrator**: Zustand stores are minimal. Each store is a single function that returns state and actions. No reducers, no action types, no boilerplate. The sidebar store, for example, tracks open/closed state and the currently selected section.

**[Visual: Show Context providers in the app root]**

**Narrator**: The app root wraps everything in a provider hierarchy: `AuthProvider` for authentication state, `WebSocketProvider` for real-time connection, and React Query's `QueryClientProvider` for server state. Protected routes use `ProtectedRoute` component that redirects unauthenticated users to the login page.

**[Visual: Open `catalog-web/src/lib/websocket.ts`]**

**Narrator**: The WebSocket integration layer connects React Query to real-time updates. When the server broadcasts a scan progress event or a new entity notification, the WebSocket handler invalidates the appropriate React Query cache keys, triggering automatic UI updates without polling.

---

## Scene 3: Component Architecture (30:00 - 50:00)

**[Visual: Component hierarchy diagram showing layout, pages, and shared components]**

**Narrator**: The component architecture follows atomic design principles. Atoms are basic elements -- buttons, inputs, badges. Molecules combine atoms -- search bars, file cards, navigation items. Organisms are complex features -- the file browser, entity detail view, collection manager.

**[Visual: Show `@vasic-digital/ui-components` submodule usage]**

**Narrator**: Shared UI components live in the `UI-Components-React` submodule, linked as `@vasic-digital/ui-components`. These are project-agnostic: buttons, modals, dropdowns, toast notifications. The Catalogizer frontend imports and extends them with domain-specific styling.

**[Visual: Show form handling with React Hook Form + Zod]**

**Narrator**: Forms use React Hook Form for state management and Zod for schema validation. The schema defines the shape and constraints of form data, and React Hook Form enforces them with zero re-renders.

```typescript
// Example form with Zod validation
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const storageRootSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  protocol: z.enum(['smb', 'ftp', 'nfs', 'webdav', 'local']),
  host: z.string().min(1, 'Host is required'),
  path: z.string().min(1, 'Path is required'),
  username: z.string().optional(),
  password: z.string().optional(),
});

type StorageRootForm = z.infer<typeof storageRootSchema>;

function AddStorageRoot() {
  const { register, handleSubmit, formState: { errors } } = useForm<StorageRootForm>({
    resolver: zodResolver(storageRootSchema),
  });
  // ...
}
```

**[Visual: Show Tailwind CSS usage]**

**Narrator**: Styling uses Tailwind CSS exclusively. No CSS modules, no styled-components. Tailwind's utility classes compose directly in JSX. The design system's colors, spacing, and typography are configured in `tailwind.config.js`.

**[Visual: Show Framer Motion animations]**

**Narrator**: Animations use Framer Motion. Page transitions, list item animations, modal entrances, and loading states are all declarative. The `motion` component wraps standard HTML elements with animation props.

```tsx
// Example Framer Motion usage
import { motion, AnimatePresence } from 'framer-motion';

function MediaCard({ item }: { item: MediaItem }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      transition={{ duration: 0.2 }}
    >
      {/* Card content */}
    </motion.div>
  );
}
```

**[Visual: Show specialized API modules]**

**Narrator**: The `src/lib/` directory contains focused API modules: `mediaApi.ts` for media entity operations, `collectionsApi.ts` for collection management, `favoritesApi.ts` for user favorites, `subtitleApi.ts` for subtitle operations, `conversionApi.ts` for format conversion, and `adminApi.ts` for administration.

---

## Scene 4: Testing Frontend (50:00 - 60:00)

**[Visual: Terminal running `npm run test`]**

**Narrator**: Frontend testing uses Vitest for unit tests and Playwright for end-to-end tests. The project has over 1600 unit tests across 101 test files with zero failures.

**[Visual: Show a unit test file]**

**Narrator**: Unit tests follow the same file convention as Go -- test files sit next to their source files in `__tests__/` directories. Each API module has a corresponding test file.

```typescript
// catalog-web/src/lib/__tests__/api.test.ts
describe('fetchFiles', () => {
  it('should return file list for valid storage root', async () => {
    // Mock fetch
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ files: [], total: 0 }),
    });

    const result = await fetchFiles({ storageRootId: 1, page: 1 });
    expect(result.files).toEqual([]);
    expect(fetch).toHaveBeenCalledWith(expect.stringContaining('/api/v1/files'));
  });
});
```

**[Visual: Show Playwright E2E test]**

**Narrator**: End-to-end tests use Playwright to automate real browser interactions. They verify complete user flows: login, navigate to the file browser, trigger a scan, wait for results, and verify entity creation. The test stack uses `docker-compose.test.yml` with the API, web, and Playwright containers all running on the host network.

**[Visual: Show test commands]**

**Narrator**: The full test suite runs with these commands:

```bash
npm run test           # Unit tests (single run, 1623 tests)
npm run test:watch     # Watch mode for development
npm run test:coverage  # Coverage report
npm run test:e2e       # Playwright E2E tests
```

**[Visual: Course title card]**

**Narrator**: That covers the frontend. You have seen how Vite, React Query, Zustand, Tailwind, and Zod work together in a production application. The architecture is modular, testable, and performant. In Module 6, we add real-time features with WebSocket.

---

## Key Code Examples

### Running the Frontend
```bash
cd catalog-web
npm install
npm run dev  # Reads ../catalog-api/.service-port for API proxy
# Access at http://localhost:3000
```

### Type-Safe API Layer
```typescript
// src/types/media.ts
interface MediaItem {
  id: number;
  media_type_id: number;
  parent_id: number | null;
  title: string;
  original_title: string | null;
  year: number | null;
  rating: number | null;
  description: string | null;
  cover_url: string | null;
  status: string;
}

// src/lib/mediaApi.ts
export async function getMediaItem(id: number): Promise<MediaItem> {
  const response = await fetch(`/api/v1/entities/${id}`);
  if (!response.ok) throw new Error('Failed to fetch media item');
  return response.json();
}
```

### Submodule Dependencies (package.json)
```json
{
  "dependencies": {
    "@vasic-digital/websocket-client": "file:../WebSocket-Client-TS",
    "@vasic-digital/ui-components": "file:../UI-Components-React",
    "@vasic-digital/media-types": "file:../Media-Types-TS",
    "@vasic-digital/catalogizer-api-client": "file:../Catalogizer-API-Client-TS",
    "@vasic-digital/auth-context": "file:../Auth-Context-React"
  }
}
```

---

## Quiz Questions

1. How does the frontend discover the backend's port during development?
   **Answer**: The Vite dev server reads the `../catalog-api/.service-port` file at startup. The Go backend writes its dynamically chosen port to this file. If the file is missing, the frontend falls back to port 8080.

2. Why does Catalogizer use both React Query and Zustand for state management?
   **Answer**: React Query manages server state (data from the API) with automatic caching, refetching, deduplication, and cache invalidation. Zustand manages client state (UI preferences, sidebar state, filter selections) with a minimal API. Separating these concerns prevents mixing server data lifecycle with UI state.

3. How are form validations implemented in the frontend?
   **Answer**: Forms use React Hook Form for state management and Zod for schema validation. Zod schemas define the shape, types, and constraints of form data. The `zodResolver` connects the schema to React Hook Form, which enforces validation without causing unnecessary re-renders.

4. What is the purpose of chunk splitting in the Vite build configuration?
   **Answer**: The build splits output into named chunks (vendor, router, ui, charts, utils) with content hashes in filenames. This enables optimal browser caching: when only application code changes, the vendor chunk (React core) is still served from cache. Users only download the chunks that have actually changed.
