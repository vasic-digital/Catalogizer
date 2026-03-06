import { render } from '@testing-library/react'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, onClick, ..._props }: any) => (
      <div className={className} onClick={onClick}>{children}</div>
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

describe('CollectionAnalytics', () => {
  it('exports CollectionAnalytics component', async () => {
    const mod = await import('../CollectionAnalytics')
    expect(mod.CollectionAnalytics).toBeDefined()
    expect(typeof mod.CollectionAnalytics).toBe('function')
  })

  it('component has correct display name or is a valid React component', async () => {
    const mod = await import('../CollectionAnalytics')
    const Component = mod.CollectionAnalytics
    // Verify it is a function (React component)
    expect(typeof Component).toBe('function')
  })

  it('renders CollectionAnalytics without initialization errors', async () => {
    const mod = await import('../CollectionAnalytics')
    const CollectionAnalytics = mod.CollectionAnalytics

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

    // formatFileSize bug is fixed - component renders without errors
    expect(() => {
      render(<CollectionAnalytics collection={mockCollection as any} />)
    }).not.toThrow()
  })
})
