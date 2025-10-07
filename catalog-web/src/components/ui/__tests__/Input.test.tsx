import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Input } from '../Input';

describe('Input', () => {
  it('renders with default props', () => {
    render(<Input />);
    const input = screen.getByRole('textbox');
    expect(input).toBeInTheDocument();
    expect(input).toHaveClass('border-gray-300');
  });

  it('handles text input', async () => {
    const user = userEvent.setup();
    render(<Input />);
    const input = screen.getByRole('textbox');

    await user.type(input, 'Hello World');
    expect(input).toHaveValue('Hello World');
  });

  it('accepts placeholder text', () => {
    render(<Input placeholder="Enter your name" />);
    const input = screen.getByPlaceholderText('Enter your name');
    expect(input).toBeInTheDocument();
  });

  it('can be disabled', () => {
    render(<Input disabled />);
    const input = screen.getByRole('textbox');
    expect(input).toBeDisabled();
  });

  it('renders with different types', () => {
    const { rerender } = render(<Input type="email" />);
    expect(screen.getByRole('textbox')).toHaveAttribute('type', 'email');

    rerender(<Input type="password" />);
    expect(screen.getByDisplayValue('')).toHaveAttribute('type', 'password');

    rerender(<Input type="number" />);
    expect(screen.getByRole('spinbutton')).toHaveAttribute('type', 'number');
  });

  it('applies custom className', () => {
    render(<Input className="custom-input" />);
    expect(screen.getByRole('textbox')).toHaveClass('custom-input');
  });

  it('forwards other props', () => {
    render(<Input data-testid="custom-input" maxLength={10} />);
    const input = screen.getByTestId('custom-input');
    expect(input).toHaveAttribute('maxLength', '10');
  });

  it('handles controlled input', () => {
    const handleChange = jest.fn();
    render(<Input value="controlled" onChange={handleChange} />);
    const input = screen.getByRole('textbox');
    expect(input).toHaveValue('controlled');
  });

  it('supports different sizes', () => {
    render(<Input className="text-sm" />);
    expect(screen.getByRole('textbox')).toHaveClass('text-sm');
  });
});