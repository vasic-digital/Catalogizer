import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { entityApi } from '@/lib/mediaApi'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import {
  ArrowLeft, Film, Tv, Music, Gamepad2, Monitor, BookOpen, Book,
  Star, Clock, Globe, Calendar, Folder, FileText, Copy, Heart,
  RefreshCw, ChevronRight,
} from 'lucide-react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import type { MediaEntity, MediaEntityDetail, EntityFile } from '@/types/media'

const TYPE_ICONS: Record<string, React.ElementType> = {
  movie: Film,
  tv_show: Tv,
  tv_season: Tv,
  tv_episode: Tv,
  music_artist: Music,
  music_album: Music,
  song: Music,
  game: Gamepad2,
  software: Monitor,
  book: BookOpen,
  comic: Book,
}

export function EntityDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const entityId = parseInt(id || '0', 10)

  const { data: entity, isLoading } = useQuery({
    queryKey: ['entity', entityId],
    queryFn: () => entityApi.getEntity(entityId),
    enabled: entityId > 0,
  })

  const { data: childrenData } = useQuery({
    queryKey: ['entityChildren', entityId],
    queryFn: () => entityApi.getEntityChildren(entityId),
    enabled: entityId > 0 && (entity?.children_count || 0) > 0,
  })

  const { data: filesData } = useQuery({
    queryKey: ['entityFiles', entityId],
    queryFn: () => entityApi.getEntityFiles(entityId),
    enabled: entityId > 0,
  })

  const { data: duplicatesData } = useQuery({
    queryKey: ['entityDuplicates', entityId],
    queryFn: () => entityApi.getEntityDuplicates(entityId),
    enabled: entityId > 0,
  })

  const refreshMutation = useMutation({
    mutationFn: () => entityApi.refreshEntityMetadata(entityId),
    onSuccess: () => {
      toast.success('Metadata refresh queued')
      queryClient.invalidateQueries({ queryKey: ['entity', entityId] })
    },
    onError: () => toast.error('Failed to refresh metadata'),
  })

  const favoriteMutation = useMutation({
    mutationFn: (favorite: boolean) =>
      entityApi.updateUserMetadata(entityId, { favorite }),
    onSuccess: () => {
      toast.success('Updated')
      queryClient.invalidateQueries({ queryKey: ['entity', entityId] })
    },
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    )
  }

  if (!entity) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 dark:text-gray-400">Entity not found</p>
        <Button variant="outline" className="mt-4" onClick={() => navigate('/browse')}>
          Back to Browse
        </Button>
      </div>
    )
  }

  const Icon = TYPE_ICONS[entity.media_type] || Film
  const children = childrenData?.items || []
  const files = filesData?.files || []
  const duplicates = duplicatesData?.duplicates || []

  return (
    <div className="space-y-6">
      {/* Back button + breadcrumb */}
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <span className="text-sm text-gray-500 dark:text-gray-400 capitalize">
          {entity.media_type.replace(/_/g, ' ')}
        </span>
        <ChevronRight className="h-4 w-4 text-gray-400" />
        <span className="text-sm font-medium text-gray-900 dark:text-white">
          {entity.title}
        </span>
      </div>

      {/* Hero section */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <Card>
          <CardContent className="p-6">
            <div className="flex items-start gap-6">
              <div className="w-20 h-20 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0">
                <Icon className="h-10 w-10 text-white" />
              </div>
              <div className="flex-1 min-w-0">
                <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                  {entity.title}
                </h1>
                {entity.original_title && entity.original_title !== entity.title && (
                  <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                    {entity.original_title}
                  </p>
                )}

                <div className="flex flex-wrap items-center gap-4 mt-3 text-sm text-gray-600 dark:text-gray-400">
                  <span className="capitalize px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded">
                    {entity.media_type.replace(/_/g, ' ')}
                  </span>
                  {entity.year && (
                    <span className="flex items-center gap-1">
                      <Calendar className="h-4 w-4" />
                      {entity.year}
                    </span>
                  )}
                  {entity.rating != null && (
                    <span className="flex items-center gap-1">
                      <Star className="h-4 w-4 text-yellow-500" />
                      {entity.rating.toFixed(1)}
                    </span>
                  )}
                  {entity.runtime != null && (
                    <span className="flex items-center gap-1">
                      <Clock className="h-4 w-4" />
                      {entity.runtime} min
                    </span>
                  )}
                  {entity.language && (
                    <span className="flex items-center gap-1">
                      <Globe className="h-4 w-4" />
                      {entity.language}
                    </span>
                  )}
                </div>

                {entity.genre && entity.genre.length > 0 && (
                  <div className="flex flex-wrap gap-2 mt-3">
                    {entity.genre.map((g) => (
                      <span
                        key={g}
                        className="px-3 py-1 text-xs font-medium bg-gray-100 dark:bg-gray-800 rounded-full text-gray-700 dark:text-gray-300"
                      >
                        {g}
                      </span>
                    ))}
                  </div>
                )}

                {entity.director && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
                    Directed by <span className="font-medium">{entity.director}</span>
                  </p>
                )}

                {entity.description && (
                  <p className="text-gray-700 dark:text-gray-300 mt-4 leading-relaxed">
                    {entity.description}
                  </p>
                )}
              </div>

              {/* Actions */}
              <div className="flex flex-col gap-2 flex-shrink-0">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => favoriteMutation.mutate(true)}
                >
                  <Heart className="h-4 w-4 mr-1" />
                  Favorite
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => refreshMutation.mutate()}
                  disabled={refreshMutation.isPending}
                >
                  <RefreshCw className={`h-4 w-4 mr-1 ${refreshMutation.isPending ? 'animate-spin' : ''}`} />
                  Refresh
                </Button>
              </div>
            </div>

            {/* Stats row */}
            <div className="flex gap-6 mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
              <div className="text-center">
                <div className="text-lg font-semibold text-gray-900 dark:text-white">{entity.file_count}</div>
                <div className="text-xs text-gray-500">Files</div>
              </div>
              <div className="text-center">
                <div className="text-lg font-semibold text-gray-900 dark:text-white">{entity.children_count}</div>
                <div className="text-xs text-gray-500">Children</div>
              </div>
              {duplicates.length > 0 && (
                <div className="text-center">
                  <div className="text-lg font-semibold text-orange-600">{duplicates.length}</div>
                  <div className="text-xs text-gray-500">Duplicates</div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* Children (seasons, episodes, songs, etc.) */}
      {children.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Folder className="h-5 w-5" />
              {entity.media_type === 'tv_show'
                ? 'Seasons'
                : entity.media_type === 'tv_season'
                  ? 'Episodes'
                  : entity.media_type === 'music_album'
                    ? 'Tracks'
                    : 'Children'}
              <span className="text-sm font-normal text-gray-500">({children.length})</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {children.map((child) => (
                <button
                  key={child.id}
                  onClick={() => navigate(`/entity/${child.id}`)}
                  className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors text-left"
                >
                  <div className="w-8 h-8 rounded bg-gray-100 dark:bg-gray-700 flex items-center justify-center text-sm font-medium text-gray-600 dark:text-gray-400">
                    {child.season_number || child.episode_number || child.track_number || '#'}
                  </div>
                  <div className="flex-1 min-w-0">
                    <span className="font-medium text-gray-900 dark:text-white truncate block">
                      {child.title}
                    </span>
                  </div>
                  {child.year && (
                    <span className="text-sm text-gray-500">{child.year}</span>
                  )}
                  <ChevronRight className="h-4 w-4 text-gray-400" />
                </button>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Files */}
      {files.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Files
              <span className="text-sm font-normal text-gray-500">({files.length})</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {files.map((file) => (
                <div
                  key={file.id}
                  className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-800"
                >
                  <FileText className="h-5 w-5 text-gray-400" />
                  <div className="flex-1 min-w-0">
                    <span className="text-sm text-gray-900 dark:text-white">
                      File #{file.file_id}
                    </span>
                    {file.quality_info && (
                      <span className="ml-2 text-xs text-gray-500">{file.quality_info}</span>
                    )}
                  </div>
                  {file.language && (
                    <span className="text-xs px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded">
                      {file.language}
                    </span>
                  )}
                  {file.is_primary && (
                    <span className="text-xs px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded">
                      Primary
                    </span>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* External Metadata */}
      {entity.external_metadata && entity.external_metadata.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-5 w-5" />
              External Metadata
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {entity.external_metadata.map((meta) => (
                <div
                  key={meta.id}
                  className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-800"
                >
                  <div className="flex-1">
                    <span className="font-medium text-gray-900 dark:text-white capitalize">
                      {meta.provider}
                    </span>
                    <span className="text-sm text-gray-500 ml-2">ID: {meta.external_id}</span>
                  </div>
                  {meta.rating != null && (
                    <span className="flex items-center gap-1 text-sm">
                      <Star className="h-3 w-3 text-yellow-500" />
                      {meta.rating}
                    </span>
                  )}
                  {meta.review_url && (
                    <a
                      href={meta.review_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-blue-600 hover:underline"
                    >
                      View
                    </a>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Duplicates */}
      {duplicates.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-orange-600">
              <Copy className="h-5 w-5" />
              Potential Duplicates
              <span className="text-sm font-normal text-gray-500">({duplicates.length})</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {duplicates.map((dup) => (
                <button
                  key={dup.id}
                  onClick={() => navigate(`/entity/${dup.id}`)}
                  className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors text-left"
                >
                  <div className="flex-1 min-w-0">
                    <span className="font-medium text-gray-900 dark:text-white">
                      {dup.title}
                    </span>
                    {dup.year && (
                      <span className="text-sm text-gray-500 ml-2">({dup.year})</span>
                    )}
                  </div>
                  <span className="text-xs text-gray-500 capitalize">
                    {dup.status}
                  </span>
                  <ChevronRight className="h-4 w-4 text-gray-400" />
                </button>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

export default EntityDetail
