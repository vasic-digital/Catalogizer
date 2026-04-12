import { render, screen, act } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import { SplashScreen } from '../SplashScreen'

// Image import removed — SplashScreen now uses a text-based brand mark

describe('SplashScreen', () => {
  let onComplete: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.useFakeTimers()
    onComplete = vi.fn()
    localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders the default title', () => {
    render(<SplashScreen onComplete={onComplete} />)
    expect(screen.getByText('Catalogizer')).toBeInTheDocument()
  })

  it('renders a custom title when provided', () => {
    render(<SplashScreen onComplete={onComplete} appTitle="Catalogizer Installation Wizard" />)
    expect(screen.getByText('Catalogizer Installation Wizard')).toBeInTheDocument()
  })

  it('renders the default subtitle', () => {
    render(<SplashScreen onComplete={onComplete} />)
    expect(screen.getByText('Your media collection, organized')).toBeInTheDocument()
  })

  it('renders a custom subtitle when provided', () => {
    render(<SplashScreen onComplete={onComplete} subtitle="Custom subtitle" />)
    expect(screen.getByText('Custom subtitle')).toBeInTheDocument()
  })

  it('renders the footer text', () => {
    render(<SplashScreen onComplete={onComplete} />)
    expect(screen.getByText(/Made with/)).toBeInTheDocument()
    expect(screen.getByText(/Vasic Digital/)).toBeInTheDocument()
  })

  it('renders the version number', () => {
    render(<SplashScreen onComplete={onComplete} />)
    expect(screen.getByText('v1.1.0')).toBeInTheDocument()
  })

  it('renders the app icon image', () => {
    render(<SplashScreen onComplete={onComplete} />)
    const mark = screen.getByLabelText('Catalogizer icon')
    expect(mark).toBeInTheDocument()
    expect(mark).toHaveTextContent('C')
  })

  it('renders the loading spinner', () => {
    const { container } = render(<SplashScreen onComplete={onComplete} />)
    const spinner = container.querySelector('.animate-spin')
    expect(spinner).toBeInTheDocument()
  })

  describe('first launch (5s minimum)', () => {
    it('does not call onComplete before 5 seconds', () => {
      render(<SplashScreen onComplete={onComplete} />)

      act(() => {
        vi.advanceTimersByTime(4999)
      })

      expect(onComplete).not.toHaveBeenCalled()
    })

    it('calls onComplete after 5 seconds', () => {
      render(<SplashScreen onComplete={onComplete} />)

      act(() => {
        vi.advanceTimersByTime(5000)
      })

      expect(onComplete).toHaveBeenCalledTimes(1)
    })

    it('sets the localStorage flag', () => {
      render(<SplashScreen onComplete={onComplete} />)
      expect(localStorage.getItem('catalogizer_launched')).toBe('true')
    })
  })

  describe('regular launch (2.5s minimum)', () => {
    beforeEach(() => {
      localStorage.setItem('catalogizer_launched', 'true')
    })

    it('does not call onComplete before 2.5 seconds', () => {
      render(<SplashScreen onComplete={onComplete} />)

      act(() => {
        vi.advanceTimersByTime(2499)
      })

      expect(onComplete).not.toHaveBeenCalled()
    })

    it('calls onComplete after 2.5 seconds', () => {
      render(<SplashScreen onComplete={onComplete} />)

      act(() => {
        vi.advanceTimersByTime(2500)
      })

      expect(onComplete).toHaveBeenCalledTimes(1)
    })
  })
})
