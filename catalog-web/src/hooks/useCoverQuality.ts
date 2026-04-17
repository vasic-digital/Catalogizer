import { useEffect, useState } from 'react';

/**
 * useCoverQuality — inspects the X-Cover-Quality / X-Cover-Source
 * response headers that catalog-api emits on `/api/v1/cover/:id`.
 * The hook is intentionally side-effect-free for non-debug builds:
 * it lazily fires a HEAD request only when the caller's `debug`
 * flag is true, and caches the result in a module-level Map so the
 * same cover never gets probed twice.
 *
 * Surface the returned `{ quality, source }` via a dev-only overlay
 * so engineers can see which provider served each cover while
 * navigating the app. Production builds that omit `debug` pay no
 * network cost.
 */
export interface CoverQualitySignal {
  quality: string;
  source: string;
}

const cache = new Map<string, CoverQualitySignal>();

export function useCoverQuality(coverId: number | string | null, debug: boolean): CoverQualitySignal | null {
  const [state, setState] = useState<CoverQualitySignal | null>(() => {
    if (coverId == null) return null;
    return cache.get(String(coverId)) ?? null;
  });

  useEffect(() => {
    if (!debug || coverId == null) {
      return;
    }
    const key = String(coverId);
    if (cache.has(key)) {
      setState(cache.get(key) ?? null);
      return;
    }
    const controller = new AbortController();
    fetch(`/api/v1/cover/${encodeURIComponent(key)}`, {
      method: 'HEAD',
      signal: controller.signal,
      credentials: 'include',
    })
      .then((resp) => {
        const signal: CoverQualitySignal = {
          quality: resp.headers.get('X-Cover-Quality') ?? 'unknown',
          source: resp.headers.get('X-Cover-Source') ?? '',
        };
        cache.set(key, signal);
        setState(signal);
      })
      .catch((err: unknown) => {
        if ((err as { name?: string }).name === 'AbortError') return;
        // Non-fatal — the overlay just shows 'unknown'.
        setState({ quality: 'unknown', source: '' });
      });
    return () => controller.abort();
  }, [coverId, debug]);

  return state;
}

/**
 * resetCoverQualityCache clears the module-level cache. Tests call it
 * before asserting fresh requests; production code should leave it
 * alone so repeated navigations do not burn extra HEAD calls.
 */
export function resetCoverQualityCache(): void {
  cache.clear();
}
