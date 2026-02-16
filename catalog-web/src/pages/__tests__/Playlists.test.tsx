import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PlaylistsPage } from '../Playlists'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

// Mock dependencies
vi.mock('../../components/layout/PageHeader', () => ({
  PageHeader: ({ title, subtitle }: any) => (
    <div data-testid="page-header">
      <h1>{title}</h1>
      <p>{subtitle}</p>
    </div>
  ),
}))

vi.mock('../../components/playlists/PlaylistGrid', () => ({
  PlaylistGrid: ({ onCreatePlaylist, onEditPlaylist }: any) => (
    <div data-testid="playlist-grid">
      <button onClick={onCreatePlaylist}>Grid Create</button>
    </div>
  ),
}))

vi.mock('../../components/playlists/PlaylistManager', () => ({
  PlaylistManager: () => <div data-testid="playlist-manager">Manager</div>,
}))

vi.mock('../../components/playlists/PlaylistPlayer', () => ({
  PlaylistPlayer: ({ onClose }: any) => (
    <div data-testid="playlist-player">
      Player
      <button onClick={onClose}>Close</button>
    </div>
  ),
}))

vi.mock('../../components/playlists/PlaylistItem', () => ({
  PlaylistItemComponent: () => <div>Playlist Item</div>,
}))

vi.mock('../../components/playlists/SmartPlaylistBuilder', () => ({
  SmartPlaylistBuilder: ({ onSave, onCancel }: any) => (
    <div data-testid="smart-builder">
      Smart Playlist Builder
      <button onClick={onCancel}>Cancel Smart</button>
    </div>
  ),
}))

vi.mock('../../lib/playlistsApi', () => ({
  playlistsApi: {
    createPlaylist: vi.fn(),
    updatePlaylist: vi.fn(),
    deletePlaylist: vi.fn(),
    getPlaylistItems: vi.fn(),
  },
}))

vi.mock('../../lib/mediaApi', () => ({
  mediaApi: {
    searchMedia: vi.fn(() => Promise.resolve({ items: [] })),
  },
}))

vi.mock('../../hooks/usePlaylists', () => ({
  usePlaylists: vi.fn(() => ({
    playlists: [
      {
        id: '1',
        name: 'Test Playlist',
        description: 'A test playlist',
        is_public: false,
        primary_media_type: 'music',
        items: [],
        item_count: 5,
      },
    ],
    isLoading: false,
    error: null,
    refetchPlaylists: vi.fn(),
  })),
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

describe('Playlists Page', () => {
  it('renders page header', () => {
    render(<PlaylistsPage />)
    expect(screen.getByText('Playlists')).toBeInTheDocument()
    expect(screen.getByText('Organize and manage your media collections')).toBeInTheDocument()
  })

  it('renders tab navigation', () => {
    render(<PlaylistsPage />)
    expect(screen.getByText('All Playlists')).toBeInTheDocument()
    expect(screen.getByText('My Playlists')).toBeInTheDocument()
    expect(screen.getByText('Public')).toBeInTheDocument()
    expect(screen.getByText('Smart Builder')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<PlaylistsPage />)
    expect(screen.getByPlaceholderText('Search playlists...')).toBeInTheDocument()
  })

  it('renders Create Playlist button', () => {
    render(<PlaylistsPage />)
    expect(screen.getByText('Create Playlist')).toBeInTheDocument()
  })

  it('renders playlist grid', () => {
    render(<PlaylistsPage />)
    expect(screen.getByTestId('playlist-grid')).toBeInTheDocument()
  })

  it('opens create form when Create Playlist is clicked', async () => {
    const user = userEvent.setup()
    render(<PlaylistsPage />)

    await user.click(screen.getByText('Create Playlist'))

    expect(screen.getByText('Create New Playlist')).toBeInTheDocument()
    expect(screen.getByText('Playlist Name *')).toBeInTheDocument()
  })

  it('shows Smart Builder when Smart Builder tab is clicked', async () => {
    const user = userEvent.setup()
    render(<PlaylistsPage />)

    await user.click(screen.getByText('Smart Builder'))

    expect(screen.getByTestId('smart-builder')).toBeInTheDocument()
  })

  it('renders media type filter dropdown', () => {
    render(<PlaylistsPage />)
    const select = screen.getByRole('combobox')
    expect(select).toBeInTheDocument()
  })

  it('renders "Favorites" tab', () => {
    render(<PlaylistsPage />)
    // There's a Favorites tab in the navigation
    expect(screen.getByText('Favorites')).toBeInTheDocument()
  })
})
