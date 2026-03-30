import type {
  Collection,
  SmartCollection,
  CollectionRule,
  CollectionTemplate,
  CollectionTemplateRule,
  CollectionAnalytics,
  CreateCollectionRequest,
  UpdateCollectionRequest,
  ShareCollectionRequest,
  CollectionShareInfo,
} from '../collections'

describe('collections types', () => {
  describe('Collection interface', () => {
    it('has all required fields', () => {
      const collection: Collection = {
        id: 'col-1',
        name: 'My Movies',
        item_count: 50,
        is_public: false,
        is_smart: false,
        primary_media_type: 'video',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
        owner_id: 'user-1',
      }
      expect(collection.id).toBe('col-1')
      expect(collection.name).toBe('My Movies')
      expect(collection.item_count).toBe(50)
      expect(collection.is_public).toBe(false)
      expect(collection.description).toBeUndefined()
      expect(collection.thumbnail_url).toBeUndefined()
      expect(collection.cover_image).toBeUndefined()
    })

    it('accepts all primary_media_type values', () => {
      const types: Collection['primary_media_type'][] = ['music', 'video', 'image', 'document', 'mixed']
      expect(types).toHaveLength(5)
    })

    it('accepts optional fields', () => {
      const collection: Collection = {
        id: 'col-2',
        name: 'Favorites',
        description: 'My favorite media',
        item_count: 10,
        is_public: true,
        is_smart: true,
        primary_media_type: 'mixed',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
        thumbnail_url: '/thumbs/fav.jpg',
        cover_image: '/covers/fav.jpg',
        owner_id: 'user-1',
      }
      expect(collection.description).toBe('My favorite media')
      expect(collection.thumbnail_url).toBe('/thumbs/fav.jpg')
      expect(collection.cover_image).toBe('/covers/fav.jpg')
    })
  })

  describe('SmartCollection interface', () => {
    it('has is_smart always true', () => {
      const smart: SmartCollection = {
        id: 'sc-1',
        name: 'Recent Movies',
        is_smart: true,
        smart_rules: [],
        item_count: 25,
        last_updated: '2024-06-01T00:00:00Z',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
        owner_id: 'user-1',
        is_public: false,
        primary_media_type: 'video',
      }
      expect(smart.is_smart).toBe(true)
      expect(smart.smart_rules).toEqual([])
    })
  })

  describe('CollectionRule interface', () => {
    it('has all required fields', () => {
      const rule: CollectionRule = {
        id: 'rule-1',
        field: 'media_type',
        operator: 'equals',
        value: 'movie',
        field_type: 'select',
        label: 'Media Type',
      }
      expect(rule.id).toBe('rule-1')
      expect(rule.field).toBe('media_type')
      expect(rule.operator).toBe('equals')
      expect(rule.value).toBe('movie')
      expect(rule.condition).toBeUndefined()
      expect(rule.nested_rules).toBeUndefined()
    })

    it('accepts nested rules', () => {
      const rule: CollectionRule = {
        id: 'rule-parent',
        field: 'media_type',
        operator: 'equals',
        value: 'movie',
        condition: 'AND',
        field_type: 'select',
        label: 'Media Type',
        nested_rules: [
          {
            id: 'rule-child',
            field: 'year',
            operator: 'greater_than',
            value: 2020,
            field_type: 'number',
            label: 'Year',
          },
        ],
      }
      expect(rule.nested_rules).toHaveLength(1)
      expect(rule.condition).toBe('AND')
    })

    it('accepts all field_type values', () => {
      const fieldTypes: CollectionRule['field_type'][] = [
        'text', 'number', 'date', 'boolean', 'select', 'multiselect',
      ]
      expect(fieldTypes).toHaveLength(6)
    })

    it('accepts both AND and OR conditions', () => {
      const andRule: CollectionRule = {
        id: '1', field: 'f', operator: 'eq', value: 'v',
        condition: 'AND', field_type: 'text', label: 'F',
      }
      const orRule: CollectionRule = {
        id: '2', field: 'f', operator: 'eq', value: 'v',
        condition: 'OR', field_type: 'text', label: 'F',
      }
      expect(andRule.condition).toBe('AND')
      expect(orRule.condition).toBe('OR')
    })
  })

  describe('CollectionTemplate interface', () => {
    it('has all required fields', () => {
      const template: CollectionTemplate = {
        id: 'tpl-1',
        name: 'Recent Movies',
        description: 'Movies added in the last 30 days',
        category: 'Movies',
        rules: [
          { field: 'media_type', operator: 'equals', value: 'movie', field_type: 'select', label: 'Type' },
        ],
      }
      expect(template.id).toBe('tpl-1')
      expect(template.rules).toHaveLength(1)
      expect(template.icon).toBeUndefined()
    })

    it('accepts optional icon', () => {
      const template: CollectionTemplate = {
        id: 'tpl-2',
        name: 'All Music',
        description: 'All music items',
        category: 'Music',
        rules: [],
        icon: 'music',
      }
      expect(template.icon).toBe('music')
    })
  })

  describe('CollectionTemplateRule type', () => {
    it('is CollectionRule without id', () => {
      const rule: CollectionTemplateRule = {
        field: 'year',
        operator: 'greater_than',
        value: 2020,
        field_type: 'number',
        label: 'Year',
      }
      expect(rule.field).toBe('year')
      expect((rule as Record<string, unknown>)['id']).toBeUndefined()
    })
  })

  describe('CollectionAnalytics interface', () => {
    it('has all required fields', () => {
      const analytics: CollectionAnalytics = {
        collection_id: 'col-1',
        total_items: 100,
        media_type_distribution: { music: 30, video: 50, image: 15, document: 5 },
        size_distribution: {
          total_size_bytes: 50000000000,
          average_size_bytes: 500000000,
          largest_item_size_bytes: 5000000000,
        },
        quality_distribution: { hd: 60, sd: 30, uhd: 10 },
        time_based_stats: {
          items_added_today: 2,
          items_added_this_week: 10,
          items_added_this_month: 30,
          oldest_item_date: '2020-01-01',
          newest_item_date: '2024-06-01',
        },
        engagement_stats: {
          total_views: 5000,
          unique_viewers: 150,
          total_plays: 3000,
          average_completion_rate: 0.85,
          last_accessed: '2024-06-01T12:00:00Z',
        },
      }
      expect(analytics.total_items).toBe(100)
      expect(analytics.media_type_distribution.video).toBe(50)
      expect(analytics.genre_distribution).toBeUndefined()
      expect(analytics.artist_distribution).toBeUndefined()
      expect(analytics.decade_distribution).toBeUndefined()
    })

    it('accepts optional distribution fields', () => {
      const analytics: CollectionAnalytics = {
        collection_id: 'col-2',
        total_items: 50,
        media_type_distribution: { music: 50, video: 0, image: 0, document: 0 },
        size_distribution: { total_size_bytes: 0, average_size_bytes: 0, largest_item_size_bytes: 0 },
        quality_distribution: { hd: 0, sd: 0, uhd: 0 },
        time_based_stats: {
          items_added_today: 0, items_added_this_week: 0, items_added_this_month: 0,
          oldest_item_date: '', newest_item_date: '',
        },
        genre_distribution: { rock: 20, jazz: 15, classical: 15 },
        artist_distribution: { 'Artist A': 10, 'Artist B': 5 },
        decade_distribution: { '1990s': 25, '2000s': 25 },
        engagement_stats: {
          total_views: 0, unique_viewers: 0, total_plays: 0,
          average_completion_rate: 0, last_accessed: '',
        },
      }
      expect(analytics.genre_distribution).toEqual({ rock: 20, jazz: 15, classical: 15 })
      expect(analytics.artist_distribution).toBeDefined()
      expect(analytics.decade_distribution).toBeDefined()
    })
  })

  describe('CreateCollectionRequest interface', () => {
    it('has required fields', () => {
      const request: CreateCollectionRequest = {
        name: 'New Collection',
        is_smart: true,
        smart_rules: [],
      }
      expect(request.name).toBe('New Collection')
      expect(request.is_smart).toBe(true)
      expect(request.description).toBeUndefined()
      expect(request.is_public).toBeUndefined()
    })
  })

  describe('UpdateCollectionRequest interface', () => {
    it('has all optional fields', () => {
      const request: UpdateCollectionRequest = {}
      expect(request.name).toBeUndefined()
      expect(request.description).toBeUndefined()
      expect(request.is_public).toBeUndefined()
      expect(request.smart_rules).toBeUndefined()
    })
  })

  describe('ShareCollectionRequest interface', () => {
    it('has required fields', () => {
      const request: ShareCollectionRequest = {
        can_view: true,
        can_comment: false,
        can_download: true,
      }
      expect(request.can_view).toBe(true)
      expect(request.can_comment).toBe(false)
      expect(request.expires_at).toBeUndefined()
      expect(request.allow_reshare).toBeUndefined()
    })
  })

  describe('CollectionShareInfo interface', () => {
    it('has required fields', () => {
      const info: CollectionShareInfo = {
        share_url: 'https://example.com/share/abc',
        share_id: 'share-1',
        permissions: {
          can_view: true,
          can_comment: false,
          can_download: true,
        },
        created_at: '2024-06-01T00:00:00Z',
        access_count: 42,
      }
      expect(info.share_url).toContain('share')
      expect(info.access_count).toBe(42)
      expect(info.expires_at).toBeUndefined()
    })
  })
})
