import React from 'react';
import { useCoverQuality } from '../../hooks/useCoverQuality';

/**
 * CoverQualityBadge renders the X-Cover-Quality verdict produced by
 * catalog-api's Nexus quality gate. Debug-only: gate via
 * `import.meta.env.DEV` at the call site or pass `debug={false}` in
 * release builds.
 */
export interface CoverQualityBadgeProps {
  coverId: number | string | null;
  debug: boolean;
  className?: string;
}

const verdictClass: Record<string, string> = {
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

export const CoverQualityBadge: React.FC<CoverQualityBadgeProps> = ({
  coverId,
  debug,
  className,
}) => {
  const signal = useCoverQuality(coverId, debug);
  if (!debug || !signal) return null;
  const color = verdictClass[signal.quality] ?? verdictClass.unknown;
  return (
    <span
      className={[
        'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium shadow',
        color,
        className ?? '',
      ].join(' ')}
      title={`Nexus quality — source: ${signal.source || 'n/a'}`}
      data-testid="cover-quality-badge"
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current opacity-80" aria-hidden />
      {signal.quality}
      {signal.source ? <span className="opacity-75">· {signal.source}</span> : null}
    </span>
  );
};
