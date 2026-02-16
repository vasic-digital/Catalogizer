import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Select } from '../Select'

describe('Select', () => {
  const options = [
    { value: 'a', label: 'Option A' },
    { value: 'b', label: 'Option B' },
    { value: 'c', label: 'Option C' },
  ]

  it('renders select element', () => {
    render(<Select options={options} />)
    const select = screen.getByRole('combobox')
    expect(select).toBeInTheDocument()
  })

  it('renders all options', () => {
    render(<Select options={options} />)
    expect(screen.getByText('Option A')).toBeInTheDocument()
    expect(screen.getByText('Option B')).toBeInTheDocument()
    expect(screen.getByText('Option C')).toBeInTheDocument()
  })

  it('sets the selected value', () => {
    render(<Select options={options} value="b" />)
    const select = screen.getByRole('combobox') as HTMLSelectElement
    expect(select.value).toBe('b')
  })

  it('calls onChange when selection changes', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<Select options={options} value="a" onChange={onChange} />)

    await user.selectOptions(screen.getByRole('combobox'), 'b')
    expect(onChange).toHaveBeenCalledWith('b')
  })

  it('calls onValueChange when selection changes', async () => {
    const user = userEvent.setup()
    const onValueChange = vi.fn()
    render(<Select options={options} value="a" onValueChange={onValueChange} />)

    await user.selectOptions(screen.getByRole('combobox'), 'c')
    expect(onValueChange).toHaveBeenCalledWith('c')
  })

  it('renders children instead of options when provided', () => {
    render(
      <Select>
        <option value="x">Custom X</option>
        <option value="y">Custom Y</option>
      </Select>
    )
    expect(screen.getByText('Custom X')).toBeInTheDocument()
    expect(screen.getByText('Custom Y')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(<Select options={options} className="custom-select" />)
    const select = screen.getByRole('combobox')
    expect(select).toHaveClass('custom-select')
  })

  it('renders chevron icon container', () => {
    const { container } = render(<Select options={options} />)
    const chevronContainer = container.querySelector('.pointer-events-none')
    expect(chevronContainer).toBeInTheDocument()
  })
})
