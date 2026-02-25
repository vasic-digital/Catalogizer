/**
 * Example component test
 */
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { generateMediaItems } from '@/test-utils/testData';
import ExampleMediaGrid from '@/components/ExampleMediaGrid';

// Mock component for example
const ExampleMediaGrid = ({ mediaItems, isLoading, onMediaClick }: any) => {
  if (isLoading) {
    return <div data-testid="loading">Loading...</div>;
  }

  return (
    <div data-testid="media-grid">
      {mediaItems.map((item: any) => (
        <div
          key={item.id}
          data-testid={`media-item-${item.id}`}
          onClick={() => onMediaClick(item)}
        >
          <h3>{item.title}</h3>
          <p>{item.year}</p>
        </div>
      ))}
    </div>
  );
};

describe('ExampleMediaGrid', () => {
  it('should render loading state', () => {
    // Given
    const props = {
      mediaItems: [],
      isLoading: true,
      onMediaClick: vi.fn(),
    };

    // When
    render(<ExampleMediaGrid {...props} />);

    // Then
    expect(screen.getByTestId('loading')).toBeInTheDocument();
    expect(screen.queryByTestId('media-grid')).not.toBeInTheDocument();
  });

  it('should render media items', () => {
    // Given
    const mediaItems = generateMediaItems(3);
    const onMediaClick = vi.fn();
    const props = {
      mediaItems,
      isLoading: false,
      onMediaClick,
    };

    // When
    render(<ExampleMediaGrid {...props} />);

    // Then
    expect(screen.getByTestId('media-grid')).toBeInTheDocument();
    expect(screen.getAllByTestId(/media-item-/)).toHaveLength(3);
    
    mediaItems.forEach(item => {
      expect(screen.getByText(item.title)).toBeInTheDocument();
      expect(screen.getByText(item.year.toString())).toBeInTheDocument();
    });
  });

  it('should call onMediaClick when item is clicked', () => {
    // Given
    const mediaItems = generateMediaItems(2);
    const onMediaClick = vi.fn();
    const props = {
      mediaItems,
      isLoading: false,
      onMediaClick,
    };

    // When
    render(<ExampleMediaGrid {...props} />);
    const firstItem = screen.getByTestId('media-item-1');
    fireEvent.click(firstItem);

    // Then
    expect(onMediaClick).toHaveBeenCalledTimes(1);
    expect(onMediaClick).toHaveBeenCalledWith(mediaItems[0]);
  });
});
