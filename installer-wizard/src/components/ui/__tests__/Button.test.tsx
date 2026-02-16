import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from '../Button'

describe('Button', () => {
  it('renders with default variant and size', () => {
    render(<Button>Click me</Button>)

    const button = screen.getByRole('button', { name: 'Click me' })
    expect(button).toBeInTheDocument()
    expect(button).toHaveClass('bg-primary')
    expect(button).toHaveClass('h-10')
  })

  it('renders with destructive variant', () => {
    render(<Button variant="destructive">Delete</Button>)

    const button = screen.getByRole('button', { name: 'Delete' })
    expect(button).toHaveClass('bg-destructive')
  })

  it('renders with outline variant', () => {
    render(<Button variant="outline">Outlined</Button>)

    const button = screen.getByRole('button', { name: 'Outlined' })
    expect(button).toHaveClass('border')
    expect(button).toHaveClass('border-input')
  })

  it('renders with secondary variant', () => {
    render(<Button variant="secondary">Secondary</Button>)

    const button = screen.getByRole('button', { name: 'Secondary' })
    expect(button).toHaveClass('bg-secondary')
  })

  it('renders with ghost variant', () => {
    render(<Button variant="ghost">Ghost</Button>)

    const button = screen.getByRole('button', { name: 'Ghost' })
    expect(button).not.toHaveClass('bg-primary')
    expect(button).not.toHaveClass('border')
  })

  it('renders with link variant', () => {
    render(<Button variant="link">Link</Button>)

    const button = screen.getByRole('button', { name: 'Link' })
    expect(button).toHaveClass('text-primary')
    expect(button).toHaveClass('underline-offset-4')
  })

  it('renders with small size', () => {
    render(<Button size="sm">Small</Button>)

    const button = screen.getByRole('button', { name: 'Small' })
    expect(button).toHaveClass('h-9')
  })

  it('renders with large size', () => {
    render(<Button size="lg">Large</Button>)

    const button = screen.getByRole('button', { name: 'Large' })
    expect(button).toHaveClass('h-11')
  })

  it('renders with icon size', () => {
    render(<Button size="icon">X</Button>)

    const button = screen.getByRole('button', { name: 'X' })
    expect(button).toHaveClass('h-10')
    expect(button).toHaveClass('w-10')
  })

  it('handles onClick events', () => {
    const handleClick = vi.fn()
    render(<Button onClick={handleClick}>Click</Button>)

    fireEvent.click(screen.getByRole('button', { name: 'Click' }))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('supports disabled state', () => {
    render(<Button disabled>Disabled</Button>)

    const button = screen.getByRole('button', { name: 'Disabled' })
    expect(button).toBeDisabled()
    expect(button).toHaveClass('disabled:opacity-50')
  })

  it('applies additional className', () => {
    render(<Button className="custom-class">Custom</Button>)

    const button = screen.getByRole('button', { name: 'Custom' })
    expect(button).toHaveClass('custom-class')
  })

  it('forwards ref to the button element', () => {
    const ref = vi.fn()
    render(<Button ref={ref}>Ref</Button>)

    expect(ref).toHaveBeenCalled()
    expect(ref.mock.calls[0][0]).toBeInstanceOf(HTMLButtonElement)
  })

  it('has the correct displayName', () => {
    expect(Button.displayName).toBe('Button')
  })
})
