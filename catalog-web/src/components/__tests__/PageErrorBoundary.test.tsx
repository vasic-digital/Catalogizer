import { render, screen } from '@testing-library/react'
import { PageErrorBoundary } from '../PageErrorBoundary'

// A component that throws an error for testing
const ThrowingComponent = ({ shouldThrow = true }: { shouldThrow?: boolean }) => {
  if (shouldThrow) {
    throw new Error('Test page error')
  }
  return <div>Page content rendered</div>
}

describe('PageErrorBoundary', () => {
  // Suppress console.error for expected error boundary logs
  const originalError = console.error
  beforeAll(() => {
    console.error = vi.fn()
  })

  afterAll(() => {
    console.error = originalError
  })

  describe('Rendering children', () => {
    it('renders children when no error occurs', () => {
      render(
        <PageErrorBoundary pageName="Dashboard">
          <div>Normal page content</div>
        </PageErrorBoundary>
      )

      expect(screen.getByText('Normal page content')).toBeInTheDocument()
    })

    it('renders multiple children without error', () => {
      render(
        <PageErrorBoundary pageName="Settings">
          <div>Section 1</div>
          <div>Section 2</div>
        </PageErrorBoundary>
      )

      expect(screen.getByText('Section 1')).toBeInTheDocument()
      expect(screen.getByText('Section 2')).toBeInTheDocument()
    })
  })

  describe('Error fallback', () => {
    it('shows error message with page name when child throws', () => {
      render(
        <PageErrorBoundary pageName="Dashboard">
          <ThrowingComponent />
        </PageErrorBoundary>
      )

      expect(screen.getByText('Something went wrong in Dashboard')).toBeInTheDocument()
    })

    it('shows descriptive message about other pages working', () => {
      render(
        <PageErrorBoundary pageName="Media Browser">
          <ThrowingComponent />
        </PageErrorBoundary>
      )

      expect(
        screen.getByText('This page encountered an error. Other pages should still work normally.')
      ).toBeInTheDocument()
    })

    it('renders a Reload Page button', () => {
      render(
        <PageErrorBoundary pageName="Settings">
          <ThrowingComponent />
        </PageErrorBoundary>
      )

      expect(screen.getByRole('button', { name: /reload page/i })).toBeInTheDocument()
    })

    it('renders the error icon SVG', () => {
      const { container } = render(
        <PageErrorBoundary pageName="Dashboard">
          <ThrowingComponent />
        </PageErrorBoundary>
      )

      const svg = container.querySelector('svg')
      expect(svg).toBeInTheDocument()
    })

    it('does not show page content when error occurs', () => {
      render(
        <PageErrorBoundary pageName="Dashboard">
          <ThrowingComponent />
        </PageErrorBoundary>
      )

      expect(screen.queryByText('Page content rendered')).not.toBeInTheDocument()
    })
  })

  describe('Different page names', () => {
    it('displays correct page name for Settings', () => {
      render(
        <PageErrorBoundary pageName="Settings">
          <ThrowingComponent />
        </PageErrorBoundary>
      )

      expect(screen.getByText('Something went wrong in Settings')).toBeInTheDocument()
    })

    it('displays correct page name for Media Browser', () => {
      render(
        <PageErrorBoundary pageName="Media Browser">
          <ThrowingComponent />
        </PageErrorBoundary>
      )

      expect(screen.getByText('Something went wrong in Media Browser')).toBeInTheDocument()
    })
  })
})
