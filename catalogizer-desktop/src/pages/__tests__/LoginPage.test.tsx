import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter, MemoryRouter } from 'react-router-dom'
import LoginPage from '../LoginPage'

// Mock authStore
const mockLogin = vi.fn()
vi.mock('../../stores/authStore', () => ({
  useAuthStore: () => ({
    login: mockLogin,
  }),
}))

// Mock configStore
vi.mock('../../stores/configStore', () => ({
  useConfigStore: () => ({
    serverUrl: 'http://localhost:8080',
  }),
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

// Mock lucide-react
vi.mock('lucide-react', () => ({
  Film: (props: any) => <span data-testid="icon-film" {...props} />,
  Loader2: (props: any) => <span data-testid="icon-loader" {...props} />,
  Eye: (props: any) => <span data-testid="icon-eye" {...props} />,
  EyeOff: (props: any) => <span data-testid="icon-eye-off" {...props} />,
}))

const renderLoginPage = () => {
  return render(
    <BrowserRouter>
      <LoginPage />
    </BrowserRouter>
  )
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the welcome title', () => {
    renderLoginPage()

    expect(screen.getByText('Welcome to Catalogizer')).toBeInTheDocument()
  })

  it('renders the subtitle', () => {
    renderLoginPage()

    expect(screen.getByText('Sign in to access your media library')).toBeInTheDocument()
  })

  it('renders username and password inputs', () => {
    renderLoginPage()

    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
  })

  it('renders the submit button', () => {
    renderLoginPage()

    expect(screen.getByText('Sign in')).toBeInTheDocument()
  })

  it('renders the server URL when configured', () => {
    renderLoginPage()

    expect(screen.getByText('Connecting to: http://localhost:8080')).toBeInTheDocument()
  })

  it('renders the configure server connection link', () => {
    renderLoginPage()

    expect(screen.getByText('Configure server connection')).toBeInTheDocument()
  })

  it('disables submit button when fields are empty', () => {
    renderLoginPage()

    const submitButton = screen.getByText('Sign in')
    expect(submitButton).toBeDisabled()
  })

  it('enables submit button when both fields are filled', async () => {
    const user = userEvent.setup()
    renderLoginPage()

    await user.type(screen.getByLabelText('Username'), 'admin')
    await user.type(screen.getByLabelText('Password'), 'password')

    const submitButton = screen.getByText('Sign in')
    expect(submitButton).not.toBeDisabled()
  })

  it('calls login and navigates on successful submit', async () => {
    mockLogin.mockResolvedValue(undefined)
    const user = userEvent.setup()

    renderLoginPage()

    await user.type(screen.getByLabelText('Username'), 'admin')
    await user.type(screen.getByLabelText('Password'), 'password')
    await user.click(screen.getByText('Sign in'))

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('admin', 'password')
    })

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/')
    })
  })

  it('shows error message on login failure', async () => {
    mockLogin.mockRejectedValue(new Error('Invalid credentials'))
    const user = userEvent.setup()

    renderLoginPage()

    await user.type(screen.getByLabelText('Username'), 'admin')
    await user.type(screen.getByLabelText('Password'), 'wrongpass')
    await user.click(screen.getByText('Sign in'))

    await waitFor(() => {
      expect(screen.getByText('Invalid credentials')).toBeInTheDocument()
    })
  })

  it('toggles password visibility', async () => {
    const user = userEvent.setup()
    renderLoginPage()

    const passwordInput = screen.getByLabelText('Password')
    expect(passwordInput).toHaveAttribute('type', 'password')

    // Click the toggle button (it's a button within the password field)
    const toggleButton = passwordInput.parentElement?.querySelector('button')
    expect(toggleButton).toBeTruthy()
    await user.click(toggleButton!)

    expect(passwordInput).toHaveAttribute('type', 'text')
  })

  it('navigates to settings when configure link is clicked', async () => {
    const user = userEvent.setup()
    renderLoginPage()

    await user.click(screen.getByText('Configure server connection'))

    expect(mockNavigate).toHaveBeenCalledWith('/settings')
  })
})
