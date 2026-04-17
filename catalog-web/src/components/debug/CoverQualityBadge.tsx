import React from 'react';
import { useCoverQuality } from '@/hooks/useCoverQuality';

/**
 * CoverQualityBadge renders a small pill showing the X-Cover-Quality
 * verdict and X-Cover-Source of a cover image. It only fires a
 * network request when `debug` is true — production builds should pass
 * `debug={false}` or gate the whole component behind a dev-only flag.
 */
export interface CoverQualityBadgeProps {
  coverId: number | string | null;
  debug: boolean;
  className?: string;
}

const verdictColors: Record<string, string> = {
  pass: 'bg-emerald-500 text-white',
  fail_lowres: 'bg-amber-500 text-white',
  fail_blurry: 'bg-amber-600 text-white',
  fail_small_bytes: 'bg-amber-600 text-white',
  fail_corrupt: 'bg-rose-600 text-white',
  fail_wrong_aspect: 'bg-amber-500 text-white',
  fail_too_large: 'bg-rose-500 text-white',
  placeholder_fallback: 'bg-slate-500 text-white',
  unknown: 'bg-slate-400 text-white',
};

export const CoverQualityBadge: React.FC<CoverQualityBadgeProps> = ({ coverId, debug, className }) => {
  const signal = useCoverQuality(coverId, debug);
  if (!debug || !signal) {
    return null;
  }
  const color = verdictColors[signal.quality] ?? verdictColors.unknown;
  return (
    <span
      className={[
        'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium shadow',
        color,
        className ?? '',
      ].join(' ')}
      title={`Nexus Quality Gate — source: ${signal.source || 'n/a'}`}
      data-testid="cover-quality-badge"
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current opacity-80" aria-hidden />
      {signal.quality}
      {signal.source ? <span className="opacity-75">· {signal.source}</span> : null}
    </span>
  );
};

CoverQualityBadge.displayName = 'CoverQualityBadge';
