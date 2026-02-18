import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import { EntityGrid } from '../EntityGrid'
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

// Mock EntityCard to simplify testing EntityGrid
vi.mock('../EntityCard', () => ({
  EntityCard: ({ entity, onClick }: any) => (
    <div data-testid={`entity-card-${entity.id}`}>
      <span>{entity.title}</span>
      <button onClick={onClick}>Select {entity.title}</button>
    </div>
  ),
}))

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  ChevronLeft: (props: any) => <svg data-testid="icon-chevron-left" {...props} />,
  ChevronRight: (props: any) => <svg data-testid="icon-chevron-right" {...props} />,
  Film: (props: any) => <svg data-testid="icon-film" {...props} />,
}))

const createMockEntities = (count: number): MediaEntity[] =>
  Array.from({ length: count }, (_, i) => ({
    id: i + 1,
    media_type_id: 1,
    title: `Entity ${i + 1}`,
    status: 'movie',
    first_detected: '2024-01-01T00:00:00Z',
    last_updated: '2024-01-01T00:00:00Z',
  }))

describe('EntityGrid', () => {
  it('renders entity cards', () => {
    const entities = createMockEntities(3)
    render(
      <EntityGrid
        entities={entities}
        total={3}
        limit={20}
        offset={0}
        page={1}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    expect(screen.getByText('Entity 1')).toBeInTheDocument()
    expect(screen.getByText('Entity 2')).toBeInTheDocument()
    expect(screen.getByText('Entity 3')).toBeInTheDocument()
  })

  it('shows "Showing X-Y of Z" text', () => {
    const entities = createMockEntities(5)
    render(
      <EntityGrid
        entities={entities}
        total={50}
        limit={5}
        offset={0}
        page={1}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    expect(screen.getByText('Showing 1-5 of 50')).toBeInTheDocument()
  })

  it('shows correct range on later pages', () => {
    const entities = createMockEntities(5)
    render(
      <EntityGrid
        entities={entities}
        total={50}
        limit={5}
        offset={10}
        page={3}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    expect(screen.getByText('Showing 11-15 of 50')).toBeInTheDocument()
  })

  it('clamps the upper range to total on last page', () => {
    const entities = createMockEntities(3)
    render(
      <EntityGrid
        entities={entities}
        total={13}
        limit={5}
        offset={10}
        page={3}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    expect(screen.getByText('Showing 11-13 of 13')).toBeInTheDocument()
  })

  it('shows "No entities found" when empty', () => {
    render(
      <EntityGrid
        entities={[]}
        total={0}
        limit={20}
        offset={0}
        page={1}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    expect(screen.getByText('No entities found')).toBeInTheDocument()
  })

  it('does not show pagination when only one page', () => {
    const entities = createMockEntities(3)
    render(
      <EntityGrid
        entities={entities}
        total={3}
        limit={20}
        offset={0}
        page={1}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    expect(screen.queryByText(/Page \d+ of \d+/)).not.toBeInTheDocument()
  })

  it('shows pagination when multiple pages', () => {
    const entities = createMockEntities(5)
    render(
      <EntityGrid
        entities={entities}
        total={50}
        limit={5}
        offset={0}
        page={1}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    expect(screen.getByText('Page 1 of 10')).toBeInTheDocument()
  })

  it('disables previous button on first page', () => {
    const entities = createMockEntities(5)
    render(
      <EntityGrid
        entities={entities}
        total={50}
        limit={5}
        offset={0}
        page={1}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    const buttons = screen.getAllByRole('button').filter(
      (btn) => btn.querySelector('[data-testid="icon-chevron-left"]')
    )
    expect(buttons[0]).toBeDisabled()
  })

  it('disables next button on last page', () => {
    const entities = createMockEntities(5)
    render(
      <EntityGrid
        entities={entities}
        total={50}
        limit={5}
        offset={45}
        page={10}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    const buttons = screen.getAllByRole('button').filter(
      (btn) => btn.querySelector('[data-testid="icon-chevron-right"]')
    )
    expect(buttons[0]).toBeDisabled()
  })

  it('calls onPageChange with previous page when prev button clicked', async () => {
    const user = userEvent.setup()
    const handlePageChange = vi.fn()
    const entities = createMockEntities(5)
    render(
      <EntityGrid
        entities={entities}
        total={50}
        limit={5}
        offset={5}
        page={2}
        onEntityClick={vi.fn()}
        onPageChange={handlePageChange}
      />
    )

    const prevButton = screen.getAllByRole('button').find(
      (btn) => btn.querySelector('[data-testid="icon-chevron-left"]')
    )!
    await user.click(prevButton)
    expect(handlePageChange).toHaveBeenCalledWith(1)
  })

  it('calls onPageChange with next page when next button clicked', async () => {
    const user = userEvent.setup()
    const handlePageChange = vi.fn()
    const entities = createMockEntities(5)
    render(
      <EntityGrid
        entities={entities}
        total={50}
        limit={5}
        offset={0}
        page={1}
        onEntityClick={vi.fn()}
        onPageChange={handlePageChange}
      />
    )

    const nextButton = screen.getAllByRole('button').find(
      (btn) => btn.querySelector('[data-testid="icon-chevron-right"]')
    )!
    await user.click(nextButton)
    expect(handlePageChange).toHaveBeenCalledWith(2)
  })

  it('calls onEntityClick when an entity card is clicked', async () => {
    const user = userEvent.setup()
    const handleEntityClick = vi.fn()
    const entities = createMockEntities(3)
    render(
      <EntityGrid
        entities={entities}
        total={3}
        limit={20}
        offset={0}
        page={1}
        onEntityClick={handleEntityClick}
        onPageChange={vi.fn()}
      />
    )

    await user.click(screen.getByText('Select Entity 2'))
    expect(handleEntityClick).toHaveBeenCalledTimes(1)
    expect(handleEntityClick).toHaveBeenCalledWith(entities[1])
  })

  it('renders the correct number of entity cards', () => {
    const entities = createMockEntities(4)
    render(
      <EntityGrid
        entities={entities}
        total={4}
        limit={20}
        offset={0}
        page={1}
        onEntityClick={vi.fn()}
        onPageChange={vi.fn()}
      />
    )

    const cards = screen.getAllByTestId(/entity-card-/)
    expect(cards).toHaveLength(4)
  })
})
