import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Textarea } from '../Textarea'

describe('Textarea', () => {
  it('renders a textarea element', () => {
    render(<Textarea />)
    const textarea = screen.getByRole('textbox')
    expect(textarea).toBeInTheDocument()
  })

  it('renders with label', () => {
    render(<Textarea label="Description" />)
    expect(screen.getByText('Description')).toBeInTheDocument()
  })

  it('renders without label when not provided', () => {
    render(<Textarea />)
    expect(screen.queryByText('Description')).not.toBeInTheDocument()
  })

  it('displays the value', () => {
    render(<Textarea value="Hello world" />)
    const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
    expect(textarea.value).toBe('Hello world')
  })

  it('calls onChange when text is entered', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<Textarea value="" onChange={onChange} />)

    await user.type(screen.getByRole('textbox'), 'test')
    expect(onChange).toHaveBeenCalled()
  })

  it('renders with placeholder', () => {
    render(<Textarea placeholder="Enter description..." />)
    expect(screen.getByPlaceholderText('Enter description...')).toBeInTheDocument()
  })

  it('sets rows attribute', () => {
    render(<Textarea rows={6} />)
    const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
    expect(textarea.rows).toBe(6)
  })

  it('uses default rows of 4', () => {
    render(<Textarea />)
    const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
    expect(textarea.rows).toBe(4)
  })

  it('applies disabled state', () => {
    render(<Textarea disabled />)
    const textarea = screen.getByRole('textbox')
    expect(textarea).toBeDisabled()
  })

  it('shows error message', () => {
    render(<Textarea error="This field is required" />)
    expect(screen.getByText('This field is required')).toBeInTheDocument()
  })

  it('applies error styling when error is present', () => {
    render(<Textarea error="Required" />)
    const textarea = screen.getByRole('textbox')
    expect(textarea).toHaveClass('border-red-500')
  })

  it('applies custom className', () => {
    const { container } = render(<Textarea className="custom-textarea" />)
    expect(container.firstChild).toHaveClass('custom-textarea')
  })
})
