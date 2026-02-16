import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CollectionSettings } from '../CollectionSettings'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, ...props }: any) => <div className={className}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
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
  onSave: vi.fn(),
}

describe('CollectionSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders heading', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('Collection Settings')).toBeInTheDocument()
  })

  it('displays collection name', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('Test Collection')).toBeInTheDocument()
  })

  it('renders sidebar navigation tabs', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('Display')).toBeInTheDocument()
    expect(screen.getByText('Behavior')).toBeInTheDocument()
    expect(screen.getByText('Playback')).toBeInTheDocument()
    expect(screen.getByText('Download')).toBeInTheDocument()
    expect(screen.getByText('Sharing')).toBeInTheDocument()
    expect(screen.getByText('Privacy')).toBeInTheDocument()
  })

  it('shows Display settings by default', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('View Settings')).toBeInTheDocument()
    expect(screen.getByText('Default View')).toBeInTheDocument()
    expect(screen.getByText('Items Per Page')).toBeInTheDocument()
  })

  it('shows Grid and List view options', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('Grid')).toBeInTheDocument()
    expect(screen.getByText('List')).toBeInTheDocument()
  })

  it('shows Display Options toggles', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('Show Thumbnails')).toBeInTheDocument()
    expect(screen.getByText('Show Metadata')).toBeInTheDocument()
    expect(screen.getByText('Compact View')).toBeInTheDocument()
  })

  it('renders Save Settings button', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('Save Settings')).toBeInTheDocument()
  })

  it('renders Cancel button', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('shows "All changes saved" initially', () => {
    render(<CollectionSettings {...defaultProps} />)
    expect(screen.getByText('All changes saved')).toBeInTheDocument()
  })

  it('switches to Playback tab', async () => {
    const user = userEvent.setup()
    render(<CollectionSettings {...defaultProps} />)

    await user.click(screen.getByText('Playback'))

    expect(screen.getByText('Auto Play Next Item')).toBeInTheDocument()
    expect(screen.getByText('Loop Playback')).toBeInTheDocument()
    expect(screen.getByText('Remember Playback Position')).toBeInTheDocument()
    expect(screen.getByText('Shuffle by Default')).toBeInTheDocument()
  })

  it('switches to Privacy tab', async () => {
    const user = userEvent.setup()
    render(<CollectionSettings {...defaultProps} />)

    await user.click(screen.getByText('Privacy'))

    expect(screen.getByText('Privacy Controls')).toBeInTheDocument()
    expect(screen.getByText('Private Collection')).toBeInTheDocument()
    expect(screen.getByText('Require Password')).toBeInTheDocument()
  })

  it('switches to Download tab', async () => {
    const user = userEvent.setup()
    render(<CollectionSettings {...defaultProps} />)

    await user.click(screen.getByText('Download'))

    expect(screen.getByText('Download Options')).toBeInTheDocument()
    expect(screen.getByText('Default Format')).toBeInTheDocument()
    expect(screen.getByText('Download Quality')).toBeInTheDocument()
  })

  it('calls onClose when Cancel is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionSettings {...defaultProps} />)

    await user.click(screen.getByText('Cancel'))

    expect(defaultProps.onClose).toHaveBeenCalled()
  })
})
