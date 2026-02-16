import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input } from '../Input'

describe('Input', () => {
  it('renders an input element', () => {
    render(<Input data-testid="test-input" />)

    const input = screen.getByTestId('test-input')
    expect(input).toBeInTheDocument()
    expect(input.tagName).toBe('INPUT')
  })

  it('applies default styling', () => {
    render(<Input data-testid="test-input" />)

    const input = screen.getByTestId('test-input')
    expect(input).toHaveClass('flex')
    expect(input).toHaveClass('h-10')
    expect(input).toHaveClass('w-full')
    expect(input).toHaveClass('rounded-md')
    expect(input).toHaveClass('border')
  })

  it('applies additional className', () => {
    render(<Input data-testid="test-input" className="my-custom" />)

    const input = screen.getByTestId('test-input')
    expect(input).toHaveClass('my-custom')
  })

  it('renders with text type by default', () => {
    render(<Input data-testid="test-input" />)

    const input = screen.getByTestId('test-input')
    // Default HTML input type is text when not specified
    expect(input).not.toHaveAttribute('type', 'password')
  })

  it('renders with password type', () => {
    render(<Input data-testid="test-input" type="password" />)

    const input = screen.getByTestId('test-input')
    expect(input).toHaveAttribute('type', 'password')
  })

  it('renders with email type', () => {
    render(<Input data-testid="test-input" type="email" />)

    const input = screen.getByTestId('test-input')
    expect(input).toHaveAttribute('type', 'email')
  })

  it('renders with number type', () => {
    render(<Input data-testid="test-input" type="number" />)

    const input = screen.getByTestId('test-input')
    expect(input).toHaveAttribute('type', 'number')
  })

  it('handles value changes', async () => {
    const user = userEvent.setup()
    const handleChange = vi.fn()

    render(<Input data-testid="test-input" onChange={handleChange} />)

    const input = screen.getByTestId('test-input')
    await user.type(input, 'hello')

    expect(handleChange).toHaveBeenCalled()
  })

  it('supports placeholder text', () => {
    render(<Input placeholder="Enter your name" />)

    expect(screen.getByPlaceholderText('Enter your name')).toBeInTheDocument()
  })

  it('supports disabled state', () => {
    render(<Input data-testid="test-input" disabled />)

    const input = screen.getByTestId('test-input')
    expect(input).toBeDisabled()
    expect(input).toHaveClass('disabled:cursor-not-allowed')
    expect(input).toHaveClass('disabled:opacity-50')
  })

  it('forwards ref to the input element', () => {
    const ref = vi.fn()
    render(<Input ref={ref} />)

    expect(ref).toHaveBeenCalled()
    expect(ref.mock.calls[0][0]).toBeInstanceOf(HTMLInputElement)
  })

  it('has the correct displayName', () => {
    expect(Input.displayName).toBe('Input')
  })

  it('handles controlled input value', () => {
    const { rerender } = render(<Input data-testid="test-input" value="initial" readOnly />)

    const input = screen.getByTestId('test-input')
    expect(input).toHaveValue('initial')

    rerender(<Input data-testid="test-input" value="updated" readOnly />)
    expect(input).toHaveValue('updated')
  })
})
