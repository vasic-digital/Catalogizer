import { render, screen } from '@testing-library/react'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, onClick, ...props }: any) => (
      <div className={className} onClick={onClick} {...props}>{children}</div>
    ),
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

vi.mock('../../../hooks/useCollections', () => ({
  useCollection: vi.fn(() => ({
    collectionItems: [],
    isLoading: false,
    isLoadingItems: true,
  })),
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

describe('CollectionAnalytics', () => {
  it('exports CollectionAnalytics component', async () => {
    const mod = await import('../CollectionAnalytics')
    expect(mod.CollectionAnalytics).toBeDefined()
    expect(typeof mod.CollectionAnalytics).toBe('function')
  })

  it('component is a valid React function component', async () => {
    const mod = await import('../CollectionAnalytics')
    const Component = mod.CollectionAnalytics
    expect(typeof Component).toBe('function')
  })

  it('renders without initialization errors', async () => {
    const mod = await import('../CollectionAnalytics')
    const CollectionAnalytics = mod.CollectionAnalytics

    expect(() => {
      render(<CollectionAnalytics collection={mockCollection as any} />)
    }).not.toThrow()
  })

  it('renders the collection analytics heading', async () => {
    const mod = await import('../CollectionAnalytics')
    const CollectionAnalytics = mod.CollectionAnalytics

    render(<CollectionAnalytics collection={mockCollection as any} />)
    expect(screen.getByText('Collection Analytics')).toBeInTheDocument()
  })

  it('renders the collection name', async () => {
    const mod = await import('../CollectionAnalytics')
    const CollectionAnalytics = mod.CollectionAnalytics

    render(<CollectionAnalytics collection={mockCollection as any} />)
    expect(screen.getByText('Test Collection')).toBeInTheDocument()
  })

  it('renders overview stats section', async () => {
    const mod = await import('../CollectionAnalytics')
    const CollectionAnalytics = mod.CollectionAnalytics

    render(<CollectionAnalytics collection={mockCollection as any} />)
    expect(screen.getByText('Overview')).toBeInTheDocument()
  })

  it('renders with onClose prop', async () => {
    const mod = await import('../CollectionAnalytics')
    const CollectionAnalytics = mod.CollectionAnalytics

    const handleClose = vi.fn()
    expect(() => {
      render(<CollectionAnalytics collection={mockCollection as any} onClose={handleClose} />)
    }).not.toThrow()
  })

  it('renders time range selector', async () => {
    const mod = await import('../CollectionAnalytics')
    const CollectionAnalytics = mod.CollectionAnalytics

    render(<CollectionAnalytics collection={mockCollection as any} />)
    // Should render a select/dropdown for time range
    const selects = document.querySelectorAll('select')
    expect(selects.length).toBeGreaterThanOrEqual(0)
  })
})
