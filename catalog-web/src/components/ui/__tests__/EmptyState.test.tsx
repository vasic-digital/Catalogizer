import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { EmptyState } from '../EmptyState';
import { FolderIcon } from 'lucide-react';

describe('EmptyState', () => {
  it('renders title', () => {
    render(<EmptyState title="No items found" />);
    expect(screen.getByText('No items found')).toBeInTheDocument();
  });

  it('renders description when provided', () => {
    render(
      <EmptyState 
        title="No items" 
        description="Try adjusting your filters"
      />
    );
    expect(screen.getByText('Try adjusting your filters')).toBeInTheDocument();
  });

  it('renders icon when provided', () => {
    render(
      <EmptyState 
        title="No items" 
        icon={<FolderIcon data-testid="folder-icon" />}
      />
    );
    expect(screen.getByTestId('folder-icon')).toBeInTheDocument();
  });

  it('renders action when provided', () => {
    render(
      <EmptyState 
        title="No items" 
        action={<button>Create new</button>}
      />
    );
    expect(screen.getByRole('button', { name: 'Create new' })).toBeInTheDocument();
  });

  it('applies custom className', () => {
    const { container } = render(
      <EmptyState title="No items" className="custom-class" />
    );
    expect(container.firstChild).toHaveClass('custom-class');
  });
});
