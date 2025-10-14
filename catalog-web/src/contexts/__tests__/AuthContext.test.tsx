import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from '../AuthContext';

// Mock the API module completely
jest.mock('@/lib/api', () => ({
  authApi: {
    getAuthStatus: jest.fn(),
    getPermissions: jest.fn(),
    login: jest.fn(),
    register: jest.fn(),
    logout: jest.fn(),
    updateProfile: jest.fn(),
    changePassword: jest.fn(),
  }
}));

// Mock react-hot-toast
jest.mock('react-hot-toast', () => ({
  __esModule: true,
  default: {
    success: jest.fn(),
    error: jest.fn(),
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
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {/* mock */});

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
