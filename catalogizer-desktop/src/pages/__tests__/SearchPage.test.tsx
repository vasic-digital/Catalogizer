import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import SearchPage from '../SearchPage'

// Mock apiService
vi.mock('../../services/apiService', () => ({
  apiService: {
    searchMedia: vi.fn(),
  },
}))

// Mock react-router-dom navigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

import { apiService } from '../../services/apiService'

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  return <BrowserRouter>{children}</BrowserRouter>
}

describe('SearchPage', () => {
  it('renders the page title and description', () => {
    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    expect(screen.getByText('Search')).toBeInTheDocument()
    expect(screen.getByText('Search across your entire media library')).toBeInTheDocument()
  })

  it('renders the search input with placeholder', () => {
    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    const searchInput = screen.getByPlaceholderText('Search for movies, TV shows, music...')
    expect(searchInput).toBeInTheDocument()
    expect(searchInput).toHaveValue('')
  })

  it('renders media type filter buttons', () => {
    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    expect(screen.getByText('All')).toBeInTheDocument()
    expect(screen.getByText('Movies')).toBeInTheDocument()
    expect(screen.getByText('Music')).toBeInTheDocument()
    expect(screen.getByText('Images')).toBeInTheDocument()
    expect(screen.getByText('Documents')).toBeInTheDocument()
  })

  it('shows empty state message when no search has been performed', () => {
    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    expect(screen.getByText('Enter a search term to find media in your library')).toBeInTheDocument()
  })

  it('updates search input value when user types', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    const searchInput = screen.getByPlaceholderText('Search for movies, TV shows, music...')
    await user.type(searchInput, 'test query')

    expect(searchInput).toHaveValue('test query')
  })

  it('performs a debounced search when user types a query', async () => {
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
      limit: 50,
      offset: 0,
    })

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    const searchInput = screen.getByPlaceholderText('Search for movies, TV shows, music...')
    await user.type(searchInput, 'Test Movie')

    await waitFor(() => {
      expect(apiService.searchMedia).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(screen.getByText('1 result found')).toBeInTheDocument()
      expect(screen.getByText('Test Movie')).toBeInTheDocument()
    })
  })

  it('displays multiple results with correct count', async () => {
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [
        {
          id: 1,
          title: 'Movie One',
          media_type: 'movie',
          year: 2022,
          directory_path: '/media/movies',
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
        {
          id: 2,
          title: 'Movie Two',
          media_type: 'movie',
          year: 2023,
          directory_path: '/media/movies',
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
      ],
      total: 2,
      limit: 50,
      offset: 0,
    })

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    const searchInput = screen.getByPlaceholderText('Search for movies, TV shows, music...')
    await user.type(searchInput, 'Movie')

    await waitFor(() => {
      expect(screen.getByText('2 results found')).toBeInTheDocument()
      expect(screen.getByText('Movie One')).toBeInTheDocument()
      expect(screen.getByText('Movie Two')).toBeInTheDocument()
    })
  })

  it('displays media type and year for results', async () => {
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [
        {
          id: 1,
          title: 'A Great Film',
          media_type: 'movie',
          year: 2021,
          directory_path: '/media/movies',
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
      ],
      total: 1,
      limit: 50,
      offset: 0,
    })

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    await user.type(screen.getByPlaceholderText('Search for movies, TV shows, music...'), 'Great')

    await waitFor(() => {
      expect(screen.getByText('movie (2021)')).toBeInTheDocument()
    })
  })

  it('shows no results message when search returns empty', async () => {
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [],
      total: 0,
      limit: 50,
      offset: 0,
    })

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    const searchInput = screen.getByPlaceholderText('Search for movies, TV shows, music...')
    await user.type(searchInput, 'nonexistent')

    await waitFor(() => {
      expect(screen.getByText('0 results found')).toBeInTheDocument()
    })
  })

  it('shows error message when search fails', async () => {
    vi.mocked(apiService.searchMedia).mockRejectedValue(new Error('Network error'))

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    const searchInput = screen.getByPlaceholderText('Search for movies, TV shows, music...')
    await user.type(searchInput, 'broken search')

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument()
    })
  })

  it('navigates to media detail when a result is clicked', async () => {
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [
        {
          id: 42,
          title: 'Clickable Movie',
          media_type: 'movie',
          directory_path: '/media/movies',
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
      ],
      total: 1,
      limit: 50,
      offset: 0,
    })

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    await user.type(screen.getByPlaceholderText('Search for movies, TV shows, music...'), 'Clickable')

    await waitFor(() => {
      expect(screen.getByText('Clickable Movie')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Clickable Movie'))

    expect(mockNavigate).toHaveBeenCalledWith('/media/42')
  })

  it('displays quality badge when result has quality info', async () => {
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [
        {
          id: 1,
          title: 'HD Movie',
          media_type: 'movie',
          quality: '1080p',
          directory_path: '/media/movies',
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
      ],
      total: 1,
      limit: 50,
      offset: 0,
    })

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    await user.type(screen.getByPlaceholderText('Search for movies, TV shows, music...'), 'HD')

    await waitFor(() => {
      expect(screen.getByText('1080p')).toBeInTheDocument()
    })
  })

  it('filters by media type when a filter button is clicked', async () => {
    vi.mocked(apiService.searchMedia).mockResolvedValue({
      items: [],
      total: 0,
      limit: 50,
      offset: 0,
    })

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SearchPage />
      </TestWrapper>
    )

    // Type a query first so search triggers
    await user.type(screen.getByPlaceholderText('Search for movies, TV shows, music...'), 'test')

    await waitFor(() => {
      expect(apiService.searchMedia).toHaveBeenCalled()
    })

    vi.mocked(apiService.searchMedia).mockClear()

    // Click Movies filter
    await user.click(screen.getByText('Movies'))

    await waitFor(() => {
      expect(apiService.searchMedia).toHaveBeenCalledWith(
        expect.objectContaining({
          media_type: 'movie',
        })
      )
    })
  })
})
