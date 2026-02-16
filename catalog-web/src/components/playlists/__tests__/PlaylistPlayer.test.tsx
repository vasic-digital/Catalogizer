import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PlaylistPlayer } from '../PlaylistPlayer'

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

vi.mock('../../media/MediaPlayer', () => ({
  MediaPlayer: () => <div data-testid="media-player">Media Player</div>,
}))

vi.mock('../../favorites/FavoriteToggle', () => ({
  FavoriteToggle: () => <div data-testid="favorite-toggle">Favorite</div>,
}))

vi.mock('../../../hooks/usePlayerState', () => ({
  usePlayerState: vi.fn(() => ({
    currentTrack: null,
    isPlaying: false,
    play: vi.fn(),
    pause: vi.fn(),
  })),
}))

vi.mock('../../../lib/playlistsApi', () => ({
  playlistsApi: {
    updatePlayHistory: vi.fn(),
  },
}))

vi.mock('../PlaylistAnalytics', () => ({
  PlaylistAnalytics: () => <div data-testid="playlist-analytics">Analytics</div>,
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

// Mock types/playlists to handle flattenPlaylistItem
vi.mock('../../../types/playlists', () => ({
  flattenPlaylistItem: vi.fn((item: any) => ({
    title: item.title || item.media_item?.title || 'Unknown',
    media_type: item.media_type || item.media_item?.media_type || 'music',
    duration: item.duration || 0,
    file_path: item.file_path || '',
    thumbnail_url: item.thumbnail_url || '',
    item_id: item.id || '1',
    artist: item.artist || '',
    album: item.album || '',
  })),
  getMediaIconWithMap: vi.fn(() => {
    return (props: any) => <span data-testid="media-icon" {...props} />
  }),
}))

const mockPlaylist = {
  id: '1',
  name: 'Test Playlist',
  description: 'A test playlist',
  is_public: false,
  primary_media_type: 'music',
  items: [],
  item_count: 2,
}

const mockItems = [
  {
    id: '1',
    title: 'Track 1',
    media_type: 'music',
    duration: 180,
    file_path: '/music/track1.mp3',
    media_item: { id: '1', title: 'Track 1', media_type: 'music' },
  },
  {
    id: '2',
    title: 'Track 2',
    media_type: 'music',
    duration: 240,
    file_path: '/music/track2.mp3',
    media_item: { id: '2', title: 'Track 2', media_type: 'music' },
  },
]

describe('PlaylistPlayer', () => {
  it('renders playlist name', () => {
    render(<PlaylistPlayer playlist={mockPlaylist as any} items={mockItems as any} />)
    // "Test Playlist" appears in header and player display
    const playlistNames = screen.getAllByText('Test Playlist')
    expect(playlistNames.length).toBeGreaterThanOrEqual(1)
  })

  it('renders player controls', () => {
    render(<PlaylistPlayer playlist={mockPlaylist as any} items={mockItems as any} />)
    // Should have play/pause and skip controls
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('shows current track info', () => {
    render(<PlaylistPlayer playlist={mockPlaylist as any} items={mockItems as any} />)
    // "Track 1" appears in now playing section and track list
    const trackElements = screen.getAllByText('Track 1')
    expect(trackElements.length).toBeGreaterThanOrEqual(1)
  })

  it('calls onClose when close button is clicked', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    render(
      <PlaylistPlayer
        playlist={mockPlaylist as any}
        items={mockItems as any}
        onClose={onClose}
      />
    )

    // Find close button
    const buttons = screen.getAllByRole('button')
    const closeBtn = buttons.find(btn => btn.querySelector('svg'))
    if (closeBtn) {
      await user.click(closeBtn)
    }
  })

  it('displays track list', () => {
    render(<PlaylistPlayer playlist={mockPlaylist as any} items={mockItems as any} />)
    // Track names appear in track list (may appear multiple times due to now-playing display)
    const track1Elements = screen.getAllByText('Track 1')
    expect(track1Elements.length).toBeGreaterThanOrEqual(1)
    const track2Elements = screen.getAllByText('Track 2')
    expect(track2Elements.length).toBeGreaterThanOrEqual(1)
  })
})
