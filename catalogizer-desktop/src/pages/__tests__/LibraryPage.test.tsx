import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import LibraryPage from '../LibraryPage'

// Mock apiService
vi.mock('../../services/apiService', () => ({
  apiService: {
    searchMedia: vi.fn(),
  },
}))

// Mock lucide-react
vi.mock('lucide-react', () => ({
  Grid: (props: any) => <span data-testid="icon-grid" {...props} />,
  List: (props: any) => <span data-testid="icon-list" {...props} />,
  Search: (props: any) => <span data-testid="icon-search" {...props} />,
  Play: (props: any) => <span data-testid="icon-play" {...props} />,
  Star: (props: any) => <span data-testid="icon-star" {...props} />,
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

describe('LibraryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [],
      total: 0,
      limit: 50,
      offset: 0,
    })
  })

  it('renders the page title', () => {
    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    expect(screen.getByText('Library')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    expect(screen.getByText('Browse and manage your media collection')).toBeInTheDocument()
  })

  it('renders the search input with placeholder', () => {
    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    const searchInput = screen.getByPlaceholderText('Search your library...')
    expect(searchInput).toBeInTheDocument()
  })

  it('renders the media type filter dropdown', () => {
    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    expect(screen.getByText('All Types')).toBeInTheDocument()
    expect(screen.getByText('Movies')).toBeInTheDocument()
    expect(screen.getByText('TV Shows')).toBeInTheDocument()
  })

  it('renders sort options dropdown', () => {
    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    expect(screen.getByText('Recently Updated')).toBeInTheDocument()
  })

  it('renders grid and list view toggle buttons', () => {
    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    expect(screen.getByTestId('icon-grid')).toBeInTheDocument()
    expect(screen.getByTestId('icon-list')).toBeInTheDocument()
  })

  it('shows no media found when search returns empty', async () => {
    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('No media found')).toBeInTheDocument()
    })

    expect(screen.getByText('Try adjusting your search or filters')).toBeInTheDocument()
  })

  it('displays media items when data is returned', async () => {
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [
        {
          id: 1,
          title: 'Action Movie',
          media_type: 'movie',
          year: 2023,
          rating: 8.5,
          directory_path: '/media/movies',
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
        {
          id: 2,
          title: 'Drama Film',
          media_type: 'movie',
          year: 2022,
          directory_path: '/media/movies',
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
      ],
      total: 2,
      limit: 50,
      offset: 0,
    })

    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Action Movie')).toBeInTheDocument()
    })

    expect(screen.getByText('Drama Film')).toBeInTheDocument()
    expect(screen.getByText('Showing 2 of 2 items')).toBeInTheDocument()
  })

  it('updates search input value when user types', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    const searchInput = screen.getByPlaceholderText('Search your library...')
    await user.type(searchInput, 'test query')

    expect(searchInput).toHaveValue('test query')
  })

  it('calls searchMedia on mount', () => {
    render(
      <TestWrapper>
        <LibraryPage />
      </TestWrapper>
    )

    expect(apiService.searchMedia).toHaveBeenCalled()
  })
})
