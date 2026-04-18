import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useCoverQuality, resetCoverQualityCache } from '../useCoverQuality';
import { useConfigStore } from '../../stores/configStore';

vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(),
}));

const originalFetch = global.fetch;

describe('useCoverQuality (desktop)', () => {
  beforeEach(() => {
    resetCoverQualityCache();
    useConfigStore.setState({
      serverUrl: 'http://localhost:8080',
      theme: 'dark',
      autoStart: false,
      isLoading: false,
    });
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it('returns null when debug=false', () => {
    const { result } = renderHook(() => useCoverQuality(42, false));
    expect(result.current).toBeNull();
  });

  it('returns null when coverId is null', () => {
    const { result } = renderHook(() => useCoverQuality(null, true));
    expect(result.current).toBeNull();
  });

  it('reads X-Cover-Quality + X-Cover-Source on HEAD success', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      headers: {
        get: (name: string) => {
          const map: Record<string, string> = {
            'X-Cover-Quality': 'pass',
            'X-Cover-Source': 'tmdb',
          };
          return map[name] ?? null;
        },
      },
    }) as unknown as typeof fetch;
    const { result } = renderHook(() => useCoverQuality(42, true));
    await waitFor(() =>
      expect(result.current).toEqual({ quality: 'pass', source: 'tmdb' }),
    );
  });

  it('falls back to unknown on network error', async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useCoverQuality(99, true));
    await waitFor(() =>
      expect(result.current).toEqual({ quality: 'unknown', source: '' }),
    );
  });

  it('does not fire a second HEAD for the same coverId', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      headers: {
        get: (name: string) =>
          ({ 'X-Cover-Quality': 'pass', 'X-Cover-Source': 'tmdb' })[name] ?? null,
      },
    });
    global.fetch = fetchMock as unknown as typeof fetch;
    const first = renderHook(() => useCoverQuality(7, true));
    await waitFor(() => expect(first.result.current?.quality).toBe('pass'));
    const second = renderHook(() => useCoverQuality(7, true));
    expect(second.result.current?.quality).toBe('pass');
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('does not probe when serverUrl is unset', () => {
    useConfigStore.setState({
      serverUrl: null,
      theme: 'dark',
      autoStart: false,
      isLoading: false,
    });
    const fetchMock = vi.fn();
    global.fetch = fetchMock as unknown as typeof fetch;
    renderHook(() => useCoverQuality(1, true));
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('aborts the in-flight HEAD on unmount', async () => {
    const abortSpy = vi.fn();
    global.fetch = vi.fn().mockImplementation(() => new Promise(() => {})) as unknown as typeof fetch;
    const origAbort = AbortController.prototype.abort;
    AbortController.prototype.abort = function (...args) {
      abortSpy();
      return origAbort.apply(this, args as []);
    };
    try {
      const { unmount } = renderHook(() => useCoverQuality(11, true));
      act(() => {
        unmount();
      });
      expect(abortSpy).toHaveBeenCalled();
    } finally {
      AbortController.prototype.abort = origAbort;
    }
  });
});
