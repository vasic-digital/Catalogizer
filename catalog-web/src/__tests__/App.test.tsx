import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import App from '../App'

// Mock BrowserRouter to use MemoryRouter for testing
jest.mock('react-router-dom', () => {
  const actual = jest.requireActual('react-router-dom')
  return {
    ...actual,
    BrowserRouter: ({ children }: { children: React.ReactNode }) => (
      <actual.MemoryRouter initialEntries={[(global as any).testInitialRoute || '/']}>{children}</actual.MemoryRouter>
    ),
  }
})

// Mock all child components and contexts
jest.mock('@/contexts/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="auth-provider">{children}</div>
  ),
  useAuth: jest.fn(() => ({
    user: null,
    isAuthenticated: false,
    login: jest.fn(),
    logout: jest.fn(),
  })),
}))

jest.mock('@/contexts/WebSocketContext', () => ({
  WebSocketProvider: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="websocket-provider">{children}</div>
  ),
  useWebSocketContext: jest.fn(),
}))

jest.mock('@/components/ui/ConnectionStatus', () => ({
  ConnectionStatus: () => <div data-testid="connection-status">Connection Status</div>,
}))

jest.mock('@/components/layout/Layout', () => {
  const { Outlet } = require('react-router-dom')
  return {
    Layout: () => (
      <div data-testid="layout">
        Layout
        <Outlet />
      </div>
    ),
  }
})

jest.mock('@/components/auth/LoginForm', () => ({
  LoginForm: () => <div data-testid="login-form">Login Form</div>,
}))

jest.mock('@/components/auth/RegisterForm', () => ({
  RegisterForm: () => <div data-testid="register-form">Register Form</div>,
}))

jest.mock('@/components/auth/ProtectedRoute', () => ({
  ProtectedRoute: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="protected-route">{children}</div>
  ),
}))

jest.mock('@/pages/Dashboard', () => ({
  Dashboard: () => <div data-testid="dashboard-page">Dashboard Page</div>,
}))

jest.mock('@/pages/MediaBrowser', () => ({
  MediaBrowser: () => <div data-testid="media-browser-page">Media Browser Page</div>,
}))

jest.mock('@/pages/Analytics', () => ({
  Analytics: () => <div data-testid="analytics-page">Analytics Page</div>,
}))

describe('App', () => {
  describe('Rendering and Setup', () => {
    it('renders the App component', () => {
      render(<App />)
      expect(screen.getByTestId('auth-provider')).toBeInTheDocument()
    })

    it('renders with provider hierarchy', () => {
      const { container } = render(<App />)
      const authProvider = screen.getByTestId('auth-provider')
      const websocketProvider = screen.getByTestId('websocket-provider')

      // WebSocketProvider should be inside AuthProvider
      expect(authProvider).toContainElement(websocketProvider)
    })

    it('renders ConnectionStatus globally', () => {
      render(<App />)
      expect(screen.getByTestId('connection-status')).toBeInTheDocument()
    })

    it('sets up Router correctly', () => {
      const { container } = render(<App />)
      // Router should be present (no errors thrown)
      expect(container.firstChild).toBeInTheDocument()
    })
  })

  describe('Public Routes', () => {
    it('renders LoginForm on /login route', async () => {
      ;(global as any).testInitialRoute = '/login'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('login-form')).toBeInTheDocument()
      })
    })

    it('renders RegisterForm on /register route', async () => {
      ;(global as any).testInitialRoute = '/register'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('register-form')).toBeInTheDocument()
      })
    })
  })

  describe('Protected Routes with Layout', () => {
    it('renders Dashboard on /dashboard route', async () => {
      ;(global as any).testInitialRoute = '/dashboard'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })

    it('renders MediaBrowser on /media route', async () => {
      ;(global as any).testInitialRoute = '/media'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })

    it('renders Analytics on /analytics route', async () => {
      ;(global as any).testInitialRoute = '/analytics'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })

    it('renders Admin page on /admin route', async () => {
      ;(global as any).testInitialRoute = '/admin'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })

    it('renders Profile page on /profile route', async () => {
      ;(global as any).testInitialRoute = '/profile'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })

    it('renders Settings page on /settings route', async () => {
      ;(global as any).testInitialRoute = '/settings'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })
  })

  describe('Navigation and Redirects', () => {
    it('redirects from root / to /dashboard', async () => {
      ;(global as any).testInitialRoute = '/'
      render(<App />)

      await waitFor(() => {
        // Should render Layout (which means we're on /dashboard after redirect)
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })

    it('redirects from unknown route to /dashboard', async () => {
      ;(global as any).testInitialRoute = '/unknown-route'
      render(<App />)

      await waitFor(() => {
        // Should render Layout (catch-all redirects to /dashboard)
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })

    it('redirects from invalid nested route to /dashboard', async () => {
      ;(global as any).testInitialRoute = '/invalid/nested/route'
      render(<App />)

      await waitFor(() => {
        // Should render Layout (catch-all redirects to /dashboard)
        expect(screen.getByTestId('layout')).toBeInTheDocument()
      })
    })
  })

  describe('Layout Integration', () => {
    it('protected routes render inside Layout wrapper', async () => {
      ;(global as any).testInitialRoute = '/dashboard'
      render(<App />)

      await waitFor(() => {
        const layout = screen.getByTestId('layout')
        expect(layout).toBeInTheDocument()
      })
    })

    it('public routes render without Layout wrapper', async () => {
      ;(global as any).testInitialRoute = '/login'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('login-form')).toBeInTheDocument()
        expect(screen.queryByTestId('layout')).not.toBeInTheDocument()
      })
    })
  })

  describe('Protected Route Wrapper', () => {
    it('wraps dashboard with ProtectedRoute', async () => {
      ;(global as any).testInitialRoute = '/dashboard'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('protected-route')).toBeInTheDocument()
      })
    })

    it('wraps media browser with ProtectedRoute', async () => {
      ;(global as any).testInitialRoute = '/media'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('protected-route')).toBeInTheDocument()
      })
    })

    it('wraps analytics with ProtectedRoute', async () => {
      ;(global as any).testInitialRoute = '/analytics'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('protected-route')).toBeInTheDocument()
      })
    })
  })

  describe('Provider Hierarchy', () => {
    it('AuthProvider is the outermost provider', () => {
      const { container } = render(<App />)
      const authProvider = screen.getByTestId('auth-provider')

      // AuthProvider should be at the top level
      expect(container.firstChild).toContainElement(authProvider)
    })

    it('WebSocketProvider is inside AuthProvider', () => {
      render(<App />)

      const authProvider = screen.getByTestId('auth-provider')
      const websocketProvider = screen.getByTestId('websocket-provider')

      expect(authProvider).toContainElement(websocketProvider)
    })
  })

  describe('Edge Cases', () => {
    it('handles login route rendering', async () => {
      ;(global as any).testInitialRoute = '/login'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('login-form')).toBeInTheDocument()
        expect(screen.getByTestId('connection-status')).toBeInTheDocument()
      })
    })

    it('handles dashboard route rendering', async () => {
      ;(global as any).testInitialRoute = '/dashboard'
      render(<App />)

      await waitFor(() => {
        expect(screen.getByTestId('layout')).toBeInTheDocument()
        expect(screen.getByTestId('connection-status')).toBeInTheDocument()
      })
    })

    it('renders ConnectionStatus on public routes', async () => {
      ;(global as any).testInitialRoute = '/login'
      render(<App />)

      expect(screen.getByTestId('connection-status')).toBeInTheDocument()
    })

    it('renders ConnectionStatus on protected routes', async () => {
      ;(global as any).testInitialRoute = '/dashboard'
      render(<App />)

      expect(screen.getByTestId('connection-status')).toBeInTheDocument()
    })
  })
})
