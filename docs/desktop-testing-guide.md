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
