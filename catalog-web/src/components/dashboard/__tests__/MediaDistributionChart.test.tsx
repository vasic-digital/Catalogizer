import { render, screen } from '@testing-library/react'
import { MediaDistributionChart } from '../MediaDistributionChart'

// Mock recharts components since they don't render well in jsdom
vi.mock('recharts', () => ({
  PieChart: ({ children }: any) => <div data-testid="pie-chart">{children}</div>,
  Pie: ({ children }: any) => <div data-testid="pie">{children}</div>,
  Cell: () => <div data-testid="cell" />,
  ResponsiveContainer: ({ children }: any) => <div data-testid="responsive-container">{children}</div>,
  Legend: () => <div data-testid="legend" />,
  Tooltip: () => <div data-testid="tooltip" />,
}))

describe('MediaDistributionChart', () => {
  it('renders the chart title', () => {
    render(<MediaDistributionChart />)
    expect(screen.getByText('Media Distribution')).toBeInTheDocument()
  })

  it('shows empty state when no data provided', () => {
    render(<MediaDistributionChart />)
    expect(screen.getByText('No media data available')).toBeInTheDocument()
    expect(screen.getByText('Media will appear here once scanned')).toBeInTheDocument()
  })

  it('shows empty state when data is empty object', () => {
    render(<MediaDistributionChart data={{}} />)
    expect(screen.getByText('No media data available')).toBeInTheDocument()
  })

  it('shows loading state when loading is true', () => {
    render(<MediaDistributionChart loading />)
    expect(screen.getByText('Media Distribution')).toBeInTheDocument()
    // Loading state shows skeleton, not empty state or chart
    expect(screen.queryByText('No media data available')).not.toBeInTheDocument()
  })

  it('renders chart when data is provided', () => {
    const data = { video: 500, audio: 300, image: 200 }
    render(<MediaDistributionChart data={data} />)

    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
  })

  it('renders with single media type', () => {
    const data = { video: 100 }
    render(<MediaDistributionChart data={data} />)

    expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
  })

  it('renders with multiple media types', () => {
    const data = {
      video: 500,
      audio: 300,
      image: 200,
      document: 50,
      other: 10,
    }
    render(<MediaDistributionChart data={data} />)

    expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
  })

  it('does not render chart when loading', () => {
    const data = { video: 500, audio: 300 }
    render(<MediaDistributionChart data={data} loading />)

    expect(screen.queryByTestId('pie-chart')).not.toBeInTheDocument()
  })
})
