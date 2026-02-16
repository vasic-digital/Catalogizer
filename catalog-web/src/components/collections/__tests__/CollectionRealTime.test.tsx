import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CollectionRealTime } from '../CollectionRealTime'

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
    collectionItems: [],
    refetchItems: vi.fn(),
  })),
}))

vi.mock('react-hot-toast', () => ({
  toast: vi.fn(),
  default: vi.fn(),
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

describe('CollectionRealTime', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders heading', () => {
    render(<CollectionRealTime collection={mockCollection as any} />)
    expect(screen.getByText('Real-time Collection')).toBeInTheDocument()
  })

  it('displays collection name', () => {
    render(<CollectionRealTime collection={mockCollection as any} />)
    expect(screen.getByText('Test Collection')).toBeInTheDocument()
  })

  it('shows connection status', async () => {
    render(<CollectionRealTime collection={mockCollection as any} />)
    // Initially connecting
    expect(screen.getByText('Connecting')).toBeInTheDocument()
  })

  it('renders activity stats section', async () => {
    render(<CollectionRealTime collection={mockCollection as any} />)

    await waitFor(() => {
      expect(screen.getByText('Activity Stats')).toBeInTheDocument()
    })
    expect(screen.getByText('Active Users')).toBeInTheDocument()
    expect(screen.getByText('Total Interactions')).toBeInTheDocument()
  })

  it('renders notifications section', () => {
    render(<CollectionRealTime collection={mockCollection as any} />)
    expect(screen.getByText('Notifications')).toBeInTheDocument()
  })

  it('shows online users heading', () => {
    render(<CollectionRealTime collection={mockCollection as any} />)
    // Initially shows 0 online users before connection
    expect(screen.getByText(/Online Users/)).toBeInTheDocument()
  })

  it('shows recent events heading', () => {
    render(<CollectionRealTime collection={mockCollection as any} />)
    expect(screen.getByText('Recent Events')).toBeInTheDocument()
  })

  it('renders connection panel', () => {
    render(<CollectionRealTime collection={mockCollection as any} />)
    expect(screen.getByText('Connection')).toBeInTheDocument()
  })
})
