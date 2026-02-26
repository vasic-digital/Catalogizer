import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { LoginForm } from '../LoginForm'
import { useAuth } from '@/contexts/AuthContext'

// Mock the AuthContext
vi.mock('@/contexts/AuthContext', async () => ({
  useAuth: vi.fn(),
}))

// Mock framer-motion to avoid animation issues in tests
vi.mock('framer-motion', async () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
}))

// Mock react-router-dom's useNavigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => ({
  ...(await vi.importActual('react-router-dom')),
  useNavigate: () => mockNavigate,
}))

const mockUseAuth = vi.mocked(useAuth)

describe('LoginForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuth.mockReturnValue({
      login: vi.fn().mockResolvedValue(undefined),
    })
  })

  describe('Rendering', () => {
    it('renders the login form with all elements', () => {
      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      expect(screen.getByText('Welcome back')).toBeInTheDocument()
      expect(screen.getByText('Enter your credentials to access Catalogizer')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Enter your username')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Enter your password')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
    })

    it('renders remember me checkbox', () => {
      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      expect(screen.getByLabelText('Remember me')).toBeInTheDocument()
      expect(screen.getByRole('checkbox')).toBeInTheDocument()
    })

    it('renders forgot password link', () => {
      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const forgotPasswordLink = screen.getByText('Forgot password?')
      expect(forgotPasswordLink).toBeInTheDocument()
      expect(forgotPasswordLink).toHaveAttribute('href', '/forgot-password')
    })

    it('renders create account link', () => {
      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      expect(screen.getByText("Don't have an account?")).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /create new account/i })).toBeInTheDocument()
    })
  })

  describe('Form Input', () => {
    it('updates username input value', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      await user.type(usernameInput, 'testuser')

      expect(usernameInput).toHaveValue('testuser')
    })

    it('updates password input value', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const passwordInput = screen.getByPlaceholderText('Enter your password')
      await user.type(passwordInput, 'password123')

      expect(passwordInput).toHaveValue('password123')
    })

    it('password input is hidden by default', () => {
      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const passwordInput = screen.getByPlaceholderText('Enter your password')
      expect(passwordInput).toHaveAttribute('type', 'password')
    })
  })

  describe('Password Visibility Toggle', () => {
    it('toggles password visibility when eye icon is clicked', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const passwordInput = screen.getByPlaceholderText('Enter your password')
      expect(passwordInput).toHaveAttribute('type', 'password')

      // Find and click the toggle button (it's the only button with type="button")
      const toggleButton = screen.getAllByRole('button').find((btn) => (btn as HTMLButtonElement).type === 'button')
      expect(toggleButton).toBeInTheDocument()

      await user.click(toggleButton as HTMLElement)
      expect(passwordInput).toHaveAttribute('type', 'text')

      await user.click(toggleButton as HTMLElement)
      expect(passwordInput).toHaveAttribute('type', 'password')
    })
  })

  describe('Form Validation', () => {
    it('submit button is disabled when username is empty', () => {
      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const submitButton = screen.getByRole('button', { name: /sign in/i })
      expect(submitButton).toBeDisabled()
    })

    it('submit button is disabled when password is empty', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      await user.type(usernameInput, 'testuser')

      const submitButton = screen.getByRole('button', { name: /sign in/i })
      expect(submitButton).toBeDisabled()
    })

    it('submit button is disabled when username is only whitespace', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      const passwordInput = screen.getByPlaceholderText('Enter your password')

      await user.type(usernameInput, '   ')
      await user.type(passwordInput, 'password123')

      const submitButton = screen.getByRole('button', { name: /sign in/i })
      expect(submitButton).toBeDisabled()
    })

    it('submit button is disabled when password is only whitespace', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      const passwordInput = screen.getByPlaceholderText('Enter your password')

      await user.type(usernameInput, 'testuser')
      await user.type(passwordInput, '   ')

      const submitButton = screen.getByRole('button', { name: /sign in/i })
      expect(submitButton).toBeDisabled()
    })

    it('submit button is enabled when both fields are filled', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      const passwordInput = screen.getByPlaceholderText('Enter your password')

      await user.type(usernameInput, 'testuser')
      await user.type(passwordInput, 'password123')

      const submitButton = screen.getByRole('button', { name: /sign in/i })
      expect(submitButton).not.toBeDisabled()
    })
  })

  describe('Form Submission', () => {
    it('calls login with trimmed username and password on submit', async () => {
      const user = userEvent.setup()
      const mockLogin = vi.fn().mockResolvedValue(undefined)
      mockUseAuth.mockReturnValue({ login: mockLogin })

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      const passwordInput = screen.getByPlaceholderText('Enter your password')
      const submitButton = screen.getByRole('button', { name: /sign in/i })

      await user.type(usernameInput, '  testuser  ')
      await user.type(passwordInput, 'password123')
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith(
          expect.objectContaining({
            username: 'testuser',
            password: 'password123',
          })
        )
      })
    })

    it('navigates to dashboard on successful login', async () => {
      const user = userEvent.setup()
      const mockLogin = vi.fn().mockResolvedValue(undefined)
      mockUseAuth.mockReturnValue({ login: mockLogin })

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      const passwordInput = screen.getByPlaceholderText('Enter your password')
      const submitButton = screen.getByRole('button', { name: /sign in/i })

      await user.type(usernameInput, 'testuser')
      await user.type(passwordInput, 'password123')
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/dashboard')
      })
    })

    it('shows loading state during login', async () => {
      const user = userEvent.setup()
      const mockLogin = vi.fn(() => new Promise((resolve) => setTimeout(resolve, 100)))
      mockUseAuth.mockReturnValue({ login: mockLogin })

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      const passwordInput = screen.getByPlaceholderText('Enter your password')
      const submitButton = screen.getByRole('button', { name: /sign in/i })

      await user.type(usernameInput, 'testuser')
      await user.type(passwordInput, 'password123')
      await user.click(submitButton)

      // Button should be disabled during loading
      expect(submitButton).toBeDisabled()

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalled()
      })
    })

    it('handles login errors gracefully', async () => {
      const user = userEvent.setup()
      const mockLogin = vi.fn().mockRejectedValue(new Error('Invalid credentials'))
      mockUseAuth.mockReturnValue({ login: mockLogin })
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => { /* Suppress error output during test */ })

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter your username')
      const passwordInput = screen.getByPlaceholderText('Enter your password')
      const submitButton = screen.getByRole('button', { name: /sign in/i })

      await user.type(usernameInput, 'testuser')
      await user.type(passwordInput, 'wrongpassword')
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalled()
      })

      // Should log error
      expect(consoleErrorSpy).toHaveBeenCalledWith('Login failed:', expect.any(Error))

      // Button should be re-enabled after error
      await waitFor(() => {
        expect(submitButton).not.toBeDisabled()
      })

      consoleErrorSpy.mockRestore()
    })

    it('does not submit form when username is empty', async () => {
      const user = userEvent.setup()
      const mockLogin = vi.fn()
      mockUseAuth.mockReturnValue({ login: mockLogin })

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const passwordInput = screen.getByPlaceholderText('Enter your password')
      await user.type(passwordInput, 'password123')

      const submitButton = screen.getByRole('button', { name: /sign in/i })

      // Button should be disabled
      expect(submitButton).toBeDisabled()

      // Login should not be called
      expect(mockLogin).not.toHaveBeenCalled()
    })
  })

  describe('User Interactions', () => {
    it('allows checking remember me checkbox', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <LoginForm />
        </MemoryRouter>
      )

      const checkbox = screen.getByLabelText('Remember me')
      expect(checkbox).not.toBeChecked()

      await user.click(checkbox)
      expect(checkbox).toBeChecked()

      await user.click(checkbox)
      expect(checkbox).not.toBeChecked()
    })
  })
})
