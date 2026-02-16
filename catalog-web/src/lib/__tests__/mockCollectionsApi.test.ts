import { mockCollectionsApi, shouldUseMockCollections } from '../mockCollectionsApi'

describe('mockCollectionsApi', () => {
  describe('getCollections', () => {
    it('returns an array of collections', async () => {
      const collections = await mockCollectionsApi.getCollections()
      expect(Array.isArray(collections)).toBe(true)
      expect(collections.length).toBeGreaterThan(0)
    })

    it('returns collections with required fields', async () => {
      const collections = await mockCollectionsApi.getCollections()
      const collection = collections[0]

      expect(collection).toHaveProperty('id')
      expect(collection).toHaveProperty('name')
      expect(collection).toHaveProperty('item_count')
      expect(collection).toHaveProperty('is_public')
      expect(collection).toHaveProperty('is_smart')
      expect(collection).toHaveProperty('primary_media_type')
      expect(collection).toHaveProperty('created_at')
      expect(collection).toHaveProperty('updated_at')
      expect(collection).toHaveProperty('owner_id')
    })

    it('includes My Music Collection in mock data', async () => {
      const collections = await mockCollectionsApi.getCollections()
      const musicCollection = collections.find(c => c.name === 'My Music Collection')
      expect(musicCollection).toBeDefined()
      expect(musicCollection?.item_count).toBe(1250)
      expect(musicCollection?.primary_media_type).toBe('music')
    })
  })

  describe('getCollection', () => {
    it('returns a single collection by ID', async () => {
      const collection = await mockCollectionsApi.getCollection('1')
      expect(collection.id).toBe('1')
      expect(collection.name).toBe('My Music Collection')
    })

    it('throws error for non-existent collection', async () => {
      await expect(mockCollectionsApi.getCollection('nonexistent')).rejects.toThrow(
        'Collection not found'
      )
    })
  })

  describe('createCollection', () => {
    it('creates a new collection with provided data', async () => {
      const newCollection = await mockCollectionsApi.createCollection({
        name: 'Test Collection',
        description: 'A test',
        items: ['a', 'b', 'c'],
      })

      expect(newCollection.name).toBe('Test Collection')
      expect(newCollection.description).toBe('A test')
      expect(newCollection.item_count).toBe(3)
      expect(newCollection.is_public).toBe(false)
      expect(newCollection.is_smart).toBe(false)
      expect(newCollection.primary_media_type).toBe('mixed')
      expect(newCollection.id).toBeDefined()
      expect(newCollection.created_at).toBeDefined()
      expect(newCollection.updated_at).toBeDefined()
    })

    it('creates collection with zero items when no items provided', async () => {
      const newCollection = await mockCollectionsApi.createCollection({
        name: 'Empty Collection',
      })

      expect(newCollection.item_count).toBe(0)
    })
  })

  describe('updateCollection', () => {
    it('updates an existing collection', async () => {
      const updated = await mockCollectionsApi.updateCollection('1', {
        name: 'Updated Music',
      })

      expect(updated.id).toBe('1')
      expect(updated.name).toBe('Updated Music')
    })

    it('throws error for non-existent collection', async () => {
      await expect(
        mockCollectionsApi.updateCollection('nonexistent', { name: 'test' })
      ).rejects.toThrow('Collection not found')
    })

    it('updates the updated_at timestamp', async () => {
      const before = new Date().toISOString()
      const updated = await mockCollectionsApi.updateCollection('2', {
        name: 'Updated Name',
      })

      expect(new Date(updated.updated_at).getTime()).toBeGreaterThanOrEqual(
        new Date(before).getTime() - 1000
      )
    })
  })

  describe('deleteCollection', () => {
    it('does not throw for existing collection', async () => {
      // First create one so we can safely delete it
      const created = await mockCollectionsApi.createCollection({
        name: 'To Delete',
      })

      await expect(
        mockCollectionsApi.deleteCollection(created.id)
      ).resolves.not.toThrow()
    })

    it('throws error for non-existent collection', async () => {
      await expect(
        mockCollectionsApi.deleteCollection('nonexistent')
      ).rejects.toThrow('Collection not found')
    })
  })

  describe('getSmartCollections', () => {
    it('returns an array of smart collections', async () => {
      const smartCollections = await mockCollectionsApi.getSmartCollections()
      expect(Array.isArray(smartCollections)).toBe(true)
      expect(smartCollections.length).toBeGreaterThan(0)
    })

    it('all returned collections have is_smart = true', async () => {
      const smartCollections = await mockCollectionsApi.getSmartCollections()
      smartCollections.forEach(sc => {
        expect(sc.is_smart).toBe(true)
      })
    })

    it('smart collections have smart_rules', async () => {
      const smartCollections = await mockCollectionsApi.getSmartCollections()
      smartCollections.forEach(sc => {
        expect(Array.isArray(sc.smart_rules)).toBe(true)
        expect(sc.smart_rules.length).toBeGreaterThan(0)
      })
    })
  })

  describe('getSmartCollection', () => {
    it('returns a single smart collection by ID', async () => {
      const sc = await mockCollectionsApi.getSmartCollection('3')
      expect(sc.id).toBe('3')
      expect(sc.name).toBe('Recently Added')
      expect(sc.is_smart).toBe(true)
    })

    it('throws error for non-existent smart collection', async () => {
      await expect(
        mockCollectionsApi.getSmartCollection('nonexistent')
      ).rejects.toThrow('Smart Collection not found')
    })
  })

  describe('createSmartCollection', () => {
    it('creates a smart collection with rules', async () => {
      const rules = [
        {
          field: 'genre',
          operator: 'equals',
          value: 'rock',
          field_type: 'select' as const,
          label: 'Genre',
        },
      ]

      const created = await mockCollectionsApi.createSmartCollection({
        name: 'Rock Songs',
        description: 'All rock music',
        rules: rules as any,
      })

      expect(created.name).toBe('Rock Songs')
      expect(created.is_smart).toBe(true)
      expect(created.smart_rules).toEqual(rules)
      expect(created.item_count).toBeGreaterThanOrEqual(10)
    })
  })

  describe('updateSmartCollection', () => {
    it('updates a smart collection', async () => {
      const updated = await mockCollectionsApi.updateSmartCollection('3', {
        name: 'Updated Recently Added',
      })

      expect(updated.id).toBe('3')
      expect(updated.name).toBe('Updated Recently Added')
    })

    it('throws for non-existent smart collection', async () => {
      await expect(
        mockCollectionsApi.updateSmartCollection('nonexistent', { name: 'test' })
      ).rejects.toThrow('Smart Collection not found')
    })
  })

  describe('deleteSmartCollection', () => {
    it('throws for non-existent smart collection', async () => {
      await expect(
        mockCollectionsApi.deleteSmartCollection('nonexistent')
      ).rejects.toThrow('Smart Collection not found')
    })
  })

  describe('getTemplates', () => {
    it('returns an array of templates', async () => {
      const templates = await mockCollectionsApi.getTemplates()
      expect(Array.isArray(templates)).toBe(true)
      expect(templates.length).toBeGreaterThan(0)
    })

    it('templates have required fields', async () => {
      const templates = await mockCollectionsApi.getTemplates()
      templates.forEach(t => {
        expect(t).toHaveProperty('id')
        expect(t).toHaveProperty('name')
        expect(t).toHaveProperty('description')
        expect(t).toHaveProperty('category')
        expect(t).toHaveProperty('rules')
      })
    })

    it('includes Recently Added template', async () => {
      const templates = await mockCollectionsApi.getTemplates()
      const recentTemplate = templates.find(t => t.id === 'recently_added')
      expect(recentTemplate).toBeDefined()
      expect(recentTemplate?.name).toBe('Recently Added')
    })
  })

  describe('getAnalytics', () => {
    it('returns analytics for a collection', async () => {
      const analytics = await mockCollectionsApi.getAnalytics('1')
      expect(analytics.collection_id).toBe('1')
      expect(analytics.total_items).toBe(1250)
      expect(analytics).toHaveProperty('media_type_distribution')
      expect(analytics).toHaveProperty('size_distribution')
      expect(analytics).toHaveProperty('quality_distribution')
      expect(analytics).toHaveProperty('time_based_stats')
      expect(analytics).toHaveProperty('engagement_stats')
      expect(analytics).toHaveProperty('genre_distribution')
    })

    it('throws for non-existent collection analytics', async () => {
      await expect(
        mockCollectionsApi.getAnalytics('nonexistent')
      ).rejects.toThrow('Analytics not found')
    })
  })

  describe('testRules', () => {
    it('returns validation result for rules', async () => {
      const result = await mockCollectionsApi.testRules([
        {
          field: 'genre',
          operator: 'equals',
          value: 'rock',
          field_type: 'select',
          label: 'Genre',
        },
      ] as any)

      expect(result.valid).toBe(true)
      expect(result.sample_count).toBeGreaterThanOrEqual(10)
    })
  })
})

describe('shouldUseMockCollections', () => {
  it('returns true (mock mode is always on)', () => {
    expect(shouldUseMockCollections()).toBe(true)
  })
})
