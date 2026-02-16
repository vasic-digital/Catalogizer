import React from 'react'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { usePlaylistReorder } from '../usePlaylistReorder'
import { playlistApi } from '../../lib/playlistsApi'
import { toast } from 'react-hot-toast'

vi.mock('../../lib/playlistsApi', () => ({
  playlistApi: {
    reorderPlaylist: vi.fn(),
  },
  playlistsApi: {
    reorderPlaylist: vi.fn(),
  },
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

const mockPlaylistApi = vi.mocked(playlistApi)

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}

const mockItems = [
  {
    id: 'item-1',
    playlist_id: 'playlist-1',
    media_id: 1,
    position: 0,
    media_item: { id: 1, title: 'Song A', media_type: 'music' },
  },
  {
    id: 'item-2',
    playlist_id: 'playlist-1',
    media_id: 2,
    position: 1,
    media_item: { id: 2, title: 'Song B', media_type: 'music' },
  },
  {
    id: 'item-3',
    playlist_id: 'playlist-1',
    media_id: 3,
    position: 2,
    media_item: { id: 3, title: 'Song C', media_type: 'music' },
  },
]

describe('usePlaylistReorder', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns a mutation object', () => {
    const { result } = renderHook(() => usePlaylistReorder(), {
      wrapper: createWrapper(),
    })

    expect(result.current).toBeDefined()
    expect(result.current.mutate).toBeDefined()
    expect(result.current.mutateAsync).toBeDefined()
  })

  it('calls playlistApi.reorderPlaylist on mutate', async () => {
    mockPlaylistApi.reorderPlaylist.mockResolvedValue(undefined)

    const { result } = renderHook(() => usePlaylistReorder(), {
      wrapper: createWrapper(),
    })

    act(() => {
      result.current.mutate({
        playlistId: 'playlist-1',
        items: mockItems as any,
      })
    })

    await waitFor(() => {
      expect(mockPlaylistApi.reorderPlaylist).toHaveBeenCalledWith(
        'playlist-1',
        ['item-1', 'item-2', 'item-3']
      )
    })
  })

  it('shows success toast on successful reorder', async () => {
    mockPlaylistApi.reorderPlaylist.mockResolvedValue(undefined)

    const { result } = renderHook(() => usePlaylistReorder(), {
      wrapper: createWrapper(),
    })

    act(() => {
      result.current.mutate({
        playlistId: 'playlist-1',
        items: mockItems as any,
      })
    })

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('Playlist reordered successfully')
    })
  })

  it('shows error toast on failed reorder', async () => {
    mockPlaylistApi.reorderPlaylist.mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => usePlaylistReorder(), {
      wrapper: createWrapper(),
    })

    act(() => {
      result.current.mutate({
        playlistId: 'playlist-1',
        items: mockItems as any,
      })
    })

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        'Failed to reorder playlist: Network error'
      )
    })
  })

  it('handles non-Error rejection', async () => {
    mockPlaylistApi.reorderPlaylist.mockRejectedValue('some string error')

    const { result } = renderHook(() => usePlaylistReorder(), {
      wrapper: createWrapper(),
    })

    act(() => {
      result.current.mutate({
        playlistId: 'playlist-1',
        items: mockItems as any,
      })
    })

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        'Failed to reorder playlist: Unknown error'
      )
    })
  })

  it('performs optimistic update on playlists cache', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })

    queryClient.setQueryData(['playlists'], {
      playlists: [
        { id: 'playlist-1', name: 'Test', items: [] },
      ],
    })

    mockPlaylistApi.reorderPlaylist.mockResolvedValue(undefined)

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )

    const { result } = renderHook(() => usePlaylistReorder(), { wrapper })

    act(() => {
      result.current.mutate({
        playlistId: 'playlist-1',
        items: mockItems as any,
      })
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })
  })

  it('rolls back on error when previous data exists', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })

    const previousPlaylists = {
      playlists: [
        { id: 'playlist-1', name: 'Test', items: [{ id: 'old-item' }] },
      ],
    }
    queryClient.setQueryData(['playlists'], previousPlaylists)

    mockPlaylistApi.reorderPlaylist.mockRejectedValue(new Error('Server error'))

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )

    const { result } = renderHook(() => usePlaylistReorder(), { wrapper })

    act(() => {
      result.current.mutate({
        playlistId: 'playlist-1',
        items: mockItems as any,
      })
    })

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })

    // After error, the cache should be rolled back
    const data = queryClient.getQueryData(['playlists']) as any
    expect(data).toEqual(previousPlaylists)
  })

  it('invalidates queries after settling', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })

    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
    mockPlaylistApi.reorderPlaylist.mockResolvedValue(undefined)

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )

    const { result } = renderHook(() => usePlaylistReorder(), { wrapper })

    act(() => {
      result.current.mutate({
        playlistId: 'playlist-1',
        items: mockItems as any,
      })
    })

    await waitFor(() => {
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['playlists'] })
      expect(invalidateSpy).toHaveBeenCalledWith({
        queryKey: ['playlist', 'playlist-1'],
      })
    })

    invalidateSpy.mockRestore()
  })
})
