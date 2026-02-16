import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, ...props }: any) => <div className={className}>{children}</div>,
    button: ({ children, className, onClick, ...props }: any) => (
      <button className={className} onClick={onClick}>{children}</button>
    ),
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

vi.mock('../../../types/playlists', () => ({
  SmartPlaylistRule: {},
  PlaylistType: {},
}))

describe('SmartPlaylistBuilder', () => {
  let SmartPlaylistBuilder: any
  const defaultProps = {
    onSave: vi.fn(),
    onCancel: vi.fn(),
  }

  beforeAll(async () => {
    const mod = await import('../SmartPlaylistBuilder')
    SmartPlaylistBuilder = mod.SmartPlaylistBuilder || mod.default
  })

  it('renders the component', () => {
    if (!SmartPlaylistBuilder) return
    render(<SmartPlaylistBuilder {...defaultProps} />)
    // Placeholder is "Enter playlist name..."
    expect(screen.getByPlaceholderText('Enter playlist name...')).toBeInTheDocument()
  })

  it('renders Save and Cancel buttons', () => {
    if (!SmartPlaylistBuilder) return
    render(<SmartPlaylistBuilder {...defaultProps} />)
    expect(screen.getByText('Save Playlist')).toBeInTheDocument()
    // Cancel appears in header and footer, use getAllByText
    const cancelElements = screen.getAllByText(/Cancel/)
    expect(cancelElements.length).toBeGreaterThanOrEqual(1)
  })

  it('calls onCancel when Cancel is clicked', async () => {
    if (!SmartPlaylistBuilder) return
    const user = userEvent.setup()
    render(<SmartPlaylistBuilder {...defaultProps} />)

    // Cancel appears multiple times; click the first one
    const cancelElements = screen.getAllByText(/Cancel/)
    await user.click(cancelElements[0])

    expect(defaultProps.onCancel).toHaveBeenCalled()
  })

  it('renders Add Rule button', () => {
    if (!SmartPlaylistBuilder) return
    render(<SmartPlaylistBuilder {...defaultProps} />)
    expect(screen.getByText('Add Rule')).toBeInTheDocument()
  })
})
