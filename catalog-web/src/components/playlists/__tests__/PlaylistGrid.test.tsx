import { render, screen } from '@testing-library/react'
import { PlaylistGrid } from '../PlaylistGrid'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, ...props }: any) => <div className={className}>{children}</div>,
    button: ({ children, className, onClick, ...props }: any) => (
      <button className={className} onClick={onClick}>{children}</button>
    ),
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

vi.mock('../../../hooks/usePlaylists', () => ({
  usePlaylists: vi.fn(() => ({
    playlists: [
      {
        id: '1',
        name: 'My Playlist',
        description: 'A test playlist',
        is_public: false,
        primary_media_type: 'music',
        items: [],
        item_count: 5,
        created_at: '2024-01-01T00:00:00Z',
      },
    ],
    isLoading: false,
    error: null,
    refetchPlaylists: vi.fn(),
  })),
}))

vi.mock('../../../hooks/useFavorites', () => ({
  useFavorites: vi.fn(() => ({
    checkFavoriteStatus: vi.fn(() => false),
  })),
}))

vi.mock('../PlaylistManager', () => ({
  PlaylistManager: () => <div data-testid="playlist-manager">Manager</div>,
}))

vi.mock('../PlaylistPlayer', () => ({
  PlaylistPlayer: () => <div data-testid="playlist-player">Player</div>,
}))

vi.mock('../PlaylistItem', () => ({
  PlaylistItemComponent: () => <div>Playlist Item</div>,
}))

vi.mock('../../../lib/playlistsApi', () => ({
  playlistApi: {
    deletePlaylist: vi.fn(),
    duplicatePlaylist: vi.fn(),
  },
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

describe('PlaylistGrid', () => {
  it('renders playlist items', () => {
    render(<PlaylistGrid />)
    expect(screen.getByText('My Playlist')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<PlaylistGrid />)
    expect(screen.getByPlaceholderText('Search playlists...')).toBeInTheDocument()
  })

  it('shows item count for playlists', () => {
    render(<PlaylistGrid />)
    expect(screen.getByText(/5 items/)).toBeInTheDocument()
  })

  it('renders view mode toggle', () => {
    render(<PlaylistGrid />)
    // Grid/List view toggles exist
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('renders sort controls', () => {
    render(<PlaylistGrid />)
    // Sort by dropdown
    const selects = screen.getAllByRole('combobox')
    expect(selects.length).toBeGreaterThan(0)
  })
})
