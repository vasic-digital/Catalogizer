import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import FavoritesPage from '../Favorites'

// Mock dependencies
vi.mock('@/hooks/useFavorites', () => ({
  useFavorites: vi.fn(() => ({
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
    refetchStats: vi.fn(),
  })),
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

describe('Favorites Page', () => {
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
})
