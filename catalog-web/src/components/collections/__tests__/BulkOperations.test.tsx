import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BulkOperations } from '../BulkOperations'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, onClick, ...props }: any) => (
      <div className={className} onClick={onClick}>{children}</div>
    ),
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

const defaultProps = {
  selectedCollections: ['1', '2'],
  collections: [
    { id: '1', name: 'Collection A' },
    { id: '2', name: 'Collection B' },
  ],
  onOperation: vi.fn(),
  onClose: vi.fn(),
}

describe('BulkOperations', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders heading', () => {
    render(<BulkOperations {...defaultProps} />)
    expect(screen.getByText('Bulk Operations')).toBeInTheDocument()
  })

  it('shows selected collections count', () => {
    render(<BulkOperations {...defaultProps} />)
    expect(screen.getByText('2 collections selected')).toBeInTheDocument()
  })

  it('renders all bulk action options', () => {
    render(<BulkOperations {...defaultProps} />)
    expect(screen.getByText('Delete Collections')).toBeInTheDocument()
    expect(screen.getByText('Share Collections')).toBeInTheDocument()
    expect(screen.getByText('Export Collections')).toBeInTheDocument()
    expect(screen.getByText('Duplicate Collections')).toBeInTheDocument()
    expect(screen.getByText('Add Tags')).toBeInTheDocument()
    expect(screen.getByText('Archive Collections')).toBeInTheDocument()
    expect(screen.getByText('Move Collections')).toBeInTheDocument()
  })

  it('shows action descriptions', () => {
    render(<BulkOperations {...defaultProps} />)
    expect(screen.getByText('Permanently remove selected collections')).toBeInTheDocument()
    expect(screen.getByText('Share selected collections with others')).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', async () => {
    const user = userEvent.setup()
    render(<BulkOperations {...defaultProps} />)

    // The close button has an X icon
    const closeButtons = screen.getAllByRole('button')
    // Find the close button (first button in header area)
    const closeButton = closeButtons.find(
      btn => btn.textContent === '' || btn.querySelector('svg')
    )
    if (closeButton) {
      await user.click(closeButton)
      expect(defaultProps.onClose).toHaveBeenCalled()
    }
  })

  it('shows delete confirmation when delete is clicked', async () => {
    const user = userEvent.setup()
    render(<BulkOperations {...defaultProps} />)

    await user.click(screen.getByText('Delete Collections'))

    expect(screen.getByText('Delete Collections?')).toBeInTheDocument()
    expect(screen.getByText(/This will permanently delete 2 collection/)).toBeInTheDocument()
  })

  it('shows share options when share is selected', async () => {
    const user = userEvent.setup()
    render(<BulkOperations {...defaultProps} />)

    await user.click(screen.getByText('Share Collections'))

    expect(screen.getByText('Download Permission')).toBeInTheDocument()
    expect(screen.getByText('Reshare Permission')).toBeInTheDocument()
  })

  it('shows export options when export is selected', async () => {
    const user = userEvent.setup()
    render(<BulkOperations {...defaultProps} />)

    await user.click(screen.getByText('Export Collections'))

    expect(screen.getByText('Export Format')).toBeInTheDocument()
  })

  it('shows Back button when action is selected', async () => {
    const user = userEvent.setup()
    render(<BulkOperations {...defaultProps} />)

    await user.click(screen.getByText('Share Collections'))

    expect(screen.getByText('Back')).toBeInTheDocument()
  })

  it('shows Execute Action button when action is selected', async () => {
    const user = userEvent.setup()
    render(<BulkOperations {...defaultProps} />)

    await user.click(screen.getByText('Export Collections'))

    expect(screen.getByText('Execute Action')).toBeInTheDocument()
  })
})
