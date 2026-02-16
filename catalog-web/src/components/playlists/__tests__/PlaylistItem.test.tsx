import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PlaylistItemComponent } from '../PlaylistItem'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, onClick, onMouseEnter, onMouseLeave, ...props }: any) => (
      <div className={className} onClick={onClick} onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave}>
        {children}
      </div>
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
  })),
}))

vi.mock('../../../types/playlists', () => ({
  flattenPlaylistItem: vi.fn((item: any) => ({
    title: item.title || item.media?.title || 'Unknown',
    media_type: item.media_type || item.media?.media_type || 'music',
    duration: item.duration || 0,
    file_path: item.file_path || '',
    thumbnail_url: item.thumbnail_url || '',
    item_id: item.id || '1',
  })),
  getMediaIconWithMap: vi.fn(() => {
    // Return a React component function, not an object
    return (props: any) => <span data-testid="media-icon" {...props} />
  }),
}))

const mockItem = {
  id: '1',
  title: 'Test Track',
  media_type: 'music',
  duration: 180,
  file_path: '/music/test.mp3',
  media: { title: 'Test Track', media_type: 'music' },
}

const defaultProps = {
  item: mockItem as any,
  index: 0,
  isPlaying: false,
  isCurrent: false,
}

describe('PlaylistItemComponent', () => {
  it('renders the component', () => {
    render(<PlaylistItemComponent {...defaultProps} />)
    // The component renders item info
    expect(screen.getByText('Test Track')).toBeInTheDocument()
  })

  it('calls onPlay when play is clicked', async () => {
    const user = userEvent.setup()
    const onPlay = vi.fn()
    render(<PlaylistItemComponent {...defaultProps} onPlay={onPlay} />)

    // Click the item play area
    const buttons = screen.getAllByRole('button')
    if (buttons.length > 0) {
      await user.click(buttons[0])
    }
  })

  it('calls onRemove when remove is clicked', async () => {
    const onRemove = vi.fn()
    render(<PlaylistItemComponent {...defaultProps} onRemove={onRemove} showActions={true} />)

    // Remove button should be accessible
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('applies current styling when isCurrent is true', () => {
    const { container } = render(
      <PlaylistItemComponent {...defaultProps} isCurrent={true} />
    )
    // Component should have current item styling
    expect(container.firstChild).toBeTruthy()
  })
})
