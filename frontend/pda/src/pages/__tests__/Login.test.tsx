// Smoke test: PDA Login page renders the login form.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '@/pages/Login';

// ── Mock useAuth hook ──────────────────────────────────────────────────

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    isAuthenticated: false,
    setTokens: vi.fn(),
    clearTokens: vi.fn(),
  }),
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
    expect(screen.getByPlaceholderText('auth.usernamePlaceholder')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('auth.passwordPlaceholder')).toBeInTheDocument();
    expect(screen.getByText('auth.signIn')).toBeInTheDocument();
  });

  it('shows the app title on the card', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('app.title')).toBeInTheDocument();
    expect(screen.getByText('app.subtitle')).toBeInTheDocument();
  });
});
