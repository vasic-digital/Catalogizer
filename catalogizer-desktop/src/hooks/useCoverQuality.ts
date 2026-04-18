import { useEffect, useState } from 'react';
import { useConfigStore } from '../stores/configStore';

/**
 * useCoverQuality — mirrors the catalog-web hook of the same name for
 * the Tauri desktop shell. Probes `/api/v1/cover/:id` with a HEAD
 * request and returns the X-Cover-Quality / X-Cover-Source verdict.
 * Debug-only: the hook short-circuits when `debug` is false so release
 * builds pay no network cost. A module-level cache prevents duplicate
 * probes for the same coverId within one process lifetime.
 */
export interface CoverQualitySignal {
  quality: string;
  source: string;
}

const cache = new Map<string, CoverQualitySignal>();

export function useCoverQuality(
  coverId: number | string | null,
  debug: boolean,
): CoverQualitySignal | null {
  const serverUrl = useConfigStore((s) => s.serverUrl);
  const [state, setState] = useState<CoverQualitySignal | null>(() => {
    if (coverId == null) return null;
    return cache.get(String(coverId)) ?? null;
  });

  useEffect(() => {
    if (!debug || coverId == null || !serverUrl) return;
    const key = String(coverId);
    if (cache.has(key)) {
      setState(cache.get(key) ?? null);
      return;
    }
    const controller = new AbortController();
    const base = serverUrl.replace(/\/$/, '');
    fetch(`${base}/api/v1/cover/${encodeURIComponent(key)}`, {
      method: 'HEAD',
      signal: controller.signal,
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
        setState({ quality: 'unknown', source: '' });
      });
    return () => controller.abort();
  }, [coverId, debug, serverUrl]);

  return state;
}

export function resetCoverQualityCache(): void {
  cache.clear();
}
