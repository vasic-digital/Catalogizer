import { render, screen } from '@testing-library/react'
import { CollectionSharing } from '../CollectionSharing'

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

describe('CollectionSharing', () => {
  it('renders heading', () => {
    render(<CollectionSharing collection={mockCollection as any} onClose={vi.fn()} />)
    expect(screen.getByText('Share Collection')).toBeInTheDocument()
  })

  it('displays collection name', () => {
    render(<CollectionSharing collection={mockCollection as any} onClose={vi.fn()} />)
    expect(screen.getByText('Test Collection')).toBeInTheDocument()
  })

  it('renders Create Share Link section', () => {
    render(<CollectionSharing collection={mockCollection as any} onClose={vi.fn()} />)
    // "Create Share Link" appears as heading and button
    const shareLinkElements = screen.getAllByText('Create Share Link')
    expect(shareLinkElements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders share buttons', () => {
    render(<CollectionSharing collection={mockCollection as any} onClose={vi.fn()} />)
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })
})
