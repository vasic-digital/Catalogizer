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
