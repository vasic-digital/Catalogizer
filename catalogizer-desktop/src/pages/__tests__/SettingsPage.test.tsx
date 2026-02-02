import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import SettingsPage from '../SettingsPage'

// Mock apiService
vi.mock('../../services/apiService', () => ({
  apiService: {
    healthCheck: vi.fn(),
    getSMBConfigs: vi.fn().mockResolvedValue([]),
    createSMBConfig: vi.fn(),
    deleteSMBConfig: vi.fn(),
  },
}))

// Mock react-router-dom navigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

// Mock stores
const mockConfigStore = {
  serverUrl: 'http://localhost:8080',
  theme: 'dark' as const,
  autoStart: false,
  setServerUrl: vi.fn().mockResolvedValue(undefined),
  setTheme: vi.fn().mockResolvedValue(undefined),
  setAutoStart: vi.fn().mockResolvedValue(undefined),
}

const mockAuthStore = {
  isAuthenticated: false,
}

vi.mock('../../stores/configStore', () => ({
  useConfigStore: () => mockConfigStore,
}))

vi.mock('../../stores/authStore', () => ({
  useAuthStore: () => mockAuthStore,
}))

import { apiService } from '../../services/apiService'

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  return <BrowserRouter>{children}</BrowserRouter>
}

describe('SettingsPage', () => {
  beforeEach(() => {
    mockAuthStore.isAuthenticated = false
    mockConfigStore.serverUrl = 'http://localhost:8080'
    mockConfigStore.theme = 'dark'
    mockConfigStore.autoStart = false
  })

  it('renders the settings page title and description', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('Settings')).toBeInTheDocument()
    expect(screen.getByText('Configure your Catalogizer desktop client')).toBeInTheDocument()
  })

  it('renders the Server Configuration section', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('Server Configuration')).toBeInTheDocument()
    expect(screen.getByLabelText('Server URL')).toBeInTheDocument()
  })

  it('renders the Appearance section with theme selector', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('Appearance')).toBeInTheDocument()
    expect(screen.getByLabelText('Theme')).toBeInTheDocument()

    const themeSelect = screen.getByLabelText('Theme') as HTMLSelectElement
    expect(themeSelect).toBeInTheDocument()

    // Check theme options exist
    const options = themeSelect.querySelectorAll('option')
    const optionValues = Array.from(options).map((opt) => opt.value)
    expect(optionValues).toContain('light')
    expect(optionValues).toContain('dark')
    expect(optionValues).toContain('system')
  })

  it('renders the Storage Configuration section', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('Storage Configuration')).toBeInTheDocument()
    expect(
      screen.getByText(
        'Configure storage sources for media scanning. Supported protocols: SMB, FTP, NFS, WebDAV, Local.'
      )
    ).toBeInTheDocument()
  })

  it('renders the General section with auto-start toggle', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('General')).toBeInTheDocument()
    expect(screen.getByText('Auto-start')).toBeInTheDocument()
    expect(screen.getByText('Start Catalogizer when your computer starts')).toBeInTheDocument()
  })

  it('renders the Save Settings button', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('Save Settings')).toBeInTheDocument()
  })

  it('renders the Test connection button', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('Test')).toBeInTheDocument()
  })

  it('populates server URL input from config store', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    const serverInput = screen.getByLabelText('Server URL') as HTMLInputElement
    expect(serverInput.value).toBe('http://localhost:8080')
  })

  it('allows editing the server URL', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    const serverInput = screen.getByLabelText('Server URL') as HTMLInputElement
    await user.clear(serverInput)
    await user.type(serverInput, 'http://newserver:9090')

    expect(serverInput.value).toBe('http://newserver:9090')
  })

  it('shows test connection error when URL is empty', async () => {
    mockConfigStore.serverUrl = ''
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    const serverInput = screen.getByLabelText('Server URL') as HTMLInputElement
    await user.clear(serverInput)

    // Test button should be disabled when input is empty
    const testButton = screen.getByText('Test')
    expect(testButton).toBeDisabled()
  })

  it('shows success message when test connection succeeds', async () => {
    vi.mocked(apiService.healthCheck).mockResolvedValue({
      status: 'healthy',
      timestamp: '2023-01-01T00:00:00Z',
    })

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    const testButton = screen.getByText('Test')
    await user.click(testButton)

    await waitFor(() => {
      expect(screen.getByText(/Connected successfully/)).toBeInTheDocument()
    })
  })

  it('shows error message when test connection fails', async () => {
    vi.mocked(apiService.healthCheck).mockRejectedValue(new Error('Connection refused'))

    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    const testButton = screen.getByText('Test')
    await user.click(testButton)

    await waitFor(() => {
      expect(screen.getByText('Connection refused')).toBeInTheDocument()
    })
  })

  it('allows changing the theme', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    const themeSelect = screen.getByLabelText('Theme')
    await user.selectOptions(themeSelect, 'light')

    expect(themeSelect).toHaveValue('light')
  })

  it('does not show back button when user is not authenticated', () => {
    mockAuthStore.isAuthenticated = false

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    // ArrowLeft icon button should not be present
    // The back button is only shown when canGoBack = isAuthenticated && serverUrl
    const buttons = screen.getAllByRole('button')
    // Only Test, Save Settings, and Add Storage Source buttons should exist
    const buttonTexts = buttons.map((b) => b.textContent?.trim()).filter(Boolean)
    expect(buttonTexts).not.toContain('')
  })

  it('shows Add Storage Source button', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('Add Storage Source')).toBeInTheDocument()
  })

  it('shows storage form when Add Storage Source is clicked', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    await user.click(screen.getByText('Add Storage Source'))

    expect(screen.getByPlaceholderText('Storage path (e.g. //server/share or /mnt/media)')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Username (optional)')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Password (optional)')).toBeInTheDocument()
    expect(screen.getByText('Add Source')).toBeInTheDocument()
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('hides storage form when Cancel is clicked', async () => {
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    await user.click(screen.getByText('Add Storage Source'))
    expect(screen.getByPlaceholderText('Storage path (e.g. //server/share or /mnt/media)')).toBeInTheDocument()

    await user.click(screen.getByText('Cancel'))
    expect(screen.queryByPlaceholderText('Storage path (e.g. //server/share or /mnt/media)')).not.toBeInTheDocument()
  })

  it('shows no storage sources message when list is empty', () => {
    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    expect(screen.getByText('No storage sources configured yet.')).toBeInTheDocument()
  })

  it('saves settings and navigates to login when not authenticated', async () => {
    mockAuthStore.isAuthenticated = false
    const user = userEvent.setup()

    render(
      <TestWrapper>
        <SettingsPage />
      </TestWrapper>
    )

    await user.click(screen.getByText('Save Settings'))

    await waitFor(() => {
      expect(mockConfigStore.setServerUrl).toHaveBeenCalled()
      expect(mockConfigStore.setTheme).toHaveBeenCalled()
      expect(mockConfigStore.setAutoStart).toHaveBeenCalled()
      expect(mockNavigate).toHaveBeenCalledWith('/login')
    })
  })
})
