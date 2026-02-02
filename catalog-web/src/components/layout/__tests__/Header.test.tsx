import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { Header } from '../Header'
import { useAuth } from '@/contexts/AuthContext'

// Mock AuthContext
const mockLogout = vi.fn()
const mockNavigate = vi.fn()

vi.mock('@/contexts/AuthContext', async () => ({
  useAuth: vi.fn(),
}))

vi.mock('react-router-dom', async () => ({
  ...(await vi.importActual('react-router-dom')),
  useNavigate: () => mockNavigate,
}))

// Mock framer-motion to avoid animation issues in tests
vi.mock('framer-motion', async () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

const mockUseAuth = vi.mocked(useAuth)

describe('Header', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Logo and Branding', () => {
    it('renders the Catalogizer logo', () => {
      mockUseAuth.mockReturnValue({
        user: null,
        isAuthenticated: false,
        logout: mockLogout,
      })

      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.getByText('Catalogizer')).toBeInTheDocument()
      expect(screen.getByText('C')).toBeInTheDocument()
    })

    it('logo links to home page', () => {
      mockUseAuth.mockReturnValue({
        user: null,
        isAuthenticated: false,
        logout: mockLogout,
      })

      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const logoLink = screen.getByText('Catalogizer').closest('a')
      expect(logoLink).toHaveAttribute('href', '/')
    })
  })

  describe('Unauthenticated State', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: null,
        isAuthenticated: false,
        logout: mockLogout,
      })
    })

    it('does not display navigation links when not authenticated', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.queryByText('Dashboard')).not.toBeInTheDocument()
      expect(screen.queryByText('Media')).not.toBeInTheDocument()
      expect(screen.queryByText('Analytics')).not.toBeInTheDocument()
    })

    it('does not display search bar when not authenticated', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.queryByPlaceholderText('Search media...')).not.toBeInTheDocument()
    })

    it('displays Login and Sign Up buttons when not authenticated', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.getByRole('button', { name: /login/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /sign up/i })).toBeInTheDocument()
    })

    it('navigates to login page when Login button is clicked', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      await user.click(screen.getByRole('button', { name: /login/i }))
      expect(mockNavigate).toHaveBeenCalledWith('/login')
    })

    it('navigates to register page when Sign Up button is clicked', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      await user.click(screen.getByRole('button', { name: /sign up/i }))
      expect(mockNavigate).toHaveBeenCalledWith('/register')
    })
  })

  describe('Authenticated State - Regular User', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'testuser',
          first_name: 'Test',
          last_name: 'User',
          role: 'user',
        },
        isAuthenticated: true,
        logout: mockLogout,
      })
    })

    it('displays navigation links when authenticated', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.getByText('Dashboard')).toBeInTheDocument()
      expect(screen.getByText('Media')).toBeInTheDocument()
      expect(screen.getByText('Analytics')).toBeInTheDocument()
    })

    it('does not display Admin link for regular users', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.queryByText('Admin')).not.toBeInTheDocument()
    })

    it('displays search bar when authenticated', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.getAllByPlaceholderText('Search media...').length).toBeGreaterThan(0)
    })

    it('displays user greeting with first name', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.getByText(/Welcome, Test/i)).toBeInTheDocument()
    })

    it('displays username when first name is not available', () => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'testuser',
          role: 'user',
        },
        isAuthenticated: true,
        logout: mockLogout,
      })

      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.getByText(/Welcome, testuser/i)).toBeInTheDocument()
    })

    it('navigates to profile page when profile button is clicked', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      // Find the profile button by its icon (User icon)
      const profileButtons = screen.getAllByRole('button')
      const profileButton = profileButtons.find(btn => btn.querySelector('svg'))

      if (profileButton) {
        await user.click(profileButton)
      }

      // Profile button is the first icon button
      expect(mockNavigate).toHaveBeenCalled()
    })

    it('calls logout when logout button is clicked', async () => {
      const user = userEvent.setup()
      mockLogout.mockResolvedValue(undefined)

      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      // Find logout button (LogOut icon) - it's the last icon button in desktop menu
      const iconButtons = screen.getAllByRole('button').filter(btn =>
        btn.querySelector('svg') && btn.className.includes('h-8 w-8')
      )
      const logoutButton = iconButtons[iconButtons.length - 1]

      await user.click(logoutButton)

      await waitFor(() => {
        expect(mockLogout).toHaveBeenCalled()
      })
    })

    it('navigates to login after successful logout', async () => {
      const user = userEvent.setup()
      mockLogout.mockResolvedValue(undefined)

      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const iconButtons = screen.getAllByRole('button').filter(btn =>
        btn.querySelector('svg') && btn.className.includes('h-8 w-8')
      )
      const logoutButton = iconButtons[iconButtons.length - 1]

      await user.click(logoutButton)

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/login')
      })
    })

    it('handles logout errors gracefully', async () => {
      const user = userEvent.setup()
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation()
      mockLogout.mockRejectedValue(new Error('Logout failed'))

      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const iconButtons = screen.getAllByRole('button').filter(btn =>
        btn.querySelector('svg') && btn.className.includes('h-8 w-8')
      )
      const logoutButton = iconButtons[iconButtons.length - 1]

      await user.click(logoutButton)

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith(
          'Logout failed:',
          expect.any(Error)
        )
      })

      consoleErrorSpy.mockRestore()
    })
  })

  describe('Authenticated State - Admin User', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'admin',
          first_name: 'Admin',
          last_name: 'User',
          role: 'admin',
        },
        isAuthenticated: true,
        logout: mockLogout,
      })
    })

    it('displays Admin link for admin users', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      expect(screen.getByText('Admin')).toBeInTheDocument()
    })

    it('Admin link navigates to admin page', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const adminLink = screen.getByText('Admin').closest('a')
      expect(adminLink).toHaveAttribute('href', '/admin')
    })
  })

  describe('Navigation Links', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'testuser',
          role: 'user',
        },
        isAuthenticated: true,
        logout: mockLogout,
      })
    })

    it('Dashboard link navigates to dashboard page', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const dashboardLink = screen.getByText('Dashboard').closest('a')
      expect(dashboardLink).toHaveAttribute('href', '/dashboard')
    })

    it('Media link navigates to media page', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const mediaLink = screen.getByText('Media').closest('a')
      expect(mediaLink).toHaveAttribute('href', '/media')
    })

    it('Analytics link navigates to analytics page', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const analyticsLink = screen.getByText('Analytics').closest('a')
      expect(analyticsLink).toHaveAttribute('href', '/analytics')
    })
  })

  describe('Mobile Menu', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'testuser',
          first_name: 'Test',
          role: 'user',
        },
        isAuthenticated: true,
        logout: mockLogout,
      })
    })

    it('mobile menu is closed by default', () => {
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      // Mobile menu content should not be visible
      const mobileLinks = screen.queryAllByText('Dashboard')
      // Should only find desktop link, not mobile
      expect(mobileLinks.length).toBe(1)
    })

    it('toggles mobile menu when menu button is clicked', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      // Find the mobile menu toggle button
      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      expect(menuToggle).toBeInTheDocument()

      if (menuToggle) {
        await user.click(menuToggle)
      }

      // After clicking, mobile menu should be open (multiple Dashboard links visible)
      await waitFor(() => {
        const dashboardLinks = screen.getAllByText('Dashboard')
        expect(dashboardLinks.length).toBeGreaterThan(1)
      })
    })

    it('displays mobile navigation links when menu is open', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      if (menuToggle) {
        await user.click(menuToggle)
      }

      await waitFor(() => {
        const dashboardLinks = screen.getAllByText('Dashboard')
        expect(dashboardLinks.length).toBeGreaterThan(1)
        const mediaLinks = screen.getAllByText('Media')
        expect(mediaLinks.length).toBeGreaterThan(1)
      })
    })

    it('displays mobile search bar when menu is open and user is authenticated', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      if (menuToggle) {
        await user.click(menuToggle)
      }

      await waitFor(() => {
        const searchInputs = screen.getAllByPlaceholderText('Search media...')
        // Should have both desktop and mobile search bars
        expect(searchInputs.length).toBeGreaterThan(1)
      })
    })

    it('displays user profile links in mobile menu', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      if (menuToggle) {
        await user.click(menuToggle)
      }

      await waitFor(() => {
        expect(screen.getByText('Profile')).toBeInTheDocument()
        expect(screen.getByText('Settings')).toBeInTheDocument()
        expect(screen.getByText('Logout')).toBeInTheDocument()
      })
    })

    it('displays username in mobile menu', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      if (menuToggle) {
        await user.click(menuToggle)
      }

      await waitFor(() => {
        // Should display first name in mobile menu
        const usernames = screen.getAllByText(/Test/)
        expect(usernames.length).toBeGreaterThan(1)
      })
    })

    it('closes mobile menu when logout is clicked', async () => {
      const user = userEvent.setup()
      mockLogout.mockResolvedValue(undefined)

      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      if (menuToggle) {
        await user.click(menuToggle)
      }

      await waitFor(() => {
        expect(screen.getByText('Logout')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Logout'))

      await waitFor(() => {
        expect(mockLogout).toHaveBeenCalled()
      })
    })
  })

  describe('Mobile Menu - Unauthenticated', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: null,
        isAuthenticated: false,
        logout: mockLogout,
      })
    })

    it('displays Login and Sign Up in mobile menu when not authenticated', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      if (menuToggle) {
        await user.click(menuToggle)
      }

      await waitFor(() => {
        const loginLinks = screen.getAllByText(/login/i)
        expect(loginLinks.length).toBeGreaterThan(1)
      })
    })

    it('does not display navigation links in mobile menu when not authenticated', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      if (menuToggle) {
        await user.click(menuToggle)
      }

      await waitFor(() => {
        expect(screen.queryByText('Dashboard')).not.toBeInTheDocument()
      })
    })
  })

  describe('Mobile Menu - Admin User', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'admin',
          role: 'admin',
        },
        isAuthenticated: true,
        logout: mockLogout,
      })
    })

    it('displays Admin link in mobile menu for admin users', async () => {
      const user = userEvent.setup()
      render(
        <MemoryRouter>
          <Header />
        </MemoryRouter>
      )

      const menuButtons = screen.getAllByRole('button')
      const menuToggle = menuButtons.find(btn =>
        btn.querySelector('svg') && btn.className.includes('md:hidden')
      )

      if (menuToggle) {
        await user.click(menuToggle)
      }

      await waitFor(() => {
        const adminLinks = screen.getAllByText('Admin')
        // Should have both desktop and mobile Admin links
        expect(adminLinks.length).toBeGreaterThan(1)
      })
    })
  })
})
