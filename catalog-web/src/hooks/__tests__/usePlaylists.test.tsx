import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import { usePlaylists, usePlaylistItems, usePlaylistAnalytics } from '../usePlaylists'

// Mock playlistsApi
jest.mock('@/lib/playlistsApi', () => ({
  playlistsApi: {
    getPlaylists: jest.fn(),
    getPlaylistItems: jest.fn(),
    createPlaylist: jest.fn(),
    updatePlaylist: jest.fn(),
    deletePlaylist: jest.fn(),
    addItemsToPlaylist: jest.fn(),
    removeFromPlaylist: jest.fn(),
    reorderPlaylistItems: jest.fn(),
    getPlaylistAnalytics: jest.fn(),
  },
}))

// Mock react-hot-toast
jest.mock('react-hot-toast', () => ({
  __esModule: true,
  default: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockPlaylistsApi = require('@/lib/playlistsApi').playlistsApi
const mockToast = require('react-hot-toast').default

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

const mockPlaylist = {
  id: 'pl-1',
  name: 'My Playlist',
  description: 'Test playlist',
  user_id: 1,
  is_public: false,
  is_smart: false,
  item_count: 3,
  total_duration: 7200,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockPlaylistsResponse = {
  playlists: [mockPlaylist],
  total: 1,
  limit: 20,
  offset: 0,
}

const mockPlaylistItem = {
  id: 'item-1',
  playlist_id: 'pl-1',
  media_id: 42,
  position: 1,
  media_item: {
    id: 42,
    title: 'Test Movie',
    media_type: 'movie',
    year: 2024,
    duration: 7200,
  },
  added_at: '2024-01-01T00:00:00Z',
}

const mockPlaylistItemsResponse = {
  items: [mockPlaylistItem],
  total: 1,
  playlist: mockPlaylist,
}

const mockAnalytics = {
  playlist_id: 'pl-1',
  total_plays: 50,
  unique_viewers: 10,
  average_completion_rate: 0.85,
  popular_items: [],
  viewing_stats: [],
}

describe('usePlaylists', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockPlaylistsApi.getPlaylists.mockResolvedValue(mockPlaylistsResponse)
    // Mock window.confirm for delete tests
    window.confirm = jest.fn(() => true)
  })

  describe('Fetching playlists', () => {
    it('fetches playlists on mount', async () => {
      const { Wrapper } = createWrapper()

      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      expect(result.current.isLoading).toBe(true)

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.playlists).toEqual([mockPlaylist])
      expect(result.current.total).toBe(1)
    })

    it('passes params to getPlaylists', async () => {
      const { Wrapper } = createWrapper()
      const params = { limit: 10, offset: 0, include_smart: true }

      renderHook(() => usePlaylists(params), { wrapper: Wrapper })

      await waitFor(() => {
        expect(mockPlaylistsApi.getPlaylists).toHaveBeenCalledWith(params)
      })
    })

    it('returns empty array when no playlists data', async () => {
      mockPlaylistsApi.getPlaylists.mockResolvedValue(undefined)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.playlists).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('exposes error when fetch fails', async () => {
      const error = new Error('Network error')
      mockPlaylistsApi.getPlaylists.mockRejectedValue(error)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.error).toBeTruthy()
      })
    })
  })

  describe('createPlaylist', () => {
    it('creates a playlist and shows success toast', async () => {
      mockPlaylistsApi.createPlaylist.mockResolvedValue(mockPlaylist)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.createPlaylist({ name: 'My Playlist' })
      })

      expect(mockPlaylistsApi.createPlaylist).toHaveBeenCalledWith({ name: 'My Playlist' })
      expect(mockToast.success).toHaveBeenCalledWith('Playlist "My Playlist" created successfully')
    })

    it('shows error toast on create failure', async () => {
      mockPlaylistsApi.createPlaylist.mockRejectedValue({
        response: { data: { message: 'Duplicate name' } },
      })

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        try {
          await result.current.createPlaylist({ name: 'Duplicate' })
        } catch {
          // Expected to throw
        }
      })

      expect(mockToast.error).toHaveBeenCalledWith('Duplicate name')
    })

    it('shows default error message when no response message', async () => {
      mockPlaylistsApi.createPlaylist.mockRejectedValue(new Error('Network error'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        try {
          await result.current.createPlaylist({ name: 'Test' })
        } catch {
          // Expected to throw
        }
      })

      expect(mockToast.error).toHaveBeenCalledWith('Failed to create playlist')
    })
  })

  describe('updatePlaylist', () => {
    it('updates a playlist and shows success toast', async () => {
      const updatedPlaylist = { ...mockPlaylist, name: 'Updated Name' }
      mockPlaylistsApi.updatePlaylist.mockResolvedValue(updatedPlaylist)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.updatePlaylist('pl-1', { name: 'Updated Name' })
      })

      await waitFor(() => {
        expect(mockPlaylistsApi.updatePlaylist).toHaveBeenCalledWith('pl-1', { name: 'Updated Name' })
        expect(mockToast.success).toHaveBeenCalledWith('Playlist "Updated Name" updated successfully')
      })
    })

    it('shows error toast on update failure', async () => {
      mockPlaylistsApi.updatePlaylist.mockRejectedValue(new Error('Server error'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.updatePlaylist('pl-1', { name: 'Fail' })
      })

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to update playlist')
      })
    })
  })

  describe('deletePlaylist', () => {
    it('deletes a playlist after user confirmation', async () => {
      mockPlaylistsApi.deletePlaylist.mockResolvedValue(undefined)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.deletePlaylist('pl-1')
      })

      expect(window.confirm).toHaveBeenCalled()

      await waitFor(() => {
        expect(mockPlaylistsApi.deletePlaylist).toHaveBeenCalledWith('pl-1')
        expect(mockToast.success).toHaveBeenCalledWith('Playlist deleted successfully')
      })
    })

    it('does not delete when user cancels confirmation', async () => {
      ;(window.confirm as jest.Mock).mockReturnValue(false)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.deletePlaylist('pl-1')
      })

      expect(window.confirm).toHaveBeenCalled()
      expect(mockPlaylistsApi.deletePlaylist).not.toHaveBeenCalled()
    })

    it('shows error toast on delete failure', async () => {
      mockPlaylistsApi.deletePlaylist.mockRejectedValue(new Error('Cannot delete'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.deletePlaylist('pl-1')
      })

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to delete playlist')
      })
    })
  })

  describe('addItemsToPlaylist', () => {
    it('adds items and shows success toast', async () => {
      mockPlaylistsApi.addItemsToPlaylist.mockResolvedValue({ added: 2, failed: 0 })

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.addItemsToPlaylist('pl-1', [1, 2])
      })

      await waitFor(() => {
        expect(mockPlaylistsApi.addItemsToPlaylist).toHaveBeenCalledWith('pl-1', [1, 2])
        expect(mockToast.success).toHaveBeenCalledWith('Added 2 items to playlist')
      })
    })

    it('shows error toast on add items failure', async () => {
      mockPlaylistsApi.addItemsToPlaylist.mockRejectedValue(new Error('Server error'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.addItemsToPlaylist('pl-1', [1, 2])
      })

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to add items to playlist')
      })
    })
  })

  describe('removeFromPlaylist', () => {
    it('removes an item and shows success toast', async () => {
      mockPlaylistsApi.removeFromPlaylist.mockResolvedValue(undefined)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.removeFromPlaylist('pl-1', 'item-1')
      })

      await waitFor(() => {
        expect(mockPlaylistsApi.removeFromPlaylist).toHaveBeenCalledWith('pl-1', 'item-1')
        expect(mockToast.success).toHaveBeenCalledWith('Item removed from playlist')
      })
    })

    it('shows error toast on remove failure', async () => {
      mockPlaylistsApi.removeFromPlaylist.mockRejectedValue(new Error('Server error'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.removeFromPlaylist('pl-1', 'item-1')
      })

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to remove item from playlist')
      })
    })
  })

  describe('reorderPlaylistItems', () => {
    it('reorders items successfully', async () => {
      mockPlaylistsApi.reorderPlaylistItems.mockResolvedValue(undefined)

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      const itemOrders = [
        { id: 'item-1', position: 2 },
        { id: 'item-2', position: 1 },
      ]

      act(() => {
        result.current.reorderPlaylistItems('pl-1', itemOrders)
      })

      await waitFor(() => {
        expect(mockPlaylistsApi.reorderPlaylistItems).toHaveBeenCalledWith('pl-1', itemOrders)
      })
    })

    it('shows error toast on reorder failure', async () => {
      mockPlaylistsApi.reorderPlaylistItems.mockRejectedValue(new Error('Server error'))

      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.reorderPlaylistItems('pl-1', [{ id: 'item-1', position: 1 }])
      })

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to reorder playlist')
      })
    })
  })

  describe('loading states', () => {
    it('exposes isCreating state', async () => {
      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.isCreating).toBe(false)
    })

    it('exposes isUpdating state', async () => {
      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.isUpdating).toBe(false)
    })

    it('exposes isDeleting state', async () => {
      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.isDeleting).toBe(false)
    })

    it('exposes isAddingItems state', async () => {
      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.isAddingItems).toBe(false)
    })
  })

  describe('refetch', () => {
    it('exposes refetchPlaylists function', async () => {
      const { Wrapper } = createWrapper()
      const { result } = renderHook(() => usePlaylists(), { wrapper: Wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(typeof result.current.refetchPlaylists).toBe('function')
    })
  })
})

describe('usePlaylistItems', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockPlaylistsApi.getPlaylistItems.mockResolvedValue(mockPlaylistItemsResponse)
  })

  it('fetches playlist items for a given playlist ID', async () => {
    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => usePlaylistItems('pl-1'), { wrapper: Wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.items).toEqual([mockPlaylistItem])
    expect(result.current.total).toBe(1)
    expect(result.current.playlist).toEqual(mockPlaylist)
  })

  it('does not fetch when playlistId is empty', () => {
    const { Wrapper } = createWrapper()
    renderHook(() => usePlaylistItems(''), { wrapper: Wrapper })

    expect(mockPlaylistsApi.getPlaylistItems).not.toHaveBeenCalled()
  })

  it('returns empty defaults when no data', async () => {
    mockPlaylistsApi.getPlaylistItems.mockResolvedValue(undefined)

    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => usePlaylistItems('pl-1'), { wrapper: Wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.items).toEqual([])
    expect(result.current.total).toBe(0)
    expect(result.current.playlist).toBeUndefined()
  })
})

describe('usePlaylistAnalytics', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockPlaylistsApi.getPlaylistAnalytics.mockResolvedValue(mockAnalytics)
  })

  it('fetches analytics for a given playlist ID', async () => {
    const { Wrapper } = createWrapper()
    const { result } = renderHook(() => usePlaylistAnalytics('pl-1'), { wrapper: Wrapper })

    await waitFor(() => {
      expect(result.current.data).toEqual(mockAnalytics)
    })

    expect(mockPlaylistsApi.getPlaylistAnalytics).toHaveBeenCalledWith('pl-1')
  })

  it('does not fetch when playlistId is empty', () => {
    const { Wrapper } = createWrapper()
    renderHook(() => usePlaylistAnalytics(''), { wrapper: Wrapper })

    expect(mockPlaylistsApi.getPlaylistAnalytics).not.toHaveBeenCalled()
  })
})
