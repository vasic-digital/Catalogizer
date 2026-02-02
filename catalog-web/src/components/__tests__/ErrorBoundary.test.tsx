import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ErrorBoundary } from '../ErrorBoundary'

// A component that throws an error for testing
const ThrowingComponent = ({ shouldThrow = true }: { shouldThrow?: boolean }) => {
  if (shouldThrow) {
    throw new Error('Test error message')
  }
  return <div>Child content rendered</div>
}

// A component that can be toggled to throw
let shouldThrow = false
const ConditionallyThrowingComponent = () => {
  if (shouldThrow) {
    throw new Error('Conditional error')
  }
  return <div>Working component</div>
}

describe('ErrorBoundary', () => {
  // Suppress console.error for expected error boundary logs
  const originalError = console.error
  beforeAll(() => {
    console.error = vi.fn()
  })

  afterAll(() => {
    console.error = originalError
  })

  beforeEach(() => {
    shouldThrow = false
  })

  describe('Rendering children', () => {
    it('renders children when no error occurs', () => {
      render(
        <ErrorBoundary>
          <div>Normal content</div>
        </ErrorBoundary>
      )

      expect(screen.getByText('Normal content')).toBeInTheDocument()
    })

    it('renders multiple children without error', () => {
      render(
        <ErrorBoundary>
          <div>First child</div>
          <div>Second child</div>
        </ErrorBoundary>
      )

      expect(screen.getByText('First child')).toBeInTheDocument()
      expect(screen.getByText('Second child')).toBeInTheDocument()
    })
  })

  describe('Catching errors', () => {
    it('catches errors and renders default fallback UI', () => {
      render(
        <ErrorBoundary>
          <ThrowingComponent />
        </ErrorBoundary>
      )

      expect(screen.getByText('Something went wrong')).toBeInTheDocument()
      expect(screen.getByText('Test error message')).toBeInTheDocument()
    })

    it('renders the "Try again" button in default fallback', () => {
      render(
        <ErrorBoundary>
          <ThrowingComponent />
        </ErrorBoundary>
      )

      expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument()
    })

    it('logs error information via componentDidCatch', () => {
      render(
        <ErrorBoundary>
          <ThrowingComponent />
        </ErrorBoundary>
      )

      expect(console.error).toHaveBeenCalledWith(
        'ErrorBoundary caught an error:',
        expect.any(Error),
        expect.objectContaining({ componentStack: expect.any(String) })
      )
    })
  })

  describe('Custom fallback', () => {
    it('renders custom fallback when provided', () => {
      render(
        <ErrorBoundary fallback={<div>Custom error UI</div>}>
          <ThrowingComponent />
        </ErrorBoundary>
      )

      expect(screen.getByText('Custom error UI')).toBeInTheDocument()
      expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument()
    })

    it('does not render custom fallback when no error', () => {
      render(
        <ErrorBoundary fallback={<div>Custom error UI</div>}>
          <div>Normal content</div>
        </ErrorBoundary>
      )

      expect(screen.getByText('Normal content')).toBeInTheDocument()
      expect(screen.queryByText('Custom error UI')).not.toBeInTheDocument()
    })
  })

  describe('Reset / Try again', () => {
    it('resets error state when "Try again" is clicked', async () => {
      const user = userEvent.setup()

      // Control whether the child should throw via external flag
      let shouldThrowError = true
      const ConditionalError = () => {
        if (shouldThrowError) {
          throw new Error('Resettable error')
        }
        return <div>Recovered content</div>
      }

      render(
        <ErrorBoundary>
          <ConditionalError />
        </ErrorBoundary>
      )

      // Error state should be shown
      expect(screen.getByText('Something went wrong')).toBeInTheDocument()
      expect(screen.getByText('Resettable error')).toBeInTheDocument()

      // Stop throwing before clicking try again
      shouldThrowError = false

      // Click try again
      const tryAgainButton = screen.getByRole('button', { name: /try again/i })
      await user.click(tryAgainButton)

      // After reset, the component re-renders and should show recovered content
      expect(screen.getByText('Recovered content')).toBeInTheDocument()
      expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument()
    })
  })

  describe('Default error message', () => {
    it('shows default message when error has no message', () => {
      const ThrowEmptyError = () => {
        throw new Error()
      }

      render(
        <ErrorBoundary>
          <ThrowEmptyError />
        </ErrorBoundary>
      )

      // The component uses error?.message || 'An unexpected error occurred.'
      expect(screen.getByText('An unexpected error occurred.')).toBeInTheDocument()
    })
  })
})
