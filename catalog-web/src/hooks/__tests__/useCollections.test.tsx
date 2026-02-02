import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import { useCollections, useCollection, useCollectionAnalytics, useSharedCollection } from '../useCollections'
import { collectionsApi } from '@/lib/collectionsApi'

// Mock collectionsApi
vi.mock('@/lib/collectionsApi', async () => ({
  collectionsApi: {
    getCollections: vi.fn(),
    getCollection: vi.fn(),
    createCollection: vi.fn(),
    updateCollection: vi.fn(),
    deleteCollection: vi.fn(),
    refreshCollection: vi.fn(),
    shareCollection: vi.fn(),
    duplicateCollection: vi.fn(),
    exportCollection: vi.fn(),
    getCollectionItems: vi.fn(),
    getCollectionAnalytics: vi.fn(),
    getSharedCollection: vi.fn(),
    bulkDeleteCollections: vi.fn(),
    bulkShareCollections: vi.fn(),
    bulkExportCollections: vi.fn(),
    bulkUpdateCollections: vi.fn(),
  },
}))

// Mock react-hot-toast
vi.mock('react-hot-toast', async () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

const mockCollectionsApi = vi.mocked(collectionsApi)
const mockToast = require('react-hot-toast').toast

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        cacheTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  })

  const Wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )

  return { Wrapper, queryClient }
}

const mockCollection = {
  id: '1',
  name: 'Test Collection',
  description: 'A test collection',
  smart_rules: [],
  item_count: 10,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  last_updated: '2024-01-01T00:00:00Z',
}

describe('useCollections', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockCollectionsApi.getCollections.mockResolvedValue([mockCollection])
  })

  describe('fetching collections', () => {
    it('fetches collections on mount', async () => {
      const { Wrapper } = createWrapper()

      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      expect(result.current.isLoading).toBe(true)

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.collections).toEqual([mockCollection])
      expect(mockCollectionsApi.getCollections).toHaveBeenCalled()
    })

    it('returns empty array when no collections exist', async () => {
      mockCollectionsApi.getCollections.mockResolvedValue([])
      const { Wrapper } = createWrapper()

      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.collections).toEqual([])
    })

    it('exposes error when fetch fails', async () => {
      mockCollectionsApi.getCollections.mockRejectedValue(new Error('Network error'))
      const { Wrapper } = createWrapper()

      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.error).toBeTruthy()
      })
    })
  })

  describe('createCollection', () => {
    it('calls createCollection API and shows success toast', async () => {
      const newCollection = { ...mockCollection, id: '2', name: 'New Collection' }
      mockCollectionsApi.createCollection.mockResolvedValue(newCollection)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.createCollection({
          collection: { name: 'New Collection', description: 'desc' } as any,
        })
      })

      expect(mockCollectionsApi.createCollection).toHaveBeenCalledWith({
        name: 'New Collection',
        description: 'desc',
      })
      expect(mockToast.success).toHaveBeenCalledWith('Created collection: New Collection')
    })

    it('shows error toast on create failure', async () => {
      mockCollectionsApi.createCollection.mockRejectedValue(new Error('Create failed'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        try {
          await result.current.createCollection({
            collection: { name: 'Fail' } as any,
          })
        } catch (e) {
          // Expected to throw
        }
      })

      expect(mockToast.error).toHaveBeenCalledWith('Create failed')
    })
  })

  describe('updateCollection', () => {
    it('calls updateCollection API and shows success toast', async () => {
      const updated = { ...mockCollection, name: 'Updated Name' }
      mockCollectionsApi.updateCollection.mockResolvedValue(updated)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.updateCollection({
          id: '1',
          updates: { name: 'Updated Name' } as any,
        })
      })

      expect(mockCollectionsApi.updateCollection).toHaveBeenCalledWith('1', { name: 'Updated Name' })
      expect(mockToast.success).toHaveBeenCalledWith('Updated collection: Updated Name')
    })
  })

  describe('deleteCollection', () => {
    it('calls deleteCollection API and shows success toast', async () => {
      mockCollectionsApi.deleteCollection.mockResolvedValue(undefined)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.deleteCollection({ id: '1' })
      })

      expect(mockCollectionsApi.deleteCollection).toHaveBeenCalledWith('1')
      expect(mockToast.success).toHaveBeenCalledWith('Collection deleted successfully')
    })

    it('shows error toast on delete failure', async () => {
      mockCollectionsApi.deleteCollection.mockRejectedValue(new Error('Delete failed'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        try {
          await result.current.deleteCollection({ id: '1' })
        } catch (e) {
          // Expected
        }
      })

      expect(mockToast.error).toHaveBeenCalledWith('Delete failed')
    })
  })

  describe('refreshCollection', () => {
    it('calls refreshCollection API and shows success toast', async () => {
      const refreshed = { ...mockCollection, item_count: 15 }
      mockCollectionsApi.refreshCollection.mockResolvedValue(refreshed)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.refreshCollection({ id: '1' })
      })

      expect(mockCollectionsApi.refreshCollection).toHaveBeenCalledWith('1')
      expect(mockToast.success).toHaveBeenCalledWith('Collection refreshed successfully')
    })
  })

  describe('duplicateCollection', () => {
    it('calls duplicateCollection API with name and shows success toast', async () => {
      const duplicated = { ...mockCollection, id: '3', name: 'Copy of Test' }
      mockCollectionsApi.duplicateCollection.mockResolvedValue(duplicated)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.duplicateCollection({ id: '1', newName: 'Copy of Test' })
      })

      expect(mockCollectionsApi.duplicateCollection).toHaveBeenCalledWith('1', 'Copy of Test')
      expect(mockToast.success).toHaveBeenCalledWith('Duplicated collection: Copy of Test')
    })
  })

  describe('bulkDeleteCollections', () => {
    it('calls bulkDeleteCollections API', async () => {
      mockCollectionsApi.bulkDeleteCollections.mockResolvedValue(undefined)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.bulkDeleteCollections({ collectionIds: ['1', '2'] })
      })

      expect(mockCollectionsApi.bulkDeleteCollections).toHaveBeenCalledWith(['1', '2'])
    })

    it('shows error toast on bulk delete failure', async () => {
      mockCollectionsApi.bulkDeleteCollections.mockRejectedValue(new Error('Bulk delete failed'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        try {
          await result.current.bulkDeleteCollections({ collectionIds: ['1'] })
        } catch (e) {
          // Expected
        }
      })

      expect(mockToast.error).toHaveBeenCalledWith('Bulk delete failed')
    })
  })

  describe('loading states', () => {
    it('exposes mutation loading states', async () => {
      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      // All mutation loading states should initially be false
      expect(result.current.isCreating).toBe(false)
      expect(result.current.isUpdating).toBe(false)
      expect(result.current.isDeleting).toBe(false)
      expect(result.current.isRefreshing).toBe(false)
      expect(result.current.isSharing).toBe(false)
      expect(result.current.isDuplicating).toBe(false)
      expect(result.current.isExporting).toBe(false)
      expect(result.current.isBulkDeleting).toBe(false)
      expect(result.current.isBulkSharing).toBe(false)
      expect(result.current.isBulkExporting).toBe(false)
      expect(result.current.isBulkUpdating).toBe(false)
    })
  })

  describe('refetchCollections', () => {
    it('exposes a refetchCollections function', async () => {
      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => useCollections(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(typeof result.current.refetchCollections).toBe('function')
    })
  })
})

describe('useCollection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches a single collection and its items by ID', async () => {
    mockCollectionsApi.getCollection.mockResolvedValue(mockCollection)
    mockCollectionsApi.getCollectionItems.mockResolvedValue({
      items: [{ id: 'item-1', title: 'Item 1' }],
      total: 1,
    })

    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => useCollection('1'), { wrapper: Wrapper })

    expect(result.current.isLoading).toBe(true)

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.collection).toEqual(mockCollection)
    expect(mockCollectionsApi.getCollection).toHaveBeenCalledWith('1')
    expect(mockCollectionsApi.getCollectionItems).toHaveBeenCalledWith('1')
  })

  it('does not fetch when id is empty', () => {
    const { Wrapper } = createWrapper()
    renderHook(() => useCollection(''), { wrapper: Wrapper })

    expect(mockCollectionsApi.getCollection).not.toHaveBeenCalled()
    expect(mockCollectionsApi.getCollectionItems).not.toHaveBeenCalled()
  })

  it('exposes refetch functions', async () => {
    mockCollectionsApi.getCollection.mockResolvedValue(mockCollection)
    mockCollectionsApi.getCollectionItems.mockResolvedValue({ items: [], total: 0 })

    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => useCollection('1'), { wrapper: Wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(typeof result.current.refetchCollection).toBe('function')
    expect(typeof result.current.refetchItems).toBe('function')
  })
})

describe('useCollectionAnalytics', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches analytics for a collection', async () => {
    const mockAnalytics = {
      collection_id: '1',
      total_items: 10,
      total_size: 5000000,
      media_type_distribution: { movie: 5, music: 5 },
    }
    mockCollectionsApi.getCollectionAnalytics.mockResolvedValue(mockAnalytics)

    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => useCollectionAnalytics('1'), { wrapper: Wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.analytics).toEqual(mockAnalytics)
    expect(mockCollectionsApi.getCollectionAnalytics).toHaveBeenCalledWith('1')
  })

  it('does not fetch when id is empty', () => {
    const { Wrapper } = createWrapper()
    renderHook(() => useCollectionAnalytics(''), { wrapper: Wrapper })

    expect(mockCollectionsApi.getCollectionAnalytics).not.toHaveBeenCalled()
  })

  it('exposes refetchAnalytics function', async () => {
    mockCollectionsApi.getCollectionAnalytics.mockResolvedValue({})

    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => useCollectionAnalytics('1'), { wrapper: Wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(typeof result.current.refetchAnalytics).toBe('function')
  })
})

describe('useSharedCollection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches a shared collection by shareId', async () => {
    mockCollectionsApi.getSharedCollection.mockResolvedValue(mockCollection)

    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => useSharedCollection('share-abc'), { wrapper: Wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.collection).toEqual(mockCollection)
    expect(mockCollectionsApi.getSharedCollection).toHaveBeenCalledWith('share-abc')
  })

  it('does not fetch when shareId is empty', () => {
    const { Wrapper } = createWrapper()
    renderHook(() => useSharedCollection(''), { wrapper: Wrapper })

    expect(mockCollectionsApi.getSharedCollection).not.toHaveBeenCalled()
  })

  it('exposes error when fetch fails', async () => {
    mockCollectionsApi.getSharedCollection.mockRejectedValue(new Error('Not found'))

    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => useSharedCollection('bad-id'), { wrapper: Wrapper })

    await waitFor(() => {
      expect(result.current.error).toBeTruthy()
    })
  })
})
