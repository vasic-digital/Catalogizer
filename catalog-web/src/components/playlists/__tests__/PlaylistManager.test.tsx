import { render, screen } from '@testing-library/react'
import { PlaylistManager } from '../PlaylistManager'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, ...props }: any) => <div className={className}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

// Mock DnD Kit
vi.mock('@dnd-kit/core', () => ({
  DndContext: ({ children }: any) => <div>{children}</div>,
  closestCenter: vi.fn(),
  KeyboardSensor: vi.fn(),
  PointerSensor: vi.fn(),
  useSensor: vi.fn(),
  useSensors: vi.fn(() => []),
}))

vi.mock('@dnd-kit/sortable', () => ({
  arrayMove: vi.fn(),
  SortableContext: ({ children }: any) => <div>{children}</div>,
  sortableKeyboardCoordinates: vi.fn(),
  verticalListSortingStrategy: vi.fn(),
}))

vi.mock('../../../hooks/usePlaylists', () => ({
  usePlaylists: vi.fn(() => ({
    playlists: [
      {
        id: '1',
        name: 'My Playlist',
        description: 'Test playlist',
        is_public: false,
        primary_media_type: 'music',
        items: [],
        item_count: 3,
      },
    ],
    isLoading: false,
    error: null,
    refetchPlaylists: vi.fn(),
  })),
}))

vi.mock('../../../hooks/usePlaylistReorder', () => ({
  usePlaylistReorder: vi.fn(() => ({
    reorderItems: vi.fn(),
    isReordering: false,
  })),
}))

vi.mock('../SortablePlaylistItem', () => ({
  SortablePlaylistItem: ({ item }: any) => <div>{item?.title || 'Item'}</div>,
}))

vi.mock('../../../lib/playlistsApi', () => ({
  playlistApi: {
    deletePlaylist: vi.fn(),
  },
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

const defaultProps = {
  onCreatePlaylist: vi.fn(),
  onEditPlaylist: vi.fn(),
  onPlaylistSelect: vi.fn(),
}

describe('PlaylistManager', () => {
  it('renders playlist items', () => {
    render(<PlaylistManager {...defaultProps} />)
    expect(screen.getByText('My Playlist')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<PlaylistManager {...defaultProps} />)
    expect(screen.getByPlaceholderText('Search playlists...')).toBeInTheDocument()
  })

  it('shows item count', () => {
    render(<PlaylistManager {...defaultProps} />)
    expect(screen.getByText(/3 items/)).toBeInTheDocument()
  })

  it('renders view toggle buttons', () => {
    render(<PlaylistManager {...defaultProps} />)
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('renders sort dropdown', () => {
    render(<PlaylistManager {...defaultProps} />)
    const selects = screen.getAllByRole('combobox')
    expect(selects.length).toBeGreaterThan(0)
  })
})
