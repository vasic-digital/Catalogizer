import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from '../AuthContext';

// Mock the API module completely
vi.mock('@/lib/api', async () => ({
  authApi: {
    getAuthStatus: vi.fn(),
    getPermissions: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    updateProfile: vi.fn(),
    changePassword: vi.fn(),
  }
}));

// Mock react-hot-toast
vi.mock('react-hot-toast', async () => ({
  __esModule: true,
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Test component that uses auth context
function TestComponent() {
  const { user, isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return (
    <div>
      <div data-testid="auth-status">
        {isAuthenticated ? 'Authenticated' : 'Not authenticated'}
      </div>
      <div data-testid="user-info">
        {user ? `User: ${user.username}` : 'No user'}
      </div>
    </div>
  );
}

describe('AuthContext', () => {
  const queryClient = new QueryClient();

  it('provides initial unauthenticated state', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </QueryClientProvider>
    );

    // The component will be loading initially
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('throws error when useAuth is used outside provider', () => {
    // Mock console.error for the error boundary
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {/* mock */});

    expect(() => {
      render(
        <QueryClientProvider client={queryClient}>
          <TestComponent />
        </QueryClientProvider>
      );
    }).toThrow('useAuth must be used within an AuthProvider');

    consoleSpy.mockRestore();
  });
});
