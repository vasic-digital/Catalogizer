import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { EntityCard } from './EntityCard'
import type { MediaEntity } from '@/types/media'

export function EntityGrid({
  entities,
  total,
  limit,
  offset,
  page,
  onEntityClick,
  onPageChange,
}: {
  entities: MediaEntity[]
  total: number
  limit: number
  offset: number
  page: number
  onEntityClick: (entity: MediaEntity) => void
  onPageChange: (page: number) => void
}) {
  const totalPages = Math.ceil(total / limit)

  if (entities.length === 0) {
    return (
      <Card>
        <CardContent className="p-12 text-center">
          <p className="text-gray-500 dark:text-gray-400">No entities found</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <>
      <p className="text-sm text-gray-500 dark:text-gray-400">
        Showing {offset + 1}-{Math.min(offset + limit, total)} of {total}
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {entities.map((entity) => (
          <EntityCard
            key={entity.id}
            entity={entity}
            onClick={() => onEntityClick(entity)}
          />
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-6">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
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
            onClick={() => onPageChange(page + 1)}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
    </>
  )
}
