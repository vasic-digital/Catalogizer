import { useState, useEffect } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { entityApi } from '@/lib/mediaApi'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import {
  Film, Tv, Music, Gamepad2, Monitor, BookOpen, Book,
  Search, ChevronLeft, ChevronRight,
  ArrowLeft,
} from 'lucide-react'
import { motion } from 'framer-motion'
import type { MediaEntity, MediaTypeInfo } from '@/types/media'

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

const TYPE_COLORS: Record<string, string> = {
  movie: 'from-blue-500 to-blue-700',
  tv_show: 'from-purple-500 to-purple-700',
  music_artist: 'from-green-500 to-green-700',
  music_album: 'from-green-400 to-green-600',
  song: 'from-green-300 to-green-500',
  game: 'from-red-500 to-red-700',
  software: 'from-cyan-500 to-cyan-700',
  book: 'from-amber-500 to-amber-700',
  comic: 'from-pink-500 to-pink-700',
}

function TypeSelectorGrid({
  types,
  onSelect,
}: {
  types: MediaTypeInfo[]
  onSelect: (type: string) => void
}) {
  // Only show types that have count or are top-level browsable
  const browsableTypes = types.filter(
    (t) => !['tv_season', 'tv_episode', 'song'].includes(t.name)
  )

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      {browsableTypes.map((type, i) => {
        const Icon = TYPE_ICONS[type.name] || Film
        const gradient = TYPE_COLORS[type.name] || 'from-gray-500 to-gray-700'

        return (
          <motion.div
            key={type.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.05 }}
          >
            <button
              onClick={() => onSelect(type.name)}
              className="w-full text-left"
            >
              <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                <CardContent className="p-6">
                  <div
                    className={`w-12 h-12 rounded-lg bg-gradient-to-br ${gradient} flex items-center justify-center mb-3`}
                  >
                    <Icon className="h-6 w-6 text-white" />
                  </div>
                  <h3 className="font-semibold text-gray-900 dark:text-white capitalize">
                    {type.name.replace(/_/g, ' ')}
                  </h3>
                  <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                    {type.count} {type.count === 1 ? 'item' : 'items'}
                  </p>
                </CardContent>
              </Card>
            </button>
          </motion.div>
        )
      })}
    </div>
  )
}

function EntityCard({
  entity,
  onClick,
}: {
  entity: MediaEntity
  onClick: () => void
}) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      whileHover={{ scale: 1.02 }}
      transition={{ duration: 0.15 }}
    >
      <Card
        className="cursor-pointer hover:shadow-lg transition-shadow h-full"
        onClick={onClick}
      >
        <CardContent className="p-4">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded bg-gray-100 dark:bg-gray-800 flex items-center justify-center flex-shrink-0">
              {(() => {
                const Icon = TYPE_ICONS[entity.status] || Film
                return <Icon className="h-5 w-5 text-gray-500" />
              })()}
            </div>
            <div className="min-w-0 flex-1">
              <h3 className="font-medium text-gray-900 dark:text-white truncate">
                {entity.title}
              </h3>
              <div className="flex items-center gap-2 mt-1 text-sm text-gray-500 dark:text-gray-400">
                {entity.year && <span>{entity.year}</span>}
                {entity.rating != null && (
                  <span className="flex items-center gap-0.5">
                    {entity.rating.toFixed(1)}
                  </span>
                )}
              </div>
              {entity.genre && entity.genre.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {entity.genre.slice(0, 3).map((g) => (
                    <span
                      key={g}
                      className="px-2 py-0.5 text-xs bg-gray-100 dark:bg-gray-800 rounded-full text-gray-600 dark:text-gray-400"
                    >
                      {g}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )
}

export function EntityBrowser() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const selectedType = searchParams.get('type') || ''
  const searchQuery = searchParams.get('q') || ''
  const page = parseInt(searchParams.get('page') || '1', 10)
  const limit = 24
  const offset = (page - 1) * limit

  const { data: typesData } = useQuery({
    queryKey: ['entityTypes'],
    queryFn: () => entityApi.getEntityTypes(),
  })

  const { data: statsData } = useQuery({
    queryKey: ['entityStats'],
    queryFn: () => entityApi.getEntityStats(),
  })

  const {
    data: entitiesData,
    isLoading: entitiesLoading,
  } = useQuery({
    queryKey: ['entities', selectedType, searchQuery, page],
    queryFn: () =>
      selectedType
        ? entityApi.browseByType(selectedType, { limit, offset })
        : entityApi.getEntities({ query: searchQuery || undefined, limit, offset }),
    enabled: !!selectedType || !!searchQuery,
  })

  const types = typesData?.types || []
  const entities = entitiesData?.items || []
  const total = entitiesData?.total || 0
  const totalPages = Math.ceil(total / limit)

  const handleTypeSelect = (type: string) => {
    setSearchParams({ type })
  }

  const handleSearch = (q: string) => {
    if (q) {
      setSearchParams({ q })
    } else {
      setSearchParams({})
    }
  }

  const handleBack = () => {
    setSearchParams({})
  }

  const handleEntityClick = (entity: MediaEntity) => {
    navigate(`/entity/${entity.id}`)
  }

  const handlePageChange = (newPage: number) => {
    const params: Record<string, string> = {}
    if (selectedType) params.type = selectedType
    if (searchQuery) params.q = searchQuery
    params.page = String(newPage)
    setSearchParams(params)
  }

  const showTypeSelector = !selectedType && !searchQuery

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {!showTypeSelector && (
            <Button variant="ghost" size="icon" onClick={handleBack}>
              <ArrowLeft className="h-5 w-5" />
            </Button>
          )}
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              {showTypeSelector
                ? 'Browse Media'
                : selectedType
                  ? selectedType.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()) + 's'
                  : `Search: "${searchQuery}"`}
            </h1>
            {statsData && showTypeSelector && (
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                {statsData.total_entities} total entities across all types
              </p>
            )}
          </div>
        </div>

        {/* Search */}
        <div className="relative w-64">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search entities..."
            defaultValue={searchQuery}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                handleSearch((e.target as HTMLInputElement).value)
              }
            }}
            className="w-full pl-10 pr-4 py-2 bg-gray-100 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-white"
          />
        </div>
      </div>

      {/* Type Selector */}
      {showTypeSelector && <TypeSelectorGrid types={types} onSelect={handleTypeSelect} />}

      {/* Entity List */}
      {!showTypeSelector && (
        <>
          {entitiesLoading ? (
            <div className="flex items-center justify-center min-h-[200px]">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
            </div>
          ) : entities.length === 0 ? (
            <Card>
              <CardContent className="p-12 text-center">
                <p className="text-gray-500 dark:text-gray-400">No entities found</p>
              </CardContent>
            </Card>
          ) : (
            <>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Showing {offset + 1}-{Math.min(offset + limit, total)} of {total}
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {entities.map((entity) => (
                  <EntityCard
                    key={entity.id}
                    entity={entity}
                    onClick={() => handleEntityClick(entity)}
                  />
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-2 mt-6">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page <= 1}
                    onClick={() => handlePageChange(page - 1)}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <span className="text-sm text-gray-600 dark:text-gray-400">
                    Page {page} of {totalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page >= totalPages}
                    onClick={() => handlePageChange(page + 1)}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              )}
            </>
          )}
        </>
      )}
    </div>
  )
}

export default EntityBrowser
