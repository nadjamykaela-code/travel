import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { ProtectedRoute } from './ProtectedRoute';
import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockUseAuth = vi.fn();

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => mockUseAuth(),
}));

beforeEach(() => {
  mockUseAuth.mockReset();
});

describe('ProtectedRoute', () => {
  it('shows spinner when loading', () => {
    mockUseAuth.mockReturnValue({ user: null, loading: true });
    render(
      <MemoryRouter>
        <ProtectedRoute><div>content</div></ProtectedRoute>
      </MemoryRouter>
    );
    expect(screen.getByLabelText('Carregando')).toBeInTheDocument();
  });

  it('redirects to login when not authenticated', () => {
    mockUseAuth.mockReturnValue({ user: null, loading: false });
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <ProtectedRoute><div>content</div></ProtectedRoute>
      </MemoryRouter>
    );
    expect(screen.queryByText('content')).not.toBeInTheDocument();
  });

  it('renders children when authenticated', () => {
    mockUseAuth.mockReturnValue({ user: { uid: 'u1', email: 'test@test.com' }, loading: false });
    render(
      <MemoryRouter>
        <ProtectedRoute><div>content</div></ProtectedRoute>
      </MemoryRouter>
    );
    expect(screen.getByText('content')).toBeInTheDocument();
  });
});
