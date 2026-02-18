import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import { EntityCard } from '../EntityCard'
import type { MediaEntity } from '@/types/media'

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

const mockEntity: MediaEntity = {
  id: 1,
  media_type_id: 1,
  title: 'The Dark Knight',
  year: 2008,
  description: 'A superhero film',
  genre: ['Action', 'Crime', 'Drama', 'Thriller'],
  director: 'Christopher Nolan',
  rating: 9.0,
  runtime: 152,
  language: 'English',
  status: 'movie',
  first_detected: '2024-01-01T00:00:00Z',
  last_updated: '2024-01-01T00:00:00Z',
}

describe('EntityCard', () => {
  it('renders the entity title', () => {
    render(<EntityCard entity={mockEntity} onClick={vi.fn()} />)
    expect(screen.getByText('The Dark Knight')).toBeInTheDocument()
  })

  it('shows year when present', () => {
    render(<EntityCard entity={mockEntity} onClick={vi.fn()} />)
    expect(screen.getByText('2008')).toBeInTheDocument()
  })

  it('does not show year when absent', () => {
    const noYearEntity = { ...mockEntity, year: undefined }
    render(<EntityCard entity={noYearEntity} onClick={vi.fn()} />)
    expect(screen.queryByText('2008')).not.toBeInTheDocument()
  })

  it('shows rating formatted to one decimal', () => {
    render(<EntityCard entity={mockEntity} onClick={vi.fn()} />)
    expect(screen.getByText('9.0')).toBeInTheDocument()
  })

  it('does not show rating when null', () => {
    const noRatingEntity = { ...mockEntity, rating: undefined }
    render(<EntityCard entity={noRatingEntity} onClick={vi.fn()} />)
    expect(screen.queryByText('9.0')).not.toBeInTheDocument()
  })

  it('shows genre badges (max 3)', () => {
    render(<EntityCard entity={mockEntity} onClick={vi.fn()} />)
    expect(screen.getByText('Action')).toBeInTheDocument()
    expect(screen.getByText('Crime')).toBeInTheDocument()
    expect(screen.getByText('Drama')).toBeInTheDocument()
    // 4th genre should not be shown
    expect(screen.queryByText('Thriller')).not.toBeInTheDocument()
  })

  it('does not render genre section when genres are empty', () => {
    const noGenreEntity = { ...mockEntity, genre: [] }
    render(<EntityCard entity={noGenreEntity} onClick={vi.fn()} />)
    expect(screen.queryByText('Action')).not.toBeInTheDocument()
  })

  it('does not render genre section when genres are undefined', () => {
    const noGenreEntity = { ...mockEntity, genre: undefined }
    render(<EntityCard entity={noGenreEntity} onClick={vi.fn()} />)
    expect(screen.queryByText('Action')).not.toBeInTheDocument()
  })

  it('calls onClick when card is clicked', async () => {
    const user = userEvent.setup()
    const handleClick = vi.fn()
    render(<EntityCard entity={mockEntity} onClick={handleClick} />)

    // The Card component with onClick wraps the content
    await user.click(screen.getByText('The Dark Knight'))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('renders with fewer than 3 genres', () => {
    const fewGenresEntity = { ...mockEntity, genre: ['Sci-Fi'] }
    render(<EntityCard entity={fewGenresEntity} onClick={vi.fn()} />)
    expect(screen.getByText('Sci-Fi')).toBeInTheDocument()
  })

  it('handles entity with minimal fields', () => {
    const minimalEntity: MediaEntity = {
      id: 2,
      media_type_id: 1,
      title: 'Minimal Entity',
      status: 'movie',
      first_detected: '2024-01-01T00:00:00Z',
      last_updated: '2024-01-01T00:00:00Z',
    }
    render(<EntityCard entity={minimalEntity} onClick={vi.fn()} />)
    expect(screen.getByText('Minimal Entity')).toBeInTheDocument()
  })

  it('renders the icon based on entity status', () => {
    render(<EntityCard entity={mockEntity} onClick={vi.fn()} />)
    // Entity status is 'movie', so Film icon is used
    expect(screen.getByTestId('icon-film')).toBeInTheDocument()
  })

  it('falls back to Film icon for unknown status', () => {
    const unknownEntity = { ...mockEntity, status: 'unknown_type' }
    render(<EntityCard entity={unknownEntity} onClick={vi.fn()} />)
    expect(screen.getByTestId('icon-film')).toBeInTheDocument()
  })
})
