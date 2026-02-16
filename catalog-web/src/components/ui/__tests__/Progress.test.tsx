import { render, screen } from '@testing-library/react'
import { Progress } from '../Progress'

describe('Progress', () => {
  it('renders progress bar', () => {
    const { container } = render(<Progress value={50} />)
    const progressBar = container.querySelector('.bg-blue-500')
    expect(progressBar).toBeInTheDocument()
  })

  it('sets correct width based on value', () => {
    const { container } = render(<Progress value={75} />)
    const progressBar = container.querySelector('.bg-blue-500')
    expect(progressBar).toHaveStyle({ width: '75%' })
  })

  it('clamps value at 100%', () => {
    const { container } = render(<Progress value={150} />)
    const progressBar = container.querySelector('.bg-blue-500')
    expect(progressBar).toHaveStyle({ width: '100%' })
  })

  it('clamps value at 0%', () => {
    const { container } = render(<Progress value={-10} />)
    const progressBar = container.querySelector('.bg-blue-500')
    expect(progressBar).toHaveStyle({ width: '0%' })
  })

  it('uses custom max value', () => {
    const { container } = render(<Progress value={50} max={200} />)
    const progressBar = container.querySelector('.bg-blue-500')
    expect(progressBar).toHaveStyle({ width: '25%' })
  })

  it('shows label when showLabel is true', () => {
    render(<Progress value={65} showLabel />)
    expect(screen.getByText('Progress')).toBeInTheDocument()
    expect(screen.getByText('65%')).toBeInTheDocument()
  })

  it('hides label by default', () => {
    render(<Progress value={65} />)
    expect(screen.queryByText('Progress')).not.toBeInTheDocument()
    expect(screen.queryByText('65%')).not.toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<Progress value={50} className="custom-class" />)
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('renders 0% progress correctly', () => {
    const { container } = render(<Progress value={0} />)
    const progressBar = container.querySelector('.bg-blue-500')
    expect(progressBar).toHaveStyle({ width: '0%' })
  })

  it('renders 100% progress correctly', () => {
    const { container } = render(<Progress value={100} />)
    const progressBar = container.querySelector('.bg-blue-500')
    expect(progressBar).toHaveStyle({ width: '100%' })
  })
})
