import { render, screen, waitFor } from '@testing-library/react'
import { ConnectionStatus } from '../ConnectionStatus'
import { useWebSocket } from '@/lib/websocket'

// Mock the websocket hook
vi.mock('@/lib/websocket', async () => ({
  useWebSocket: vi.fn(),
}))

// Mock framer-motion to avoid animation issues in tests
vi.mock('framer-motion', async () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

const mockUseWebSocket = vi.mocked(useWebSocket)

describe('ConnectionStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  describe('Connection States', () => {
    it('displays connecting status when connection state is connecting', () => {
      mockUseWebSocket.mockReturnValue({
        getConnectionState: vi.fn().mockReturnValue('connecting'),
      })

      render(<ConnectionStatus />)

      expect(screen.getByText('Connecting...')).toBeInTheDocument()
      expect(document.querySelector('.animate-spin')).toBeInTheDocument()
    })

    it('does not display status when connection state is open', () => {
      mockUseWebSocket.mockReturnValue({
        getConnectionState: vi.fn().mockReturnValue('open'),
      })

      render(<ConnectionStatus />)

      expect(screen.queryByText('Connected')).not.toBeInTheDocument()
      expect(screen.queryByText('Connecting...')).not.toBeInTheDocument()
      expect(screen.queryByText('Disconnected')).not.toBeInTheDocument()
    })

    it('displays disconnecting status when connection state is closing', () => {
      mockUseWebSocket.mockReturnValue({
        getConnectionState: vi.fn().mockReturnValue('closing'),
      })

      render(<ConnectionStatus />)

      expect(screen.getByText('Disconnecting...')).toBeInTheDocument()
    })

    it('displays disconnected status when connection state is closed', () => {
      mockUseWebSocket.mockReturnValue({
        getConnectionState: vi.fn().mockReturnValue('closed'),
      })

      render(<ConnectionStatus />)

      expect(screen.getByText('Disconnected')).toBeInTheDocument()
    })
  })

  describe('Status Colors', () => {
    it('applies yellow background for connecting state', () => {
      mockUseWebSocket.mockReturnValue({
        getConnectionState: vi.fn().mockReturnValue('connecting'),
      })

      render(<ConnectionStatus />)

      const statusElement = screen.getByText('Connecting...').parentElement
      expect(statusElement).toHaveClass('bg-yellow-500')
    })

    it('applies red background for disconnected state', () => {
      mockUseWebSocket.mockReturnValue({
        getConnectionState: vi.fn().mockReturnValue('closed'),
      })

      render(<ConnectionStatus />)

      const statusElement = screen.getByText('Disconnected').parentElement
      expect(statusElement).toHaveClass('bg-red-500')
    })

    it('applies orange background for disconnecting state', () => {
      mockUseWebSocket.mockReturnValue({
        getConnectionState: vi.fn().mockReturnValue('closing'),
      })

      render(<ConnectionStatus />)

      const statusElement = screen.getByText('Disconnecting...').parentElement
      expect(statusElement).toHaveClass('bg-orange-500')
    })
  })

  describe('Dynamic State Changes', () => {
    it('updates status when connection state changes', async () => {
      const getConnectionState = vi.fn().mockReturnValue('connecting')
      mockUseWebSocket.mockReturnValue({
        getConnectionState,
      })

      const { rerender } = render(<ConnectionStatus />)

      expect(screen.getByText('Connecting...')).toBeInTheDocument()

      // Change state to closed
      getConnectionState.mockReturnValue('closed')

      // Fast-forward time to trigger the interval
      vi.advanceTimersByTime(1000)

      // Wait for the state update
      await waitFor(() => {
        expect(screen.getByText('Disconnected')).toBeInTheDocument()
      })
    })

    it('hides status when connection becomes open', async () => {
      const getConnectionState = vi.fn().mockReturnValue('connecting')
      mockUseWebSocket.mockReturnValue({
        getConnectionState,
      })

      render(<ConnectionStatus />)

      expect(screen.getByText('Connecting...')).toBeInTheDocument()

      // Change state to open
      getConnectionState.mockReturnValue('open')

      // Fast-forward time to trigger the interval
      vi.advanceTimersByTime(1000)

      // Wait for the status to be hidden
      await waitFor(() => {
        expect(screen.queryByText('Connected')).not.toBeInTheDocument()
      })
    })
  })

  describe('Interval Updates', () => {
    it('checks connection state every second', () => {
      const getConnectionState = vi.fn().mockReturnValue('closed')
      mockUseWebSocket.mockReturnValue({
        getConnectionState,
      })

      render(<ConnectionStatus />)

      // Initial call
      expect(getConnectionState).toHaveBeenCalledTimes(1)

      // After 1 second
      vi.advanceTimersByTime(1000)
      expect(getConnectionState).toHaveBeenCalledTimes(2)

      // After 3 more seconds
      vi.advanceTimersByTime(3000)
      expect(getConnectionState).toHaveBeenCalledTimes(5)
    })

    it('cleans up interval on unmount', () => {
      const getConnectionState = vi.fn().mockReturnValue('closed')
      mockUseWebSocket.mockReturnValue({
        getConnectionState,
      })

      const { unmount } = render(<ConnectionStatus />)

      expect(getConnectionState).toHaveBeenCalledTimes(1)

      unmount()

      // Advance time after unmount
      vi.advanceTimersByTime(5000)

      // Should not call again after unmount
      expect(getConnectionState).toHaveBeenCalledTimes(1)
    })
  })

  describe('Visibility Logic', () => {
    it('shows status only when not connected', () => {
      const testCases = [
        { state: 'connecting', shouldShow: true },
        { state: 'open', shouldShow: false },
        { state: 'closing', shouldShow: true },
        { state: 'closed', shouldShow: true },
      ]

      testCases.forEach(({ state, shouldShow }) => {
        mockUseWebSocket.mockReturnValue({
          getConnectionState: vi.fn().mockReturnValue(state),
        })

        const { unmount } = render(<ConnectionStatus />)

        const statusTexts = ['Connecting...', 'Connected', 'Disconnecting...', 'Disconnected']
        const hasAnyStatus = statusTexts.some((text) => screen.queryByText(text) !== null)

        expect(hasAnyStatus).toBe(shouldShow)

        unmount()
      })
    })
  })
})
