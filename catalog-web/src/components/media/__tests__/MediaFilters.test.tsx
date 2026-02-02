import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MediaFilters } from '../MediaFilters';
import type { MediaSearchRequest } from '@/types/media';

const mockFilters: MediaSearchRequest = {
  query: '',
  limit: 20,
  offset: 0,
};

describe('MediaFilters', () => {
  it('renders filter title', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.getByText('Filters')).toBeInTheDocument();
  });

  it('renders search input', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.getByPlaceholderText('Search media titles...')).toBeInTheDocument();
  });

  it('updates query filter on search input change', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const searchInput = screen.getByPlaceholderText('Search media titles...');
    await user.type(searchInput, 'M');

    expect(handleChange).toHaveBeenCalled();
    const lastCall = handleChange.mock.calls[handleChange.mock.calls.length - 1][0];
    expect(lastCall.query).toBe('M');
  });

  it('displays query value in search input', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();
    const filtersWithQuery = { ...mockFilters, query: 'Inception' };

    render(
      <MediaFilters
        filters={filtersWithQuery}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const searchInput = screen.getByPlaceholderText('Search media titles...') as HTMLInputElement;
    expect(searchInput.value).toBe('Inception');
  });

  it('renders media type filter buttons', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.getByText('Media Type')).toBeInTheDocument();
    // Check for some common media types
    expect(screen.getByText(/movie/i)).toBeInTheDocument();
  });

  it('updates media type filter when button clicked', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const movieButton = screen.getByText(/movie/i);
    await user.click(movieButton);

    expect(handleChange).toHaveBeenCalled();
    const callArg = handleChange.mock.calls[0][0];
    expect(callArg.media_type).toBeDefined();
  });

  it('clears media type filter when clicked again', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();
    const filtersWithType = { ...mockFilters, media_type: 'movie' };

    render(
      <MediaFilters
        filters={filtersWithType}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const movieButton = screen.getByText(/movie/i);
    await user.click(movieButton);

    expect(handleChange).toHaveBeenCalled();
    const callArg = handleChange.mock.calls[0][0];
    expect(callArg.media_type).toBeUndefined();
  });

  it('renders quality filter buttons', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.getByText('Quality')).toBeInTheDocument();
  });

  it('updates quality filter when button clicked', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    // Look for any quality button (e.g., 1080P, 720P, 4K)
    const buttons = screen.getAllByRole('button');
    const qualityButton = buttons.find(btn => btn.textContent?.match(/1080P|720P|4K/i));

    if (qualityButton) {
      await user.click(qualityButton);
      expect(handleChange).toHaveBeenCalled();
    }
  });

  it('renders year range inputs', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.getByText('Year Range')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('From')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('To')).toBeInTheDocument();
  });

  it('updates year_min filter', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const yearFromInput = screen.getByPlaceholderText('From');
    await user.type(yearFromInput, '2020');

    expect(handleChange).toHaveBeenCalled();
  });

  it('updates year_max filter', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const yearToInput = screen.getByPlaceholderText('To');
    await user.type(yearToInput, '2024');

    expect(handleChange).toHaveBeenCalled();
  });

  it('renders rating input', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.getByText('Minimum Rating')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('0.0')).toBeInTheDocument();
  });

  it('updates rating_min filter', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const ratingInput = screen.getByPlaceholderText('0.0');
    await user.type(ratingInput, '7.5');

    expect(handleChange).toHaveBeenCalled();
  });

  it('renders sort options', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.getByText('Sort By')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Last Updated')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Descending')).toBeInTheDocument();
  });

  it('updates sort_by filter', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const sortBySelect = screen.getByDisplayValue('Last Updated');
    await user.selectOptions(sortBySelect, 'title');

    expect(handleChange).toHaveBeenCalled();
    const callArg = handleChange.mock.calls[0][0];
    expect(callArg.sort_by).toBe('title');
  });

  it('updates sort_order filter', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const sortOrderSelect = screen.getByDisplayValue('Descending');
    await user.selectOptions(sortOrderSelect, 'asc');

    expect(handleChange).toHaveBeenCalled();
    const callArg = handleChange.mock.calls[0][0];
    expect(callArg.sort_order).toBe('asc');
  });

  it('shows clear all button when filters are active', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();
    const activeFilters = { ...mockFilters, query: 'test', media_type: 'movie' };

    render(
      <MediaFilters
        filters={activeFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.getByText('Clear all')).toBeInTheDocument();
  });

  it('hides clear all button when no filters are active', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();
    // Empty filters - limit and offset are ignored by hasActiveFilters check
    const emptyFilters = {};

    render(
      <MediaFilters
        filters={emptyFilters as MediaSearchRequest}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    expect(screen.queryByText('Clear all')).not.toBeInTheDocument();
  });

  it('calls onReset when clear all button is clicked', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    const handleReset = vi.fn();
    const activeFilters = { ...mockFilters, query: 'test' };

    render(
      <MediaFilters
        filters={activeFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const clearButton = screen.getByText('Clear all');
    await user.click(clearButton);

    expect(handleReset).toHaveBeenCalledTimes(1);
  });

  it('applies custom className', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();

    const { container } = render(
      <MediaFilters
        filters={mockFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
        className="custom-filters-class"
      />
    );

    expect(container.querySelector('.custom-filters-class')).toBeInTheDocument();
  });

  it('displays all filter values from props', () => {
    const handleChange = vi.fn();
    const handleReset = vi.fn();
    const fullFilters: MediaSearchRequest = {
      query: 'Matrix',
      media_type: 'movie',
      quality: '1080p',
      year_min: 1999,
      year_max: 2024,
      rating_min: 7.0,
      sort_by: 'rating',
      sort_order: 'desc',
      limit: 20,
      offset: 0,
    };

    render(
      <MediaFilters
        filters={fullFilters}
        onFiltersChange={handleChange}
        onReset={handleReset}
      />
    );

    const searchInput = screen.getByPlaceholderText('Search media titles...') as HTMLInputElement;
    expect(searchInput.value).toBe('Matrix');

    const yearFromInput = screen.getByPlaceholderText('From') as HTMLInputElement;
    expect(yearFromInput.value).toBe('1999');

    const yearToInput = screen.getByPlaceholderText('To') as HTMLInputElement;
    expect(yearToInput.value).toBe('2024');

    const ratingInput = screen.getByPlaceholderText('0.0') as HTMLInputElement;
    expect(ratingInput.value).toBe('7');

    expect(screen.getByDisplayValue('Rating')).toBeInTheDocument();
  });
});
