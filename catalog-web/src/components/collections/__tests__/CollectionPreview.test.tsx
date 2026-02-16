import { render, screen } from '@testing-library/react'
import { CollectionPreview } from '../CollectionPreview'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, onClick, ...props }: any) => (
      <div className={className} onClick={onClick}>{children}</div>
    ),
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

vi.mock('../../../hooks/useCollections', () => ({
  useCollection: vi.fn(() => ({
    collectionItems: {
      items: [
        {
          id: '1',
          title: 'Song A',
          media_type: 'music',
          file_size: 5000000,
          duration: 240,
          date_added: '2024-01-01T00:00:00Z',
          rating: 4,
        },
        {
          id: '2',
          title: 'Video B',
          media_type: 'video',
          file_size: 1000000000,
          duration: 3600,
          date_added: '2024-01-02T00:00:00Z',
          rating: 5,
        },
      ],
      total: 2,
    },
    isLoadingItems: false,
  })),
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

const mockCollection = {
  id: '1',
  name: 'Test Collection',
  description: 'A test collection',
  is_smart: true,
  smart_rules: [],
  item_count: 2,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  is_public: false,
  primary_media_type: 'music',
  owner_id: 'user1',
}

const defaultProps = {
  collection: mockCollection as any,
  onClose: vi.fn(),
}

describe('CollectionPreview', () => {
  it('renders collection name', () => {
    render(<CollectionPreview {...defaultProps} />)
    expect(screen.getByText('Test Collection')).toBeInTheDocument()
  })

  it('displays total items count', () => {
    render(<CollectionPreview {...defaultProps} />)
    expect(screen.getByText('2 items total')).toBeInTheDocument()
  })

  it('displays collection items', () => {
    render(<CollectionPreview {...defaultProps} />)
    expect(screen.getByText('Song A')).toBeInTheDocument()
    expect(screen.getByText('Video B')).toBeInTheDocument()
  })

  it('shows loading state when isLoading is true', () => {
    render(<CollectionPreview {...defaultProps} isLoading={true} />)
    // Should show a loading indicator (pulse animation)
    const pulseElement = document.querySelector('.animate-pulse')
    expect(pulseElement).toBeTruthy()
  })

  it('renders close button', () => {
    render(<CollectionPreview {...defaultProps} />)
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })
})
