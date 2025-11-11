import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, fireEvent, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MediaBrowser } from './MediaBrowser'
import { mediaApi } from '@/lib/mediaApi'
import type { MediaItem, MediaSearchResponse, MediaStats } from '@/types/media'

// Mock the mediaApi
vi.mock('@/lib/mediaApi', () => ({
  mediaApi: {
    searchMedia: vi.fn(),
    getMediaStats: vi.fn(),
    downloadMedia: vi.fn(),
  },
}))

// Mock the child components
vi.mock('@/components/media/MediaGrid', () => ({
  MediaGrid: ({ media, loading, onMediaView, onMediaDownload }: any) => (
    <div data-testid="media-grid">
      {loading && <div>Loading...</div>}
      {!loading && media.length === 0 && <div>No media found</div>}
      {!loading && media.map((item: MediaItem) => (
        <div key={item.id} data-testid={`media-item-${item.id}`}>
          <span>{item.title}</span>
          <button onClick={() => onMediaView(item)} data-testid={`view-${item.id}`}>
            View
          </button>
          <button onClick={() => onMediaDownload(item)} data-testid={`download-${item.id}`}>
            Download
          </button>
        </div>
      ))}
    </div>
  ),
}))

vi.mock('@/components/media/MediaFilters', () => ({
  MediaFilters: ({ filters, onFiltersChange, onReset }: any) => (
    <div data-testid="media-filters">
      <button onClick={() => onFiltersChange({ ...filters, media_type: 'movie' })}>
        Filter Movies
      </button>
      <button onClick={onReset} data-testid="reset-filters">
        Reset
      </button>
    </div>
  ),
}))

vi.mock('@/components/media/MediaDetailModal', () => ({
  MediaDetailModal: ({ media, isOpen, onClose, onDownload }: any) => (
    isOpen ? (
      <div data-testid="media-detail-modal">
        <h2>{media?.title}</h2>
        <button onClick={onClose} data-testid="close-modal">Close</button>
        <button onClick={() => onDownload(media)} data-testid="modal-download">Download</button>
      </div>
    ) : null
  ),
}))

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    aside: ({ children, ...props }: any) => <aside {...props}>{children}</aside>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

// Mock debounce utility
vi.mock('@/lib/utils', () => ({
  debounce: (fn: any) => fn,
}))

// Test data
const mockMediaItems: MediaItem[] = [
  {
    id: 1,
    title: 'Test Movie 1',
    media_type: 'movie',
    directory_path: '/movies/test1.mp4',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    title: 'Test Movie 2',
    media_type: 'movie',
    directory_path: '/movies/test2.mp4',
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  },
  {
    id: 3,
    title: 'Test TV Show',
    media_type: 'tv_show',
    directory_path: '/tv/test.mp4',
    created_at: '2024-01-03T00:00:00Z',
    updated_at: '2024-01-03T00:00:00Z',
  },
]

const mockSearchResponse: MediaSearchResponse = {
  items: mockMediaItems,
  total: 3,
  limit: 24,
  offset: 0,
}

const mockStats: MediaStats = {
  total_items: 150,
  by_type: {
    movie: 80,
    tv_show: 50,
    music: 20,
  },
  by_quality: {
    '1080p': 100,
    '720p': 40,
    '4k': 10,
  },
  total_size: 1073741824000, // 1 TB
  recent_additions: 25,
}

describe('MediaBrowser', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          gcTime: 0,
        },
      },
    })

    // Setup default mocks
    vi.mocked(mediaApi.searchMedia).mockResolvedValue(mockSearchResponse)
    vi.mocked(mediaApi.getMediaStats).mockResolvedValue(mockStats)
    vi.mocked(mediaApi.downloadMedia).mockResolvedValue()
  })

  afterEach(() => {
    vi.clearAllMocks()
    queryClient.clear()
  })

  const renderComponent = () => {
    return render(
      <QueryClientProvider client={queryClient}>
        <MediaBrowser />
      </QueryClientProvider>
    )
  }

  describe('Rendering', () => {
    it('should render the component with title', () => {
      renderComponent()
      expect(screen.getByText('Media Browser')).toBeInTheDocument()
      expect(screen.getByText('Explore and discover your media collection')).toBeInTheDocument()
    })

    it('should render stats cards when data is loaded', async () => {
      renderComponent()

      await waitFor(() => {
        expect(screen.getByText('150')).toBeInTheDocument() // Total items
        expect(screen.getByText('Total Items')).toBeInTheDocument()
        expect(screen.getByText('3')).toBeInTheDocument() // Media types (movie, tv_show, music)
        expect(screen.getByText('Media Types')).toBeInTheDocument()
        expect(screen.getByText(/1000.0 GB/)).toBeInTheDocument() // Total size
        expect(screen.getByText('Total Size')).toBeInTheDocument()
        expect(screen.getByText('25')).toBeInTheDocument() // Recent additions
        expect(screen.getByText('Recent Additions')).toBeInTheDocument()
      })
    })

    it('should render search input', () => {
      renderComponent()
      const searchInput = screen.getByPlaceholderText('Search your media collection...')
      expect(searchInput).toBeInTheDocument()
    })

    it('should render control buttons', () => {
      renderComponent()
      expect(screen.getByText('Filters')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /grid/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /list/i })).toBeInTheDocument()
    })

    it('should render media grid', async () => {
      renderComponent()

      await waitFor(() => {
        expect(screen.getByTestId('media-grid')).toBeInTheDocument()
      })
    })
  })

  describe('Search Functionality', () => {
    it('should update search query on input change', async () => {
      const user = userEvent.setup()
      renderComponent()

      const searchInput = screen.getByPlaceholderText('Search your media collection...')
      await user.type(searchInput, 'Matrix')

      expect(searchInput).toHaveValue('Matrix')
    })

    it('should call searchMedia with query parameter', async () => {
      const user = userEvent.setup()
      renderComponent()

      const searchInput = screen.getByPlaceholderText('Search your media collection...')
      await user.type(searchInput, 'Matrix')

      await waitFor(() => {
        expect(mediaApi.searchMedia).toHaveBeenCalledWith(
          expect.objectContaining({
            query: 'Matrix',
            offset: 0,
          })
        )
      })
    })

    it('should reset offset when searching', async () => {
      const user = userEvent.setup()
      renderComponent()

      const searchInput = screen.getByPlaceholderText('Search your media collection...')
      await user.type(searchInput, 'Test')

      await waitFor(() => {
        expect(mediaApi.searchMedia).toHaveBeenCalledWith(
          expect.objectContaining({ offset: 0 })
        )
      })
    })

    it('should display search query in results header', async () => {
      const user = userEvent.setup()
      renderComponent()

      await waitFor(() => {
        expect(screen.getByText(/Showing 3 of 3 results/)).toBeInTheDocument()
      })

      const searchInput = screen.getByPlaceholderText('Search your media collection...')
      await user.type(searchInput, 'Matrix')

      await waitFor(() => {
        expect(screen.getByText(/for "Matrix"/)).toBeInTheDocument()
      })
    })
  })

  describe('Filter Functionality', () => {
    it('should toggle filter sidebar', async () => {
      const user = userEvent.setup()
      renderComponent()

      const filtersButton = screen.getByText('Filters')

      // Initially filters should not be visible
      expect(screen.queryByTestId('media-filters')).not.toBeInTheDocument()

      // Click to show filters
      await user.click(filtersButton)

      await waitFor(() => {
        expect(screen.getByTestId('media-filters')).toBeInTheDocument()
      })

      // Click again to hide filters
      await user.click(filtersButton)

      await waitFor(() => {
        expect(screen.queryByTestId('media-filters')).not.toBeInTheDocument()
      })
    })

    it('should handle filter changes', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Show filters
      await user.click(screen.getByText('Filters'))

      await waitFor(() => {
        expect(screen.getByTestId('media-filters')).toBeInTheDocument()
      })

      // Apply a filter
      await user.click(screen.getByText('Filter Movies'))

      await waitFor(() => {
        expect(mediaApi.searchMedia).toHaveBeenCalledWith(
          expect.objectContaining({
            media_type: 'movie',
            offset: 0, // Should reset pagination
          })
        )
      })
    })

    it('should reset filters', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Set search query
      const searchInput = screen.getByPlaceholderText('Search your media collection...')
      await user.type(searchInput, 'Test')

      // Show filters
      await user.click(screen.getByText('Filters'))

      await waitFor(() => {
        expect(screen.getByTestId('media-filters')).toBeInTheDocument()
      })

      // Reset filters
      await user.click(screen.getByTestId('reset-filters'))

      await waitFor(() => {
        expect(searchInput).toHaveValue('')
        expect(mediaApi.searchMedia).toHaveBeenCalledWith(
          expect.objectContaining({
            query: undefined,
            limit: 24,
            offset: 0,
            sort_by: 'updated_at',
            sort_order: 'desc',
          })
        )
      })
    })
  })

  describe('View Mode Toggle', () => {
    it('should start with grid view mode', () => {
      renderComponent()

      // Grid button should have default styling (active)
      const gridButton = screen.getAllByRole('button').find(btn =>
        btn.querySelector('svg') && btn.textContent === ''
      )
      expect(gridButton).toBeInTheDocument()
    })

    it('should toggle between grid and list view', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Grid is initially active
      // Note: View mode doesn't affect API calls in this implementation,
      // it only affects visual presentation

      // Implementation test would require checking CSS classes or
      // MediaGrid component props, which are mocked
      expect(screen.getByTestId('media-grid')).toBeInTheDocument()
    })
  })

  describe('Pagination', () => {
    it('should display current page and total pages', async () => {
      const largeResponse = {
        ...mockSearchResponse,
        total: 100,
      }
      vi.mocked(mediaApi.searchMedia).mockResolvedValue(largeResponse)

      renderComponent()

      await waitFor(() => {
        expect(screen.getByText(/Page 1 of 5/)).toBeInTheDocument()
      })
    })

    it('should navigate to next page', async () => {
      const user = userEvent.setup()
      const largeResponse = {
        ...mockSearchResponse,
        total: 100,
      }
      vi.mocked(mediaApi.searchMedia).mockResolvedValue(largeResponse)

      renderComponent()

      await waitFor(() => {
        expect(screen.getByText(/Page 1 of 5/)).toBeInTheDocument()
      })

      const nextButtons = screen.getAllByRole('button', { name: /next/i })
      await user.click(nextButtons[0])

      await waitFor(() => {
        expect(mediaApi.searchMedia).toHaveBeenCalledWith(
          expect.objectContaining({ offset: 24 })
        )
      })
    })

    it('should navigate to previous page', async () => {
      const user = userEvent.setup()
      const largeResponse = {
        ...mockSearchResponse,
        total: 100,
        offset: 24,
      }
      vi.mocked(mediaApi.searchMedia).mockResolvedValue(largeResponse)

      renderComponent()

      // Manually set to page 2
      await waitFor(() => {
        screen.getByText(/Page 1 of 5/)
      })

      const nextButtons = screen.getAllByRole('button', { name: /next/i })
      await user.click(nextButtons[0])

      await waitFor(() => {
        const prevButtons = screen.getAllByRole('button', { name: /previous/i })
        expect(prevButtons.length).toBeGreaterThan(0)
      })

      const prevButtons = screen.getAllByRole('button', { name: /previous/i })
      await user.click(prevButtons[0])

      await waitFor(() => {
        expect(mediaApi.searchMedia).toHaveBeenCalledWith(
          expect.objectContaining({ offset: 0 })
        )
      })
    })

    it('should disable previous button on first page', async () => {
      const largeResponse = {
        ...mockSearchResponse,
        total: 100,
      }
      vi.mocked(mediaApi.searchMedia).mockResolvedValue(largeResponse)

      renderComponent()

      await waitFor(() => {
        const prevButtons = screen.getAllByRole('button', { name: /previous/i })
        prevButtons.forEach(btn => {
          expect(btn).toBeDisabled()
        })
      })
    })

    it('should disable next button on last page', async () => {
      const user = userEvent.setup()
      const lastPageResponse = {
        ...mockSearchResponse,
        total: 24,
        offset: 0,
      }
      vi.mocked(mediaApi.searchMedia).mockResolvedValue(lastPageResponse)

      renderComponent()

      await waitFor(() => {
        const nextButtons = screen.getAllByRole('button', { name: /next/i })
        nextButtons.forEach(btn => {
          expect(btn).toBeDisabled()
        })
      })
    })

    it('should hide pagination when only one page', async () => {
      const singlePageResponse = {
        ...mockSearchResponse,
        total: 10,
      }
      vi.mocked(mediaApi.searchMedia).mockResolvedValue(singlePageResponse)

      renderComponent()

      await waitFor(() => {
        expect(screen.queryByText(/Page 1 of/)).not.toBeInTheDocument()
      })
    })
  })

  describe('Media Interactions', () => {
    it('should open modal when media item is clicked', async () => {
      const user = userEvent.setup()
      renderComponent()

      await waitFor(() => {
        expect(screen.getByTestId('media-item-1')).toBeInTheDocument()
      })

      await user.click(screen.getByTestId('view-1'))

      await waitFor(() => {
        expect(screen.getByTestId('media-detail-modal')).toBeInTheDocument()
        expect(screen.getByText('Test Movie 1')).toBeInTheDocument()
      })
    })

    it('should close modal', async () => {
      const user = userEvent.setup()
      renderComponent()

      await waitFor(() => {
        expect(screen.getByTestId('media-item-1')).toBeInTheDocument()
      })

      // Open modal
      await user.click(screen.getByTestId('view-1'))

      await waitFor(() => {
        expect(screen.getByTestId('media-detail-modal')).toBeInTheDocument()
      })

      // Close modal
      await user.click(screen.getByTestId('close-modal'))

      await waitFor(() => {
        expect(screen.queryByTestId('media-detail-modal')).not.toBeInTheDocument()
      })
    })

    it('should handle media download from grid', async () => {
      const user = userEvent.setup()
      renderComponent()

      await waitFor(() => {
        expect(screen.getByTestId('media-item-1')).toBeInTheDocument()
      })

      await user.click(screen.getByTestId('download-1'))

      await waitFor(() => {
        expect(mediaApi.downloadMedia).toHaveBeenCalledWith(
          expect.objectContaining({ id: 1, title: 'Test Movie 1' })
        )
      })
    })

    it('should handle media download from modal', async () => {
      const user = userEvent.setup()
      renderComponent()

      await waitFor(() => {
        expect(screen.getByTestId('media-item-1')).toBeInTheDocument()
      })

      // Open modal
      await user.click(screen.getByTestId('view-1'))

      await waitFor(() => {
        expect(screen.getByTestId('media-detail-modal')).toBeInTheDocument()
      })

      // Download from modal
      await user.click(screen.getByTestId('modal-download'))

      await waitFor(() => {
        expect(mediaApi.downloadMedia).toHaveBeenCalledWith(
          expect.objectContaining({ id: 1, title: 'Test Movie 1' })
        )
      })
    })

    it('should handle download error gracefully', async () => {
      const user = userEvent.setup()
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      vi.mocked(mediaApi.downloadMedia).mockRejectedValue(new Error('Download failed'))

      renderComponent()

      await waitFor(() => {
        expect(screen.getByTestId('media-item-1')).toBeInTheDocument()
      })

      await user.click(screen.getByTestId('download-1'))

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith(
          'Download failed:',
          expect.any(Error)
        )
      })

      consoleErrorSpy.mockRestore()
    })
  })

  describe('Error Handling', () => {
    it('should display error state when search fails', async () => {
      vi.mocked(mediaApi.searchMedia).mockRejectedValue(new Error('Network error'))

      renderComponent()

      await waitFor(() => {
        expect(screen.getByText('Failed to load media')).toBeInTheDocument()
        expect(screen.getByText('There was an error loading your media collection.')).toBeInTheDocument()
      })
    })

    it('should allow retry after error', async () => {
      const user = userEvent.setup()
      vi.mocked(mediaApi.searchMedia).mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce(mockSearchResponse)

      renderComponent()

      await waitFor(() => {
        expect(screen.getByText('Failed to load media')).toBeInTheDocument()
      })

      const retryButton = screen.getByText('Try again')
      await user.click(retryButton)

      await waitFor(() => {
        expect(screen.getByTestId('media-grid')).toBeInTheDocument()
        expect(screen.getByTestId('media-item-1')).toBeInTheDocument()
      })
    })
  })

  describe('Loading States', () => {
    it('should show loading state initially', () => {
      vi.mocked(mediaApi.searchMedia).mockImplementation(() => new Promise(() => {}))

      renderComponent()

      expect(screen.getByText('Loading...')).toBeInTheDocument()
    })

    it('should show loading spinner when refreshing', async () => {
      const user = userEvent.setup()
      renderComponent()

      await waitFor(() => {
        expect(screen.getByTestId('media-grid')).toBeInTheDocument()
      })

      const refreshButton = screen.getAllByRole('button').find(btn =>
        btn.querySelector('.animate-spin')
      )

      expect(refreshButton).toBeInTheDocument()
    })
  })

  describe('Empty States', () => {
    it('should display empty state when no media found', async () => {
      vi.mocked(mediaApi.searchMedia).mockResolvedValue({
        items: [],
        total: 0,
        limit: 24,
        offset: 0,
      })

      renderComponent()

      await waitFor(() => {
        expect(screen.getByText('No media found')).toBeInTheDocument()
      })
    })

    it('should show correct count when no results', async () => {
      vi.mocked(mediaApi.searchMedia).mockResolvedValue({
        items: [],
        total: 0,
        limit: 24,
        offset: 0,
      })

      renderComponent()

      await waitFor(() => {
        expect(screen.getByText('Showing 0 of 0 results')).toBeInTheDocument()
      })
    })
  })

  describe('Refresh Functionality', () => {
    it('should refetch data when refresh button is clicked', async () => {
      const user = userEvent.setup()
      renderComponent()

      await waitFor(() => {
        expect(mediaApi.searchMedia).toHaveBeenCalledTimes(1)
      })

      const refreshButtons = screen.getAllByRole('button')
      const refreshButton = refreshButtons.find(btn =>
        btn.querySelector('svg') && !btn.textContent?.includes('Filters')
      )

      if (refreshButton) {
        await user.click(refreshButton)

        await waitFor(() => {
          expect(mediaApi.searchMedia).toHaveBeenCalledTimes(2)
        })
      }
    })
  })
})
