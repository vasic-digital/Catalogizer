import { collectionsApi } from '../collectionsApi'

// Mock the api and mockCollectionsApi modules
vi.mock('../api', async () => {
  const mockApi = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  }
  return {
    __esModule: true,
    default: mockApi,
    api: mockApi,
  }
})

vi.mock('../mockCollectionsApi', () => ({
  shouldUseMockCollections: vi.fn(() => false),
  mockCollectionsApi: {
    getSmartCollections: vi.fn(),
    getSmartCollection: vi.fn(),
    createSmartCollection: vi.fn(),
    updateSmartCollection: vi.fn(),
    deleteSmartCollection: vi.fn(),
    getAnalytics: vi.fn(),
    getTemplates: vi.fn(),
    testRules: vi.fn(),
  },
}))

import { api } from '../api'
const mockApi = vi.mocked(api)

describe('collectionsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getCollections', () => {
    it('calls GET /api/collections and returns collections array', async () => {
      const mockCollections = [
        { id: '1', name: 'Test Collection', item_count: 5 },
        { id: '2', name: 'Second Collection', item_count: 10 },
      ]
      mockApi.get.mockResolvedValue({ data: mockCollections })

      const result = await collectionsApi.getCollections()

      expect(mockApi.get).toHaveBeenCalledWith('/api/collections')
      expect(result).toEqual(mockCollections)
    })

    it('propagates errors on failure', async () => {
      const error = new Error('Server error')
      ;(error as any).response = { status: 500 }
      mockApi.get.mockRejectedValue(error)

      await expect(collectionsApi.getCollections()).rejects.toThrow('Server error')
    })
  })

  describe('getCollection', () => {
    it('calls GET /api/collections/:id and returns single collection', async () => {
      const mockCollection = { id: '1', name: 'Test Collection', item_count: 5 }
      mockApi.get.mockResolvedValue({ data: mockCollection })

      const result = await collectionsApi.getCollection('1')

      expect(mockApi.get).toHaveBeenCalledWith('/api/collections/1')
      expect(result).toEqual(mockCollection)
    })
  })

  describe('createCollection', () => {
    it('calls POST /api/collections with collection data', async () => {
      const newCollection = {
        name: 'New Collection',
        description: 'A test collection',
        is_public: false,
        is_smart: true as const,
        smart_rules: [],
      }
      const created = { ...newCollection, id: '3', item_count: 0 }
      mockApi.post.mockResolvedValue({ data: created })

      const result = await collectionsApi.createCollection(newCollection)

      expect(mockApi.post).toHaveBeenCalledWith('/api/collections', newCollection)
      expect(result).toEqual(created)
    })
  })

  describe('updateCollection', () => {
    it('calls PUT /api/collections/:id with updates', async () => {
      const updates = { name: 'Updated Name' }
      const updated = { id: '1', name: 'Updated Name', item_count: 5 }
      mockApi.put.mockResolvedValue({ data: updated })

      const result = await collectionsApi.updateCollection('1', updates)

      expect(mockApi.put).toHaveBeenCalledWith('/api/collections/1', updates)
      expect(result).toEqual(updated)
    })
  })

  describe('deleteCollection', () => {
    it('calls DELETE /api/collections/:id', async () => {
      mockApi.delete.mockResolvedValue({})

      await collectionsApi.deleteCollection('1')

      expect(mockApi.delete).toHaveBeenCalledWith('/api/collections/1')
    })
  })

  describe('getCollectionItems', () => {
    it('calls GET /api/collections/:id/items with pagination params', async () => {
      const mockItems = {
        items: [{ id: 'item1', title: 'Item 1' }],
        total: 1,
        page: 1,
        limit: 50,
      }
      mockApi.get.mockResolvedValue({ data: mockItems })

      const result = await collectionsApi.getCollectionItems('1', 1, 50)

      expect(mockApi.get).toHaveBeenCalledWith('/api/collections/1/items', {
        params: { page: 1, limit: 50 },
      })
      expect(result).toEqual(mockItems)
    })

    it('uses default pagination parameters', async () => {
      mockApi.get.mockResolvedValue({ data: { items: [], total: 0 } })

      await collectionsApi.getCollectionItems('1')

      expect(mockApi.get).toHaveBeenCalledWith('/api/collections/1/items', {
        params: { page: 1, limit: 50 },
      })
    })
  })

  describe('refreshCollection', () => {
    it('calls POST /api/collections/:id/refresh', async () => {
      const refreshed = { id: '1', name: 'Test', item_count: 15, last_updated: '2024-01-01' }
      mockApi.post.mockResolvedValue({ data: refreshed })

      const result = await collectionsApi.refreshCollection('1')

      expect(mockApi.post).toHaveBeenCalledWith('/api/collections/1/refresh')
      expect(result).toEqual(refreshed)
    })
  })

  describe('getCollectionAnalytics', () => {
    it('calls GET /api/collections/:id/analytics', async () => {
      const mockAnalytics = {
        collection_id: '1',
        total_items: 100,
        media_type_distribution: { music: 50, video: 30, image: 20, document: 0 },
      }
      mockApi.get.mockResolvedValue({ data: mockAnalytics })

      const result = await collectionsApi.getCollectionAnalytics('1')

      expect(mockApi.get).toHaveBeenCalledWith('/api/collections/1/analytics')
      expect(result).toEqual(mockAnalytics)
    })
  })

  describe('shareCollection', () => {
    it('calls POST /api/collections/:id/share with share request', async () => {
      const shareRequest = { can_view: true, can_comment: false, can_download: true }
      const shareInfo = {
        share_id: 'share_1',
        share_url: 'http://localhost/shared/share_1',
        permissions: shareRequest,
      }
      mockApi.post.mockResolvedValue({ data: shareInfo })

      const result = await collectionsApi.shareCollection('1', shareRequest)

      expect(mockApi.post).toHaveBeenCalledWith('/api/collections/1/share', shareRequest)
      expect(result).toEqual(shareInfo)
    })
  })

  describe('exportCollection', () => {
    it('calls GET /api/collections/:id/export with format param', async () => {
      const blob = new Blob(['test data'], { type: 'application/json' })
      mockApi.get.mockResolvedValue({ data: blob })

      const result = await collectionsApi.exportCollection('1', 'json')

      expect(mockApi.get).toHaveBeenCalledWith('/api/collections/1/export', {
        params: { format: 'json' },
        responseType: 'blob',
      })
      expect(result).toBeInstanceOf(Blob)
    })

    it('defaults to json format', async () => {
      const blob = new Blob(['{}'], { type: 'application/json' })
      mockApi.get.mockResolvedValue({ data: blob })

      await collectionsApi.exportCollection('1')

      expect(mockApi.get).toHaveBeenCalledWith('/api/collections/1/export', {
        params: { format: 'json' },
        responseType: 'blob',
      })
    })
  })

  describe('duplicateCollection', () => {
    it('calls POST /api/collections/:id/duplicate with new name', async () => {
      const duplicated = { id: '4', name: 'Copy of Collection' }
      mockApi.post.mockResolvedValue({ data: duplicated })

      const result = await collectionsApi.duplicateCollection('1', 'Copy of Collection')

      expect(mockApi.post).toHaveBeenCalledWith('/api/collections/1/duplicate', {
        name: 'Copy of Collection',
      })
      expect(result).toEqual(duplicated)
    })
  })

  describe('bulkDeleteCollections', () => {
    it('calls DELETE /api/collections/bulk with collection IDs', async () => {
      mockApi.delete.mockResolvedValue({})

      await collectionsApi.bulkDeleteCollections(['1', '2', '3'])

      expect(mockApi.delete).toHaveBeenCalledWith('/api/collections/bulk', {
        data: { collection_ids: ['1', '2', '3'] },
      })
    })
  })

  describe('testRules', () => {
    it('calls POST /api/collections/test-rules with rules array', async () => {
      const rules = [{ field: 'genre', operator: 'equals', value: 'rock' }]
      const testResult = { valid: true, sample_count: 25 }
      mockApi.post.mockResolvedValue({ data: testResult })

      const result = await collectionsApi.testRules(rules as any)

      expect(mockApi.post).toHaveBeenCalledWith('/api/collections/test-rules', { rules })
      expect(result).toEqual(testResult)
    })
  })
})
