import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { RegisterForm } from '../RegisterForm'
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

describe('RegisterForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuth.mockReturnValue({
      register: vi.fn().mockResolvedValue(undefined),
    })
  })

  describe('Rendering', () => {
    it('renders the registration form with all elements', () => {
      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      expect(screen.getByRole('heading', { name: /create account/i })).toBeInTheDocument()
      expect(screen.getByText('Join Catalogizer to start organizing your media')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('John')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Doe')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Enter username')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Enter email address')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Create password')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Confirm password')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /create account/i })).toBeInTheDocument()
    })

    it('renders sign in link', () => {
      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      expect(screen.getByText('Already have an account?')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /sign in instead/i })).toBeInTheDocument()
    })
  })

  describe('Form Input', () => {
    it('updates first name input value', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      const firstNameInput = screen.getByPlaceholderText('John')
      await user.type(firstNameInput, 'Jane')

      expect(firstNameInput).toHaveValue('Jane')
    })

    it('updates last name input value', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      const lastNameInput = screen.getByPlaceholderText('Doe')
      await user.type(lastNameInput, 'Smith')

      expect(lastNameInput).toHaveValue('Smith')
    })

    it('updates username input value', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      const usernameInput = screen.getByPlaceholderText('Enter username')
      await user.type(usernameInput, 'testuser')

      expect(usernameInput).toHaveValue('testuser')
    })

    it('updates email input value', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      const emailInput = screen.getByPlaceholderText('Enter email address')
      await user.type(emailInput, 'test@example.com')

      expect(emailInput).toHaveValue('test@example.com')
    })

    it('updates password input value', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      const passwordInput = screen.getByPlaceholderText('Create password')
      await user.type(passwordInput, 'password123')

      expect(passwordInput).toHaveValue('password123')
    })

    it('updates confirm password input value', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      const confirmPasswordInput = screen.getByPlaceholderText('Confirm password')
      await user.type(confirmPasswordInput, 'password123')

      expect(confirmPasswordInput).toHaveValue('password123')
    })
  })

  describe('Password Visibility Toggle', () => {
    it('toggles password visibility when eye icon is clicked', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      const passwordInput = screen.getByPlaceholderText('Create password')
      expect(passwordInput).toHaveAttribute('type', 'password')

      // Find password toggle button
      const toggleButtons = screen.getAllByRole('button').filter((btn) => (btn as HTMLButtonElement).type === 'button')
      const passwordToggle = toggleButtons[0]

      await user.click(passwordToggle)
      expect(passwordInput).toHaveAttribute('type', 'text')

      await user.click(passwordToggle)
      expect(passwordInput).toHaveAttribute('type', 'password')
    })

    it('toggles confirm password visibility when eye icon is clicked', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      const confirmPasswordInput = screen.getByPlaceholderText('Confirm password')
      expect(confirmPasswordInput).toHaveAttribute('type', 'password')

      // Find confirm password toggle button (second toggle button)
      const toggleButtons = screen.getAllByRole('button').filter((btn) => (btn as HTMLButtonElement).type === 'button')
      const confirmPasswordToggle = toggleButtons[1]

      await user.click(confirmPasswordToggle)
      expect(confirmPasswordInput).toHaveAttribute('type', 'text')

      await user.click(confirmPasswordToggle)
      expect(confirmPasswordInput).toHaveAttribute('type', 'password')
    })
  })

  describe('Form Validation', () => {
    it('shows error when username is empty on submit', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      // Fill all fields except username
      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      // Use fireEvent.submit to bypass native HTML5 required validation
      const form = screen.getByRole('button', { name: /create account/i }).closest('form')!
      fireEvent.submit(form)

      await waitFor(() => {
        expect(screen.getByText('Username is required')).toBeInTheDocument()
      })
    })

    it('shows error when username is too short', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'ab')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Username must be at least 3 characters')).toBeInTheDocument()
      })
    })

    it('shows error when email is invalid', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'testuser')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'invalid-email')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      // Use fireEvent.submit to bypass native HTML5 email type validation
      const form = screen.getByRole('button', { name: /create account/i }).closest('form')!
      fireEvent.submit(form)

      await waitFor(() => {
        expect(screen.getByText('Email is invalid')).toBeInTheDocument()
      })
    })

    it('shows error when password is too short', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'testuser')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'short')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'short')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Password must be at least 8 characters')).toBeInTheDocument()
      })
    })

    it('shows error when passwords do not match', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'testuser')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'different123')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
      })
    })

    it('shows error when first name is empty', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'testuser')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      // Use fireEvent.submit to bypass native HTML5 required validation
      const form = screen.getByRole('button', { name: /create account/i }).closest('form')!
      fireEvent.submit(form)

      await waitFor(() => {
        expect(screen.getByText('First name is required')).toBeInTheDocument()
      })
    })

    it('shows error when last name is empty', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Enter username'), 'testuser')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      // Use fireEvent.submit to bypass native HTML5 required validation
      const form = screen.getByRole('button', { name: /create account/i }).closest('form')!
      fireEvent.submit(form)

      await waitFor(() => {
        expect(screen.getByText('Last name is required')).toBeInTheDocument()
      })
    })

    it('clears error when field is corrected', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      // Submit with invalid username
      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'ab')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Username must be at least 3 characters')).toBeInTheDocument()
      })

      // Correct the username
      const usernameInput = screen.getByPlaceholderText('Enter username')
      await user.clear(usernameInput)
      await user.type(usernameInput, 'validuser')

      await waitFor(() => {
        expect(screen.queryByText('Username must be at least 3 characters')).not.toBeInTheDocument()
      })
    })
  })

  describe('Form Submission', () => {
    it('calls register with correct data on valid submission', async () => {
      const user = userEvent.setup()
      const mockRegister = vi.fn().mockResolvedValue(undefined)
      mockUseAuth.mockReturnValue({ register: mockRegister })

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), '  Jane  ')
      await user.type(screen.getByPlaceholderText('Doe'), '  Smith  ')
      await user.type(screen.getByPlaceholderText('Enter username'), '  testuser  ')
      await user.type(screen.getByPlaceholderText('Enter email address'), '  test@example.com  ')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockRegister).toHaveBeenCalledWith({
          username: 'testuser',
          email: 'test@example.com',
          password: 'password123',
          first_name: 'Jane',
          last_name: 'Smith',
        })
      })
    })

    it('navigates to login on successful registration', async () => {
      const user = userEvent.setup()
      const mockRegister = vi.fn().mockResolvedValue(undefined)
      mockUseAuth.mockReturnValue({ register: mockRegister })

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'testuser')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/login')
      })
    })

    it('shows loading state during registration', async () => {
      const user = userEvent.setup()
      const mockRegister = vi.fn(() => new Promise((resolve) => setTimeout(resolve, 100)))
      mockUseAuth.mockReturnValue({ register: mockRegister })

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'testuser')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      // Button should show loading state
      expect(submitButton).toBeDisabled()

      await waitFor(() => {
        expect(mockRegister).toHaveBeenCalled()
      })
    })

    it('handles registration errors gracefully', async () => {
      const user = userEvent.setup()
      const mockRegister = vi.fn().mockRejectedValue(new Error('Registration failed'))
      mockUseAuth.mockReturnValue({ register: mockRegister })
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'testuser')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockRegister).toHaveBeenCalled()
      })

      // Should log error
      expect(consoleErrorSpy).toHaveBeenCalledWith('Registration failed:', expect.any(Error))

      // Button should be re-enabled after error
      await waitFor(() => {
        expect(submitButton).not.toBeDisabled()
      })

      consoleErrorSpy.mockRestore()
    })

    it('does not submit form when validation fails', async () => {
      const user = userEvent.setup()
      const mockRegister = vi.fn()
      mockUseAuth.mockReturnValue({ register: mockRegister })

      render(
        <MemoryRouter>
          <RegisterForm />
        </MemoryRouter>
      )

      // Submit with invalid data (short username)
      await user.type(screen.getByPlaceholderText('John'), 'Jane')
      await user.type(screen.getByPlaceholderText('Doe'), 'Smith')
      await user.type(screen.getByPlaceholderText('Enter username'), 'ab')
      await user.type(screen.getByPlaceholderText('Enter email address'), 'test@example.com')
      await user.type(screen.getByPlaceholderText('Create password'), 'password123')
      await user.type(screen.getByPlaceholderText('Confirm password'), 'password123')

      const submitButton = screen.getByRole('button', { name: /create account/i })
      await user.click(submitButton)

      // Register should not be called
      expect(mockRegister).not.toHaveBeenCalled()
    })
  })
})
