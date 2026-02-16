import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BundleAnalyzer } from '../BundleAnalyzer'

vi.mock('lucide-react', () => ({
  RefreshCw: () => <span data-testid="icon-refresh">Refresh</span>,
  Download: () => <span data-testid="icon-download">Download</span>,
  Copy: () => <span data-testid="icon-copy">Copy</span>,
}))

// Mock URL.createObjectURL and URL.revokeObjectURL
const mockCreateObjectURL = vi.fn().mockReturnValue('blob:test-url')
const mockRevokeObjectURL = vi.fn()
global.URL.createObjectURL = mockCreateObjectURL
global.URL.revokeObjectURL = mockRevokeObjectURL

describe('BundleAnalyzer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state initially', () => {
    render(<BundleAnalyzer />)
    expect(screen.getByText('Analyzing bundle...')).toBeInTheDocument()
  })

  it('renders the analysis after loading', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Bundle Analysis')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )
  })

  it('displays total bundle size', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Total Bundle Size')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    // The total size is calculated from the mock data - multiple elements contain "MB"
    const sizeTexts = screen.getAllByText(/MB/)
    expect(sizeTexts.length).toBeGreaterThan(0)
  })

  it('displays potential savings', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Potential Savings')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    expect(screen.getByText('900 KB')).toBeInTheDocument()
  })

  it('displays total bundles count', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Total Bundles')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('displays JS bundles count', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('JS Bundles')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    expect(screen.getByText('4')).toBeInTheDocument()
  })

  it('renders bundle breakdown section', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Bundle Breakdown')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )
  })

  it('shows individual bundle names', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('main')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    expect(screen.getByText('vendor')).toBeInTheDocument()
    expect(screen.getByText('collections')).toBeInTheDocument()
    expect(screen.getByText('main.css')).toBeInTheDocument()
    expect(screen.getByText('components')).toBeInTheDocument()
  })

  it('shows bundle sizes', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('1.85 MB')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    expect(screen.getByText('3.2 MB')).toBeInTheDocument()
    expect(screen.getByText('450 KB')).toBeInTheDocument()
    expect(screen.getByText('185 KB')).toBeInTheDocument()
    expect(screen.getByText('620 KB')).toBeInTheDocument()
  })

  it('shows bundle types (JS and CSS)', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        const jsBadges = screen.getAllByText('JS')
        expect(jsBadges.length).toBe(4)
      },
      { timeout: 3000 }
    )

    expect(screen.getByText('CSS')).toBeInTheDocument()
  })

  it('shows bundle paths', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(
          screen.getByText('Path: /static/js/main.js')
        ).toBeInTheDocument()
      },
      { timeout: 3000 }
    )
  })

  it('shows optimization recommendations', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(
          screen.getByText('Optimization Recommendations')
        ).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    expect(
      screen.getByText(/Consider lazy loading CollectionTemplates/)
    ).toBeInTheDocument()
  })

  it('shows Show Details toggle', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Show Details')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )
  })

  it('toggles chunk analysis details', async () => {
    const user = userEvent.setup()
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Show Details')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    await user.click(screen.getByText('Show Details'))

    expect(screen.getByText('Chunk Analysis')).toBeInTheDocument()
    expect(screen.getByText('collection-templates')).toBeInTheDocument()
    expect(screen.getByText('advanced-search')).toBeInTheDocument()
    expect(screen.getByText('280 KB')).toBeInTheDocument()

    expect(screen.getByText('Hide Details')).toBeInTheDocument()
  })

  it('shows chunk module names in detail view', async () => {
    const user = userEvent.setup()
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Show Details')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    await user.click(screen.getByText('Show Details'))

    expect(
      screen.getByText('Modules: CollectionTemplates, TemplatePreview, TemplateCard')
    ).toBeInTheDocument()
  })

  it('shows chunk parent names in detail view', async () => {
    const user = userEvent.setup()
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Show Details')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    await user.click(screen.getByText('Show Details'))

    // Multiple chunks share the same parents "collections, main"
    const parentElements = screen.getAllByText('Parents: collections, main')
    expect(parentElements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders Optimize Bundle button', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Optimize Bundle')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )
  })

  it('renders Export Report button', async () => {
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        const exportBtns = screen.getAllByText('Export Report')
        expect(exportBtns.length).toBeGreaterThan(0)
      },
      { timeout: 3000 }
    )
  })

  it('copies stats to clipboard', async () => {
    const writeTextSpy = vi.fn().mockResolvedValue(undefined)
    // Use vi.stubGlobal to properly mock clipboard for jsdom
    const originalClipboard = navigator.clipboard
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: writeTextSpy, readText: vi.fn() },
      writable: true,
      configurable: true,
    })

    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Bundle Analysis')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    const copyBtn = screen.getByTitle('Copy Stats')
    // Use fireEvent instead of userEvent to avoid userEvent's clipboard interception
    const { fireEvent } = await import('@testing-library/react')
    fireEvent.click(copyBtn)

    expect(writeTextSpy).toHaveBeenCalledWith(
      expect.stringContaining('Bundle Analysis Report')
    )

    // Restore
    Object.defineProperty(navigator, 'clipboard', {
      value: originalClipboard,
      writable: true,
      configurable: true,
    })
  })

  it('exports report as JSON', async () => {
    const mockClick = vi.fn()
    const mockCreateElement = vi.spyOn(document, 'createElement')

    const user = userEvent.setup()
    render(<BundleAnalyzer />)

    await waitFor(
      () => {
        expect(screen.getByText('Bundle Analysis')).toBeInTheDocument()
      },
      { timeout: 3000 }
    )

    const exportBtn = screen.getByTitle('Export Report')
    await user.click(exportBtn)

    expect(mockCreateObjectURL).toHaveBeenCalled()
    mockCreateElement.mockRestore()
  })
})
