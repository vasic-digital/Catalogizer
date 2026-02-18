import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import { TypeSelectorGrid, TYPE_ICONS, TYPE_COLORS } from '../TypeSelector'
import type { MediaTypeInfo } from '@/types/media'

// Mock framer-motion to simplify rendering
vi.mock('framer-motion', () => ({
  motion: {
    div: React.forwardRef(({ children, className, ...props }: any, ref: any) => (
      <div ref={ref} className={className} data-testid={props['data-testid']}>
        {children}
      </div>
    )),
  },
}))

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  Film: (props: any) => <svg data-testid="icon-film" {...props} />,
  Tv: (props: any) => <svg data-testid="icon-tv" {...props} />,
  Music: (props: any) => <svg data-testid="icon-music" {...props} />,
  Gamepad2: (props: any) => <svg data-testid="icon-gamepad" {...props} />,
  Monitor: (props: any) => <svg data-testid="icon-monitor" {...props} />,
  BookOpen: (props: any) => <svg data-testid="icon-bookopen" {...props} />,
  Book: (props: any) => <svg data-testid="icon-book" {...props} />,
}))

const mockTypes: MediaTypeInfo[] = [
  { id: 1, name: 'movie', description: 'Movies', count: 42 },
  { id: 2, name: 'tv_show', description: 'TV Shows', count: 15 },
  { id: 3, name: 'tv_season', description: 'TV Seasons', count: 30 },
  { id: 4, name: 'tv_episode', description: 'TV Episodes', count: 200 },
  { id: 5, name: 'music_artist', description: 'Music Artists', count: 8 },
  { id: 6, name: 'music_album', description: 'Music Albums', count: 25 },
  { id: 7, name: 'song', description: 'Songs', count: 300 },
  { id: 8, name: 'game', description: 'Games', count: 10 },
  { id: 9, name: 'software', description: 'Software', count: 5 },
  { id: 10, name: 'book', description: 'Books', count: 18 },
  { id: 11, name: 'comic', description: 'Comics', count: 7 },
]

describe('TypeSelectorGrid', () => {
  it('renders browsable type cards', () => {
    const handleSelect = vi.fn()
    render(<TypeSelectorGrid types={mockTypes} onSelect={handleSelect} />)

    expect(screen.getByText('movie')).toBeInTheDocument()
    expect(screen.getByText('tv show')).toBeInTheDocument()
    expect(screen.getByText('music artist')).toBeInTheDocument()
    expect(screen.getByText('music album')).toBeInTheDocument()
    expect(screen.getByText('game')).toBeInTheDocument()
    expect(screen.getByText('software')).toBeInTheDocument()
    expect(screen.getByText('book')).toBeInTheDocument()
    expect(screen.getByText('comic')).toBeInTheDocument()
  })

  it('filters out sub-types (tv_season, tv_episode, song)', () => {
    const handleSelect = vi.fn()
    render(<TypeSelectorGrid types={mockTypes} onSelect={handleSelect} />)

    // Sub-types should not be rendered
    expect(screen.queryByText('tv season')).not.toBeInTheDocument()
    expect(screen.queryByText('tv episode')).not.toBeInTheDocument()
    // 'song' would display as 'song' after replace
    expect(screen.queryByText('300 items')).not.toBeInTheDocument()
  })

  it('shows count for each type', () => {
    const handleSelect = vi.fn()
    render(<TypeSelectorGrid types={mockTypes} onSelect={handleSelect} />)

    expect(screen.getByText('42 items')).toBeInTheDocument()
    expect(screen.getByText('15 items')).toBeInTheDocument()
    expect(screen.getByText('8 items')).toBeInTheDocument()
    expect(screen.getByText('25 items')).toBeInTheDocument()
    expect(screen.getByText('10 items')).toBeInTheDocument()
    expect(screen.getByText('5 items')).toBeInTheDocument()
    expect(screen.getByText('18 items')).toBeInTheDocument()
    expect(screen.getByText('7 items')).toBeInTheDocument()
  })

  it('shows singular "item" for count of 1', () => {
    const singleType: MediaTypeInfo[] = [
      { id: 1, name: 'movie', description: 'Movies', count: 1 },
    ]
    const handleSelect = vi.fn()
    render(<TypeSelectorGrid types={singleType} onSelect={handleSelect} />)

    expect(screen.getByText('1 item')).toBeInTheDocument()
  })

  it('shows icon for each type', () => {
    const handleSelect = vi.fn()
    render(<TypeSelectorGrid types={mockTypes} onSelect={handleSelect} />)

    // Icons are rendered inside each card
    expect(screen.getAllByTestId('icon-film').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByTestId('icon-tv').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByTestId('icon-music').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByTestId('icon-gamepad').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByTestId('icon-monitor').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByTestId('icon-bookopen').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByTestId('icon-book').length).toBeGreaterThanOrEqual(1)
  })

  it('calls onSelect with the type name when clicked', async () => {
    const user = userEvent.setup()
    const handleSelect = vi.fn()
    render(<TypeSelectorGrid types={mockTypes} onSelect={handleSelect} />)

    await user.click(screen.getByText('movie'))
    expect(handleSelect).toHaveBeenCalledWith('movie')

    await user.click(screen.getByText('tv show'))
    expect(handleSelect).toHaveBeenCalledWith('tv_show')
  })

  it('renders correct number of cards (excluding sub-types)', () => {
    const handleSelect = vi.fn()
    render(<TypeSelectorGrid types={mockTypes} onSelect={handleSelect} />)

    // 11 types minus 3 sub-types (tv_season, tv_episode, song) = 8 buttons
    const buttons = screen.getAllByRole('button')
    expect(buttons).toHaveLength(8)
  })

  it('handles empty types array', () => {
    const handleSelect = vi.fn()
    const { container } = render(<TypeSelectorGrid types={[]} onSelect={handleSelect} />)

    const buttons = screen.queryAllByRole('button')
    expect(buttons).toHaveLength(0)
    // Grid should still render but be empty
    expect(container.querySelector('.grid')).toBeInTheDocument()
  })

  it('uses default icon for unknown type', () => {
    const unknownType: MediaTypeInfo[] = [
      { id: 99, name: 'unknown_type', description: 'Unknown', count: 3 },
    ]
    const handleSelect = vi.fn()
    render(<TypeSelectorGrid types={unknownType} onSelect={handleSelect} />)

    expect(screen.getByText('unknown type')).toBeInTheDocument()
    // Falls back to Film icon
    expect(screen.getByTestId('icon-film')).toBeInTheDocument()
  })

  it('exports TYPE_ICONS mapping', () => {
    expect(TYPE_ICONS).toBeDefined()
    expect(TYPE_ICONS.movie).toBeDefined()
    expect(TYPE_ICONS.tv_show).toBeDefined()
    expect(TYPE_ICONS.music_artist).toBeDefined()
    expect(TYPE_ICONS.game).toBeDefined()
    expect(TYPE_ICONS.software).toBeDefined()
    expect(TYPE_ICONS.book).toBeDefined()
    expect(TYPE_ICONS.comic).toBeDefined()
  })

  it('exports TYPE_COLORS mapping', () => {
    expect(TYPE_COLORS).toBeDefined()
    expect(TYPE_COLORS.movie).toContain('blue')
    expect(TYPE_COLORS.tv_show).toContain('purple')
    expect(TYPE_COLORS.game).toContain('red')
  })
})
