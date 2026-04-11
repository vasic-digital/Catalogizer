import React, { useEffect, useMemo, useState } from 'react'
import { MediaCard } from './MediaCard'
import { HistoryDrawer } from './HistoryDrawer'
import { getEntityProgress } from '@/lib/playbackApi'
import type { MediaItem } from '@/types/media'
import type { UiPlaybackProgress } from '@/types/playback'
import { motion } from 'framer-motion'

interface MediaGridProps {
  media: MediaItem[]
  loading?: boolean
  viewMode?: 'grid' | 'list'
  onMediaView?: (media: MediaItem) => void
  onMediaPlay?: (media: MediaItem) => void
  onMediaDownload?: (media: MediaItem) => void
  onSelect?: (mediaId: number, selected: boolean) => void
  selectedItems?: number[]
  showActions?: boolean
  className?: string
}

const LoadingSkeleton: React.FC = () => (
  <div className="animate-pulse">
    <div className="aspect-[3/4] bg-gray-300 dark:bg-gray-600 rounded-lg mb-4" />
    <div className="space-y-2">
      <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-3/4" />
      <div className="h-3 bg-gray-300 dark:bg-gray-600 rounded w-1/2" />
      <div className="h-3 bg-gray-300 dark:bg-gray-600 rounded w-2/3" />
    </div>
  </div>
)

/**
 * MediaGrid renders a responsive grid or list of MediaCard components
 * with loading skeletons and empty-state handling.
 */
export const MediaGrid: React.FC<MediaGridProps> = ({
  media,
  loading = false,
  viewMode = 'grid',
  onMediaView,
  onMediaPlay,
  onMediaDownload,
  onSelect: _onSelect,
  selectedItems: _selectedItems = [],
  showActions: _showActions = true,
  className = ''
}) => {
  // Always call hooks at the top level, before any early returns
  const gridCols = useMemo(() => viewMode === 'grid'
    ? 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6'
    : 'grid-cols-1', [viewMode])

  // Per-card reproduction progress + HistoryDrawer wiring. Loads
  // progress in parallel once the media list is known so badges
  // populate without blocking initial render.
  const [progressById, setProgressById] = useState<Record<number, UiPlaybackProgress>>({})
  const [historyTarget, setHistoryTarget] = useState<MediaItem | null>(null)
  const mediaIdsKey = media.map((m) => m.id).join(',')

  useEffect(() => {
    if (media.length === 0) return
    let cancelled = false
    const ids = media.map((m) => m.id)
    Promise.all(ids.map((id) => getEntityProgress(id))).then((results) => {
      if (cancelled) return
      const next: Record<number, UiPlaybackProgress> = {}
      results.forEach((prog, idx) => {
        if (prog) next[ids[idx]] = prog
      })
      setProgressById(next)
    })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mediaIdsKey])

  if (loading) {
    return (
      <div className={`grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6 ${className}`}>
        {Array.from({ length: 12 }).map((_, index) => (
          <LoadingSkeleton key={index} />
        ))}
      </div>
    )
  }

  if (media.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 dark:text-gray-500 mb-4">
          <svg
            className="mx-auto h-12 w-12"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
            />
          </svg>
        </div>
        <h3 className="text-sm font-medium text-gray-900 dark:text-white">No media found</h3>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Try adjusting your search criteria or add some media to your collection.
        </p>
      </div>
    )
  }

  return (
    <>
      <div className={`grid ${gridCols} gap-6 ${className}`}>
        {media.map((item, index) => (
          <motion.div
            key={item.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: index * 0.05 }}
          >
            <MediaCard
              media={item}
              onView={onMediaView}
              onPlay={onMediaPlay}
              onDownload={onMediaDownload}
              progress={progressById[item.id] ?? null}
              onOpenHistory={(m) => setHistoryTarget(m)}
            />
          </motion.div>
        ))}
      </div>
      <HistoryDrawer
        mediaItemId={historyTarget?.id ?? null}
        mediaTitle={historyTarget?.title ?? ''}
        open={historyTarget !== null}
        onClose={() => setHistoryTarget(null)}
      />
    </>
  )
}