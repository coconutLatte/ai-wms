// Smoke test: Login page renders the login form without crashing.
// Mocks the auth hook to simulate an unauthenticated state.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '@/pages/Login';

// ── Mock useAuth hook ──────────────────────────────────────────────────

const mockSetTokens = vi.fn();

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    isAuthenticated: false,
    setTokens: mockSetTokens,
  }),
}));

// ── Mock API ───────────────────────────────────────────────────────────

vi.mock('@/api/auth', () => ({
  login: vi.fn(),
}));

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the login form without crashing', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    // The form should display username/password inputs and a sign-in button
    expect(screen.getByPlaceholderText('auth.username')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('auth.password')).toBeInTheDocument();
    expect(screen.getByText('auth.signIn')).toBeInTheDocument();
  });

  it('shows the admin title on the card', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('auth.adminTitle')).toBeInTheDocument();
  });
});
