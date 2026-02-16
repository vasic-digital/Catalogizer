import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Collections } from '../Collections'

// Mock all heavy dependencies
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

vi.mock('../../hooks/useCollections', () => ({
  useCollections: vi.fn(() => ({
    collections: [
      {
        id: '1',
        name: 'Test Collection',
        description: 'A test collection',
        is_smart: true,
        smart_rules: [],
        item_count: 42,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        is_public: false,
        primary_media_type: 'music',
        owner_id: 'user1',
      },
    ],
    isLoading: false,
    error: null,
    refetchCollections: vi.fn(),
    createCollection: vi.fn(),
    updateCollection: vi.fn(),
    deleteCollection: vi.fn(),
    shareCollection: vi.fn(),
    duplicateCollection: vi.fn(),
    exportCollection: vi.fn(),
    bulkDeleteCollections: vi.fn(),
    bulkShareCollections: vi.fn(),
    bulkExportCollections: vi.fn(),
    bulkUpdateCollections: vi.fn(),
    isSharing: false,
    isDuplicating: false,
    isExporting: false,
  })),
}))

vi.mock('../../components/collections/SmartCollectionBuilder', () => ({
  SmartCollectionBuilder: ({ onSave, onCancel }: any) => (
    <div data-testid="smart-builder">Smart Collection Builder</div>
  ),
}))

vi.mock('../../components/collections/CollectionPreview', () => ({
  CollectionPreview: () => <div data-testid="collection-preview">Collection Preview</div>,
}))

vi.mock('../../components/collections/BulkOperations', () => ({
  BulkOperations: () => <div data-testid="bulk-ops">Bulk Operations</div>,
}))

vi.mock('../../components/collections/PerformanceOptimizer', () => ({
  PerformanceOptimizer: ({ children }: any) => <div>{children}</div>,
}))

vi.mock('../../components/collections/CollectionSettings', () => ({
  CollectionSettings: () => <div>Collection Settings</div>,
}))

vi.mock('../../components/collections/CollectionAnalytics', () => ({
  CollectionAnalytics: () => <div>Collection Analytics</div>,
}))

vi.mock('../../components/collections/CollectionSharing', () => ({
  CollectionSharing: () => <div>Collection Sharing</div>,
}))

vi.mock('../../components/collections/CollectionExport', () => ({
  CollectionExport: () => <div>Collection Export</div>,
}))

vi.mock('../../components/collections/CollectionRealTime', () => ({
  CollectionRealTime: () => <div>Collection RealTime</div>,
}))

vi.mock('../../components/performance/LazyComponents', () => ({
  ComponentLoader: ({ children }: any) => <>{children}</>,
  preloadComponent: vi.fn(),
  CollectionTemplates: () => <div>Collection Templates</div>,
  AdvancedSearch: () => <div>Advanced Search</div>,
  CollectionAutomation: () => <div>Collection Automation</div>,
  ExternalIntegrations: () => <div>External Integrations</div>,
  CollectionAnalytics: () => <div>Lazy Analytics</div>,
}))

vi.mock('../../components/performance/VirtualScroller', () => ({
  VirtualList: ({ children }: any) => <div>{children}</div>,
  VirtualizedTable: ({ children }: any) => <div>{children}</div>,
}))

vi.mock('../../components/performance/MemoCache', () => ({
  useMemoized: (fn: any, deps: any[]) => fn(),
  useOptimizedData: (data: any) => data,
  usePagination: (data: any, pageSize: number) => ({
    page: 1,
    paginatedData: data,
    totalPages: 1,
    nextPage: vi.fn(),
    prevPage: vi.fn(),
    goToPage: vi.fn(),
  }),
}))

vi.mock('../../components/performance/BundleAnalyzer', () => ({
  BundleAnalyzer: () => <div>Bundle Analyzer</div>,
}))

vi.mock('../../components/ai/AIComponents', () => ({
  AICollectionSuggestions: () => <div>AI Suggestions</div>,
  AINaturalSearch: () => <div>AI Search</div>,
  AIContentCategorizer: () => <div>AI Categorizer</div>,
}))

vi.mock('../../components/ai/AIAnalytics', () => ({
  AIUserBehaviorAnalytics: () => <div>AI Behavior</div>,
  AIPredictions: () => <div>AI Predictions</div>,
  AISmartOrganization: () => <div>AI Organization</div>,
}))

vi.mock('../../components/ai/AIMetadata', () => ({
  AIMetadataExtractor: () => <div>AI Metadata</div>,
  AIAutomationRules: () => <div>AI Rules</div>,
  AIContentQualityAnalyzer: () => <div>AI Quality</div>,
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

describe('Collections Page', () => {
  it('renders page heading', () => {
    render(<Collections />)
    expect(screen.getByText('Collections')).toBeInTheDocument()
  })

  it('renders page description', () => {
    render(<Collections />)
    expect(
      screen.getByText('Organize your media with smart and manual collections')
    ).toBeInTheDocument()
  })

  it('renders tab navigation', () => {
    render(<Collections />)
    expect(screen.getByText('All Collections')).toBeInTheDocument()
    expect(screen.getByText('Smart Collections')).toBeInTheDocument()
    // "Templates" appears as both a tab label and a button
    const templatesElements = screen.getAllByText('Templates')
    expect(templatesElements.length).toBeGreaterThanOrEqual(1)
  })

  it('displays collection items', () => {
    render(<Collections />)
    expect(screen.getByText('Test Collection')).toBeInTheDocument()
    expect(screen.getByText('42 items')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<Collections />)
    expect(screen.getByPlaceholderText('Search collections...')).toBeInTheDocument()
  })

  it('renders action buttons', () => {
    render(<Collections />)
    expect(screen.getByText('Smart Collection')).toBeInTheDocument()
    // "Templates" appears multiple times (tab + button)
    const templatesElements = screen.getAllByText('Templates')
    expect(templatesElements.length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Advanced Search')).toBeInTheDocument()
  })

  it('shows AI Features button', () => {
    render(<Collections />)
    // "AI Features" appears as both a tab and a button
    const aiElements = screen.getAllByText('AI Features')
    expect(aiElements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders select all checkbox', () => {
    render(<Collections />)
    expect(screen.getByText(/Select all/)).toBeInTheDocument()
  })

  it('shows empty state when no collections match filter', async () => {
    const { useCollections } = await import('../../hooks/useCollections')
    vi.mocked(useCollections).mockReturnValue({
      collections: [],
      isLoading: false,
      error: null,
      refetchCollections: vi.fn(),
      createCollection: vi.fn(),
      updateCollection: vi.fn(),
      deleteCollection: vi.fn(),
      shareCollection: vi.fn(),
      duplicateCollection: vi.fn(),
      exportCollection: vi.fn(),
      bulkDeleteCollections: vi.fn(),
      bulkShareCollections: vi.fn(),
      bulkExportCollections: vi.fn(),
      bulkUpdateCollections: vi.fn(),
      isSharing: false,
      isDuplicating: false,
      isExporting: false,
    } as any)

    render(<Collections />)
    expect(screen.getByText('No collections found')).toBeInTheDocument()
  })
})
