import { render, screen } from '@testing-library/react'
import { PlaylistAnalytics } from '../PlaylistAnalytics'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, ...props }: any) => <div className={className}>{children}</div>,
  },
}))

const mockAnalytics = {
  total_plays: 15000,
  unique_viewers: 250,
  average_completion_rate: 75,
  popular_items: [
    { id: '1', title: 'Popular Song', play_count: 500 },
  ],
  viewing_stats: [
    { date: '2024-01-01', plays: 100, viewers: 50 },
  ],
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-15T00:00:00Z',
}

describe('PlaylistAnalytics', () => {
  it('renders analytics stats', () => {
    render(<PlaylistAnalytics analytics={mockAnalytics as any} />)
    expect(screen.getByText('Total Plays')).toBeInTheDocument()
  })

  it('displays formatted play count', () => {
    render(<PlaylistAnalytics analytics={mockAnalytics as any} />)
    // 15000 should be formatted as 15.0K
    expect(screen.getByText('15.0K')).toBeInTheDocument()
  })

  it('displays unique viewers stat', () => {
    render(<PlaylistAnalytics analytics={mockAnalytics as any} />)
    expect(screen.getByText('Unique Viewers')).toBeInTheDocument()
  })

  it('displays completion rate stat', () => {
    render(<PlaylistAnalytics analytics={mockAnalytics as any} />)
    expect(screen.getByText('Completion Rate')).toBeInTheDocument()
  })

  it('shows popular items section', () => {
    render(<PlaylistAnalytics analytics={mockAnalytics as any} />)
    expect(screen.getByText('Popular Items')).toBeInTheDocument()
    expect(screen.getByText('Popular Song')).toBeInTheDocument()
  })

  it('shows viewing trends section', () => {
    render(<PlaylistAnalytics analytics={mockAnalytics as any} />)
    expect(screen.getByText('Viewing Trends')).toBeInTheDocument()
  })
})
