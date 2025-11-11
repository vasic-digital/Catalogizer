import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MediaGrid } from '../MediaGrid';
import type { MediaItem } from '@/types/media';

// Mock MediaCard component
jest.mock('../MediaCard', () => ({
  MediaCard: ({ media, onView, onDownload }: any) => (
    <div data-testid={`media-card-${media.id}`}>
      <h3>{media.title}</h3>
      {onView && (
        <button onClick={() => onView(media)}>View {media.title}</button>
      )}
      {onDownload && (
        <button onClick={() => onDownload(media)}>Download {media.title}</button>
      )}
    </div>
  ),
}));

const mockMediaItems: MediaItem[] = [
  {
    id: 1,
    title: 'Test Movie 1',
    media_type: 'movie',
    directory_path: '/movies/test1',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    title: 'Test Movie 2',
    media_type: 'movie',
    directory_path: '/movies/test2',
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  },
  {
    id: 3,
    title: 'Test Movie 3',
    media_type: 'movie',
    directory_path: '/movies/test3',
    created_at: '2024-01-03T00:00:00Z',
    updated_at: '2024-01-03T00:00:00Z',
  },
];

describe('MediaGrid', () => {
  it('renders media items correctly', () => {
    render(<MediaGrid media={mockMediaItems} />);

    expect(screen.getByText('Test Movie 1')).toBeInTheDocument();
    expect(screen.getByText('Test Movie 2')).toBeInTheDocument();
    expect(screen.getByText('Test Movie 3')).toBeInTheDocument();
  });

  it('renders correct number of media cards', () => {
    render(<MediaGrid media={mockMediaItems} />);

    const cards = screen.getAllByTestId(/media-card-/);
    expect(cards).toHaveLength(3);
  });

  it('displays loading skeletons when loading', () => {
    render(<MediaGrid media={[]} loading={true} />);

    // Should render 12 loading skeletons
    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons).toHaveLength(12);
  });

  it('does not display media items when loading', () => {
    render(<MediaGrid media={mockMediaItems} loading={true} />);

    expect(screen.queryByText('Test Movie 1')).not.toBeInTheDocument();
  });

  it('displays empty state when no media', () => {
    render(<MediaGrid media={[]} />);

    expect(screen.getByText('No media found')).toBeInTheDocument();
    expect(
      screen.getByText(/Try adjusting your search criteria/i)
    ).toBeInTheDocument();
  });

  it('does not display empty state when loading', () => {
    render(<MediaGrid media={[]} loading={true} />);

    expect(screen.queryByText('No media found')).not.toBeInTheDocument();
  });

  it('passes onMediaView callback to MediaCard', async () => {
    const user = userEvent.setup();
    const handleView = jest.fn();

    render(<MediaGrid media={mockMediaItems} onMediaView={handleView} />);

    const viewButton = screen.getByText('View Test Movie 1');
    await user.click(viewButton);

    expect(handleView).toHaveBeenCalledTimes(1);
    expect(handleView).toHaveBeenCalledWith(mockMediaItems[0]);
  });

  it('passes onMediaDownload callback to MediaCard', async () => {
    const user = userEvent.setup();
    const handleDownload = jest.fn();

    render(<MediaGrid media={mockMediaItems} onMediaDownload={handleDownload} />);

    const downloadButton = screen.getByText('Download Test Movie 2');
    await user.click(downloadButton);

    expect(handleDownload).toHaveBeenCalledTimes(1);
    expect(handleDownload).toHaveBeenCalledWith(mockMediaItems[1]);
  });

  it('renders with custom className', () => {
    const { container } = render(
      <MediaGrid media={mockMediaItems} className="custom-grid-class" />
    );

    const grid = container.querySelector('.custom-grid-class');
    expect(grid).toBeInTheDocument();
  });

  it('applies grid layout classes', () => {
    const { container } = render(<MediaGrid media={mockMediaItems} />);

    const grid = container.querySelector('.grid');
    expect(grid).toHaveClass('grid-cols-2');
    expect(grid).toHaveClass('sm:grid-cols-3');
    expect(grid).toHaveClass('md:grid-cols-4');
  });

  it('handles single media item', () => {
    const singleItem = [mockMediaItems[0]];
    render(<MediaGrid media={singleItem} />);

    expect(screen.getByText('Test Movie 1')).toBeInTheDocument();
    expect(screen.queryByText('Test Movie 2')).not.toBeInTheDocument();
  });

  it('handles large number of media items', () => {
    const manyItems = Array.from({ length: 50 }, (_, i) => ({
      id: i + 1,
      title: `Movie ${i + 1}`,
      media_type: 'movie',
      directory_path: `/movies/movie${i + 1}`,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }));

    render(<MediaGrid media={manyItems} />);

    const cards = screen.getAllByTestId(/media-card-/);
    expect(cards).toHaveLength(50);
  });

  it('does not render callbacks when not provided', () => {
    render(<MediaGrid media={mockMediaItems} />);

    expect(screen.queryByText(/View/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Download/)).not.toBeInTheDocument();
  });

  it('renders only view callback when only onMediaView provided', () => {
    const handleView = jest.fn();
    render(<MediaGrid media={mockMediaItems} onMediaView={handleView} />);

    expect(screen.getByText('View Test Movie 1')).toBeInTheDocument();
    expect(screen.queryByText('Download Test Movie 1')).not.toBeInTheDocument();
  });

  it('renders only download callback when only onMediaDownload provided', () => {
    const handleDownload = jest.fn();
    render(<MediaGrid media={mockMediaItems} onMediaDownload={handleDownload} />);

    expect(screen.getByText('Download Test Movie 1')).toBeInTheDocument();
    expect(screen.queryByText('View Test Movie 1')).not.toBeInTheDocument();
  });

  it('updates when media prop changes', () => {
    const { rerender } = render(<MediaGrid media={mockMediaItems} />);

    expect(screen.getByText('Test Movie 1')).toBeInTheDocument();

    const newMedia = [
      {
        id: 4,
        title: 'New Movie',
        media_type: 'movie',
        directory_path: '/movies/new',
        created_at: '2024-01-04T00:00:00Z',
        updated_at: '2024-01-04T00:00:00Z',
      },
    ];

    rerender(<MediaGrid media={newMedia} />);

    expect(screen.queryByText('Test Movie 1')).not.toBeInTheDocument();
    expect(screen.getByText('New Movie')).toBeInTheDocument();
  });

  it('maintains proper grid structure with empty slots', () => {
    const { container } = render(<MediaGrid media={mockMediaItems} />);

    const grid = container.querySelector('.grid');
    expect(grid?.children).toHaveLength(3);
  });

  it('renders empty state svg icon', () => {
    render(<MediaGrid media={[]} />);

    const svg = document.querySelector('svg');
    expect(svg).toBeInTheDocument();
    expect(svg).toHaveClass('h-12');
  });
});
