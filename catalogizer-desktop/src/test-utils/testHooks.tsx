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
