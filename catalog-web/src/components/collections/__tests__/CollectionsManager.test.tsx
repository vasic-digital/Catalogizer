import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CollectionsManager } from '../CollectionsManager'

vi.mock('lucide-react', () => ({
  Plus: () => <span data-testid="icon-plus">+</span>,
  Search: () => <span data-testid="icon-search">Search</span>,
  Grid: () => <span data-testid="icon-grid">Grid</span>,
  List: () => <span data-testid="icon-list">List</span>,
  PlayCircle: () => <span data-testid="icon-play">Play</span>,
  Star: () => <span data-testid="icon-star">Star</span>,
  Clock: () => <span data-testid="icon-clock">Clock</span>,
  MoreHorizontal: () => <span data-testid="icon-more">More</span>,
  Edit2: () => <span data-testid="icon-edit">Edit</span>,
  Trash2: () => <span data-testid="icon-trash">Trash</span>,
}))

const mockCollections = [
  {
    id: '1',
    name: 'Movie Favorites',
    description: 'My favorite movies of all time',
    mediaCount: 42,
    duration: 7200,
    thumbnail: '/img/movies.jpg',
    isSmart: false,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-15T00:00:00Z',
  },
  {
    id: '2',
    name: 'Jazz Collection',
    description: 'Best jazz music tracks',
    mediaCount: 128,
    duration: 25000,
    isSmart: true,
    createdAt: '2024-01-05T00:00:00Z',
    updatedAt: '2024-01-20T00:00:00Z',
  },
  {
    id: '3',
    name: 'Workout Playlist',
    description: 'High energy tracks for working out',
    mediaCount: 35,
    duration: 3600,
    isSmart: false,
    createdAt: '2024-01-10T00:00:00Z',
    updatedAt: '2024-01-18T00:00:00Z',
  },
]

describe('CollectionsManager', () => {
  const defaultProps = {
    collections: mockCollections,
    onCreateCollection: vi.fn(),
    onUpdateCollection: vi.fn(),
    onDeleteCollection: vi.fn(),
    onPlayCollection: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders all collections', () => {
    render(<CollectionsManager {...defaultProps} />)
    expect(screen.getByText('Movie Favorites')).toBeInTheDocument()
    expect(screen.getByText('Jazz Collection')).toBeInTheDocument()
    expect(screen.getByText('Workout Playlist')).toBeInTheDocument()
  })

  it('renders collection descriptions', () => {
    render(<CollectionsManager {...defaultProps} />)
    expect(screen.getByText('My favorite movies of all time')).toBeInTheDocument()
    expect(screen.getByText('Best jazz music tracks')).toBeInTheDocument()
  })

  it('renders media counts', () => {
    render(<CollectionsManager {...defaultProps} />)
    expect(screen.getByText('42 items')).toBeInTheDocument()
    expect(screen.getByText('128 items')).toBeInTheDocument()
    expect(screen.getByText('35 items')).toBeInTheDocument()
  })

  it('formats durations correctly', () => {
    render(<CollectionsManager {...defaultProps} />)
    expect(screen.getByText('2h 0m')).toBeInTheDocument() // 7200s
    expect(screen.getByText('1h 0m')).toBeInTheDocument() // 3600s
  })

  it('shows Smart badge for smart collections', () => {
    render(<CollectionsManager {...defaultProps} />)
    expect(screen.getByText('Smart')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<CollectionsManager {...defaultProps} />)
    expect(screen.getByPlaceholderText('Search collections...')).toBeInTheDocument()
  })

  it('renders New Collection button', () => {
    render(<CollectionsManager {...defaultProps} />)
    expect(screen.getByText('New Collection')).toBeInTheDocument()
  })

  it('renders grid and list view toggle buttons', () => {
    render(<CollectionsManager {...defaultProps} />)
    expect(screen.getByTestId('icon-grid')).toBeInTheDocument()
    expect(screen.getByTestId('icon-list')).toBeInTheDocument()
  })

  it('filters collections by search query', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    const searchInput = screen.getByPlaceholderText('Search collections...')
    await user.type(searchInput, 'Jazz')

    expect(screen.getByText('Jazz Collection')).toBeInTheDocument()
    expect(screen.queryByText('Movie Favorites')).not.toBeInTheDocument()
    expect(screen.queryByText('Workout Playlist')).not.toBeInTheDocument()
  })

  it('filters by description too', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    const searchInput = screen.getByPlaceholderText('Search collections...')
    await user.type(searchInput, 'energy')

    expect(screen.getByText('Workout Playlist')).toBeInTheDocument()
    expect(screen.queryByText('Movie Favorites')).not.toBeInTheDocument()
  })

  it('shows empty state when no collections match search', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    const searchInput = screen.getByPlaceholderText('Search collections...')
    await user.type(searchInput, 'xyznonexistent')

    expect(screen.getByText('No collections found')).toBeInTheDocument()
    expect(screen.getByText('Try adjusting your search terms')).toBeInTheDocument()
  })

  it('shows empty state with create button when no collections', () => {
    render(<CollectionsManager collections={[]} />)
    expect(screen.getByText('No collections yet')).toBeInTheDocument()
    expect(screen.getByText('Create Collection')).toBeInTheDocument()
  })

  it('opens create modal when New Collection is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    await user.click(screen.getByText('New Collection'))

    expect(screen.getByText('Create New Collection')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Collection name')).toBeInTheDocument()
  })

  it('closes create modal when Cancel is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    await user.click(screen.getByText('New Collection'))
    expect(screen.getByText('Create New Collection')).toBeInTheDocument()

    await user.click(screen.getByText('Cancel'))
    expect(screen.queryByText('Create New Collection')).not.toBeInTheDocument()
  })

  it('calls onCreateCollection when Create button is clicked in modal', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    await user.click(screen.getByText('New Collection'))
    await user.click(screen.getByText('Create'))

    expect(defaultProps.onCreateCollection).toHaveBeenCalled()
  })

  it('calls onDeleteCollection when delete button is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    // Find all trash icons and click the first one
    const trashIcons = screen.getAllByTestId('icon-trash')
    const deleteBtn = trashIcons[0].closest('button')
    if (deleteBtn) {
      await user.click(deleteBtn)
    }

    expect(defaultProps.onDeleteCollection).toHaveBeenCalledWith('1')
  })

  it('calls onPlayCollection when play button is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    const playIcons = screen.getAllByTestId('icon-play')
    // Find the standalone play button (not in card header)
    const playBtn = playIcons[playIcons.length - 1].closest('button')
    if (playBtn) {
      await user.click(playBtn)
    }

    expect(defaultProps.onPlayCollection).toHaveBeenCalled()
  })

  it('switches to list view mode', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    const listIcon = screen.getByTestId('icon-list')
    const listBtn = listIcon.closest('button')
    if (listBtn) {
      await user.click(listBtn)
    }

    // In list view, Smart Collection badge text is different
    expect(screen.getByText('Smart Collection')).toBeInTheDocument()
  })

  it('renders collection thumbnails in grid view', () => {
    render(<CollectionsManager {...defaultProps} />)
    const images = document.querySelectorAll('img')
    expect(images.length).toBeGreaterThan(0)
    expect(images[0]).toHaveAttribute('alt', 'Movie Favorites')
  })

  it('renders placeholder for collections without thumbnails', () => {
    render(<CollectionsManager {...defaultProps} />)
    // Collections without thumbnails get a gradient background
    const gradients = document.querySelectorAll('.bg-gradient-to-br')
    expect(gradients.length).toBeGreaterThan(0)
  })

  it('shows edit modal when edit button is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    const editIcons = screen.getAllByTestId('icon-edit')
    const editBtn = editIcons[0].closest('button')
    if (editBtn) {
      await user.click(editBtn)
    }

    expect(screen.getByText('Edit Collection')).toBeInTheDocument()
    expect(screen.getByText('Update')).toBeInTheDocument()
  })

  it('calls onUpdateCollection when Update is clicked', async () => {
    const user = userEvent.setup()
    render(<CollectionsManager {...defaultProps} />)

    // Open edit modal
    const editIcons = screen.getAllByTestId('icon-edit')
    const editBtn = editIcons[0].closest('button')
    if (editBtn) {
      await user.click(editBtn)
    }

    // Click update
    await user.click(screen.getByText('Update'))

    expect(defaultProps.onUpdateCollection).toHaveBeenCalledWith('1', {
      name: 'Updated Collection Name',
      description: 'Updated description',
    })
  })

  it('displays updated dates', () => {
    render(<CollectionsManager {...defaultProps} />)
    // Check for "Updated" text patterns
    const updatedTexts = screen.getAllByText(/Updated/)
    expect(updatedTexts.length).toBeGreaterThan(0)
  })
})
