#!/bin/bash
# Desktop App Test Infrastructure Setup Script for Catalogizer
# This script sets up comprehensive testing for Tauri desktop app

set -e

echo "💻 Catalogizer Desktop Test Infrastructure Setup"
echo "=============================================="

# Check if we're in the right directory
if [ ! -d "catalogizer-desktop" ]; then
    echo "❌ catalogizer-desktop directory not found. Run from project root."
    exit 1
fi

cd catalogizer-desktop

echo "📋 Checking desktop project structure..."
if [ ! -f "package.json" ]; then
    echo "❌ package.json not found."
    exit 1
fi

if [ ! -f "vitest.config.ts" ]; then
    echo "❌ vitest.config.ts not found."
    exit 1
fi

echo "✅ Desktop project structure verified"

# Create test utilities directory
echo "📁 Creating test utilities..."
mkdir -p src/test-utils

# Create test data generator
cat > src/test-utils/testData.ts << 'EOF'
/**
 * Test data generator for Catalogizer Desktop tests
 */

export interface MediaItem {
  id: number;
  title: string;
  type: 'movie' | 'tv_show' | 'music_album' | 'game' | 'book';
  year: number;
  posterPath?: string;
  backdropPath?: string;
  overview: string;
  rating: number;
  runtime?: number;
  genres: string[];
}

export interface User {
  id: number;
  username: string;
  email: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

/**
 * Generates mock media items for testing
 */
export function generateMediaItems(count: number = 10): MediaItem[] {
  const items: MediaItem[] = [];
  const types: MediaItem['type'][] = ['movie', 'tv_show', 'music_album', 'game', 'book'];
  const genres = [
    'Action', 'Adventure', 'Comedy', 'Drama', 'Horror',
    'Sci-Fi', 'Fantasy', 'Romance', 'Thriller', 'Documentary'
  ];

  for (let i = 1; i <= count; i++) {
    const type = types[i % types.length];
    const title = `Test ${type.replace('_', ' ')} ${i}`;
    const itemGenres = [...genres].sort(() => Math.random() - 0.5).slice(0, 3);

    items.push({
      id: i,
      title,
      type,
      year: 2010 + (i % 15),
      posterPath: `/posters/poster_${i}.jpg`,
      backdropPath: `/backdrops/backdrop_${i}.jpg`,
      overview: `This is a test overview for ${title}. It's a great piece of media that everyone should experience.`,
      rating: 5.0 + (i % 5),
      runtime: type === 'movie' || type === 'tv_show' ? 90 + (i % 60) : undefined,
      genres: itemGenres,
    });
  }

  return items;
}

/**
 * Generates mock users for testing
 */
export function generateUsers(count: number = 5): User[] {
  const users: User[] = [];

  for (let i = 1; i <= count; i++) {
    users.push({
      id: i,
      username: `user${i}`,
      email: `user${i}@example.com`,
      createdAt: new Date(Date.now() - i * 86400000),
      updatedAt: new Date(),
    });
  }

  return users;
}

/**
 * Creates a successful API response
 */
export function createSuccessResponse<T>(data: T): ApiResponse<T> {
  return {
    success: true,
    data,
  };
}

/**
 * Creates an error API response
 */
export function createErrorResponse(message: string): ApiResponse<never> {
  return {
    success: false,
    error: message,
  };
}

/**
 * Creates a loading API response
 */
export function createLoadingResponse<T>(): ApiResponse<T> {
  return {
    success: false,
    message: 'Loading...',
  };
}

/**
 * Mock Tauri API functions for testing
 */
export const mockTauriApi = {
  invoke: vi.fn(),
  listen: vi.fn(),
  emit: vi.fn(),
};

/**
 * Resets all mock Tauri API functions
 */
export function resetTauriMocks() {
  mockTauriApi.invoke.mockReset();
  mockTauriApi.listen.mockReset();
  mockTauriApi.emit.mockReset();
}

/**
 * Sets up a successful Tauri API response
 */
export function setupTauriSuccessResponse(data: any) {
  mockTauriApi.invoke.mockResolvedValue(data);
}

/**
 * Sets up a failed Tauri API response
 */
export function setupTauriErrorResponse(error: string) {
  mockTauriApi.invoke.mockRejectedValue(new Error(error));
}
EOF

# Create test render utilities
cat > src/test-utils/testRender.tsx << 'EOF'
/**
 * Test render utilities for React components
 */
import { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';

/**
 * Creates a test QueryClient with default options
 */
export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

/**
 * Custom render function that wraps components with necessary providers
 */
export function renderWithProviders(
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) {
  const queryClient = createTestQueryClient();

  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          {children}
        </BrowserRouter>
      </QueryClientProvider>
    );
  }

  return {
    ...render(ui, { wrapper: Wrapper, ...options }),
    queryClient,
  };
}

/**
 * Renders a component with router context
 */
export function renderWithRouter(
  ui: ReactElement,
  { route = '/' } = {}
) {
  window.history.pushState({}, 'Test page', route);

  return renderWithProviders(ui);
}

/**
 * Waits for the next tick (useful for async updates)
 */
export function waitForNextTick() {
  return new Promise(resolve => setTimeout(resolve, 0));
}
EOF

# Create test hooks utilities
cat > src/test-utils/testHooks.tsx << 'EOF'
/**
 * Utilities for testing React hooks
 */
import { renderHook, RenderHookResult } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';
import { BrowserRouter } from 'react-router-dom';

/**
 * Creates a wrapper for hook tests
 */
export function createHookWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  });

  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          {children}
        </BrowserRouter>
      </QueryClientProvider>
    );
  };
}

/**
 * Renders a hook with all necessary providers
 */
export function renderHookWithProviders<TResult, TProps>(
  render: (props: TProps) => TResult,
  options?: {
    initialProps?: TProps;
    wrapper?: React.ComponentType;
  }
): RenderHookResult<TResult, TProps> {
  const wrapper = options?.wrapper || createHookWrapper();

  return renderHook(render, {
    wrapper,
    initialProps: options?.initialProps,
  });
}
EOF

# Create component test templates
cat > src/test-utils/componentTemplates.tsx << 'EOF'
/**
 * Component test templates and utilities
 */
import { vi } from 'vitest';
import { generateMediaItems } from './testData';

/**
 * Mock useNavigate hook
 */
export const mockNavigate = vi.fn();

/**
 * Mock useLocation hook
 */
export const mockLocation = {
  pathname: '/',
  search: '',
  hash: '',
  state: null,
};

/**
 * Mock React Router hooks
 */
export function mockRouterHooks() {
  return {
    useNavigate: () => mockNavigate,
    useLocation: () => mockLocation,
    useParams: () => ({}),
    useSearchParams: () => [new URLSearchParams(), vi.fn()],
  };
}

/**
 * Mock Zustand store
 */
export function createMockStore<T>(initialState: T) {
  const state = { ...initialState };
  const listeners = new Set<() => void>();

  return {
    getState: () => state,
    setState: (partial: Partial<T> | ((state: T) => Partial<T>)) => {
      const newPartial = typeof partial === 'function' ? partial(state) : partial;
      Object.assign(state, newPartial);
      listeners.forEach(listener => listener());
    },
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
  };
}

/**
 * Common test props for media components
 */
export const mediaTestProps = {
  mediaItems: generateMediaItems(5),
  isLoading: false,
  error: null as string | null,
  onMediaClick: vi.fn(),
  onLoadMore: vi.fn(),
  hasMore: true,
};

/**
 * Common test props for user components
 */
export const userTestProps = {
  user: {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    createdAt: new Date(),
    updatedAt: new Date(),
  },
  isAuthenticated: true,
  isLoading: false,
  onLogin: vi.fn(),
  onLogout: vi.fn(),
  onRegister: vi.fn(),
};
EOF

# Update vitest config for better test coverage
echo "🔧 Updating vitest.config.ts for better test coverage..."
if [ -f "vitest.config.ts" ]; then
    cat > vitest.config.ts << 'EOF'
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test-utils/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test-utils/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/__tests__/**',
        '**/*.test.*',
        '**/*.spec.*',
      ],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 80,
        statements: 80,
      },
    },
    include: ['src/**/*.{test,spec}.{js,jsx,ts,tsx}'],
    exclude: ['node_modules', 'dist', '.idea', '.git', '.cache'],
  },
  resolve: {
    alias: {
      '@': '/src',
    },
  },
});
EOF
    echo "✅ Updated vitest.config.ts"
else
    echo "⚠️ vitest.config.ts not found, creating new one..."
    cat > vitest.config.ts << 'EOF'
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test-utils/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test-utils/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/__tests__/**',
        '**/*.test.*',
        '**/*.spec.*',
      ],
    },
    include: ['src/**/*.{test,spec}.{js,jsx,ts,tsx}'],
  },
  resolve: {
    alias: {
      '@': '/src',
    },
  },
});
EOF
    echo "✅ Created vitest.config.ts"
fi

# Create test setup file
echo "📝 Creating test setup file..."
cat > src/test-utils/setup.ts << 'EOF'
/**
 * Global test setup file for Vitest
 */
import '@testing-library/jest-dom';
import { vi } from 'vitest';
import { cleanup } from '@testing-library/react';

// Mock Tauri API
vi.mock('@tauri-apps/api', () => ({
  invoke: vi.fn(),
  listen: vi.fn(),
  emit: vi.fn(),
}));

// Mock Tauri plugin shell
vi.mock('@tauri-apps/plugin-shell', () => ({
  open: vi.fn(),
}));

// Clear all mocks and cleanup after each test
afterEach(() => {
  vi.clearAllMocks();
  cleanup();
});

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock IntersectionObserver
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
  takeRecords: vi.fn(),
}));
EOF

# Create test coverage script
echo "📝 Creating test coverage script..."
cat > ../scripts/run-desktop-tests.sh << 'EOF'
#!/bin/bash
# Run desktop app tests with coverage reporting

set -e

echo "💻 Running Catalogizer Desktop Tests"
echo "==================================="

cd catalogizer-desktop

echo "🔧 Installing dependencies if needed..."
if [ ! -d "node_modules" ]; then
    npm install
fi

echo "🧪 Running tests with coverage..."
npm run test:coverage

echo "📊 Coverage reports generated:"
echo "   - HTML: coverage/index.html"
echo "   - Text: coverage/coverage-final.json"

# Check if coverage meets threshold
COVERAGE_SUMMARY="coverage/coverage-summary.json"
if [ -f "$COVERAGE_SUMMARY" ]; then
    echo "📈 Checking coverage threshold..."
    
    # Extract coverage percentages (simplified)
    if command -v jq &> /dev/null; then
        LINES_COV=$(jq -r '.total.lines.pct' "$COVERAGE_SUMMARY")
        STATEMENTS_COV=$(jq -r '.total.statements.pct' "$COVERAGE_SUMMARY")
        FUNCTIONS_COV=$(jq -r '.total.functions.pct' "$COVERAGE_SUMMARY")
        BRANCHES_COV=$(jq -r '.total.branches.pct' "$COVERAGE_SUMMARY")
        
        echo "✅ Coverage Summary:"
        echo "   - Lines: ${LINES_COV}%"
        echo "   - Statements: ${STATEMENTS_COV}%"
        echo "   - Functions: ${FUNCTIONS_COV}%"
        echo "   - Branches: ${BRANCHES_COV}%"
        
        # Check against thresholds
        THRESHOLD=80
        if (( $(echo "$LINES_COV < $THRESHOLD" | bc -l) )); then
            echo "⚠️ Lines coverage below ${THRESHOLD}% target. Consider adding more tests."
        else
            echo "🎉 Lines coverage meets ${THRESHOLD}% target!"
        fi
    else
        echo "⚠️ jq not installed. Install jq to parse coverage summary."
        echo "   Coverage report available at: coverage/index.html"
    fi
else
    echo "⚠️ Coverage summary not found at $COVERAGE_SUMMARY"
    echo "   Raw coverage data available at: coverage/coverage-final.json"
fi

echo ""
echo "🚀 To view coverage report:"
echo "   open coverage/index.html"
EOF

chmod +x ../scripts/run-desktop-tests.sh

# Create example test for Tauri commands
echo "📝 Creating example Tauri command test..."
mkdir -p src/__tests__

cat > src/__tests__/tauriCommands.test.ts << 'EOF'
/**
 * Example tests for Tauri commands
 */
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { invoke } from '@tauri-apps/api';
import { mockTauriApi, setupTauriSuccessResponse, setupTauriErrorResponse } from '@/test-utils/testData';

// Mock the Tauri API
vi.mock('@tauri-apps/api', () => ({
  invoke: mockTauriApi.invoke,
}));

describe('Tauri Commands', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('get_app_info', () => {
    it('should return app info successfully', async () => {
      // Given
      const mockAppInfo = {
        name: 'Catalogizer Desktop',
        version: '1.0.0',
        author: 'Catalogizer Team',
      };
      setupTauriSuccessResponse(mockAppInfo);

      // When
      const result = await invoke('get_app_info');

      // Then
      expect(result).toEqual(mockAppInfo);
      expect(mockTauriApi.invoke).toHaveBeenCalledWith('get_app_info');
    });

    it('should handle errors when getting app info', async () => {
      // Given
      const errorMessage = 'Failed to get app info';
      setupTauriErrorResponse(errorMessage);

      // When & Then
      await expect(invoke('get_app_info')).rejects.toThrow(errorMessage);
    });
  });

  describe('get_system_info', () => {
    it('should return system info successfully', async () => {
      // Given
      const mockSystemInfo = {
        os: 'Linux',
        arch: 'x86_64',
        memory: 8192,
        cores: 8,
      };
      setupTauriSuccessResponse(mockSystemInfo);

      // When
      const result = await invoke('get_system_info');

      // Then
      expect(result).toEqual(mockSystemInfo);
      expect(mockTauriApi.invoke).toHaveBeenCalledWith('get_system_info');
    });
  });
});
EOF

# Create example component test
cat > src/__tests__/exampleComponent.test.tsx << 'EOF'
/**
 * Example component test
 */
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { generateMediaItems } from '@/test-utils/testData';
import ExampleMediaGrid from '@/components/ExampleMediaGrid';

// Mock component for example
const ExampleMediaGrid = ({ mediaItems, isLoading, onMediaClick }: any) => {
  if (isLoading) {
    return <div data-testid="loading">Loading...</div>;
  }

  return (
    <div data-testid="media-grid">
      {mediaItems.map((item: any) => (
        <div
          key={item.id}
          data-testid={`media-item-${item.id}`}
          onClick={() => onMediaClick(item)}
        >
          <h3>{item.title}</h3>
          <p>{item.year}</p>
        </div>
      ))}
    </div>
  );
};

describe('ExampleMediaGrid', () => {
  it('should render loading state', () => {
    // Given
    const props = {
      mediaItems: [],
      isLoading: true,
      onMediaClick: vi.fn(),
    };

    // When
    render(<ExampleMediaGrid {...props} />);

    // Then
    expect(screen.getByTestId('loading')).toBeInTheDocument();
    expect(screen.queryByTestId('media-grid')).not.toBeInTheDocument();
  });

  it('should render media items', () => {
    // Given
    const mediaItems = generateMediaItems(3);
    const onMediaClick = vi.fn();
    const props = {
      mediaItems,
      isLoading: false,
      onMediaClick,
    };

    // When
    render(<ExampleMediaGrid {...props} />);

    // Then
    expect(screen.getByTestId('media-grid')).toBeInTheDocument();
    expect(screen.getAllByTestId(/media-item-/)).toHaveLength(3);
    
    mediaItems.forEach(item => {
      expect(screen.getByText(item.title)).toBeInTheDocument();
      expect(screen.getByText(item.year.toString())).toBeInTheDocument();
    });
  });

  it('should call onMediaClick when item is clicked', () => {
    // Given
    const mediaItems = generateMediaItems(2);
    const onMediaClick = vi.fn();
    const props = {
      mediaItems,
      isLoading: false,
      onMediaClick,
    };

    // When
    render(<ExampleMediaGrid {...props} />);
    const firstItem = screen.getByTestId('media-item-1');
    fireEvent.click(firstItem);

    // Then
    expect(onMediaClick).toHaveBeenCalledTimes(1);
    expect(onMediaClick).toHaveBeenCalledWith(mediaItems[0]);
  });
});
EOF

# Create README for desktop testing
cat > ../docs/desktop-testing-guide.md << 'EOF'
# Desktop App Testing Guide for Catalogizer

## Overview

This guide covers the test infrastructure setup for Catalogizer Desktop application (Tauri + React). The goal is to achieve **80%+ test coverage** across all components.

## Test Structure

### 1. Unit Tests (`src/__tests__/`)
- **Location**: `src/__tests__/` and `src/**/__tests__/`
- **Purpose**: Test React components, hooks, utilities, services
- **Frameworks**: Vitest, React Testing Library, Jest DOM
- **Coverage Target**: 80%+

### 2. Integration Tests
- **Location**: `src/__tests__/integration/`
- **Purpose**: Test component interactions, API integrations
- **Frameworks**: Vitest, React Testing Library
- **Coverage Target**: 70%+

### 3. Tauri Command Tests
- **Location**: `src/__tests__/tauri/`
- **Purpose**: Test Rust Tauri command mocking and integration
- **Frameworks**: Vitest
- **Coverage Target**: 90%+

## Test Utilities

We've created several test utility files:

### `src/test-utils/testData.ts`
- Mock data generators for media items, users, API responses
- Tauri API mocking utilities
- Response factory functions

### `src/test-utils/testRender.tsx`
- Custom render functions with providers (QueryClient, Router)
- Test wrapper creation utilities
- Async test helpers

### `src/test-utils/testHooks.tsx`
- Hook testing utilities
- Provider wrappers for hook tests

### `src/test-utils/componentTemplates.tsx`
- Common test props and mocks
- Router hook mocking
- Zustand store mocking

### `src/test-utils/setup.ts`
- Global test setup
- Tauri API mocking
- Browser API polyfills

## Running Tests

### Run All Tests
```bash
./scripts/run-desktop-tests.sh
```

### Run Tests in Watch Mode
```bash
cd catalogizer-desktop
npm run test:watch
```

### Run Tests Once
```bash
cd catalogizer-desktop
npm run test
```

### Generate Coverage Report
```bash
cd catalogizer-desktop
npm run test:coverage
```

## Coverage Reports

After running tests, coverage reports are available at:
- **HTML Report**: `coverage/index.html`
- **JSON Report**: `coverage/coverage-final.json`
- **Summary**: `coverage/coverage-summary.json`

## Test Patterns

### Component Testing
```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import MediaGrid from '@/components/MediaGrid';
import { generateMediaItems } from '@/test-utils/testData';

describe('MediaGrid', () => {
  it('should render media items', () => {
    // Given
    const mediaItems = generateMediaItems(3);
    const onItemClick = vi.fn();
    
    // When
    render(<MediaGrid items={mediaItems} onItemClick={onItemClick} />);
    
    // Then
    expect(screen.getAllByRole('article')).toHaveLength(3);
    mediaItems.forEach(item => {
      expect(screen.getByText(item.title)).toBeInTheDocument();
    });
  });
  
  it('should call onItemClick when item is clicked', () => {
    // Given
    const mediaItems = generateMediaItems(1);
    const onItemClick = vi.fn();
    
    // When
    render(<MediaGrid items={mediaItems} onItemClick={onItemClick} />);
    fireEvent.click(screen.getByRole('article'));
    
    // Then
    expect(onItemClick).toHaveBeenCalledWith(mediaItems[0]);
  });
});
```

### Hook Testing
```typescript
import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { useMediaSearch } from '@/hooks/useMediaSearch';
import { renderHookWithProviders } from '@/test-utils/testHooks';

describe('useMediaSearch', () => {
  it('should search for media', async () => {
    // Given
    const mockResults = generateMediaItems(5);
    const mockApi = {
      searchMedia: vi.fn().mockResolvedValue(mockResults),
    };
    
    // When
    const { result } = renderHookWithProviders(() => 
      useMediaSearch(mockApi)
    );
    
    await act(async () => {
      await result.current.search('action');
    });
    
    // Then
    expect(mockApi.searchMedia).toHaveBeenCalledWith('action');
    expect(result.current.results).toEqual(mockResults);
    expect(result.current.isLoading).toBe(false);
  });
});
```

### Tauri Command Testing
```typescript
import { vi, describe, it, expect } from 'vitest';
import { invoke } from '@tauri-apps/api';
import { mockTauriApi, setupTauriSuccessResponse } from '@/test-utils/testData';

// Mock Tauri API
vi.mock('@tauri-apps/api', () => ({
  invoke: mockTauriApi.invoke,
}));

describe('Tauri Commands', () => {
  it('should get app version', async () => {
    // Given
    const mockVersion = '1.0.0';
    setupTauriSuccessResponse(mockVersion);
    
    // When
    const result = await invoke('get_version');
    
    // Then
    expect(result).toBe(mockVersion);
    expect(mockTauriApi.invoke).toHaveBeenCalledWith('get_version');
  });
});
```

## Best Practices

1. **Test Naming**: Use descriptive test names
2. **Arrange-Act-Assert**: Structure tests clearly
3. **Mock External Dependencies**: Mock Tauri API, network requests, browser APIs
4. **Test User Interactions**: Test clicks, inputs, navigation
5. **Test Error States**: Include error cases and edge cases
6. **Test Loading States**: Test loading indicators and async behavior
7. **Clean Up**: Reset mocks and cleanup after each test
8. **Use Test Utilities**: Leverage the provided test utilities

## Coverage Goals

| Component | Target Coverage | Current Coverage |
|-----------|----------------|------------------|
| React Components | 85% | TBD |
| Custom Hooks | 90% | TBD |
| Services/API | 90% | TBD |
| Utilities | 95% | TBD |
| Tauri Integration | 80% | TBD |
| **Overall** | **85%** | **0%** |

## Next Steps

1. Run existing tests: `./scripts/run-desktop-tests.sh`
2. Review coverage report
3. Identify gaps in test coverage
4. Add tests for untested components
5. Aim for incremental coverage improvement
6. Integrate with CI/CD pipeline

## Troubleshooting

### Tests Not Running
- Check if `node_modules` is installed
- Verify Vitest configuration in `vitest.config.ts`
- Ensure test files match the pattern `*.test.ts` or `*.spec.ts`

### Coverage Not Reported
- Run `npm run test:coverage` specifically
- Check Vitest coverage configuration
- Verify test files are actually executing

### Tauri Mocking Issues
- Ensure Tauri APIs are mocked in `setup.ts`
- Use the provided `mockTauriApi` utilities
- Mock specific commands as needed for each test

### React Testing Issues
- Use `renderWithProviders` for components with context
- Mock hooks that use Tauri or other external APIs
- Use `waitFor` for async updates
EOF

echo "✅ Desktop testing guide created"

cd ..

echo ""
echo "🎉 Desktop Test Infrastructure Setup Complete!"
echo "============================================="
echo ""
echo "📋 Available commands:"
echo "   ./scripts/run-desktop-tests.sh      - Run desktop tests with coverage"
echo ""
echo "📚 Documentation:"
echo "   docs/desktop-testing-guide.md       - Complete desktop testing guide"
echo ""
echo "🔧 Next Steps:"
echo "   1. Review the desktop testing guide"
echo "   2. Run: ./scripts/run-desktop-tests.sh"
echo "   3. Check test coverage report"
echo "   4. Add tests for components with low coverage"
echo "   5. Aim for 80%+ test coverage"
echo ""
echo "🚀 Happy testing!"