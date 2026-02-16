import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import HomePage from '../HomePage'

// Mock apiService
vi.mock('../../services/apiService', () => ({
  apiService: {
    getMediaStats: vi.fn(),
    searchMedia: vi.fn(),
  },
}))

// Mock lucide-react
vi.mock('lucide-react', () => ({
  Play: (props: any) => <span data-testid="icon-play" {...props} />,
  Clock: (props: any) => <span data-testid="icon-clock" {...props} />,
  Star: (props: any) => <span data-testid="icon-star" {...props} />,
  TrendingUp: (props: any) => <span data-testid="icon-trending" {...props} />,
  Calendar: (props: any) => <span data-testid="icon-calendar" {...props} />,
}))

import { apiService } from '../../services/apiService'

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = createQueryClient()
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>{children}</BrowserRouter>
    </QueryClientProvider>
  )
}

describe('HomePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Default mocks returning no data
    vi.mocked(apiService.getMediaStats).mockResolvedValue({
      total_items: 0,
      by_type: {},
      by_quality: {},
      total_size: 0,
      recent_additions: 0,
    })
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [],
      total: 0,
      limit: 12,
      offset: 0,
    })
  })

  it('renders the welcome heading', () => {
    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    // The greeting depends on time of day
    expect(screen.getByText(/Welcome back/)).toBeInTheDocument()
  })

  it('renders the subtitle', () => {
    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    expect(screen.getByText("Here's what's new in your media library")).toBeInTheDocument()
  })

  it('displays stat cards when stats are loaded', async () => {
    vi.mocked(apiService.getMediaStats).mockResolvedValue({
      total_items: 1500,
      by_type: { movie: 800, tv_show: 400 },
      by_quality: { '1080p': 500 },
      total_size: 5000000000,
      recent_additions: 25,
    })

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('1,500')).toBeInTheDocument()
    })

    expect(screen.getByText('25')).toBeInTheDocument()
    expect(screen.getByText('800')).toBeInTheDocument()
    expect(screen.getByText('400')).toBeInTheDocument()
  })

  it('displays stat card labels', async () => {
    vi.mocked(apiService.getMediaStats).mockResolvedValue({
      total_items: 100,
      by_type: { movie: 50, tv_show: 30 },
      by_quality: {},
      total_size: 0,
      recent_additions: 5,
    })

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Total Items')).toBeInTheDocument()
    })

    expect(screen.getByText('Recent Additions')).toBeInTheDocument()
    expect(screen.getByText('Movies')).toBeInTheDocument()
    expect(screen.getByText('TV Shows')).toBeInTheDocument()
  })

  it('displays recently added section when items exist', async () => {
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [
        {
          id: 1,
          title: 'Test Movie',
          media_type: 'movie',
          year: 2023,
          directory_path: '/media/movies',
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
      ],
      total: 1,
      limit: 12,
      offset: 0,
    })

    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    await waitFor(() => {
      // searchMedia is called multiple times (recent, continue watching, trending)
      // so the same item may appear in multiple sections
      const matches = screen.getAllByText('Test Movie')
      expect(matches.length).toBeGreaterThan(0)
    })
  })

  it('shows the correct time-based greeting', () => {
    // We cannot easily mock Date, but we can verify the greeting pattern
    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    const greetingElement = screen.getByText(/Welcome back/)
    const text = greetingElement.textContent || ''
    expect(
      text.includes('Good morning') ||
        text.includes('Good afternoon') ||
        text.includes('Good evening')
    ).toBe(true)
  })

  it('calls apiService.getMediaStats on mount', () => {
    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    expect(apiService.getMediaStats).toHaveBeenCalled()
  })

  it('calls apiService.searchMedia on mount for recent items', () => {
    render(
      <TestWrapper>
        <HomePage />
      </TestWrapper>
    )

    expect(apiService.searchMedia).toHaveBeenCalled()
  })
})
