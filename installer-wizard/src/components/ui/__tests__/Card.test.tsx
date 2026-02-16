import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '../Card'

describe('Card', () => {
  it('renders card with children', () => {
    render(<Card>Card content</Card>)

    expect(screen.getByText('Card content')).toBeInTheDocument()
  })

  it('applies default card styling', () => {
    render(<Card data-testid="card">Content</Card>)

    const card = screen.getByTestId('card')
    expect(card).toHaveClass('rounded-lg')
    expect(card).toHaveClass('border')
    expect(card).toHaveClass('bg-card')
    expect(card).toHaveClass('shadow-sm')
  })

  it('applies additional className', () => {
    render(<Card data-testid="card" className="my-custom-class">Content</Card>)

    const card = screen.getByTestId('card')
    expect(card).toHaveClass('my-custom-class')
  })

  it('forwards ref', () => {
    const ref = vi.fn()
    render(<Card ref={ref}>Content</Card>)

    expect(ref).toHaveBeenCalled()
    expect(ref.mock.calls[0][0]).toBeInstanceOf(HTMLDivElement)
  })

  it('has the correct displayName', () => {
    expect(Card.displayName).toBe('Card')
  })
})

describe('CardHeader', () => {
  it('renders header content', () => {
    render(<CardHeader>Header content</CardHeader>)

    expect(screen.getByText('Header content')).toBeInTheDocument()
  })

  it('applies default header styling', () => {
    render(<CardHeader data-testid="card-header">Header</CardHeader>)

    const header = screen.getByTestId('card-header')
    expect(header).toHaveClass('flex')
    expect(header).toHaveClass('flex-col')
    expect(header).toHaveClass('p-6')
  })

  it('has the correct displayName', () => {
    expect(CardHeader.displayName).toBe('CardHeader')
  })
})

describe('CardTitle', () => {
  it('renders title text', () => {
    render(<CardTitle>My Title</CardTitle>)

    expect(screen.getByText('My Title')).toBeInTheDocument()
  })

  it('renders as h3 element', () => {
    render(<CardTitle>My Title</CardTitle>)

    const title = screen.getByText('My Title')
    expect(title.tagName).toBe('H3')
  })

  it('applies default title styling', () => {
    render(<CardTitle>Title</CardTitle>)

    const title = screen.getByText('Title')
    expect(title).toHaveClass('text-2xl')
    expect(title).toHaveClass('font-semibold')
  })

  it('has the correct displayName', () => {
    expect(CardTitle.displayName).toBe('CardTitle')
  })
})

describe('CardDescription', () => {
  it('renders description text', () => {
    render(<CardDescription>A description</CardDescription>)

    expect(screen.getByText('A description')).toBeInTheDocument()
  })

  it('renders as p element', () => {
    render(<CardDescription>A description</CardDescription>)

    const desc = screen.getByText('A description')
    expect(desc.tagName).toBe('P')
  })

  it('applies default description styling', () => {
    render(<CardDescription>Description</CardDescription>)

    const desc = screen.getByText('Description')
    expect(desc).toHaveClass('text-sm')
    expect(desc).toHaveClass('text-muted-foreground')
  })

  it('has the correct displayName', () => {
    expect(CardDescription.displayName).toBe('CardDescription')
  })
})

describe('CardContent', () => {
  it('renders content', () => {
    render(<CardContent>Content here</CardContent>)

    expect(screen.getByText('Content here')).toBeInTheDocument()
  })

  it('applies default content styling', () => {
    render(<CardContent data-testid="card-content">Content</CardContent>)

    const content = screen.getByTestId('card-content')
    expect(content).toHaveClass('p-6')
    expect(content).toHaveClass('pt-0')
  })

  it('has the correct displayName', () => {
    expect(CardContent.displayName).toBe('CardContent')
  })
})

describe('CardFooter', () => {
  it('renders footer content', () => {
    render(<CardFooter>Footer here</CardFooter>)

    expect(screen.getByText('Footer here')).toBeInTheDocument()
  })

  it('applies default footer styling', () => {
    render(<CardFooter data-testid="card-footer">Footer</CardFooter>)

    const footer = screen.getByTestId('card-footer')
    expect(footer).toHaveClass('flex')
    expect(footer).toHaveClass('items-center')
    expect(footer).toHaveClass('p-6')
    expect(footer).toHaveClass('pt-0')
  })

  it('has the correct displayName', () => {
    expect(CardFooter.displayName).toBe('CardFooter')
  })
})
