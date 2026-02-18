import { Card, CardContent } from '@/components/ui/Card'
import {
  Film, Tv, Music, Gamepad2, Monitor, BookOpen, Book,
} from 'lucide-react'
import { motion } from 'framer-motion'
import type { MediaTypeInfo } from '@/types/media'

export const TYPE_ICONS: Record<string, React.ElementType> = {
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

export const TYPE_COLORS: Record<string, string> = {
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

export function TypeSelectorGrid({
  types,
  onSelect,
}: {
  types: MediaTypeInfo[]
  onSelect: (type: string) => void
}) {
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
