import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WebSocketProvider, useWebSocketContext } from '../WebSocketContext'

// Mock dependencies
const mockConnect = jest.fn()
const mockDisconnect = jest.fn()
const mockSend = jest.fn()
const mockSubscribe = jest.fn()
const mockUnsubscribe = jest.fn()
const mockGetConnectionState = jest.fn(() => 'closed')

jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}))

jest.mock('@/lib/websocket', () => ({
  useWebSocket: jest.fn(),
}))

const mockUseAuth = require('@/contexts/AuthContext').useAuth
const mockUseWebSocket = require('@/lib/websocket').useWebSocket

// Test component that uses the context
const TestComponent = () => {
  const { connect, disconnect, send, subscribe, unsubscribe, getConnectionState } =
    useWebSocketContext()

  return (
    <div>
      <button onClick={connect}>Connect</button>
      <button onClick={disconnect}>Disconnect</button>
      <button onClick={() => send({ type: 'test' })}>Send</button>
      <button onClick={() => subscribe('test-channel')}>Subscribe</button>
      <button onClick={() => unsubscribe('test-channel')}>Unsubscribe</button>
      <span data-testid="state">{getConnectionState()}</span>
    </div>
  )
}

describe('WebSocketContext', () => {
  beforeEach(() => {
    jest.clearAllMocks()

    // Default mock implementation
    mockUseWebSocket.mockReturnValue({
      connect: mockConnect,
      disconnect: mockDisconnect,
      send: mockSend,
      subscribe: mockSubscribe,
      unsubscribe: mockUnsubscribe,
      getConnectionState: mockGetConnectionState,
    })
  })

  describe('WebSocketProvider', () => {
    it('provides WebSocket context to children', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      expect(screen.getByRole('button', { name: /^connect$/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /^disconnect$/i })).toBeInTheDocument()
    })

    it('connects to WebSocket when user is authenticated', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      expect(mockConnect).toHaveBeenCalledTimes(1)
    })

    it('does not connect when user is not authenticated', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: false,
        user: null,
      })

      render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      expect(mockConnect).not.toHaveBeenCalled()
    })

    it('disconnects when user logs out', () => {
      const { rerender } = render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      // Initially authenticated
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      rerender(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      expect(mockConnect).toHaveBeenCalled()

      // User logs out
      mockUseAuth.mockReturnValue({
        isAuthenticated: false,
        user: null,
      })

      rerender(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      expect(mockDisconnect).toHaveBeenCalled()
    })

    it('disconnects on unmount', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      const { unmount } = render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      unmount()

      expect(mockDisconnect).toHaveBeenCalled()
    })

    it('reconnects when user changes', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'user1' },
      })

      const { rerender } = render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      const connectCallCount1 = mockConnect.mock.calls.length

      // Different user logs in
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 2, username: 'user2' },
      })

      rerender(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      // Should disconnect old connection and connect new one
      expect(mockDisconnect).toHaveBeenCalled()
      expect(mockConnect.mock.calls.length).toBeGreaterThan(connectCallCount1)
    })
  })

  describe('useWebSocketContext hook', () => {
    it('provides connect method', async () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      const user = userEvent.setup()

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      await user.click(screen.getByRole('button', { name: /^connect$/i }))

      expect(mockConnect).toHaveBeenCalled()
    })

    it('provides disconnect method', async () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      const user = userEvent.setup()

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      await user.click(screen.getByRole('button', { name: /^disconnect$/i }))

      expect(mockDisconnect).toHaveBeenCalled()
    })

    it('provides send method', async () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      const user = userEvent.setup()

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      await user.click(screen.getByRole('button', { name: /^send$/i }))

      expect(mockSend).toHaveBeenCalledWith({ type: 'test' })
    })

    it('provides subscribe method', async () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      const user = userEvent.setup()

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      await user.click(screen.getByRole('button', { name: /^subscribe$/i }))

      expect(mockSubscribe).toHaveBeenCalledWith('test-channel')
    })

    it('provides unsubscribe method', async () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      const user = userEvent.setup()

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      await user.click(screen.getByRole('button', { name: /^unsubscribe$/i }))

      expect(mockUnsubscribe).toHaveBeenCalledWith('test-channel')
    })

    it('provides getConnectionState method', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      mockGetConnectionState.mockReturnValue('open')

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      expect(screen.getByTestId('state')).toHaveTextContent('open')
      expect(mockGetConnectionState).toHaveBeenCalled()
    })

    it('throws error when used outside provider', () => {
      // Suppress console.error for this test
      const consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation()

      expect(() => {
        render(<TestComponent />)
      }).toThrow('useWebSocketContext must be used within a WebSocketProvider')

      consoleErrorSpy.mockRestore()
    })
  })

  describe('Connection States', () => {
    it('reflects connecting state', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      mockGetConnectionState.mockReturnValue('connecting')

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      expect(screen.getByTestId('state')).toHaveTextContent('connecting')
    })

    it('reflects open state', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      mockGetConnectionState.mockReturnValue('open')

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      expect(screen.getByTestId('state')).toHaveTextContent('open')
    })

    it('reflects closing state', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      mockGetConnectionState.mockReturnValue('closing')

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      expect(screen.getByTestId('state')).toHaveTextContent('closing')
    })

    it('reflects closed state', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      mockGetConnectionState.mockReturnValue('closed')

      render(
        <WebSocketProvider>
          <TestComponent />
        </WebSocketProvider>
      )

      expect(screen.getByTestId('state')).toHaveTextContent('closed')
    })
  })

  describe('Authentication Integration', () => {
    it('handles authentication without user object', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: null, // No user object
      })

      render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      // Should not connect without user
      expect(mockConnect).not.toHaveBeenCalled()
    })

    it('handles unauthenticated state', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: false,
        user: { id: 1, username: 'testuser' }, // User object exists but not authenticated
      })

      render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      expect(mockConnect).not.toHaveBeenCalled()
      expect(mockDisconnect).toHaveBeenCalled()
    })

    it('maintains connection when authentication state does not change', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      const { rerender } = render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      const connectCallCount = mockConnect.mock.calls.length

      // Re-render without changing auth state
      rerender(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      // Should not trigger additional connect calls
      expect(mockConnect.mock.calls.length).toBe(connectCallCount)
    })
  })

  describe('Edge Cases', () => {
    it('handles multiple children', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      render(
        <WebSocketProvider>
          <div>Child 1</div>
          <div>Child 2</div>
          <div>Child 3</div>
        </WebSocketProvider>
      )

      expect(screen.getByText('Child 1')).toBeInTheDocument()
      expect(screen.getByText('Child 2')).toBeInTheDocument()
      expect(screen.getByText('Child 3')).toBeInTheDocument()
    })

    it('handles nested providers', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, username: 'testuser' },
      })

      const NestedComponent = () => {
        const ws = useWebSocketContext()
        return <div data-testid="nested">Has context: {ws ? 'yes' : 'no'}</div>
      }

      render(
        <WebSocketProvider>
          <WebSocketProvider>
            <NestedComponent />
          </WebSocketProvider>
        </WebSocketProvider>
      )

      expect(screen.getByTestId('nested')).toHaveTextContent('Has context: yes')
    })

    it('handles rapid authentication changes', () => {
      const { rerender } = render(
        <WebSocketProvider>
          <div>Content</div>
        </WebSocketProvider>
      )

      // Cycle through states rapidly
      mockUseAuth.mockReturnValue({ isAuthenticated: true, user: { id: 1 } })
      rerender(<WebSocketProvider><div>Content</div></WebSocketProvider>)

      mockUseAuth.mockReturnValue({ isAuthenticated: false, user: null })
      rerender(<WebSocketProvider><div>Content</div></WebSocketProvider>)

      mockUseAuth.mockReturnValue({ isAuthenticated: true, user: { id: 2 } })
      rerender(<WebSocketProvider><div>Content</div></WebSocketProvider>)

      // Should handle all state changes gracefully
      expect(mockConnect).toHaveBeenCalled()
      expect(mockDisconnect).toHaveBeenCalled()
    })
  })
})
