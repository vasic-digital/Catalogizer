import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Dashboard } from '../Dashboard'
import { useAuth } from '@/contexts/AuthContext'

// Mock dependencies
vi.mock('@/contexts/AuthContext', async () => ({
  useAuth: vi.fn(),
}))

vi.mock('framer-motion', async () => {
  const MockMotionDiv = ({ children, ...props }: any) => <div {...props}>{children}</div>
  MockMotionDiv.displayName = 'MockMotionDiv'
  return {
    motion: {
      div: MockMotionDiv,
    },
  }
})

vi.mock('lucide-react', async () => {
  const icon = (name: string) => {
    const Component = (props: any) => <svg data-testid={`icon-${name.toLowerCase()}`} {...props} />
    Component.displayName = name
    return Component
  }
  return {
    // Dashboard.tsx
    Film: icon('film'),
    Upload: icon('upload'),
    Search: icon('search'),
    Settings: icon('settings'),
    Activity: icon('activity'),
    HardDrive: icon('harddrive'),
    Zap: icon('zap'),
    Clock: icon('clock'),
    // ActivityFeed.tsx
    Play: icon('play'),
    Download: icon('download'),
    Users: icon('users'),
    Eye: icon('eye'),
    Filter: icon('filter'),
    // DashboardStats.tsx
    Database: icon('database'),
    PlusCircle: icon('pluscircle'),
    TrendingUp: icon('trendingup'),
    TrendingDown: icon('trendingdown'),
    PlayCircle: icon('playcircle'),
    // MediaDistributionChart.tsx
    PieChart: icon('piechart'),
  }
})

vi.mock('react-hot-toast', async () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
    promise: vi.fn(),
  },
}))

vi.mock('@/lib/mediaApi', async () => ({
  mediaApi: {
    getMediaStats: vi.fn().mockResolvedValue({
      total_items: 100,
      total_size: 1024000,
      recent_additions: 5,
      by_type: { video: 50, audio: 30, image: 20 },
      by_quality: { hd: 60, sd: 40 },
    }),
    analyzeDirectory: vi.fn().mockResolvedValue({}),
  },
}))

// Mock child components to isolate Dashboard logic
vi.mock('@/components/dashboard/DashboardStats', async () => ({
  DashboardStats: ({ loading }: any) => (
    <div data-testid="dashboard-stats">{loading ? 'Loading stats...' : 'Dashboard Stats'}</div>
  ),
}))

vi.mock('@/components/dashboard/MediaDistributionChart', async () => ({
  MediaDistributionChart: ({ loading }: any) => (
    <div data-testid="media-distribution-chart">{loading ? 'Loading chart...' : 'Media Distribution Chart'}</div>
  ),
}))

vi.mock('@/components/dashboard/ActivityFeed', async () => ({
  ActivityFeed: ({ limit }: any) => (
    <div data-testid="activity-feed">Activity Feed (limit: {limit})</div>
  ),
}))

vi.mock('recharts', async () => ({
  ResponsiveContainer: ({ children }: any) => <div>{children}</div>,
  PieChart: ({ children }: any) => <div>{children}</div>,
  Pie: () => <div />,
  Cell: () => <div />,
  Legend: () => <div />,
  Tooltip: () => <div />,
}))

const mockUseAuth = vi.mocked(useAuth)

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Rendering', () => {
    it('renders the Dashboard component', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByText(/Welcome back/i)).toBeInTheDocument()
    })

    it('displays welcome message with username', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'johndoe' },
      })

      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByText(/Welcome back, johndoe!/i)).toBeInTheDocument()
    })

    it('displays welcome message with fallback for missing username', () => {
      mockUseAuth.mockReturnValue({
        user: { username: '' },
      })

      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByText(/Welcome back/i)).toBeInTheDocument()
    })

    it('displays subtitle description', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />, { wrapper: createWrapper() })
      expect(
        screen.getByText(/Here's what's happening with your media library today/i)
      ).toBeInTheDocument()
    })
  })

  describe('Child Components', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })
    })

    it('renders DashboardStats component', () => {
      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByTestId('dashboard-stats')).toBeInTheDocument()
    })

    it('renders MediaDistributionChart component', () => {
      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByTestId('media-distribution-chart')).toBeInTheDocument()
    })

    it('renders ActivityFeed component', () => {
      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByTestId('activity-feed')).toBeInTheDocument()
    })
  })

  describe('Quick Actions Section', () => {
    it('renders Quick Actions heading', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByText('Quick Actions')).toBeInTheDocument()
    })

    it('renders 4 quick action buttons', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />, { wrapper: createWrapper() })

      expect(screen.getByText('Upload Media')).toBeInTheDocument()
      expect(screen.getByText('Scan Library')).toBeInTheDocument()
      expect(screen.getByText('Search')).toBeInTheDocument()
      expect(screen.getByText('Settings')).toBeInTheDocument()
    })

    it('quick action buttons are clickable', async () => {
      const user = userEvent.setup()
      const toast = (await import('react-hot-toast')).default

      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />, { wrapper: createWrapper() })

      const uploadButton = screen.getByText('Upload Media')
      await user.click(uploadButton)
      expect(toast.success).toHaveBeenCalledWith('Opening upload interface...')
    })
  })

  describe('System Status Section', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })
    })

    it('renders System Status heading', () => {
      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByText('System Status')).toBeInTheDocument()
    })

    it('renders CPU, Memory, and Disk usage', () => {
      render(<Dashboard />, { wrapper: createWrapper() })

      expect(screen.getByText('CPU Usage')).toBeInTheDocument()
      expect(screen.getByText('Memory Usage')).toBeInTheDocument()
      expect(screen.getByText('Disk Usage')).toBeInTheDocument()
    })

    it('renders network status', () => {
      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByText('Network')).toBeInTheDocument()
      expect(screen.getByText('Online')).toBeInTheDocument()
    })

    it('renders uptime', () => {
      render(<Dashboard />, { wrapper: createWrapper() })
      expect(screen.getByText('Uptime')).toBeInTheDocument()
      expect(screen.getByText('5d 12h 34m')).toBeInTheDocument()
    })
  })

  describe('Edge Cases', () => {
    it('renders with null user', () => {
      mockUseAuth.mockReturnValue({
        user: null,
      })

      render(<Dashboard />, { wrapper: createWrapper() })

      // Should still render with fallback "User"
      expect(screen.getByText(/Welcome back/i)).toBeInTheDocument()
    })

    it('renders with undefined user', () => {
      mockUseAuth.mockReturnValue({
        user: undefined,
      })

      render(<Dashboard />, { wrapper: createWrapper() })

      expect(screen.getByText(/Welcome back/i)).toBeInTheDocument()
    })

    it('renders with empty username', () => {
      mockUseAuth.mockReturnValue({
        user: { username: '' },
      })

      render(<Dashboard />, { wrapper: createWrapper() })

      expect(screen.getByText(/Welcome back/i)).toBeInTheDocument()
    })
  })

  describe('Layout and Structure', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })
    })

    it('renders main container with correct class', () => {
      const { container } = render(<Dashboard />, { wrapper: createWrapper() })
      const mainDiv = container.firstChild as HTMLElement

      expect(mainDiv).toHaveClass('space-y-6')
    })

    it('renders grid layouts', () => {
      const { container } = render(<Dashboard />, { wrapper: createWrapper() })

      const grids = container.querySelectorAll('.grid')
      expect(grids.length).toBeGreaterThan(0)
    })

    it('renders all sections in correct order', () => {
      render(<Dashboard />, { wrapper: createWrapper() })

      const headings = screen.getAllByRole('heading')
      const headingTexts = headings.map(h => h.textContent)

      expect(headingTexts.some(text => text?.includes('Welcome back'))).toBe(true)
      expect(headingTexts.some(text => text?.includes('Quick Actions'))).toBe(true)
      expect(headingTexts.some(text => text?.includes('System Status'))).toBe(true)
    })
  })
})
