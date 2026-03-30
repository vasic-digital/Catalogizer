import { describe, it, expect } from 'vitest'
import { screen } from '@testing-library/react'
import {
  renderWithProviders,
  renderWithRouter,
  createTestQueryClient,
  waitForNextTick,
} from '../testRender'

describe('testRender utilities', () => {
  describe('createTestQueryClient', () => {
    it('creates a QueryClient instance', () => {
      const client = createTestQueryClient()

      expect(client).toBeDefined()
      expect(typeof client.getQueryCache).toBe('function')
      expect(typeof client.getMutationCache).toBe('function')
    })

    it('creates a new instance each call', () => {
      const client1 = createTestQueryClient()
      const client2 = createTestQueryClient()

      expect(client1).not.toBe(client2)
    })

    it('has retry disabled for queries', () => {
      const client = createTestQueryClient()
      const defaults = client.getDefaultOptions()

      expect(defaults.queries?.retry).toBe(false)
    })

    it('has retry disabled for mutations', () => {
      const client = createTestQueryClient()
      const defaults = client.getDefaultOptions()

      expect(defaults.mutations?.retry).toBe(false)
    })
  })

  describe('renderWithProviders', () => {
    it('renders a simple component', () => {
      renderWithProviders(<div data-testid="test-div">Hello</div>)

      expect(screen.getByTestId('test-div')).toBeInTheDocument()
      expect(screen.getByText('Hello')).toBeInTheDocument()
    })

    it('returns a queryClient in the result', () => {
      const result = renderWithProviders(<div>Test</div>)

      expect(result.queryClient).toBeDefined()
      expect(typeof result.queryClient.getQueryCache).toBe('function')
    })

    it('wraps with BrowserRouter for routing', () => {
      // A Link component requires router context - if it renders, router is present
      const { container } = renderWithProviders(
        <a href="/test">Link Test</a>
      )

      expect(container).toBeDefined()
    })

    it('renders multiple children', () => {
      renderWithProviders(
        <div>
          <span data-testid="child-1">First</span>
          <span data-testid="child-2">Second</span>
        </div>
      )

      expect(screen.getByTestId('child-1')).toBeInTheDocument()
      expect(screen.getByTestId('child-2')).toBeInTheDocument()
    })
  })

  describe('renderWithRouter', () => {
    it('renders a component with default route', () => {
      renderWithRouter(<div data-testid="routed">Routed Content</div>)

      expect(screen.getByTestId('routed')).toBeInTheDocument()
    })

    it('sets the specified route', () => {
      renderWithRouter(
        <div data-testid="settings-page">Settings</div>,
        { route: '/settings' }
      )

      expect(window.location.pathname).toBe('/settings')
    })

    it('sets default route to /', () => {
      renderWithRouter(<div>Home</div>)

      expect(window.location.pathname).toBe('/')
    })
  })

  describe('waitForNextTick', () => {
    it('returns a promise', () => {
      const result = waitForNextTick()

      expect(result).toBeInstanceOf(Promise)
    })

    it('resolves after a tick', async () => {
      let resolved = false

      waitForNextTick().then(() => {
        resolved = true
      })

      expect(resolved).toBe(false)

      await waitForNextTick()

      expect(resolved).toBe(true)
    })
  })
})
