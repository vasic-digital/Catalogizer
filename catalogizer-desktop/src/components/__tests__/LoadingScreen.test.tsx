import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import LoadingScreen from '../LoadingScreen'

// Mock lucide-react
vi.mock('lucide-react', () => ({
  Loader2: (props: any) => <span data-testid="icon-loader" {...props} />,
}))

describe('LoadingScreen', () => {
  it('renders the loading title', () => {
    render(<LoadingScreen />)

    expect(screen.getByText('Loading Catalogizer')).toBeInTheDocument()
  })

  it('renders the loading description', () => {
    render(<LoadingScreen />)

    expect(screen.getByText('Initializing your media library...')).toBeInTheDocument()
  })

  it('renders the loading spinner icon', () => {
    render(<LoadingScreen />)

    expect(screen.getByTestId('icon-loader')).toBeInTheDocument()
  })

  it('has the correct container styling for full screen centering', () => {
    const { container } = render(<LoadingScreen />)

    const outerDiv = container.firstChild as HTMLElement
    expect(outerDiv).toHaveClass('min-h-screen')
    expect(outerDiv).toHaveClass('flex')
    expect(outerDiv).toHaveClass('items-center')
    expect(outerDiv).toHaveClass('justify-center')
  })

  it('applies the animate-spin class to the loader icon', () => {
    render(<LoadingScreen />)

    const loader = screen.getByTestId('icon-loader')
    expect(loader).toHaveClass('animate-spin')
  })
})
