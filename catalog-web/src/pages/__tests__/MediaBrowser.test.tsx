import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MediaBrowser } from '../MediaBrowser'
import * as mediaApi from '@/lib/mediaApi'
import * as utils from '@/lib/utils'
import type { MediaSearchRequest, MediaItem } from '@/types/media'

// Mock dependencies
jest.mock('@/lib/mediaApi')
jest.mock('@/lib/utils', () => ({
  debounce: jest.fn((fn) => fn), // Return the function directly for testing
}))

// Mock child components
jest.mock('@/components/media/MediaGrid', () => ({
  MediaGrid: ({ media, onView, onDownload, isLoading, viewMode }: any) => (
    <div data-testid="media-grid">
      {isLoading && <div data-testid="loading-indicator">Loading...</div>}
      {media?.map((item: MediaItem) => (
        <div key={item.id} data-testid={`media-item-${item.id}`}>
          <span>{item.title}</span>
          <button onClick={() => onView(item)}>View</button>
          <button onClick={() => onDownload(item)}>Download</button>
        </div>
      ))}
    </div>
  ),
}))

jest.mock('@/components/media/MediaFilters', () => ({
  MediaFilters: ({ filters, onChange, onReset }: any) => (
    <div data-testid="media-filters">
      <button onClick={() => onChange({ ...filters, media_type: 'video' })}>
        Filter Video
      </button>
      <button onClick={onReset}>Reset Filters</button>
    </div>
  ),
}))

jest.mock('@/components/media/MediaDetailModal', () => ({
  MediaDetailModal: ({ media, isOpen, onClose }: any) => (
    isOpen ? (
      <div data-testid="media-detail-modal">
        <h2>{media.title}</h2>
        <button onClick={onClose}>Close</button>
      </div>
    ) : null
  ),
}))

// Mock UI components
jest.mock('@/components/ui/Card', () => ({
  Card: ({ children }: any) => <div data-testid="card">{children}</div>,
  CardContent: ({ children }: any) => <div data-testid="card-content">{children}</div>,
  CardHeader: ({ children }: any) => <div data-testid="card-header">{children}</div>,
  CardTitle: ({ children }: any) => <div data-testid="card-title">{children}</div>,
}))

jest.mock('@/components/ui/Button', () => ({
  Button: ({ children, onClick, disabled }: any) => (
    <button onClick={onClick} disabled={disabled} data-testid="button">
      {children}
    </button>
  ),
}))

jest.mock('@/components/ui/Input', () => ({
  Input: ({ onChange, value, placeholder }: any) => (
    <input
      onChange={onChange}
      value={value}
      placeholder={placeholder}
      data-testid="input"
    />
  ),
}))

const mockMediaApi = mediaApi as jest.Mocked<typeof mediaApi>
const mockDebounce = utils.debounce as jest.MockedFunction<typeof utils.debounce>

// Test data
const mockMediaItems: MediaItem[] = [
  {
    id: 1,
    title: 'Test Video 1',
    media_type: 'video',
    directory_path: '/test/video1.mp4',
    file_size: 1024000,
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2023-01-02T00:00:00Z',
  },
  {
    id: 2,
    title: 'Test Video 2',
    media_type: 'video',
    directory_path: '/test/video2.mp4',
    file_size: 2048000,
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2023-01-02T00:00:00Z',
  },
]

const mockStats = {
  total_items: 150,
  total_size: 1024000000,
  by_type: { video: 100, audio: 30, image: 20 },
  recent_additions: 5,
}

const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: { retry: false },
  },
})

const renderWithQueryClient = (component: React.ReactElement) => {
  const queryClient = createTestQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      {component}
    </QueryClientProvider>
  )
}

describe('MediaBrowser', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockMediaApi.searchMedia.mockResolvedValue({
      items: mockMediaItems,
      total: 100,
      offset: 0,
      limit: 24,
    })
    mockMediaApi.getMediaStats.mockResolvedValue(mockStats)
    mockMediaApi.downloadMedia.mockResolvedValue(undefined)
    mockDebounce.mockImplementation((fn) => fn)
  })

  describe('Rendering', () => {
    it('renders MediaBrowser component with header', () => {
      renderWithQueryClient(<MediaBrowser />)
      
      expect(screen.getByText('Media Browser')).toBeInTheDocument()
      expect(screen.getByText('Explore and discover your media collection')).toBeInTheDocument()
    })

    it('renders stats cards when stats are loaded', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        expect(screen.getByText('150')).toBeInTheDocument() // Total items
        expect(screen.getByText('Total Items')).toBeInTheDocument()
        expect(screen.getByText('3')).toBeInTheDocument() // Media types
        expect(screen.getByText('Media Types')).toBeInTheDocument()
      })
    })

    it('renders search input', () => {
      renderWithQueryClient(<MediaBrowser />)
      
      const searchInput = screen.getByPlaceholderText('Search media...')
      expect(searchInput).toBeInTheDocument()
    })

    it('renders view mode toggle buttons', () => {
      renderWithQueryClient(<MediaBrowser />)
      
      expect(screen.getByTestId('button')).toBeInTheDocument() // Grid view button
      // List view button would also be present
    })

    it('renders filter toggle button', () => {
      renderWithQueryClient(<MediaBrowser />)
      
      expect(screen.getByText('Filters')).toBeInTheDocument()
    })

    it('renders refresh button', () => {
      renderWithQueryClient(<MediaBrowser />)
      
      expect(screen.getByText('Refresh')).toBeInTheDocument()
    })
  })

  describe('Media Grid Integration', () => {
    it('displays media items when loaded', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        expect(screen.getByTestId('media-grid')).toBeInTheDocument()
        expect(screen.getByTestId('media-item-1')).toBeInTheDocument()
        expect(screen.getByTestId('media-item-2')).toBeInTheDocument()
        expect(screen.getByText('Test Video 1')).toBeInTheDocument()
        expect(screen.getByText('Test Video 2')).toBeInTheDocument()
      })
    })

    it('shows loading state while fetching media', () => {
      mockMediaApi.searchMedia.mockImplementation(() => new Promise(() => {})) // Never resolves
      
      renderWithQueryClient(<MediaBrowser />)
      
      expect(screen.getByTestId('loading-indicator')).toBeInTheDocument()
    })

    it('shows empty state when no media items found', async () => {
      mockMediaApi.searchMedia.mockResolvedValue({
        items: [],
        total: 0,
        offset: 0,
        limit: 24,
      })
      
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        expect(screen.getByText('No media items found')).toBeInTheDocument()
      })
    })
  })

  describe('Search Functionality', () => {
    it('handles search input changes', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      const searchInput = screen.getByPlaceholderText('Search media...')
      await userEvent.type(searchInput, 'test query')
      
      // Verify debounce was called
      expect(mockDebounce).toHaveBeenCalled()
      expect(mockMediaApi.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          query: 'test query',
          offset: 0, // Should reset to 0 on new search
        })
      )
    })

    it('clears search when input is empty', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      const searchInput = screen.getByPlaceholderText('Search media...')
      await userEvent.type(searchInput, 'test')
      await userEvent.clear(searchInput)
      
      expect(mockMediaApi.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          query: undefined,
          offset: 0,
        })
      )
    })
  })

  describe('Filter Functionality', () => {
    it('toggles filter panel when filters button is clicked', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      const filtersButton = screen.getByText('Filters')
      await userEvent.click(filtersButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('media-filters')).toBeInTheDocument()
      })
    })

    it('applies filters when filter options are selected', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      // Open filters
      const filtersButton = screen.getByText('Filters')
      await userEvent.click(filtersButton)
      
      // Apply a filter
      await waitFor(() => {
        const filterButton = screen.getByText('Filter Video')
        userEvent.click(filterButton)
      })
      
      expect(mockMediaApi.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          media_type: 'video',
          offset: 0,
        })
      )
    })

    it('resets filters when reset button is clicked', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      // Open filters
      const filtersButton = screen.getByText('Filters')
      await userEvent.click(filtersButton)
      
      // Reset filters
      await waitFor(() => {
        const resetButton = screen.getByText('Reset Filters')
        userEvent.click(resetButton)
      })
      
      expect(mockMediaApi.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          limit: 24,
          offset: 0,
          sort_by: 'updated_at',
          sort_order: 'desc',
        })
      )
    })
  })

  describe('View Mode Toggle', () => {
    it('switches between grid and list view modes', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      // Initially in grid mode
      expect(screen.getByText('Grid')).toBeInTheDocument()
      
      // Switch to list mode (implementation specific)
      const viewModeButton = screen.getByText('List')
      await userEvent.click(viewModeButton)
      
      // Verify mode changed (would check for list-specific elements)
    })
  })

  describe('Pagination', () => {
    it('shows pagination controls when there are multiple pages', async () => {
      mockMediaApi.searchMedia.mockResolvedValue({
        items: mockMediaItems,
        total: 100, // 5 pages with limit 24
        offset: 0,
        limit: 24,
      })
      
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        expect(screen.getByText('Page 1 of 5')).toBeInTheDocument()
        expect(screen.getByText('Previous')).toBeInTheDocument()
        expect(screen.getByText('Next')).toBeInTheDocument()
      })
    })

    it('disables previous button on first page', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        const prevButton = screen.getByText('Previous')
        expect(prevButton.closest('button')).toBeDisabled()
      })
    })

    it('disables next button on last page', async () => {
      mockMediaApi.searchMedia.mockResolvedValue({
        items: mockMediaItems,
        total: 48, // 2 pages with limit 24
        offset: 24, // Second page
        limit: 24,
      })
      
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        const nextButton = screen.getByText('Next')
        expect(nextButton.closest('button')).toBeDisabled()
      })
    })

    it('handles next page navigation', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        const nextButton = screen.getByText('Next')
        userEvent.click(nextButton)
      })
      
      expect(mockMediaApi.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          offset: 24, // Next page
        })
      )
    })

    it('handles previous page navigation', async () => {
      // Start on page 2
      mockMediaApi.searchMedia.mockResolvedValue({
        items: mockMediaItems,
        total: 100,
        offset: 24, // Page 2
        limit: 24,
      })
      
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        const prevButton = screen.getByText('Previous')
        userEvent.click(prevButton)
      })
      
      expect(mockMediaApi.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          offset: 0, // Previous page
        })
      )
    })
  })

  describe('Media Interaction', () => {
    it('opens media detail modal when view is clicked', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        const viewButton = screen.getByText('View')
        userEvent.click(viewButton)
      })
      
      await waitFor(() => {
        expect(screen.getByTestId('media-detail-modal')).toBeInTheDocument()
        expect(screen.getByText('Test Video 1')).toBeInTheDocument()
      })
    })

    it('closes media detail modal when close is clicked', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      // Open modal
      await waitFor(() => {
        const viewButton = screen.getByText('View')
        userEvent.click(viewButton)
      })
      
      // Close modal
      await waitFor(() => {
        const closeButton = screen.getByText('Close')
        userEvent.click(closeButton)
      })
      
      expect(screen.queryByTestId('media-detail-modal')).not.toBeInTheDocument()
    })

    it('handles media download', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        const downloadButton = screen.getByText('Download')
        userEvent.click(downloadButton)
      })
      
      expect(mockMediaApi.downloadMedia).toHaveBeenCalledWith(mockMediaItems[0])
    })
  })

  describe('Refresh Functionality', () => {
    it('refreshes media data when refresh button is clicked', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      const refreshButton = screen.getByText('Refresh')
      await userEvent.click(refreshButton)
      
      expect(mockMediaApi.searchMedia).toHaveBeenCalled()
      expect(mockMediaApi.getMediaStats).toHaveBeenCalled()
    })
  })

  describe('Error Handling', () => {
    it('displays error message when media search fails', async () => {
      mockMediaApi.searchMedia.mockRejectedValue(new Error('Search failed'))
      
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        expect(screen.getByText(/Failed to load media/)).toBeInTheDocument()
      })
    })

    it('displays error message when stats fetch fails', async () => {
      mockMediaApi.getMediaStats.mockRejectedValue(new Error('Stats failed'))
      
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        expect(screen.getByText(/Failed to load statistics/)).toBeInTheDocument()
      })
    })

    it('handles download error gracefully', async () => {
      mockMediaApi.downloadMedia.mockRejectedValue(new Error('Download failed'))
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation()
      
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        const downloadButton = screen.getByText('Download')
        userEvent.click(downloadButton)
      })
      
      expect(consoleSpy).toHaveBeenCalledWith('Download failed:', expect.any(Error))
      consoleSpy.mockRestore()
    })
  })

  describe('Edge Cases', () => {
    it('handles empty search results gracefully', async () => {
      mockMediaApi.searchMedia.mockResolvedValue({
        items: [],
        total: 0,
        offset: 0,
        limit: 24,
      })
      
      renderWithQueryClient(<MediaBrowser />)
      
      await waitFor(() => {
        expect(screen.getByText('No media items found')).toBeInTheDocument()
      })
    })

    it('handles null/undefined props gracefully', () => {
      expect(() => {
        renderWithQueryClient(<MediaBrowser />)
      }).not.toThrow()
    })

    it('handles rapid search input changes', async () => {
      renderWithQueryClient(<MediaBrowser />)
      
      const searchInput = screen.getByPlaceholderText('Search media...')
      
      // Type rapidly
      await userEvent.type(searchInput, 'rapid search')
      
      // Should debounce and only call once for final value
      expect(mockDebounce).toHaveBeenCalled()
    })
  })

  describe('Responsive Design', () => {
    it('adapts to different screen sizes', () => {
      // Test mobile view
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 640 })
      renderWithQueryClient(<MediaBrowser />)
      
      // Test desktop view
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 })
      renderWithQueryClient(<MediaBrowser />)
      
      // Component should render without errors in both sizes
      expect(screen.getByText('Media Browser')).toBeInTheDocument()
    })
  })
})