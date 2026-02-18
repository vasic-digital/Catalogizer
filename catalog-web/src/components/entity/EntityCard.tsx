import { Card, CardContent } from '@/components/ui/Card'
import { Film } from 'lucide-react'
import { motion } from 'framer-motion'
import { TYPE_ICONS } from './TypeSelector'
import type { MediaEntity } from '@/types/media'

export function EntityCard({
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
