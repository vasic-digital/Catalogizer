import React from 'react'
import { PlaybackFormatter } from '@/lib/playbackFormatter'
import type { UiPlaybackProgress } from '@/types/playback'

interface ProgressBadgeProps {
  progress: UiPlaybackProgress | null | undefined
  onOpenHistory?: () => void
  className?: string
}

/**
 * Reproduction progress overlay for MediaCard. Shows total duration /
 * current position / last-session amount / play count and a progress
 * bar. When `onOpenHistory` is provided, the badge becomes a button
 * that opens the HistoryDrawer — clicks stop propagation so the parent
 * card's onView handler never fires.
 */
export const ProgressBadge: React.FC<ProgressBadgeProps> = ({
  progress,
  onOpenHistory,
  className = '',
}) => {
  if (!progress) return null

  const duration = PlaybackFormatter.formatAmount(
    progress.durationTotal ?? 0,
    progress.positionUnit,
  )
  const currentLabel = PlaybackFormatter.formatProgress(
    progress.lastPosition,
    progress.durationTotal,
    progress.positionUnit,
  )
  const lastSessionLabel =
    PlaybackFormatter.formatAmount(
      progress.lastSessionAmount,
      progress.positionUnit,
    ) + ' last'
  const reps = PlaybackFormatter.formatReproductionCount(
    progress.totalReproductions,
  )
  const pct = PlaybackFormatter.progressFraction(
    progress.lastPosition,
    progress.durationTotal,
  )

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onOpenHistory?.()
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      className={`absolute bottom-2 left-2 right-2 rounded-md bg-black/75 px-2 py-1.5 text-left text-white backdrop-blur-sm hover:bg-black/85 focus:outline-none focus:ring-2 focus:ring-white/60 ${className}`}
      data-testid={`progress-badge-${progress.mediaItemId}`}
      aria-label={`View reproduction history (${reps} played)`}
    >
      <div className="text-[11px] font-semibold leading-tight">{duration}</div>
      <div className="text-[10px] leading-tight opacity-95">{currentLabel}</div>
      <div className="text-[10px] leading-tight opacity-75">{lastSessionLabel}</div>
      <div className="text-[10px] leading-tight opacity-60">{reps} played</div>
      {pct > 0 && (
        <div className="mt-1 h-[3px] w-full overflow-hidden rounded-full bg-white/25">
          <div
            className="h-full rounded-full bg-sky-400"
            style={{ width: `${pct * 100}%` }}
          />
        </div>
      )}
    </button>
  )
}
