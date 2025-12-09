import React from 'react'
import { MediaCard } from './MediaCard'
import type { MediaItem } from '@/types/media'
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

export const MediaGrid: React.FC<MediaGridProps> = ({
  media,
  loading = false,
  viewMode = 'grid',
  onMediaView,
  onMediaPlay,
  onMediaDownload,
  onSelect,
  selectedItems = [],
  showActions = true,
  className = ''
}) => {
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

  const gridCols = viewMode === 'grid' 
    ? 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6'
    : 'grid-cols-1'

  return (
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
          />
        </motion.div>
      ))}
    </div>
  )
}