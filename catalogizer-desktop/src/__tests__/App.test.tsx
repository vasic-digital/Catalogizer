import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter, MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import App from '../App'

// Mock Tauri invoke (already global from setup, but let's be explicit for config)
const mockInvoke = vi.fn()
vi.mock('@tauri-apps/api/core', () => ({
  invoke: (...args: any[]) => mockInvoke(...args),
}))

// Mock authStore
const mockSetAuthToken = vi.fn()
let mockIsAuthenticated = false
let mockUser = null as any

vi.mock('../stores/authStore', () => ({
  useAuthStore: () => ({
    isAuthenticated: mockIsAuthenticated,
    setAuthToken: mockSetAuthToken,
    user: mockUser,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))

// Mock configStore
let mockServerUrl: string | null = null
const mockLoadConfig = vi.fn()

vi.mock('../stores/configStore', () => ({
  useConfigStore: () => ({
    loadConfig: mockLoadConfig,
    serverUrl: mockServerUrl,
  }),
}))

// Mock child components to simplify tests
vi.mock('../components/Layout', () => ({
  default: ({ children }: any) => <div data-testid="layout">{children}</div>,
}))

vi.mock('../components/LoadingScreen', () => ({
  default: () => <div data-testid="loading-screen">Loading...</div>,
}))

vi.mock('../pages/LoginPage', () => ({
  default: () => <div data-testid="login-page">Login</div>,
}))

vi.mock('../pages/HomePage', () => ({
  default: () => <div data-testid="home-page">Home</div>,
}))

vi.mock('../pages/LibraryPage', () => ({
  default: () => <div data-testid="library-page">Library</div>,
}))

vi.mock('../pages/SearchPage', () => ({
  default: () => <div data-testid="search-page">Search</div>,
}))

vi.mock('../pages/SettingsPage', () => ({
  default: () => <div data-testid="settings-page">Settings</div>,
}))

vi.mock('../pages/MediaDetailPage', () => ({
  default: () => <div data-testid="media-detail-page">MediaDetail</div>,
}))

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

const renderApp = (initialEntries = ['/']) => {
  const queryClient = createQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={initialEntries}>
        <App />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockIsAuthenticated = false
    mockServerUrl = null
    mockUser = null
    mockLoadConfig.mockResolvedValue(undefined)
    mockInvoke.mockResolvedValue({
      server_url: null,
      auth_token: null,
      theme: 'dark',
      auto_start: false,
    })
  })

  it('shows loading screen initially', () => {
    // Make invoke never resolve to keep loading state
    mockLoadConfig.mockImplementation(() => new Promise(() => {}))

    renderApp()

    expect(screen.getByTestId('loading-screen')).toBeInTheDocument()
  })

  it('redirects to settings when no server URL is configured', async () => {
    mockServerUrl = null

    renderApp()

    await waitFor(() => {
      expect(screen.getByTestId('settings-page')).toBeInTheDocument()
    })
  })

  it('redirects to login when not authenticated but server URL exists', async () => {
    mockServerUrl = 'http://localhost:8080'
    mockIsAuthenticated = false

    renderApp()

    await waitFor(() => {
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })
  })

  it('shows main layout with home page when authenticated', async () => {
    mockServerUrl = 'http://localhost:8080'
    mockIsAuthenticated = true

    renderApp()

    await waitFor(() => {
      expect(screen.getByTestId('layout')).toBeInTheDocument()
    })

    expect(screen.getByTestId('home-page')).toBeInTheDocument()
  })

  it('calls loadConfig on initialization', async () => {
    renderApp()

    await waitFor(() => {
      expect(mockLoadConfig).toHaveBeenCalled()
    })
  })

  it('calls invoke to get_config on initialization', async () => {
    renderApp()

    await waitFor(() => {
      expect(mockInvoke).toHaveBeenCalledWith('get_config')
    })
  })

  it('sets auth token when config has one', async () => {
    mockInvoke.mockResolvedValue({
      server_url: 'http://localhost:8080',
      auth_token: 'stored-token',
      theme: 'dark',
      auto_start: false,
    })

    renderApp()

    await waitFor(() => {
      expect(mockSetAuthToken).toHaveBeenCalledWith('stored-token')
    })
  })

  it('handles initialization errors gracefully', async () => {
    mockLoadConfig.mockRejectedValue(new Error('Config load failed'))

    renderApp()

    // Should still render (setIsInitialized(true) in catch block)
    await waitFor(() => {
      // App should not be stuck on loading screen
      expect(screen.queryByTestId('loading-screen')).not.toBeInTheDocument()
    })
  })
})
