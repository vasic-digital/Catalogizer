import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SmartCollectionBuilder } from '../SmartCollectionBuilder'

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

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

vi.mock('../../../lib/collectionRules', () => ({
  COLLECTION_TEMPLATES: [
    {
      id: 'recent',
      name: 'Recently Added',
      description: 'Items added in the last 30 days',
      icon: 'Clock',
      category: 'Time-based',
      rules: [{ field: 'created_at', operator: 'greater_than', value: '30d' }],
    },
  ],
  COLLECTION_FIELD_OPTIONS: [
    { value: 'media_type', label: 'Media Type' },
    { value: 'year', label: 'Year' },
  ],
  COLLECTION_OPERATORS: {
    text: [
      { value: 'equals', label: 'Equals' },
      { value: 'contains', label: 'Contains' },
    ],
    number: [
      { value: 'equals', label: 'Equals' },
      { value: 'greater_than', label: 'Greater than' },
    ],
  },
  getFieldOptions: vi.fn(() => []),
  getFieldLabel: vi.fn((field: string) => field),
  getFieldType: vi.fn(() => 'text'),
  validateRules: vi.fn(() => ({ isValid: true, errors: [] })),
}))

const defaultProps = {
  onSave: vi.fn(),
  onCancel: vi.fn(),
}

describe('SmartCollectionBuilder', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the component', () => {
    render(<SmartCollectionBuilder {...defaultProps} />)
    expect(screen.getByPlaceholderText('Enter collection name')).toBeInTheDocument()
  })

  it('renders name input field', () => {
    render(<SmartCollectionBuilder {...defaultProps} />)
    const nameInput = screen.getByPlaceholderText('Enter collection name')
    expect(nameInput).toBeInTheDocument()
  })

  it('renders description textarea', () => {
    render(<SmartCollectionBuilder {...defaultProps} />)
    const descInput = screen.getByPlaceholderText('Optional description for this collection')
    expect(descInput).toBeInTheDocument()
  })

  it('renders Create Collection and Cancel buttons', () => {
    render(<SmartCollectionBuilder {...defaultProps} />)
    expect(screen.getByText('Create Collection')).toBeInTheDocument()
    // Cancel button may appear multiple times
    const cancelButtons = screen.getAllByText('Cancel')
    expect(cancelButtons.length).toBeGreaterThanOrEqual(1)
  })

  it('calls onCancel when Cancel is clicked', async () => {
    const user = userEvent.setup()
    render(<SmartCollectionBuilder {...defaultProps} />)

    const cancelButtons = screen.getAllByText('Cancel')
    await user.click(cancelButtons[0])

    expect(defaultProps.onCancel).toHaveBeenCalled()
  })

  it('renders Add Rule button', () => {
    render(<SmartCollectionBuilder {...defaultProps} />)
    expect(screen.getByText('Add Rule')).toBeInTheDocument()
  })

  it('renders Quick Start Templates section', () => {
    render(<SmartCollectionBuilder {...defaultProps} />)
    // Templates section header is visible (collapsed by default)
    expect(screen.getByText('Quick Start Templates')).toBeInTheDocument()
  })

  it('accepts initial values', () => {
    render(
      <SmartCollectionBuilder
        {...defaultProps}
        initialName="Test Smart Collection"
        initialDescription="Test description"
      />
    )
    const nameInput = screen.getByPlaceholderText('Enter collection name') as HTMLInputElement
    expect(nameInput.value).toBe('Test Smart Collection')
  })
})
