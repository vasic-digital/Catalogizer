import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import HomePage from '../HomePage'

// Mock apiService
vi.mock('../../services/apiService', () => ({
  apiService: {
    getMediaStats: vi.fn(),
    searchMedia: vi.fn(),
  },
}))

import { apiService } from '../../services/apiService'

const createTestWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>{children}</BrowserRouter>
    </QueryClientProvider>
  )
}

describe('HomePage (Dashboard)', () => {
  beforeEach(() => {
    vi.mocked(apiService.getMediaStats).mockResolvedValue({
      total_items: 150,
      by_type: { movie: 80, tv_show: 40 },
      by_quality: { '1080p': 100, '4k': 50 },
      total_size: 1099511627776,
      recent_additions: 12,
    })

    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [],
      total: 0,
      limit: 12,
      offset: 0,
    })
  })

  it('renders a time-based greeting', () => {
    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    // The greeting depends on the current time of day
    const greetingEl = screen.getByText(/Welcome back/)
    expect(greetingEl).toBeInTheDocument()

    // Should contain one of the greeting variants
    const greetingText = greetingEl.textContent || ''
    expect(
      greetingText.includes('Good morning') ||
        greetingText.includes('Good afternoon') ||
        greetingText.includes('Good evening')
    ).toBe(true)
  })

  it('renders the subtitle description', () => {
    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    expect(screen.getByText("Here's what's new in your media library")).toBeInTheDocument()
  })

  it('displays stats cards when stats are loaded', async () => {
    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Total Items')).toBeInTheDocument()
      expect(screen.getByText('150')).toBeInTheDocument()
    })

    expect(screen.getByText('Recent Additions')).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()

    expect(screen.getByText('Movies')).toBeInTheDocument()
    expect(screen.getByText('80')).toBeInTheDocument()

    expect(screen.getByText('TV Shows')).toBeInTheDocument()
    expect(screen.getByText('40')).toBeInTheDocument()
  })

  it('displays stat card descriptions', async () => {
    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Media files')).toBeInTheDocument()
    })

    expect(screen.getByText('This week')).toBeInTheDocument()
    expect(screen.getByText('In collection')).toBeInTheDocument()
    expect(screen.getByText('Series available')).toBeInTheDocument()
  })

  it('calls getMediaStats on mount', async () => {
    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(apiService.getMediaStats).toHaveBeenCalled()
    })
  })

  it('calls searchMedia for recently added items', async () => {
    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(apiService.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          sort_by: 'created_at',
          sort_order: 'desc',
          limit: 12,
        })
      )
    })
  })

  it('calls searchMedia for trending/highly rated items', async () => {
    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(apiService.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          sort_by: 'rating',
          sort_order: 'desc',
          limit: 8,
        })
      )
    })
  })

  it('renders recently added section when items exist', async () => {
    vi.mocked(apiService.searchMedia).mockImplementation(async (params) => {
      if (params?.sort_by === 'created_at') {
        return {
          items: [
            {
              id: 1,
              title: 'New Movie',
              media_type: 'movie',
              year: 2023,
              directory_path: '/media/movies',
              created_at: '2023-12-01',
              updated_at: '2023-12-01',
            },
          ],
          total: 1,
          limit: 12,
          offset: 0,
        }
      }
      return { items: [], total: 0, limit: 8, offset: 0 }
    })

    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Recently Added')).toBeInTheDocument()
      expect(screen.getByText('New Movie')).toBeInTheDocument()
    })
  })

  it('renders highly rated section when items exist', async () => {
    vi.mocked(apiService.searchMedia).mockImplementation(async (params) => {
      if (params?.sort_by === 'rating') {
        return {
          items: [
            {
              id: 2,
              title: 'Top Rated Film',
              media_type: 'movie',
              year: 2022,
              rating: 9.2,
              directory_path: '/media/movies',
              created_at: '2023-01-01',
              updated_at: '2023-01-01',
            },
          ],
          total: 1,
          limit: 8,
          offset: 0,
        }
      }
      return { items: [], total: 0, limit: 12, offset: 0 }
    })

    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Highly Rated')).toBeInTheDocument()
      expect(screen.getByText('Top Rated Film')).toBeInTheDocument()
    })
  })

  it('displays year and rating on media cards', async () => {
    vi.mocked(apiService.searchMedia).mockImplementation(async (params) => {
      if (params?.sort_by === 'created_at') {
        return {
          items: [
            {
              id: 1,
              title: 'Detailed Movie',
              media_type: 'movie',
              year: 2024,
              rating: 8.5,
              directory_path: '/media/movies',
              created_at: '2023-12-01',
              updated_at: '2023-12-01',
            },
          ],
          total: 1,
          limit: 12,
          offset: 0,
        }
      }
      return { items: [], total: 0, limit: 8, offset: 0 }
    })

    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Detailed Movie')).toBeInTheDocument()
      expect(screen.getByText('2024')).toBeInTheDocument()
      expect(screen.getByText('8.5')).toBeInTheDocument()
    })
  })

  it('renders media cards as links to detail pages', async () => {
    vi.mocked(apiService.searchMedia).mockImplementation(async (params) => {
      if (params?.sort_by === 'created_at') {
        return {
          items: [
            {
              id: 99,
              title: 'Linked Movie',
              media_type: 'movie',
              directory_path: '/media/movies',
              created_at: '2023-12-01',
              updated_at: '2023-12-01',
            },
          ],
          total: 1,
          limit: 12,
          offset: 0,
        }
      }
      return { items: [], total: 0, limit: 8, offset: 0 }
    })

    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      const link = screen.getByText('Linked Movie').closest('a')
      expect(link).toHaveAttribute('href', '/media/99')
    })
  })

  it('handles zero stats gracefully', async () => {
    vi.mocked(apiService.getMediaStats).mockResolvedValue({
      total_items: 0,
      by_type: {},
      by_quality: {},
      total_size: 0,
      recent_additions: 0,
    })

    const TestWrapper = createTestWrapper()

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Total Items')).toBeInTheDocument()
    })

    // All four stat values should render as "0"
    const zeroElements = screen.getAllByText('0')
    expect(zeroElements.length).toBe(4)
  })
})
