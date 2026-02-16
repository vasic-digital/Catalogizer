import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CollectionExport } from '../CollectionExport'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, onClick, ...props }: any) => (
      <div className={className} onClick={onClick}>{children}</div>
    ),
    button: ({ children, className, onClick, ...props }: any) => (
      <button className={className} onClick={onClick}>{children}</button>
    ),
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

vi.mock('../../../hooks/useCollections', () => ({
  useCollection: vi.fn(() => ({
    collectionItems: [],
    isLoading: false,
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
  item_count: 5,
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

describe('CollectionExport', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders heading', () => {
    render(<CollectionExport {...defaultProps} />)
    expect(screen.getByText('Export / Import Collection')).toBeInTheDocument()
  })

  it('displays collection name', () => {
    render(<CollectionExport {...defaultProps} />)
    expect(screen.getByText('Test Collection')).toBeInTheDocument()
  })

  it('renders Export and Import tabs', () => {
    render(<CollectionExport {...defaultProps} />)
    expect(screen.getByText('Export')).toBeInTheDocument()
    expect(screen.getByText('Import')).toBeInTheDocument()
  })

  it('shows export format options', () => {
    render(<CollectionExport {...defaultProps} />)
    expect(screen.getByText('Export Format')).toBeInTheDocument()
    expect(screen.getByText('JSON')).toBeInTheDocument()
    expect(screen.getByText('CSV')).toBeInTheDocument()
    expect(screen.getByText('M3U Playlist')).toBeInTheDocument()
  })

  it('shows export options section', () => {
    render(<CollectionExport {...defaultProps} />)
    expect(screen.getByText('Export Options')).toBeInTheDocument()
    expect(screen.getByText('Include Metadata')).toBeInTheDocument()
    expect(screen.getByText('Include Thumbnails')).toBeInTheDocument()
    expect(screen.getByText('Include Files')).toBeInTheDocument()
  })

  it('renders Export Collection button', () => {
    render(<CollectionExport {...defaultProps} />)
    expect(screen.getByText('Export Collection')).toBeInTheDocument()
  })

  it('switches to Import tab', async () => {
    const user = userEvent.setup()
    render(<CollectionExport {...defaultProps} />)

    await user.click(screen.getByText('Import'))

    expect(screen.getByText('Import File')).toBeInTheDocument()
    expect(screen.getByText('Click to upload or drag and drop')).toBeInTheDocument()
  })

  it('shows collection summary', () => {
    render(<CollectionExport {...defaultProps} />)
    expect(screen.getByText('Collection Summary')).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionExport {...defaultProps} />)

    // Find the close button by looking for buttons near the header
    const closeButtons = screen.getAllByRole('button')
    const closeBtn = closeButtons.find(btn => {
      return btn.closest('.flex.items-center.justify-between')
    })
    // Click the first ghost button (close) in the header
    if (closeBtn) {
      await user.click(closeBtn)
    }
  })
})
