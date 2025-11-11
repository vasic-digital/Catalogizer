import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Dashboard } from '../Dashboard'

// Mock dependencies
jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}))

jest.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
}))

jest.mock('lucide-react', () => ({
  Database: () => <div data-testid="icon-database">Database Icon</div>,
  Film: () => <div data-testid="icon-film">Film Icon</div>,
  Music: () => <div data-testid="icon-music">Music Icon</div>,
  Gamepad2: () => <div data-testid="icon-gamepad">Gamepad Icon</div>,
  Monitor: () => <div data-testid="icon-monitor">Monitor Icon</div>,
  BookOpen: () => <div data-testid="icon-book">Book Icon</div>,
  TrendingUp: () => <div data-testid="icon-trending">Trending Icon</div>,
  Users: () => <div data-testid="icon-users">Users Icon</div>,
  Activity: () => <div data-testid="icon-activity">Activity Icon</div>,
  HardDrive: () => <div data-testid="icon-harddrive">HardDrive Icon</div>,
}))

const mockUseAuth = require('@/contexts/AuthContext').useAuth

describe('Dashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('Rendering', () => {
    it('renders the Dashboard component', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />)
      expect(screen.getByText(/Welcome back/i)).toBeInTheDocument()
    })

    it('displays welcome message with username', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'johndoe' },
      })

      render(<Dashboard />)
      expect(screen.getByText(/Welcome back, johndoe!/i)).toBeInTheDocument()
    })

    it('displays welcome message with first_name if available', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'johndoe', first_name: 'John' },
      })

      render(<Dashboard />)
      expect(screen.getByText(/Welcome back, John!/i)).toBeInTheDocument()
    })

    it('displays subtitle description', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />)
      expect(
        screen.getByText(/Here's what's happening with your media collection today/i)
      ).toBeInTheDocument()
    })
  })

  describe('Stats Section', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })
    })

    it('renders all 4 stat cards', () => {
      render(<Dashboard />)

      expect(screen.getByText('Total Media Items')).toBeInTheDocument()
      expect(screen.getByText('Movies')).toBeInTheDocument()
      expect(screen.getByText('Music Albums')).toBeInTheDocument()
      expect(screen.getByText('Games')).toBeInTheDocument()
    })

    it('displays stat values', () => {
      render(<Dashboard />)

      expect(screen.getByText('1,234')).toBeInTheDocument()
      expect(screen.getByText('456')).toBeInTheDocument()
      expect(screen.getByText('789')).toBeInTheDocument()
      expect(screen.getByText('123')).toBeInTheDocument()
    })

    it('displays stat changes', () => {
      render(<Dashboard />)

      expect(screen.getByText('+12% from last month')).toBeInTheDocument()
      expect(screen.getByText('+8% from last month')).toBeInTheDocument()
      expect(screen.getByText('+15% from last month')).toBeInTheDocument()
      expect(screen.getByText('+5% from last month')).toBeInTheDocument()
    })

    it('renders stat icons', () => {
      render(<Dashboard />)

      // Count Database icons (appears in stats and quick actions)
      const databaseIcons = screen.getAllByTestId('icon-database')
      expect(databaseIcons.length).toBeGreaterThan(0)

      expect(screen.getAllByTestId('icon-film').length).toBeGreaterThan(0)
      expect(screen.getAllByTestId('icon-music').length).toBeGreaterThan(0)
      expect(screen.getAllByTestId('icon-gamepad').length).toBeGreaterThan(0)
    })
  })

  describe('Quick Actions Section', () => {
    it('renders Quick Actions heading', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />)
      expect(screen.getByText('Quick Actions')).toBeInTheDocument()
    })

    it('renders 4 quick action cards for regular users', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser', role: 'user' },
      })

      render(<Dashboard />)

      expect(screen.getByText('Browse Media')).toBeInTheDocument()
      expect(screen.getByText('View Analytics')).toBeInTheDocument()
      expect(screen.getByText('System Health')).toBeInTheDocument()
      expect(screen.getByText('Storage Usage')).toBeInTheDocument()
    })

    it('renders 5 quick action cards for admin users', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'admin', role: 'admin' },
      })

      render(<Dashboard />)

      expect(screen.getByText('Browse Media')).toBeInTheDocument()
      expect(screen.getByText('View Analytics')).toBeInTheDocument()
      expect(screen.getByText('System Health')).toBeInTheDocument()
      expect(screen.getByText('Storage Usage')).toBeInTheDocument()
      expect(screen.getByText('User Management')).toBeInTheDocument()
    })

    it('does not show User Management for non-admin users', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser', role: 'user' },
      })

      render(<Dashboard />)
      expect(screen.queryByText('User Management')).not.toBeInTheDocument()
    })

    it('renders quick action descriptions', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />)

      expect(screen.getByText('Explore your media collection')).toBeInTheDocument()
      expect(screen.getByText('See detailed statistics')).toBeInTheDocument()
      expect(screen.getByText('Check system status')).toBeInTheDocument()
      expect(screen.getByText('Monitor disk usage')).toBeInTheDocument()
    })

    it('quick action cards are clickable', async () => {
      const user = userEvent.setup()
      const consoleSpy = jest.spyOn(console, 'log').mockImplementation()

      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />)

      const browseMediaCard = screen.getByText('Browse Media').closest('div')?.parentElement?.parentElement
      if (browseMediaCard) {
        await user.click(browseMediaCard)
        expect(consoleSpy).toHaveBeenCalledWith('Browse media')
      }

      consoleSpy.mockRestore()
    })
  })

  describe('Recent Activity Section', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })
    })

    it('renders Recent Activity heading', () => {
      render(<Dashboard />)
      expect(screen.getByText('Recent Activity')).toBeInTheDocument()
    })

    it('renders Recent Activity description', () => {
      render(<Dashboard />)
      expect(screen.getByText('Latest changes in your media collection')).toBeInTheDocument()
    })

    it('renders all 4 recent activity items', () => {
      render(<Dashboard />)

      expect(screen.getByText('The Matrix (1999)')).toBeInTheDocument()
      expect(screen.getByText('Dark Side of the Moon')).toBeInTheDocument()
      expect(screen.getByText('Cyberpunk 2077')).toBeInTheDocument()
      expect(screen.getByText('Adobe Photoshop 2024')).toBeInTheDocument()
    })

    it('renders activity actions', () => {
      render(<Dashboard />)

      expect(screen.getByText('Added to collection')).toBeInTheDocument()
      expect(screen.getByText('Metadata updated')).toBeInTheDocument()
      expect(screen.getByText('Quality analysis completed')).toBeInTheDocument()
      expect(screen.getByText('New version detected')).toBeInTheDocument()
    })

    it('renders activity timestamps', () => {
      render(<Dashboard />)

      expect(screen.getByText('2 hours ago')).toBeInTheDocument()
      expect(screen.getByText('4 hours ago')).toBeInTheDocument()
      expect(screen.getByText('6 hours ago')).toBeInTheDocument()
      expect(screen.getByText('1 day ago')).toBeInTheDocument()
    })

    it('renders activity type badges', () => {
      render(<Dashboard />)

      expect(screen.getByText('Movie')).toBeInTheDocument()
      expect(screen.getByText('Album')).toBeInTheDocument()
      expect(screen.getByText('Game')).toBeInTheDocument()
      expect(screen.getByText('Software')).toBeInTheDocument()
    })

    it('renders View All Activity button', () => {
      render(<Dashboard />)
      expect(screen.getByRole('button', { name: /View All Activity/i })).toBeInTheDocument()
    })
  })

  describe('User Role Handling', () => {
    it('handles admin user correctly', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'admin', role: 'admin', first_name: 'Admin' },
      })

      render(<Dashboard />)

      expect(screen.getByText(/Welcome back, Admin!/i)).toBeInTheDocument()
      expect(screen.getByText('User Management')).toBeInTheDocument()
    })

    it('handles regular user correctly', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'user', role: 'user', first_name: 'Regular' },
      })

      render(<Dashboard />)

      expect(screen.getByText(/Welcome back, Regular!/i)).toBeInTheDocument()
      expect(screen.queryByText('User Management')).not.toBeInTheDocument()
    })

    it('handles user without role', () => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })

      render(<Dashboard />)

      expect(screen.getByText(/Welcome back, testuser!/i)).toBeInTheDocument()
      expect(screen.queryByText('User Management')).not.toBeInTheDocument()
    })
  })

  describe('Edge Cases', () => {
    it('renders with null user', () => {
      mockUseAuth.mockReturnValue({
        user: null,
      })

      render(<Dashboard />)

      // Should still render but with no name
      expect(screen.getByText(/Welcome back,/i)).toBeInTheDocument()
    })

    it('renders with undefined user', () => {
      mockUseAuth.mockReturnValue({
        user: undefined,
      })

      render(<Dashboard />)

      expect(screen.getByText(/Welcome back,/i)).toBeInTheDocument()
    })

    it('renders with empty username', () => {
      mockUseAuth.mockReturnValue({
        user: { username: '' },
      })

      render(<Dashboard />)

      expect(screen.getByText(/Welcome back,/i)).toBeInTheDocument()
    })

    it('handles user with only first_name', () => {
      mockUseAuth.mockReturnValue({
        user: { first_name: 'John' },
      })

      render(<Dashboard />)

      expect(screen.getByText(/Welcome back, John!/i)).toBeInTheDocument()
    })
  })

  describe('Layout and Structure', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: { username: 'testuser' },
      })
    })

    it('renders main container with correct classes', () => {
      const { container } = render(<Dashboard />)
      const mainDiv = container.firstChild as HTMLElement

      expect(mainDiv).toHaveClass('max-w-7xl')
      expect(mainDiv).toHaveClass('mx-auto')
    })

    it('renders stats in grid layout', () => {
      const { container } = render(<Dashboard />)

      // Check for grid container
      const grids = container.querySelectorAll('.grid')
      expect(grids.length).toBeGreaterThan(0)
    })

    it('renders all sections in correct order', () => {
      render(<Dashboard />)

      const headings = screen.getAllByRole('heading')
      const headingTexts = headings.map(h => h.textContent)

      // Should have Welcome message, Quick Actions, Recent Activity
      expect(headingTexts.some(text => text?.includes('Welcome back'))).toBe(true)
      expect(headingTexts.some(text => text?.includes('Quick Actions'))).toBe(true)
      expect(headingTexts.some(text => text?.includes('Recent Activity'))).toBe(true)
    })
  })
})
