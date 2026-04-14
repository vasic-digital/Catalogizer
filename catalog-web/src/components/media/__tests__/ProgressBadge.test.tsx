import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import { ProgressBadge } from '../ProgressBadge'
import type { UiPlaybackProgress } from '@/types/playback'

const sampleProgress: UiPlaybackProgress = {
  mediaItemId: 42,
  positionUnit: 'seconds',
  durationTotal: 7200,
  lastPosition: 3600,
  lastSessionAmount: 1800,
  totalReproductions: 5,
  aggregateAmount: 9000,
  lastSessionEndedAt: '2026-04-10T15:30:00Z',
}

describe('ProgressBadge', () => {
  it('renders nothing when progress is null', () => {
    const { container } = render(<ProgressBadge progress={null} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders nothing when progress is undefined', () => {
    const { container } = render(<ProgressBadge progress={undefined} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders the badge when progress is provided', () => {
    render(<ProgressBadge progress={sampleProgress} />)
    const badge = screen.getByTestId('progress-badge-42')
    expect(badge).toBeInTheDocument()
  })

  it('displays total duration', () => {
    render(<ProgressBadge progress={sampleProgress} />)
    // 7200 seconds = 2h 0m
    expect(screen.getByText('2h 0m')).toBeInTheDocument()
  })

  it('displays current position / total as progress label', () => {
    render(<ProgressBadge progress={sampleProgress} />)
    // 3600s / 7200s = "1h 0m / 2h 0m"
    expect(screen.getByText('1h 0m / 2h 0m')).toBeInTheDocument()
  })

  it('displays last session amount with "last" suffix', () => {
    render(<ProgressBadge progress={sampleProgress} />)
    // 1800s = 30m
    expect(screen.getByText('30m last')).toBeInTheDocument()
  })

  it('displays reproduction count with "played" suffix', () => {
    render(<ProgressBadge progress={sampleProgress} />)
    expect(screen.getByText(/5× played/)).toBeInTheDocument()
  })

  it('renders a progress bar when fraction is greater than zero', () => {
    render(<ProgressBadge progress={sampleProgress} />)
    const badge = screen.getByTestId('progress-badge-42')
    // 3600/7200 = 50%
    const progressBar = badge.querySelector('[style*="width"]')
    expect(progressBar).not.toBeNull()
    expect(progressBar?.getAttribute('style')).toContain('50%')
  })

  it('does not render a progress bar when position is zero', () => {
    const zeroProgress: UiPlaybackProgress = {
      ...sampleProgress,
      lastPosition: 0,
    }
    render(<ProgressBadge progress={zeroProgress} />)
    const badge = screen.getByTestId('progress-badge-42')
    const progressBar = badge.querySelector('[style*="width"]')
    expect(progressBar).toBeNull()
  })

  it('has correct aria-label reflecting reproduction count', () => {
    render(<ProgressBadge progress={sampleProgress} />)
    const badge = screen.getByTestId('progress-badge-42')
    expect(badge).toHaveAttribute(
      'aria-label',
      'View reproduction history (5× played)',
    )
  })

  it('calls onOpenHistory when clicked', async () => {
    const onOpenHistory = vi.fn()
    render(
      <ProgressBadge
        progress={sampleProgress}
        onOpenHistory={onOpenHistory}
      />,
    )
    const user = userEvent.setup()
    await user.click(screen.getByTestId('progress-badge-42'))
    expect(onOpenHistory).toHaveBeenCalledTimes(1)
  })

  it('stops event propagation on click', async () => {
    const onOpenHistory = vi.fn()
    const outerClick = vi.fn()
    render(
      <div onClick={outerClick}>
        <ProgressBadge
          progress={sampleProgress}
          onOpenHistory={onOpenHistory}
        />
      </div>,
    )
    const user = userEvent.setup()
    await user.click(screen.getByTestId('progress-badge-42'))
    expect(onOpenHistory).toHaveBeenCalledTimes(1)
    expect(outerClick).not.toHaveBeenCalled()
  })

  it('does not throw when onOpenHistory is not provided', async () => {
    render(<ProgressBadge progress={sampleProgress} />)
    const user = userEvent.setup()
    await expect(
      user.click(screen.getByTestId('progress-badge-42')),
    ).resolves.not.toThrow()
  })

  it('applies custom className', () => {
    render(
      <ProgressBadge progress={sampleProgress} className="custom-class" />,
    )
    const badge = screen.getByTestId('progress-badge-42')
    expect(badge.className).toContain('custom-class')
  })

  it('renders as a button element', () => {
    render(<ProgressBadge progress={sampleProgress} />)
    const badge = screen.getByTestId('progress-badge-42')
    expect(badge.tagName).toBe('BUTTON')
    expect(badge).toHaveAttribute('type', 'button')
  })

  it('handles page-based position units', () => {
    const pageProgress: UiPlaybackProgress = {
      mediaItemId: 99,
      positionUnit: 'pages',
      durationTotal: 320,
      lastPosition: 140,
      lastSessionAmount: 30,
      totalReproductions: 2,
      aggregateAmount: 280,
      lastSessionEndedAt: null,
    }
    render(<ProgressBadge progress={pageProgress} />)
    expect(screen.getByText('320 pages')).toBeInTheDocument()
    expect(screen.getByText('140 pages / 320 pages')).toBeInTheDocument()
    expect(screen.getByText('30 pages last')).toBeInTheDocument()
    expect(screen.getByText(/2× played/)).toBeInTheDocument()
  })

  it('handles null durationTotal gracefully', () => {
    const nullDuration: UiPlaybackProgress = {
      ...sampleProgress,
      durationTotal: null,
      lastPosition: 600,
    }
    render(<ProgressBadge progress={nullDuration} />)
    const badge = screen.getByTestId('progress-badge-42')
    expect(badge).toBeInTheDocument()
    // duration shows as "-" for 0, position shows just current
    expect(screen.getByText('-')).toBeInTheDocument()
    expect(screen.getByText('10m')).toBeInTheDocument()
  })

  it('clamps progress bar at 100% when position exceeds total', () => {
    const overProgress: UiPlaybackProgress = {
      ...sampleProgress,
      lastPosition: 10000,
      durationTotal: 7200,
    }
    render(<ProgressBadge progress={overProgress} />)
    const badge = screen.getByTestId('progress-badge-42')
    const progressBar = badge.querySelector('[style*="width"]')
    expect(progressBar).not.toBeNull()
    expect(progressBar?.getAttribute('style')).toContain('100%')
  })
})
