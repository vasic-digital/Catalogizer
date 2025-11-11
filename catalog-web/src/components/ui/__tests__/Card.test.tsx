import React from 'react'
import { render, screen } from '@testing-library/react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '../Card'

describe('Card', () => {
  describe('Card Component', () => {
    it('renders children correctly', () => {
      render(<Card>Card content</Card>)
      expect(screen.getByText('Card content')).toBeInTheDocument()
    })

    it('applies default classes', () => {
      render(<Card data-testid="card">Card content</Card>)
      const card = screen.getByTestId('card')
      expect(card).toHaveClass('rounded-xl')
      expect(card).toHaveClass('border')
      expect(card).toHaveClass('bg-white')
      expect(card).toHaveClass('shadow-lg')
    })

    it('applies custom className', () => {
      render(
        <Card className="custom-class" data-testid="card">
          Card content
        </Card>
      )
      const card = screen.getByTestId('card')
      expect(card).toHaveClass('custom-class')
      expect(card).toHaveClass('rounded-xl') // Still has default classes
    })

    it('forwards ref correctly', () => {
      const ref = React.createRef<HTMLDivElement>()
      render(<Card ref={ref}>Card content</Card>)
      expect(ref.current).toBeInstanceOf(HTMLDivElement)
      expect(ref.current?.textContent).toBe('Card content')
    })

    it('spreads additional props', () => {
      render(
        <Card data-testid="card" data-custom="value">
          Card content
        </Card>
      )
      const card = screen.getByTestId('card')
      expect(card).toHaveAttribute('data-custom', 'value')
    })
  })

  describe('CardHeader Component', () => {
    it('renders children correctly', () => {
      render(<CardHeader>Header content</CardHeader>)
      expect(screen.getByText('Header content')).toBeInTheDocument()
    })

    it('applies default classes', () => {
      render(<CardHeader data-testid="header">Header content</CardHeader>)
      const header = screen.getByTestId('header')
      expect(header).toHaveClass('flex')
      expect(header).toHaveClass('flex-col')
      expect(header).toHaveClass('space-y-1.5')
      expect(header).toHaveClass('p-6')
    })

    it('applies custom className', () => {
      render(
        <CardHeader className="custom-header" data-testid="header">
          Header content
        </CardHeader>
      )
      const header = screen.getByTestId('header')
      expect(header).toHaveClass('custom-header')
      expect(header).toHaveClass('flex') // Still has default classes
    })

    it('forwards ref correctly', () => {
      const ref = React.createRef<HTMLDivElement>()
      render(<CardHeader ref={ref}>Header content</CardHeader>)
      expect(ref.current).toBeInstanceOf(HTMLDivElement)
    })
  })

  describe('CardTitle Component', () => {
    it('renders children correctly', () => {
      render(<CardTitle>Title text</CardTitle>)
      expect(screen.getByText('Title text')).toBeInTheDocument()
    })

    it('renders as h3 by default', () => {
      render(<CardTitle>Title text</CardTitle>)
      const title = screen.getByText('Title text')
      expect(title.tagName).toBe('H3')
    })

    it('applies default classes', () => {
      render(<CardTitle>Title text</CardTitle>)
      const title = screen.getByText('Title text')
      expect(title).toHaveClass('font-semibold')
      expect(title).toHaveClass('leading-none')
      expect(title).toHaveClass('tracking-tight')
    })

    it('applies custom className', () => {
      render(<CardTitle className="custom-title">Title text</CardTitle>)
      const title = screen.getByText('Title text')
      expect(title).toHaveClass('custom-title')
      expect(title).toHaveClass('font-semibold') // Still has default classes
    })

    it('forwards ref correctly', () => {
      const ref = React.createRef<HTMLParagraphElement>()
      render(<CardTitle ref={ref}>Title text</CardTitle>)
      expect(ref.current).toBeInstanceOf(HTMLHeadingElement)
    })
  })

  describe('CardDescription Component', () => {
    it('renders children correctly', () => {
      render(<CardDescription>Description text</CardDescription>)
      expect(screen.getByText('Description text')).toBeInTheDocument()
    })

    it('renders as paragraph', () => {
      render(<CardDescription>Description text</CardDescription>)
      const description = screen.getByText('Description text')
      expect(description.tagName).toBe('P')
    })

    it('applies default classes', () => {
      render(<CardDescription>Description text</CardDescription>)
      const description = screen.getByText('Description text')
      expect(description).toHaveClass('text-sm')
      expect(description).toHaveClass('text-gray-600')
    })

    it('applies custom className', () => {
      render(
        <CardDescription className="custom-description">
          Description text
        </CardDescription>
      )
      const description = screen.getByText('Description text')
      expect(description).toHaveClass('custom-description')
      expect(description).toHaveClass('text-sm') // Still has default classes
    })

    it('forwards ref correctly', () => {
      const ref = React.createRef<HTMLParagraphElement>()
      render(<CardDescription ref={ref}>Description text</CardDescription>)
      expect(ref.current).toBeInstanceOf(HTMLParagraphElement)
    })
  })

  describe('CardContent Component', () => {
    it('renders children correctly', () => {
      render(<CardContent>Content text</CardContent>)
      expect(screen.getByText('Content text')).toBeInTheDocument()
    })

    it('applies default classes', () => {
      render(<CardContent data-testid="content">Content text</CardContent>)
      const content = screen.getByTestId('content')
      expect(content).toHaveClass('p-6')
      expect(content).toHaveClass('pt-0')
    })

    it('applies custom className', () => {
      render(
        <CardContent className="custom-content" data-testid="content">
          Content text
        </CardContent>
      )
      const content = screen.getByTestId('content')
      expect(content).toHaveClass('custom-content')
      expect(content).toHaveClass('p-6') // Still has default classes
    })

    it('forwards ref correctly', () => {
      const ref = React.createRef<HTMLDivElement>()
      render(<CardContent ref={ref}>Content text</CardContent>)
      expect(ref.current).toBeInstanceOf(HTMLDivElement)
    })
  })

  describe('CardFooter Component', () => {
    it('renders children correctly', () => {
      render(<CardFooter>Footer content</CardFooter>)
      expect(screen.getByText('Footer content')).toBeInTheDocument()
    })

    it('applies default classes', () => {
      render(<CardFooter data-testid="footer">Footer content</CardFooter>)
      const footer = screen.getByTestId('footer')
      expect(footer).toHaveClass('flex')
      expect(footer).toHaveClass('items-center')
      expect(footer).toHaveClass('p-6')
      expect(footer).toHaveClass('pt-0')
    })

    it('applies custom className', () => {
      render(
        <CardFooter className="custom-footer" data-testid="footer">
          Footer content
        </CardFooter>
      )
      const footer = screen.getByTestId('footer')
      expect(footer).toHaveClass('custom-footer')
      expect(footer).toHaveClass('flex') // Still has default classes
    })

    it('forwards ref correctly', () => {
      const ref = React.createRef<HTMLDivElement>()
      render(<CardFooter ref={ref}>Footer content</CardFooter>)
      expect(ref.current).toBeInstanceOf(HTMLDivElement)
    })
  })

  describe('Card Composition', () => {
    it('renders complete card with all sub-components', () => {
      render(
        <Card data-testid="complete-card">
          <CardHeader>
            <CardTitle>Card Title</CardTitle>
            <CardDescription>Card Description</CardDescription>
          </CardHeader>
          <CardContent>Card Content</CardContent>
          <CardFooter>Card Footer</CardFooter>
        </Card>
      )

      expect(screen.getByTestId('complete-card')).toBeInTheDocument()
      expect(screen.getByText('Card Title')).toBeInTheDocument()
      expect(screen.getByText('Card Description')).toBeInTheDocument()
      expect(screen.getByText('Card Content')).toBeInTheDocument()
      expect(screen.getByText('Card Footer')).toBeInTheDocument()
    })

    it('renders card with only some sub-components', () => {
      render(
        <Card>
          <CardHeader>
            <CardTitle>Title Only</CardTitle>
          </CardHeader>
          <CardContent>Content Only</CardContent>
        </Card>
      )

      expect(screen.getByText('Title Only')).toBeInTheDocument()
      expect(screen.getByText('Content Only')).toBeInTheDocument()
      expect(screen.queryByText('Card Footer')).not.toBeInTheDocument()
    })

    it('renders nested content correctly', () => {
      render(
        <Card>
          <CardHeader>
            <CardTitle>Product Card</CardTitle>
            <CardDescription>
              <span>Price: $99.99</span>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div>
              <p>Feature 1</p>
              <p>Feature 2</p>
            </div>
          </CardContent>
          <CardFooter>
            <button>Buy Now</button>
          </CardFooter>
        </Card>
      )

      expect(screen.getByText('Product Card')).toBeInTheDocument()
      expect(screen.getByText('Price: $99.99')).toBeInTheDocument()
      expect(screen.getByText('Feature 1')).toBeInTheDocument()
      expect(screen.getByText('Feature 2')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /buy now/i })).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('CardTitle has proper heading semantics', () => {
      render(
        <Card>
          <CardHeader>
            <CardTitle>Accessible Title</CardTitle>
          </CardHeader>
        </Card>
      )

      const title = screen.getByText('Accessible Title')
      expect(title).toHaveAttribute('class')
      expect(title.tagName).toBe('H3')
    })

    it('CardDescription has proper paragraph semantics', () => {
      render(
        <Card>
          <CardHeader>
            <CardDescription>Accessible Description</CardDescription>
          </CardHeader>
        </Card>
      )

      const description = screen.getByText('Accessible Description')
      expect(description.tagName).toBe('P')
    })

    it('supports aria attributes', () => {
      render(
        <Card aria-label="Product card" data-testid="card">
          <CardContent>Content</CardContent>
        </Card>
      )

      const card = screen.getByTestId('card')
      expect(card).toHaveAttribute('aria-label', 'Product card')
    })

    it('supports role attribute', () => {
      render(
        <Card role="article" data-testid="card">
          <CardContent>Content</CardContent>
        </Card>
      )

      const card = screen.getByTestId('card')
      expect(card).toHaveAttribute('role', 'article')
    })
  })

  describe('Edge Cases', () => {
    it('renders empty Card', () => {
      render(<Card data-testid="empty-card" />)
      const card = screen.getByTestId('empty-card')
      expect(card).toBeInTheDocument()
      expect(card.textContent).toBe('')
    })

    it('renders Card with null children', () => {
      render(
        <Card data-testid="null-card">
          {null}
        </Card>
      )
      const card = screen.getByTestId('null-card')
      expect(card).toBeInTheDocument()
    })

    it('renders Card with boolean children', () => {
      render(
        <Card data-testid="boolean-card">
          {true && <div>Conditional content</div>}
          {false && <div>Hidden content</div>}
        </Card>
      )
      const card = screen.getByTestId('boolean-card')
      expect(screen.getByText('Conditional content')).toBeInTheDocument()
      expect(screen.queryByText('Hidden content')).not.toBeInTheDocument()
    })

    it('handles multiple CardTitle components', () => {
      render(
        <Card>
          <CardHeader>
            <CardTitle>First Title</CardTitle>
            <CardTitle>Second Title</CardTitle>
          </CardHeader>
        </Card>
      )

      expect(screen.getByText('First Title')).toBeInTheDocument()
      expect(screen.getByText('Second Title')).toBeInTheDocument()
    })

    it('handles very long content', () => {
      const longText = 'A'.repeat(1000)
      render(
        <Card>
          <CardContent>{longText}</CardContent>
        </Card>
      )

      expect(screen.getByText(longText)).toBeInTheDocument()
    })
  })
})
