import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import CollectionAutomation from '../CollectionAutomation'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, ...props }: any) => <div className={className}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

describe('CollectionAutomation', () => {
  it('renders heading and description', () => {
    render(<CollectionAutomation />)
    expect(screen.getByText('Automation Rules')).toBeInTheDocument()
    expect(
      screen.getByText('Create workflows to automatically manage your collections')
    ).toBeInTheDocument()
  })

  it('renders Create Rule button', () => {
    render(<CollectionAutomation />)
    expect(screen.getByText('Create Rule')).toBeInTheDocument()
  })

  it('displays stats cards', () => {
    render(<CollectionAutomation />)
    expect(screen.getByText('Total Rules')).toBeInTheDocument()
    expect(screen.getByText('Active')).toBeInTheDocument()
    expect(screen.getByText('Total Runs')).toBeInTheDocument()
    expect(screen.getByText('Success Rate')).toBeInTheDocument()
  })

  it('displays mock automation rules', () => {
    render(<CollectionAutomation />)
    expect(screen.getByText('Auto-Tag New Movies')).toBeInTheDocument()
    expect(screen.getByText('Weekly Collection Cleanup')).toBeInTheDocument()
    expect(screen.getByText('Sync to External Drive')).toBeInTheDocument()
  })

  it('shows rule descriptions', () => {
    render(<CollectionAutomation />)
    expect(
      screen.getByText('Automatically tag new movie files with genre and year')
    ).toBeInTheDocument()
  })

  it('renders filter buttons', () => {
    render(<CollectionAutomation />)
    expect(screen.getByText('All')).toBeInTheDocument()
    // 'Enabled' appears both as filter button and badge, use getAllByText
    const enabledElements = screen.getAllByText('Enabled')
    expect(enabledElements.length).toBeGreaterThanOrEqual(1)
    // 'Disabled' also appears as both filter button and badge
    const disabledElements = screen.getAllByText('Disabled')
    expect(disabledElements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders search input', () => {
    render(<CollectionAutomation />)
    expect(screen.getByPlaceholderText('Search rules...')).toBeInTheDocument()
  })

  it('opens create modal when Create Rule is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionAutomation />)

    await user.click(screen.getByText('Create Rule'))

    expect(screen.getByText('Create Automation Rule')).toBeInTheDocument()
  })

  it('shows rule status badges', () => {
    render(<CollectionAutomation />)
    const enabledBadges = screen.getAllByText('Enabled')
    expect(enabledBadges.length).toBeGreaterThanOrEqual(1)
  })
})
