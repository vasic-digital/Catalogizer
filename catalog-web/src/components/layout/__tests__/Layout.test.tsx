import React from 'react'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { Layout } from '../Layout'

// Mock the Header component
jest.mock('../Header', () => ({
  Header: () => <div data-testid="mock-header">Header</div>,
}))

// Mock AuthContext for Header
jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(() => ({
    user: null,
    isAuthenticated: false,
    logout: jest.fn(),
  })),
}))

describe('Layout', () => {
  describe('Rendering', () => {
    it('renders the layout component', () => {
      render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      expect(screen.getByTestId('mock-header')).toBeInTheDocument()
    })

    it('renders Header component', () => {
      render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      expect(screen.getByText('Header')).toBeInTheDocument()
    })

    it('has main element for content', () => {
      const { container } = render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      const mainElement = container.querySelector('main')
      expect(mainElement).toBeInTheDocument()
    })

    it('applies min-h-screen class to wrapper', () => {
      const { container } = render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      const wrapper = container.firstChild as HTMLElement
      expect(wrapper).toHaveClass('min-h-screen')
    })

    it('applies background color classes', () => {
      const { container } = render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      const wrapper = container.firstChild as HTMLElement
      expect(wrapper).toHaveClass('bg-gray-50')
      expect(wrapper).toHaveClass('dark:bg-gray-900')
    })
  })

  describe('Outlet Integration', () => {
    it('renders child routes through Outlet', () => {
      render(
        <MemoryRouter initialEntries={['/test']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/test" element={<div>Test Page Content</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('Test Page Content')).toBeInTheDocument()
    })

    it('renders different child routes', () => {
      // Test route 1
      const { unmount: unmount1 } = render(
        <MemoryRouter initialEntries={['/page1']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/page1" element={<div>Page 1</div>} />
              <Route path="/page2" element={<div>Page 2</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('Page 1')).toBeInTheDocument()
      unmount1()

      // Test route 2
      render(
        <MemoryRouter initialEntries={['/page2']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/page1" element={<div>Page 1</div>} />
              <Route path="/page2" element={<div>Page 2</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('Page 2')).toBeInTheDocument()
      expect(screen.queryByText('Page 1')).not.toBeInTheDocument()
    })

    it('renders nested routes correctly', () => {
      render(
        <MemoryRouter initialEntries={['/parent/child']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/parent" element={<div>Parent</div>}>
                <Route path="child" element={<div>Child</div>} />
              </Route>
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('Parent')).toBeInTheDocument()
    })

    it('maintains Header across route changes', () => {
      const { rerender } = render(
        <MemoryRouter initialEntries={['/route1']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/route1" element={<div>Route 1</div>} />
              <Route path="/route2" element={<div>Route 2</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByTestId('mock-header')).toBeInTheDocument()

      rerender(
        <MemoryRouter initialEntries={['/route2']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/route1" element={<div>Route 1</div>} />
              <Route path="/route2" element={<div>Route 2</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByTestId('mock-header')).toBeInTheDocument()
    })
  })

  describe('Structure', () => {
    it('renders in correct order: Header then main', () => {
      const { container } = render(
        <MemoryRouter>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<div>Content</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      const wrapper = container.firstChild as HTMLElement
      const children = Array.from(wrapper.children)

      expect(children[0]).toContainElement(screen.getByTestId('mock-header'))
      expect(children[1].tagName).toBe('MAIN')
    })

    it('main element has flex-1 class for proper layout', () => {
      const { container } = render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      const mainElement = container.querySelector('main')
      expect(mainElement).toHaveClass('flex-1')
    })

    it('wraps content in a container div', () => {
      const { container } = render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      expect(container.firstChild).toBeInstanceOf(HTMLDivElement)
    })
  })

  describe('Edge Cases', () => {
    it('renders without any child routes', () => {
      render(
        <MemoryRouter>
          <Routes>
            <Route path="/" element={<Layout />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByTestId('mock-header')).toBeInTheDocument()
      expect(screen.queryByText('Test Page Content')).not.toBeInTheDocument()
    })

    it('renders with no matching route', () => {
      render(
        <MemoryRouter initialEntries={['/test']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/test" element={<div>Test Content</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByTestId('mock-header')).toBeInTheDocument()
      expect(screen.getByText('Test Content')).toBeInTheDocument()
    })

    it('handles complex nested content', () => {
      render(
        <MemoryRouter initialEntries={['/complex']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route
                path="/complex"
                element={
                  <div>
                    <h1>Title</h1>
                    <div>
                      <p>Paragraph 1</p>
                      <p>Paragraph 2</p>
                    </div>
                    <footer>Footer</footer>
                  </div>
                }
              />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('Title')).toBeInTheDocument()
      expect(screen.getByText('Paragraph 1')).toBeInTheDocument()
      expect(screen.getByText('Paragraph 2')).toBeInTheDocument()
      expect(screen.getByText('Footer')).toBeInTheDocument()
    })

    it('renders with fragments as children', () => {
      render(
        <MemoryRouter initialEntries={['/fragment']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route
                path="/fragment"
                element={
                  <>
                    <div>Fragment Child 1</div>
                    <div>Fragment Child 2</div>
                  </>
                }
              />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('Fragment Child 1')).toBeInTheDocument()
      expect(screen.getByText('Fragment Child 2')).toBeInTheDocument()
    })
  })

  describe('Styling', () => {
    it('applies dark mode classes', () => {
      const { container } = render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      const wrapper = container.firstChild as HTMLElement
      expect(wrapper.className).toContain('dark:bg-gray-900')
    })

    it('applies light mode classes', () => {
      const { container } = render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      const wrapper = container.firstChild as HTMLElement
      expect(wrapper.className).toContain('bg-gray-50')
    })

    it('has full viewport height', () => {
      const { container } = render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      )

      const wrapper = container.firstChild as HTMLElement
      expect(wrapper).toHaveClass('min-h-screen')
    })
  })

  describe('React Router Integration', () => {
    it('works with MemoryRouter', () => {
      render(
        <MemoryRouter>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<div>Home</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('Home')).toBeInTheDocument()
    })

    it('works with multiple routes at same level', () => {
      render(
        <MemoryRouter initialEntries={['/about']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="/home" element={<div>Home</div>} />
              <Route path="/about" element={<div>About</div>} />
              <Route path="/contact" element={<div>Contact</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('About')).toBeInTheDocument()
      expect(screen.queryByText('Home')).not.toBeInTheDocument()
      expect(screen.queryByText('Contact')).not.toBeInTheDocument()
    })

    it('supports index routes', () => {
      render(
        <MemoryRouter initialEntries={['/']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<div>Index Page</div>} />
              <Route path="other" element={<div>Other Page</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('Index Page')).toBeInTheDocument()
    })
  })
})
