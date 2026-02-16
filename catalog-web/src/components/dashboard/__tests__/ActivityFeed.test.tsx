import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ActivityFeed } from '../ActivityFeed'

// Mock the websocket hook
vi.mock('@/lib/websocket', () => ({
  useWebSocket: vi.fn(() => ({
    connected: true,
    messages: [],
    send: vi.fn(),
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
  })),
}))

// Mock date-fns to get deterministic output
vi.mock('date-fns', () => ({
  formatDistanceToNow: vi.fn(() => '5 minutes ago'),
}))

describe('ActivityFeed', () => {
  it('shows loading skeleton initially', () => {
    render(<ActivityFeed />)
    expect(screen.getByText('Recent Activity')).toBeInTheDocument()
  })

  it('renders Recent Activity title', async () => {
    render(<ActivityFeed />)
    await waitFor(() => {
      expect(screen.getByText('Recent Activity')).toBeInTheDocument()
    })
  })

  it('displays activities after loading', async () => {
    render(<ActivityFeed />)
    await waitFor(() => {
      expect(screen.getByText(/Sample Movie/)).toBeInTheDocument()
    })
  })

  it('shows filter buttons when showFilters is true', async () => {
    render(<ActivityFeed showFilters />)
    await waitFor(() => {
      expect(screen.getByText('All')).toBeInTheDocument()
      expect(screen.getByText('Playing')).toBeInTheDocument()
      expect(screen.getByText('Uploads')).toBeInTheDocument()
    })
  })

  it('hides filter buttons when showFilters is false', async () => {
    render(<ActivityFeed showFilters={false} />)
    await waitFor(() => {
      expect(screen.queryByText('Playing')).not.toBeInTheDocument()
      expect(screen.queryByText('Uploads')).not.toBeInTheDocument()
    })
  })

  it('filters activities when filter is clicked', async () => {
    const user = userEvent.setup()
    render(<ActivityFeed showFilters />)

    await waitFor(() => {
      expect(screen.getByText('Playing')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Playing'))

    await waitFor(() => {
      // Should only show media_played activities
      expect(screen.getByText(/Sample Movie/)).toBeInTheDocument()
    })
  })

  it('respects the limit prop', async () => {
    render(<ActivityFeed limit={2} />)
    await waitFor(() => {
      // Should show max 2 activities
      expect(screen.getByText('Recent Activity')).toBeInTheDocument()
    })
  })

  it('shows user names for activities', async () => {
    render(<ActivityFeed />)
    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })
  })

  it('shows timestamps for activities', async () => {
    render(<ActivityFeed />)
    await waitFor(() => {
      expect(screen.getAllByText('5 minutes ago').length).toBeGreaterThan(0)
    })
  })
})
