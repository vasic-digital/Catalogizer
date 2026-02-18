import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { EntityBrowser } from '../EntityBrowser'
import type { MediaTypeInfo, MediaEntity } from '@/types/media'

// Mock the mediaApi
vi.mock('@/lib/mediaApi', () => ({
  entityApi: {
    getEntityTypes: vi.fn().mockResolvedValue({
      types: [
        { id: 1, name: 'movie', description: 'Movies', count: 42 },
        { id: 2, name: 'tv_show', description: 'TV Shows', count: 15 },
        { id: 3, name: 'music_artist', description: 'Music Artists', count: 8 },
      ],
    }),
    getEntityStats: vi.fn().mockResolvedValue({
      total_entities: 65,
      by_type: { movie: 42, tv_show: 15, music_artist: 8 },
    }),
    getEntities: vi.fn().mockResolvedValue({
      items: [
        { id: 1, title: 'The Matrix', year: 1999, status: 'movie', genre: ['Sci-Fi'] },
      ],
      total: 1,
      limit: 24,
      offset: 0,
    }),
    browseByType: vi.fn().mockResolvedValue({
      items: [
        { id: 1, title: 'The Matrix', year: 1999, status: 'movie', genre: ['Sci-Fi'] },
      ],
      total: 1,
      limit: 24,
      offset: 0,
      type: 'movie',
    }),
  },
}))

// Mock entity sub-components
vi.mock('@/components/entity/TypeSelector', () => ({
  TypeSelectorGrid: ({ types, onSelect }: { types: MediaTypeInfo[]; onSelect: (type: string) => void }) => (
    <div data-testid="type-selector-grid">
      {types.map((type) => (
        <button key={type.id} onClick={() => onSelect(type.name)} data-testid={`type-card-${type.name}`}>
          <span>{type.name.replace(/_/g, ' ')}</span>
          <span>{type.count} {type.count === 1 ? 'item' : 'items'}</span>
        </button>
      ))}
    </div>
  ),
  TYPE_ICONS: {},
  TYPE_COLORS: {},
}))

vi.mock('@/components/entity/EntityGrid', () => ({
  EntityGrid: ({ entities, total, limit, offset, page, onEntityClick, onPageChange }: any) => (
    <div data-testid="entity-grid">
      {entities.length === 0 ? (
        <p>No entities found</p>
      ) : (
        entities.map((entity: MediaEntity) => (
          <div key={entity.id} data-testid={`entity-item-${entity.id}`}>
            <span>{entity.title}</span>
            <button onClick={() => onEntityClick(entity)}>View</button>
          </div>
        ))
      )}
      <span>Showing {offset + 1}-{Math.min(offset + limit, total)} of {total}</span>
    </div>
  ),
}))

// Mock UI components
vi.mock('@/components/ui/Button', async () => ({
  Button: ({ children, onClick, disabled, variant, size, ...rest }: any) => (
    <button onClick={onClick} disabled={disabled} {...rest}>
      {children}
    </button>
  ),
  buttonVariants: () => '',
}))

function renderWithProviders(ui: React.ReactElement, { route = '/browse' } = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>
        {ui}
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('EntityBrowser', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the Browse Media heading', async () => {
    renderWithProviders(<EntityBrowser />)
    await waitFor(() => {
      expect(screen.getByText('Browse Media')).toBeInTheDocument()
    })
  })

  it('shows entity type cards when types load', async () => {
    renderWithProviders(<EntityBrowser />)
    await waitFor(() => {
      expect(screen.getByText('movie')).toBeInTheDocument()
    })
    expect(screen.getByText('42 items')).toBeInTheDocument()
  })

  it('shows tv show type card with correct count', async () => {
    renderWithProviders(<EntityBrowser />)
    await waitFor(() => {
      expect(screen.getByText('tv show')).toBeInTheDocument()
    })
    expect(screen.getByText('15 items')).toBeInTheDocument()
  })

  it('shows music artist type card with correct count', async () => {
    renderWithProviders(<EntityBrowser />)
    await waitFor(() => {
      expect(screen.getByText('music artist')).toBeInTheDocument()
    })
    expect(screen.getByText('8 items')).toBeInTheDocument()
  })

  it('shows total entity count from stats', async () => {
    renderWithProviders(<EntityBrowser />)
    await waitFor(() => {
      expect(screen.getByText('65 total entities across all types')).toBeInTheDocument()
    })
  })

  it('has a search input', () => {
    renderWithProviders(<EntityBrowser />)
    expect(screen.getByPlaceholderText('Search entities...')).toBeInTheDocument()
  })

  it('does not show entity grid on initial load (no type or search selected)', async () => {
    renderWithProviders(<EntityBrowser />)
    await waitFor(() => {
      expect(screen.getByText('Browse Media')).toBeInTheDocument()
    })
    // On initial load without a type or search query, the entity list is not shown
    expect(screen.queryByTestId('entity-grid')).not.toBeInTheDocument()
  })

  it('shows the type selector grid on initial load', async () => {
    renderWithProviders(<EntityBrowser />)
    await waitFor(() => {
      expect(screen.getByTestId('type-selector-grid')).toBeInTheDocument()
    })
  })

  it('renders all three type cards', async () => {
    renderWithProviders(<EntityBrowser />)
    await waitFor(() => {
      expect(screen.getByTestId('type-card-movie')).toBeInTheDocument()
      expect(screen.getByTestId('type-card-tv_show')).toBeInTheDocument()
      expect(screen.getByTestId('type-card-music_artist')).toBeInTheDocument()
    })
  })
})
