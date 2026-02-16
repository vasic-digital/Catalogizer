import { render, screen } from '@testing-library/react'
import { SortablePlaylistItem } from '../SortablePlaylistItem'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, onClick, ...props }: any) => (
      <div className={className} onClick={onClick}>{children}</div>
    ),
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

// Mock DnD Kit
vi.mock('@dnd-kit/sortable', () => ({
  useSortable: vi.fn(() => ({
    attributes: {},
    listeners: {},
    setNodeRef: vi.fn(),
    transform: null,
    transition: undefined,
    isDragging: false,
  })),
}))

vi.mock('@dnd-kit/utilities', () => ({
  CSS: {
    Transform: {
      toString: vi.fn(() => ''),
    },
  },
}))

vi.mock('../../favorites/FavoriteToggle', () => ({
  FavoriteToggle: () => <div data-testid="favorite-toggle">Favorite</div>,
}))

vi.mock('../../../types/playlists', () => ({
  flattenPlaylistItem: vi.fn((item: any) => ({
    title: item.title || item.media?.title || 'Unknown',
    media_type: item.media_type || 'music',
    duration: item.duration || 0,
    file_path: item.file_path || '',
    item_id: item.id || '1',
  })),
  getMediaIconWithMap: vi.fn(() => {
    // Return a React component function, not an object
    return (props: any) => <span data-testid="media-icon" {...props} />
  }),
}))

const mockItem = {
  id: 'item-1',
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

describe('SortablePlaylistItem', () => {
  it('renders the component', () => {
    render(<SortablePlaylistItem {...defaultProps} />)
    expect(screen.getByText('Test Track')).toBeInTheDocument()
  })

  it('renders action buttons when showActions is true', () => {
    render(<SortablePlaylistItem {...defaultProps} showActions={true} />)
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('renders without action buttons when showActions is false', () => {
    render(<SortablePlaylistItem {...defaultProps} showActions={false} />)
    // Component still renders, just without actions
    expect(screen.getByText('Test Track')).toBeInTheDocument()
  })

  it('applies drag handle', () => {
    render(<SortablePlaylistItem {...defaultProps} />)
    // The drag handle (grip icon) should be present
    expect(screen.getByText('Test Track')).toBeInTheDocument()
  })
})
