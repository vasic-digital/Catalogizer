import React from 'react'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Settings } from '../Settings'

// Mock framer-motion
vi.mock('framer-motion', async () => {
  const MockMotionDiv = ({ children, ...props }: any) => <div {...props}>{children}</div>
  MockMotionDiv.displayName = 'MockMotionDiv'
  return {
    motion: {
      div: MockMotionDiv,
    },
  }
})

// Mock lucide-react icons
vi.mock('lucide-react', async () => {
  const icon = (name: string) => {
    const Component = (props: any) => <svg data-testid={`icon-${name.toLowerCase()}`} {...props} />
    Component.displayName = name
    return Component
  }
  return {
    Scan: icon('scan'),
    Network: icon('network'),
    RefreshCw: icon('refreshcw'),
    Plus: icon('plus'),
    Play: icon('play'),
    Trash2: icon('trash2'),
    CheckCircle: icon('checkcircle'),
    XCircle: icon('xcircle'),
    Clock: icon('clock'),
    Server: icon('server'),
    FolderSearch: icon('foldersearch'),
    Wifi: icon('wifi'),
  }
})

// Mock react-hot-toast
vi.mock('react-hot-toast', async () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

// Mock API modules
vi.mock('@/lib/scansApi', async () => ({
  scansApi: {
    listScans: vi.fn().mockResolvedValue([]),
    queueScan: vi.fn().mockResolvedValue({}),
    getScanStatus: vi.fn().mockResolvedValue({}),
  },
}))

vi.mock('@/lib/smbApi', async () => ({
  smbApi: {
    discover: vi.fn().mockResolvedValue([]),
    testConnection: vi.fn().mockResolvedValue({ success: true, message: 'OK' }),
    browse: vi.fn().mockResolvedValue([]),
  },
}))

vi.mock('@/lib/syncApi', async () => ({
  syncApi: {
    listEndpoints: vi.fn().mockResolvedValue([]),
    createEndpoint: vi.fn().mockResolvedValue({}),
    startSync: vi.fn().mockResolvedValue({}),
    deleteEndpoint: vi.fn().mockResolvedValue(undefined),
    getSessions: vi.fn().mockResolvedValue([]),
    getStatistics: vi.fn().mockResolvedValue({
      total_syncs: 10,
      successful_syncs: 8,
      failed_syncs: 2,
      total_items_synced: 500,
    }),
  },
}))

vi.mock('@/lib/mediaApi', async () => ({
  mediaApi: {
    getStorageRoots: vi.fn().mockResolvedValue([]),
  },
}))

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
  const Wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  Wrapper.displayName = 'TestWrapper'
  return Wrapper
}

describe('Settings', () => {
  describe('Rendering', () => {
    it('renders the Settings page heading', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(screen.getByText('Settings')).toBeInTheDocument()
    })

    it('renders the description text', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(
        screen.getByText('Manage scans, network discovery, and synchronization')
      ).toBeInTheDocument()
    })
  })

  describe('Tabs', () => {
    it('renders Scans tab', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(screen.getByText('Scans')).toBeInTheDocument()
    })

    it('renders SMB Discovery tab', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(screen.getByText('SMB Discovery')).toBeInTheDocument()
    })

    it('renders Sync tab', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(screen.getByText('Sync')).toBeInTheDocument()
    })
  })

  describe('Scans Section (default tab)', () => {
    it('renders Start New Scan card', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(screen.getByText('Start New Scan')).toBeInTheDocument()
    })

    it('renders Storage Root select', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(screen.getByLabelText('Storage Root')).toBeInTheDocument()
    })

    it('renders Start Scan button', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(screen.getByText('Start Scan')).toBeInTheDocument()
    })

    it('renders Scan History card', () => {
      render(<Settings />, { wrapper: createWrapper() })

      expect(screen.getByText('Scan History')).toBeInTheDocument()
    })
  })

  describe('Layout', () => {
    it('renders within a max-width container', () => {
      const { container } = render(<Settings />, { wrapper: createWrapper() })

      const wrapper = container.querySelector('.max-w-7xl')
      expect(wrapper).toBeInTheDocument()
    })
  })
})
