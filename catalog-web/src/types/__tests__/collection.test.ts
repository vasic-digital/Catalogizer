import type { Collection } from '../collection'

describe('collection type (single)', () => {
  describe('Collection interface', () => {
    it('has all required fields', () => {
      const collection: Collection = {
        id: 'col-1',
        name: 'Movies',
        item_count: 100,
        is_public: false,
        is_smart: false,
        primary_media_type: 'video',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
        owner_id: 'user-1',
      }
      expect(collection.id).toBe('col-1')
      expect(collection.name).toBe('Movies')
      expect(collection.item_count).toBe(100)
      expect(collection.is_public).toBe(false)
      expect(collection.is_smart).toBe(false)
      expect(collection.primary_media_type).toBe('video')
      expect(collection.owner_id).toBe('user-1')
      expect(collection.description).toBeUndefined()
      expect(collection.thumbnail_url).toBeUndefined()
      expect(collection.cover_image).toBeUndefined()
    })

    it('accepts optional description', () => {
      const collection: Collection = {
        id: 'col-2',
        name: 'Music',
        description: 'My music collection',
        item_count: 500,
        is_public: true,
        is_smart: true,
        primary_media_type: 'music',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
        owner_id: 'user-1',
      }
      expect(collection.description).toBe('My music collection')
    })

    it('accepts optional thumbnail_url and cover_image', () => {
      const collection: Collection = {
        id: 'col-3',
        name: 'Photos',
        item_count: 1000,
        is_public: false,
        is_smart: false,
        primary_media_type: 'image',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
        thumbnail_url: '/thumbs/photos.jpg',
        cover_image: '/covers/photos.jpg',
        owner_id: 'user-1',
      }
      expect(collection.thumbnail_url).toBe('/thumbs/photos.jpg')
      expect(collection.cover_image).toBe('/covers/photos.jpg')
    })

    it('accepts all primary_media_type values', () => {
      const types: Collection['primary_media_type'][] = [
        'music', 'video', 'image', 'document', 'mixed',
      ]
      expect(types).toHaveLength(5)
      types.forEach(type => {
        expect(typeof type).toBe('string')
      })
    })
  })
})
