import React, { Fragment } from 'react'
import { Dialog, Transition } from '@headlessui/react'
import { X, Download, Play, Star, Calendar, Film, HardDrive, Clock, Info } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import type { MediaItem } from '@/types/media'

interface MediaDetailModalProps {
  media: MediaItem | null
  isOpen: boolean
  onClose: () => void
  onDownload?: (media: MediaItem) => void
  onPlay?: (media: MediaItem) => void
}

export const MediaDetailModal: React.FC<MediaDetailModalProps> = ({
  media,
  isOpen,
  onClose,
  onDownload,
  onPlay,
}) => {
  if (!media) return null

  const externalMeta = media.external_metadata?.[0]
  const posterUrl = externalMeta?.poster_url || media.cover_image
  const backdropUrl = externalMeta?.backdrop_url

  const formatFileSize = (bytes?: number) => {
    if (!bytes) return 'Unknown'
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(1024))
    return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`
  }

  const formatDuration = (seconds?: number) => {
    if (!seconds) return 'Unknown'
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    if (hours > 0) {
      return `${hours}h ${minutes}m`
    }
    return `${minutes}m`
  }

  return (
    <Transition appear show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={onClose}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black bg-opacity-75" />
        </Transition.Child>

        <div className="fixed inset-0 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4 text-center">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 scale-95"
              enterTo="opacity-100 scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 scale-100"
              leaveTo="opacity-0 scale-95"
            >
              <Dialog.Panel className="w-full max-w-4xl transform overflow-hidden rounded-2xl bg-white dark:bg-gray-800 text-left align-middle shadow-xl transition-all">
                {/* Backdrop Image */}
                {backdropUrl && (
                  <div className="relative h-64 overflow-hidden">
                    <img
                      src={backdropUrl}
                      alt={media.title}
                      className="w-full h-full object-cover"
                    />
                    <div className="absolute inset-0 bg-gradient-to-t from-white dark:from-gray-800 to-transparent" />
                  </div>
                )}

                {/* Close Button */}
                <button
                  onClick={onClose}
                  className="absolute top-4 right-4 p-2 rounded-full bg-gray-900 bg-opacity-50 text-white hover:bg-opacity-75 transition-all z-10"
                >
                  <X className="h-5 w-5" />
                </button>

                {/* Content */}
                <div className="p-8 -mt-20 relative">
                  <div className="flex gap-8">
                    {/* Poster */}
                    {posterUrl && (
                      <div className="flex-shrink-0">
                        <img
                          src={posterUrl}
                          alt={media.title}
                          className="w-48 h-72 object-cover rounded-lg shadow-2xl"
                        />
                      </div>
                    )}

                    {/* Details */}
                    <div className="flex-1 min-w-0">
                      <Dialog.Title
                        as="h2"
                        className="text-3xl font-bold text-gray-900 dark:text-white mb-2"
                      >
                        {externalMeta?.title || media.title}
                      </Dialog.Title>

                      {/* Meta Info */}
                      <div className="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400 mb-4">
                        {media.year && (
                          <div className="flex items-center gap-1">
                            <Calendar className="h-4 w-4" />
                            <span>{media.year}</span>
                          </div>
                        )}
                        {media.rating && (
                          <div className="flex items-center gap-1">
                            <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                            <span>{media.rating.toFixed(1)}</span>
                          </div>
                        )}
                        {media.media_type && (
                          <div className="flex items-center gap-1">
                            <Film className="h-4 w-4" />
                            <span className="capitalize">{media.media_type.replace('_', ' ')}</span>
                          </div>
                        )}
                        {media.quality && (
                          <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded text-xs font-semibold uppercase">
                            {media.quality}
                          </span>
                        )}
                      </div>

                      {/* Genres */}
                      {externalMeta?.genres && externalMeta.genres.length > 0 && (
                        <div className="flex flex-wrap gap-2 mb-4">
                          {externalMeta.genres.map((genre, index) => (
                            <span
                              key={index}
                              className="px-3 py-1 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-full text-sm"
                            >
                              {genre}
                            </span>
                          ))}
                        </div>
                      )}

                      {/* Description */}
                      {(externalMeta?.description || media.description) && (
                        <p className="text-gray-600 dark:text-gray-300 mb-6 line-clamp-4">
                          {externalMeta?.description || media.description}
                        </p>
                      )}

                      {/* Action Buttons */}
                      <div className="flex gap-3 mb-6">
                        {onPlay && (
                          <Button
                            onClick={() => onPlay(media)}
                            className="flex items-center gap-2"
                          >
                            <Play className="h-4 w-4" />
                            Play
                          </Button>
                        )}
                        {onDownload && (
                          <Button
                            variant="outline"
                            onClick={() => onDownload(media)}
                            className="flex items-center gap-2"
                          >
                            <Download className="h-4 w-4" />
                            Download
                          </Button>
                        )}
                      </div>

                      {/* Technical Details */}
                      <div className="grid grid-cols-2 gap-4">
                        {media.file_size && (
                          <div className="flex items-start gap-2">
                            <HardDrive className="h-4 w-4 text-gray-400 mt-0.5" />
                            <div>
                              <div className="text-xs text-gray-500 dark:text-gray-400">File Size</div>
                              <div className="text-sm font-medium text-gray-900 dark:text-white">
                                {formatFileSize(media.file_size)}
                              </div>
                            </div>
                          </div>
                        )}
                        {media.duration && (
                          <div className="flex items-start gap-2">
                            <Clock className="h-4 w-4 text-gray-400 mt-0.5" />
                            <div>
                              <div className="text-xs text-gray-500 dark:text-gray-400">Duration</div>
                              <div className="text-sm font-medium text-gray-900 dark:text-white">
                                {formatDuration(media.duration)}
                              </div>
                            </div>
                          </div>
                        )}
                        {media.storage_root_name && (
                          <div className="flex items-start gap-2">
                            <Info className="h-4 w-4 text-gray-400 mt-0.5" />
                            <div>
                              <div className="text-xs text-gray-500 dark:text-gray-400">Storage</div>
                              <div className="text-sm font-medium text-gray-900 dark:text-white">
                                {media.storage_root_name}
                              </div>
                            </div>
                          </div>
                        )}
                        {media.storage_root_protocol && (
                          <div className="flex items-start gap-2">
                            <Info className="h-4 w-4 text-gray-400 mt-0.5" />
                            <div>
                              <div className="text-xs text-gray-500 dark:text-gray-400">Protocol</div>
                              <div className="text-sm font-medium text-gray-900 dark:text-white uppercase">
                                {media.storage_root_protocol}
                              </div>
                            </div>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Cast */}
                  {externalMeta?.cast && externalMeta.cast.length > 0 && (
                    <div className="mt-8">
                      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Cast</h3>
                      <div className="flex flex-wrap gap-3">
                        {externalMeta.cast.slice(0, 10).map((actor, index) => (
                          <span
                            key={index}
                            className="px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg text-sm"
                          >
                            {actor}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Versions */}
                  {media.versions && media.versions.length > 0 && (
                    <div className="mt-8">
                      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Available Versions</h3>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {media.versions.map((version) => (
                          <Card key={version.id}>
                            <CardContent className="p-4">
                              <div className="flex justify-between items-start">
                                <div>
                                  <div className="font-medium text-gray-900 dark:text-white">
                                    {version.quality} - {version.resolution}
                                  </div>
                                  <div className="text-sm text-gray-500 dark:text-gray-400">
                                    {version.codec} • {formatFileSize(version.file_size)}
                                  </div>
                                </div>
                                {version.language && (
                                  <span className="text-xs px-2 py-1 bg-gray-200 dark:bg-gray-600 rounded">
                                    {version.language}
                                  </span>
                                )}
                              </div>
                            </CardContent>
                          </Card>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  )
}
