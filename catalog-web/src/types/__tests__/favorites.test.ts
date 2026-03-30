import type {
  Favorite,
  FavoriteToggleRequest,
  FavoritesResponse,
  FavoriteStats,
} from '../favorites'

describe('favorites types', () => {
  describe('Favorite interface', () => {
    it('has all required fields', () => {
      const favorite: Favorite = {
        id: 'fav-1',
        user_id: 1,
        media_id: 42,
        media_item: {
          id: 42,
          title: 'The Matrix',
          media_type: 'movie',
        },
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      expect(favorite.id).toBe('fav-1')
      expect(favorite.user_id).toBe(1)
      expect(favorite.media_id).toBe(42)
      expect(favorite.media_item.title).toBe('The Matrix')
      expect(favorite.media_item.year).toBeUndefined()
      expect(favorite.media_item.cover_image).toBeUndefined()
      expect(favorite.media_item.duration).toBeUndefined()
      expect(favorite.media_item.rating).toBeUndefined()
      expect(favorite.media_item.quality).toBeUndefined()
    })

    it('accepts media_item with all optional fields', () => {
      const favorite: Favorite = {
        id: 'fav-2',
        user_id: 1,
        media_id: 99,
        media_item: {
          id: 99,
          title: 'Inception',
          media_type: 'movie',
          year: 2010,
          cover_image: '/covers/inception.jpg',
          duration: 148,
          rating: 8.8,
          quality: '4k',
        },
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
      }
      expect(favorite.media_item.year).toBe(2010)
      expect(favorite.media_item.cover_image).toBe('/covers/inception.jpg')
      expect(favorite.media_item.duration).toBe(148)
      expect(favorite.media_item.rating).toBe(8.8)
      expect(favorite.media_item.quality).toBe('4k')
    })
  })

  describe('FavoriteToggleRequest interface', () => {
    it('has all required fields', () => {
      const request: FavoriteToggleRequest = {
        media_id: 42,
        is_favorite: true,
      }
      expect(request.media_id).toBe(42)
      expect(request.is_favorite).toBe(true)
    })

    it('can toggle off', () => {
      const request: FavoriteToggleRequest = {
        media_id: 42,
        is_favorite: false,
      }
      expect(request.is_favorite).toBe(false)
    })
  })

  describe('FavoritesResponse interface', () => {
    it('has all required fields', () => {
      const response: FavoritesResponse = {
        items: [],
        total: 0,
        limit: 20,
        offset: 0,
      }
      expect(response.items).toEqual([])
      expect(response.total).toBe(0)
      expect(response.limit).toBe(20)
      expect(response.offset).toBe(0)
    })

    it('accepts items array with Favorite objects', () => {
      const response: FavoritesResponse = {
        items: [
          {
            id: 'fav-1',
            user_id: 1,
            media_id: 42,
            media_item: { id: 42, title: 'The Matrix', media_type: 'movie' },
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        ],
        total: 1,
        limit: 20,
        offset: 0,
      }
      expect(response.items).toHaveLength(1)
      expect(response.total).toBe(1)
    })
  })

  describe('FavoriteStats interface', () => {
    it('has all required fields', () => {
      const stats: FavoriteStats = {
        total_count: 150,
        media_type_breakdown: {
          movie: 50,
          tv_show: 30,
          music: 40,
          game: 10,
          documentary: 8,
          anime: 5,
          concert: 2,
          other: 5,
        },
        recent_additions: [],
      }
      expect(stats.total_count).toBe(150)
      expect(stats.media_type_breakdown.movie).toBe(50)
      expect(stats.media_type_breakdown.tv_show).toBe(30)
      expect(stats.media_type_breakdown.music).toBe(40)
      expect(stats.media_type_breakdown.game).toBe(10)
      expect(stats.media_type_breakdown.documentary).toBe(8)
      expect(stats.media_type_breakdown.anime).toBe(5)
      expect(stats.media_type_breakdown.concert).toBe(2)
      expect(stats.media_type_breakdown.other).toBe(5)
      expect(stats.recent_additions).toEqual([])
    })

    it('accepts recent_additions with Favorite objects', () => {
      const stats: FavoriteStats = {
        total_count: 1,
        media_type_breakdown: {
          movie: 1, tv_show: 0, music: 0, game: 0,
          documentary: 0, anime: 0, concert: 0, other: 0,
        },
        recent_additions: [
          {
            id: 'fav-1',
            user_id: 1,
            media_id: 42,
            media_item: { id: 42, title: 'Test', media_type: 'movie' },
            created_at: '2024-06-01T00:00:00Z',
            updated_at: '2024-06-01T00:00:00Z',
          },
        ],
      }
      expect(stats.recent_additions).toHaveLength(1)
    })
  })
})
