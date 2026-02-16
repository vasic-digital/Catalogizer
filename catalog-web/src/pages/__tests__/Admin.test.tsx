import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Admin } from '../Admin'

// Mock the adminApi
vi.mock('@/lib/adminApi', () => ({
  adminApi: {
    getSystemInfo: vi.fn(() =>
      Promise.resolve({
        version: '1.0.0',
        uptime: 86400,
        cpuUsage: 45,
        memoryUsage: 62,
        diskUsage: { total: 1073741824000, used: 536870912000, free: 536870912000 },
        activeConnections: 12,
        totalRequests: 15420,
      })
    ),
    getUsers: vi.fn(() =>
      Promise.resolve([
        { id: '1', username: 'admin', email: 'admin@test.com', role: 'admin', status: 'active', createdAt: '2023-01-01' },
      ])
    ),
    getStorageInfo: vi.fn(() =>
      Promise.resolve([
        { path: '/media/movies', totalSpace: 1073741824000, usedSpace: 536870912000, availableSpace: 536870912000, mediaCount: 1250 },
      ])
    ),
    getBackups: vi.fn(() =>
      Promise.resolve([
        { id: '1', filename: 'backup.tar.gz', size: 1073741824, createdAt: '2024-01-01', type: 'full', status: 'completed' },
      ])
    ),
    createBackup: vi.fn(() => Promise.resolve()),
    restoreBackup: vi.fn(() => Promise.resolve()),
    scanStorage: vi.fn(() => Promise.resolve()),
    updateUser: vi.fn(() => Promise.resolve()),
  },
}))

// Mock react-hot-toast
vi.mock('react-hot-toast', () => ({
  __esModule: true,
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('Admin Page', () => {
  it('renders Admin Panel heading', async () => {
    render(<Admin />, { wrapper: createWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Admin Panel')).toBeInTheDocument()
    })
  })

  it('renders overview tab with system stats', async () => {
    render(<Admin />, { wrapper: createWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Overview')).toBeInTheDocument()
    })
  })

  it('shows navigation tabs', async () => {
    render(<Admin />, { wrapper: createWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Users')).toBeInTheDocument()
      expect(screen.getByText('Storage')).toBeInTheDocument()
      expect(screen.getByText('Backups')).toBeInTheDocument()
    })
  })

  it('displays system version when data loads', async () => {
    render(<Admin />, { wrapper: createWrapper() })
    await waitFor(() => {
      expect(screen.getByText('1.0.0')).toBeInTheDocument()
    })
  })

  it('shows default values when data is loading', () => {
    render(<Admin />, { wrapper: createWrapper() })
    // Default values are shown immediately from the fallback props
    expect(screen.getByText('Admin Panel')).toBeInTheDocument()
  })

  it('renders quick action buttons', async () => {
    render(<Admin />, { wrapper: createWrapper() })
    await waitFor(() => {
      expect(screen.getByText('Create Full Backup')).toBeInTheDocument()
      expect(screen.getByText('Create Incremental Backup')).toBeInTheDocument()
    })
  })

  it('renders resource usage section', async () => {
    render(<Admin />, { wrapper: createWrapper() })
    await waitFor(() => {
      expect(screen.getByText('System Resources')).toBeInTheDocument()
      expect(screen.getByText('CPU Usage')).toBeInTheDocument()
      expect(screen.getByText('Memory Usage')).toBeInTheDocument()
    })
  })

  it('wraps content in max-width container', () => {
    const { container } = render(<Admin />, { wrapper: createWrapper() })
    const wrapper = container.firstChild as HTMLElement
    expect(wrapper).toHaveClass('max-w-7xl')
  })
})
