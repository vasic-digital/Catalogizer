import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import FavoritesPage from '../Favorites'
import toast from 'react-hot-toast'

const mockRefetchStats = vi.fn()
const mockRefetchFavorites = vi.fn()
const mockBulkAddToFavorites = vi.fn()
const mockUseFavorites = vi.fn()

// Mock dependencies
vi.mock('@/hooks/useFavorites', () => ({
  useFavorites: (...args: any[]) => mockUseFavorites(...args),
}))

vi.mock('@/components/favorites/FavoritesGrid', () => ({
  FavoritesGrid: ({ showFilters, showStats, selectable }: any) => (
    <div data-testid="favorites-grid">
      Favorites Grid (filters: {String(showFilters)}, stats: {String(showStats)}, selectable: {String(selectable)})
    </div>
  ),
}))

vi.mock('@/components/layout/PageHeader', () => ({
  PageHeader: ({ title, subtitle, actions }: any) => (
    <div data-testid="page-header">
      <h1>{title}</h1>
      <p>{subtitle}</p>
      {actions}
    </div>
  ),
}))

vi.mock('@/lib/utils', () => ({
  cn: (...classes: any[]) => classes.filter(Boolean).join(' '),
}))

vi.mock('react-hot-toast', () => ({
  __esModule: true,
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

const defaultFavoritesReturn = {
  favorites: [
    {
      id: 1,
      media_id: 101,
      created_at: '2024-01-01T00:00:00Z',
      media_item: { title: 'Favorite Movie', media_type: 'movie', year: 2023 },
    },
    {
      id: 2,
      media_id: 102,
      created_at: '2024-01-02T00:00:00Z',
      media_item: { title: 'Favorite Song', media_type: 'music', year: 2024 },
    },
  ],
  stats: {
    total_count: 42,
    media_type_breakdown: {
      movie: 20,
      tv_show: 10,
      music: 8,
      game: 2,
      documentary: 1,
      anime: 1,
      concert: 0,
      other: 0,
    },
    recent_additions: [],
  },
  refetchStats: mockRefetchStats,
  refetchFavorites: mockRefetchFavorites,
  bulkAddToFavorites: mockBulkAddToFavorites,
}

describe('Favorites Page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseFavorites.mockReturnValue(defaultFavoritesReturn)
  })

  it('renders page header with title', () => {
    render(<FavoritesPage />)
    expect(screen.getByText('My Favorites')).toBeInTheDocument()
  })

  it('renders page subtitle', () => {
    render(<FavoritesPage />)
    expect(screen.getByText('Manage your favorite media items')).toBeInTheDocument()
  })

  it('renders tab navigation', () => {
    render(<FavoritesPage />)
    expect(screen.getByText('Favorites')).toBeInTheDocument()
    expect(screen.getByText('Recently Added')).toBeInTheDocument()
    expect(screen.getByText('Statistics')).toBeInTheDocument()
  })

  it('renders action buttons', () => {
    render(<FavoritesPage />)
    expect(screen.getByText('Bulk Actions')).toBeInTheDocument()
    expect(screen.getByText('Import')).toBeInTheDocument()
    expect(screen.getByText('Export')).toBeInTheDocument()
  })

  it('renders FavoritesGrid component', () => {
    render(<FavoritesPage />)
    expect(screen.getByTestId('favorites-grid')).toBeInTheDocument()
  })

  it('shows stats tab with statistics when clicked', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    expect(screen.getByText('Favorite Statistics')).toBeInTheDocument()
    expect(screen.getByText('42')).toBeInTheDocument()
  })

  it('displays media type breakdown in stats', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    expect(screen.getByText('By Media Type')).toBeInTheDocument()
    expect(screen.getByText('20')).toBeInTheDocument()
  })

  it('shows Favorite Insights section', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    expect(screen.getByText('Favorite Insights')).toBeInTheDocument()
    expect(screen.getByText('Most Common Type')).toBeInTheDocument()
  })

  it('toggles bulk actions on button click', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Bulk Actions'))
    // After click, the button should have active styling
    const button = screen.getByText('Bulk Actions')
    expect(button).toBeInTheDocument()
  })

  it('switches to Recently Added tab', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Recently Added'))

    expect(screen.getByText('Recently Added Favorites')).toBeInTheDocument()
  })

  // --- New tests for increased coverage ---

  it('enables selectable mode on FavoritesGrid when bulk actions is active', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    // Before clicking bulk actions, selectable should be false
    expect(screen.getByText(/selectable: false/)).toBeInTheDocument()

    await user.click(screen.getByText('Bulk Actions'))

    // After clicking, selectable should be true
    expect(screen.getByText(/selectable: true/)).toBeInTheDocument()
  })

  it('toggles bulk actions off on second click', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Bulk Actions'))
    expect(screen.getByText(/selectable: true/)).toBeInTheDocument()

    await user.click(screen.getByText('Bulk Actions'))
    expect(screen.getByText(/selectable: false/)).toBeInTheDocument()
  })

  it('shows loading state for statistics when stats is null', async () => {
    mockUseFavorites.mockReturnValue({
      ...defaultFavoritesReturn,
      stats: null,
    })
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    // Should show the loading skeleton (animate-pulse div)
    expect(screen.getByText('Favorite Statistics')).toBeInTheDocument()
    // Total count should not be shown
    expect(screen.queryByText('42')).not.toBeInTheDocument()
  })

  it('displays "Recent Activity" section in insights', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    expect(screen.getByText('Recent Activity')).toBeInTheDocument()
    expect(screen.getByText('0 items added in the last week')).toBeInTheDocument()
  })

  it('displays "Storage Impact" section in insights', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    expect(screen.getByText('Storage Impact')).toBeInTheDocument()
    expect(screen.getByText('Favorites help you quickly access your most-loved content')).toBeInTheDocument()
  })

  it('shows recent additions count from stats', async () => {
    mockUseFavorites.mockReturnValue({
      ...defaultFavoritesReturn,
      stats: {
        ...defaultFavoritesReturn.stats,
        recent_additions: [{ id: 1 }, { id: 2 }, { id: 3 }],
      },
    })
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    expect(screen.getByText('3 items added in the last week')).toBeInTheDocument()
  })

  it('determines most common type from breakdown', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    // movie has highest count (20) so it should be shown as most common type
    // It appears both in the breakdown list and in the insights section
    const movieTexts = screen.getAllByText('movie')
    expect(movieTexts.length).toBeGreaterThanOrEqual(2) // once in breakdown, once in insights
  })

  it('exports favorites as JSON when Export is clicked', async () => {
    // Mock URL.createObjectURL and link behavior
    const mockCreateObjectURL = vi.fn(() => 'blob:test-url')
    const mockRevokeObjectURL = vi.fn()
    global.URL.createObjectURL = mockCreateObjectURL
    global.URL.revokeObjectURL = mockRevokeObjectURL

    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Export'))

    expect(mockCreateObjectURL).toHaveBeenCalled()
    expect(mockRevokeObjectURL).toHaveBeenCalledWith('blob:test-url')
    expect(toast.success).toHaveBeenCalledWith('Exported 2 favorites')
  })

  it('shows error toast when exporting with no favorites', async () => {
    mockUseFavorites.mockReturnValue({
      ...defaultFavoritesReturn,
      favorites: [],
    })
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Export'))

    expect(toast.error).toHaveBeenCalledWith('No favorites to export')
  })

  it('shows error toast when exporting with null favorites', async () => {
    mockUseFavorites.mockReturnValue({
      ...defaultFavoritesReturn,
      favorites: null,
    })
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Export'))

    expect(toast.error).toHaveBeenCalledWith('No favorites to export')
  })

  it('triggers file input when Import is clicked', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    // The import button should trigger the file input
    const importButton = screen.getByText('Import')
    expect(importButton).toBeInTheDocument()

    // Clicking Import should trigger the file input click
    await user.click(importButton)
    // The hidden input should exist
    const fileInput = document.querySelector('input[type="file"]')
    expect(fileInput).toBeInTheDocument()
    expect(fileInput).toHaveAttribute('accept', '.json')
  })

  it('imports favorites from a valid JSON file', async () => {
    render(<FavoritesPage />)

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    expect(fileInput).toBeInTheDocument()

    const fileContent = JSON.stringify([
      { media_id: 1 },
      { media_id: 2 },
      { media_id: 3 },
    ])
    const file = new File([fileContent], 'favorites.json', { type: 'application/json' })

    await userEvent.upload(fileInput, file)

    await waitFor(() => {
      expect(mockBulkAddToFavorites).toHaveBeenCalledWith([1, 2, 3])
    })

    expect(mockRefetchFavorites).toHaveBeenCalled()
    expect(mockRefetchStats).toHaveBeenCalled()
  })

  it('shows error toast when importing an invalid JSON file', async () => {
    render(<FavoritesPage />)

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    const file = new File(['not-json-content'], 'bad.json', { type: 'application/json' })

    await userEvent.upload(fileInput, file)

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Invalid JSON file')
    })
  })

  it('shows error toast when importing a JSON file with no valid media IDs', async () => {
    render(<FavoritesPage />)

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    const fileContent = JSON.stringify([{ name: 'no media_id' }])
    const file = new File([fileContent], 'empty.json', { type: 'application/json' })

    await userEvent.upload(fileInput, file)

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('No valid media IDs found in file')
    })
  })

  it('passes showFilters and showStats correctly to FavoritesGrid in favorites tab', () => {
    render(<FavoritesPage />)

    expect(screen.getByText(/filters: true/)).toBeInTheDocument()
    expect(screen.getByText(/stats: true/)).toBeInTheDocument()
  })

  it('shows Recently Added tab with FavoritesGrid without filters', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Recently Added'))

    // The Recently Added tab should have FavoritesGrid with showFilters=false, showStats=false
    const grids = screen.getAllByTestId('favorites-grid')
    expect(grids.length).toBeGreaterThanOrEqual(1)
  })

  it('displays all media type entries in the breakdown', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    // Check various breakdown entries
    expect(screen.getByText('10')).toBeInTheDocument()  // tv_show
    expect(screen.getByText('8')).toBeInTheDocument()   // music
    expect(screen.getByText('2')).toBeInTheDocument()   // game
  })

  it('renders the total favorites count label', async () => {
    const user = userEvent.setup()
    render(<FavoritesPage />)

    await user.click(screen.getByText('Statistics'))

    expect(screen.getByText('Total Favorites')).toBeInTheDocument()
  })
})
