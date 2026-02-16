import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Switch } from '../Switch'

describe('Switch', () => {
  it('renders as a switch role element', () => {
    render(<Switch />)
    const switchEl = screen.getByRole('switch')
    expect(switchEl).toBeInTheDocument()
  })

  it('renders in unchecked state by default', () => {
    render(<Switch />)
    const switchEl = screen.getByRole('switch')
    expect(switchEl).toHaveAttribute('aria-checked', 'false')
  })

  it('renders in checked state when checked is true', () => {
    render(<Switch checked />)
    const switchEl = screen.getByRole('switch')
    expect(switchEl).toHaveAttribute('aria-checked', 'true')
  })

  it('calls onCheckedChange when clicked', async () => {
    const user = userEvent.setup()
    const onCheckedChange = vi.fn()
    render(<Switch checked={false} onCheckedChange={onCheckedChange} />)

    await user.click(screen.getByRole('switch'))
    expect(onCheckedChange).toHaveBeenCalledWith(true)
  })

  it('calls onCheckedChange with false when unchecking', async () => {
    const user = userEvent.setup()
    const onCheckedChange = vi.fn()
    render(<Switch checked={true} onCheckedChange={onCheckedChange} />)

    await user.click(screen.getByRole('switch'))
    expect(onCheckedChange).toHaveBeenCalledWith(false)
  })

  it('does not call onCheckedChange when disabled', async () => {
    const user = userEvent.setup()
    const onCheckedChange = vi.fn()
    render(<Switch checked={false} onCheckedChange={onCheckedChange} disabled />)

    await user.click(screen.getByRole('switch'))
    expect(onCheckedChange).not.toHaveBeenCalled()
  })

  it('applies disabled styling when disabled', () => {
    render(<Switch disabled />)
    const switchEl = screen.getByRole('switch')
    expect(switchEl).toBeDisabled()
    expect(switchEl).toHaveClass('opacity-50', 'cursor-not-allowed')
  })

  it('applies checked styling when checked', () => {
    render(<Switch checked />)
    const switchEl = screen.getByRole('switch')
    expect(switchEl).toHaveClass('bg-blue-600')
  })

  it('applies unchecked styling when not checked', () => {
    render(<Switch checked={false} />)
    const switchEl = screen.getByRole('switch')
    expect(switchEl).toHaveClass('bg-gray-200')
  })

  it('applies custom className', () => {
    render(<Switch className="custom-switch" />)
    const switchEl = screen.getByRole('switch')
    expect(switchEl).toHaveClass('custom-switch')
  })

  it('is a button element', () => {
    render(<Switch />)
    const switchEl = screen.getByRole('switch')
    expect(switchEl.tagName).toBe('BUTTON')
  })
})
