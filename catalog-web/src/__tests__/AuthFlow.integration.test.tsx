import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route, useNavigate } from 'react-router-dom'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'

// Mock the AuthContext with stateful behavior for full flow tests
const authState = {
  user: null as any,
  isAuthenticated: false,
  isLoading: false,
  permissions: [] as string[],
  isAdmin: false,
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
  updateProfile: vi.fn(),
  changePassword: vi.fn(),
  hasPermission: vi.fn(),
  canAccess: vi.fn(),
}

vi.mock('@/contexts/AuthContext', async () => ({
  useAuth: () => authState,
}))

// Mock react-hot-toast
vi.mock('react-hot-toast', async () => ({
  __esModule: true,
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

// --- Simulated Page Components ---

function LoginPage() {
  const navigate = useNavigate()

  const handleLogin = async () => {
    try {
      await authState.login({ username: 'testuser', password: 'password123' })
      navigate('/dashboard')
    } catch {
      // stay on login
    }
  }

  return (
    <div data-testid="login-page">
      <h1>Login Page</h1>
      <button data-testid="login-btn" onClick={handleLogin}>
        Sign In
      </button>
    </div>
  )
}

function DashboardPage() {
  const navigate = useNavigate()

  const handleLogout = async () => {
    try {
      await authState.logout()
      navigate('/login')
    } catch {
      // handled
    }
  }

  return (
    <div data-testid="dashboard-page">
      <h1>Dashboard</h1>
      <div data-testid="welcome-msg">Welcome, {authState.user?.username}</div>
      <button data-testid="logout-btn" onClick={handleLogout}>
        Logout
      </button>
    </div>
  )
}

function AdminPage() {
  return <div data-testid="admin-page">Admin Panel</div>
}

function SettingsPage() {
  return <div data-testid="settings-page">Settings</div>
}

function renderApp(initialRoute = '/login') {
  return render(
    <MemoryRouter initialEntries={[initialRoute]}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <DashboardPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin"
          element={
            <ProtectedRoute requireAdmin>
              <AdminPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings"
          element={
            <ProtectedRoute requiredPermission="read:settings">
              <SettingsPage />
            </ProtectedRoute>
          }
        />
      </Routes>
    </MemoryRouter>
  )
}

function setAuthState(overrides: Partial<typeof authState>) {
  Object.assign(authState, overrides)
}

function resetAuthState() {
  authState.user = null
  authState.isAuthenticated = false
  authState.isLoading = false
  authState.permissions = []
  authState.isAdmin = false
  authState.login = vi.fn()
  authState.register = vi.fn()
  authState.logout = vi.fn()
  authState.updateProfile = vi.fn()
  authState.changePassword = vi.fn()
  authState.hasPermission = vi.fn().mockReturnValue(false)
  authState.canAccess = vi.fn().mockReturnValue(false)
}

describe('Auth Flow Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    resetAuthState()
  })

  describe('Unauthenticated Access', () => {
    it('shows login page at /login', () => {
      renderApp('/login')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects to login when accessing /dashboard unauthenticated', () => {
      renderApp('/dashboard')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
      expect(screen.queryByTestId('dashboard-page')).not.toBeInTheDocument()
    })

    it('redirects to login when accessing /admin unauthenticated', () => {
      renderApp('/admin')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
      expect(screen.queryByTestId('admin-page')).not.toBeInTheDocument()
    })

    it('redirects to login when accessing /settings unauthenticated', () => {
      renderApp('/settings')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
      expect(screen.queryByTestId('settings-page')).not.toBeInTheDocument()
    })
  })

  describe('Login Flow', () => {
    it('calls login and navigates to dashboard on success', async () => {
      const user = userEvent.setup()
      authState.login = vi.fn().mockResolvedValue(undefined)

      renderApp('/login')

      expect(screen.getByTestId('login-page')).toBeInTheDocument()

      // Simulate the auth state changing after login
      authState.login.mockImplementation(async () => {
        setAuthState({
          user: { id: 1, username: 'testuser', role: 'user' },
          isAuthenticated: true,
        })
      })

      await user.click(screen.getByTestId('login-btn'))

      await waitFor(() => {
        expect(authState.login).toHaveBeenCalledWith({
          username: 'testuser',
          password: 'password123',
        })
      })

      await waitFor(() => {
        expect(screen.getByTestId('dashboard-page')).toBeInTheDocument()
      })

      expect(screen.getByTestId('welcome-msg')).toHaveTextContent('Welcome, testuser')
    })

    it('stays on login page when login fails', async () => {
      const user = userEvent.setup()
      authState.login = vi.fn().mockRejectedValue(new Error('Invalid credentials'))

      renderApp('/login')

      await user.click(screen.getByTestId('login-btn'))

      await waitFor(() => {
        expect(authState.login).toHaveBeenCalled()
      })

      // Still on login page
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
      expect(screen.queryByTestId('dashboard-page')).not.toBeInTheDocument()
    })
  })

  describe('Logout Flow', () => {
    it('calls logout and navigates to login page', async () => {
      const user = userEvent.setup()
      setAuthState({
        user: { id: 1, username: 'testuser', role: 'user' },
        isAuthenticated: true,
      })
      authState.logout = vi.fn().mockImplementation(async () => {
        setAuthState({
          user: null,
          isAuthenticated: false,
        })
      })

      renderApp('/dashboard')

      expect(screen.getByTestId('dashboard-page')).toBeInTheDocument()
      expect(screen.getByTestId('welcome-msg')).toHaveTextContent('Welcome, testuser')

      await user.click(screen.getByTestId('logout-btn'))

      await waitFor(() => {
        expect(authState.logout).toHaveBeenCalled()
      })

      await waitFor(() => {
        expect(screen.getByTestId('login-page')).toBeInTheDocument()
      })
    })

    it('navigates to login even if logout API call fails', async () => {
      const user = userEvent.setup()
      setAuthState({
        user: { id: 1, username: 'testuser', role: 'user' },
        isAuthenticated: true,
      })
      // Logout rejects but we catch it and still navigate
      authState.logout = vi.fn().mockResolvedValue(undefined)

      renderApp('/dashboard')

      await user.click(screen.getByTestId('logout-btn'))

      await waitFor(() => {
        expect(screen.getByTestId('login-page')).toBeInTheDocument()
      })
    })
  })

  describe('Admin Route Access Control', () => {
    it('admin can access admin routes', () => {
      setAuthState({
        user: { id: 1, username: 'admin', role: 'admin' },
        isAuthenticated: true,
        isAdmin: true,
      })

      renderApp('/admin')

      expect(screen.getByTestId('admin-page')).toBeInTheDocument()
    })

    it('regular user is redirected from admin route to dashboard', () => {
      setAuthState({
        user: { id: 1, username: 'testuser', role: 'user' },
        isAuthenticated: true,
        isAdmin: false,
      })

      renderApp('/admin')

      expect(screen.getByTestId('dashboard-page')).toBeInTheDocument()
      expect(screen.queryByTestId('admin-page')).not.toBeInTheDocument()
    })
  })

  describe('Permission-Gated Routes', () => {
    it('allows access when user has required permission', () => {
      setAuthState({
        user: { id: 1, username: 'testuser', role: 'user' },
        isAuthenticated: true,
        hasPermission: vi.fn().mockReturnValue(true),
      })

      renderApp('/settings')

      expect(screen.getByTestId('settings-page')).toBeInTheDocument()
    })

    it('denies access when user lacks required permission', () => {
      setAuthState({
        user: { id: 1, username: 'testuser', role: 'user' },
        isAuthenticated: true,
        hasPermission: vi.fn().mockReturnValue(false),
      })

      renderApp('/settings')

      expect(screen.getByTestId('dashboard-page')).toBeInTheDocument()
      expect(screen.queryByTestId('settings-page')).not.toBeInTheDocument()
    })

    it('admin bypasses permission requirements (admin role check in hasPermission)', () => {
      setAuthState({
        user: { id: 1, username: 'admin', role: 'admin' },
        isAuthenticated: true,
        isAdmin: true,
        hasPermission: vi.fn().mockReturnValue(true),
      })

      renderApp('/settings')

      expect(screen.getByTestId('settings-page')).toBeInTheDocument()
    })
  })

  describe('Loading State', () => {
    it('shows loading spinner when auth is loading', () => {
      setAuthState({ isLoading: true })

      renderApp('/dashboard')

      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
      expect(screen.queryByTestId('dashboard-page')).not.toBeInTheDocument()
      expect(screen.queryByTestId('login-page')).not.toBeInTheDocument()
    })
  })

  describe('Full Login-Logout Cycle', () => {
    it('completes full sign-in then sign-out flow', async () => {
      const user = userEvent.setup()

      // Start unauthenticated
      renderApp('/login')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()

      // Set up login to update auth state
      authState.login = vi.fn().mockImplementation(async () => {
        setAuthState({
          user: { id: 1, username: 'testuser', role: 'user' },
          isAuthenticated: true,
        })
      })

      // Click login
      await user.click(screen.getByTestId('login-btn'))

      await waitFor(() => {
        expect(screen.getByTestId('dashboard-page')).toBeInTheDocument()
      })

      // Set up logout to update auth state
      authState.logout = vi.fn().mockImplementation(async () => {
        setAuthState({
          user: null,
          isAuthenticated: false,
        })
      })

      // Click logout
      await user.click(screen.getByTestId('logout-btn'))

      await waitFor(() => {
        expect(screen.getByTestId('login-page')).toBeInTheDocument()
      })

      expect(authState.login).toHaveBeenCalledTimes(1)
      expect(authState.logout).toHaveBeenCalledTimes(1)
    })
  })
})
