import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { BrowserRouter, MemoryRouter } from 'react-router-dom'
import Layout from '../Layout'

// Mock authStore
const mockLogout = vi.fn()
vi.mock('../../stores/authStore', () => ({
  useAuthStore: () => ({
    user: {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      first_name: 'John',
      last_name: 'Doe',
      is_admin: false,
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    },
    logout: mockLogout,
  }),
}))

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  Home: (props: any) => <span data-testid="icon-home" {...props} />,
  Library: (props: any) => <span data-testid="icon-library" {...props} />,
  Search: (props: any) => <span data-testid="icon-search" {...props} />,
  Settings: (props: any) => <span data-testid="icon-settings" {...props} />,
  Film: (props: any) => <span data-testid="icon-film" {...props} />,
  LogOut: (props: any) => <span data-testid="icon-logout" {...props} />,
  User: (props: any) => <span data-testid="icon-user" {...props} />,
}))

const renderWithRouter = (ui: React.ReactElement, initialEntries = ['/']) => {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      {ui}
    </MemoryRouter>
  )
}

describe('Layout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the Catalogizer logo and title', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>)

    expect(screen.getByText('Catalogizer')).toBeInTheDocument()
  })

  it('renders navigation links', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>)

    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getByText('Library')).toBeInTheDocument()
    expect(screen.getByText('Search')).toBeInTheDocument()
  })

  it('renders the Settings link', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>)

    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('renders the Logout button', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>)

    expect(screen.getByText('Logout')).toBeInTheDocument()
  })

  it('renders user full name when available', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>)

    expect(screen.getByText('John Doe')).toBeInTheDocument()
  })

  it('renders user email', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>)

    expect(screen.getByText('test@example.com')).toBeInTheDocument()
  })

  it('renders children content', () => {
    renderWithRouter(<Layout><div>Test Content Here</div></Layout>)

    expect(screen.getByText('Test Content Here')).toBeInTheDocument()
  })

  it('calls logout when Logout button is clicked', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>)

    const logoutButton = screen.getByText('Logout')
    fireEvent.click(logoutButton)

    expect(mockLogout).toHaveBeenCalledTimes(1)
  })

  it('highlights the active navigation link for Home', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>, ['/'])

    const homeLink = screen.getByText('Home').closest('a')
    expect(homeLink).toHaveClass('bg-primary')
  })

  it('highlights the active navigation link for Library', () => {
    renderWithRouter(<Layout><div>Content</div></Layout>, ['/library'])

    const libraryLink = screen.getByText('Library').closest('a')
    expect(libraryLink).toHaveClass('bg-primary')
  })
})
